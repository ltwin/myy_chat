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

// userProfileRepo 用户画像仓储实现
type userProfileRepo struct {
	db  *pgxpool.Pool
	log *log.Helper
}

// NewUserProfileRepo 创建用户画像仓储
func NewUserProfileRepo(db *pgxpool.Pool, logger log.Logger) biz.UserProfileRepo {
	return &userProfileRepo{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create 创建用户画像
func (r *userProfileRepo) Create(ctx context.Context, profile *biz.UserProfile) error {
	locationJSON, err := json.Marshal(profile.Location)
	if err != nil {
		return err
	}

	var birthDate sql.NullTime
	if profile.BirthDate != nil {
		birthDate = sql.NullTime{Time: *profile.BirthDate.ToTime(), Valid: true}
	}

	query := `
		INSERT INTO user_profiles (user_id, full_name, gender, birth_date, location, interests, occupation, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = r.db.Exec(ctx, query,
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
	)
	return err
}

// GetByUserID 根据用户ID获取画像
func (r *userProfileRepo) GetByUserID(ctx context.Context, userID int64) (*biz.UserProfile, error) {
	query := `
		SELECT user_id, full_name, gender, birth_date, location, interests, occupation, bio, created_at, updated_at
		FROM user_profiles
		WHERE user_id = $1
	`

	var profile biz.UserProfile
	var fullName, gender, occupation, bio sql.NullString
	var birthDate sql.NullTime
	var locationJSON []byte
	var interests []string

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&fullName,
		&gender,
		&birthDate,
		&locationJSON,
		&interests,
		&occupation,
		&bio,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrUserNotFound
		}
		return nil, err
	}

	if fullName.Valid {
		profile.FullName = fullName.String
	}
	if gender.Valid {
		profile.Gender = gender.String
	}
	if birthDate.Valid {
		t := birthDate.Time
		profile.BirthDate = &biz.Date{
			Year:  t.Year(),
			Month: int(t.Month()),
			Day:   t.Day(),
		}
	}
	if len(locationJSON) > 0 && string(locationJSON) != "null" {
		var loc biz.Location
		if err := json.Unmarshal(locationJSON, &loc); err == nil {
			profile.Location = &loc
		}
	}
	profile.Interests = interests
	if occupation.Valid {
		profile.Occupation = occupation.String
	}
	if bio.Valid {
		profile.Bio = bio.String
	}

	return &profile, nil
}

// Update 更新用户画像
func (r *userProfileRepo) Update(ctx context.Context, profile *biz.UserProfile) error {
	locationJSON, err := json.Marshal(profile.Location)
	if err != nil {
		return err
	}

	var birthDate sql.NullTime
	if profile.BirthDate != nil {
		birthDate = sql.NullTime{Time: *profile.BirthDate.ToTime(), Valid: true}
	}

	query := `
		INSERT INTO user_profiles (user_id, full_name, gender, birth_date, location, interests, occupation, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			gender = EXCLUDED.gender,
			birth_date = EXCLUDED.birth_date,
			location = EXCLUDED.location,
			interests = EXCLUDED.interests,
			occupation = EXCLUDED.occupation,
			bio = EXCLUDED.bio,
			updated_at = EXCLUDED.updated_at
	`
	_, err = r.db.Exec(ctx, query,
		profile.UserID,
		nullString(profile.FullName),
		nullString(profile.Gender),
		birthDate,
		locationJSON,
		pq.Array(profile.Interests),
		nullString(profile.Occupation),
		nullString(profile.Bio),
		time.Now(),
		profile.UpdatedAt,
	)
	return err
}

// Delete 删除用户画像
func (r *userProfileRepo) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_profiles WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
