// Package data 用户数据访问层
package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"

	"github.com/myy-chat/backend/app/user/internal/biz"
)

// userRepo 用户仓储实现
type userRepo struct {
	db  *pgxpool.Pool
	log *log.Helper
}

// NewUserRepo 创建用户仓储
func NewUserRepo(db *pgxpool.Pool, logger log.Logger) biz.UserRepo {
	return &userRepo{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, user *biz.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		biz.NormalizeEmail(user.Email),
		user.PasswordHash,
		nullString(user.Phone),
		nullString(user.AvatarURL),
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
		user.IsDeleted,
		nullTime(user.DeletionScheduledAt),
	)
	return err
}

// CreateWithInitialResources 在单事务中创建用户、用户画像和初始积分账户
func (r *userRepo) CreateWithInitialResources(ctx context.Context, user *biz.User, profile *biz.UserProfile, initialCredits int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	createUserSQL := `
		INSERT INTO users (id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	if _, err := tx.Exec(ctx, createUserSQL,
		user.ID,
		user.Username,
		biz.NormalizeEmail(user.Email),
		user.PasswordHash,
		nullString(user.Phone),
		nullString(user.AvatarURL),
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
		user.IsDeleted,
		nullTime(user.DeletionScheduledAt),
	); err != nil {
		return err
	}

	createProfileSQL := `
		INSERT INTO user_profiles (user_id, full_name, gender, birth_date, location, interests, occupation, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	locationJSON, err := json.Marshal(profile.Location)
	if err != nil {
		return err
	}
	var birthDate sql.NullTime
	if profile.BirthDate != nil {
		birthDate = sql.NullTime{Time: *profile.BirthDate.ToTime(), Valid: true}
	}

	if _, err := tx.Exec(ctx, createProfileSQL,
		profile.UserID,
		nullString(profile.FullName),
		nullString(profile.Gender),
		birthDate,
		locationJSON,
		pq.Array(profile.Interests),
		nullString(profile.Occupation),
		nullString(profile.Bio),
		profile.CreatedAt,
		profile.UpdatedAt,
	); err != nil {
		return err
	}

	createCreditSQL := `
		INSERT INTO credit_accounts (user_id, balance, total_charged, total_consumed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.Exec(ctx, createCreditSQL,
		user.ID,
		initialCredits,
		initialCredits,
		0,
		user.CreatedAt,
		user.UpdatedAt,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetByID 根据ID获取用户
func (r *userRepo) GetByID(ctx context.Context, id int64) (*biz.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at
		FROM users
		WHERE id = $1 AND NOT is_deleted
	`
	return r.scanUser(r.db.QueryRow(ctx, query, id))
}

// GetByEmail 根据邮箱获取用户
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at
		FROM users
		WHERE lower(email) = lower($1) AND NOT is_deleted
	`
	return r.scanUser(r.db.QueryRow(ctx, query, biz.NormalizeEmail(email)))
}

// GetByUsername 根据用户名获取用户
func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at
		FROM users
		WHERE username = $1 AND NOT is_deleted
	`
	return r.scanUser(r.db.QueryRow(ctx, query, username))
}

// GetByPhone 根据手机号获取用户
func (r *userRepo) GetByPhone(ctx context.Context, phone string) (*biz.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at
		FROM users
		WHERE phone = $1 AND NOT is_deleted
	`
	return r.scanUser(r.db.QueryRow(ctx, query, phone))
}

// scanUser 扫描用户行
func (r *userRepo) scanUser(row pgx.Row) (*biz.User, error) {
	var user biz.User
	var phone, avatarURL sql.NullString
	var lastLoginAt, deletionScheduledAt sql.NullTime

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&phone,
		&avatarURL,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.IsDeleted,
		&deletionScheduledAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}

	if phone.Valid {
		user.Phone = phone.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if deletionScheduledAt.Valid {
		user.DeletionScheduledAt = &deletionScheduledAt.Time
	}

	return &user, nil
}

// Update 更新用户
func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	query := `
		UPDATE users SET
			username = $2,
			email = $3,
			password_hash = $4,
			phone = $5,
			avatar_url = $6,
			email_verified = $7,
			updated_at = $8,
			last_login_at = $9,
			is_deleted = $10,
			deletion_scheduled_at = $11
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.Username,
		biz.NormalizeEmail(user.Email),
		user.PasswordHash,
		nullString(user.Phone),
		nullString(user.AvatarURL),
		user.EmailVerified,
		user.UpdatedAt,
		nullTime(user.LastLoginAt),
		user.IsDeleted,
		nullTime(user.DeletionScheduledAt),
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrUserNotFound
	}

	return nil
}

// Delete 删除用户 (软删除)
func (r *userRepo) Delete(ctx context.Context, id int64) error {
	query := `UPDATE users SET is_deleted = true, updated_at = $2 WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrUserNotFound
	}

	return nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1) AND NOT is_deleted)`
	var exists bool
	err := r.db.QueryRow(ctx, query, biz.NormalizeEmail(email)).Scan(&exists)
	return exists, err
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND NOT is_deleted)`
	var exists bool
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	return exists, err
}

// ExistsByPhone 检查手机号是否存在
func (r *userRepo) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	if phone == "" {
		return false, nil
	}
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND NOT is_deleted)`
	var exists bool
	err := r.db.QueryRow(ctx, query, phone).Scan(&exists)
	return exists, err
}

// ListScheduledForDeletion 列出计划删除的用户
func (r *userRepo) ListScheduledForDeletion(ctx context.Context, before time.Time) ([]*biz.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone, avatar_url,
			email_verified, created_at, updated_at, last_login_at, is_deleted, deletion_scheduled_at
		FROM users
		WHERE deletion_scheduled_at IS NOT NULL
			AND deletion_scheduled_at < $1
			AND NOT is_deleted
	`
	rows, err := r.db.Query(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*biz.User
	for rows.Next() {
		var user biz.User
		var phone, avatarURL sql.NullString
		var lastLoginAt, deletionScheduledAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&phone,
			&avatarURL,
			&user.EmailVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
			&user.IsDeleted,
			&deletionScheduledAt,
		)
		if err != nil {
			return nil, err
		}

		if phone.Valid {
			user.Phone = phone.String
		}
		if avatarURL.Valid {
			user.AvatarURL = avatarURL.String
		}
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}
		if deletionScheduledAt.Valid {
			user.DeletionScheduledAt = &deletionScheduledAt.Time
		}

		users = append(users, &user)
	}

	return users, rows.Err()
}

// 辅助函数

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
