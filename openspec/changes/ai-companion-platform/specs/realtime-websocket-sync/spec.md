## ADDED Requirements

### Requirement: Authenticated WebSocket Session
WebSocket connections for conversation sync MUST be authenticated with JWT at connection establishment.

#### Scenario: Unauthorized connection is rejected
- **WHEN** a client connects without a valid JWT
- **THEN** server MUST reject or close the connection with an authorization error

#### Scenario: Revoked session token is rejected immediately
- **WHEN** JWT is cryptographically valid but its `sid/jti` is present in server-side blacklist
- **THEN** server MUST reject the WebSocket connection (or close it immediately) with an authorization error

#### Scenario: Token in URL query is not accepted as primary auth channel
- **WHEN** client attempts to pass JWT via WebSocket URL query parameter
- **THEN** server MUST reject or ignore that token source according to security policy, and require authenticated handshake via approved channel

#### Scenario: Active websocket is terminated after session revocation
- **WHEN** a connected client session is revoked by logout or logout-all
- **THEN** server MUST proactively close all WebSocket connections bound to that revoked `sid`

### Requirement: Heartbeat and Reconnect Safety
Realtime channel MUST implement heartbeat and reconnect semantics to keep session liveness predictable.

#### Scenario: Lost heartbeat is detected and disconnected
- **WHEN** ping/pong heartbeat fails within configured interval
- **THEN** server MUST close stale connection and client MUST transition to reconnect flow

#### Scenario: Client reconnect uses exponential backoff
- **WHEN** network interruption occurs
- **THEN** client reconnect strategy MUST use exponential backoff with bounded maximum delay

### Requirement: Multi-Device Consistency
A single user MAY have multiple concurrent device connections, and updates MUST be broadcast to all active sessions for that user.

#### Scenario: Message event fan-out reaches all active devices
- **WHEN** one device sends a new message and server generates updates
- **THEN** server MUST push corresponding events to every active connection associated with the same user

### Requirement: Ordered Streaming Event Delivery
Streaming message chunks and completion events MUST preserve causal order per message.

#### Scenario: Chunk order is stable per message stream
- **WHEN** assistant response is emitted over websocket
- **THEN** chunk events MUST be delivered in generation order before final completion event
