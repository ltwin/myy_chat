// Package data 积分服务客户端实现
package data

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	userv1 "github.com/myy-chat/backend/api/user/v1"
	"github.com/myy-chat/backend/app/conversation/internal/biz"
)

const (
	defaultUserServiceGRPCAddr = "127.0.0.1:9000"
	userServiceGRPCAddrEnv     = "USER_SERVICE_GRPC_ADDR"
)

// creditClient 积分服务客户端实现
// 通过 gRPC 调用 User Service 的积分接口
type creditClient struct {
	log    *log.Helper
	conn   *grpc.ClientConn
	client userv1.UserServiceClient
}

// NewCreditClient 创建积分客户端
func NewCreditClient(logger log.Logger) (biz.CreditService, func(), error) {
	helper := log.NewHelper(logger)
	addr := strings.TrimSpace(os.Getenv(userServiceGRPCAddrEnv))
	if addr == "" {
		addr = defaultUserServiceGRPCAddr
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect user service grpc addr=%s: %w", addr, err)
	}
	helper.Infof("credit client connected to user service: %s", addr)

	cleanup := func() {
		if err := conn.Close(); err != nil {
			helper.Warnf("failed to close credit grpc conn: %v", err)
		}
	}

	return &creditClient{
		log:    helper,
		conn:   conn,
		client: userv1.NewUserServiceClient(conn),
	}, cleanup, nil
}

// DeductCredits 扣除积分
func (c *creditClient) DeductCredits(ctx context.Context, userID int64, amount float64, reason string) error {
	c.log.Infof("Deduct credits: user_id=%d, amount=%.4f, reason=%s", userID, amount, reason)

	credits := creditsFromAmount(amount)
	if credits <= 0 {
		return nil
	}

	_, err := c.client.DeductCredits(ctx, &userv1.DeductCreditsRequest{
		UserId:         userID,
		Amount:         credits,
		Reason:         reason,
		ReferenceType:  "conversation",
		ReferenceId:    reason,
		IdempotencyKey: buildConversationIdempotencyKey(userID, reason, credits),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
			return biz.ErrInsufficientCredits
		}
		return err
	}

	return nil
}

// GetBalance 获取积分余额
func (c *creditClient) GetBalance(ctx context.Context, userID int64) (float64, error) {
	c.log.Infof("Get credit balance: user_id=%d", userID)

	resp, err := c.client.GetCreditBalance(ctx, &userv1.GetCreditBalanceRequest{UserId: userID})
	if err != nil {
		return 0, err
	}
	return float64(resp.Balance), nil
}

func creditsFromAmount(amount float64) int64 {
	if amount <= 0 {
		return 0
	}
	// MVP 阶段使用整数积分，浮点成本按向上取整折算成积分单位。
	return int64(math.Ceil(amount))
}

func buildConversationIdempotencyKey(userID int64, reason string, credits int64) string {
	normalizedReason := strings.TrimSpace(reason)
	if normalizedReason == "" {
		normalizedReason = "conversation"
	}
	return fmt.Sprintf("conv-deduct:%d:%s:%d", userID, normalizedReason, credits)
}
