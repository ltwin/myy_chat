# llm_agent Service Makefile (大仓模式)
# ============================================
# 注意: 此 Makefile 用于单个服务的开发调试
# 大部分操作应在根目录使用统一 Makefile

GOHOSTOS := $(shell go env GOHOSTOS)
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "v0.0.1")
SERVICE := llm_agent

# 跨平台兼容性
ifeq ($(GOHOSTOS), windows)
	Git_Bash := $(subst \,/,$(subst cmd\git.exe,bin\bash.exe,"$(shell where git)"))
	INTERNAL_PROTO_FILES := $(shell $(Git_Bash) -c "find internal -name '*.proto'")
else
	INTERNAL_PROTO_FILES := $(shell find internal -name '*.proto' 2>/dev/null)
endif

.PHONY: help
help:
	@echo ""
	@echo "════════════════════════════════════════"
	@echo "  $(SERVICE) Service (单服务开发工具)"
	@echo "════════════════════════════════════════"
	@echo ""
	@echo "建议使用根目录 Makefile 进行统一操作:"
	@echo "  cd ../../ && make proto-$(SERVICE)      # 生成 proto"
	@echo "  cd ../../ && make build-$(SERVICE)      # 构建服务"
	@echo "  cd ../../ && make run-$(SERVICE)        # 运行服务"
	@echo ""
	@echo "本地开发命令 (当前目录):"
	@echo "  make config        - 生成 internal proto (conf.proto)"
	@echo "  make generate      - 生成 wire 依赖注入"
	@echo "  make build         - 构建当前服务"
	@echo "  make run           - 运行当前服务"
	@echo "  make test          - 运行单元测试"
	@echo "  make test-coverage - 测试并生成覆盖率"
	@echo "  make debug         - 调试模式 (Delve, port 2345)"
	@echo "  make clean         - 清理构建产物"
	@echo ""

# ================================================================
# 本地 Proto 生成 (仅限 internal 配置)
# ================================================================

.PHONY: config
# 生成 internal 的 proto 配置文件
config:
	@if [ -n "$(INTERNAL_PROTO_FILES)" ]; then \
		echo ">>> 生成 internal proto 配置..."; \
		protoc \
			--proto_path=./internal \
			--proto_path=../../third_party \
			--go_out=paths=source_relative:./internal \
			$(INTERNAL_PROTO_FILES); \
		echo "✓ Internal proto 配置生成完成!"; \
	else \
		echo "✓ 无 internal proto 文件，跳过"; \
	fi

# ================================================================
# Wire 依赖注入
# ================================================================

.PHONY: generate
# 生成 wire 依赖注入代码
generate:
	@if [ -f cmd/$(SERVICE)/wire.go ]; then \
		echo ">>> 生成 wire 代码..."; \
		cd cmd/$(SERVICE) && wire && cd ../..; \
		echo "✓ Wire 代码生成完成!"; \
	else \
		echo "✗ wire.go 不存在，跳过"; \
	fi
	@cd ../../ && go mod tidy

# ================================================================
# 构建和运行
# ================================================================

.PHONY: build
# 构建当前服务
build:
	@echo ">>> 构建 $(SERVICE) 服务..."
	@mkdir -p bin
	@if [ -d cmd/$(SERVICE) ]; then \
		go build -ldflags "-X main.Version=$(VERSION)" -o bin/$(SERVICE) ./cmd/$(SERVICE); \
		echo "✓ 构建完成: bin/$(SERVICE)"; \
	else \
		echo "✗ cmd/$(SERVICE) 不存在"; \
		exit 1; \
	fi

.PHONY: run
# 运行当前服务
run:
	@echo ">>> 运行 $(SERVICE) 服务..."
	@if [ -d cmd/$(SERVICE) ]; then \
		go run ./cmd/$(SERVICE) -conf ./configs; \
	else \
		echo "✗ cmd/$(SERVICE) 不存在"; \
		exit 1; \
	fi

.PHONY: clean
# 清理构建产物
clean:
	@echo ">>> 清理 $(SERVICE) 构建产物..."
	@rm -rf bin/
	@rm -f cmd/$(SERVICE)/wire_gen.go
	@echo "✓ 清理完成!"

# ================================================================
# 测试
# ================================================================

.PHONY: test
# 运行当前服务的测试
test:
	@echo ">>> 运行 $(SERVICE) 测试..."
	go test -v -race ./...

.PHONY: test-coverage
# 生成测试覆盖率报告
test-coverage:
	@echo ">>> 生成 $(SERVICE) 覆盖率报告..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ 覆盖率报告: coverage.html"

.PHONY: debug
# 调试模式运行 (使用 Delve)
debug:
	@echo ">>> 调试 $(SERVICE) (Delve headless mode, port 2345)..."
	@echo "    连接方式: dlv connect :2345"
	@echo "    或使用 IDE 远程调试连接 localhost:2345"
	@if [ -d cmd/$(SERVICE) ]; then \
		dlv debug ./cmd/$(SERVICE) --headless --listen=:2345 --api-version=2 --accept-multiclient -- -conf ./configs; \
	else \
		echo "✗ cmd/$(SERVICE) 不存在"; \
		exit 1; \
	fi

# ================================================================
# 快捷命令
# ================================================================

.PHONY: dev
# 开发模式: 生成代码 + 构建 + 运行
dev:
	@make config
	@make generate
	@make build
	@make run

.PHONY: all
# 完整流程
all:
	@make config
	@make generate
	@make build

.DEFAULT_GOAL := help
