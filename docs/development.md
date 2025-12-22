# MyY Chat 开发规范文档

> 本文档定义了 MyY Chat AI角色对话平台的开发标准、代码规范和最佳实践

**版本**: 1.0
**更新日期**: 2025-11-08
**适用范围**: 所有后端(Golang/Python)和前端(TypeScript)代码

---

## 1. 代码注释规范

### 1.1 中文注释原则

**核心要求**: 所有代码注释必须使用**中文**，便于团队沟通和维护

#### Golang 注释规范

```go
// 用户服务 - 负责用户认证、授权和画像管理
type UserService struct {
    repo UserRepository
    jwt  JWTManager
}

// CreateUser 创建新用户
// 参数:
//   - ctx: 上下文对象，用于超时控制和取消
//   - req: 创建用户请求，包含邮箱和密码
// 返回:
//   - *User: 创建成功的用户对象
//   - error: 创建失败时返回错误
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // 验证邮箱格式
    if !isValidEmail(req.Email) {
        return nil, errors.New("邮箱格式不正确")
    }

    // 检查邮箱是否已注册
    exists, err := s.repo.ExistsByEmail(ctx, req.Email)
    if err != nil {
        return nil, fmt.Errorf("检查邮箱失败: %w", err)
    }
    if exists {
        return nil, errors.New("邮箱已被注册")
    }

    // 使用 bcrypt 哈希密码 (cost=12)
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(req.Password),
        12, // bcrypt cost factor
    )
    if err != nil {
        return nil, fmt.Errorf("密码哈希失败: %w", err)
    }

    // 生成 Snowflake ID
    userID := s.idGenerator.NextID()

    // 创建用户对象
    user := &User{
        ID:           userID,
        Email:        req.Email,
        PasswordHash: string(hashedPassword),
        CreatedAt:    time.Now(),
    }

    // 保存到数据库
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("保存用户失败: %w", err)
    }

    return user, nil
}
```

**注释要点**:
- 包/类型/函数的顶层注释使用 `//` 开头
- 复杂逻辑前添加注释说明
- 函数注释包含：用途、参数、返回值、错误处理
- 关键算法和业务规则必须注释

#### Python 注释规范

```python
class MemoryProcessor:
    """记忆处理服务 - 负责记忆提取、Embedding生成和矛盾检测

    Attributes:
        embedding_client: SiliconFlow Embedding API 客户端
        llm_client: LLM 客户端，用于记忆提取
    """

    def __init__(self, embedding_client: EmbeddingClient, llm_client: LLMClient):
        self.embedding_client = embedding_client
        self.llm_client = llm_client

    async def extract_memories(
        self,
        conversation: Conversation,
        user_id: int,
        character_id: int
    ) -> list[Memory]:
        """从对话中提取用户记忆

        Args:
            conversation: 对话对象，包含用户和AI的消息
            user_id: 用户ID (Snowflake ID)
            character_id: 角色ID (Snowflake ID)

        Returns:
            提取的记忆列表，每条记忆包含内容和重要性评分

        Raises:
            LLMError: LLM 调用失败
            EmbeddingError: Embedding 生成失败
        """
        # 构建提取记忆的 prompt
        prompt = self._build_extraction_prompt(conversation)

        # 调用 LLM 提取结构化记忆
        response = await self.llm_client.complete(prompt)
        memories = self._parse_memory_response(response)

        # 为每条记忆生成 Embedding
        for memory in memories:
            embedding = await self.embedding_client.embed(memory.content)
            memory.embedding = embedding

        return memories
```

**注释要点**:
- 类使用 docstring (三引号) 说明用途和属性
- 函数使用 Google 风格 docstring (Args, Returns, Raises)
- 复杂逻辑用行内注释说明
- 类型提示 (type hints) 必须完整

#### TypeScript 注释规范

```typescript
/**
 * 用户认证服务
 * 负责登录、注册、Token 管理
 */
export class AuthService {
  /**
   * 用户登录
   * @param email - 用户邮箱
   * @param password - 用户密码
   * @returns 包含访问令牌和刷新令牌的认证响应
   * @throws {AuthError} 登录失败时抛出错误
   */
  async login(email: string, password: string): Promise<AuthResponse> {
    // 验证邮箱格式
    if (!this.isValidEmail(email)) {
      throw new AuthError('邮箱格式不正确');
    }

    // 调用后端 API 登录
    const response = await this.apiClient.post('/auth/login', {
      email,
      password,
    });

    // 保存 Token 到本地存储
    this.saveTokens(response.data);

    return response.data;
  }
}
```

**注释要点**:
- 使用 JSDoc 风格注释
- 类和方法必须注释
- 参数用 `@param`，返回值用 `@returns`
- 异常用 `@throws`

---

## 2. 日志规范

### 2.1 英文日志原则

**核心要求**: 所有日志输出必须使用**英文**，便于国际化和日志分析工具处理

#### Golang 日志规范

```go
package main

import (
    "go.uber.org/zap"
)

var logger *zap.Logger

func init() {
    // 初始化结构化日志
    logger, _ = zap.NewProduction()
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // ✅ 正确：使用英文 + 结构化字段
    logger.Info("creating new user",
        zap.String("email", req.Email),
        zap.String("request_id", getRequestID(ctx)),
    )

    user, err := s.repo.Create(ctx, req)
    if err != nil {
        // ✅ 正确：错误日志包含上下文
        logger.Error("failed to create user",
            zap.String("email", req.Email),
            zap.Error(err),
        )
        return nil, err
    }

    // ✅ 正确：成功日志包含关键信息
    logger.Info("user created successfully",
        zap.Int64("user_id", user.ID),
        zap.String("email", user.Email),
    )

    return user, nil
}

// ❌ 错误示例：不要使用中文日志
func badExample() {
    logger.Info("用户创建成功") // 错误：使用了中文
}
```

**日志级别**:
- `DEBUG`: 调试信息 (开发环境)
- `INFO`: 正常流程信息
- `WARN`: 警告，可能的问题
- `ERROR`: 错误，需要关注
- `FATAL`: 致命错误，程序退出

#### Python 日志规范

```python
import structlog

logger = structlog.get_logger()

class LLMService:
    async def generate_response(
        self,
        prompt: str,
        user_id: int,
        character_id: int
    ) -> str:
        # ✅ 正确：结构化日志 + 英文
        logger.info(
            "generating AI response",
            user_id=user_id,
            character_id=character_id,
            prompt_length=len(prompt)
        )

        try:
            response = await self.llm_client.complete(prompt)

            logger.info(
                "AI response generated successfully",
                user_id=user_id,
                response_length=len(response),
                tokens_used=response.usage.total_tokens
            )

            return response.content

        except Exception as e:
            # ✅ 正确：错误日志包含异常信息
            logger.error(
                "failed to generate AI response",
                user_id=user_id,
                error=str(e),
                exc_info=True  # 包含堆栈跟踪
            )
            raise
```

**日志配置** (pyproject.toml):
```toml
[tool.structlog]
# 生产环境使用 JSON 格式
# 开发环境使用彩色控制台输出
```

#### TypeScript 日志规范

```typescript
import { logger } from '@/lib/logger';

export class ConversationService {
  async sendMessage(
    conversationId: string,
    message: string
  ): Promise<Message> {
    // ✅ 正确：英文日志 + 结构化数据
    logger.info('sending message', {
      conversation_id: conversationId,
      message_length: message.length,
    });

    try {
      const response = await this.apiClient.post('/messages', {
        conversation_id: conversationId,
        content: message,
      });

      logger.info('message sent successfully', {
        conversation_id: conversationId,
        message_id: response.data.id,
      });

      return response.data;
    } catch (error) {
      // ✅ 正确：错误日志包含详细信息
      logger.error('failed to send message', {
        conversation_id: conversationId,
        error: error.message,
      });
      throw error;
    }
  }
}
```

---

## 3. 命名规范

### 3.1 Golang 命名

- **包名**: 小写，单词，简短 (`user`, `biz`, `data`)
- **文件名**: 小写+下划线 (`user_service.go`, `user_repository.go`)
- **类型名**: PascalCase (`UserService`, `ConversationRepository`)
- **函数名**: PascalCase (公开), camelCase (私有)
- **常量**: PascalCase (`MaxRetryCount`, `DefaultTimeout`)
- **变量**: camelCase (`userID`, `requestContext`)

```go
// ✅ 正确示例
const (
    MaxRetryCount = 3
    DefaultTimeout = 10 * time.Second
)

type UserService struct {
    repo UserRepository
}

func (s *UserService) CreateUser(ctx context.Context) error {
    var userID int64 // camelCase
    return nil
}
```

### 3.2 Python 命名

- **模块名**: 小写+下划线 (`user_service.py`, `llm_client.py`)
- **类名**: PascalCase (`MemoryProcessor`, `EmbeddingClient`)
- **函数名**: snake_case (`extract_memories`, `generate_embedding`)
- **常量**: UPPER_SNAKE_CASE (`MAX_RETRY_COUNT`, `DEFAULT_TIMEOUT`)
- **变量**: snake_case (`user_id`, `conversation_id`)

```python
# ✅ 正确示例
MAX_RETRY_COUNT = 3
DEFAULT_TIMEOUT = 10

class UserService:
    def create_user(self, email: str, password: str) -> User:
        user_id = generate_snowflake_id()
        return User(id=user_id, email=email)
```

### 3.3 TypeScript 命名

- **文件名**: kebab-case (`user-service.ts`, `api-client.ts`)
- **类名**: PascalCase (`AuthService`, `ApiClient`)
- **函数名**: camelCase (`createUser`, `sendMessage`)
- **常量**: UPPER_SNAKE_CASE (`MAX_RETRY_COUNT`, `API_BASE_URL`)
- **接口**: PascalCase + `I` 前缀 (可选) (`IUser`, `IConversation`)
- **类型**: PascalCase (`User`, `Message`)

```typescript
// ✅ 正确示例
const MAX_RETRY_COUNT = 3;

interface User {
  id: string;
  email: string;
}

class AuthService {
  async createUser(email: string, password: string): Promise<User> {
    const userId = generateId();
    return { id: userId, email };
  }
}
```

---

## 4. 测试规范

### 4.1 测试覆盖率要求

根据 **MyY Chat Constitution**:
- **最低要求**: ≥70% 代码覆盖率
- **推荐目标**: ≥80% 代码覆盖率

### 4.2 Golang 测试

**文件命名**: `*_test.go`

```go
// user_service_test.go
package biz

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/mock/gomock"
)

func TestUserService_CreateUser(t *testing.T) {
    // 使用表驱动测试
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {
            name:    "有效邮箱",
            email:   "user@example.com",
            wantErr: false,
        },
        {
            name:    "无效邮箱",
            email:   "invalid",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockRepo := NewMockUserRepository(ctrl)
            service := NewUserService(mockRepo)

            // Act
            _, err := service.CreateUser(context.Background(), tt.email)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**运行测试**:
```bash
# 运行所有测试并显示覆盖率
go test -v -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 4.3 Python 测试

**文件命名**: `test_*.py` 或 `*_test.py`

```python
# test_memory_processor.py
import pytest
from unittest.mock import AsyncMock, MagicMock

from app.services.memory_processor import MemoryProcessor

@pytest.mark.asyncio
async def test_extract_memories_success():
    """测试成功提取记忆"""
    # Arrange
    embedding_client = AsyncMock()
    llm_client = AsyncMock()
    processor = MemoryProcessor(embedding_client, llm_client)

    conversation = create_mock_conversation()

    # Act
    memories = await processor.extract_memories(
        conversation,
        user_id=123,
        character_id=456
    )

    # Assert
    assert len(memories) > 0
    assert memories[0].embedding is not None
    llm_client.complete.assert_called_once()
```

**运行测试**:
```bash
# 运行所有测试并显示覆盖率
pytest --cov=. --cov-report=term --cov-report=html

# 检查覆盖率是否达标 (≥70%)
pytest --cov=. --cov-fail-under=70
```

### 4.4 TypeScript 测试

使用 **Vitest** 或 **Jest**:

```typescript
// user-service.test.ts
import { describe, it, expect, vi } from 'vitest';
import { AuthService } from './auth-service';

describe('AuthService', () => {
  it('should login successfully with valid credentials', async () => {
    // Arrange
    const apiClient = {
      post: vi.fn().mockResolvedValue({
        data: { access_token: 'token123' },
      }),
    };
    const service = new AuthService(apiClient);

    // Act
    const result = await service.login('user@example.com', 'password');

    // Assert
    expect(result.access_token).toBe('token123');
    expect(apiClient.post).toHaveBeenCalledWith('/auth/login', {
      email: 'user@example.com',
      password: 'password',
    });
  });
});
```

---

## 5. 错误处理

### 5.1 Golang 错误处理

```go
import "fmt"

// ✅ 正确：使用 fmt.Errorf 包装错误
func (s *UserService) GetUser(ctx context.Context, id int64) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("find user by id %d: %w", id, err)
    }
    return user, nil
}

// ✅ 正确：定义自定义错误类型
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
)

// ❌ 错误：直接返回底层错误
func badExample() error {
    _, err := someFunction()
    return err // 缺少上下文
}
```

### 5.2 Python 错误处理

```python
class UserNotFoundError(Exception):
    """用户未找到错误"""
    pass

class UserService:
    def get_user(self, user_id: int) -> User:
        # ✅ 正确：使用自定义异常
        user = self.repo.find_by_id(user_id)
        if not user:
            raise UserNotFoundError(f"User {user_id} not found")
        return user
```

---

## 6. 代码审查清单

提交 Pull Request 前自查：

- [ ] 所有注释使用中文
- [ ] 所有日志使用英文
- [ ] 测试覆盖率 ≥70%
- [ ] 通过 `golangci-lint` / `ruff` / `eslint` 检查
- [ ] 通过 `go test` / `pytest` / `npm test`
- [ ] 更新相关文档
- [ ] 遵循命名规范
- [ ] 错误处理完整

---

## 7. 工具链

### 7.1 代码检查

```bash
# Golang
golangci-lint run --config .golangci.yml

# Python
ruff check .
black --check .
mypy .

# TypeScript
npm run lint
npm run type-check
```

### 7.2 格式化

```bash
# Golang
gofmt -s -w .
goimports -w .

# Python
black .
ruff format .

# TypeScript
npm run format
```

---

## 8. 参考资料

- [Effective Go](https://go.dev/doc/effective_go)
- [PEP 8 - Python 代码风格](https://pep8.org/)
- [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html)
- [MyY Chat Constitution](../.specify/memory/constitution.md)

---

**最后更新**: 2025-11-08
**维护者**: MyY Chat 开发团队
