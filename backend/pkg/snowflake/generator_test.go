package snowflake

import (
	"os"
	"sync"
	"testing"
)

// TestNewGenerator 测试创建生成器
func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  int64
		wantErr bool
	}{
		{"valid node ID 0", 0, false},
		{"valid node ID 1", 1, false},
		{"valid node ID 1023", 1023, false},
		{"valid node ID 512", 512, false},
		{"invalid node ID -1", -1, true},
		{"invalid node ID 1024", 1024, true},
		{"invalid node ID 2000", 2000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGenerator(tt.nodeID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewGenerator(%d) expected error, got nil", tt.nodeID)
				}
			} else {
				if err != nil {
					t.Errorf("NewGenerator(%d) unexpected error: %v", tt.nodeID, err)
				}
				if gen == nil {
					t.Errorf("NewGenerator(%d) returned nil generator", tt.nodeID)
				}
			}
		})
	}
}

// TestGenerate 测试ID生成
func TestGenerate(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// 生成多个ID
	ids := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id := gen.Generate()
		if id <= 0 {
			t.Errorf("Generate() returned non-positive ID: %d", id)
		}
		if ids[id] {
			t.Errorf("Generate() returned duplicate ID: %d", id)
		}
		ids[id] = true
	}

	// 验证唯一性
	if len(ids) != 1000 {
		t.Errorf("Expected 1000 unique IDs, got %d", len(ids))
	}
}

// TestGenerateString 测试字符串ID生成
func TestGenerateString(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	id := gen.GenerateString()
	if id == "" {
		t.Error("GenerateString() returned empty string")
	}
}

// TestParse 测试ID解析
func TestParse(t *testing.T) {
	gen, err := NewGenerator(42)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// 生成ID并解析
	id := gen.Generate()
	nodeID, timestamp, sequence := gen.Parse(id)

	// 验证节点ID
	if nodeID != 42 {
		t.Errorf("Parse() nodeID = %d, want 42", nodeID)
	}

	// 验证时间戳是正数
	if timestamp <= 0 {
		t.Errorf("Parse() timestamp = %d, want positive", timestamp)
	}

	// 验证序列号在有效范围
	if sequence < 0 || sequence > 4095 {
		t.Errorf("Parse() sequence = %d, want 0-4095", sequence)
	}
}

// TestConcurrentGenerate 测试并发生成ID
func TestConcurrentGenerate(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var wg sync.WaitGroup
	idsChan := make(chan int64, 10000)

	// 启动10个goroutine，每个生成1000个ID
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				idsChan <- gen.Generate()
			}
		}()
	}

	wg.Wait()
	close(idsChan)

	// 收集所有ID并检查唯一性
	ids := make(map[int64]bool)
	for id := range idsChan {
		if ids[id] {
			t.Errorf("Concurrent Generate() returned duplicate ID: %d", id)
		}
		ids[id] = true
	}

	if len(ids) != 10000 {
		t.Errorf("Expected 10000 unique IDs, got %d", len(ids))
	}
}

// TestIDOrdering 测试ID顺序性
func TestIDOrdering(t *testing.T) {
	gen, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	var prevID int64 = 0
	for i := 0; i < 100; i++ {
		id := gen.Generate()
		if id <= prevID {
			t.Errorf("ID %d is not greater than previous ID %d", id, prevID)
		}
		prevID = id
	}
}

// TestNewGeneratorFromEnv 测试从环境变量创建生成器
func TestNewGeneratorFromEnv(t *testing.T) {
	// 保存原始环境变量
	origDatacenter := os.Getenv("DATACENTER_ID")
	origWorker := os.Getenv("WORKER_ID")
	defer func() {
		os.Setenv("DATACENTER_ID", origDatacenter)
		os.Setenv("WORKER_ID", origWorker)
	}()

	tests := []struct {
		name         string
		datacenterID string
		workerID     string
		wantErr      bool
	}{
		{"default values", "", "", false},
		{"valid datacenter and worker", "1", "2", false},
		{"max datacenter", "31", "31", false},
		{"invalid datacenter", "32", "0", true},
		{"invalid worker", "0", "32", true},
		{"negative datacenter", "-1", "0", true},
		{"non-numeric datacenter", "abc", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DATACENTER_ID", tt.datacenterID)
			os.Setenv("WORKER_ID", tt.workerID)

			gen, err := NewGeneratorFromEnv()
			if tt.wantErr {
				if err == nil {
					t.Error("NewGeneratorFromEnv() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("NewGeneratorFromEnv() unexpected error: %v", err)
				}
				if gen == nil {
					t.Error("NewGeneratorFromEnv() returned nil generator")
				}
			}
		})
	}
}

// TestNodeIDCalculation 测试NodeID计算
func TestNodeIDCalculation(t *testing.T) {
	// 保存原始环境变量
	origDatacenter := os.Getenv("DATACENTER_ID")
	origWorker := os.Getenv("WORKER_ID")
	defer func() {
		os.Setenv("DATACENTER_ID", origDatacenter)
		os.Setenv("WORKER_ID", origWorker)
	}()

	tests := []struct {
		datacenterID string
		workerID     string
		wantNodeID   int64
	}{
		{"0", "0", 0},
		{"0", "1", 1},
		{"1", "0", 32},
		{"1", "1", 33},
		{"31", "31", 1023},
		{"15", "15", 495}, // 15*32 + 15 = 495
	}

	for _, tt := range tests {
		t.Run(tt.datacenterID+"_"+tt.workerID, func(t *testing.T) {
			os.Setenv("DATACENTER_ID", tt.datacenterID)
			os.Setenv("WORKER_ID", tt.workerID)

			gen, err := NewGeneratorFromEnv()
			if err != nil {
				t.Fatalf("NewGeneratorFromEnv() error: %v", err)
			}

			id := gen.Generate()
			nodeID, _, _ := gen.Parse(id)

			if nodeID != tt.wantNodeID {
				t.Errorf("NodeID = %d, want %d", nodeID, tt.wantNodeID)
			}
		})
	}
}

// TestGlobalFunctions 测试全局函数
func TestGlobalFunctions(t *testing.T) {
	// 重置默认节点用于测试
	defaultNode = nil
	initOnce = sync.Once{}
	initErr = nil

	// 保存原始环境变量
	origDatacenter := os.Getenv("DATACENTER_ID")
	origWorker := os.Getenv("WORKER_ID")
	defer func() {
		os.Setenv("DATACENTER_ID", origDatacenter)
		os.Setenv("WORKER_ID", origWorker)
		// 重置默认节点
		defaultNode = nil
		initOnce = sync.Once{}
		initErr = nil
	}()

	os.Setenv("DATACENTER_ID", "1")
	os.Setenv("WORKER_ID", "1")

	// 测试Generate
	id, err := Generate()
	if err != nil {
		t.Errorf("Generate() error: %v", err)
	}
	if id <= 0 {
		t.Errorf("Generate() returned non-positive ID: %d", id)
	}

	// 测试GenerateString
	idStr, err := GenerateString()
	if err != nil {
		t.Errorf("GenerateString() error: %v", err)
	}
	if idStr == "" {
		t.Error("GenerateString() returned empty string")
	}

	// 测试Parse
	nodeID, timestamp, sequence := Parse(id)
	if nodeID != 33 { // 1*32 + 1 = 33
		t.Errorf("Parse() nodeID = %d, want 33", nodeID)
	}
	if timestamp <= 0 {
		t.Errorf("Parse() timestamp = %d, want positive", timestamp)
	}
	if sequence < 0 {
		t.Errorf("Parse() sequence = %d, want non-negative", sequence)
	}
}

// TestMustGenerate 测试MustGenerate
func TestMustGenerate(t *testing.T) {
	// 重置默认节点用于测试
	defaultNode = nil
	initOnce = sync.Once{}
	initErr = nil

	// 保存原始环境变量
	origDatacenter := os.Getenv("DATACENTER_ID")
	origWorker := os.Getenv("WORKER_ID")
	defer func() {
		os.Setenv("DATACENTER_ID", origDatacenter)
		os.Setenv("WORKER_ID", origWorker)
		// 重置默认节点
		defaultNode = nil
		initOnce = sync.Once{}
		initErr = nil
	}()

	os.Setenv("DATACENTER_ID", "0")
	os.Setenv("WORKER_ID", "0")

	// 测试正常情况不会panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustGenerate() panicked unexpectedly: %v", r)
			}
		}()
		id := MustGenerate()
		if id <= 0 {
			t.Errorf("MustGenerate() returned non-positive ID: %d", id)
		}
	}()
}

// BenchmarkGenerate 性能测试
func BenchmarkGenerate(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate()
	}
}

// BenchmarkConcurrentGenerate 并发性能测试
func BenchmarkConcurrentGenerate(b *testing.B) {
	gen, err := NewGenerator(1)
	if err != nil {
		b.Fatalf("Failed to create generator: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gen.Generate()
		}
	})
}
