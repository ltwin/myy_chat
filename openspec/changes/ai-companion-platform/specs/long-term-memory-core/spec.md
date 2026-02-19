## ADDED Requirements

### Requirement: Five-Layer Memory Architecture
System MUST implement a five-layer long-term memory architecture composed of conversation summaries (Layer 1), emotional states (Layer 2), important events (Layer 3), graph memory nodes/edges (Layer 4), and user portraits (Layer 5).

#### Scenario: Memory write pipeline updates multi-layer structures
- **WHEN** a new user-assistant interaction is finalized
- **THEN** system MUST process the interaction through configured memory pipeline and persist relevant updates to one or more target layers

### Requirement: Persona-Aligned Memory Conflict Clarification
When new user information conflicts with existing memory, system MUST pause direct overwrite and initiate a persona-aligned clarification turn before committing the update.

#### Scenario: Contradictory profile information triggers clarification
- **WHEN** user provides profile info that contradicts stored memory facts
- **THEN** assistant MUST ask for confirmation in current role persona and MUST NOT expose internal system wording such as “conflict detected”

#### Scenario: Conflict resolution runs through asynchronous claim state machine
- **WHEN** new memory claims are extracted from finalized conversation messages
- **THEN** system MUST process claim state transitions `PENDING → CONFLICTING → CLARIFYING → (CONFIRMED | REJECTED | MERGED)` and related conflict transitions `OPEN → CLARIFYING → RESOLVED*` in memory-service async pipeline, and MUST NOT block SendMessage primary response

### Requirement: AI-First Profile Capture With Field Tiering
User portrait acquisition MUST default to AI extraction from conversation, and profile fields MUST be managed by tiered rules rather than one-size-fits-all overwrite.

#### Scenario: Auto extraction does not directly overwrite canonical portrait
- **WHEN** assistant extracts new profile facts from conversation
- **THEN** system MUST create claims first and MUST NOT directly overwrite canonical single-value profile fields before conflict policy resolution

#### Scenario: Critical single-value fields require lightweight confirmation on conflict
- **WHEN** a critical single-value field (e.g., `full_name`, `birth_date`, `current_occupation`) receives contradictory value
- **THEN** system MUST trigger persona-aligned clarification and update canonical field only after explicit resolution

#### Scenario: Multi-value fields allow coexistence
- **WHEN** extracted facts belong to multi-value profile fields (e.g., interests, skills, role tags)
- **THEN** system MUST allow coexistence and maintain ranking metadata instead of forcing replacement

#### Scenario: Occupation change preserves temporal history
- **WHEN** user role changes across time (e.g., developer → product manager)
- **THEN** system MUST preserve historical role records and update current role projection instead of deleting prior facts

### Requirement: Dual Profile SoT Boundary
Account profile storage and conversational portrait storage MUST have explicit service ownership and non-destructive synchronization rules.

#### Scenario: Manual profile edit enters claim pipeline as high-trust source
- **WHEN** user updates profile fields in account settings (`user_profiles`)
- **THEN** system MUST emit corresponding claim with `USER_STATED` priority and resolve through memory-service policy before portrait canonical update

#### Scenario: Portrait canonical update does not overwrite account settings by default
- **WHEN** memory-service confirms portrait facts in `user_portraits`
- **THEN** system MUST NOT auto-overwrite manually maintained account fields in `user_profiles` unless explicit sync policy is configured

### Requirement: MCP Gateway Access Layer for Agent Tools
Agent-side tool calls in current change scope (profile tools) MUST be exposed through `mcp-service` gateway. For portrait CRUD, memory-service MUST remain the only data source of truth for persisted profile state.

#### Scenario: Authorized tool discovery is scope-filtered
- **WHEN** agent requests available MCP tools
- **THEN** mcp-service MUST return only tools allowed by caller identity and granted scopes

#### Scenario: Agent uses MCP gateway tools for portrait operations
- **WHEN** agent needs portrait read/write actions
- **THEN** it MUST call MCP tools (`profile.get`, `profile.claim_upsert`, `profile.conflict_list`, `profile.conflict_resolve`, `profile.fact_delete`, `profile.audit_query`) rather than direct storage access

#### Scenario: Domain plugin routing dispatches to target service
- **WHEN** agent invokes a registered tool by `tool_name`
- **THEN** mcp-service MUST route request via domain registry to corresponding downstream microservice API

#### Scenario: MCP profile write path enforces SoT boundary
- **WHEN** mcp-service handles profile write intent
- **THEN** it MUST route persistence through memory-service APIs and MUST NOT bypass service boundary with direct database mutations

#### Scenario: Server-side identity context overrides client-supplied identity
- **WHEN** tool request payload includes user/tenant identifiers
- **THEN** mcp-service MUST use server-side trusted context for authorization and downstream calls

#### Scenario: Unauthorized tool invocation is blocked
- **WHEN** caller identity or scope is insufficient for target profile mutation
- **THEN** mcp-service MUST reject request and emit auditable security event

#### Scenario: Unregistered tool call is denied by default
- **WHEN** caller invokes an unknown or unregistered tool
- **THEN** mcp-service MUST reject request with auditable denial and MUST NOT forward request to downstream services

### Requirement: Memory Visibility Isolation
Memory retrieval and prompt injection MUST enforce visibility scopes, at minimum `Owner-Private` and `Public`, so private memories never leak to unauthorized context.

#### Scenario: Visitor cannot retrieve owner-private memories
- **WHEN** a non-owner request attempts to access private memory context
- **THEN** retrieval result MUST exclude owner-private records and prompt assembly MUST inject only authorized memory snippets

#### Scenario: Visibility filter is applied before rerank
- **WHEN** retrieval stage returns mixed-scope memory candidates
- **THEN** system MUST hard-filter unauthorized candidates before rerank and prompt assembly, rather than treating visibility as a ranking factor

### Requirement: Temporal Compression and Priority Adjustment
System MUST compress transient dialogue context when context usage crosses configured threshold and MUST adjust memory priority based on recency, importance, and reinforcement signals.

#### Scenario: Context pressure triggers memory compression
- **WHEN** conversation context usage reaches configured compression threshold
- **THEN** system MUST summarize transient context into structured memory artifacts and preserve key facts for future retrieval

#### Scenario: Priority model decays unimportant old memories
- **WHEN** memory maintenance job runs on aging memory records
- **THEN** lower-value stale memories MUST decay in retrieval priority while high-importance memories remain retrievable

### Requirement: Graph Memory Synchronization
Graph memory metadata in PostgreSQL MUST synchronize to internal graph projection representation with retryable and idempotent sync semantics.

#### Scenario: Unsynced graph node/edge is retried safely
- **WHEN** graph sync for node or edge fails transiently
- **THEN** background sync worker MUST retry with idempotent semantics and update sync status fields on success or tracked failure

### Requirement: Temporal Relationship Evolution
For graph relationships that can change over time, system MUST preserve temporal history using `valid_from` and `valid_until` semantics instead of destructive overwrite.

#### Scenario: Relationship changes without losing history
- **WHEN** an existing relation (e.g. friend/enemy/blocked) changes based on new evidence
- **THEN** system MUST close previous edge validity (`valid_until`) and persist new edge validity window (`valid_from`) so historical relationship can still be queried

### Requirement: User Memory Management APIs
System MUST provide user-facing memory management capabilities to list memory profile and allow controlled edit/delete operations on eligible memory entries.

#### Scenario: User lists remembered profile facts
- **WHEN** user opens memory list view
- **THEN** system MUST return current user-memory profile entries with visibility-safe projection

#### Scenario: User edits or removes memory entry
- **WHEN** user submits memory correction or deletion for a mutable memory item
- **THEN** system MUST persist change with audit metadata and reflect update in subsequent retrieval
