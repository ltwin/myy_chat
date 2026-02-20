// Package data 用户数据访问层
package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/app/user/internal/biz"
)

// sessionRepo 会话仓储实现
type sessionRepo struct {
	db  *pgxpool.Pool
	log *log.Helper
}

// NewSessionRepo 创建会话仓储
func NewSessionRepo(db *pgxpool.Pool, logger log.Logger) biz.SessionRepo {
	return &sessionRepo{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create 创建会话
func (r *sessionRepo) Create(ctx context.Context, session *biz.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		nullString(session.UserAgent),
		nullString(session.IPAddress),
		session.ExpiresAt,
		session.CreatedAt,
		nullTime(session.RevokedAt),
	)
	return err
}

// GetByID 根据ID获取会话
func (r *sessionRepo) GetByID(ctx context.Context, id int64) (*biz.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM sessions
		WHERE id = $1
	`
	return r.scanSession(r.db.QueryRow(ctx, query, id))
}

// GetByRefreshTokenHash 根据刷新令牌哈希获取会话
func (r *sessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*biz.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM sessions
		WHERE refresh_token_hash = $1 AND revoked_at IS NULL
	`
	return r.scanSession(r.db.QueryRow(ctx, query, hash))
}

// scanSession 扫描会话行
func (r *sessionRepo) scanSession(row pgx.Row) (*biz.Session, error) {
	var session biz.Session
	var userAgent, ipAddress sql.NullString
	var revokedAt sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&userAgent,
		&ipAddress,
		&session.ExpiresAt,
		&session.CreatedAt,
		&revokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrSessionNotFound
		}
		return nil, err
	}

	if userAgent.Valid {
		session.UserAgent = userAgent.String
	}
	if ipAddress.Valid {
		session.IPAddress = ipAddress.String
	}
	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return &session, nil
}

// Update 更新会话
func (r *sessionRepo) Update(ctx context.Context, session *biz.Session) error {
	query := `
		UPDATE sessions SET
			refresh_token_hash = $2,
			expires_at = $3,
			revoked_at = $4
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.RefreshTokenHash,
		session.ExpiresAt,
		nullTime(session.RevokedAt),
	)
	return err
}

// Revoke 撤销会话
func (r *sessionRepo) Revoke(ctx context.Context, id int64) error {
	query := `UPDATE sessions SET revoked_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, time.Now())
	return err
}

// RevokeAllByUserID 撤销用户所有会话
func (r *sessionRepo) RevokeAllByUserID(ctx context.Context, userID int64) error {
	query := `UPDATE sessions SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, query, userID, time.Now())
	return err
}

// DeleteExpired 删除过期会话
func (r *sessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < $1`
	result, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// ListByUserID 列出用户所有活跃会话
func (r *sessionRepo) ListByUserID(ctx context.Context, userID int64) ([]*biz.Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, created_at, revoked_at
		FROM sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*biz.Session
	for rows.Next() {
		var session biz.Session
		var userAgent, ipAddress sql.NullString
		var revokedAt sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshTokenHash,
			&userAgent,
			&ipAddress,
			&session.ExpiresAt,
			&session.CreatedAt,
			&revokedAt,
		)
		if err != nil {
			return nil, err
		}

		if userAgent.Valid {
			session.UserAgent = userAgent.String
		}
		if ipAddress.Valid {
			session.IPAddress = ipAddress.String
		}
		if revokedAt.Valid {
			session.RevokedAt = &revokedAt.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}
