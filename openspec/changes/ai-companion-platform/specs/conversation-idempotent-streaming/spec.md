## ADDED Requirements

### Requirement: Client Message Id Idempotency Contract
The send-message API MUST accept `client_message_id` as a required client-generated idempotency key, and server MUST guarantee at-most-once business effects per unique key.

#### Scenario: Retries with same key do not duplicate business effects
- **WHEN** a client retries a send-message request with the same `client_message_id`
- **THEN** system MUST return the previously computed result and MUST NOT duplicate message storage or billing transactions

### Requirement: Dual Identifier Persistence
Conversation storage MUST maintain both server-generated `message.id` (Snowflake BIGINT) and client-generated `client_message_id` (string unique key).

#### Scenario: Message persistence keeps both identities
- **WHEN** a user message is persisted
- **THEN** database MUST store `message.id` as primary key and MUST enforce uniqueness of `client_message_id` for idempotency

### Requirement: Streaming Response with Bounded First Token Time
Conversation service MUST stream assistant responses incrementally and target TTFT (time to first token) within 3 seconds under normal operation.

#### Scenario: Successful LLM response is streamed incrementally
- **WHEN** LLM backend returns streaming chunks
- **THEN** server MUST forward chunks in order and emit final completion state for the client

#### Scenario: Streaming timeout yields explicit failure outcome
- **WHEN** TTFT threshold is exceeded or upstream stream fails
- **THEN** server MUST return an explicit timeout/error response and trigger billing compensation flow if reserve exists

### Requirement: Reserve Compensation on Downstream Failure
If reserve succeeds but downstream processing fails before settle, conversation flow MUST invoke release to restore user balance.

#### Scenario: Post-reserve failure triggers release
- **WHEN** reserve is successful and LLM call fails prior to settlement
- **THEN** system MUST call `ReleaseCredits` with the existing reserve identifier and record the failure path outcome

