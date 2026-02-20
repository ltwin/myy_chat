# Conversation Service

AI companion chat platform conversation service implementation.

## Overview

The Conversation Service handles all chat-related functionality including:
- Creating and managing conversations between users and AI characters
- Sending messages and receiving AI responses via LLM service
- Retrieving conversation history with pagination
- Archiving conversations
- Token counting and credit billing

## Architecture

### Layers

```
┌─────────────────────────────────────────┐
│          gRPC/HTTP Service              │  service/conversation_service.go
├─────────────────────────────────────────┤
│         Business Logic                  │  biz/conversation_service.go
│  - ConversationService                  │  biz/conversation.go
│  - LLM Integration                      │  biz/message.go
│  - Credit Billing                       │
├─────────────────────────────────────────┤
│         Data Access                     │  data/conversation.go
│  - ConversationRepo                     │  data/message.go
│  - MessageRepo                          │
│  - PostgreSQL (partitioned tables)     │
└─────────────────────────────────────────┘
```

## API Endpoints

### gRPC Service

- `CreateConversation` - Create a new conversation with a character
- `GetConversation` - Get conversation details
- `ListConversations` - List user's conversations (paginated, filtered)
- `SendMessage` - Send a user message and receive AI response
- `GetMessages` - Get conversation messages (cursor-based pagination)
- `ArchiveConversation` - Archive/unarchive a conversation

### HTTP REST (via gRPC-Gateway)

- `POST /api/v1/conversations` - Create conversation
- `GET /api/v1/conversations/{id}` - Get conversation
- `GET /api/v1/conversations` - List conversations
- `POST /api/v1/conversations/{id}/messages` - Send message
- `GET /api/v1/conversations/{id}/messages` - Get messages
- `POST /api/v1/conversations/{id}/archive` - Archive conversation

## Database Schema

### Conversations Table (Partitioned by Month)

```sql
CREATE TABLE conversations (
    id BIGINT NOT NULL,               -- Snowflake ID
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    title VARCHAR(255),
    message_count INT DEFAULT 0,
    token_count INT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    last_message_at TIMESTAMP WITH TIME ZONE,
    is_archived BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (id, started_at)
) PARTITION BY RANGE (started_at);
```

### Messages Table (Partitioned by Month)

```sql
CREATE TABLE messages (
    id BIGINT NOT NULL,               -- Snowflake ID
    conversation_id BIGINT NOT NULL,
    role VARCHAR(20) NOT NULL,        -- user, assistant, system
    content TEXT NOT NULL,
    token_count INT NOT NULL,
    metadata JSONB,                   -- {latency_ms, model, cost}
    created_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

## Business Logic

### ConversationService

Core business operations:

1. **CreateConversation**
   - Validates character ID
   - Generates Snowflake ID
   - Creates conversation record

2. **SendMessage**
   - Validates conversation ownership
   - Checks if archived (prevents sending)
   - Checks credit balance
   - Retrieves conversation history (last 10 messages)
   - Calls LLM service via gRPC
   - Saves user + assistant messages atomically
   - Updates conversation counters
   - Deducts credits

3. **GetMessages**
   - Validates ownership
   - Cursor-based pagination (before_message_id)
   - Returns messages in descending order (newest first)

4. **ArchiveConversation**
   - Validates ownership
   - Updates archive status
   - Archived conversations cannot receive new messages

## Dependencies

### External Services

- **LLM Agent Service** (gRPC): AI response generation
- **Billing Service**: Credit balance checks and deductions

### Internal Dependencies

- Snowflake ID Generator
- PostgreSQL with pgx driver
- Kratos v2 framework
- Wire dependency injection

## Testing

### Unit Tests

Run unit tests:
```bash
cd backend
make test-conversation
```

Test files:
- `biz/conversation_test.go` - Conversation entity tests
- `biz/message_test.go` - Message entity tests

### Test Coverage

- Entity creation and validation
- Business rule enforcement (e.g., archived conversations)
- Error handling
- Message role validation
- Token count validation

## Configuration

Example configuration (internal/conf/conf.proto):

```protobuf
message Data {
  message Database {
    string driver = 1;
    string source = 2;
  }
  Database database = 1;
}
```

Example values:
```yaml
data:
  database:
    driver: postgres
    source: postgresql://user:pass@localhost:5432/myy_chat?sslmode=disable
```

## Development

### Running Locally

```bash
cd backend
make run-conversation
```

### Debugging

```bash
make debug-conversation
# Delve debugger listens on :2345
```

### Regenerate Proto

```bash
make proto-conversation
```

## Error Handling

### Business Errors

- `ErrConversationNotFound` - Conversation does not exist
- `ErrConversationAccessDenied` - User doesn't own conversation
- `ErrConversationArchived` - Cannot send to archived conversation
- `ErrInsufficientCredits` - Not enough credits
- `ErrLLMServiceFailure` - LLM service error
- `ErrInvalidMessageContent` - Empty message content
- `ErrInvalidMessageRole` - Invalid role (not user/assistant/system)

### Error Responses

Errors are returned as gRPC status codes:
- `NotFound` - Conversation/message not found
- `PermissionDenied` - Access denied
- `FailedPrecondition` - Archived, insufficient credits
- `Internal` - Database errors, LLM failures

## Performance Considerations

### Database Optimization

1. **Partitioning**: Monthly partitions reduce query scan size
2. **Indexes**:
   - `conversations(user_id, last_message_at DESC)` - List conversations
   - `messages(conversation_id, created_at DESC)` - Get messages
3. **Atomic Updates**: `IncrementCounts` uses single UPDATE
4. **Batch Inserts**: User + assistant messages saved in transaction

### Caching Strategy (Future)

- Cache recent conversation metadata in Redis
- Cache last N messages per conversation
- TTL: 1 hour

### Query Optimization

- Limit history to last 10 messages for LLM context
- Cursor pagination (more efficient than offset)
- Partition pruning via date ranges

## Integration Points

### LLM Agent Service

```go
type LLMClient interface {
    Chat(ctx context.Context, messages []ChatMessage, model string) (*ChatResponse, error)
}
```

### Credit Service

```go
type CreditService interface {
    DeductCredits(ctx context.Context, userID int64, amount float64, reason string) error
    GetBalance(ctx context.Context, userID int64) (float64, error)
}
```

## Future Enhancements

1. **Streaming Responses**: SSE for real-time AI responses
2. **Context Management**: Automatic summarization for long conversations
3. **Multi-modal**: Support image/audio messages
4. **Conversation Search**: Full-text search on messages
5. **Conversation Sharing**: Share conversation links
6. **Export**: Export conversation as PDF/Markdown

## Migration

Migrations are located in `/backend/migrations/`:
- `000_init.sql` - Database initialization (functions, extensions)
- `004_create_conversations_table.sql` - Conversations table + partitions
- `005_create_messages_table.sql` - Messages table + partitions

Run migrations:
```bash
# Using your migration tool
migrate -path ./migrations -database "postgresql://..." up
```

## Monitoring

### Metrics (TODO)

- Conversation creation rate
- Message send rate
- LLM latency (p50, p95, p99)
- Token usage per conversation
- Error rates

### Logging

Logs include:
- Conversation creation: `conversation created: id=..., user_id=..., character_id=...`
- Message sent: `message sent: conversation_id=..., tokens=..., cost=..., latency=...ms`
- Errors: All errors logged with context

## License

Copyright 2025 myy_chat
