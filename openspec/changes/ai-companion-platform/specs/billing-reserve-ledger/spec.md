## ADDED Requirements

### Requirement: Billing Service as Ledger Source of Truth
All credit balance mutations MUST be executed through `billing-service` APIs. Other services MUST NOT directly write `credit_accounts` or `credit_transactions`.

#### Scenario: Cross-service credit mutation is routed through billing
- **WHEN** user registration, message consumption, timeout compensation, or recharge changes user credits
- **THEN** the caller MUST invoke billing APIs and MUST NOT perform direct SQL writes to ledger tables from non-billing services

### Requirement: Reserve-Settle-Release Workflow
Billing MUST provide a reserve-first charging workflow with explicit operations for `GrantCredits`, `ReserveCredits`, `SettleCredits`, and `ReleaseCredits`.

#### Scenario: Successful message flow performs reserve and settle
- **WHEN** a message is accepted for LLM processing
- **THEN** system MUST reserve estimated credits before LLM call and MUST settle final amount after model response

#### Scenario: Failed message flow releases reserved credits
- **WHEN** LLM call fails or times out after reserve succeeds
- **THEN** system MUST release the reserved credits and record a release transaction

### Requirement: Idempotent Reserve Keying
Reserve and related ledger actions MUST be idempotent by `client_message_id` so retries do not cause duplicate charging.

#### Scenario: Duplicate reserve request returns prior reservation
- **WHEN** the same `client_message_id` is submitted repeatedly
- **THEN** `ReserveCredits` MUST return the existing reservation result and MUST NOT create duplicate reserve transactions

### Requirement: Compatibility During API Evolution
During migration from legacy direct charge APIs, billing proto MUST support new reserve workflow while preserving temporary compatibility for existing `DeductCredits` and `AddCredits` callers.

#### Scenario: Legacy and new callers can coexist during rollout
- **WHEN** one caller uses `ReserveCredits/SettleCredits` and another still uses legacy APIs
- **THEN** billing service MUST keep both API families callable and maintain consistent ledger invariants during the compatibility window

