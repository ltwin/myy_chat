// Package data 角色数据访问层
package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/golang/app/character/internal/biz"
)

// characterRepo 角色仓储实现
type characterRepo struct {
	db  *pgxpool.Pool
	log *log.Helper
}

// NewCharacterRepo 创建角色仓储
func NewCharacterRepo(db *pgxpool.Pool, logger log.Logger) biz.CharacterRepo {
	return &characterRepo{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// Create 创建角色
func (r *characterRepo) Create(ctx context.Context, character *biz.Character) error {
	query := `
		INSERT INTO characters (id, user_id, name, avatar_url, description, personality,
			background_story, speaking_style, system_prompt, world_view,
			is_public, is_preset, version, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.Exec(ctx, query,
		character.ID,
		nullInt64(character.UserID),
		character.Name,
		nullString(character.AvatarURL),
		nullString(character.Description),
		character.Personality,
		nullString(character.BackgroundStory),
		nullString(character.SpeakingStyle),
		character.SystemPrompt,
		character.WorldView,
		character.IsPublic,
		character.IsPreset,
		character.Version,
		character.CreatedAt,
		character.UpdatedAt,
		character.IsDeleted,
	)
	return err
}

// GetByID 根据ID获取角色
func (r *characterRepo) GetByID(ctx context.Context, id int64) (*biz.Character, error) {
	query := `
		SELECT id, user_id, name, avatar_url, description, personality,
			background_story, speaking_style, system_prompt, world_view,
			is_public, is_preset, version, created_at, updated_at, is_deleted
		FROM characters
		WHERE id = $1 AND NOT is_deleted
	`
	return r.scanCharacter(r.db.QueryRow(ctx, query, id))
}

// Update 更新角色
func (r *characterRepo) Update(ctx context.Context, character *biz.Character) error {
	query := `
		UPDATE characters SET
			user_id = $2,
			name = $3,
			avatar_url = $4,
			description = $5,
			personality = $6,
			background_story = $7,
			speaking_style = $8,
			system_prompt = $9,
			world_view = $10,
			is_public = $11,
			is_preset = $12,
			version = $13,
			updated_at = $14,
			is_deleted = $15
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query,
		character.ID,
		nullInt64(character.UserID),
		character.Name,
		nullString(character.AvatarURL),
		nullString(character.Description),
		character.Personality,
		nullString(character.BackgroundStory),
		nullString(character.SpeakingStyle),
		character.SystemPrompt,
		character.WorldView,
		character.IsPublic,
		character.IsPreset,
		character.Version,
		character.UpdatedAt,
		character.IsDeleted,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrCharacterNotFound
	}

	return nil
}

// Delete 删除角色 (软删除)
func (r *characterRepo) Delete(ctx context.Context, id int64) error {
	query := `UPDATE characters SET is_deleted = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrCharacterNotFound
	}

	return nil
}

// List 列出角色 (支持分页和过滤)
func (r *characterRepo) List(ctx context.Context, userID int64, filter string, page, pageSize int) ([]*biz.Character, int, error) {
	// 构建 WHERE 条件
	whereClause, args := r.buildWhereClause(userID, filter)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM characters WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 如果没有结果，直接返回
	if total == 0 {
		return []*biz.Character{}, 0, nil
	}

	// 查询分页数据
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT id, user_id, name, avatar_url, description, personality,
			background_story, speaking_style, system_prompt, world_view,
			is_public, is_preset, version, created_at, updated_at, is_deleted
		FROM characters
		WHERE %s
		ORDER BY is_preset DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var characters []*biz.Character
	for rows.Next() {
		char, err := r.scanCharacterFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		characters = append(characters, char)
	}

	return characters, total, rows.Err()
}

// ExistsByName 检查角色名称是否存在 (针对特定用户)
func (r *characterRepo) ExistsByName(ctx context.Context, userID int64, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM characters WHERE user_id = $1 AND name = $2 AND NOT is_deleted)`
	var exists bool
	err := r.db.QueryRow(ctx, query, userID, name).Scan(&exists)
	return exists, err
}

// buildWhereClause 构建 WHERE 条件
func (r *characterRepo) buildWhereClause(userID int64, filter string) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 基础条件: 未删除
	conditions = append(conditions, "NOT is_deleted")

	switch filter {
	case "preset":
		// 仅预设角色
		conditions = append(conditions, "is_preset = true")

	case "public":
		// 公开角色 + 预设角色
		conditions = append(conditions, "(is_public = true OR is_preset = true)")

	case "mine":
		// 仅用户自己的角色
		if userID > 0 {
			conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)+1))
			args = append(args, userID)
		} else {
			// 未登录用户没有"我的角色"
			conditions = append(conditions, "FALSE")
		}

	case "all":
		fallthrough
	default:
		// 预设角色 + 公开角色 + 用户自己的角色
		if userID > 0 {
			conditions = append(conditions, fmt.Sprintf("(is_preset = true OR is_public = true OR user_id = $%d)", len(args)+1))
			args = append(args, userID)
		} else {
			// 未登录用户只能看到预设和公开角色
			conditions = append(conditions, "(is_preset = true OR is_public = true)")
		}
	}

	whereClause := conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	return whereClause, args
}

// scanCharacter 扫描角色行
func (r *characterRepo) scanCharacter(row pgx.Row) (*biz.Character, error) {
	var char biz.Character
	var userID sql.NullInt64
	var avatarURL, description, backgroundStory, speakingStyle sql.NullString
	var personality, worldView []byte

	err := row.Scan(
		&char.ID,
		&userID,
		&char.Name,
		&avatarURL,
		&description,
		&personality,
		&backgroundStory,
		&speakingStyle,
		&char.SystemPrompt,
		&worldView,
		&char.IsPublic,
		&char.IsPreset,
		&char.Version,
		&char.CreatedAt,
		&char.UpdatedAt,
		&char.IsDeleted,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrCharacterNotFound
		}
		return nil, err
	}

	// 处理可空字段
	if userID.Valid {
		char.UserID = userID.Int64
	}
	if avatarURL.Valid {
		char.AvatarURL = avatarURL.String
	}
	if description.Valid {
		char.Description = description.String
	}
	if backgroundStory.Valid {
		char.BackgroundStory = backgroundStory.String
	}
	if speakingStyle.Valid {
		char.SpeakingStyle = speakingStyle.String
	}

	// 解析 JSONB 字段
	char.Personality = json.RawMessage(personality)
	char.WorldView = json.RawMessage(worldView)

	return &char, nil
}

// scanCharacterFromRows 从 Rows 扫描角色
func (r *characterRepo) scanCharacterFromRows(rows pgx.Rows) (*biz.Character, error) {
	var char biz.Character
	var userID sql.NullInt64
	var avatarURL, description, backgroundStory, speakingStyle sql.NullString
	var personality, worldView []byte

	err := rows.Scan(
		&char.ID,
		&userID,
		&char.Name,
		&avatarURL,
		&description,
		&personality,
		&backgroundStory,
		&speakingStyle,
		&char.SystemPrompt,
		&worldView,
		&char.IsPublic,
		&char.IsPreset,
		&char.Version,
		&char.CreatedAt,
		&char.UpdatedAt,
		&char.IsDeleted,
	)

	if err != nil {
		return nil, err
	}

	// 处理可空字段
	if userID.Valid {
		char.UserID = userID.Int64
	}
	if avatarURL.Valid {
		char.AvatarURL = avatarURL.String
	}
	if description.Valid {
		char.Description = description.String
	}
	if backgroundStory.Valid {
		char.BackgroundStory = backgroundStory.String
	}
	if speakingStyle.Valid {
		char.SpeakingStyle = speakingStyle.String
	}

	// 解析 JSONB 字段
	char.Personality = json.RawMessage(personality)
	char.WorldView = json.RawMessage(worldView)

	return &char, nil
}

// 辅助函数

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}
