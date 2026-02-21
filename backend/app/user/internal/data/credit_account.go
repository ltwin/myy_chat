// Package data 用户数据访问层
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/app/user/internal/biz"
	"github.com/myy-chat/backend/pkg/snowflake"
)

// creditAccountRepo 积分账户仓储实现
type creditAccountRepo struct {
	db    *pgxpool.Pool
	idGen snowflake.Generator
	log   *log.Helper
}

// NewCreditAccountRepo 创建积分账户仓储
func NewCreditAccountRepo(db *pgxpool.Pool, idGen snowflake.Generator, logger log.Logger) biz.CreditAccountRepo {
	return &creditAccountRepo{
		db:    db,
		idGen: idGen,
		log:   log.NewHelper(logger),
	}
}

// Create 创建积分账户
func (r *creditAccountRepo) Create(ctx context.Context, account *biz.CreditAccount) error {
	query := `
		INSERT INTO credit_accounts (user_id, balance, total_charged, total_consumed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		account.UserID,
		account.Balance,
		account.TotalCharged,
		account.TotalConsumed,
		account.CreatedAt,
		account.UpdatedAt,
	)
	return err
}

// GetByUserID 根据用户ID获取账户
func (r *creditAccountRepo) GetByUserID(ctx context.Context, userID int64) (*biz.CreditAccount, error) {
	query := `
		SELECT user_id, balance, total_charged, total_consumed, created_at, updated_at
		FROM credit_accounts
		WHERE user_id = $1
	`

	var account biz.CreditAccount
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&account.UserID,
		&account.Balance,
		&account.TotalCharged,
		&account.TotalConsumed,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, biz.ErrCreditAccountNotFound
		}
		return nil, err
	}

	return &account, nil
}

// Update 更新积分账户
func (r *creditAccountRepo) Update(ctx context.Context, account *biz.CreditAccount) error {
	query := `
		UPDATE credit_accounts SET
			balance = $2,
			total_charged = $3,
			total_consumed = $4,
			updated_at = NOW()
		WHERE user_id = $1
	`
	result, err := r.db.Exec(ctx, query,
		account.UserID,
		account.Balance,
		account.TotalCharged,
		account.TotalConsumed,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return biz.ErrCreditAccountNotFound
	}

	return nil
}

// AddCredits 增加积分 (带交易记录)
func (r *creditAccountRepo) AddCredits(ctx context.Context, userID int64, amount int64, reason, refType string, refID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 更新账户余额
	var balanceAfter int64
	updateQuery := `
		UPDATE credit_accounts
		SET balance = balance + $2, total_charged = total_charged + $2, updated_at = NOW()
		WHERE user_id = $1
		RETURNING balance
	`
	err = tx.QueryRow(ctx, updateQuery, userID, amount).Scan(&balanceAfter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return biz.ErrCreditAccountNotFound
		}
		return err
	}

	// 创建交易记录
	transactionID := r.idGen.Generate()
	insertQuery := `
		INSERT INTO credit_transactions (
			id, user_id, type, transaction_type, amount, balance_after,
			description, reference_type, reference_id, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'SUCCESS', NOW())
	`
	_, err = tx.Exec(ctx, insertQuery,
		transactionID,
		userID,
		biz.TransactionTypeBonus,
		"GRANT",
		amount,
		balanceAfter,
		reason,
		refType,
		refID,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeductCredits 扣除积分 (带交易记录)
func (r *creditAccountRepo) DeductCredits(ctx context.Context, userID int64, amount int64, reason, refType string, refID int64, idempotencyKey string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 检查幂等性
	if idempotencyKey != "" {
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM credit_transactions WHERE idempotency_key = $1)`
		err = tx.QueryRow(ctx, checkQuery, idempotencyKey).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			return biz.ErrDuplicateTransaction
		}
	}

	// 检查余额并更新
	var balanceAfter int64
	updateQuery := `
		UPDATE credit_accounts
		SET balance = balance - $2, total_consumed = total_consumed + $2, updated_at = NOW()
		WHERE user_id = $1 AND balance >= $2
		RETURNING balance
	`
	err = tx.QueryRow(ctx, updateQuery, userID, amount).Scan(&balanceAfter)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 可能是余额不足或账户不存在
			var currentBalance int64
			checkBalanceQuery := `SELECT balance FROM credit_accounts WHERE user_id = $1`
			checkErr := tx.QueryRow(ctx, checkBalanceQuery, userID).Scan(&currentBalance)
			if checkErr != nil {
				if errors.Is(checkErr, pgx.ErrNoRows) {
					return biz.ErrCreditAccountNotFound
				}
				return checkErr
			}
			return biz.ErrInsufficientBalance
		}
		return err
	}

	// 创建交易记录
	transactionID := r.idGen.Generate()
	insertQuery := `
		INSERT INTO credit_transactions (
			id, user_id, type, transaction_type, amount, balance_after,
			description, reference_type, reference_id, idempotency_key, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'SUCCESS', NOW())
	`
	var idempKey sql.NullString
	if idempotencyKey != "" {
		idempKey = sql.NullString{String: idempotencyKey, Valid: true}
	}
	_, err = tx.Exec(ctx, insertQuery,
		transactionID,
		userID,
		biz.TransactionTypeConsume,
		"SETTLE",
		-amount, // 消费记为负数
		balanceAfter,
		reason,
		refType,
		refID,
		idempKey,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetTransactions 获取交易记录
func (r *creditAccountRepo) GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*biz.CreditTransaction, int, error) {
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM credit_transactions WHERE user_id = $1`
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取记录
	query := `
		SELECT
			id,
			user_id,
			COALESCE(type, lower(transaction_type)) AS type,
			amount,
			balance_after,
			description,
			reference_type,
			reference_id,
			idempotency_key,
			created_at
		FROM credit_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*biz.CreditTransaction
	for rows.Next() {
		var t biz.CreditTransaction
		var description, refType, idempKey sql.NullString
		var refID sql.NullInt64

		err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Type,
			&t.Amount,
			&t.BalanceAfter,
			&description,
			&refType,
			&refID,
			&idempKey,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if description.Valid {
			t.Description = description.String
		}
		if refType.Valid {
			t.ReferenceType = refType.String
		}
		if refID.Valid {
			t.ReferenceID = refID.Int64
		}
		if idempKey.Valid {
			t.IdempotencyKey = idempKey.String
		}

		transactions = append(transactions, &t)
	}

	return transactions, total, rows.Err()
}

// Delete 删除积分账户
func (r *creditAccountRepo) Delete(ctx context.Context, userID int64) error {
	// 先删除交易记录
	_, err := r.db.Exec(ctx, `DELETE FROM credit_transactions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete transactions: %w", err)
	}

	// 删除账户
	_, err = r.db.Exec(ctx, `DELETE FROM credit_accounts WHERE user_id = $1`, userID)
	return err
}
