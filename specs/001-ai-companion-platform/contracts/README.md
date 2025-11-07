# API Contracts

本目录包含AI陪伴精灵平台的所有gRPC API定义(Protobuf格式)。

## Overview

- **user_service.proto**: 用户服务 - 认证、授权、画像管理
- **conversation_service.proto**: 对话服务 - 会话管理、消息路由
- **memory_service.proto**: 记忆服务 - 记忆存储与检索
- **character_service.proto**: 角色服务 - AI角色CRUD (待生成)
- **billing_service.proto**: 计费服务 - 积分管理 (待生成)
- **llm_service.proto**: LLM服务 - AI推理与Agent (待生成)

## Code Generation

### Golang

```bash
# 安装protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       user_service.proto
```

### Python

```bash
# 安装grpcio-tools
pip install grpcio-tools

# 生成代码
python -m grpc_tools.protoc -I. \
       --python_out=. \
       --grpc_python_out=. \
       user_service.proto
```

## API Design Principles

1. **统一错误处理**: 使用gRPC标准错误码
2. **幂等性**: 关键操作(发送消息、扣除积分)使用idempotency_key
3. **分页**: 使用cursor-based pagination
4. **流式响应**: 长时间操作(LLM生成)使用stream
5. **版本管理**: 包名包含版本号(v1, v2...)

## Error Codes

| gRPC Code | HTTP Status | 说明 |
|-----------|-------------|------|
| OK | 200 | 成功 |
| INVALID_ARGUMENT | 400 | 请求参数错误 |
| UNAUTHENTICATED | 401 | 未认证 |
| PERMISSION_DENIED | 403 | 无权限 |
| NOT_FOUND | 404 | 资源不存在 |
| ALREADY_EXISTS | 409 | 资源已存在 |
| RESOURCE_EXHAUSTED | 429 | 请求过多 |
| INTERNAL | 500 | 服务器内部错误 |
| UNAVAILABLE | 503 | 服务不可用 |

## Authentication

所有API调用需要在metadata中携带JWT token:

```go
// Golang客户端
md := metadata.Pairs("authorization", "Bearer "+token)
ctx := metadata.NewOutgoingContext(context.Background(), md)
response, err := client.GetUser(ctx, &GetUserRequest{...})
```

```python
# Python客户端
metadata = [('authorization', f'Bearer {token}')]
response = stub.GetUser(GetUserRequest(...), metadata=metadata)
```

## Next Steps

1. 生成其他服务的Proto文件(character, billing, llm)
2. 实现服务端和客户端代码
3. 编写API集成测试
4. 生成OpenAPI文档(通过grpc-gateway)
