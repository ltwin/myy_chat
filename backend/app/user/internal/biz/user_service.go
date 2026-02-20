// Package biz 用户业务逻辑层
package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/myy-chat/backend/pkg/snowflake"
)

// UserService 用户业务服务
type UserService struct {
	userRepo    UserRepo
	profileRepo UserProfileRepo
	creditRepo  CreditAccountRepo
	sessionRepo SessionRepo
	idGen       snowflake.Generator
	log         *log.Helper
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo UserRepo,
	profileRepo UserProfileRepo,
	creditRepo CreditAccountRepo,
	sessionRepo SessionRepo,
	idGen snowflake.Generator,
	logger log.Logger,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		creditRepo:  creditRepo,
		sessionRepo: sessionRepo,
		idGen:       idGen,
		log:         log.NewHelper(logger),
	}
}

// RegisterInput 注册输入
type RegisterInput struct {
	Username string
	Email    string
	Password string
	Phone    string
}

// RegisterOutput 注册输出
type RegisterOutput struct {
	User         *User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Register 用户注册 (FR-001)
// 创建用户、用户画像、积分账户，并发放初始积分
func (s *UserService) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// 检查手机号是否已存在 (如果提供)
	if input.Phone != "" {
		exists, err = s.userRepo.ExistsByPhone(ctx, input.Phone)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPhoneAlreadyExists
		}
	}

	// 生成雪花ID
	userID := s.idGen.Generate()

	// 创建用户实体
	user, err := NewUser(userID, input.Username, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	// 设置手机号
	if input.Phone != "" {
		user.Phone = input.Phone
	}

	// 保存用户
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 创建用户画像
	profile := NewUserProfile(userID)
	if err := s.profileRepo.Create(ctx, profile); err != nil {
		s.log.Warnf("failed to create user profile: %v", err)
		// 非致命错误，继续
	}

	// 创建积分账户并发放初始积分 (FR-031a: 100积分)
	creditAccount := NewCreditAccount(userID)
	if err := s.creditRepo.Create(ctx, creditAccount); err != nil {
		s.log.Warnf("failed to create credit account: %v", err)
		// 非致命错误，继续
	}

	// 发放初始积分
	initialCredits := 100.0
	if err := s.creditRepo.AddCredits(ctx, userID, initialCredits, "registration_bonus", "registration", userID); err != nil {
		s.log.Warnf("failed to grant initial credits: %v", err)
	}

	s.log.Infof("user registered: id=%d, email=%s", userID, input.Email)

	// TODO: 生成 JWT token
	return &RegisterOutput{
		User:         user,
		AccessToken:  "", // 由调用方生成
		RefreshToken: "",
		ExpiresIn:    3600,
	}, nil
}

// LoginInput 登录输入
type LoginInput struct {
	Email    string // 邮箱或用户名
	Password string
}

// LoginOutput 登录输出
type LoginOutput struct {
	User         *User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Login 用户登录 (FR-003)
func (s *UserService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	// 尝试通过邮箱查找用户
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == ErrUserNotFound {
		// 尝试通过用户名查找
		user, err = s.userRepo.GetByUsername(ctx, input.Email)
	}

	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 检查用户是否可以登录
	if err := user.CanLogin(); err != nil {
		return nil, err
	}

	// 验证密码
	if !user.CheckPassword(input.Password) {
		return nil, ErrInvalidCredentials
	}

	// 记录登录
	user.RecordLogin()
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Warnf("failed to update last login: %v", err)
	}

	s.log.Infof("user logged in: id=%d", user.ID)

	// TODO: 生成 JWT token
	return &LoginOutput{
		User:         user,
		AccessToken:  "",
		RefreshToken: "",
		ExpiresIn:    3600,
	}, nil
}

// IssueSessionOutput 创建会话输出
type IssueSessionOutput struct {
	Session      *Session
	RefreshToken string
}

// IssueSession 创建服务端会话并发放 refresh token（明文仅返回给调用方）
func (s *UserService) IssueSession(ctx context.Context, userID int64, userAgent, ipAddress string) (*IssueSessionOutput, error) {
	sessionID := s.idGen.Generate()
	session, refreshToken, err := NewSession(sessionID, userID, userAgent, ipAddress, DefaultSessionConfig)
	if err != nil {
		return nil, err
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &IssueSessionOutput{
		Session:      session,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshSessionOutput 刷新会话输出
type RefreshSessionOutput struct {
	User         *User
	Session      *Session
	RefreshToken string
}

// RefreshSession 通过 refresh token 刷新会话并执行 token rotation
func (s *UserService) RefreshSession(ctx context.Context, refreshToken string) (*RefreshSessionOutput, error) {
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	session, err := s.sessionRepo.GetByRefreshTokenHash(ctx, HashRefreshToken(refreshToken))
	if err != nil {
		// 对外统一返回无效 refresh token，避免会话枚举。
		if err == ErrSessionExpired {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	if !session.VerifyRefreshToken(refreshToken) {
		return nil, ErrInvalidRefreshToken
	}
	if session.IsRevoked() {
		return nil, ErrSessionRevoked
	}
	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := session.RotateRefreshToken(DefaultSessionConfig.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return &RefreshSessionOutput{
		User:         user,
		Session:      session,
		RefreshToken: newRefreshToken,
	}, nil
}

// RevokeSession 撤销指定会话（当前设备登出）
func (s *UserService) RevokeSession(ctx context.Context, sessionID int64) error {
	if sessionID <= 0 {
		return ErrInvalidRefreshToken
	}
	return s.sessionRepo.Revoke(ctx, sessionID)
}

// RevokeAllSessions 撤销用户所有活跃会话（全部设备登出）
func (s *UserService) RevokeAllSessions(ctx context.Context, userID int64) ([]*Session, error) {
	sessions, err := s.sessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.sessionRepo.RevokeAllByUserID(ctx, userID); err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(ctx context.Context, userID int64) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.IsDeleted {
		return nil, ErrUserDeleted
	}

	return user, nil
}

// UpdateUserInput 更新用户输入
type UpdateUserInput struct {
	UserID    int64
	Username  *string
	Phone     *string
	AvatarURL *string
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(ctx context.Context, input UpdateUserInput) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// 更新用户名
	if input.Username != nil && *input.Username != user.Username {
		// 检查用户名是否已存在
		exists, err := s.userRepo.ExistsByUsername(ctx, *input.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUsernameAlreadyExists
		}
		if err := user.UpdateUsername(*input.Username); err != nil {
			return nil, err
		}
	}

	// 更新手机号
	if input.Phone != nil && *input.Phone != user.Phone {
		if *input.Phone != "" {
			exists, err := s.userRepo.ExistsByPhone(ctx, *input.Phone)
			if err != nil {
				return nil, err
			}
			if exists {
				return nil, ErrPhoneAlreadyExists
			}
		}
		user.UpdatePhone(*input.Phone)
	}

	// 更新头像
	if input.AvatarURL != nil {
		user.UpdateAvatarURL(*input.AvatarURL)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdatePassword 更新密码 (FR-003c)
func (s *UserService) UpdatePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := user.UpdatePassword(currentPassword, newPassword); err != nil {
		return err
	}

	return s.userRepo.Update(ctx, user)
}

// GetProfile 获取用户画像
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*UserProfile, error) {
	return s.profileRepo.GetByUserID(ctx, userID)
}

// UpdateProfileInput 更新画像输入
type UpdateProfileInput struct {
	UserID     int64
	FullName   string
	Gender     string
	BirthDate  string
	Location   *Location
	Interests  []string
	Occupation string
	Bio        string
}

// UpdateProfile 更新用户画像
func (s *UserService) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*UserProfile, error) {
	profile, err := s.profileRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		// 如果画像不存在，创建新的
		if err == ErrUserNotFound {
			profile = NewUserProfile(input.UserID)
		} else {
			return nil, err
		}
	}

	// 解析生日
	birthDate, err := ParseDate(input.BirthDate)
	if err != nil {
		return nil, err
	}

	profile.Update(
		input.FullName,
		input.Gender,
		birthDate,
		input.Location,
		input.Interests,
		input.Occupation,
		input.Bio,
	)

	if err := s.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// RequestAccountDeletion 请求账号删除 (FR-004a)
func (s *UserService) RequestAccountDeletion(ctx context.Context, userID int64, password string) (time.Time, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return time.Time{}, err
	}

	// 验证密码
	if !user.CheckPassword(password) {
		return time.Time{}, ErrPasswordMismatch
	}

	// 计划删除
	deletionTime := user.ScheduleDeletion()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return time.Time{}, err
	}

	s.log.Infof("user deletion scheduled: id=%d, scheduled_at=%v", userID, deletionTime)
	return deletionTime, nil
}

// CancelAccountDeletion 取消账号删除 (FR-004b)
func (s *UserService) CancelAccountDeletion(ctx context.Context, userID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsDeletionScheduled() {
		return nil // 无需取消
	}

	user.CancelDeletion()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	s.log.Infof("user deletion cancelled: id=%d", userID)
	return nil
}

// VerifyEmail 验证邮箱 (FR-002)
func (s *UserService) VerifyEmail(ctx context.Context, userID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.MarkEmailVerified()
	return s.userRepo.Update(ctx, user)
}

// GetCreditBalance 获取用户积分余额
func (s *UserService) GetCreditBalance(ctx context.Context, userID int64) (float64, error) {
	account, err := s.creditRepo.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}
