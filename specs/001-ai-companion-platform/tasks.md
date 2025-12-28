---
description: "Task list for AI角色对话平台 MVP implementation"
---

# Tasks: AI角色对话平台

**Input**: Design documents from `/specs/001-ai-companion-platform/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Unit tests are included for every implementation task to achieve 70% coverage requirement.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

## Path Conventions

This is a microservices web application:
- **Backend Golang**: `backend/golang/app/[service-name]/` (Kratos v2 大仓模式)
- **Backend Python**: `backend/python/[service-name]/`
- **Frontend**: `frontend/src/` (React 19 + Vite + TypeScript)

## MyY Chat Constitution Quality Tasks

Each implementation phase MUST include:

- **Testing Tasks**: Unit tests, Integration tests, Coverage verification (≥70%)
- **Documentation Tasks**: API docs, Design docs, Operation manuals
- **Security Tasks**: Static analysis, Security review, Concurrency verification
- **Deployment Tasks**: Docker Compose config, Kubernetes manifests, Config switching
- **Code Quality Tasks**: Chinese comments, English logging, Code reviews

---

## Phase 1: Setup & Infrastructure (19 tasks)

**Purpose**: Project initialization and basic structure

- [X] T001 [P] Initialize repository structure at /root/workspace/AI/myy_chat with backend/golang/, backend/python/, frontend/, docs/, deployments/
- [X] T002 [P] Initialize Golang module structure with go.mod for each microservice (user-service, character-service, conversation-service, memory-service, billing-service, admin-service, analytics-service)
- [X] T003 [P] Initialize Python project structure with requirements.txt for each service (llm-agent-service, memory-processor, compression-service)
- [X] T004 [P] Initialize React 19.2.0 + Vite 7.2.4 frontend project at frontend/ with TypeScript 5.9.3, React Router 7.11.0, TailwindCSS 3.4.17, and Framer Motion
- [X] T005 [P] Configure Golang linting tools (golangci-lint) and formatting (gofmt) in .golangci.yml
- [X] T006 [P] Configure Python linting tools (ruff, black, mypy) and testing (pytest) in pyproject.toml
- [X] T007 [P] Configure TypeScript ESLint and Prettier for frontend in eslint.config.js
- [X] T008 [P] Setup Go testing framework with testify and gomock in each Golang service
- [X] T009 [P] Setup Python testing framework with pytest, pytest-asyncio, pytest-cov in each Python service
- [X] T010 [P] Setup Playwright for frontend E2E testing at frontend/tests/e2e/
- [ ] T236 [P] Setup Vitest for frontend unit testing at frontend/ with vitest.config.ts, @testing-library/react, and sample component tests
- [X] T011 [P] Configure Chinese comment and English logging standards document in docs/development.md
- [X] T012 [P] Setup pre-commit hooks for linting and testing in .pre-commit-config.yaml
- [X] T013 [P] Create shared Protobuf definitions directory at backend/proto/ with user.proto, conversation.proto, memory.proto
- [X] T014 [P] Configure Snowflake ID environment variables template in .env.example (DATACENTER_ID, WORKER_ID)
- [X] T015 [P] Setup Git workflow with branch protection rules in .github/workflows/
- [X] T016 [P] Create initial Docker Compose development environment skeleton in deployments/docker-compose.dev.yml
- [X] T017 [P] Configure CI pipeline for testing and building in .github/workflows/ci.yml
- [X] T018 [P] Create project README.md with setup instructions and architecture overview

---

## Phase 2: Foundational (Blocking Prerequisites) (57 tasks)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database & ID Generation (10 tasks)

- [ ] T019 Implement Snowflake ID generator library in backend/golang/pkg/snowflake/generator.go using github.com/bwmarrin/snowflake
- [ ] T020 Write unit tests for Snowflake ID generator in backend/golang/pkg/snowflake/generator_test.go
- [ ] T021 Implement Snowflake ID generator library in backend/python/common/snowflake.py using pysnowflake
- [ ] T022 Write unit tests for Python Snowflake ID generator in backend/python/common/test_snowflake.py
- [ ] T023 Create PostgreSQL initialization script at deployments/postgres/init.sql with pgvector extension
- [ ] T024 Create database migration 001_create_users_table.sql in backend/golang/migrations/ with BIGINT id (Snowflake ID)
- [ ] T025 Create database migration 002_create_user_profiles_table.sql with BIGINT user_id
- [ ] T026 Create database migration 003_create_characters_table.sql with BIGINT id and user_id
- [ ] T027 Create database migration 004_create_conversations_table.sql with BIGINT id, user_id, character_id
- [ ] T028 Create database migration 005_create_messages_table.sql with BIGINT id, conversation_id

### Protobuf & gRPC Setup (8 tasks)

- [ ] T029 [P] Copy user_service.proto from specs/001-ai-companion-platform/contracts/ to backend/golang/api/user/v1/
- [ ] T030 [P] Copy conversation_service.proto from specs/001-ai-companion-platform/contracts/ to backend/golang/api/conversation/v1/
- [ ] T031 [P] Copy memory_service.proto from specs/001-ai-companion-platform/contracts/ to backend/golang/api/memory/v1/
- [ ] T032 [P] Create Makefile for Protobuf compilation in backend/golang/Makefile with Go and Python targets
- [ ] T033 [P] Compile user_service.proto to generate Go code in backend/golang/api/user/v1/
- [ ] T034 [P] Compile conversation_service.proto to generate Go code in backend/golang/api/conversation/v1/
- [ ] T035 [P] Compile memory_service.proto to generate Go code in backend/golang/api/memory/v1/
- [ ] T036 [P] Compile all proto files to generate Python code in backend/python/proto/

### Shared Middleware & Infrastructure (12 tasks)

- [ ] T037 Implement JWT authentication middleware in backend/golang/pkg/middleware/auth.go with token validation
- [ ] T038 Write unit tests for JWT middleware in backend/golang/pkg/middleware/auth_test.go
- [ ] T039 [P] Implement logging middleware with English logs in backend/golang/pkg/middleware/logging.go using structured logging
- [ ] T040 [P] Write unit tests for logging middleware in backend/golang/pkg/middleware/logging_test.go
- [ ] T041 [P] Implement error handling middleware in backend/golang/pkg/middleware/error.go with standard error codes
- [ ] T042 [P] Write unit tests for error handling middleware in backend/golang/pkg/middleware/error_test.go
- [ ] T043 [P] Implement distributed tracing middleware in backend/golang/pkg/middleware/tracing.go using OpenTelemetry
- [ ] T044 [P] Implement Redis connection pool library in backend/golang/pkg/redis/client.go
- [ ] T045 [P] Write unit tests for Redis client in backend/golang/pkg/redis/client_test.go
- [ ] T046 Implement distributed lock utility in backend/golang/pkg/redis/lock.go using Redis SET NX
- [ ] T253 Implement cascade delete utilities in backend/golang/pkg/data/cascade.go (FR-055: application-layer referential integrity for user deletion cascading to conversations, memories, credits)
- [ ] T254 Write unit tests for cascade delete utilities in backend/golang/pkg/data/cascade_test.go

### Service Scaffolding (7 tasks)

- [ ] T047 [P] Scaffold user service with Kratos framework in backend/golang/app/user/ (cmd/user/main.go, internal/biz/, internal/data/, internal/service/)
- [ ] T048 [P] Scaffold character service with Kratos framework in backend/golang/app/character/
- [ ] T049 [P] Scaffold conversation service with Kratos framework in backend/golang/app/conversation/
- [ ] T050 [P] Scaffold memory service with Kratos framework in backend/golang/app/memory/
- [ ] T051 [P] Scaffold billing service with Kratos framework in backend/golang/app/billing/
- [ ] T052 [P] Scaffold llm-agent-service with FastAPI + gRPC in backend/python/llm-agent-service/ (FastAPI for health/debug, gRPC for service communication, calls LiteLLM Proxy for LLM)
- [ ] T053 [P] Scaffold memory-processor with FastAPI + gRPC in backend/python/memory-processor/ (FastAPI for health/debug, gRPC for service communication)

### Docker & Deployment (10 tasks)

- [ ] T054 Create PostgreSQL 16 service configuration in deployments/docker-compose.dev.yml with pgvector extension
- [ ] T055 Create Redis 7 service configuration in deployments/docker-compose.dev.yml
- [ ] T056 Create Kafka 3.5 service configuration (optional for MVP, can use Redis Streams) in deployments/docker-compose.dev.yml
- [ ] T057 [P] Create Dockerfile for user service in backend/golang/app/user/Dockerfile
- [ ] T058 [P] Create Dockerfile for character service in backend/golang/app/character/Dockerfile
- [ ] T059 [P] Create Dockerfile for conversation service in backend/golang/app/conversation/Dockerfile
- [ ] T060 [P] Create Dockerfile for memory service in backend/golang/app/memory/Dockerfile
- [ ] T061 [P] Create Dockerfile for llm-agent-service in backend/python/llm-agent-service/Dockerfile
- [ ] T062 [P] Create Dockerfile for memory-processor in backend/python/memory-processor/Dockerfile
- [ ] T063 Complete docker-compose.dev.yml with all services, networks, and volumes

### LiteLLM Proxy 部署 (FR-061 ~ FR-068) (6 tasks)

- [ ] T222 [P] Create LiteLLM Proxy configuration in deployments/litellm/config.yaml (multi-model routing: Azure OpenAI/Claude/Gemini, fallback strategy, Redis cache, user budget settings)
- [ ] T223 [P] Add LiteLLM Proxy service to deployments/docker-compose.dev.yml with PostgreSQL/Redis integration, health check, and environment variables
- [ ] T224 [P] Implement user budget sync between billing-service and LiteLLM Proxy via /user/new and /user/update APIs in backend/golang/app/billing/internal/biz/litellm_sync.go
- [ ] T225 [P] Write integration tests for LiteLLM Proxy multi-model routing and failover in tests/integration/test_litellm_proxy.py
- [ ] T226 [P] Configure LiteLLM Proxy Prometheus metrics export and create Grafana dashboard in deployments/monitoring/litellm-dashboard.json
- [ ] T227 [P] Document LiteLLM Proxy configuration and API usage in docs/llm-gateway.md

### 积分计费与 LLM 成本对接 (FR-032a, FR-035a, FR-063) (4 tasks)

- [ ] T228 [P] Add credit_price_mapping initial data to system_configs in deployments/postgres/init.sql (USD-to-credit conversion rate, model multipliers, platform margin)
- [ ] T229 [P] Create credit_packages table migration and initial data in deployments/postgres/init.sql (5 preset packages: 体验包/基础包/标准包/豪华包/尊享包)
- [ ] T230 [P] Define SyncLLMCost RPC in backend/golang/api/billing/v1/billing.proto (request: user_id, cost_usd, model, input_tokens, output_tokens; response: credits_deducted)
- [ ] T231 [P] Define UpdateCreditFormula and ListCreditPackages RPCs in backend/golang/api/admin/v1/admin.proto (admin management APIs for credit configuration)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 用户与AI角色基础对话 (Priority: P1) 🎯 MVP (68 tasks)

**Goal**: 用户可以注册账号、选择预设AI角色、发送消息并获得回复

**Independent Test**: 用户从注册到发送第一条消息并获得AI回复的完整流程，无需依赖其他功能模块

### User Service Implementation (18 tasks)

- [ ] T064 [P] [US1] Create User entity model in backend/golang/app/user/internal/biz/user.go with Snowflake ID, username, email, password_hash
- [ ] T065 [P] [US1] Write unit tests for User entity in backend/golang/app/user/internal/biz/user_test.go
- [ ] T066 [P] [US1] Create UserProfile entity model in backend/golang/app/user/internal/biz/user_profile.go
- [ ] T067 [P] [US1] Write unit tests for UserProfile entity in backend/golang/app/user/internal/biz/user_profile_test.go
- [ ] T068 [US1] Implement UserRepository interface in backend/golang/app/user/internal/data/user.go with pgx connection and Snowflake ID generation
- [ ] T069 [US1] Write unit tests for UserRepository in backend/golang/app/user/internal/data/user_test.go
- [ ] T070 [US1] Implement UserService with Register logic in backend/golang/app/user/internal/biz/user_service.go (email validation, password bcrypt, Snowflake ID)
- [ ] T071 [US1] Write unit tests for UserService Register in backend/golang/app/user/internal/biz/user_service_test.go
- [ ] T072 [US1] Implement UserService Login logic in backend/golang/app/user/internal/biz/user_service.go (JWT generation)
- [ ] T073 [US1] Write unit tests for UserService Login in backend/golang/app/user/internal/biz/user_service_test.go
- [ ] T234 [US1] Implement login failure lockout middleware in backend/golang/pkg/middleware/login_lockout.go (5 failed attempts → 15min lockout, Redis counter with TTL)
- [ ] T235 [US1] Write unit tests for login lockout middleware in backend/golang/pkg/middleware/login_lockout_test.go
- [ ] T074 [US1] Implement UserService GetUser logic in backend/golang/app/user/internal/biz/user_service.go
- [ ] T075 [US1] Write unit tests for UserService GetUser in backend/golang/app/user/internal/biz/user_service_test.go
- [ ] T076 [US1] Implement gRPC service endpoints (Register, Login, GetUser) in backend/golang/app/user/internal/service/user_service.go
- [ ] T077 [US1] Write integration tests for user service gRPC endpoints in backend/golang/app/user/tests/integration/user_test.go
- [ ] T078 [US1] Add initial credit grant (100 credits) logic in UserService Register method
- [ ] T079 [US1] Write unit tests for initial credit grant in backend/golang/app/user/internal/biz/user_service_test.go

### Character Service Implementation (12 tasks)

- [ ] T080 [P] [US1] Create Character entity model in backend/golang/app/character/internal/biz/character.go with Snowflake ID, name, personality, background
- [ ] T081 [P] [US1] Write unit tests for Character entity in backend/golang/app/character/internal/biz/character_test.go
- [ ] T082 [US1] Implement CharacterRepository in backend/golang/app/character/internal/data/character.go with pgx connection
- [ ] T083 [US1] Write unit tests for CharacterRepository in backend/golang/app/character/internal/data/character_test.go
- [ ] T084 [US1] Implement CharacterService with ListCharacters logic in backend/golang/app/character/internal/biz/character_service.go
- [ ] T085 [US1] Write unit tests for CharacterService ListCharacters in backend/golang/app/character/internal/biz/character_service_test.go
- [ ] T086 [US1] Implement CharacterService GetCharacter logic in backend/golang/app/character/internal/biz/character_service.go
- [ ] T087 [US1] Write unit tests for CharacterService GetCharacter in backend/golang/app/character/internal/biz/character_service_test.go
- [ ] T088 [US1] Create database seed script with 5 preset characters in backend/golang/migrations/006_seed_characters.sql
- [ ] T089 [US1] Implement gRPC service endpoints (ListCharacters, GetCharacter) in backend/golang/app/character/internal/service/character_service.go
- [ ] T090 [US1] Write integration tests for character service gRPC endpoints in backend/golang/app/character/tests/integration/character_test.go
- [ ] T091 [US1] Add system prompt generation logic based on character personality in backend/golang/app/character/internal/biz/prompt.go

### Conversation Service Implementation (16 tasks)

- [ ] T092 [P] [US1] Create Conversation entity model in backend/golang/app/conversation/internal/biz/conversation.go with Snowflake ID, user_id, character_id
- [ ] T093 [P] [US1] Write unit tests for Conversation entity in backend/golang/app/conversation/internal/biz/conversation_test.go
- [ ] T094 [P] [US1] Create Message entity model in backend/golang/app/conversation/internal/biz/message.go with Snowflake ID, conversation_id, role, content
- [ ] T095 [P] [US1] Write unit tests for Message entity in backend/golang/app/conversation/internal/biz/message_test.go
- [ ] T096 [US1] Implement ConversationRepository in backend/golang/app/conversation/internal/data/conversation.go with pgx
- [ ] T097 [US1] Write unit tests for ConversationRepository in backend/golang/app/conversation/internal/data/conversation_test.go
- [ ] T098 [US1] Implement MessageRepository in backend/golang/app/conversation/internal/data/message.go with pgx
- [ ] T099 [US1] Write unit tests for MessageRepository in backend/golang/app/conversation/internal/data/message_test.go
- [ ] T100 [US1] Implement ConversationService CreateConversation logic in backend/golang/app/conversation/internal/biz/conversation_service.go
- [ ] T101 [US1] Write unit tests for CreateConversation in backend/golang/app/conversation/internal/biz/conversation_service_test.go
- [ ] T102 [US1] Implement ConversationService SendMessage logic with LLM service call in backend/golang/app/conversation/internal/biz/conversation_service.go
- [ ] T103 [US1] Write unit tests for SendMessage in backend/golang/app/conversation/internal/biz/conversation_service_test.go
- [ ] T104 [US1] Implement ConversationService GetMessages logic in backend/golang/app/conversation/internal/biz/conversation_service.go
- [ ] T105 [US1] Write unit tests for GetMessages in backend/golang/app/conversation/internal/biz/conversation_service_test.go
- [ ] T106 [US1] Implement gRPC service endpoints (CreateConversation, SendMessage, GetMessages) in backend/golang/app/conversation/internal/service/conversation_service.go
- [ ] T107 [US1] Write integration tests for conversation service gRPC endpoints in backend/golang/app/conversation/tests/integration/conversation_test.go

### LLM Agent Service Implementation (Python) (6 tasks)

> **Note**: llm-agent-service 通过 LiteLLM Proxy 调用 LLM，无需直接集成各提供商 SDK

- [ ] T108 [US1] Implement OpenAI-compatible client for LiteLLM Proxy in backend/python/llm-agent-service/app/core/llm_client.py (base_url pointing to LiteLLM Proxy, user parameter for budget)
- [ ] T109 [US1] Write unit tests for LLM client in backend/python/llm-agent-service/tests/test_llm_client.py (mock LiteLLM Proxy responses, test budget rejection handling)
- [ ] T110 [US1] Implement basic LangGraph agent workflow in backend/python/llm-agent-service/app/services/agent.py (no tools, just conversation)
- [ ] T111 [US1] Write unit tests for LangGraph agent in backend/python/llm-agent-service/tests/test_agent.py
- [ ] T112 [US1] Implement gRPC server for ChatCompletion in backend/python/llm-agent-service/app/api/grpc_server.py
- [ ] T113 [US1] Write integration tests for llm-agent-service gRPC server in backend/python/llm-agent-service/tests/test_grpc_integration.py

### LLM Proxy Integration & Failover (FR-040a, FR-066) (4 tasks)

> **Note**: 熔断和故障切换由 LiteLLM Proxy 内置功能处理，以下任务聚焦于集成和测试

- [ ] T114 [US1] Implement LiteLLM Proxy client wrapper in backend/python/llm-agent-service/app/core/llm_proxy_client.py (OpenAI SDK pointing to LiteLLM Proxy, user parameter for budget tracking)
- [ ] T115 [US1] Write unit tests for LiteLLM Proxy client in backend/python/llm-agent-service/tests/test_llm_proxy_client.py (test connection, error handling, user budget rejection)
- [ ] T116 [US1] Configure LiteLLM Proxy fallback routes in deployments/litellm/config.yaml (gpt-4o → claude-3-5-sonnet → gemini-2.0-flash, with 3-retry and latency-based routing)
- [ ] T117 [US1] Write E2E tests for LLM failover via LiteLLM Proxy in tests/integration/test_llm_failover_e2e.py (simulate provider failures, verify automatic switch <10s, test recovery)

### Frontend Chat UI (FR-012a/b) (4 tasks)

- [ ] T237 [P] [US1] Implement ChatLoadingState component in frontend/src/components/chat/ChatLoadingState.tsx (3s: "正在思考中...", 10s: "AI正在努力思考，请稍候...", 30s: timeout with retry button)
- [ ] T238 [P] [US1] Write unit tests for ChatLoadingState component in frontend/src/components/chat/ChatLoadingState.test.tsx
- [ ] T239 [US1] Integrate ChatLoadingState into ChatPage.tsx with timeout handling and retry logic
- [ ] T240 [US1] Write E2E tests for chat timeout scenarios in frontend/tests/e2e/chat-timeout.spec.ts

### Content Moderation (FR-069) (5 tasks)

> **Note**: 通过 System Prompt 指导 LLM 识别敏感内容并以角色口吻拒绝，无需前置分类器

- [ ] T248 [US1] Update System Prompt template with content moderation instructions in backend/python/llm-agent-service/app/services/prompt_builder.py (FR-069a: political/profanity/illegal content detection, character-appropriate refusal)
- [ ] T249 [US1] Create moderation category configuration file at backend/python/llm-agent-service/configs/moderation_config.yaml (categories: political, profanity, illegal; refusal style templates per character type)
- [ ] T250 [US1] Write unit tests for content moderation scenarios in backend/python/llm-agent-service/tests/test_moderation.py (FR-069b: verify character consistency in refusal responses)
- [ ] T251 [US1] Implement moderation event audit logging in backend/golang/app/analytics/internal/service/moderation_audit.go (FR-069c: log trigger type only, no message content)
- [ ] T252 [US1] Write unit tests for moderation audit logging in backend/golang/app/analytics/internal/service/moderation_audit_test.go

---

## Phase 4: User Story 2 - AI角色记忆用户信息 (Priority: P1) 🎯 MVP (44 tasks)

**Goal**: AI能够记住用户告诉它的个人信息,在多次对话中表现出对用户的了解

**Independent Test**: 用户在对话中告诉AI个人信息,在后续对话或新会话中AI能够主动使用这些信息

### Memory Database Setup (4 tasks)

- [ ] T118 [P] [US2] Create database migration 007_create_memories_table.sql with BIGINT id, user_id, character_id, content, embedding VECTOR(1536)
- [ ] T119 [P] [US2] Create database migration 008_create_relationships_table.sql with BIGINT user_id, character_id, closeness, emotional_bond
- [ ] T120 [P] [US2] Create HNSW index on memories embedding column in migration 009_create_memory_indexes.sql
- [ ] T121 [P] [US2] Add pgvector similarity search function in migration 010_create_similarity_functions.sql

### Memory Service Implementation (16 tasks)

- [ ] T122 [P] [US2] Create Memory entity model in backend/golang/app/memory/internal/biz/memory.go with Snowflake ID, type, importance, embedding
- [ ] T123 [P] [US2] Write unit tests for Memory entity in backend/golang/app/memory/internal/biz/memory_test.go
- [ ] T124 [P] [US2] Create Relationship entity model in backend/golang/app/memory/internal/biz/relationship.go with closeness, emotional_bond
- [ ] T125 [P] [US2] Write unit tests for Relationship entity in backend/golang/app/memory/internal/biz/relationship_test.go
- [ ] T126 [P] [US2] Create UserPortrait entity model in backend/golang/app/memory/internal/biz/user_portrait.go
- [ ] T127 [P] [US2] Write unit tests for UserPortrait entity in backend/golang/app/memory/internal/biz/user_portrait_test.go
- [ ] T128 [US2] Implement MemoryRepository with pgvector similarity search in backend/golang/app/memory/internal/data/memory.go
- [ ] T129 [US2] Write unit tests for MemoryRepository in backend/golang/app/memory/internal/data/memory_test.go
- [ ] T130 [US2] Implement RelationshipRepository in backend/golang/app/memory/internal/data/relationship.go
- [ ] T131 [US2] Write unit tests for RelationshipRepository in backend/golang/app/memory/internal/data/relationship_test.go
- [ ] T132 [US2] Implement MemoryService SaveMemory logic in backend/golang/app/memory/internal/biz/memory_service.go
- [ ] T133 [US2] Write unit tests for SaveMemory in backend/golang/app/memory/internal/biz/memory_service_test.go
- [ ] T134 [US2] Implement MemoryService RetrieveMemories logic with vector similarity in backend/golang/app/memory/internal/biz/memory_service.go
- [ ] T135 [US2] Write unit tests for RetrieveMemories in backend/golang/app/memory/internal/biz/memory_service_test.go
- [ ] T136 [US2] Implement gRPC service endpoints (SaveMemory, RetrieveMemories, GetUserPortrait) in backend/golang/app/memory/internal/service/memory_service.go
- [ ] T137 [US2] Write integration tests for memory service gRPC endpoints in backend/golang/app/memory/tests/integration/memory_test.go

### Memory Processor Implementation (Python) (8 tasks)

- [ ] T138 [US2] Implement SiliconFlow Embedding API client in backend/python/memory-processor/app/services/embedding.py
- [ ] T139 [US2] Write unit tests for SiliconFlow client in backend/python/memory-processor/tests/test_embedding.py
- [ ] T140 [US2] Implement memory extraction logic from conversation in backend/python/memory-processor/app/services/extractor.py
- [ ] T141 [US2] Write unit tests for memory extractor in backend/python/memory-processor/tests/test_extractor.py
- [ ] T142 [US2] Implement contradiction detection logic in backend/python/memory-processor/app/services/contradiction.py
- [ ] T143 [US2] Write unit tests for contradiction detection in backend/python/memory-processor/tests/test_contradiction.py
- [ ] T144 [US2] Implement gRPC server for memory processing in backend/python/memory-processor/app/api/grpc_server.py
- [ ] T145 [US2] Write integration tests for memory-processor gRPC in backend/python/memory-processor/tests/test_integration.py

### LLM Service Memory Integration (8 tasks)

- [ ] T146 [US2] Integrate memory retrieval into LangGraph agent workflow in backend/python/llm-agent-service/app/services/agent.py
- [ ] T147 [US2] Write unit tests for memory integration in agent in backend/python/llm-agent-service/tests/test_agent_memory.py
- [ ] T148 [US2] Implement system prompt injection with user portrait in backend/python/llm-agent-service/app/services/prompt_builder.py
- [ ] T149 [US2] Write unit tests for prompt builder in backend/python/llm-agent-service/tests/test_prompt_builder.py
- [ ] T150 [US2] Implement contradiction handling workflow in backend/python/llm-agent-service/app/services/contradiction_handler.py
- [ ] T151 [US2] Write unit tests for contradiction handler in backend/python/llm-agent-service/tests/test_contradiction_handler.py
- [ ] T152 [US2] Update ConversationService to call memory-processor after each message in backend/golang/app/conversation/internal/biz/conversation_service.go
- [ ] T153 [US2] Write unit tests for memory integration in ConversationService in backend/golang/app/conversation/internal/biz/conversation_service_test.go

### Frontend Memory UI (FR-024, FR-025) (8 tasks)

> **Note**: FR-024 (查看记忆) 和 FR-025 (编辑/删除记忆) 基础功能在此阶段实现；高级记忆管理 UI (MEM-010) 在 Phase 5

- [ ] T154 [P] [US2] Create UserPortrait page component in frontend/src/pages/ProfilePage.tsx (extend existing) displaying user profile and memory info
- [ ] T155 [P] [US2] Write unit tests for UserPortrait component in frontend/src/pages/ProfilePage.test.tsx
- [ ] T156 [P] [US2] Create MemoryList component in frontend/src/components/memory/MemoryList.tsx with view/delete actions
- [ ] T157 [P] [US2] Write unit tests for MemoryList component in frontend/src/components/memory/MemoryList.test.tsx
- [ ] T158 [US2] Implement API client for memory service in frontend/src/services/memoryApi.ts (including deleteMemory, updateMemory)
- [ ] T159 [US2] Write integration tests for memory UI flow in frontend/tests/e2e/memory.spec.ts
- [ ] T241 [US2] Implement MemoryEditDialog component in frontend/src/components/memory/MemoryEditDialog.tsx (FR-025: edit memory content)
- [ ] T242 [US2] Write unit tests for MemoryEditDialog in frontend/src/components/memory/MemoryEditDialog.test.tsx

### Frontend Multi-Device Sync (FR-005a) (2 tasks)

- [ ] T160 [US1] Implement frontend polling mechanism in frontend/src/lib/hooks/useMessagePolling.ts with 3-5s interval for multi-device conversation history sync
- [ ] T161 [US1] Write unit tests for polling hook in frontend/src/lib/hooks/useMessagePolling.test.tsx (test interval timing, error handling, cleanup on unmount, reconnection logic)

### Scheduled Account Deletion (FR-004d) (1 task)

- [ ] T162 [US1] Create scheduled job for automatic account deletion in backend/golang/app/user/internal/job/deletion_scheduler.go (daily cron at 2 AM, deletes accounts where deletion_scheduled_at < NOW() - 30 days, transactional cascade delete across all user tables, logs deletion events for compliance audit)

**Checkpoint**: MVP Core Complete (US1 + US2) - Basic conversation with AI memory system functional

---

## Phase 5: Long-Term Memory System (五层记忆架构) (90 tasks)

**Purpose**: Implement advanced memory system based on Mnemosyne, PRIME, MemoryBank, and Titans research for human-like memory capabilities

**Reference**: research.md Section 4.5, data-model.md Sections 8-10

**v1.4 Updates**:
- Added MEM-012: Multi-Dimensional Surprise Scoring (8 tasks)
- Added MEM-013: Memory Information Extraction Pipeline (10 tasks)
- Added MEM-014: Memory Storage Decision Logic (6 tasks)

### Story MEM-001: Neo4j Infrastructure Setup (8 tasks)

**Goal**: Set up Neo4j graph database and PostgreSQL-Neo4j synchronization

- [ ] T-MEM-001a [P] Add Neo4j 5.x service to deployments/docker-compose.dev.yml with vector index support
- [ ] T-MEM-001b [P] Implement Neo4j connection pool in backend/golang/pkg/neo4j/client.go using github.com/neo4j/neo4j-go-driver/v5
- [ ] T-MEM-001c Write unit tests for Neo4j client in backend/golang/pkg/neo4j/client_test.go
- [ ] T-MEM-001d Create Neo4j schema initialization script at deployments/neo4j/init.cypher with constraints and vector indexes
- [ ] T-MEM-001e [P] Implement Neo4j Python client in backend/python/common/neo4j_client.py
- [ ] T-MEM-001f Write unit tests for Python Neo4j client in backend/python/common/test_neo4j_client.py
- [ ] T-MEM-001g Implement PostgreSQL-Neo4j sync service in backend/golang/app/memory/internal/sync/neo4j_sync.go
- [ ] T-MEM-001h Write unit tests for sync service in backend/golang/app/memory/internal/sync/neo4j_sync_test.go

### Story MEM-002: Layer 1 - Conversation Compression (6 tasks)

**Goal**: Implement ConversationSummaryBufferMemory for efficient context management

- [ ] T-MEM-002a Create database migration for conversation_summaries table in backend/golang/migrations/011_create_conversation_summaries.sql
- [ ] T-MEM-002b [P] Create ConversationSummary entity in backend/golang/app/memory/internal/biz/conversation_summary.go
- [ ] T-MEM-002c Write unit tests for ConversationSummary entity in backend/golang/app/memory/internal/biz/conversation_summary_test.go
- [ ] T-MEM-002d Implement conversation summarization logic in backend/python/memory-processor/app/services/summarizer.py using LLM
- [ ] T-MEM-002e Write unit tests for summarizer in backend/python/memory-processor/tests/test_summarizer.py
- [ ] T-MEM-002f Integrate summarization into ConversationService with sliding window in backend/golang/app/conversation/internal/biz/conversation_service.go

### Story MEM-003: Layer 2 - Emotional Tracking (8 tasks)

**Goal**: Track user emotional states using Russell Circumplex Model

- [ ] T-MEM-003a Create database migration for emotional_states table in backend/golang/migrations/012_create_emotional_states.sql
- [ ] T-MEM-003b [P] Create EmotionalState entity in backend/golang/app/memory/internal/biz/emotional_state.go with valence/arousal
- [ ] T-MEM-003c Write unit tests for EmotionalState entity in backend/golang/app/memory/internal/biz/emotional_state_test.go
- [ ] T-MEM-003d Implement EmotionalStateRepository in backend/golang/app/memory/internal/data/emotional_state.go
- [ ] T-MEM-003e Write unit tests for EmotionalStateRepository in backend/golang/app/memory/internal/data/emotional_state_test.go
- [ ] T-MEM-003f Implement emotion detection prompts in backend/python/memory-processor/app/services/emotion_detector.py
- [ ] T-MEM-003g Write unit tests for emotion detector in backend/python/memory-processor/tests/test_emotion_detector.py
- [ ] T-MEM-003h Implement emotion trend analysis API in backend/golang/app/memory/internal/service/emotion_service.go

### Story MEM-004: Layer 3 - Important Events (Mnemosyne) (8 tasks)

**Goal**: Detect and store important life events with surprise scoring

- [ ] T-MEM-004a Create database migration for important_events table in backend/golang/migrations/013_create_important_events.sql
- [ ] T-MEM-004b [P] Create ImportantEvent entity in backend/golang/app/memory/internal/biz/important_event.go with surprise_score
- [ ] T-MEM-004c Write unit tests for ImportantEvent entity in backend/golang/app/memory/internal/biz/important_event_test.go
- [ ] T-MEM-004d Implement ImportantEventRepository in backend/golang/app/memory/internal/data/important_event.go with HNSW search
- [ ] T-MEM-004e Write unit tests for ImportantEventRepository in backend/golang/app/memory/internal/data/important_event_test.go
- [ ] T-MEM-004f Implement Mnemosyne surprise scoring algorithm in backend/python/memory-processor/app/services/surprise_scorer.py
- [ ] T-MEM-004g Write unit tests for surprise scorer in backend/python/memory-processor/tests/test_surprise_scorer.py
- [ ] T-MEM-004h Implement temporal context extraction in backend/python/memory-processor/app/services/temporal_extractor.py

### Story MEM-005: Layer 4 - Knowledge Graph (Neo4j) (10 tasks)

**Goal**: Build entity relationship graph for user's social network and knowledge

- [ ] T-MEM-005a Create database migration for memory_graph_nodes table in backend/golang/migrations/014_create_memory_graph_nodes.sql
- [ ] T-MEM-005b Create database migration for memory_graph_edges table in backend/golang/migrations/015_create_memory_graph_edges.sql
- [ ] T-MEM-005c [P] Create GraphNode entity in backend/golang/app/memory/internal/biz/graph_node.go with Neo4j sync flag
- [ ] T-MEM-005d Write unit tests for GraphNode entity in backend/golang/app/memory/internal/biz/graph_node_test.go
- [ ] T-MEM-005e [P] Create GraphEdge entity in backend/golang/app/memory/internal/biz/graph_edge.go
- [ ] T-MEM-005f Write unit tests for GraphEdge entity in backend/golang/app/memory/internal/biz/graph_edge_test.go
- [ ] T-MEM-005g Implement LLM-based entity extraction in backend/python/memory-processor/app/services/entity_extractor.py
- [ ] T-MEM-005h Write unit tests for entity extractor in backend/python/memory-processor/tests/test_entity_extractor.py
- [ ] T-MEM-005i Implement relationship extraction in backend/python/memory-processor/app/services/relationship_extractor.py
- [ ] T-MEM-005j Write unit tests for relationship extractor in backend/python/memory-processor/tests/test_relationship_extractor.py

### Story MEM-006: Layer 5 - User Portrait (PRIME) (6 tasks)

**Goal**: Build dynamic user profiles based on semantic memory

- [ ] T-MEM-006a Create database migration for user_portraits table in backend/golang/migrations/016_create_user_portraits.sql
- [ ] T-MEM-006b Implement UserPortraitRepository in backend/golang/app/memory/internal/data/user_portrait.go with core_summary embedding
- [ ] T-MEM-006c Write unit tests for UserPortraitRepository in backend/golang/app/memory/internal/data/user_portrait_test.go
- [ ] T-MEM-006d Implement PRIME-style semantic memory extraction in backend/python/memory-processor/app/services/portrait_builder.py
- [ ] T-MEM-006e Write unit tests for portrait builder in backend/python/memory-processor/tests/test_portrait_builder.py
- [ ] T-MEM-006f Implement core summary generation (Mnemosyne) in backend/python/memory-processor/app/services/core_summary.py

### Story MEM-007: Memory Decay & Reinforcement (MemoryBank) (6 tasks)

**Goal**: Implement Ebbinghaus forgetting curve with reinforcement mechanism

- [ ] T-MEM-007a Extend memories table with decay fields in backend/golang/migrations/017_alter_memories_add_decay.sql
- [ ] T-MEM-007b Implement Ebbinghaus decay algorithm in backend/golang/app/memory/internal/biz/decay.go (R = e^(-t/S), S = stability * (1 + 0.1 * boost_count), daily decay job updates retention_score)
- [ ] T-MEM-007c Write unit tests for decay algorithm in backend/golang/app/memory/internal/biz/decay_test.go (test decay curve, boost mechanics, edge cases)
- [ ] T-MEM-007d Implement memory reinforcement (boost) logic in backend/golang/app/memory/internal/biz/reinforcement.go
- [ ] T-MEM-007e Write unit tests for reinforcement in backend/golang/app/memory/internal/biz/reinforcement_test.go
- [ ] T-MEM-007f Create scheduled job for memory decay in backend/golang/app/memory/internal/job/decay_scheduler.go (daily 3AM)

### Story MEM-008: Hybrid Memory Retrieval (6 tasks)

**Goal**: Combine vector search + graph traversal for context building

- [ ] T-MEM-008a Implement hybrid retrieval service in backend/golang/app/memory/internal/biz/hybrid_retrieval.go
- [ ] T-MEM-008b Write unit tests for hybrid retrieval in backend/golang/app/memory/internal/biz/hybrid_retrieval_test.go
- [ ] T-MEM-008c Implement BM25 reranking in backend/python/memory-processor/app/services/reranker.py
- [ ] T-MEM-008d Write unit tests for reranker in backend/python/memory-processor/tests/test_reranker.py
- [ ] T-MEM-008e Implement context builder combining all 5 layers in backend/python/llm-agent-service/app/services/context_builder.py
- [ ] T-MEM-008f Write integration tests for full memory pipeline in backend/python/memory-processor/tests/test_memory_integration.py

### Story MEM-009: Memory Quality Assurance (6 tasks)

**Goal**: Implement memory deduplication, merging, and surprise scoring based on embedding distance

- [ ] T-MEM-009a Implement memory deduplicator service in backend/python/memory-processor/app/services/deduplicator.py (similarity threshold 0.9)
- [ ] T-MEM-009b Write unit tests for deduplicator in backend/python/memory-processor/tests/test_deduplicator.py
- [ ] T-MEM-009c Implement memory merger with LLM content merging in backend/python/memory-processor/app/services/merger.py
- [ ] T-MEM-009d Write unit tests for merger in backend/python/memory-processor/tests/test_merger.py
- [ ] T-MEM-009e Enhance surprise scorer with embedding distance metrics in backend/python/memory-processor/app/services/surprise_scorer.py (extends T-MEM-004f base implementation, adds novelty_score based on cosine distance)
- [ ] T-MEM-009f Write unit tests for surprise scorer in backend/python/memory-processor/tests/test_surprise_scorer.py

### Story MEM-010: Privacy Control API (6 tasks)

**Goal**: Allow users to view, delete, and manage their memories

- [ ] T-MEM-010a Extend memory_service.proto with privacy control APIs (ListUserMemories, DeleteMemory, UpdateMemory, ForgetTopic)
- [ ] T-MEM-010b Implement ListUserMemories gRPC endpoint in backend/golang/app/memory/internal/service/memory_service.go
- [ ] T-MEM-010c Implement DeleteMemory gRPC endpoint with cascade delete in Neo4j
- [ ] T-MEM-010d Implement ForgetTopic gRPC endpoint (delete memories by topic similarity)
- [ ] T-MEM-010e Create MemoryManagePage in frontend/src/pages/MemoryManagePage.tsx (list, delete, edit memories)
- [ ] T-MEM-010f Write integration tests for privacy control APIs in backend/golang/app/memory/tests/integration/privacy_test.go

### Story MEM-011: Sync Enhancement (6 tasks)

**Goal**: Implement robust PostgreSQL-Neo4j sync with retry, idempotency, and dead letter queue

- [ ] T-MEM-011a Create migration for sync enhancement fields (sync_version, sync_failed_count, last_sync_error) in backend/golang/migrations/018_alter_graph_nodes_sync.sql
- [ ] T-MEM-011b Implement retry-enabled sync service in backend/golang/app/memory/internal/sync/neo4j_sync.go (max 3 retries, exponential backoff)
- [ ] T-MEM-011c Implement dead letter queue handler in backend/golang/app/memory/internal/sync/dead_letter.go
- [ ] T-MEM-011d Implement idempotent version comparison in Neo4j MERGE (compare sync_version)
- [ ] T-MEM-011e Create sync monitoring dashboard config in deployments/monitoring/sync_dashboard.json
- [ ] T-MEM-011f Write integration tests for sync service in backend/golang/app/memory/tests/integration/sync_test.go

### Story MEM-012: Multi-Dimensional Surprise Scoring (8 tasks)

**Goal**: Implement enhanced surprise scoring with novelty, emotional intensity, event type, and user emphasis dimensions

- [ ] T-MEM-012a Create EventType enum and ExtractedInfo dataclass in backend/python/memory-processor/app/models/memory_types.py
- [ ] T-MEM-012b Implement EmotionalAnalysis dataclass with valence/arousal in backend/python/memory-processor/app/models/emotional.py
- [ ] T-MEM-012c Implement EnhancedSurpriseScorer with multi-dimensional scoring in backend/python/memory-processor/app/services/enhanced_surprise_scorer.py
- [ ] T-MEM-012d Write unit tests for EnhancedSurpriseScorer in backend/python/memory-processor/tests/test_enhanced_surprise_scorer.py
- [ ] T-MEM-012e Implement emphasis detection (emphasis markers, repeated mentions) in backend/python/memory-processor/app/services/emphasis_detector.py
- [ ] T-MEM-012f Write unit tests for emphasis detector in backend/python/memory-processor/tests/test_emphasis_detector.py
- [ ] T-MEM-012g Implement score breakdown API for debugging in backend/python/memory-processor/app/api/surprise_api.py
- [ ] T-MEM-012h Write integration tests for multi-dimensional scoring in backend/python/memory-processor/tests/test_scoring_integration.py

### Story MEM-013: Memory Information Extraction Pipeline (10 tasks)

**Goal**: Implement LLM-based structured information extraction from user conversations

- [ ] T-MEM-013a Create MEMORY_EXTRACTION_PROMPT template in backend/python/memory-processor/app/prompts/extraction_prompt.py
- [ ] T-MEM-013b Implement ExtractedMemory dataclass in backend/python/memory-processor/app/models/extracted_memory.py
- [ ] T-MEM-013c Implement MemoryExtractor service with LLM extraction in backend/python/memory-processor/app/services/memory_extractor.py
- [ ] T-MEM-013d Write unit tests for MemoryExtractor in backend/python/memory-processor/tests/test_memory_extractor.py
- [ ] T-MEM-013e Implement extraction result parser (_parse_extraction) in backend/python/memory-processor/app/services/extraction_parser.py
- [ ] T-MEM-013f Write unit tests for extraction parser in backend/python/memory-processor/tests/test_extraction_parser.py
- [ ] T-MEM-013g Implement mentioned_before detection with similarity threshold in backend/python/memory-processor/app/services/mention_detector.py
- [ ] T-MEM-013h Write unit tests for mention detector in backend/python/memory-processor/tests/test_mention_detector.py
- [ ] T-MEM-013i Implement arousal estimation based on event type and strength in backend/python/memory-processor/app/services/arousal_estimator.py
- [ ] T-MEM-013j Write unit tests for arousal estimator in backend/python/memory-processor/tests/test_arousal_estimator.py

### Story MEM-014: Memory Storage Decision Logic (6 tasks)

**Goal**: Implement intelligent storage decision based on surprise score with deduplication

- [ ] T-MEM-014a Implement MemoryStorageDecider with threshold-based decision in backend/python/memory-processor/app/services/storage_decider.py
- [ ] T-MEM-014b Write unit tests for MemoryStorageDecider in backend/python/memory-processor/tests/test_storage_decider.py
- [ ] T-MEM-014c Implement similar memory finder with configurable threshold in backend/python/memory-processor/app/services/similar_finder.py
- [ ] T-MEM-014d Write unit tests for similar finder in backend/python/memory-processor/tests/test_similar_finder.py
- [ ] T-MEM-014e Integrate storage decider into memory processing pipeline in backend/python/memory-processor/app/services/pipeline.py
- [ ] T-MEM-014f Write end-to-end integration tests for full extraction pipeline in backend/python/memory-processor/tests/test_pipeline_e2e.py

---

## Phase 6: User Story 3-11 (Priority: P2/P3) - Post-MVP Features (40 tasks)

**Purpose**: Placeholder tasks for future user stories implementation

### User Story 3 - 用户自定义AI角色 (P2)

- [ ] T163 [US3] Implement CharacterService CreateCharacter logic in backend/golang/app/character/internal/biz/character_service.go
- [ ] T164 [US3] Write unit tests for CreateCharacter
- [ ] T165 [US3] Implement CharacterService UpdateCharacter logic
- [ ] T166 [US3] Write unit tests for UpdateCharacter
- [ ] T167 [US3] Create character creation wizard UI in frontend/src/pages/CreatePage.tsx (extend existing)

### User Story 4 - AI角色使用工具能力 (P2)

- [ ] T168 [US4] Implement WeatherTool in backend/python/llm-agent-service/app/services/tools/weather.py
- [ ] T169 [US4] Write unit tests for WeatherTool
- [ ] T170 [US4] Implement SearchTool in backend/python/llm-agent-service/app/services/tools/search.py
- [ ] T171 [US4] Write unit tests for SearchTool
- [ ] T172 [US4] Implement ImageTool in backend/python/llm-agent-service/app/services/tools/image.py
- [ ] T173 [US4] Write unit tests for ImageTool
- [ ] T174 [US4] Integrate tools into LangGraph agent workflow
- [ ] T175 [US4] Write integration tests for tool-enabled agent

### User Story 5 - 积分充值与消费管理 (P2)

- [ ] T176 [US5] Create database migration for credit_accounts and credit_transactions tables
- [ ] T177 [US5] Implement BillingService with credit deduction logic in backend/golang/app/billing/internal/biz/billing_service.go
- [ ] T178 [US5] Write unit tests for BillingService
- [ ] T179 [US5] Implement idempotency for credit transactions using Redis locks
- [ ] T180 [US5] Write unit tests for idempotency logic
- [ ] T181 [US5] Integrate BillingService with ConversationService
- [ ] T232 [US5] Implement SyncLLMCost RPC handler in billing-service with USD→credit conversion logic (reads credit_price_mapping from system_configs, applies model multiplier and platform margin)
- [ ] T182 [US5] Create credit center UI in frontend/src/pages/ProfilePage.tsx (extend credits section)
- [ ] T243 [US5] *(P3)* Implement WeChat Pay integration in backend/golang/app/billing/internal/service/wechat_pay.go (FR-036)
- [ ] T244 [US5] *(P3)* Implement Alipay integration in backend/golang/app/billing/internal/service/alipay.go (FR-036)
- [ ] T245 [US5] *(P3)* Write integration tests for payment flows in backend/golang/app/billing/tests/integration/payment_test.go

### User Story 6-10 Placeholders (P3)

- [ ] T183 [US6] Implement emotion state tracking in backend/golang/app/memory/internal/biz/emotion.go
- [ ] T184 [US7] Implement admin service for LLM provider configuration in backend/golang/app/admin/
- [ ] T233 [US8] Implement admin-service credit formula and packages management API (UpdateCreditFormula, ListCreditPackages, CreateCreditPackage, UpdateCreditPackage) in backend/golang/app/admin/internal/biz/credit_config.go
- [ ] T185 [US8] *(DEFERRED v1.1)* Implement analytics service for sales CRM in backend/golang/app/analytics/
- [ ] T186 [US9] Setup centralized logging with OpenTelemetry
- [ ] T187 [US9] Setup Prometheus metrics and Grafana dashboards
- [ ] T188 [US9] Setup Jaeger for distributed tracing
- [ ] T189 [US9] Create api_logs table and logging middleware
- [ ] T190 [US10] Implement memory compression service
- [ ] T191 [US10] Implement memory importance scoring algorithm
- [ ] T192 [US10] Create background job for memory cleanup

### User Story 11 - AI角色社区与多人交互 (P3) *(DEFERRED v1.1)* (5 tasks)

> **Note**: FR-056 至 FR-060 已标记为 v1.1 计划，以下为占位任务

- [ ] T216 [US11] *(DEFERRED v1.1)* Implement character publish/unpublish API for community (FR-056) in backend/golang/app/character/internal/service/community_service.go
- [ ] T217 [US11] *(DEFERRED v1.1)* Implement Owner/Visitor permission system (FR-057) in backend/golang/app/character/internal/biz/permission.go
- [ ] T218 [US11] *(DEFERRED v1.1)* Implement memory partition with visibility control (FR-058) in backend/golang/app/memory/internal/biz/partition.go
- [ ] T219 [US11] *(DEFERRED v1.1)* Implement security lock for ownership protection (FR-059) in backend/python/llm-agent-service/app/services/security_lock.py
- [ ] T220 [US11] *(DEFERRED v1.1)* Implement prompt injection and social engineering detection (FR-060) in backend/python/llm-agent-service/app/services/attack_detector.py

---

## Phase 7: Polish & Production Readiness (24 tasks)

**Purpose**: Improvements that affect multiple user stories

### Documentation & Deployment (15 tasks)

- [ ] T193 [P] Generate Swagger documentation for all REST APIs in docs/api/swagger.yaml
- [ ] T194 [P] Generate Protobuf documentation from proto files in docs/api/grpc.md
- [ ] T195 [P] Create architecture documentation in docs/architecture.md
- [ ] T196 [P] Create deployment documentation in docs/deployment.md
- [ ] T197 [P] Create development guide in docs/development.md
- [ ] T198 [P] Create Kubernetes Helm chart for user service in deployments/k8s/user/
- [ ] T199 [P] Create Kubernetes Helm chart for conversation service in deployments/k8s/conversation/
- [ ] T200 [P] Create Kubernetes Helm chart for memory service in deployments/k8s/memory/
- [ ] T201 [P] Create Kubernetes Helm chart for llm-agent-service in deployments/k8s/llm-agent-service/
- [ ] T202 [P] Create Kubernetes StatefulSet configuration with Snowflake ID worker allocation in deployments/k8s/statefulset.yaml
- [ ] T203 [P] Implement HPA (Horizontal Pod Autoscaler) configuration in deployments/k8s/hpa.yaml
- [ ] T204 [P] Optimize Dockerfiles with multi-stage builds
- [ ] T205 [P] Run security audit with static analysis tools (gosec, bandit)
- [ ] T206 [P] Implement rate limiting middleware using Redis in backend/golang/pkg/middleware/ratelimit.go
- [ ] T207 [P] Run quickstart.md validation and update with latest setup instructions

### Security Monitoring Tasks (FR-052, FR-053, FR-054) (10 tasks)

- [ ] T208 [P] [US9] Implement DDOS detection middleware using sliding window rate limiting in backend/golang/pkg/middleware/ddos_detector.go (threshold: 100 req/s per IP, 1000 req/s total)
- [ ] T209 [P] [US9] Write unit tests for DDOS detector in backend/golang/pkg/middleware/ddos_detector_test.go
- [ ] T210 [P] [US9] Implement SQL injection detection filter in backend/golang/pkg/middleware/sql_injection_filter.go (parameterized query enforcement, pattern detection)
- [ ] T211 [P] [US9] Write unit tests for SQL injection filter in backend/golang/pkg/middleware/sql_injection_filter_test.go
- [ ] T212 [US9] Implement alert notification service in backend/golang/app/admin/internal/biz/alert_service.go (email, webhook, Slack integration)
- [ ] T213 [US9] Write unit tests for alert service in backend/golang/app/admin/internal/biz/alert_service_test.go
- [ ] T246 [P] [US9] Add Loki + Grafana services to deployments/docker-compose.dev.yml with volume mounts and network configuration
- [ ] T247 [P] [US9] Create Loki configuration file at deployments/loki/loki-config.yaml (ingester, storage, schema)
- [ ] T214 [US9] Implement log rotation and 30-day retention policy in deployments/loki/retention-config.yaml (Loki/Grafana stack)
- [ ] T215 [US9] Create audit log query API in backend/golang/app/admin/internal/service/audit_service.go with date range and keyword filters

### Quality Assurance Tasks (QR-015) (1 task)

- [ ] T221 [P] Write E2E test for character consistency (QR-015) in frontend/tests/e2e/character-consistency.spec.ts (verify AI responses maintain persona across error handling, contradictions, and edge cases - no "system detected" or machine language allowed)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-4)**: All depend on Foundational phase completion
  - US1 and US2 can proceed in parallel (if staffed)
  - US2 has slight dependency on US1 (needs conversation flow to test memory)
- **Long-Term Memory (Phase 5)**: Depends on US1+US2 completion, implements 5-layer architecture
  - MEM-001 (Neo4j) can start in parallel with late Phase 4
  - MEM-002 through MEM-008 depend on MEM-001 completion
- **Post-MVP (Phase 6)**: Depends on Phase 5 completion, all US3-10 can proceed in parallel
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - Slight integration with US1 but independently testable
- **User Story 3-10 (P2/P3)**: Can start after Foundational - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Unit tests MUST be written immediately after implementation (not batched)
- Models before services
- Services before gRPC endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (T001-T018)
- All Foundational tasks marked [P] can run in parallel within their subsections
- Once Foundational phase completes, US1 and US2 can start in parallel (if team capacity allows)
- All unit test tasks marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories (US3-US10) can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all User entity model tasks together:
Task T064: "Create User entity model"
Task T065: "Write unit tests for User entity"
Task T066: "Create UserProfile entity model"
Task T067: "Write unit tests for UserProfile entity"

# Launch all Character entity model tasks together:
Task T080: "Create Character entity model"
Task T081: "Write unit tests for Character entity"

# Launch all Conversation entity model tasks together:
Task T092: "Create Conversation entity model"
Task T093: "Write unit tests for Conversation entity"
Task T094: "Create Message entity model"
Task T095: "Write unit tests for Message entity"
```

---

## Implementation Strategy

### MVP First (US1 + US2 Only) - Recommended ✅

1. Complete Phase 1: Setup (19 tasks)
2. Complete Phase 2: Foundational (57 tasks) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 1 (68 tasks) - Basic conversation with LLM failover, multi-device sync, content moderation, and account deletion
4. Complete Phase 4: User Story 2 (44 tasks) - AI memory
5. **STOP and VALIDATE**: Test US1+US2 independently (188 tasks total)
6. Deploy MVP and gather user feedback

### Advanced Memory (Recommended for AI Companion) ✅

After MVP validation:

1. Complete Phase 5: Long-Term Memory System (90 tasks)
   - Neo4j setup and PostgreSQL sync
   - 5-layer memory architecture (Mnemosyne, PRIME, MemoryBank)
   - Ebbinghaus decay + reinforcement
   - Hybrid retrieval (vector + graph)
2. **STOP and VALIDATE**: Test advanced memory features (278 tasks total)
3. Deploy Advanced Memory and gather user feedback

### Incremental Delivery (Post-MVP)

1. Complete US1+US2 → Foundation ready (MVP deployed)
2. Complete Phase 5 → Advanced memory ready (Long-term memory deployed)
3. Add User Story 3 → Test independently → Deploy (Custom characters)
4. Add User Story 4 → Test independently → Deploy (AI tools)
5. Add User Story 5 → Test independently → Deploy (Billing system)
6. Add User Stories 6-10 in priority order
7. Complete Phase 7: Polish → Production-ready

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (Phases 1-2)
2. Once Foundational is done:
   - **Team A**: User Story 1 (User, Character, Conversation services)
   - **Team B**: Start User Story 2 scaffolding (Memory, Memory Processor)
   - **Team C**: Frontend setup and reusable components
3. After US1 backend complete:
   - **Team A**: Frontend integration for US1
   - **Team B**: Complete US2 backend implementation
4. After US2 complete:
   - **Team A**: User Story 3 (Custom characters)
   - **Team B**: User Story 4 (AI tools)
   - **Team C**: User Story 5 (Billing)

---

## Task Summary

**Total Tasks**: 344 tasks

**By Phase**:
- Phase 1 (Setup): 19 tasks (18 原有 + 1 Vitest 配置)
- Phase 2 (Foundational): 57 tasks (45 原有 + 6 LiteLLM Proxy + 4 积分计费 + 2 级联删除)
- Phase 3 (US1 - Basic Conversation): 68 tasks (57 原有 + 2 登录锁定 + 4 超时加载状态 + 5 内容审核)
- Phase 4 (US2 - AI Memory): 44 tasks (42 原有 + 2 记忆编辑组件)
- Phase 5 (Long-Term Memory): 90 tasks (5-layer architecture with Neo4j)
- Phase 6 (US3-US11 - Post-MVP): 40 tasks (37 原有 + 3 支付集成)
- Phase 7 (Polish + Security + QA): 26 tasks (24 原有 + 2 Loki 部署)

**MVP Scope (US1 + US2)**: 188 tasks (Phases 1-4)

**Advanced Memory (US1 + US2 + Memory)**: 278 tasks (Phases 1-5)

**New in this update**:
- T222-T227: LiteLLM Proxy 部署和集成 (FR-061 to FR-068)
- T228-T231: 积分计费与 LLM 成本对接基础设施 (FR-032a, FR-035a, FR-063)
- T232-T233: billing-service 和 admin-service 积分管理实现
- T234-T235: 登录失败锁定中间件 (安全)
- T236: Vitest 前端单元测试配置
- T237-T240: 聊天超时加载状态组件 (FR-012a/b)
- T241-T242: 记忆编辑对话框组件 (FR-025)
- T243-T245: 支付集成 (微信支付/支付宝)
- T246-T247: Loki 日志服务部署
- T248-T252: 内容审核与角色化拒绝 (FR-069)
- T253-T254: 应用层级联删除工具 (FR-055)
- llm-service 重命名为 llm-agent-service

**Deferred to v1.1**:
- T185: Sales CRM system (FR-044 to FR-048)
- T216-T220: Community features (FR-056 to FR-060)
- FR-070: 长链任务异步 multi-agent 工作流

**Test Coverage**: Unit test task for every implementation task = 70%+ coverage guaranteed

**Parallel Opportunities**:
- Phase 1: 17 parallelizable tasks
- Phase 2: 25 parallelizable tasks
- Phase 3: 10 parallelizable tasks (entity models)
- Phase 4: 8 parallelizable tasks (entity models)
- Phase 5: 8 parallelizable tasks (Neo4j + entity models)

**Critical Path**:
Setup (Phase 1) → Foundational (Phase 2) → US1 (Phase 3) → US2 (Phase 4) → Long-Term Memory (Phase 5) → MVP+ Complete

---

## Notes

- [P] tasks = different files, no dependencies, safe to parallelize
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Unit tests must be written immediately after implementation (not deferred)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Follow MyY Chat Constitution: Chinese comments, English logs, 70% coverage
- All IDs use Snowflake algorithm (NOT UUID)
- No database foreign keys (application-layer cascade deletes)
- SiliconFlow API for embeddings in MVP
