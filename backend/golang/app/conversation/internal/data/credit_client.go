// Package data 积分服务客户端实现
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/myy-chat/backend/golang/app/conversation/internal/biz"
)

// creditClient 积分服务客户端实现
// 通过 gRPC 调用 User Service 的积分接口
type creditClient struct {
	log *log.Helper
	// TODO: 添加 gRPC 客户端连接
	// client user_v1.UserServiceClient
}

// NewCreditClient 创建积分客户端
func NewCreditClient(logger log.Logger) biz.CreditService {
	return &creditClient{
		log: log.NewHelper(logger),
	}
}

// DeductCredits 扣除积分
func (c *creditClient) DeductCredits(ctx context.Context, userID int64, amount float64, reason string) error {
	c.log.Infof("Deduct credits: user_id=%d, amount=%.4f, reason=%s", userID, amount, reason)

	// TODO: 实现实际的 gRPC 调用
	return nil
}

// GetBalance 获取积分余额
func (c *creditClient) GetBalance(ctx context.Context, userID int64) (float64, error) {
	c.log.Infof("Get credit balance: user_id=%d", userID)

	// TODO: 实现实际的 gRPC 调用
	// 临时返回模拟余额
	return 100.0, nil
}
