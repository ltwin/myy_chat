# MyY Chat Backend Monorepo Makefile (Kratos v2 大仓模式)
# ================================================================
# 参照项目: buzzy-workflow (分仓模式)
# 当前模式: 大仓模式 - 7个服务共享 go.mod 和 api/ 目录

GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "v0.0.1")
SEP:=/

# 服务列表 (Go 服务)
SERVICES = user conversation memory character billing admin analytics demo llm_agent

# API 目录列表
API_SERVICES = user conversation memory character billing admin analytics demo llm_agent mcp

# 跨平台兼容性处理
ifeq ($(GOHOSTOS), windows)
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
	THIRD_PARTY_FILES=$(shell $(Git_Bash) -c "find third_party -name *.proto")
else
	API_PROTO_FILES=$(shell find api -name "*.proto")
	THIRD_PARTY_FILES=$(shell find third_party -name "*.proto")
endif

.PHONY: help
help:
	@echo ""
	@echo "═══════════════════════════════════════════════════════════"
	@echo "  MyY Chat Backend Monorepo (Kratos v2 大仓模式)"
	@echo "═══════════════════════════════════════════════════════════"
	@echo ""
	@echo "初始化命令:"
	@echo "  make init               - 安装所有 protoc 插件和工具"
	@echo "  make download           - 下载 Go 依赖 (go mod tidy)"
	@echo ""
	@echo "代码生成命令:"
	@echo "  make proto              - 生成所有服务的 API proto 代码"
	@echo "  make openapi            - 生成聚合 OpenAPI 文档 (openapi.yaml)"
	@echo "  make errors             - 生成所有服务的 error 定义"
	@echo "  make validate           - 生成所有服务的 validate 代码"
	@echo "  make generate           - 生成所有服务的 wire 依赖注入代码"
	@echo "  make setup              - 执行完整设置 (proto + openapi + errors + validate + generate)"
	@echo ""
	@echo "构建和测试:"
	@echo "  make build              - 构建所有服务到 bin/ 目录"
	@echo "  make test               - 运行所有单元测试"
	@echo "  make test-coverage      - 运行测试并生成覆盖率报告"
	@echo "  make lint               - 运行 golangci-lint 代码检查"
	@echo ""
	@echo "单服务操作 (替换 {service} 为服务名):"
	@echo "  make proto-{service}    - 生成指定服务的 proto 代码"
	@echo "  make build-{service}    - 构建指定服务"
	@echo "  make run-{service}      - 运行指定服务"
	@echo "  make wire-{service}     - 生成指定服务的 wire 代码"
	@echo "  make test-{service}     - 测试指定服务"
	@echo "  make test-coverage-{service} - 测试指定服务并生成覆盖率"
	@echo "  make debug-{service}    - 调试模式运行 (Delve, port 2345)"
	@echo ""
	@echo "工具命令:"
	@echo "  make clean              - 清理所有构建产物"
	@echo "  make new SERVICE=xxx    - 创建新服务 (基于模板)"
	@echo ""
	@echo "可用服务: $(SERVICES)"
	@echo ""

# ================================================================
# 初始化和依赖管理
# ================================================================

.PHONY: init
# 安装 Kratos v2 所需的所有 protoc 插件和工具
init:
	@echo ">>> 安装 Kratos v2 工具链..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "✓ 工具安装完成!"

.PHONY: download
# 下载依赖
download:
	@echo ">>> 下载 Go 依赖..."
	go mod tidy
	@echo "✓ 依赖下载完成!"

# ================================================================
# Proto 代码生成 (大仓模式: 统一在 api/ 目录生成)
# ================================================================

.PHONY: proto
# 生成所有服务的 API proto 代码 (gRPC, HTTP) 并同步生成聚合 OpenAPI
proto:
	@echo ">>> 生成所有服务的 proto 代码..."
	@for svc in $(API_SERVICES); do \
		echo "  - 生成 $$svc proto..."; \
		if [ -d api/$$svc/v1 ]; then \
			protoc \
				--proto_path=./api \
				--proto_path=./third_party \
				--go_out=paths=source_relative:./api \
				--go-http_out=paths=source_relative:./api \
				--go-grpc_out=paths=source_relative:./api \
				api/$$svc/v1/*.proto; \
		else \
			echo "    [跳过] api/$$svc/v1 不存在"; \
		fi; \
	done
	@$(MAKE) --no-print-directory openapi
	@echo "✓ Proto 代码生成完成!"

.PHONY: openapi
# 生成聚合 OpenAPI 文档 (供 Apifox 等平台导入)
openapi:
	@echo ">>> 生成聚合 OpenAPI 文档..."
	@protoc \
		--proto_path=./api \
		--proto_path=./third_party \
		--openapi_out=fq_schema_naming=true,default_response=false:. \
		$(API_PROTO_FILES)
	@echo "✓ OpenAPI 文档生成完成: openapi.yaml"

.PHONY: proto-all
# 生成所有协议代码 (当前仅 Go)
proto-all: proto
	@echo "✓ Proto 代码生成完成!"

.PHONY: errors
# 生成 Kratos errors 定义
errors:
	@echo ">>> 生成 Kratos errors..."
	@for svc in $(SERVICES); do \
		if [ -d api/$$svc/v1 ]; then \
			protoc \
				--proto_path=./api \
				--proto_path=./third_party \
				--go_out=paths=source_relative:./api \
				--go-errors_out=paths=source_relative:./api \
				api/$$svc/v1/*.proto; \
		fi; \
	done
	@echo "✓ Errors 生成完成!"

.PHONY: validate
# 生成 protoc-gen-validate 校验代码
validate:
	@echo ">>> 生成 validate 代码..."
	@for svc in $(SERVICES); do \
		if [ -d api/$$svc/v1 ]; then \
			protoc \
				--proto_path=./api \
				--proto_path=./third_party \
				--go_out=paths=source_relative:./api \
				--validate_out=paths=source_relative,lang=go:./api \
				api/$$svc/v1/*.proto; \
		fi; \
	done
	@echo "✓ Validate 代码生成完成!"

# ================================================================
# Wire 依赖注入代码生成
# ================================================================

.PHONY: generate
# 为所有服务生成 wire 依赖注入代码
generate:
	@echo ">>> 生成所有服务的 wire 代码..."
	@for svc in $(SERVICES); do \
		if [ -f app/$$svc/cmd/$$svc/wire.go ]; then \
			echo "  - 生成 $$svc wire..."; \
			cd app/$$svc/cmd/$$svc && wire && cd ../../../..; \
		else \
			echo "  - [跳过] app/$$svc/cmd/$$svc/wire.go 不存在"; \
		fi; \
	done
	@make download
	@echo "✓ Wire 代码生成完成!"

.PHONY: setup
# 完整设置: proto + openapi + errors + validate + wire
setup:
	@echo ">>> 执行完整设置流程..."
	@make proto
	@make errors
	@make validate
	@make generate
	@echo "✓ 完整设置完成!"

# ================================================================
# 构建和测试
# ================================================================

.PHONY: build
# 构建所有服务
build:
	@echo ">>> 构建所有服务..."
	@mkdir -p bin
	@for svc in $(SERVICES); do \
		if [ -d app/$$svc/cmd/$$svc ]; then \
			echo "  - 构建 $$svc..."; \
			go build -ldflags "-X main.Version=$(VERSION)" -o bin/$$svc ./app/$$svc/cmd/$$svc; \
		else \
			echo "  - [跳过] app/$$svc/cmd/$$svc 不存在"; \
		fi; \
	done
	@echo "✓ 构建完成! 二进制文件在 ./bin/"

.PHONY: test
# 运行所有单元测试
test:
	@echo ">>> 运行单元测试..."
	go test -v -race ./...

.PHONY: test-coverage
# 运行测试并生成覆盖率报告
test-coverage:
	@echo ">>> 运行测试并生成覆盖率报告..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ 覆盖率报告生成: coverage.html"

.PHONY: lint
# 运行 golangci-lint
lint:
	@echo ">>> 运行代码检查..."
	golangci-lint run ./...

# ================================================================
# 单服务操作
# ================================================================

# 生成指定服务的 proto
.PHONY: $(addprefix proto-,$(SERVICES))
$(addprefix proto-,$(SERVICES)):
	@svc=$(@:proto-%=%); \
	echo ">>> 生成 $$svc proto..."; \
	if [ -d api/$$svc/v1 ]; then \
		protoc \
			--proto_path=./api \
			--proto_path=./third_party \
			--go_out=paths=source_relative:./api \
			--go-http_out=paths=source_relative:./api \
			--go-grpc_out=paths=source_relative:./api \
			api/$$svc/v1/*.proto; \
		$(MAKE) --no-print-directory openapi; \
		echo "✓ $$svc proto 生成完成!"; \
	else \
		echo "✗ api/$$svc/v1 不存在"; \
		exit 1; \
	fi

# 构建指定服务
.PHONY: $(addprefix build-,$(SERVICES))
$(addprefix build-,$(SERVICES)):
	@svc=$(@:build-%=%); \
	echo ">>> 构建 $$svc..."; \
	if [ -d app/$$svc/cmd/$$svc ]; then \
		mkdir -p bin; \
		go build -ldflags "-X main.Version=$(VERSION)" -o bin/$$svc ./app/$$svc/cmd/$$svc; \
		echo "✓ $$svc 构建完成: bin/$$svc"; \
	else \
		echo "✗ app/$$svc/cmd/$$svc 不存在"; \
		exit 1; \
	fi

# 运行指定服务
.PHONY: $(addprefix run-,$(SERVICES))
$(addprefix run-,$(SERVICES)):
	@svc=$(@:run-%=%); \
	echo ">>> 运行 $$svc..."; \
	if [ -d app/$$svc/cmd/$$svc ]; then \
		cd app/$$svc && go run ./cmd/$$svc -conf ./configs; \
	else \
		echo "✗ app/$$svc/cmd/$$svc 不存在"; \
		exit 1; \
	fi

# 生成指定服务的 wire 代码
.PHONY: $(addprefix wire-,$(SERVICES))
$(addprefix wire-,$(SERVICES)):
	@svc=$(@:wire-%=%); \
	echo ">>> 生成 $$svc wire..."; \
	if [ -f app/$$svc/cmd/$$svc/wire.go ]; then \
		cd app/$$svc/cmd/$$svc && wire && cd ../../../..; \
		echo "✓ $$svc wire 生成完成!"; \
	else \
		echo "✗ app/$$svc/cmd/$$svc/wire.go 不存在"; \
		exit 1; \
	fi

# 测试指定服务
.PHONY: $(addprefix test-,$(SERVICES))
$(addprefix test-,$(SERVICES)):
	@svc=$(@:test-%=%); \
	echo ">>> 测试 $$svc..."; \
	if [ -d app/$$svc ]; then \
		go test -v -race ./app/$$svc/...; \
		echo "✓ $$svc 测试完成!"; \
	else \
		echo "✗ app/$$svc 不存在"; \
		exit 1; \
	fi

# 测试指定服务并生成覆盖率
.PHONY: $(addprefix test-coverage-,$(SERVICES))
$(addprefix test-coverage-,$(SERVICES)):
	@svc=$(@:test-coverage-%=%); \
	echo ">>> 测试 $$svc 并生成覆盖率..."; \
	if [ -d app/$$svc ]; then \
		go test -v -race -coverprofile=app/$$svc/coverage.out -covermode=atomic ./app/$$svc/...; \
		go tool cover -html=app/$$svc/coverage.out -o app/$$svc/coverage.html; \
		echo "✓ $$svc 覆盖率报告: app/$$svc/coverage.html"; \
	else \
		echo "✗ app/$$svc 不存在"; \
		exit 1; \
	fi

# 调试模式运行指定服务 (使用 delve)
.PHONY: $(addprefix debug-,$(SERVICES))
$(addprefix debug-,$(SERVICES)):
	@svc=$(@:debug-%=%); \
	echo ">>> 调试 $$svc (Delve headless mode, port 2345)..."; \
	echo "    连接方式: dlv connect :2345"; \
	echo "    或使用 IDE 远程调试连接 localhost:2345"; \
	if [ -d app/$$svc/cmd/$$svc ]; then \
		cd app/$$svc && dlv debug ./cmd/$$svc --headless --listen=:2345 --api-version=2 --accept-multiclient -- -conf ./configs; \
	else \
		echo "✗ app/$$svc/cmd/$$svc 不存在"; \
		exit 1; \
	fi

# ================================================================
# 工具命令
# ================================================================

.PHONY: clean
# 清理构建产物
clean:
	@echo ">>> 清理构建产物..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@for svc in $(SERVICES); do \
		rm -f app/$$svc/cmd/$$svc/wire_gen.go; \
	done
	@echo "✓ 清理完成!"

.PHONY: new
# 创建新服务 (用法: make new SERVICE=your-service-name)
new:
ifndef SERVICE
	@echo "✗ 错误: 请指定服务名称"
	@echo "用法: make new SERVICE=your-service-name"
	@exit 1
endif
	@echo ">>> 创建新服务: $(SERVICE)..."
	@if [ -d app/$(SERVICE) ]; then \
		echo "✗ 服务已存在: app/$(SERVICE)"; \
		exit 1; \
	fi
	@echo "  - 创建服务目录结构..."
	@mkdir -p app/$(SERVICE)/cmd/$(SERVICE)
	@mkdir -p app/$(SERVICE)/internal/conf
	@mkdir -p app/$(SERVICE)/internal/data
	@mkdir -p app/$(SERVICE)/internal/biz
	@mkdir -p app/$(SERVICE)/internal/service
	@mkdir -p app/$(SERVICE)/internal/server
	@mkdir -p app/$(SERVICE)/configs
	@mkdir -p api/$(SERVICE)/v1
	@echo "  - 生成服务代码模板..."
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/cmd/main.go.tpl > app/$(SERVICE)/cmd/$(SERVICE)/main.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/cmd/wire.go.tpl > app/$(SERVICE)/cmd/$(SERVICE)/wire.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/server/server.go.tpl > app/$(SERVICE)/internal/server/server.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/server/grpc.go.tpl > app/$(SERVICE)/internal/server/grpc.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/server/http.go.tpl > app/$(SERVICE)/internal/server/http.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/service/service.go.tpl > app/$(SERVICE)/internal/service/service.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/biz/biz.go.tpl > app/$(SERVICE)/internal/biz/biz.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/data/data.go.tpl > app/$(SERVICE)/internal/data/data.go
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/internal/conf/conf.proto.tpl > app/$(SERVICE)/internal/conf/conf.proto
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/configs/config.yaml.tpl > app/$(SERVICE)/configs/config.yaml
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/api/service.proto.tpl > api/$(SERVICE)/v1/$(SERVICE).proto
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/Makefile.tpl > app/$(SERVICE)/Makefile
	@sed 's/{{SERVICE}}/$(SERVICE)/g' templates/Dockerfile.tpl > app/$(SERVICE)/Dockerfile
	@echo "✓ 新服务创建完成: app/$(SERVICE)"
	@echo ""
	@echo "后续步骤:"
	@echo "  0. 将 $(SERVICE) 添加到 Makefile 的 SERVICES 列表中"
	@echo "  1. 编辑 api/$(SERVICE)/v1/$(SERVICE).proto 完善 API 定义"
	@echo "  2. 运行 make proto-$(SERVICE) 生成 proto 代码"
	@echo "  3. 运行 cd app/$(SERVICE) && make config 生成配置 proto"
	@echo "  4. 实现 internal/ 下的业务逻辑"
	@echo "  5. 运行 make wire-$(SERVICE) 生成依赖注入代码"
	@echo "  6. 运行 make build-$(SERVICE) 构建服务"
	@echo "  7. 运行 make run-$(SERVICE) 启动服务"

# ================================================================
# 快捷命令
# ================================================================

.PHONY: all
# 初始化并完整设置
all:
	@make init
	@make setup

.PHONY: rebuild
# 重新构建所有服务
rebuild:
	@make clean
	@make build

.DEFAULT_GOAL := help
