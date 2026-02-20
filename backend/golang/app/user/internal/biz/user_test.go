package biz

import (
	"testing"
	"time"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		username string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "valid user",
			id:       1234567890,
			username: "testuser",
			email:    "test@example.com",
			password: "password123",
			wantErr:  nil,
		},
		{
			name:     "username too short",
			id:       1234567890,
			username: "ab",
			email:    "test@example.com",
			password: "password123",
			wantErr:  ErrInvalidUsername,
		},
		{
			name:     "invalid email format",
			id:       1234567890,
			username: "testuser",
			email:    "invalid-email",
			password: "password123",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "password too short",
			id:       1234567890,
			username: "testuser",
			email:    "test@example.com",
			password: "short",
			wantErr:  ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.id, tt.username, tt.email, tt.password)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error = %v", err)
				return
			}

			if user.ID != tt.id {
				t.Errorf("user.ID = %v, want %v", user.ID, tt.id)
			}
			if user.Username != tt.username {
				t.Errorf("user.Username = %v, want %v", user.Username, tt.username)
			}
			if user.Email != tt.email {
				t.Errorf("user.Email = %v, want %v", user.Email, tt.email)
			}
			if user.PasswordHash == "" {
				t.Error("user.PasswordHash should not be empty")
			}
			if user.EmailVerified {
				t.Error("user.EmailVerified should be false by default")
			}
		})
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user, err := NewUser(1, "testuser", "test@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			password: "correctpassword",
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.CheckPassword(tt.password); got != tt.want {
				t.Errorf("User.CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdatePassword(t *testing.T) {
	user, err := NewUser(1, "testuser", "test@example.com", "oldpassword1")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		wantErr         error
	}{
		{
			name:            "successful update",
			currentPassword: "oldpassword1",
			newPassword:     "newpassword1",
			wantErr:         nil,
		},
		{
			name:            "wrong current password",
			currentPassword: "wrongpassword",
			newPassword:     "newpassword1",
			wantErr:         ErrPasswordMismatch,
		},
		{
			name:            "new password too short",
			currentPassword: "newpassword1",
			newPassword:     "short",
			wantErr:         ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.UpdatePassword(tt.currentPassword, tt.newPassword)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("User.UpdatePassword() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("User.UpdatePassword() unexpected error = %v", err)
				return
			}

			// 验证新密码可用
			if !user.CheckPassword(tt.newPassword) {
				t.Error("new password should be valid after update")
			}
		})
	}
}

func TestUser_ScheduleDeletion(t *testing.T) {
	user, _ := NewUser(1, "testuser", "test@example.com", "password123")

	if user.IsDeletionScheduled() {
		t.Error("new user should not have deletion scheduled")
	}

	deletionTime := user.ScheduleDeletion()

	if !user.IsDeletionScheduled() {
		t.Error("user should have deletion scheduled after calling ScheduleDeletion")
	}

	// 删除时间应该在30天后
	expectedTime := time.Now().AddDate(0, 0, 30)
	diff := deletionTime.Sub(expectedTime)
	if diff > time.Minute || diff < -time.Minute {
		t.Errorf("deletion time should be around 30 days from now, got diff: %v", diff)
	}
}

func TestUser_CancelDeletion(t *testing.T) {
	user, _ := NewUser(1, "testuser", "test@example.com", "password123")

	user.ScheduleDeletion()
	if !user.IsDeletionScheduled() {
		t.Fatal("deletion should be scheduled")
	}

	user.CancelDeletion()
	if user.IsDeletionScheduled() {
		t.Error("deletion should be cancelled")
	}
}

func TestUser_CanLogin(t *testing.T) {
	user, _ := NewUser(1, "testuser", "test@example.com", "password123")

	// 正常用户可以登录
	if err := user.CanLogin(); err != nil {
		t.Errorf("normal user should be able to login, got error: %v", err)
	}

	// 已删除用户不能登录
	user.MarkDeleted()
	if err := user.CanLogin(); err != ErrUserDeleted {
		t.Errorf("deleted user should not be able to login, got error: %v", err)
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"test@example.com", false},
		{"user.name@domain.org", false},
		{"user+tag@example.com", false},
		{"invalid", true},
		{"@domain.com", true},
		{"user@", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		username string
		wantErr  bool
	}{
		{"abc", false},             // 最小长度
		{"testuser", false},        // 正常长度
		{"用户名", false},             // 中文 (3个字符)
		{"ab", true},               // 太短
		{"a", true},                // 太短
		{"", true},                 // 空
		{string(make([]byte, 51)), true}, // 太长
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		wantErr  bool
	}{
		{"12345678", false},   // 最小长度
		{"password123", false},
		{"1234567", true},     // 太短
		{"", true},            // 空
		{string(make([]byte, 129)), true}, // 太长
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
