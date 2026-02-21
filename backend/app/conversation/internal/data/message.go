// Package data 消息数据访问层
package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/app/conversation/internal/biz"
)

// messageRepo 消息仓储实现
type messageRepo struct {
	db  *pgxpool.Pool
	log *log.Helper
}

// NewMessageRepo 创建消息仓储
func NewMessageRepo(data *Data, logger log.Logger) biz.MessageRepo {
	return &messageRepo{
		db:  data.db,
		log: log.NewHelper(logger),
	}
}

// Create 创建消息
func (r *messageRepo) Create(ctx context.Context, message *biz.Message) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insertMessageQuery := `
		INSERT INTO messages (id, conversation_id, client_message_id, role, content, token_count, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	metadata, err := serializeMetadata(message.Metadata)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, insertMessageQuery,
		message.ID,
		message.ConversationID,
		fallbackClientMessageID(message.ID),
		string(message.Role),
		message.Content,
		message.TokenCount,
		metadata,
		message.CreatedAt,
	)
	if err != nil {
		return err
	}

	insertKeyQuery := `
		INSERT INTO message_partition_keys (message_id, created_at, conversation_id, created_on)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id) DO UPDATE SET created_at = EXCLUDED.created_at, conversation_id = EXCLUDED.conversation_id
	`
	if _, err := tx.Exec(ctx, insertKeyQuery, message.ID, message.CreatedAt, message.ConversationID, time.Now()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CreateBatch 批量创建消息
func (r *messageRepo) CreateBatch(ctx context.Context, messages []*biz.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// 使用事务确保原子性
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO messages (id, conversation_id, client_message_id, role, content, token_count, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, message := range messages {
		metadata, err := serializeMetadata(message.Metadata)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, query,
			message.ID,
			message.ConversationID,
			fallbackClientMessageID(message.ID),
			string(message.Role),
			message.Content,
			message.TokenCount,
			metadata,
			message.CreatedAt,
		)
		if err != nil {
			return err
		}

		insertKeyQuery := `
			INSERT INTO message_partition_keys (message_id, created_at, conversation_id, created_on)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (message_id) DO UPDATE SET created_at = EXCLUDED.created_at, conversation_id = EXCLUDED.conversation_id
		`
		if _, err := tx.Exec(ctx, insertKeyQuery, message.ID, message.CreatedAt, message.ConversationID, time.Now()); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetByID 根据ID获取消息
func (r *messageRepo) GetByID(ctx context.Context, id int64) (*biz.Message, error) {
	createdAt, err := r.lookupMessageCreatedAt(ctx, id)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, conversation_id, role, content, token_count, metadata, created_at
		FROM messages
		WHERE id = $1 AND created_at = $2
	`
	return r.scanMessage(r.db.QueryRow(ctx, query, id, createdAt))
}

// ListByConversationID 列出会话的消息列表（分页，按时间倒序）
func (r *messageRepo) ListByConversationID(ctx context.Context, conversationID, beforeMessageID int64, limit int32) ([]*biz.Message, error) {
	var query string
	var args []interface{}

	if beforeMessageID > 0 {
		// 游标分页：获取指定消息之前的消息
		query = `
			SELECT id, conversation_id, role, content, token_count, metadata, created_at
			FROM messages
			WHERE conversation_id = $1 AND id < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{conversationID, beforeMessageID, limit}
	} else {
		// 首次查询：获取最新的消息
		query = `
			SELECT id, conversation_id, role, content, token_count, metadata, created_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{conversationID, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*biz.Message
	for rows.Next() {
		message, err := r.scanMessageRow(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

// CountByConversationID 统计会话的消息数量
func (r *messageRepo) CountByConversationID(ctx context.Context, conversationID int64) (int32, error) {
	query := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`
	var count int32
	err := r.db.QueryRow(ctx, query, conversationID).Scan(&count)
	return count, err
}

// scanMessage 扫描消息行
func (r *messageRepo) scanMessage(row pgx.Row) (*biz.Message, error) {
	var message biz.Message
	var role string
	var metadataJSON []byte

	err := row.Scan(
		&message.ID,
		&message.ConversationID,
		&role,
		&message.Content,
		&message.TokenCount,
		&metadataJSON,
		&message.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrMessageNotFound
		}
		return nil, err
	}

	message.Role = biz.MessageRole(role)

	// 解析元数据
	if len(metadataJSON) > 0 {
		var metadata biz.MessageMetadata
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			r.log.Warnf("failed to unmarshal metadata: %v", err)
		} else {
			message.Metadata = &metadata
		}
	}

	return &message, nil
}

// scanMessageRow 扫描消息行（用于rows）
func (r *messageRepo) scanMessageRow(row pgx.Row) (*biz.Message, error) {
	var message biz.Message
	var role string
	var metadataJSON []byte

	err := row.Scan(
		&message.ID,
		&message.ConversationID,
		&role,
		&message.Content,
		&message.TokenCount,
		&metadataJSON,
		&message.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	message.Role = biz.MessageRole(role)

	// 解析元数据
	if len(metadataJSON) > 0 {
		var metadata biz.MessageMetadata
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			r.log.Warnf("failed to unmarshal metadata: %v", err)
		} else {
			message.Metadata = &metadata
		}
	}

	return &message, nil
}

// serializeMetadata 序列化元数据
func serializeMetadata(metadata *biz.MessageMetadata) (interface{}, error) {
	if metadata == nil {
		return sql.NullString{}, nil
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func fallbackClientMessageID(messageID int64) string {
	return fmt.Sprintf("legacy-%d", messageID)
}

func (r *messageRepo) lookupMessageCreatedAt(ctx context.Context, messageID int64) (time.Time, error) {
	query := `SELECT created_at FROM message_partition_keys WHERE message_id = $1`
	var createdAt time.Time
	if err := r.db.QueryRow(ctx, query, messageID).Scan(&createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, biz.ErrMessageNotFound
		}
		return time.Time{}, err
	}
	return createdAt, nil
}
