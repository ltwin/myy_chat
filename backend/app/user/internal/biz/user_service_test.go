package biz

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// mockUserRepo 模拟用户仓储
type mockUserRepo struct {
	users       map[int64]*User
	emailIndex  map[string]*User
	usernameIdx map[string]*User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:       make(map[int64]*User),
		emailIndex:  make(map[string]*User),
		usernameIdx: make(map[string]*User),
	}
}

func (r *mockUserRepo) Create(ctx context.Context, user *User) error {
	r.users[user.ID] = user
	r.emailIndex[user.Email] = user
	r.usernameIdx[user.Username] = user
	return nil
}

func (r *mockUserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	if user, ok := r.emailIndex[email]; ok {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (r *mockUserRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	if user, ok := r.usernameIdx[username]; ok {
		return user, nil
	}
	return nil, ErrUserNotFound
}

func (r *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*User, error) {
	return nil, ErrUserNotFound
}

func (r *mockUserRepo) Update(ctx context.Context, user *User) error {
	if _, ok := r.users[user.ID]; !ok {
		return ErrUserNotFound
	}
	r.users[user.ID] = user
	r.emailIndex[user.Email] = user
	r.usernameIdx[user.Username] = user
	return nil
}

func (r *mockUserRepo) Delete(ctx context.Context, id int64) error {
	if user, ok := r.users[id]; ok {
		user.IsDeleted = true
		return nil
	}
	return ErrUserNotFound
}

func (r *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, ok := r.emailIndex[email]
	return ok, nil
}

func (r *mockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	_, ok := r.usernameIdx[username]
	return ok, nil
}

func (r *mockUserRepo) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	return false, nil
}

func (r *mockUserRepo) ListScheduledForDeletion(ctx context.Context, before time.Time) ([]*User, error) {
	return nil, nil
}

// mockProfileRepo 模拟画像仓储
type mockProfileRepo struct {
	profiles map[int64]*UserProfile
}

func newMockProfileRepo() *mockProfileRepo {
	return &mockProfileRepo{
		profiles: make(map[int64]*UserProfile),
	}
}

func (r *mockProfileRepo) Create(ctx context.Context, profile *UserProfile) error {
	r.profiles[profile.UserID] = profile
	return nil
}

func (r *mockProfileRepo) GetByUserID(ctx context.Context, userID int64) (*UserProfile, error) {
	if p, ok := r.profiles[userID]; ok {
		return p, nil
	}
	return nil, ErrUserNotFound
}

func (r *mockProfileRepo) Update(ctx context.Context, profile *UserProfile) error {
	r.profiles[profile.UserID] = profile
	return nil
}

func (r *mockProfileRepo) Delete(ctx context.Context, userID int64) error {
	delete(r.profiles, userID)
	return nil
}

// mockCreditRepo 模拟积分仓储
type mockCreditRepo struct {
	accounts map[int64]*CreditAccount
}

func newMockCreditRepo() *mockCreditRepo {
	return &mockCreditRepo{
		accounts: make(map[int64]*CreditAccount),
	}
}

func (r *mockCreditRepo) Create(ctx context.Context, account *CreditAccount) error {
	r.accounts[account.UserID] = account
	return nil
}

func (r *mockCreditRepo) GetByUserID(ctx context.Context, userID int64) (*CreditAccount, error) {
	if a, ok := r.accounts[userID]; ok {
		return a, nil
	}
	return nil, ErrCreditAccountNotFound
}

func (r *mockCreditRepo) Update(ctx context.Context, account *CreditAccount) error {
	r.accounts[account.UserID] = account
	return nil
}

func (r *mockCreditRepo) AddCredits(ctx context.Context, userID int64, amount float64, reason, refType string, refID int64) error {
	if account, ok := r.accounts[userID]; ok {
		account.AddCredits(amount)
		return nil
	}
	return ErrCreditAccountNotFound
}

func (r *mockCreditRepo) DeductCredits(ctx context.Context, userID int64, amount float64, reason, refType string, refID int64, idempotencyKey string) error {
	if account, ok := r.accounts[userID]; ok {
		return account.DeductCredits(amount)
	}
	return ErrCreditAccountNotFound
}

func (r *mockCreditRepo) GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*CreditTransaction, int, error) {
	return nil, 0, nil
}

func (r *mockCreditRepo) Delete(ctx context.Context, userID int64) error {
	delete(r.accounts, userID)
	return nil
}

// mockSessionRepo 模拟会话仓储
type mockSessionRepo struct {
	sessions map[int64]*Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions: make(map[int64]*Session),
	}
}

func (r *mockSessionRepo) Create(ctx context.Context, session *Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSessionRepo) GetByID(ctx context.Context, id int64) (*Session, error) {
	if s, ok := r.sessions[id]; ok {
		return s, nil
	}
	return nil, ErrSessionExpired
}

func (r *mockSessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*Session, error) {
	return nil, ErrSessionExpired
}

func (r *mockSessionRepo) Update(ctx context.Context, session *Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSessionRepo) Revoke(ctx context.Context, id int64) error {
	if s, ok := r.sessions[id]; ok {
		s.Revoke()
		return nil
	}
	return nil
}

func (r *mockSessionRepo) RevokeAllByUserID(ctx context.Context, userID int64) error {
	return nil
}

func (r *mockSessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

func (r *mockSessionRepo) ListByUserID(ctx context.Context, userID int64) ([]*Session, error) {
	return nil, nil
}

// mockIDGen 模拟ID生成器
type mockIDGen struct {
	counter int64
}

func (g *mockIDGen) Generate() int64 {
	g.counter++
	return g.counter
}

func (g *mockIDGen) GenerateString() string {
	return ""
}

func (g *mockIDGen) Parse(id int64) (nodeID, timestamp, sequence int64) {
	return 0, 0, 0
}

func TestUserService_Register(t *testing.T) {
	userRepo := newMockUserRepo()
	profileRepo := newMockProfileRepo()
	creditRepo := newMockCreditRepo()
	sessionRepo := newMockSessionRepo()
	idGen := &mockIDGen{}

	service := NewUserService(userRepo, profileRepo, creditRepo, sessionRepo, idGen, log.DefaultLogger)

	ctx := context.Background()

	// 测试正常注册
	output, err := service.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.User.ID == 0 {
		t.Error("user ID should not be 0")
	}
	if output.User.Username != "testuser" {
		t.Errorf("username = %s, want testuser", output.User.Username)
	}
	if output.User.Email != "test@example.com" {
		t.Errorf("email = %s, want test@example.com", output.User.Email)
	}

	// 测试重复邮箱
	_, err = service.Register(ctx, RegisterInput{
		Username: "another",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}

	// 测试重复用户名
	_, err = service.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "another@example.com",
		Password: "password123",
	})
	if err != ErrUsernameAlreadyExists {
		t.Errorf("expected ErrUsernameAlreadyExists, got %v", err)
	}
}

func TestUserService_Login(t *testing.T) {
	userRepo := newMockUserRepo()
	profileRepo := newMockProfileRepo()
	creditRepo := newMockCreditRepo()
	sessionRepo := newMockSessionRepo()
	idGen := &mockIDGen{}

	service := NewUserService(userRepo, profileRepo, creditRepo, sessionRepo, idGen, log.DefaultLogger)

	ctx := context.Background()

	// 先注册一个用户
	service.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})

	// 测试正确登录
	output, err := service.Login(ctx, LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.User.Email != "test@example.com" {
		t.Error("login should return correct user")
	}

	// 测试错误密码
	_, err = service.Login(ctx, LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// 测试不存在的用户
	_, err = service.Login(ctx, LoginInput{
		Email:    "notexist@example.com",
		Password: "password123",
	})
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	userRepo := newMockUserRepo()
	profileRepo := newMockProfileRepo()
	creditRepo := newMockCreditRepo()
	sessionRepo := newMockSessionRepo()
	idGen := &mockIDGen{}

	service := NewUserService(userRepo, profileRepo, creditRepo, sessionRepo, idGen, log.DefaultLogger)

	ctx := context.Background()

	// 注册用户
	output, _ := service.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	userID := output.User.ID

	// 更新用户名
	newUsername := "newname"
	updated, err := service.UpdateUser(ctx, UpdateUserInput{
		UserID:   userID,
		Username: &newUsername,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Username != "newname" {
		t.Errorf("username = %s, want newname", updated.Username)
	}
}

func TestUserService_RequestAccountDeletion(t *testing.T) {
	userRepo := newMockUserRepo()
	profileRepo := newMockProfileRepo()
	creditRepo := newMockCreditRepo()
	sessionRepo := newMockSessionRepo()
	idGen := &mockIDGen{}

	service := NewUserService(userRepo, profileRepo, creditRepo, sessionRepo, idGen, log.DefaultLogger)

	ctx := context.Background()

	// 注册用户
	output, _ := service.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	userID := output.User.ID

	// 请求删除 - 错误密码
	_, err := service.RequestAccountDeletion(ctx, userID, "wrongpassword")
	if err != ErrPasswordMismatch {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}

	// 请求删除 - 正确密码
	deletionTime, err := service.RequestAccountDeletion(ctx, userID, "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 删除时间应该在30天后
	expectedTime := time.Now().AddDate(0, 0, 30)
	diff := deletionTime.Sub(expectedTime)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("deletion time should be around 30 days from now")
	}

	// 取消删除
	err = service.CancelAccountDeletion(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证取消
	user, _ := service.GetUser(ctx, userID)
	if user.IsDeletionScheduled() {
		t.Error("deletion should be cancelled")
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	userRepo := newMockUserRepo()
	profileRepo := newMockProfileRepo()
	creditRepo := newMockCreditRepo()
	sessionRepo := newMockSessionRepo()
	idGen := &mockIDGen{}

	service := NewUserService(userRepo, profileRepo, creditRepo, sessionRepo, idGen, log.DefaultLogger)

	ctx := context.Background()

	// 注册用户
	output, _ := service.Register(ctx, RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	userID := output.User.ID

	// 更新密码 - 错误的当前密码
	err := service.UpdatePassword(ctx, userID, "wrongpassword", "newpassword1")
	if err != ErrPasswordMismatch {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}

	// 更新密码 - 正确
	err = service.UpdatePassword(ctx, userID, "password123", "newpassword1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 验证新密码可用
	_, err = service.Login(ctx, LoginInput{
		Email:    "test@example.com",
		Password: "newpassword1",
	})
	if err != nil {
		t.Error("login with new password should succeed")
	}
}
