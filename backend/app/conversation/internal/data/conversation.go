// Package data 会话数据访问层
package data

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/app/conversation/internal/biz"
)

// conversationRepo 会话仓储实现
type conversationRepo struct {
	db  *pgxpool.Pool
	log *log.Helper
}

// NewConversationRepo 创建会话仓储
func NewConversationRepo(data *Data, logger log.Logger) biz.ConversationRepo {
	return &conversationRepo{
		db:  data.db,
		log: log.NewHelper(logger),
	}
}

// Create 创建会话
func (r *conversationRepo) Create(ctx context.Context, conversation *biz.Conversation) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insertConversationQuery := `
		INSERT INTO conversations (id, user_id, character_id, title, message_count, token_count,
			started_at, last_message_at, is_archived)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if _, err := tx.Exec(ctx, insertConversationQuery,
		conversation.ID,
		conversation.UserID,
		conversation.CharacterID,
		conversation.Title,
		conversation.MessageCount,
		conversation.TokenCount,
		conversation.StartedAt,
		conversation.LastMessageAt,
		conversation.IsArchived,
	); err != nil {
		return err
	}

	insertPartitionKeyQuery := `
		INSERT INTO conversation_partition_keys (conversation_id, started_at, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (conversation_id) DO UPDATE SET started_at = EXCLUDED.started_at
	`
	if _, err := tx.Exec(ctx, insertPartitionKeyQuery, conversation.ID, conversation.StartedAt, time.Now()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetByID 根据ID获取会话
func (r *conversationRepo) GetByID(ctx context.Context, id int64) (*biz.Conversation, error) {
	startedAt, err := r.lookupConversationStartedAt(ctx, id)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, user_id, character_id, title, message_count, token_count,
			started_at, last_message_at, is_archived
		FROM conversations
		WHERE id = $1 AND started_at = $2
	`
	return r.scanConversation(r.db.QueryRow(ctx, query, id, startedAt))
}

// GetByIDAndUserID 根据ID和用户ID获取会话（权限检查）
func (r *conversationRepo) GetByIDAndUserID(ctx context.Context, id, userID int64) (*biz.Conversation, error) {
	startedAt, err := r.lookupConversationStartedAt(ctx, id)
	if err != nil {
		if errors.Is(err, biz.ErrConversationNotFound) {
			return nil, biz.ErrConversationAccessDenied
		}
		return nil, err
	}

	query := `
		SELECT id, user_id, character_id, title, message_count, token_count,
			started_at, last_message_at, is_archived
		FROM conversations
		WHERE id = $1 AND user_id = $2 AND started_at = $3
	`
	conversation, err := r.scanConversation(r.db.QueryRow(ctx, query, id, userID, startedAt))
	if err != nil {
		if errors.Is(err, biz.ErrConversationNotFound) {
			return nil, biz.ErrConversationAccessDenied
		}
		return nil, err
	}
	return conversation, nil
}

// List 列出用户的会话列表（分页）
func (r *conversationRepo) List(ctx context.Context, userID int64, characterID int64, archived bool, page, pageSize int32) ([]*biz.Conversation, int32, error) {
	offset := (page - 1) * pageSize

	// 构建查询条件
	baseWhere := "user_id = $1 AND is_archived = $2"
	args := []interface{}{userID, archived}
	argIndex := 3

	if characterID > 0 {
		baseWhere += " AND character_id = $3"
		args = append(args, characterID)
		argIndex++
	}

	// 查询总数
	countQuery := "SELECT COUNT(*) FROM conversations WHERE " + baseWhere
	var total int32
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 查询列表
	listQuery := `
		SELECT id, user_id, character_id, title, message_count, token_count,
			started_at, last_message_at, is_archived
		FROM conversations
		WHERE ` + baseWhere + `
		ORDER BY last_message_at DESC
		LIMIT $` + sqlPlaceholder(argIndex) + ` OFFSET $` + sqlPlaceholder(argIndex+1)

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var conversations []*biz.Conversation
	for rows.Next() {
		conversation, err := r.scanConversationRow(rows)
		if err != nil {
			return nil, 0, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, total, rows.Err()
}

// Update 更新会话
func (r *conversationRepo) Update(ctx context.Context, conversation *biz.Conversation) error {
	startedAt, err := r.lookupConversationStartedAt(ctx, conversation.ID)
	if err != nil {
		return err
	}

	query := `
		UPDATE conversations SET
			title = $2,
			message_count = $3,
			token_count = $4,
			last_message_at = $5,
			is_archived = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND started_at = $7
	`
	result, err := r.db.Exec(ctx, query,
		conversation.ID,
		conversation.Title,
		conversation.MessageCount,
		conversation.TokenCount,
		conversation.LastMessageAt,
		conversation.IsArchived,
		startedAt,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrConversationNotFound
	}

	return nil
}

// IncrementCounts 原子性增加消息计数和Token计数
func (r *conversationRepo) IncrementCounts(ctx context.Context, conversationID int64, tokenCount int32) error {
	startedAt, err := r.lookupConversationStartedAt(ctx, conversationID)
	if err != nil {
		return err
	}

	query := `
		UPDATE conversations SET
			message_count = message_count + 2,
			token_count = token_count + $2,
			last_message_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND started_at = $3
	`
	result, err := r.db.Exec(ctx, query, conversationID, tokenCount, startedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrConversationNotFound
	}

	return nil
}

// scanConversation 扫描会话行
func (r *conversationRepo) scanConversation(row pgx.Row) (*biz.Conversation, error) {
	var conversation biz.Conversation
	var title sql.NullString

	err := row.Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.CharacterID,
		&title,
		&conversation.MessageCount,
		&conversation.TokenCount,
		&conversation.StartedAt,
		&conversation.LastMessageAt,
		&conversation.IsArchived,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrConversationNotFound
		}
		return nil, err
	}

	if title.Valid {
		conversation.Title = title.String
	}

	return &conversation, nil
}

// scanConversationRow 扫描会话行（用于rows）
func (r *conversationRepo) scanConversationRow(row pgx.Row) (*biz.Conversation, error) {
	var conversation biz.Conversation
	var title sql.NullString

	err := row.Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.CharacterID,
		&title,
		&conversation.MessageCount,
		&conversation.TokenCount,
		&conversation.StartedAt,
		&conversation.LastMessageAt,
		&conversation.IsArchived,
	)

	if err != nil {
		return nil, err
	}

	if title.Valid {
		conversation.Title = title.String
	}

	return &conversation, nil
}

// sqlPlaceholder 生成SQL占位符数字部分
func sqlPlaceholder(n int) string {
	return strconv.Itoa(n)
}

func (r *conversationRepo) lookupConversationStartedAt(ctx context.Context, conversationID int64) (time.Time, error) {
	query := `SELECT started_at FROM conversation_partition_keys WHERE conversation_id = $1`
	var startedAt time.Time
	if err := r.db.QueryRow(ctx, query, conversationID).Scan(&startedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, biz.ErrConversationNotFound
		}
		return time.Time{}, err
	}
	return startedAt, nil
}
