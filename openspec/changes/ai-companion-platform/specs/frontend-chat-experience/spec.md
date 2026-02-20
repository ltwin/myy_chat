## ADDED Requirements

### Requirement: Runtime API Mode Switching
Frontend API client MUST support environment-driven switching between mock and real backend modes.

#### Scenario: Environment mode selects API endpoint strategy
- **WHEN** `VITE_API_MODE` is configured as `mock` or `real`
- **THEN** frontend API layer MUST route requests to the corresponding provider without code changes

### Requirement: Auth Token Injection and Recovery
Frontend request pipeline MUST attach JWT to authenticated API requests and handle unauthorized responses consistently.

#### Scenario: Expired token triggers recovery action
- **WHEN** backend returns unauthorized response for an authenticated request
- **THEN** client MUST clear invalid session or perform refresh flow according to configured auth policy

#### Scenario: Access token is memory-only
- **WHEN** user completes login and obtains a valid access token
- **THEN** frontend MUST keep access token only in runtime memory and MUST NOT persist it into `localStorage` or `sessionStorage`

#### Scenario: Refresh flow relies on HttpOnly cookie
- **WHEN** frontend needs to refresh an expired access token
- **THEN** client MUST call refresh API without reading refresh token from JavaScript, and rely on browser-attached HttpOnly cookie for refresh authentication

#### Scenario: CSRF token is attached for cookie-auth write endpoints
- **WHEN** frontend calls refresh/logout/logout-all endpoints that depend on cookie authentication
- **THEN** request MUST include configured CSRF protection signal (for example `X-CSRF-Token`) according to backend policy

### Requirement: Streaming Chat Rendering Feedback
Chat UI MUST render assistant output incrementally and present explicit loading/timeout states.

#### Scenario: Streaming response updates visible assistant content
- **WHEN** streaming chunks are received
- **THEN** message UI MUST append content incrementally and preserve text order

#### Scenario: Slow response displays staged loading hints
- **WHEN** assistant response has not completed at 10s and 30s thresholds
- **THEN** UI MUST show staged status hints and provide retry guidance at timeout threshold

### Requirement: Credit-Aware Send Guard
Frontend send-message action MUST reflect current balance and block sends when balance is insufficient.

#### Scenario: Zero balance prevents send action
- **WHEN** current user credit balance is zero
- **THEN** send input action MUST be blocked and UI MUST display recharge guidance

### Requirement: Basic Memory Visibility UI
Frontend MUST expose a memory list view that displays stored user memory summaries from memory APIs.

#### Scenario: Memory list renders backend results
- **WHEN** memory API returns records for current user context
- **THEN** MemoryList UI MUST render those records and handle empty-state presentation
