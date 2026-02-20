// Package data 用户数据访问层
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/myy-chat/backend/app/user/internal/biz"
	"github.com/myy-chat/backend/app/user/internal/conf"
	"github.com/myy-chat/backend/pkg/snowflake"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewPostgresPool,
	NewSnowflakeGenerator,
	NewUserRepo,
	NewUserProfileRepo,
	NewCreditAccountRepo,
	NewSessionRepo,
)

// Data is the data layer wrapper.
type Data struct {
	DB    *pgxpool.Pool
	IDGen snowflake.Generator
}

// NewData creates a new data layer.
func NewData(db *pgxpool.Pool, idGen snowflake.Generator, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		db.Close()
	}
	return &Data{
		DB:    db,
		IDGen: idGen,
	}, cleanup, nil
}

// NewPostgresPool 创建 PostgreSQL 连接池
func NewPostgresPool(c *conf.Data, logger log.Logger) (*pgxpool.Pool, error) {
	log := log.NewHelper(logger)

	config, err := pgxpool.ParseConfig(c.Database.Source)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	config.MaxConns = 50
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	log.Info("connected to PostgreSQL")
	return pool, nil
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

// Repos 返回所有仓储接口（便于测试时 mock）
type Repos struct {
	UserRepo       biz.UserRepo
	ProfileRepo    biz.UserProfileRepo
	CreditRepo     biz.CreditAccountRepo
	SessionRepo    biz.SessionRepo
}
