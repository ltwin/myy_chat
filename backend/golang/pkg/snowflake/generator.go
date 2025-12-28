// Package snowflake 提供雪花ID生成器
// 使用Twitter的Snowflake算法生成64位分布式唯一ID
package snowflake

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	// 默认节点实例
	defaultNode *snowflake.Node
	// 初始化锁
	initOnce sync.Once
	// 初始化错误
	initErr error
)

// Config 雪花ID生成器配置
type Config struct {
	// NodeID 节点ID，范围 0-1023
	// 由 DATACENTER_ID (0-31) * 32 + WORKER_ID (0-31) 组成
	NodeID int64
}

// Generator 雪花ID生成器接口
type Generator interface {
	// Generate 生成一个新的雪花ID
	Generate() int64
	// GenerateString 生成一个新的雪花ID字符串
	GenerateString() string
	// Parse 解析雪花ID，返回节点ID、时间戳、序列号
	Parse(id int64) (nodeID int64, timestamp int64, sequence int64)
}

// generator 雪花ID生成器实现
type generator struct {
	node *snowflake.Node
}

// NewGenerator 创建新的雪花ID生成器
// nodeID: 节点ID，范围 0-1023
func NewGenerator(nodeID int64) (Generator, error) {
	if nodeID < 0 || nodeID > 1023 {
		return nil, fmt.Errorf("node ID must be between 0 and 1023, got %d", nodeID)
	}

	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	return &generator{node: node}, nil
}

// NewGeneratorFromConfig 从配置创建雪花ID生成器
func NewGeneratorFromConfig(cfg Config) (Generator, error) {
	return NewGenerator(cfg.NodeID)
}

// NewGeneratorFromEnv 从环境变量创建雪花ID生成器
// 读取 DATACENTER_ID (0-31) 和 WORKER_ID (0-31) 环境变量
// NodeID = DATACENTER_ID * 32 + WORKER_ID
func NewGeneratorFromEnv() (Generator, error) {
	datacenterIDStr := os.Getenv("DATACENTER_ID")
	workerIDStr := os.Getenv("WORKER_ID")

	// 默认值
	var datacenterID, workerID int64 = 0, 0

	if datacenterIDStr != "" {
		id, err := strconv.ParseInt(datacenterIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid DATACENTER_ID: %w", err)
		}
		if id < 0 || id > 31 {
			return nil, fmt.Errorf("DATACENTER_ID must be between 0 and 31, got %d", id)
		}
		datacenterID = id
	}

	if workerIDStr != "" {
		id, err := strconv.ParseInt(workerIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid WORKER_ID: %w", err)
		}
		if id < 0 || id > 31 {
			return nil, fmt.Errorf("WORKER_ID must be between 0 and 31, got %d", id)
		}
		workerID = id
	}

	// NodeID = DATACENTER_ID * 32 + WORKER_ID
	// 这样可以支持 32 * 32 = 1024 个节点
	nodeID := datacenterID*32 + workerID

	return NewGenerator(nodeID)
}

// Generate 生成一个新的雪花ID
func (g *generator) Generate() int64 {
	return g.node.Generate().Int64()
}

// GenerateString 生成一个新的雪花ID字符串
func (g *generator) GenerateString() string {
	return g.node.Generate().String()
}

// Parse 解析雪花ID，返回节点ID、时间戳（毫秒）、序列号
func (g *generator) Parse(id int64) (nodeID int64, timestamp int64, sequence int64) {
	sfID := snowflake.ParseInt64(id)
	return sfID.Node(), sfID.Time(), sfID.Step()
}

// InitDefault 初始化默认的雪花ID生成器
// 从环境变量读取配置，仅初始化一次
func InitDefault() error {
	initOnce.Do(func() {
		gen, err := NewGeneratorFromEnv()
		if err != nil {
			initErr = err
			return
		}
		defaultNode = gen.(*generator).node
	})
	return initErr
}

// Generate 使用默认节点生成雪花ID
// 需要先调用 InitDefault() 初始化
func Generate() (int64, error) {
	if defaultNode == nil {
		if err := InitDefault(); err != nil {
			return 0, err
		}
	}
	return defaultNode.Generate().Int64(), nil
}

// MustGenerate 使用默认节点生成雪花ID
// 如果初始化失败会panic
func MustGenerate() int64 {
	id, err := Generate()
	if err != nil {
		panic(err)
	}
	return id
}

// GenerateString 使用默认节点生成雪花ID字符串
func GenerateString() (string, error) {
	if defaultNode == nil {
		if err := InitDefault(); err != nil {
			return "", err
		}
	}
	return defaultNode.Generate().String(), nil
}

// Parse 解析雪花ID
func Parse(id int64) (nodeID int64, timestamp int64, sequence int64) {
	sfID := snowflake.ParseInt64(id)
	return sfID.Node(), sfID.Time(), sfID.Step()
}
