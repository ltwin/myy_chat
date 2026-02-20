// Package biz 用户业务逻辑层
package biz

import (
	"context"
	"errors"
	"time"
)

// 积分账户相关错误
var (
	ErrInsufficientBalance   = errors.New("insufficient credit balance")
	ErrCreditAccountNotFound = errors.New("credit account not found")
	ErrDuplicateTransaction  = errors.New("duplicate transaction")
)

// CreditAccount 积分账户实体
type CreditAccount struct {
	UserID        int64     // 用户ID
	Balance       int64     // 当前余额
	TotalCharged  int64     // 累计充值
	TotalConsumed int64     // 累计消费
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// NewCreditAccount 创建新的积分账户
func NewCreditAccount(userID int64) *CreditAccount {
	now := time.Now()
	return &CreditAccount{
		UserID:        userID,
		Balance:       0,
		TotalCharged:  0,
		TotalConsumed: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// AddCredits 增加积分
func (a *CreditAccount) AddCredits(amount int64) {
	a.Balance += amount
	a.TotalCharged += amount
	a.UpdatedAt = time.Now()
}

// DeductCredits 扣除积分
func (a *CreditAccount) DeductCredits(amount int64) error {
	if a.Balance < amount {
		return ErrInsufficientBalance
	}
	a.Balance -= amount
	a.TotalConsumed += amount
	a.UpdatedAt = time.Now()
	return nil
}

// CanAfford 检查是否有足够积分
func (a *CreditAccount) CanAfford(amount int64) bool {
	return a.Balance >= amount
}

// CreditTransaction 积分交易记录
type CreditTransaction struct {
	ID             int64     // 交易ID (雪花ID)
	UserID         int64     // 用户ID
	Type           string    // 交易类型: charge, consume, refund, bonus
	Amount         int64     // 交易金额 (正数增加，负数减少)
	BalanceAfter   int64     // 交易后余额
	Description    string    // 交易描述
	ReferenceType  string    // 关联类型: order, llm_call, admin, registration
	ReferenceID    int64     // 关联ID
	IdempotencyKey string    // 幂等键
	CreatedAt      time.Time // 创建时间
}

// TransactionType 交易类型常量
const (
	TransactionTypeCharge  = "charge"  // 充值
	TransactionTypeConsume = "consume" // 消费
	TransactionTypeRefund  = "refund"  // 退款
	TransactionTypeBonus   = "bonus"   // 赠送
)

// ReferenceType 关联类型常量
const (
	ReferenceTypeOrder        = "order"        // 订单
	ReferenceTypeLLMCall      = "llm_call"     // LLM调用
	ReferenceTypeAdmin        = "admin"        // 管理员操作
	ReferenceTypeRegistration = "registration" // 注册赠送
)

// CreditAccountRepo 积分账户仓储接口
type CreditAccountRepo interface {
	// Create 创建积分账户
	Create(ctx context.Context, account *CreditAccount) error
	// GetByUserID 根据用户ID获取账户
	GetByUserID(ctx context.Context, userID int64) (*CreditAccount, error)
	// Update 更新积分账户
	Update(ctx context.Context, account *CreditAccount) error
	// AddCredits 增加积分 (带交易记录)
	AddCredits(ctx context.Context, userID int64, amount int64, reason, refType string, refID int64) error
	// DeductCredits 扣除积分 (带交易记录)
	DeductCredits(ctx context.Context, userID int64, amount int64, reason, refType string, refID int64, idempotencyKey string) error
	// GetTransactions 获取交易记录
	GetTransactions(ctx context.Context, userID int64, limit, offset int) ([]*CreditTransaction, int, error)
	// Delete 删除积分账户
	Delete(ctx context.Context, userID int64) error
}
