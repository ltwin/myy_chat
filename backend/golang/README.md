# MyY Chat Backend 构建指南

本项目采用 **Kratos v2 大仓模式 (Monorepo)**，所有 Go 微服务共享一个 `go.mod` 和统一的 `api/` 目录。

## 目录结构

```
backend/golang/
├── Makefile              # 根 Makefile (统一管理所有服务)
├── go.mod                # 共享的 Go 模块
├── api/                  # 统一的 API proto 定义
│   ├── user/v1/
│   ├── character/v1/
│   ├── conversation/v1/
│   ├── memory/v1/
│   ├── billing/v1/
│   ├── admin/v1/
│   └── analytics/v1/
├── app/                  # 各服务实现
│   ├── user/
│   ├── character/
│   ├── conversation/
│   ├── memory/
│   ├── billing/
│   ├── admin/
│   └── analytics/
├── pkg/                  # 共享包
└── third_party/          # 第三方 proto 依赖
```

## 快速开始

### 1. 初始化开发环境

```bash
# 安装所有 Kratos v2 工具链
make init

# 下载 Go 依赖
make download
```

### 2. 生成代码

```bash
# 生成所有服务的 proto 代码 + openapi + errors + validate + wire
make setup

# 或分步执行
make proto      # 生成 proto 代码（并自动生成 openapi.yaml）
make openapi    # 单独生成聚合 OpenAPI 文档（openapi.yaml）
make errors     # 生成 error 定义
make validate   # 生成 validate 校验
make generate   # 生成 wire 依赖注入
```

### 3. 构建和运行

```bash
# 构建所有服务
make build

# 构建单个服务
make build-user
make build-admin

# 运行单个服务
make run-user
make run-admin
```

## 常用命令

### 全局操作 (根目录)

| 命令 | 说明 |
|------|------|
| `make help` | 显示所有可用命令 |
| `make init` | 安装 protoc 插件和工具 |
| `make proto` | 生成所有服务的 proto 代码（并自动生成 `openapi.yaml`） |
| `make openapi` | 单独生成聚合 OpenAPI 文档 (`openapi.yaml`) |
| `make setup` | 完整设置 (proto + openapi + errors + validate + wire) |
| `make build` | 构建所有服务到 `bin/` 目录 |
| `make test` | 运行所有单元测试 |
| `make test-coverage` | 生成测试覆盖率报告 |
| `make lint` | 运行 golangci-lint 检查 |
| `make clean` | 清理所有构建产物 |

### 单服务操作

| 命令 | 说明 |
|------|------|
| `make proto-{service}` | 生成指定服务的 proto |
| `make build-{service}` | 构建指定服务 |
| `make run-{service}` | 运行指定服务 |
| `make wire-{service}` | 生成指定服务的 wire 代码 |

**示例:**
```bash
make proto-user      # 生成 user 服务的 proto
make build-admin     # 构建 admin 服务
make run-character   # 运行 character 服务
```

### 创建新服务

```bash
# 创建新服务 (自动生成目录结构和模板文件)
make new SERVICE=notification

# 后续步骤:
# 1. 编辑 api/notification/v1/notification.proto 定义 API
# 2. 运行 make proto-notification 生成代码
# 3. 实现 app/notification/internal/ 下的业务逻辑
# 4. 创建 app/notification/cmd/notification/wire.go
# 5. 运行 make wire-notification 生成依赖注入
# 6. 运行 make build-notification 构建服务
```

## 开发工作流

### 新功能开发流程

1. **定义 API**
   ```bash
   # 编辑 proto 文件
   vim api/user/v1/user.proto

   # 生成代码
   make proto-user
   ```

2. **实现业务逻辑**
   ```bash
   # 实现 internal/{data,biz,service} 层
   vim app/user/internal/service/user.go
   ```

3. **生成依赖注入**
   ```bash
   # 编辑 wire.go (如果需要新增依赖)
   vim app/user/cmd/user/wire.go

   # 生成 wire 代码
   make wire-user
   ```

4. **构建和测试**
   ```bash
   # 运行测试
   make test

   # 构建服务
   make build-user

   # 运行服务
   make run-user
   ```

### 单服务开发 (在服务目录内)

如果你只想专注于某个服务的开发，可以在服务目录内使用服务级 Makefile：

```bash
cd app/admin

# 查看服务级命令
make help

# 生成 internal proto 配置 (如果有)
make config

# 生成 wire 代码
make generate

# 构建当前服务
make build

# 运行当前服务
make run

# 开发模式 (一键执行: config + generate + build + run)
make dev
```

## Proto 代码生成规则

### API Proto (api/*/v1/*.proto)

- **位置**: `api/{service}/v1/*.proto`
- **生成目标**: 同目录 (`source_relative` 模式)
- **生成内容**:
  - `*.pb.go` - Protocol Buffers 消息定义
  - `*_grpc.pb.go` - gRPC 服务定义
  - `*_http.pb.go` - Kratos HTTP 路由 (可选)
  - `*.errors.pb.go` - Kratos 错误定义 (可选)
  - `openapi.yaml` - 聚合 OpenAPI 3.0 文档（根目录，可导入 Apifox）

### Internal Proto (app/*/internal/conf/*.proto)

- **位置**: `app/{service}/internal/conf/*.proto`
- **生成目标**: 同目录 (`source_relative` 模式)
- **用途**: 服务私有的配置定义
- **生成**: 在服务目录内运行 `make config`

## 大仓模式 vs 分仓模式

| 特性 | 大仓模式 (当前) | 分仓模式 (buzzy-workflow) |
|------|----------------|--------------------------|
| **go.mod** | 共享一个 | 每服务一个 |
| **api/ 目录** | 根目录统一管理 | 每服务独立 |
| **third_party/** | 根目录共享 | 每服务独立 |
| **proto 生成** | 根 Makefile 统一生成 | 服务 Makefile 各自生成 |
| **依赖管理** | 统一版本 | 各自管理 |
| **适用场景** | 7+ 个服务，共享大量代码 | 独立部署，版本独立 |

## 常见问题

### Q: 为什么 proto 文件生成在同目录？

**A:** 大仓模式使用 `source_relative` 生成模式，proto 和生成的 Go 代码在同一目录，便于统一管理和版本控制。

### Q: 如何添加新的共享依赖？

**A:** 在根目录执行 `go get <package>` 和 `make download`，所有服务自动共享。

### Q: wire.go 应该放在哪里？

**A:** 每个服务的 `app/{service}/cmd/{service}/wire.go`，用于定义该服务的依赖注入。

### Q: 如何调试单个服务？

**A:**
```bash
# 方式 1: 使用根 Makefile
make run-user

# 方式 2: 进入服务目录
cd app/user
make run

# 方式 3: 直接运行
cd app/user
go run ./cmd/user -conf ./configs
```

### Q: proto 修改后需要重新生成什么？

**A:**
```bash
# 如果修改了 api/*.proto
make proto-{service}  # 重新生成 proto 代码

# 如果修改了 wire.go
make wire-{service}   # 重新生成依赖注入

# 完整重新生成
make setup
```

## 技术栈

- **框架**: Kratos v2.8
- **RPC**: gRPC + Protobuf
- **依赖注入**: Google Wire
- **API 风格**: gRPC + HTTP (通过 go-http 插件)
- **配置**: YAML + Proto
- **代码检查**: golangci-lint
- **测试**: Go test + race detector

## 参考资源

- [Kratos 官方文档](https://go-kratos.dev/)
- [Google Wire 文档](https://github.com/google/wire)
- [Protocol Buffers 文档](https://protobuf.dev/)
- [gRPC 文档](https://grpc.io/docs/languages/go/)

## 下一步

1. 查看 [开发文档](../../docs/development.md) 了解详细开发规范
2. 阅读 [项目规范](../../specs/001-ai-companion-platform/) 了解业务需求
3. 开始实现你的第一个服务！
