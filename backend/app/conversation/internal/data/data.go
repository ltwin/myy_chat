package data

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/golang/app/conversation/internal/conf"
	"github.com/myy-chat/backend/golang/pkg/snowflake"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewConversationRepo,
	NewMessageRepo,
	NewLLMClient,
	NewCreditClient,
	NewSnowflakeGenerator,
)

// Data is the data layer wrapper.
type Data struct {
	db *pgxpool.Pool
}

// NewData creates a new data layer.
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	helper := log.NewHelper(logger)

	// 创建数据库连接池
	db, err := pgxpool.New(context.Background(), c.Database.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	cleanup := func() {
		helper.Info("closing the data resources")
		db.Close()
	}

	return &Data{db: db}, cleanup, nil
}

// NewSnowflakeGenerator 创建雪花ID生成器
func NewSnowflakeGenerator(logger log.Logger) (snowflake.Generator, error) {
	gen, err := snowflake.NewGeneratorFromEnv()
	if err != nil {
		// 使用默认节点ID
		gen, err = snowflake.NewGenerator(0)
		if err != nil {
			return nil, err
		}
	}
	return gen, nil
}
