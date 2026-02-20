// Package biz 用户业务逻辑层
package biz

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// 用户相关错误定义
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrPhoneAlreadyExists    = errors.New("phone already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserDeleted           = errors.New("user has been deleted")
	ErrUserDeletionPending   = errors.New("user deletion is pending")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidUsername       = errors.New("invalid username: must be 3-50 characters")
	ErrInvalidPassword       = errors.New("invalid password: must be 8-128 characters")
	ErrPasswordMismatch      = errors.New("current password is incorrect")
	ErrEmailNotVerified      = errors.New("email not verified")
	ErrAccountLocked         = errors.New("account is locked due to too many failed attempts")
	ErrSessionExpired        = errors.New("session expired")
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionRevoked        = errors.New("session has been revoked")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
)

// emailRegex 邮箱格式验证正则
var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

const passwordHashCost = 12

// User 用户实体
type User struct {
	ID                  int64      // 雪花ID
	Username            string     // 用户名 (3-50字符)
	Email               string     // 邮箱
	PasswordHash        string     // 密码哈希
	Phone               string     // 手机号 (可选)
	AvatarURL           string     // 头像URL
	EmailVerified       bool       // 邮箱是否验证
	CreatedAt           time.Time  // 创建时间
	UpdatedAt           time.Time  // 更新时间
	LastLoginAt         *time.Time // 最后登录时间
	IsDeleted           bool       // 是否已删除
	DeletionScheduledAt *time.Time // 计划删除时间
}

// NewUser 创建新用户
func NewUser(id int64, username, email, password string) (*User, error) {
	normalizedEmail := NormalizeEmail(email)

	// 验证用户名
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	// 验证邮箱
	if err := ValidateEmail(normalizedEmail); err != nil {
		return nil, err
	}

	// 验证密码
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// 生成密码哈希
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:            id,
		Username:      username,
		Email:         normalizedEmail,
		PasswordHash:  hash,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) error {
	length := utf8.RuneCountInString(username)
	if length < 3 || length > 50 {
		return ErrInvalidUsername
	}
	return nil
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(NormalizeEmail(email)) {
		return ErrInvalidEmail
	}
	return nil
}

// NormalizeEmail 归一化邮箱为 lower-case
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	length := len(password)
	if length < 8 || length > 128 {
		return ErrInvalidPassword
	}
	return nil
}

// HashPassword 生成密码哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword 更新密码
func (u *User) UpdatePassword(currentPassword, newPassword string) error {
	// 验证当前密码
	if !u.CheckPassword(currentPassword) {
		return ErrPasswordMismatch
	}

	// 验证新密码
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	// 生成新密码哈希
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	u.PasswordHash = hash
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateUsername 更新用户名
func (u *User) UpdateUsername(username string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	u.Username = username
	u.UpdatedAt = time.Now()
	return nil
}

// UpdatePhone 更新手机号
func (u *User) UpdatePhone(phone string) {
	u.Phone = phone
	u.UpdatedAt = time.Now()
}

// UpdateAvatarURL 更新头像
func (u *User) UpdateAvatarURL(url string) {
	u.AvatarURL = url
	u.UpdatedAt = time.Now()
}

// MarkEmailVerified 标记邮箱已验证
func (u *User) MarkEmailVerified() {
	u.EmailVerified = true
	u.UpdatedAt = time.Now()
}

// RecordLogin 记录登录
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// ScheduleDeletion 计划删除账号 (30天冷静期)
func (u *User) ScheduleDeletion() time.Time {
	deletionTime := time.Now().AddDate(0, 0, 30) // 30天后
	u.DeletionScheduledAt = &deletionTime
	u.UpdatedAt = time.Now()
	return deletionTime
}

// CancelDeletion 取消删除账号
func (u *User) CancelDeletion() {
	u.DeletionScheduledAt = nil
	u.UpdatedAt = time.Now()
}

// MarkDeleted 标记用户已删除
func (u *User) MarkDeleted() {
	u.IsDeleted = true
	u.UpdatedAt = time.Now()
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() error {
	// 设计约定：处于删除冷静期（DeletionScheduledAt != nil）的账号仍允许登录，
	// 以便用户在 30 天内执行取消删除。
	if u.IsDeleted {
		return ErrUserDeleted
	}
	return nil
}

// IsDeletionScheduled 检查是否已计划删除
func (u *User) IsDeletionScheduled() bool {
	return u.DeletionScheduledAt != nil
}

// UserRepo 用户仓储接口
type UserRepo interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error
	// CreateWithInitialResources 在单事务内创建用户、用户画像与初始积分账户
	CreateWithInitialResources(ctx context.Context, user *User, profile *UserProfile, initialCredits int64) error
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id int64) (*User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(ctx context.Context, email string) (*User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)
	// GetByPhone 根据手机号获取用户
	GetByPhone(ctx context.Context, phone string) (*User, error)
	// Update 更新用户
	Update(ctx context.Context, user *User) error
	// Delete 删除用户 (软删除)
	Delete(ctx context.Context, id int64) error
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	// ExistsByPhone 检查手机号是否存在
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	// ListScheduledForDeletion 列出计划删除的用户
	ListScheduledForDeletion(ctx context.Context, before time.Time) ([]*User, error)
}
