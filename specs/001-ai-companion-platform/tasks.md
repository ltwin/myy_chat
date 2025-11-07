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
- **Backend Golang**: `backend/golang/[service-name]/`
- **Backend Python**: `backend/python/[service-name]/`
- **Frontend**: `frontend/`

## MyY Chat Constitution Quality Tasks

Each implementation phase MUST include:

- **Testing Tasks**: Unit tests, Integration tests, Coverage verification (≥70%)
- **Documentation Tasks**: API docs, Design docs, Operation manuals
- **Security Tasks**: Static analysis, Security review, Concurrency verification
- **Deployment Tasks**: Docker Compose config, Kubernetes manifests, Config switching
- **Code Quality Tasks**: Chinese comments, English logging, Code reviews

---

## Phase 1: Setup & Infrastructure (18 tasks)

**Purpose**: Project initialization and basic structure

- [ ] T001 [P] Initialize repository structure at /root/workspace/AI/myy_chat with backend/golang/, backend/python/, frontend/, docs/, deployments/
- [ ] T002 [P] Initialize Golang module structure with go.mod for each microservice (user-service, character-service, conversation-service, memory-service, billing-service, admin-service, analytics-service)
- [ ] T003 [P] Initialize Python project structure with requirements.txt for each service (llm-service, memory-processor, compression-service)
- [ ] T004 [P] Initialize Next.js 14 frontend project at frontend/ with App Router, TypeScript 5.x, TailwindCSS 3+, and Shadcn/ui
- [ ] T005 [P] Configure Golang linting tools (golangci-lint) and formatting (gofmt) in .golangci.yml
- [ ] T006 [P] Configure Python linting tools (ruff, black, mypy) and testing (pytest) in pyproject.toml
- [ ] T007 [P] Configure TypeScript ESLint and Prettier for frontend in eslint.config.js
- [ ] T008 [P] Setup Go testing framework with testify and gomock in each Golang service
- [ ] T009 [P] Setup Python testing framework with pytest, pytest-asyncio, pytest-cov in each Python service
- [ ] T010 [P] Setup Playwright for frontend E2E testing at frontend/tests/e2e/
- [ ] T011 [P] Configure Chinese comment and English logging standards document in docs/development.md
- [ ] T012 [P] Setup pre-commit hooks for linting and testing in .pre-commit-config.yaml
- [ ] T013 [P] Create shared Protobuf definitions directory at backend/proto/ with user.proto, conversation.proto, memory.proto
- [ ] T014 [P] Configure Snowflake ID environment variables template in .env.example (DATACENTER_ID, WORKER_ID)
- [ ] T015 [P] Setup Git workflow with branch protection rules in .github/workflows/
- [ ] T016 [P] Create initial Docker Compose development environment skeleton in deployments/docker-compose.dev.yml
- [ ] T017 [P] Configure CI pipeline for testing and building in .github/workflows/ci.yml
- [ ] T018 [P] Create project README.md with setup instructions and architecture overview

---

## Phase 2: Foundational (Blocking Prerequisites) (45 tasks)

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

- [ ] T029 [P] Copy user_service.proto from specs/001-ai-companion-platform/contracts/ to backend/proto/
- [ ] T030 [P] Copy conversation_service.proto from specs/001-ai-companion-platform/contracts/ to backend/proto/
- [ ] T031 [P] Copy memory_service.proto from specs/001-ai-companion-platform/contracts/ to backend/proto/
- [ ] T032 [P] Create Makefile for Protobuf compilation in backend/proto/Makefile with Go and Python targets
- [ ] T033 [P] Compile user_service.proto to generate Go code in backend/golang/user-service/api/user/v1/
- [ ] T034 [P] Compile conversation_service.proto to generate Go code in backend/golang/conversation-service/api/conversation/v1/
- [ ] T035 [P] Compile memory_service.proto to generate Go code in backend/golang/memory-service/api/memory/v1/
- [ ] T036 [P] Compile all proto files to generate Python code in backend/python/proto/

### Shared Middleware & Infrastructure (10 tasks)

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

### Service Scaffolding (7 tasks)

- [ ] T047 [P] Scaffold user-service with Kratos framework in backend/golang/user-service/ (cmd/main.go, internal/biz/, internal/data/, internal/service/)
- [ ] T048 [P] Scaffold character-service with Kratos framework in backend/golang/character-service/
- [ ] T049 [P] Scaffold conversation-service with Kratos framework in backend/golang/conversation-service/
- [ ] T050 [P] Scaffold memory-service with Kratos framework in backend/golang/memory-service/
- [ ] T051 [P] Scaffold billing-service with Kratos framework in backend/golang/billing-service/
- [ ] T052 [P] Scaffold llm-service with FastAPI in backend/python/llm-service/ (app/api/, app/core/, app/services/)
- [ ] T053 [P] Scaffold memory-processor with FastAPI in backend/python/memory-processor/

### Docker & Deployment (10 tasks)

- [ ] T054 Create PostgreSQL 15 service configuration in deployments/docker-compose.dev.yml with pgvector extension
- [ ] T055 Create Redis 7 service configuration in deployments/docker-compose.dev.yml
- [ ] T056 Create Kafka 3.5 service configuration (optional for MVP, can use Redis Streams) in deployments/docker-compose.dev.yml
- [ ] T057 [P] Create Dockerfile for user-service in backend/golang/user-service/Dockerfile
- [ ] T058 [P] Create Dockerfile for character-service in backend/golang/character-service/Dockerfile
- [ ] T059 [P] Create Dockerfile for conversation-service in backend/golang/conversation-service/Dockerfile
- [ ] T060 [P] Create Dockerfile for memory-service in backend/golang/memory-service/Dockerfile
- [ ] T061 [P] Create Dockerfile for llm-service in backend/python/llm-service/Dockerfile
- [ ] T062 [P] Create Dockerfile for memory-processor in backend/python/memory-processor/Dockerfile
- [ ] T063 Complete docker-compose.dev.yml with all services, networks, and volumes

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 用户与AI角色基础对话 (Priority: P1) 🎯 MVP (57 tasks)

**Goal**: 用户可以注册账号、选择预设AI角色、发送消息并获得回复

**Independent Test**: 用户从注册到发送第一条消息并获得AI回复的完整流程，无需依赖其他功能模块

### User Service Implementation (16 tasks)

- [ ] T064 [P] [US1] Create User entity model in backend/golang/user-service/internal/biz/user.go with Snowflake ID, username, email, password_hash
- [ ] T065 [P] [US1] Write unit tests for User entity in backend/golang/user-service/internal/biz/user_test.go
- [ ] T066 [P] [US1] Create UserProfile entity model in backend/golang/user-service/internal/biz/user_profile.go
- [ ] T067 [P] [US1] Write unit tests for UserProfile entity in backend/golang/user-service/internal/biz/user_profile_test.go
- [ ] T068 [US1] Implement UserRepository interface in backend/golang/user-service/internal/data/user.go with pgx connection and Snowflake ID generation
- [ ] T069 [US1] Write unit tests for UserRepository in backend/golang/user-service/internal/data/user_test.go
- [ ] T070 [US1] Implement UserService with Register logic in backend/golang/user-service/internal/biz/user_service.go (email validation, password bcrypt, Snowflake ID)
- [ ] T071 [US1] Write unit tests for UserService Register in backend/golang/user-service/internal/biz/user_service_test.go
- [ ] T072 [US1] Implement UserService Login logic in backend/golang/user-service/internal/biz/user_service.go (JWT generation)
- [ ] T073 [US1] Write unit tests for UserService Login in backend/golang/user-service/internal/biz/user_service_test.go
- [ ] T074 [US1] Implement UserService GetUser logic in backend/golang/user-service/internal/biz/user_service.go
- [ ] T075 [US1] Write unit tests for UserService GetUser in backend/golang/user-service/internal/biz/user_service_test.go
- [ ] T076 [US1] Implement gRPC service endpoints (Register, Login, GetUser) in backend/golang/user-service/internal/service/user_service.go
- [ ] T077 [US1] Write integration tests for user-service gRPC endpoints in backend/golang/user-service/tests/integration/user_test.go
- [ ] T078 [US1] Add initial credit grant (100 credits) logic in UserService Register method
- [ ] T079 [US1] Write unit tests for initial credit grant in backend/golang/user-service/internal/biz/user_service_test.go

### Character Service Implementation (12 tasks)

- [ ] T080 [P] [US1] Create Character entity model in backend/golang/character-service/internal/biz/character.go with Snowflake ID, name, personality, background
- [ ] T081 [P] [US1] Write unit tests for Character entity in backend/golang/character-service/internal/biz/character_test.go
- [ ] T082 [US1] Implement CharacterRepository in backend/golang/character-service/internal/data/character.go with pgx connection
- [ ] T083 [US1] Write unit tests for CharacterRepository in backend/golang/character-service/internal/data/character_test.go
- [ ] T084 [US1] Implement CharacterService with ListCharacters logic in backend/golang/character-service/internal/biz/character_service.go
- [ ] T085 [US1] Write unit tests for CharacterService ListCharacters in backend/golang/character-service/internal/biz/character_service_test.go
- [ ] T086 [US1] Implement CharacterService GetCharacter logic in backend/golang/character-service/internal/biz/character_service.go
- [ ] T087 [US1] Write unit tests for CharacterService GetCharacter in backend/golang/character-service/internal/biz/character_service_test.go
- [ ] T088 [US1] Create database seed script with 5 preset characters in backend/golang/migrations/006_seed_characters.sql
- [ ] T089 [US1] Implement gRPC service endpoints (ListCharacters, GetCharacter) in backend/golang/character-service/internal/service/character_service.go
- [ ] T090 [US1] Write integration tests for character-service gRPC endpoints in backend/golang/character-service/tests/integration/character_test.go
- [ ] T091 [US1] Add system prompt generation logic based on character personality in backend/golang/character-service/internal/biz/prompt.go

### Conversation Service Implementation (16 tasks)

- [ ] T092 [P] [US1] Create Conversation entity model in backend/golang/conversation-service/internal/biz/conversation.go with Snowflake ID, user_id, character_id
- [ ] T093 [P] [US1] Write unit tests for Conversation entity in backend/golang/conversation-service/internal/biz/conversation_test.go
- [ ] T094 [P] [US1] Create Message entity model in backend/golang/conversation-service/internal/biz/message.go with Snowflake ID, conversation_id, role, content
- [ ] T095 [P] [US1] Write unit tests for Message entity in backend/golang/conversation-service/internal/biz/message_test.go
- [ ] T096 [US1] Implement ConversationRepository in backend/golang/conversation-service/internal/data/conversation.go with pgx
- [ ] T097 [US1] Write unit tests for ConversationRepository in backend/golang/conversation-service/internal/data/conversation_test.go
- [ ] T098 [US1] Implement MessageRepository in backend/golang/conversation-service/internal/data/message.go with pgx
- [ ] T099 [US1] Write unit tests for MessageRepository in backend/golang/conversation-service/internal/data/message_test.go
- [ ] T100 [US1] Implement ConversationService CreateConversation logic in backend/golang/conversation-service/internal/biz/conversation_service.go
- [ ] T101 [US1] Write unit tests for CreateConversation in backend/golang/conversation-service/internal/biz/conversation_service_test.go
- [ ] T102 [US1] Implement ConversationService SendMessage logic with LLM service call in backend/golang/conversation-service/internal/biz/conversation_service.go
- [ ] T103 [US1] Write unit tests for SendMessage in backend/golang/conversation-service/internal/biz/conversation_service_test.go
- [ ] T104 [US1] Implement ConversationService GetMessages logic in backend/golang/conversation-service/internal/biz/conversation_service.go
- [ ] T105 [US1] Write unit tests for GetMessages in backend/golang/conversation-service/internal/biz/conversation_service_test.go
- [ ] T106 [US1] Implement gRPC service endpoints (CreateConversation, SendMessage, GetMessages) in backend/golang/conversation-service/internal/service/conversation_service.go
- [ ] T107 [US1] Write integration tests for conversation-service gRPC endpoints in backend/golang/conversation-service/tests/integration/conversation_test.go

### LLM Service Implementation (Python) (6 tasks)

- [ ] T108 [US1] Implement LiteLLM client wrapper in backend/python/llm-service/app/core/llm_client.py with Azure OpenAI and Claude providers
- [ ] T109 [US1] Write unit tests for LiteLLM client in backend/python/llm-service/tests/test_llm_client.py
- [ ] T110 [US1] Implement basic LangGraph agent workflow in backend/python/llm-service/app/services/agent.py (no tools, just conversation)
- [ ] T111 [US1] Write unit tests for LangGraph agent in backend/python/llm-service/tests/test_agent.py
- [ ] T112 [US1] Implement gRPC server for ChatCompletion in backend/python/llm-service/app/api/grpc_server.py
- [ ] T113 [US1] Write integration tests for llm-service gRPC server in backend/python/llm-service/tests/test_grpc_integration.py

### LLM Circuit Breaker & Failover (FR-040a) (4 tasks)

- [ ] T114 [US1] Implement circuit breaker for LLM provider in backend/python/llm-service/app/core/circuit_breaker.py with 3-failure threshold, 10s timeout, and exponential backoff
- [ ] T115 [US1] Write unit tests for circuit breaker in backend/python/llm-service/tests/test_circuit_breaker.py (test failure detection, auto-recovery, threshold logic, state transitions)
- [ ] T116 [US1] Implement automatic failover logic with LiteLLM router in backend/python/llm-service/app/services/llm_router.py (Azure OpenAI → Claude fallback, logs switch events to api_logs)
- [ ] T117 [US1] Write integration tests for LLM failover in backend/python/llm-service/tests/test_failover_integration.py (simulate provider failures, verify switch <10s, test recovery to primary)

---

## Phase 4: User Story 2 - AI角色记忆用户信息 (Priority: P1) 🎯 MVP (42 tasks)

**Goal**: AI能够记住用户告诉它的个人信息,在多次对话中表现出对用户的了解

**Independent Test**: 用户在对话中告诉AI个人信息,在后续对话或新会话中AI能够主动使用这些信息

### Memory Database Setup (4 tasks)

- [ ] T118 [P] [US2] Create database migration 007_create_memories_table.sql with BIGINT id, user_id, character_id, content, embedding VECTOR(1536)
- [ ] T119 [P] [US2] Create database migration 008_create_relationships_table.sql with BIGINT user_id, character_id, closeness, emotional_bond
- [ ] T120 [P] [US2] Create HNSW index on memories embedding column in migration 009_create_memory_indexes.sql
- [ ] T121 [P] [US2] Add pgvector similarity search function in migration 010_create_similarity_functions.sql

### Memory Service Implementation (16 tasks)

- [ ] T122 [P] [US2] Create Memory entity model in backend/golang/memory-service/internal/biz/memory.go with Snowflake ID, type, importance, embedding
- [ ] T123 [P] [US2] Write unit tests for Memory entity in backend/golang/memory-service/internal/biz/memory_test.go
- [ ] T124 [P] [US2] Create Relationship entity model in backend/golang/memory-service/internal/biz/relationship.go with closeness, emotional_bond
- [ ] T125 [P] [US2] Write unit tests for Relationship entity in backend/golang/memory-service/internal/biz/relationship_test.go
- [ ] T126 [P] [US2] Create UserPortrait entity model in backend/golang/memory-service/internal/biz/user_portrait.go
- [ ] T127 [P] [US2] Write unit tests for UserPortrait entity in backend/golang/memory-service/internal/biz/user_portrait_test.go
- [ ] T128 [US2] Implement MemoryRepository with pgvector similarity search in backend/golang/memory-service/internal/data/memory.go
- [ ] T129 [US2] Write unit tests for MemoryRepository in backend/golang/memory-service/internal/data/memory_test.go
- [ ] T130 [US2] Implement RelationshipRepository in backend/golang/memory-service/internal/data/relationship.go
- [ ] T131 [US2] Write unit tests for RelationshipRepository in backend/golang/memory-service/internal/data/relationship_test.go
- [ ] T132 [US2] Implement MemoryService SaveMemory logic in backend/golang/memory-service/internal/biz/memory_service.go
- [ ] T133 [US2] Write unit tests for SaveMemory in backend/golang/memory-service/internal/biz/memory_service_test.go
- [ ] T134 [US2] Implement MemoryService RetrieveMemories logic with vector similarity in backend/golang/memory-service/internal/biz/memory_service.go
- [ ] T135 [US2] Write unit tests for RetrieveMemories in backend/golang/memory-service/internal/biz/memory_service_test.go
- [ ] T136 [US2] Implement gRPC service endpoints (SaveMemory, RetrieveMemories, GetUserPortrait) in backend/golang/memory-service/internal/service/memory_service.go
- [ ] T137 [US2] Write integration tests for memory-service gRPC endpoints in backend/golang/memory-service/tests/integration/memory_test.go

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

- [ ] T146 [US2] Integrate memory retrieval into LangGraph agent workflow in backend/python/llm-service/app/services/agent.py
- [ ] T147 [US2] Write unit tests for memory integration in agent in backend/python/llm-service/tests/test_agent_memory.py
- [ ] T148 [US2] Implement system prompt injection with user portrait in backend/python/llm-service/app/services/prompt_builder.py
- [ ] T149 [US2] Write unit tests for prompt builder in backend/python/llm-service/tests/test_prompt_builder.py
- [ ] T150 [US2] Implement contradiction handling workflow in backend/python/llm-service/app/services/contradiction_handler.py
- [ ] T151 [US2] Write unit tests for contradiction handler in backend/python/llm-service/tests/test_contradiction_handler.py
- [ ] T152 [US2] Update ConversationService to call memory-processor after each message in backend/golang/conversation-service/internal/biz/conversation_service.go
- [ ] T153 [US2] Write unit tests for memory integration in ConversationService in backend/golang/conversation-service/internal/biz/conversation_service_test.go

### Frontend Memory UI (6 tasks)

- [ ] T154 [P] [US2] Create UserPortrait page component in frontend/app/portrait/page.tsx displaying user profile info
- [ ] T155 [P] [US2] Write unit tests for UserPortrait component in frontend/app/portrait/page.test.tsx
- [ ] T156 [P] [US2] Create MemoryList component in frontend/components/memory/MemoryList.tsx
- [ ] T157 [P] [US2] Write unit tests for MemoryList component in frontend/components/memory/MemoryList.test.tsx
- [ ] T158 [US2] Implement API client for memory service in frontend/lib/api/memory.ts
- [ ] T159 [US2] Write integration tests for memory UI flow in frontend/tests/e2e/memory.spec.ts

### Frontend Multi-Device Sync (FR-005a) (2 tasks)

- [ ] T160 [US1] Implement frontend polling mechanism in frontend/lib/hooks/useMessagePolling.ts with 3-5s interval for multi-device conversation history sync
- [ ] T161 [US1] Write unit tests for polling hook in frontend/lib/hooks/useMessagePolling.test.tsx (test interval timing, error handling, cleanup on unmount, reconnection logic)

### Scheduled Account Deletion (FR-004d) (1 task)

- [ ] T162 [US1] Create scheduled job for automatic account deletion in backend/golang/user-service/internal/job/deletion_scheduler.go (daily cron at 2 AM, deletes accounts where deletion_scheduled_at < NOW() - 30 days, transactional cascade delete across all user tables, logs deletion events for compliance audit)

**Checkpoint**: MVP Core Complete (US1 + US2) - Basic conversation with AI memory system functional

---

## Phase 5: User Story 3-10 (Priority: P2/P3) - Post-MVP Features (30 tasks)

**Purpose**: Placeholder tasks for future user stories implementation

### User Story 3 - 用户自定义AI角色 (P2)

- [ ] T163 [US3] Implement CharacterService CreateCharacter logic in backend/golang/character-service/internal/biz/character_service.go
- [ ] T164 [US3] Write unit tests for CreateCharacter
- [ ] T165 [US3] Implement CharacterService UpdateCharacter logic
- [ ] T166 [US3] Write unit tests for UpdateCharacter
- [ ] T167 [US3] Create character creation wizard UI in frontend/app/characters/create/page.tsx

### User Story 4 - AI角色使用工具能力 (P2)

- [ ] T168 [US4] Implement WeatherTool in backend/python/llm-service/app/services/tools/weather.py
- [ ] T169 [US4] Write unit tests for WeatherTool
- [ ] T170 [US4] Implement SearchTool in backend/python/llm-service/app/services/tools/search.py
- [ ] T171 [US4] Write unit tests for SearchTool
- [ ] T172 [US4] Implement ImageTool in backend/python/llm-service/app/services/tools/image.py
- [ ] T173 [US4] Write unit tests for ImageTool
- [ ] T174 [US4] Integrate tools into LangGraph agent workflow
- [ ] T175 [US4] Write integration tests for tool-enabled agent

### User Story 5 - 积分充值与消费管理 (P2)

- [ ] T176 [US5] Create database migration for credit_accounts and credit_transactions tables
- [ ] T177 [US5] Implement BillingService with credit deduction logic
- [ ] T178 [US5] Write unit tests for BillingService
- [ ] T179 [US5] Implement idempotency for credit transactions using Redis locks
- [ ] T180 [US5] Write unit tests for idempotency logic
- [ ] T181 [US5] Integrate BillingService with ConversationService
- [ ] T182 [US5] Create credit center UI in frontend/app/credits/page.tsx

### User Story 6-10 Placeholders (P3)

- [ ] T183 [US6] Implement emotion state tracking in memory-service
- [ ] T184 [US7] Implement admin service for LLM provider configuration
- [ ] T185 [US8] Implement analytics service for sales CRM
- [ ] T186 [US9] Setup centralized logging with OpenTelemetry
- [ ] T187 [US9] Setup Prometheus metrics and Grafana dashboards
- [ ] T188 [US9] Setup Jaeger for distributed tracing
- [ ] T189 [US9] Create api_logs table and logging middleware
- [ ] T190 [US10] Implement memory compression service
- [ ] T191 [US10] Implement memory importance scoring algorithm
- [ ] T192 [US10] Create background job for memory cleanup

---

## Phase 6: Polish & Production Readiness (15 tasks)

**Purpose**: Improvements that affect multiple user stories

- [ ] T193 [P] Generate Swagger documentation for all REST APIs in docs/api/swagger.yaml
- [ ] T194 [P] Generate Protobuf documentation from proto files in docs/api/grpc.md
- [ ] T195 [P] Create architecture documentation in docs/architecture.md
- [ ] T196 [P] Create deployment documentation in docs/deployment.md
- [ ] T197 [P] Create development guide in docs/development.md
- [ ] T198 [P] Create Kubernetes Helm chart for user-service in deployments/k8s/user-service/
- [ ] T199 [P] Create Kubernetes Helm chart for conversation-service in deployments/k8s/conversation-service/
- [ ] T200 [P] Create Kubernetes Helm chart for memory-service in deployments/k8s/memory-service/
- [ ] T201 [P] Create Kubernetes Helm chart for llm-service in deployments/k8s/llm-service/
- [ ] T202 [P] Create Kubernetes StatefulSet configuration with Snowflake ID worker allocation in deployments/k8s/statefulset.yaml
- [ ] T203 [P] Implement HPA (Horizontal Pod Autoscaler) configuration in deployments/k8s/hpa.yaml
- [ ] T204 [P] Optimize Dockerfiles with multi-stage builds
- [ ] T205 [P] Run security audit with static analysis tools (gosec, bandit)
- [ ] T206 [P] Implement rate limiting middleware using Redis
- [ ] T207 [P] Run quickstart.md validation and update with latest setup instructions

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-4)**: All depend on Foundational phase completion
  - US1 and US2 can proceed in parallel (if staffed)
  - US2 has slight dependency on US1 (needs conversation flow to test memory)
- **Post-MVP (Phase 5)**: Depends on US1+US2 completion, all US3-10 can proceed in parallel
- **Polish (Phase 6)**: Depends on all desired user stories being complete

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

1. Complete Phase 1: Setup (18 tasks)
2. Complete Phase 2: Foundational (45 tasks) - CRITICAL, blocks all stories
3. Complete Phase 3: User Story 1 (57 tasks) - Basic conversation with LLM failover, multi-device sync, and account deletion
4. Complete Phase 4: User Story 2 (42 tasks) - AI memory
5. **STOP and VALIDATE**: Test US1+US2 independently (162 tasks total)
6. Deploy MVP and gather user feedback

### Incremental Delivery (Post-MVP)

1. Complete US1+US2 → Foundation ready (MVP deployed)
2. Add User Story 3 → Test independently → Deploy (Custom characters)
3. Add User Story 4 → Test independently → Deploy (AI tools)
4. Add User Story 5 → Test independently → Deploy (Billing system)
5. Add User Stories 6-10 in priority order
6. Complete Phase 6: Polish → Production-ready

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

**Total Tasks**: 207 tasks

**By Phase**:
- Phase 1 (Setup): 18 tasks
- Phase 2 (Foundational): 45 tasks
- Phase 3 (US1 - Basic Conversation): 57 tasks (includes LLM failover, polling, deletion job)
- Phase 4 (US2 - AI Memory): 42 tasks
- Phase 5 (US3-US10 - Post-MVP): 30 tasks
- Phase 6 (Polish): 15 tasks

**MVP Scope (US1 + US2)**: 162 tasks (Phases 1-4)

**Test Coverage**: Unit test task for every implementation task = 70%+ coverage guaranteed

**Parallel Opportunities**:
- Phase 1: 17 parallelizable tasks
- Phase 2: 25 parallelizable tasks
- Phase 3: 10 parallelizable tasks (entity models)
- Phase 4: 8 parallelizable tasks (entity models)

**Critical Path**:
Setup (Phase 1) → Foundational (Phase 2) → US1 (Phase 3) → US2 (Phase 4) → MVP Complete

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
