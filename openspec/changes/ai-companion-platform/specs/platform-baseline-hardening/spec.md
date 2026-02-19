## ADDED Requirements

### Requirement: Unified Health Check Contract
All externally exposed service health endpoints MUST use the unified path `/api/v1/health` across application code, local compose probes, and Kubernetes probe configuration.

#### Scenario: Health path is consistent across runtime layers
- **WHEN** a service is started in Docker Compose and in Kubernetes
- **THEN** liveness and readiness checks MUST both target `/api/v1/health` and return success for healthy instances

### Requirement: Runtime Baseline Version Alignment
Go services MUST use a single runtime baseline where `go.mod` and Docker build images are aligned to Go `1.24.x` to avoid build/runtime drift.

#### Scenario: Build pipeline validates Go version consistency
- **WHEN** CI executes build steps for a Go service
- **THEN** the pipeline MUST fail if Dockerfile Go version and `go.mod` Go version are not aligned to the same major/minor baseline

### Requirement: Idempotent and Monotonic Migrations
Database migrations MUST be repeatable and non-conflicting: SQL scripts MUST use idempotent guards (`IF NOT EXISTS` or equivalent), and new migration numbers MUST be strictly monotonic relative to existing files.

#### Scenario: Re-running migrations does not introduce errors
- **WHEN** `migrate up` is executed repeatedly on an already migrated environment
- **THEN** execution MUST succeed without duplicate-object errors and without changing already-correct schema state

#### Scenario: New migration numbering avoids collisions
- **WHEN** a new migration is added after existing `000`-`006` migrations
- **THEN** the new files MUST use subsequent numbers (`007+`) and MUST NOT overwrite semantics of existing migration IDs

