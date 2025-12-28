package data

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCascadeStats_Add(t *testing.T) {
	stats := &CascadeStats{}

	stats.Add("users", 1)
	stats.Add("conversations", 5)
	stats.Add("messages", 100)

	assert.Len(t, stats.Results, 3)
	assert.Equal(t, int64(106), stats.TotalRowsAffected)
	assert.Equal(t, "users", stats.Results[0].Table)
	assert.Equal(t, int64(1), stats.Results[0].RowsAffected)
}

func TestErrors(t *testing.T) {
	assert.EqualError(t, ErrNoRowsAffected, "cascade: no rows affected")
	assert.EqualError(t, ErrNilTransaction, "cascade: transaction is nil")
}

func TestDeleteUserWithTx_NilTransaction(t *testing.T) {
	_, err := DeleteUserWithTx(context.Background(), nil, 123)
	assert.ErrorIs(t, err, ErrNilTransaction)
}

func TestDeleteCharacterWithTx_NilTransaction(t *testing.T) {
	_, err := DeleteCharacterWithTx(context.Background(), nil, 123)
	assert.ErrorIs(t, err, ErrNilTransaction)
}

func TestDeleteConversationWithTx_NilTransaction(t *testing.T) {
	_, err := DeleteConversationWithTx(context.Background(), nil, 123)
	assert.ErrorIs(t, err, ErrNilTransaction)
}

func TestNewCascadeManager(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewCascadeManager(db)
	assert.NotNil(t, manager)
}

// Integration-style tests with sqlmock

func TestDeleteUserCascade_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(12345)

	mock.ExpectBegin()

	// user_profiles
	mock.ExpectExec("DELETE FROM user_profiles WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// sessions
	mock.ExpectExec("DELETE FROM sessions WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// conversations query
	mock.ExpectQuery("SELECT id FROM conversations WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100).AddRow(101))

	// messages
	mock.ExpectExec("DELETE FROM messages WHERE conversation_id IN").
		WithArgs(int64(100), int64(101)).
		WillReturnResult(sqlmock.NewResult(0, 50))

	// conversations
	mock.ExpectExec("DELETE FROM conversations WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// memories
	mock.ExpectExec("DELETE FROM memories WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 10))

	// relationships
	mock.ExpectExec("DELETE FROM relationships WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// important_events
	mock.ExpectExec("DELETE FROM important_events WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 3))

	// user_portraits
	mock.ExpectExec("DELETE FROM user_portraits WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// credit_transactions
	mock.ExpectExec("DELETE FROM credit_transactions WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 5))

	// credit_accounts
	mock.ExpectExec("DELETE FROM credit_accounts WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// orders
	mock.ExpectExec("DELETE FROM orders WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// characters query - no characters
	mock.ExpectQuery("SELECT id FROM characters WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// users
	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	stats, err := manager.DeleteUserCascade(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Greater(t, stats.TotalRowsAffected, int64(0))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUserCascade_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(99999)

	mock.ExpectBegin()

	// user_profiles
	mock.ExpectExec("DELETE FROM user_profiles").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// sessions
	mock.ExpectExec("DELETE FROM sessions").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// conversations query
	mock.ExpectQuery("SELECT id FROM conversations").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// conversations
	mock.ExpectExec("DELETE FROM conversations").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// memories
	mock.ExpectExec("DELETE FROM memories").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// relationships
	mock.ExpectExec("DELETE FROM relationships").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// important_events
	mock.ExpectExec("DELETE FROM important_events").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// user_portraits
	mock.ExpectExec("DELETE FROM user_portraits").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// credit_transactions
	mock.ExpectExec("DELETE FROM credit_transactions").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// credit_accounts
	mock.ExpectExec("DELETE FROM credit_accounts").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// orders
	mock.ExpectExec("DELETE FROM orders").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// characters query
	mock.ExpectQuery("SELECT id FROM characters").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// users - not found
	mock.ExpectExec("DELETE FROM users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectRollback()

	manager := NewCascadeManager(db)
	_, err = manager.DeleteUserCascade(context.Background(), userID)

	assert.ErrorIs(t, err, ErrNoRowsAffected)
}

func TestDeleteCharacterCascade_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	characterID := int64(54321)

	mock.ExpectBegin()

	// conversations query
	mock.ExpectQuery("SELECT id FROM conversations WHERE character_id = \\$1").
		WithArgs(characterID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(200))

	// messages
	mock.ExpectExec("DELETE FROM messages WHERE conversation_id IN").
		WithArgs(int64(200)).
		WillReturnResult(sqlmock.NewResult(0, 20))

	// conversations
	mock.ExpectExec("DELETE FROM conversations WHERE character_id = \\$1").
		WithArgs(characterID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// memories
	mock.ExpectExec("DELETE FROM memories WHERE character_id = \\$1").
		WithArgs(characterID).
		WillReturnResult(sqlmock.NewResult(0, 5))

	// relationships
	mock.ExpectExec("DELETE FROM relationships WHERE character_id = \\$1").
		WithArgs(characterID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// important_events
	mock.ExpectExec("DELETE FROM important_events WHERE character_id = \\$1").
		WithArgs(characterID).
		WillReturnResult(sqlmock.NewResult(0, 2))

	// character
	mock.ExpectExec("DELETE FROM characters WHERE id = \\$1").
		WithArgs(characterID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	stats, err := manager.DeleteCharacterCascade(context.Background(), characterID)

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(30), stats.TotalRowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteConversationCascade_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conversationID := int64(1000)

	mock.ExpectBegin()

	// messages
	mock.ExpectExec("DELETE FROM messages WHERE conversation_id = \\$1").
		WithArgs(conversationID).
		WillReturnResult(sqlmock.NewResult(0, 25))

	// conversation
	mock.ExpectExec("DELETE FROM conversations WHERE id = \\$1").
		WithArgs(conversationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	stats, err := manager.DeleteConversationCascade(context.Background(), conversationID)

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(26), stats.TotalRowsAffected)
	assert.Len(t, stats.Results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSoftDeleteUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(12345)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	err = manager.SoftDeleteUser(context.Background(), userID)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSoftDeleteUser_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(99999)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	manager := NewCascadeManager(db)
	err = manager.SoftDeleteUser(context.Background(), userID)

	assert.ErrorIs(t, err, ErrNoRowsAffected)
}

func TestRecoverUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(12345)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	err = manager.RecoverUser(context.Background(), userID)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecoverUser_ExpiredOrNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(99999)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	manager := NewCascadeManager(db)
	err = manager.RecoverUser(context.Background(), userID)

	assert.ErrorIs(t, err, ErrNoRowsAffected)
}

func TestSoftDeleteCharacter_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	characterID := int64(54321)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE characters").
		WithArgs(characterID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	err = manager.SoftDeleteCharacter(context.Background(), characterID)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestArchiveConversation_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conversationID := int64(1000)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE conversations").
		WithArgs(conversationID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	err = manager.ArchiveConversation(context.Background(), conversationID)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBatchDeleteConversations_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conversationIDs := []int64{100, 101, 102}

	mock.ExpectBegin()

	// messages
	mock.ExpectExec("DELETE FROM messages WHERE conversation_id IN").
		WithArgs(int64(100), int64(101), int64(102)).
		WillReturnResult(sqlmock.NewResult(0, 60))

	// conversations
	mock.ExpectExec("DELETE FROM conversations WHERE id IN").
		WithArgs(int64(100), int64(101), int64(102)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	stats, err := manager.BatchDeleteConversations(context.Background(), conversationIDs)

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(63), stats.TotalRowsAffected)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBatchDeleteConversations_EmptyList(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	manager := NewCascadeManager(db)
	stats, err := manager.BatchDeleteConversations(context.Background(), []int64{})

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalRowsAffected)
}

func TestPurgeExpiredUsers_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	// Query for expired users
	mock.ExpectQuery("SELECT id FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(111)).AddRow(int64(222)))

	// Delete first user (111)
	mock.ExpectExec("DELETE FROM user_profiles").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM sessions").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id FROM conversations WHERE user_id").WithArgs(int64(111)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("DELETE FROM conversations").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM memories").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM relationships").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM important_events").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM user_portraits").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM credit_transactions").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM credit_accounts").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM orders").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id FROM characters WHERE user_id").WithArgs(int64(111)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("DELETE FROM users").WithArgs(int64(111)).WillReturnResult(sqlmock.NewResult(0, 1))

	// Delete second user (222)
	mock.ExpectExec("DELETE FROM user_profiles").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM sessions").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id FROM conversations WHERE user_id").WithArgs(int64(222)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("DELETE FROM conversations").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM memories").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM relationships").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM important_events").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM user_portraits").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM credit_transactions").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM credit_accounts").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM orders").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT id FROM characters WHERE user_id").WithArgs(int64(222)).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("DELETE FROM users").WithArgs(int64(222)).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	count, err := manager.PurgeExpiredUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPurgeExpiredUsers_NoUsersToDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()

	manager := NewCascadeManager(db)
	count, err := manager.PurgeExpiredUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteUserCascade_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin().WillReturnError(errors.New("connection error"))

	manager := NewCascadeManager(db)
	_, err = manager.DeleteUserCascade(context.Background(), 12345)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
}

func TestDeleteUserCascade_DeleteError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	userID := int64(12345)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM user_profiles").
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	manager := NewCascadeManager(db)
	_, err = manager.DeleteUserCascade(context.Background(), userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_profiles")
}

// Helper function tests

func TestDeleteMessagesByConversations_EmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()

	rows, err := deleteMessagesByConversations(context.Background(), tx, []int64{})

	require.NoError(t, err)
	assert.Equal(t, int64(0), rows)
}
