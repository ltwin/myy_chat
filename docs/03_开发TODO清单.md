# AI聊天服务器 - 开发TODO清单

**项目名称**: AI Chat Server with Layered Memory
**开发模式**: 敏捷迭代，分阶段交付
**预计总工期**: 8周（单人）

---

## 📋 总览

### 开发阶段规划

```
Week 1-2  ███████████░░░░░░░░░░░░░░░  阶段1: MVP
Week 3-4  ░░░░░░░░░░░███████████░░░░  阶段2: 记忆系统
Week 5-6  ░░░░░░░░░░░░░░░░░░░░███████  阶段3: 高级功能
Week 7-8  ░░░░░░░░░░░░░░░░░░░░░░░░███  阶段4: 优化和发布
```

### 技术栈检查清单

**在开始之前，请确保安装**:

```bash
# Golang环境
□ Go 1.21+
□ Kratos CLI: go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
□ Wire: go install github.com/google/wire/cmd/wire@latest
□ Protoc + Go插件

# Python环境
□ Python 3.11+
□ Poetry或pip-tools
□ grpcio-tools

# 数据库
□ PostgreSQL 15+
□ pgvector扩展
□ Redis 7+

# 前端环境
□ Node.js 18+
□ pnpm或npm

# 开发工具
□ Docker + Docker Compose
□ Git
□ IDE (VSCode / GoLand / PyCharm)
```

---

## 🚀 阶段1: MVP (第1-2周)

**目标**: 完成最小可用系统，可以注册、创建角色、对话

### 第1周任务

#### Day 1: 项目初始化和环境搭建

- [ ] **TASK-0101**: 创建项目目录结构 (30分钟)
  ```bash
  ai-chat-server/
  ├── services/
  │   ├── user-service/      # Golang用户服务
  │   ├── character-service/ # Golang角色服务
  │   ├── chat-service/      # Golang对话服务
  │   ├── memory-service/    # Golang记忆服务
  │   └── llm-service/       # Python LLM服务
  ├── frontend/              # Vue3前端
  ├── api/                   # Proto定义
  ├── deployments/           # 部署配置
  │   └── docker-compose.yml
  └── docs/                  # 文档
  ```

- [ ] **TASK-0102**: 初始化PostgreSQL数据库 (1小时)
  ```bash
  # 1. 创建数据库
  # 2. 执行docs/database_schema.sql
  # 3. 验证所有表创建成功
  # 4. 插入测试数据
  ```
  **验收**: `SELECT count(*) FROM pg_tables WHERE schemaname='public';` 返回12+

- [ ] **TASK-0103**: 配置Redis (30分钟)
  ```bash
  # docker-compose启动Redis
  # 验证连接
  redis-cli ping  # PONG
  ```

- [ ] **TASK-0104**: 创建Golang User Service脚手架 (2小时)
  ```bash
  cd services/user-service
  kratos new . --nomod
  # 配置go.mod
  # 安装依赖
  ```
  **包含**:
  - api/v1/user.proto
  - internal/biz/user.go
  - internal/data/user.go
  - internal/service/user.go

- [ ] **TASK-0105**: 创建Python LLM Service脚手架 (2小时)
  ```bash
  cd services/llm-service
  poetry init
  poetry add fastapi grpcio grpcio-tools langchain openai
  ```
  **包含**:
  - app/api/grpc_server.py
  - app/core/config.py
  - app/services/chat.py
  - proto/llm_service.proto

**Day 1 验收**: 两个服务可以启动，监听端口

---

#### Day 2: 用户服务开发

- [ ] **TASK-0201**: 定义User Service gRPC接口 (1小时)
  ```protobuf
  // api/proto/user.proto
  service UserService {
      rpc Register(RegisterRequest) returns (User);
      rpc Login(LoginRequest) returns (LoginResponse);
      rpc GetProfile(GetProfileRequest) returns (UserProfile);
  }
  ```
  **生成代码**: `make proto-gen`

- [ ] **TASK-0202**: 实现用户注册逻辑 (2小时)
  ```go
  // internal/biz/user.go
  func (uc *UserUseCase) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
      // 1. 验证邮箱格式
      // 2. 检查邮箱是否已存在
      // 3. bcrypt加密密码
      // 4. 创建用户记录
      // 5. 返回用户信息（不含密码）
  }
  ```
  **技术要点**:
  - 使用`golang.org/x/crypto/bcrypt`加密密码
  - 使用`github.com/go-playground/validator`验证输入

- [ ] **TASK-0203**: 实现用户登录逻辑 (2小时)
  ```go
  func (uc *UserUseCase) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
      // 1. 查询用户
      // 2. 验证密码
      // 3. 生成JWT Token (24小时有效期)
      // 4. 更新last_login_at
      // 5. 返回Token和用户信息
  }
  ```
  **技术要点**:
  - 使用`github.com/golang-jwt/jwt`生成JWT
  - JWT Payload: `{"user_id": "uuid", "exp": timestamp}`

- [ ] **TASK-0204**: 实现JWT中间件 (1.5小时)
  ```go
  // internal/middleware/auth.go
  func JWTAuth() middleware.Middleware {
      return func(handler middleware.Handler) middleware.Handler {
          return func(ctx context.Context, req interface{}) (interface{}, error) {
              // 1. 从Header提取Token
              // 2. 验证Token
              // 3. 将user_id注入context
              return handler(ctx, req)
          }
      }
  }
  ```

- [ ] **TASK-0205**: 编写单元测试 (1.5小时)
  ```go
  // internal/biz/user_test.go
  func TestRegister(t *testing.T) {
      // 测试正常注册
      // 测试邮箱重复
      // 测试密码太弱
  }
  ```
  **验收**: `go test ./... -cover` 覆盖率 > 70%

**Day 2 验收**: 用户可以注册、登录，获得JWT Token

---

#### Day 3: 角色服务开发

- [ ] **TASK-0301**: 定义Character Service接口 (1小时)
  ```protobuf
  service CharacterService {
      rpc CreateCharacter(CreateCharacterRequest) returns (Character);
      rpc GetCharacter(GetCharacterRequest) returns (Character);
      rpc ListCharacters(ListCharactersRequest) returns (ListCharactersResponse);
      rpc UpdateCharacter(UpdateCharacterRequest) returns (Character);
      rpc DeleteCharacter(DeleteCharacterRequest) returns (Empty);
  }
  ```

- [ ] **TASK-0302**: 实现角色CRUD (3小时)
  ```go
  // internal/biz/character.go
  func (cc *CharacterUseCase) CreateCharacter(ctx context.Context, req *CreateRequest) (*Character, error) {
      // 1. 验证用户登录
      // 2. 验证必填字段
      // 3. 创建角色记录 (version=1)
      // 4. 创建初始版本快照
      // 5. 返回角色信息
  }
  ```

- [ ] **TASK-0303**: 实现权限控制 (1.5小时)
  ```go
  // 只有创建者可以编辑/删除角色
  // 公开角色所有人可以查看
  // 私有角色只有创建者可以查看
  ```

- [ ] **TASK-0304**: 编写测试 (1.5小时)

**Day 3 验收**: 用户可以创建、查看、编辑角色

---

#### Day 4-5: LLM Service基础功能

- [ ] **TASK-0401**: 实现gRPC Server (2小时)
  ```python
  # app/api/grpc_server.py
  class LLMServiceServicer(llm_pb2_grpc.LLMServiceServicer):
      async def Chat(self, request, context):
          # 调用ChatService
          response = await chat_service.chat(request)
          return response
  ```

- [ ] **TASK-0402**: 实现OpenAI客户端封装 (2小时)
  ```python
  # app/core/llm_client.py
  class OpenAIClient:
      async def chat_completion(self, messages, model="gpt-3.5-turbo", **kwargs):
          response = await openai.ChatCompletion.acreate(
              model=model,
              messages=messages,
              **kwargs
          )
          return response
  ```

- [ ] **TASK-0403**: 实现对话服务 (3小时)
  ```python
  # app/services/chat.py
  async def chat(request: ChatRequest) -> ChatResponse:
      # 1. 构建系统提示词
      system_prompt = build_system_prompt(request.character_settings)

      # 2. 构建消息列表
      messages = [
          {"role": "system", "content": system_prompt},
          *[{"role": msg.role, "content": msg.content} for msg in request.message_history],
          {"role": "user", "content": request.user_message}
      ]

      # 3. 调用LLM
      response = await llm_client.chat_completion(messages)

      # 4. 返回响应
      return ChatResponse(response=response.choices[0].message.content, ...)
  ```

- [ ] **TASK-0404**: 实现Token计数 (1小时)
  ```python
  import tiktoken

  def count_tokens(text: str, model: str = "gpt-3.5-turbo") -> int:
      encoding = tiktoken.encoding_for_model(model)
      return len(encoding.encode(text))
  ```

- [ ] **TASK-0405**: 配置文件和环境变量 (1小时)
  ```python
  # app/core/config.py
  from pydantic_settings import BaseSettings

  class Settings(BaseSettings):
      openai_api_key: str
      openai_api_base: str = "https://api.openai.com/v1"
      default_model: str = "gpt-3.5-turbo"

      class Config:
          env_file = ".env"
  ```

- [ ] **TASK-0406**: 编写Python单元测试 (2小时)
  ```python
  # tests/test_chat.py
  @pytest.mark.asyncio
  async def test_chat():
      request = ChatRequest(
          user_message="你好",
          character_settings=...,
      )
      response = await chat_service.chat(request)
      assert len(response.response) > 0
  ```

**Day 4-5 验收**: Python服务可以调用OpenAI API并返回回复

---

#### Day 6-7: Chat Service和前端MVP

- [ ] **TASK-0601**: 实现Chat Service (4小时)
  ```go
  // services/chat-service/internal/biz/chat.go
  func (cc *ChatUseCase) SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error) {
      // 1. 保存用户消息
      userMsg := cc.data.SaveMessage(ctx, &Message{
          ConversationID: req.ConversationID,
          Role:          "user",
          Content:       req.Content,
      })

      // 2. 加载临时上下文（Redis）
      tempContext := cc.loadTempContext(req.ConversationID)

      // 3. 调用LLM Service (gRPC)
      llmResp, err := cc.llmClient.Chat(ctx, &llm.ChatRequest{
          UserMessage:    req.Content,
          MessageHistory: tempContext.Messages,
          CharacterSettings: cc.getCharacterSettings(req.CharacterID),
      })

      // 4. 保存AI回复
      aiMsg := cc.data.SaveMessage(ctx, &Message{
          ConversationID: req.ConversationID,
          Role:          "assistant",
          Content:       llmResp.Response,
          TokenCount:    llmResp.TokenUsage.TotalTokens,
      })

      // 5. 更新Redis临时上下文
      cc.updateTempContext(req.ConversationID, userMsg, aiMsg)

      return aiMsg, nil
  }
  ```

- [ ] **TASK-0602**: 实现Redis临时上下文管理 (2小时)
  ```go
  // internal/data/context.go
  type TempContext struct {
      Messages   []*Message
      TokenCount int
  }

  func (r *chatRepo) LoadTempContext(conversationID string) *TempContext {
      // 从Redis加载: conversation:{id}:context
      // 反序列化JSON
  }

  func (r *chatRepo) UpdateTempContext(conversationID string, newMsgs ...*Message) {
      // 添加新消息到列表
      // 保留最近10-20条
      // 重新计算Token
      // 存入Redis (TTL 24小时)
  }
  ```

- [ ] **TASK-0603**: 实现对话会话管理 (1.5小时)
  ```go
  func (cc *ChatUseCase) CreateConversation(ctx context.Context, characterID string) (*Conversation, error) {
      // 创建conversation记录
      // 初始化Redis上下文
  }
  ```

- [ ] **TASK-0604**: 前端项目初始化 (2小时)
  ```bash
  cd frontend
  pnpm create vite . --template vue-ts
  pnpm add pinia vue-router naive-ui @vueuse/core axios
  ```

- [ ] **TASK-0605**: 前端登录注册页面 (3小时)
  ```vue
  <!-- src/views/Login.vue -->
  <template>
    <n-form>
      <n-form-item label="邮箱">
        <n-input v-model:value="form.email" />
      </n-form-item>
      <n-form-item label="密码">
        <n-input type="password" v-model:value="form.password" />
      </n-form-item>
      <n-button @click="handleLogin">登录</n-button>
    </n-form>
  </template>
  ```

- [ ] **TASK-0606**: 前端对话界面 (4小时)
  ```vue
  <!-- src/views/Chat.vue -->
  <template>
    <div class="chat-container">
      <div class="messages">
        <div v-for="msg in messages" :key="msg.id" :class="msg.role">
          {{ msg.content }}
        </div>
      </div>
      <div class="input-area">
        <n-input v-model:value="userInput" @keyup.enter="sendMessage" />
        <n-button @click="sendMessage">发送</n-button>
      </div>
    </div>
  </template>
  ```

**Day 6-7 验收**: 完整的MVP流程走通

---

### 第2周任务

#### Day 8-9: 集成测试和Bug修复

- [ ] **TASK-0801**: 端到端测试 (4小时)
  ```
  测试流程:
  1. 用户注册 → 登录 → 获得Token
  2. 创建角色 → 验证角色存储
  3. 开始对话 → 发送3轮消息
  4. 验证上下文记忆（第3轮能记住第1轮内容）
  5. 验证Token统计准确
  ```

- [ ] **TASK-0802**: 性能测试 (2小时)
  ```bash
  # 使用Apache Bench测试
  ab -n 100 -c 10 http://localhost:8000/api/v1/chat/messages
  # 目标: P95 < 3s
  ```

- [ ] **TASK-0803**: 修复发现的Bug (4小时)
  - 记录Bug清单
  - 逐个修复
  - 回归测试

- [ ] **TASK-0804**: 代码审查和重构 (2小时)
  - 消除重复代码
  - 优化错误处理
  - 添加日志

#### Day 10: Docker化和文档

- [ ] **TASK-1001**: 编写Dockerfile (2小时)
  ```dockerfile
  # services/user-service/Dockerfile
  FROM golang:1.21-alpine AS builder
  WORKDIR /app
  COPY . .
  RUN go build -o user-service ./cmd/user-service

  FROM alpine:latest
  COPY --from=builder /app/user-service /usr/local/bin/
  CMD ["user-service"]
  ```

- [ ] **TASK-1002**: 编写docker-compose.yml (1.5小时)
  ```yaml
  version: '3.8'
  services:
    postgres:
      image: pgvector/pgvector:pg15
    redis:
      image: redis:7-alpine
    user-service:
      build: ./services/user-service
    # ... 其他服务
  ```

- [ ] **TASK-1003**: 编写README.md (1.5小时)
  ```markdown
  # AI Chat Server

  ## 快速开始
  ```bash
  docker-compose up -d
  ```

  ## API文档
  见 docs/api.md
  ```

- [ ] **TASK-1004**: 录制演示视频 (1小时)

**阶段1验收**: MVP可以通过docker-compose一键启动

---

## 🧠 阶段2: 记忆系统 (第3-4周)

**目标**: 实现完整的分层记忆系统

### 第3周任务

#### Day 11-12: Embedding和向量检索

- [ ] **TASK-1101**: 实现Embedding生成服务 (2小时)
  ```python
  # app/core/embedding.py
  async def generate_embedding(text: str, model: str = "text-embedding-3-small") -> List[float]:
      response = await openai.Embedding.acreate(
          input=text,
          model=model
      )
      return response.data[0].embedding
  ```

- [ ] **TASK-1102**: 实现批量Embedding (1.5小时)
  ```python
  async def generate_embeddings_batch(texts: List[str]) -> List[List[float]]:
      # 批量处理提升效率
      response = await openai.Embedding.acreate(input=texts, model=model)
      return [item.embedding for item in response.data]
  ```

- [ ] **TASK-1103**: 实现向量检索功能 (3小时)
  ```python
  # app/services/memory.py
  async def retrieve_memories(
      user_id: str,
      character_id: str,
      query_embedding: List[float],
      limit: int = 5
  ) -> List[Memory]:
      # 调用PostgreSQL函数: search_memories_with_decay
      memories = await db.fetch("""
          SELECT * FROM search_memories_with_decay($1, $2, $3, $4, 30)
      """, user_id, character_id, query_embedding, limit)
      return memories
  ```

- [ ] **TASK-1104**: 优化向量索引 (1.5小时)
  ```sql
  -- 验证HNSW索引
  EXPLAIN ANALYZE
  SELECT * FROM long_term_memories
  ORDER BY embedding <=> '[...]' LIMIT 5;

  -- 调优参数
  ALTER INDEX idx_memories_embedding SET (m = 16, ef_construction = 64);
  ```

- [ ] **TASK-1105**: 编写向量检索测试 (2小时)
  ```python
  @pytest.mark.asyncio
  async def test_vector_search():
      # 插入测试记忆
      await insert_test_memories()

      # 搜索"Python项目"
      results = await retrieve_memories(
          user_id="test_user",
          query_text="Python项目怎么样了"
      )

      # 验证返回相关记忆
      assert len(results) > 0
      assert results[0].similarity > 0.7
  ```

#### Day 13-14: 记忆提取和存储

- [ ] **TASK-1301**: 实现关键信息提取 (4小时)
  ```python
  # app/services/extraction.py
  async def extract_key_info(messages: List[Message]) -> List[KeyInfo]:
      """从对话中提取关键信息"""
      prompt = """
      分析以下对话，提取关键信息并分类：
      1. 用户画像 (兴趣、偏好、背景)
      2. 关键事件 (重要的决策、发现)
      3. 关系变化 (情感连接的变化)
      4. 情感状态 (当前情绪)

      对话内容:
      {conversation}

      以JSON格式返回:
      [
        {
          "type": "user_profile",
          "content": "用户喜欢Python编程",
          "importance": 7,
          "structured_data": {"interest": "Python"}
        },
        ...
      ]
      """

      response = await llm_client.chat_completion(
          messages=[{"role": "user", "content": prompt}],
          response_format={"type": "json_object"}
      )

      return parse_key_info(response)
  ```

- [ ] **TASK-1302**: 实现重要性评估 (2小时)
  ```python
  async def evaluate_importance(content: str, context: str) -> int:
      """评估记忆的重要性 (1-10)"""
      prompt = f"""
      评估以下信息的重要性（1-10分）:

      信息: {content}
      上下文: {context}

      考虑因素:
      - 是否是关键决策
      - 是否影响长期行为
      - 是否具有情感意义

      只返回数字。
      """
      response = await llm_client.chat_completion(messages=[...])
      return int(response.content.strip())
  ```

- [ ] **TASK-1303**: 实现Memory Service (3小时)
  ```go
  // services/memory-service/internal/biz/memory.go
  func (mc *MemoryUseCase) SaveMemory(ctx context.Context, req *SaveMemoryRequest) error {
      // 1. 验证输入
      // 2. 生成embedding (调用LLM Service)
      embedding := mc.llmClient.GenerateEmbedding(ctx, req.Content)

      // 3. 保存到数据库
      memory := &Memory{
          UserID:      req.UserID,
          CharacterID: req.CharacterID,
          MemoryType:  req.MemoryType,
          Content:     req.Content,
          Embedding:   embedding,
          Importance:  req.Importance,
          StructuredData: req.StructuredData,
      }
      return mc.data.CreateMemory(ctx, memory)
  }
  ```

- [ ] **TASK-1304**: 集成记忆提取到对话流程 (2小时)
  ```go
  // 在Chat Service中，对话完成后
  go func() {
      // 异步提取关键信息
      keyInfo := llmClient.ExtractKeyInfo(ctx, messages)

      // 为每个关键信息生成embedding并保存
      for _, info := range keyInfo {
          memoryService.SaveMemory(ctx, &SaveMemoryRequest{
              UserID:      userID,
              CharacterID: characterID,
              MemoryType:  info.Type,
              Content:     info.Content,
              Importance:  info.Importance,
              StructuredData: info.StructuredData,
          })
      }
  }()
  ```

#### Day 15: 上下文压缩

- [ ] **TASK-1501**: 实现压缩检测逻辑 (1.5小时)
  ```go
  func (cc *ChatUseCase) shouldCompress(context *TempContext) bool {
      threshold := cc.getModelMaxTokens() * 0.8
      return context.TokenCount >= int(threshold)
  }
  ```

- [ ] **TASK-1502**: 实现压缩服务 (3小时)
  ```python
  # app/services/compression.py
  async def compress_context(messages: List[Message], strategy: str = "hybrid") -> CompressionResult:
      """压缩对话上下文"""
      if strategy == "summary":
          summary = await _summarize(messages)
          return CompressionResult(summary=summary, key_points=[])

      elif strategy == "key_points":
          key_points = await _extract_key_points(messages)
          return CompressionResult(summary="", key_points=key_points)

      else:  # hybrid
          summary, key_points = await asyncio.gather(
              _summarize(messages),
              _extract_key_points(messages)
          )
          return CompressionResult(summary=summary, key_points=key_points)

  async def _summarize(messages: List[Message]) -> str:
      """生成摘要"""
      prompt = """
      总结以下对话的核心内容，保留关键信息点。要求:
      1. 包含讨论的主题
      2. 包含重要的结论
      3. 包含用户的主要意图
      4. 简洁明了，不超过200字

      对话内容:
      {conversation}
      """
      # 调用LLM生成摘要
      ...
  ```

- [ ] **TASK-1503**: 集成压缩到对话流程 (2小时)
  ```go
  if cc.shouldCompress(tempContext) {
      // 调用压缩服务
      compressed := cc.llmClient.CompressContext(ctx, &CompressRequest{
          Messages: tempContext.Messages,
          Strategy: "hybrid",
      })

      // 保存压缩记录
      cc.data.SaveCompression(ctx, compressed)

      // 提取的记忆存入长期记忆
      for _, memory := range compressed.ExtractedMemories {
          cc.memoryService.SaveMemory(ctx, memory)
      }

      // 更新临时上下文
      tempContext = &TempContext{
          Messages: [
              {Role: "system", Content: compressed.Summary},
              ...recentMessages,  // 保留最近3条
          ],
      }
  }
  ```

- [ ] **TASK-1504**: 测试压缩功能 (1.5小时)

---

### 第4周任务

#### Day 16-17: 关系图谱和情感追踪

- [ ] **TASK-1601**: 实现关系图谱更新逻辑 (3小时)
  ```go
  // services/memory-service/internal/biz/relationship.go
  func (rc *RelationshipUseCase) UpdateAfterInteraction(
      ctx context.Context,
      userID, characterID string,
      interactionQuality float64,  // 0.0-1.0
  ) error {
      rel := rc.data.GetRelationship(ctx, userID, characterID)

      // 时间衰减
      daysSinceLast := time.Since(rel.UpdatedAt).Hours() / 24
      decay := math.Exp(-daysSinceLast / 7)  // 7天半衰期

      // 亲密度增长
      increment := 1.0 * interactionQuality
      newCloseness := math.Min(100,
          float64(rel.Closeness)*decay + increment,
      )

      // 更新关系
      rel.Closeness = int(newCloseness)
      rel.InteractionCount++
      rel.UpdatedAt = time.Now()

      return rc.data.UpdateRelationship(ctx, rel)
  }
  ```

- [ ] **TASK-1602**: 实现情感分析 (2小时)
  ```python
  # app/services/sentiment.py
  async def analyze_sentiment(text: str) -> SentimentAnalysis:
      """分析文本的情感"""
      prompt = """
      分析以下文本的情感状态，返回JSON格式:
      {
        "emotions": {
          "joy": 0.0-1.0,
          "sadness": 0.0-1.0,
          "anger": 0.0-1.0,
          "fear": 0.0-1.0,
          "surprise": 0.0-1.0,
          "trust": 0.0-1.0
        },
        "overall_sentiment": "positive/neutral/negative",
        "sentiment_score": -1.0 到 1.0
      }

      文本: {text}
      """

      response = await llm_client.chat_completion(
          messages=[{"role": "user", "content": prompt}],
          response_format={"type": "json_object"}
      )

      return SentimentAnalysis(**json.loads(response.content))
  ```

- [ ] **TASK-1603**: 集成情感追踪 (2小时)
  ```go
  // 对话完成后，保存情感状态
  sentiment := llmResp.Sentiment
  cc.memoryService.RecordEmotionalState(ctx, &EmotionalState{
      UserID:        userID,
      CharacterID:   characterID,
      ConversationID: conversationID,
      Emotions:      sentiment.Emotions,
      SentimentScore: sentiment.SentimentScore,
  })
  ```

- [ ] **TASK-1604**: 实现互动质量评估 (2小时)
  ```python
  async def evaluate_interaction_quality(user_msg: str, ai_response: str) -> float:
      """评估本轮对话的质量 (0.0-1.0)"""
      # 考虑因素:
      # 1. 用户是否满意 (从回复中推断)
      # 2. 对话是否流畅
      # 3. AI回答是否有帮助
      ...
  ```

#### Day 18-19: 前端记忆可视化

- [ ] **TASK-1801**: 实现记忆列表API (2小时)
  ```go
  // GET /api/v1/conversations/{id}/memories
  func (mc *MemoryService) ListMemories(ctx context.Context, req *ListMemoriesRequest) (*ListMemoriesResponse, error) {
      memories := mc.data.ListMemories(ctx, &ListOptions{
          ConversationID: req.ConversationID,
          MemoryTypes:   req.MemoryTypes,
          Limit:         req.Limit,
      })
      return &ListMemoriesResponse{Memories: memories}, nil
  }
  ```

- [ ] **TASK-1802**: 前端记忆时间线组件 (3小时)
  ```vue
  <!-- src/components/MemoryTimeline.vue -->
  <template>
    <div class="memory-timeline">
      <n-timeline>
        <n-timeline-item v-for="memory in memories" :key="memory.id">
          <div class="memory-card">
            <div class="type">{{ memory.memoryType }}</div>
            <div class="content">{{ memory.content }}</div>
            <div class="importance">重要性: {{ memory.importance }}</div>
            <div class="time">{{ formatTime(memory.createdAt) }}</div>
          </div>
        </n-timeline-item>
      </n-timeline>
    </div>
  </template>
  ```

- [ ] **TASK-1803**: 关系图谱可视化 (3小时)
  ```vue
  <!-- 使用ECharts展示关系图谱 -->
  <template>
    <div ref="chartRef" style="width: 100%; height: 400px"></div>
  </template>

  <script setup>
  import * as echarts from 'echarts';

  const option = {
    series: [{
      type: 'graph',
      layout: 'force',
      data: [
        { name: '用户', category: 0 },
        { name: '角色1', category: 1 },
        { name: '角色2', category: 1 },
      ],
      links: [
        { source: '用户', target: '角色1', value: closeness1 },
        { source: '用户', target: '角色2', value: closeness2 },
      ],
    }]
  };
  </script>
  ```

- [ ] **TASK-1804**: 情感曲线图 (2小时)
  ```vue
  <!-- 使用折线图展示情感变化趋势 -->
  <template>
    <div ref="emotionChartRef"></div>
  </template>
  ```

#### Day 20: 阶段2测试和优化

- [ ] **TASK-2001**: 端到端测试记忆系统 (3小时)
  ```
  测试场景:
  1. 对话10轮 → 验证记忆被正确提取和存储
  2. 查询"我上次提到的项目" → 验证能检索到相关记忆
  3. 对话30轮触发压缩 → 验证压缩后仍能记住关键信息
  4. 长时间不互动 → 验证亲密度衰减
  ```

- [ ] **TASK-2002**: 性能优化 (2小时)
  - 向量检索批量优化
  - Embedding缓存
  - 数据库查询优化

- [ ] **TASK-2003**: 修复Bug (3小时)

**阶段2验收**: 记忆系统完整可用，能记住30天前的对话

---

## 🎯 阶段3: 高级功能 (第5-6周)

**目标**: 实现流式响应、角色版本控制、成本统计

### 第5周任务

#### Day 21-22: 流式响应

- [ ] **TASK-2101**: Python实现流式响应 (3小时)
  ```python
  async def chat_stream(request: ChatRequest) -> AsyncIterator[ChatStreamResponse]:
      """流式返回对话"""
      stream = await openai.ChatCompletion.acreate(
          messages=messages,
          model=model,
          stream=True
      )

      async for chunk in stream:
          delta = chunk.choices[0].delta.get("content", "")
          if delta:
              yield ChatStreamResponse(delta=delta, is_final=False)

      # 最后一个chunk包含完整信息
      yield ChatStreamResponse(delta="", is_final=True, final_response=...)
  ```

- [ ] **TASK-2102**: Golang实现SSE (2小时)
  ```go
  func (cs *ChatService) StreamMessage(req *StreamRequest, stream pb.ChatService_StreamMessageServer) error {
      // 调用LLM Service的流式接口
      llmStream, err := cs.llmClient.ChatStream(ctx, &llmReq)

      for {
          chunk, err := llmStream.Recv()
          if err == io.EOF {
              break
          }

          // 转发给前端
          stream.Send(&StreamResponse{
              Delta:   chunk.Delta,
              IsFinal: chunk.IsFinal,
          })
      }
  }
  ```

- [ ] **TASK-2103**: 前端实现流式显示 (2小时)
  ```typescript
  // src/api/chat.ts
  async function* streamChat(conversationId: string, message: string) {
      const response = await fetch(`/api/v1/conversations/${conversationId}/stream`, {
          method: 'POST',
          body: JSON.stringify({ content: message }),
      });

      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          const chunk = decoder.decode(value);
          yield JSON.parse(chunk);
      }
  }
  ```

- [ ] **TASK-2104**: 测试流式响应 (1小时)

#### Day 23-24: 角色版本控制

- [ ] **TASK-2301**: 实现版本快照 (2小时)
  ```go
  func (cc *CharacterUseCase) CreateVersion(ctx context.Context, characterID string, reason string) error {
      char := cc.data.GetCharacter(ctx, characterID)

      // 创建版本快照
      version := &CharacterVersion{
          CharacterID:    characterID,
          Version:        char.Version,
          Personality:    char.Personality,
          BackgroundStory: char.BackgroundStory,
          SystemPrompt:   char.SystemPrompt,
          ChangeReason:   reason,
      }

      return cc.data.CreateVersion(ctx, version)
  }
  ```

- [ ] **TASK-2302**: 实现版本回滚 (2小时)
  ```go
  func (cc *CharacterUseCase) RollbackToVersion(ctx context.Context, versionID string) error {
      version := cc.data.GetVersion(ctx, versionID)

      // 更新角色为历史版本的内容
      char := cc.data.GetCharacter(ctx, version.CharacterID)
      char.Personality = version.Personality
      char.BackgroundStory = version.BackgroundStory
      char.SystemPrompt = version.SystemPrompt
      char.Version++  // 版本号继续递增

      // 保存
      cc.data.UpdateCharacter(ctx, char)

      // 创建回滚快照
      cc.CreateVersion(ctx, char.ID, fmt.Sprintf("Rollback to version %d", version.Version))

      return nil
  }
  ```

- [ ] **TASK-2303**: 前端版本历史页面 (3小时)

#### Day 25: 成本统计

- [ ] **TASK-2501**: 实现API调用日志 (2小时)
  ```go
  func (cs *ChatService) logAPICall(ctx context.Context, req *ChatRequest, resp *ChatResponse, err error) {
      log := &LLMAPICall{
          UserID:        req.UserID,
          ConversationID: req.ConversationID,
          Provider:      "openai",
          Model:         resp.ModelUsed,
          PromptTokens:  resp.TokenUsage.PromptTokens,
          CompletionTokens: resp.TokenUsage.CompletionTokens,
          TotalTokens:   resp.TokenUsage.TotalTokens,
          EstimatedCost: calculateCost(resp.TokenUsage),
          LatencyMs:     resp.LatencyMs,
          Status:        getStatus(err),
      }
      cs.data.SaveAPICall(ctx, log)
  }
  ```

- [ ] **TASK-2502**: 成本计算函数 (1小时)
  ```go
  func calculateCost(usage *TokenUsage) float64 {
      // GPT-3.5-turbo价格
      inputPrice := 0.0005  // $0.50 / 1M tokens
      outputPrice := 0.0015 // $1.50 / 1M tokens

      cost := (float64(usage.PromptTokens) / 1000000 * inputPrice) +
              (float64(usage.CompletionTokens) / 1000000 * outputPrice)

      return cost
  }
  ```

- [ ] **TASK-2503**: 统计API (2小时)
  ```go
  // GET /api/v1/analytics/cost
  func (as *AnalyticsService) GetCostStats(ctx context.Context, req *CostStatsRequest) (*CostStatsResponse, error) {
      stats := as.data.QueryCostStats(ctx, &QueryOptions{
          UserID:    req.UserID,
          StartDate: req.StartDate,
          EndDate:   req.EndDate,
      })

      return &CostStatsResponse{
          TotalCost:    stats.TotalCost,
          TotalTokens:  stats.TotalTokens,
          CallCount:    stats.CallCount,
          DailyBreakdown: stats.DailyBreakdown,
      }, nil
  }
  ```

- [ ] **TASK-2504**: 前端成本仪表盘 (3小时)

---

### 第6周任务

#### Day 26-28: 完善和优化

- [ ] **TASK-2601**: Rate Limiting实现 (2小时)
  ```go
  // 使用Redis实现滑动窗口限流
  func (rl *RateLimiter) Allow(userID string) (bool, error) {
      key := fmt.Sprintf("rate_limit:%s", userID)
      now := time.Now().Unix()

      // 清理1小时前的记录
      rl.redis.ZRemRangeByScore(key, 0, now-3600)

      // 计数
      count := rl.redis.ZCard(key)
      if count >= 100 {  // 每小时100次
          return false, nil
      }

      // 添加当前请求
      rl.redis.ZAdd(key, now, uuid.New().String())
      rl.redis.Expire(key, 3600)

      return true, nil
  }
  ```

- [ ] **TASK-2602**: 监控和日志 (3小时)
  - 集成Prometheus
  - 自定义metrics
  - 结构化日志

- [ ] **TASK-2603**: 错误处理优化 (2小时)
  - 统一错误码
  - 友好的错误信息
  - 错误追踪

- [ ] **TASK-2604**: 安全加固 (2小时)
  - CORS配置
  - XSS防护
  - SQL注入防护验证

- [ ] **TASK-2605**: 性能优化 (3小时)
  - 数据库连接池调优
  - Redis Pipeline
  - gRPC连接复用

#### Day 29-30: 功能测试

- [ ] **TASK-2901**: 完整的集成测试 (4小时)
- [ ] **TASK-2902**: 压力测试 (2小时)
- [ ] **TASK-2903**: Bug修复 (4小时)

**阶段3验收**: 所有高级功能可用

---

## 📦 阶段4: 优化和发布 (第7-8周)

### 第7周: 测试和文档

- [ ] **TASK-3001**: 编写API文档 (4小时)
- [ ] **TASK-3002**: 编写部署文档 (3小时)
- [ ] **TASK-3003**: 编写用户手册 (3小时)
- [ ] **TASK-3004**: 代码注释完善 (2小时)
- [ ] **TASK-3005**: 单元测试覆盖率 > 80% (6小时)

### 第8周: 上线准备

- [ ] **TASK-3101**: Kubernetes部署配置 (4小时)
- [ ] **TASK-3102**: CI/CD Pipeline (3小时)
- [ ] **TASK-3103**: 监控和告警配置 (3小时)
- [ ] **TASK-3104**: 备份恢复测试 (2小时)
- [ ] **TASK-3105**: 生产环境部署 (4小时)

---

## ✅ 每日检查清单

### 开发前
- [ ] 拉取最新代码
- [ ] 检查依赖是否更新
- [ ] 查看今日任务

### 开发中
- [ ] 遵循代码规范
- [ ] 编写必要的注释
- [ ] 编写单元测试
- [ ] 本地测试通过

### 开发后
- [ ] 代码格式化
- [ ] 提交代码（有意义的commit message）
- [ ] 更新文档
- [ ] 更新TODO状态

---

## 📚 学习资源

### Golang
- Kratos官方文档: https://go-kratos.dev/
- Wire教程: https://github.com/google/wire/blob/main/docs/guide.md
- gRPC Go教程: https://grpc.io/docs/languages/go/

### Python
- LangChain文档: https://python.langchain.com/
- FastAPI文档: https://fastapi.tiangolo.com/
- OpenAI API文档: https://platform.openai.com/docs/

### 前端
- Vue 3文档: https://vuejs.org/
- Naive UI: https://www.naiveui.com/
- Pinia: https://pinia.vuejs.org/

### 数据库
- pgvector GitHub: https://github.com/pgvector/pgvector
- PostgreSQL性能调优: https://wiki.postgresql.org/wiki/Performance_Optimization

---

## 🎉 完成标准

项目被认为"完成"需要满足:

1. ✅ 所有P0和P1功能实现
2. ✅ 单元测试覆盖率 > 70%
3. ✅ 集成测试通过
4. ✅ 性能指标达标 (P95响应时间 < 3s)
5. ✅ 可以通过docker-compose一键启动
6. ✅ 文档齐全（README、API文档、部署文档）
7. ✅ 至少在本地环境稳定运行1周无重大bug

**预祝开发顺利！🚀**
