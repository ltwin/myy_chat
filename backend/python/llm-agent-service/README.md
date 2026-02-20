# LLM Agent Service

AI 对话代理服务，为 myy_chat 平台提供智能对话能力。

## 功能特性

- **LiteLLM 集成**: 使用 OpenAI SDK 与 LiteLLM Proxy 通信，支持多模型调用
- **预算追踪**: 用户级别的积分预算管理和限流
- **LangGraph 工作流**: 基于 LangGraph 的对话代理，支持流式和非流式响应
- **系统提示词**: 动态构建角色人设、用户画像、内容审核规则
- **错误处理**: 智能重试、预算拒绝检测、降级处理
- **内容审核**: 符合角色性格的敏感内容拒绝策略（FR-069）

## 项目结构

```
llm-agent-service/
├── app/
│   ├── api/                    # API 层
│   │   ├── grpc_server.py     # gRPC 服务实现
│   │   └── health.py          # 健康检查
│   ├── core/                   # 核心模块
│   │   ├── llm_client.py      # OpenAI SDK 客户端
│   │   └── llm_proxy_client.py # 高级包装器
│   ├── services/               # 业务逻辑
│   │   ├── agent.py           # LangGraph 对话代理
│   │   └── prompt_builder.py  # 系统提示词构建
│   └── config/                 # 配置
│       └── settings.py
├── configs/
│   └── moderation_config.yaml  # 内容审核配置
├── proto/
│   └── llm_agent.proto         # gRPC 协议定义
├── tests/                      # 单元测试
│   ├── test_llm_client.py
│   ├── test_llm_proxy_client.py
│   ├── test_prompt_builder.py
│   └── test_agent.py
├── requirements.txt
├── pytest.ini
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
pip install -r requirements.txt
```

### 2. 生成 Proto 代码

```bash
bash generate_proto.sh
```

### 3. 配置环境变量

创建 `.env` 文件：

```bash
# LiteLLM Proxy
LLM_AGENT_LITELLM_PROXY_URL=http://litellm:4000
LLM_AGENT_LITELLM_API_KEY=your-api-key

# Server
LLM_AGENT_HTTP_PORT=8080
LLM_AGENT_GRPC_PORT=50051
LLM_AGENT_DEBUG=false

# Redis
LLM_AGENT_REDIS_URL=redis://localhost:6379

# Database (read-only)
LLM_AGENT_DATABASE_URL=postgresql://localhost:5432/myychat
```

### 4. 运行服务

```bash
python -m app.main
```

服务将在以下端口启动：
- HTTP: `http://localhost:8080`
- gRPC: `localhost:50051`

## 开发

### 运行测试

```bash
# 运行所有测试
pytest

# 运行特定测试文件
pytest tests/test_agent.py

# 查看覆盖率
pytest --cov=app --cov-report=html
```

### 代码质量检查

```bash
# Ruff linting
ruff check .

# Black formatting
black .

# Type checking
mypy app/
```

## API 使用

### gRPC API

#### ChatCompletion (非流式)

```python
import grpc
from app.api.generated import llm_agent_pb2, llm_agent_pb2_grpc

# 创建连接
channel = grpc.insecure_channel('localhost:50051')
stub = llm_agent_pb2_grpc.LLMAgentServiceStub(channel)

# 构建请求
character = llm_agent_pb2.Character(
    name="小艾",
    personality="活泼开朗",
    moderation_style="cheerful"
)

request = llm_agent_pb2.ChatCompletionRequest(
    user_id="user123",
    user_message="你好",
    character=character,
    model="gpt-4o-mini",
    temperature=0.7
)

# 发送请求
response = stub.ChatCompletion(request)
print(response.response)
```

#### StreamChatCompletion (流式)

```python
request = llm_agent_pb2.ChatCompletionRequest(
    user_id="user123",
    user_message="讲个故事",
    character=character
)

# 流式接收响应
for chunk in stub.StreamChatCompletion(request):
    print(chunk.delta, end='', flush=True)
```

## 配置

### 内容审核配置 (configs/moderation_config.yaml)

```yaml
# 敏感内容类别
categories:
  - political
  - profanity
  - illegal
  - nsfw

# 拒绝模板（保持角色一致性）
refusal_templates:
  cheerful: "哎呀，这个话题我不太方便聊呢~"
  strict: "我不能回应这类不当内容。"
  gentle: "对不起呢，这个我不太方便说..."
```

### 模型配置

默认使用 `gpt-4o-mini`，可通过请求参数指定其他模型：
- `gpt-4o-mini` (默认)
- `gpt-4`
- `claude-3-5-sonnet`
- 其他 LiteLLM 支持的模型

## 架构设计

### 对话流程

```
用户请求
  ↓
gRPC Server (grpc_server.py)
  ↓
ConversationAgent (agent.py)
  ↓
PromptBuilder (prompt_builder.py) → 构建系统提示词
  ↓
LLMProxyClient (llm_proxy_client.py) → 错误处理 + 重试
  ↓
LLMClient (llm_client.py) → OpenAI SDK
  ↓
LiteLLM Proxy → 多模型路由
  ↓
OpenAI / Claude / Gemini / ...
```

### 错误处理策略

1. **预算不足**: 检测 `AuthenticationError` 中的 "budget" 关键字，抛出 `BudgetRejectionError`
2. **限流错误**: 指数退避重试（最多 3 次）
3. **连接错误**: 指数退避重试（最多 3 次）
4. **服务器错误 (5xx)**: 重试
5. **客户端错误 (4xx)**: 不重试，直接抛出

### 系统提示词构成

1. **角色设定**: 名称、性格、背景、说话风格、世界观
2. **用户画像**: 姓名、年龄、兴趣、偏好（Phase 4）
3. **对话上下文**: 关系类型、交互模式
4. **内容审核**: 敏感内容识别 + 角色化拒绝（FR-069）
5. **一致性约束**: 保持角色一致性，不暴露系统逻辑（QR-015）
6. **安全锁**: 公共交互时的隐私保护和权限边界（FR-059）

## 测试覆盖率

目标覆盖率: **≥70%**

当前覆盖的模块：
- ✅ `core/llm_client.py` - OpenAI SDK 客户端
- ✅ `core/llm_proxy_client.py` - 错误处理和重试
- ✅ `services/prompt_builder.py` - 提示词构建
- ✅ `services/agent.py` - LangGraph 对话代理

## 依赖服务

- **LiteLLM Proxy**: `http://litellm:4000` (必需)
- **Redis**: `redis://localhost:6379` (可选，用于限流)
- **PostgreSQL**: `postgresql://localhost:5432/myychat` (可选，读取角色/用户数据)

## 性能指标

- **响应延迟**: P95 < 3s（非流式），首字节 < 1s（流式）
- **并发能力**: 支持 100+ 并发请求
- **可用性**: 99.5% (通过 LiteLLM 多模型故障切换)

## License

Proprietary - MyY Chat Platform

## 联系方式

如有问题，请联系开发团队。
