// Package data provides data access utilities including cascade operations
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNoRowsAffected = errors.New("cascade: no rows affected")
	ErrNilTransaction = errors.New("cascade: transaction is nil")
)

// CascadeResult contains the result of cascade delete operations
type CascadeResult struct {
	// Table is the table name
	Table string
	// RowsAffected is the number of rows deleted
	RowsAffected int64
}

// CascadeStats aggregates cascade operation statistics
type CascadeStats struct {
	// Results contains per-table deletion results
	Results []CascadeResult
	// TotalRowsAffected is the sum of all rows deleted
	TotalRowsAffected int64
}

// Add adds a result to the stats
func (s *CascadeStats) Add(table string, rows int64) {
	s.Results = append(s.Results, CascadeResult{
		Table:        table,
		RowsAffected: rows,
	})
	s.TotalRowsAffected += rows
}

// Executor represents a database executor (either *sql.DB or *sql.Tx)
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// TxProvider provides transaction management
type TxProvider interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// CascadeManager handles cascade delete operations
type CascadeManager struct {
	db TxProvider
}

// NewCascadeManager creates a new cascade manager
func NewCascadeManager(db TxProvider) *CascadeManager {
	return &CascadeManager{db: db}
}

// DeleteUserCascade deletes a user and all related data in a transaction
// Order of deletion (child-to-parent):
// 1. user_profiles
// 2. sessions
// 3. messages (via conversations)
// 4. conversations
// 5. memories
// 6. relationships
// 7. important_events
// 8. user_portraits
// 9. credit_transactions
// 10. credit_accounts
// 11. orders
// 12. characters (user-created)
// 13. users
func (m *CascadeManager) DeleteUserCascade(ctx context.Context, userID int64) (*CascadeStats, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stats, err := DeleteUserWithTx(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}

// DeleteUserWithTx deletes a user within an existing transaction
func DeleteUserWithTx(ctx context.Context, tx *sql.Tx, userID int64) (*CascadeStats, error) {
	if tx == nil {
		return nil, ErrNilTransaction
	}

	stats := &CascadeStats{}

	// 1. Delete user profiles
	rows, err := execDelete(ctx, tx, "DELETE FROM user_profiles WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete user_profiles: %w", err)
	}
	stats.Add("user_profiles", rows)

	// 2. Delete sessions
	rows, err = execDelete(ctx, tx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete sessions: %w", err)
	}
	stats.Add("sessions", rows)

	// 3. Get conversation IDs first for message deletion
	convIDs, err := getConversationIDs(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("get conversation_ids: %w", err)
	}

	// 4. Delete messages (via conversation IDs)
	if len(convIDs) > 0 {
		rows, err = deleteMessagesByConversations(ctx, tx, convIDs)
		if err != nil {
			return nil, fmt.Errorf("delete messages: %w", err)
		}
		stats.Add("messages", rows)
	}

	// 5. Delete conversations
	rows, err = execDelete(ctx, tx, "DELETE FROM conversations WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete conversations: %w", err)
	}
	stats.Add("conversations", rows)

	// 6. Delete memories
	rows, err = execDelete(ctx, tx, "DELETE FROM memories WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete memories: %w", err)
	}
	stats.Add("memories", rows)

	// 7. Delete relationships
	rows, err = execDelete(ctx, tx, "DELETE FROM relationships WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete relationships: %w", err)
	}
	stats.Add("relationships", rows)

	// 8. Delete important events
	rows, err = execDelete(ctx, tx, "DELETE FROM important_events WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete important_events: %w", err)
	}
	stats.Add("important_events", rows)

	// 9. Delete user portraits
	rows, err = execDelete(ctx, tx, "DELETE FROM user_portraits WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete user_portraits: %w", err)
	}
	stats.Add("user_portraits", rows)

	// 10. Delete credit transactions
	rows, err = execDelete(ctx, tx, "DELETE FROM credit_transactions WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete credit_transactions: %w", err)
	}
	stats.Add("credit_transactions", rows)

	// 11. Delete credit accounts
	rows, err = execDelete(ctx, tx, "DELETE FROM credit_accounts WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete credit_accounts: %w", err)
	}
	stats.Add("credit_accounts", rows)

	// 12. Delete orders
	rows, err = execDelete(ctx, tx, "DELETE FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete orders: %w", err)
	}
	stats.Add("orders", rows)

	// 13. Delete user-created characters (this will also cascade to their data)
	charIDs, err := getCharacterIDs(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("get character_ids: %w", err)
	}

	for _, charID := range charIDs {
		charStats, err := DeleteCharacterWithTx(ctx, tx, charID)
		if err != nil {
			return nil, fmt.Errorf("delete character %d: %w", charID, err)
		}
		// Merge stats
		for _, r := range charStats.Results {
			stats.Add(r.Table+" (char "+fmt.Sprint(charID)+")", r.RowsAffected)
		}
	}

	// 14. Finally delete the user
	rows, err = execDelete(ctx, tx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("delete users: %w", err)
	}
	if rows == 0 {
		return nil, ErrNoRowsAffected
	}
	stats.Add("users", rows)

	return stats, nil
}

// DeleteCharacterCascade deletes a character and all related data
// Order of deletion:
// 1. messages (via conversations with this character)
// 2. conversations (with this character)
// 3. memories (for this character)
// 4. relationships (for this character)
// 5. important_events (for this character)
// 6. character
func (m *CascadeManager) DeleteCharacterCascade(ctx context.Context, characterID int64) (*CascadeStats, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stats, err := DeleteCharacterWithTx(ctx, tx, characterID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}

// DeleteCharacterWithTx deletes a character within an existing transaction
func DeleteCharacterWithTx(ctx context.Context, tx *sql.Tx, characterID int64) (*CascadeStats, error) {
	if tx == nil {
		return nil, ErrNilTransaction
	}

	stats := &CascadeStats{}

	// 1. Get conversation IDs for this character
	convIDs, err := getCharacterConversationIDs(ctx, tx, characterID)
	if err != nil {
		return nil, fmt.Errorf("get conversation_ids: %w", err)
	}

	// 2. Delete messages
	if len(convIDs) > 0 {
		rows, err := deleteMessagesByConversations(ctx, tx, convIDs)
		if err != nil {
			return nil, fmt.Errorf("delete messages: %w", err)
		}
		stats.Add("messages", rows)
	}

	// 3. Delete conversations
	rows, err := execDelete(ctx, tx, "DELETE FROM conversations WHERE character_id = $1", characterID)
	if err != nil {
		return nil, fmt.Errorf("delete conversations: %w", err)
	}
	stats.Add("conversations", rows)

	// 4. Delete memories
	rows, err = execDelete(ctx, tx, "DELETE FROM memories WHERE character_id = $1", characterID)
	if err != nil {
		return nil, fmt.Errorf("delete memories: %w", err)
	}
	stats.Add("memories", rows)

	// 5. Delete relationships
	rows, err = execDelete(ctx, tx, "DELETE FROM relationships WHERE character_id = $1", characterID)
	if err != nil {
		return nil, fmt.Errorf("delete relationships: %w", err)
	}
	stats.Add("relationships", rows)

	// 6. Delete important events
	rows, err = execDelete(ctx, tx, "DELETE FROM important_events WHERE character_id = $1", characterID)
	if err != nil {
		return nil, fmt.Errorf("delete important_events: %w", err)
	}
	stats.Add("important_events", rows)

	// 7. Delete the character
	rows, err = execDelete(ctx, tx, "DELETE FROM characters WHERE id = $1", characterID)
	if err != nil {
		return nil, fmt.Errorf("delete characters: %w", err)
	}
	if rows == 0 {
		return nil, ErrNoRowsAffected
	}
	stats.Add("characters", rows)

	return stats, nil
}

// DeleteConversationCascade deletes a conversation and all its messages
func (m *CascadeManager) DeleteConversationCascade(ctx context.Context, conversationID int64) (*CascadeStats, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stats, err := DeleteConversationWithTx(ctx, tx, conversationID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}

// DeleteConversationWithTx deletes a conversation within an existing transaction
func DeleteConversationWithTx(ctx context.Context, tx *sql.Tx, conversationID int64) (*CascadeStats, error) {
	if tx == nil {
		return nil, ErrNilTransaction
	}

	stats := &CascadeStats{}

	// 1. Delete messages
	rows, err := execDelete(ctx, tx, "DELETE FROM messages WHERE conversation_id = $1", conversationID)
	if err != nil {
		return nil, fmt.Errorf("delete messages: %w", err)
	}
	stats.Add("messages", rows)

	// 2. Delete conversation
	rows, err = execDelete(ctx, tx, "DELETE FROM conversations WHERE id = $1", conversationID)
	if err != nil {
		return nil, fmt.Errorf("delete conversations: %w", err)
	}
	if rows == 0 {
		return nil, ErrNoRowsAffected
	}
	stats.Add("conversations", rows)

	return stats, nil
}

// SoftDeleteUser marks a user as deleted instead of hard deleting
// This sets is_deleted = true and deletion_requested_at = now()
func (m *CascadeManager) SoftDeleteUser(ctx context.Context, userID int64) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET is_deleted = true,
		    deletion_requested_at = NOW()
		WHERE id = $1 AND is_deleted = false
	`, userID)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNoRowsAffected
	}

	return tx.Commit()
}

// RecoverUser recovers a soft-deleted user
// This only works within the 30-day cooldown period
func (m *CascadeManager) RecoverUser(ctx context.Context, userID int64) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE users
		SET is_deleted = false,
		    deletion_requested_at = NULL
		WHERE id = $1
		  AND is_deleted = true
		  AND deletion_requested_at > NOW() - INTERVAL '30 days'
	`, userID)
	if err != nil {
		return fmt.Errorf("recover user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNoRowsAffected
	}

	return tx.Commit()
}

// PurgeExpiredUsers permanently deletes users whose deletion cooldown has passed
// Returns the number of users purged
func (m *CascadeManager) PurgeExpiredUsers(ctx context.Context) (int, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get users to purge
	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM users
		WHERE is_deleted = true
		  AND deletion_requested_at <= NOW() - INTERVAL '30 days'
	`)
	if err != nil {
		return 0, fmt.Errorf("query expired users: %w", err)
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("scan user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate users: %w", err)
	}

	// Delete each user
	for _, userID := range userIDs {
		_, err := DeleteUserWithTx(ctx, tx, userID)
		if err != nil {
			return 0, fmt.Errorf("delete user %d: %w", userID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return len(userIDs), nil
}

// SoftDeleteCharacter marks a character as deleted
func (m *CascadeManager) SoftDeleteCharacter(ctx context.Context, characterID int64) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE characters
		SET is_deleted = true
		WHERE id = $1 AND is_deleted = false
	`, characterID)
	if err != nil {
		return fmt.Errorf("soft delete character: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNoRowsAffected
	}

	return tx.Commit()
}

// Helper functions

func execDelete(ctx context.Context, exec Executor, query string, args ...interface{}) (int64, error) {
	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func getConversationIDs(ctx context.Context, exec Executor, userID int64) ([]int64, error) {
	rows, err := exec.QueryContext(ctx, "SELECT id FROM conversations WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func getCharacterIDs(ctx context.Context, exec Executor, userID int64) ([]int64, error) {
	rows, err := exec.QueryContext(ctx, "SELECT id FROM characters WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func getCharacterConversationIDs(ctx context.Context, exec Executor, characterID int64) ([]int64, error) {
	rows, err := exec.QueryContext(ctx, "SELECT id FROM conversations WHERE character_id = $1", characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func deleteMessagesByConversations(ctx context.Context, tx *sql.Tx, conversationIDs []int64) (int64, error) {
	if len(conversationIDs) == 0 {
		return 0, nil
	}

	// Build IN clause with positional parameters
	// PostgreSQL style: $1, $2, $3...
	query := "DELETE FROM messages WHERE conversation_id IN ("
	args := make([]interface{}, len(conversationIDs))
	for i, id := range conversationIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query += ")"

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// BatchDeleteConversations deletes multiple conversations efficiently
func (m *CascadeManager) BatchDeleteConversations(ctx context.Context, conversationIDs []int64) (*CascadeStats, error) {
	if len(conversationIDs) == 0 {
		return &CascadeStats{}, nil
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stats := &CascadeStats{}

	// Delete messages first
	rows, err := deleteMessagesByConversations(ctx, tx, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("delete messages: %w", err)
	}
	stats.Add("messages", rows)

	// Delete conversations
	query := "DELETE FROM conversations WHERE id IN ("
	args := make([]interface{}, len(conversationIDs))
	for i, id := range conversationIDs {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query += ")"

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("delete conversations: %w", err)
	}
	rows, _ = result.RowsAffected()
	stats.Add("conversations", rows)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return stats, nil
}

// ArchiveConversation moves a conversation to archived state
func (m *CascadeManager) ArchiveConversation(ctx context.Context, conversationID int64) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		UPDATE conversations
		SET status = 'archived',
		    updated_at = NOW()
		WHERE id = $1 AND status = 'active'
	`, conversationID)
	if err != nil {
		return fmt.Errorf("archive conversation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNoRowsAffected
	}

	return tx.Commit()
}
