-- 010_memory_extensions.sql
-- Memory claim/conflict/audit extensions.

BEGIN;

CREATE TABLE IF NOT EXISTS memory_claims (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    source_message_id BIGINT,
    client_message_id VARCHAR(64),
    source_type VARCHAR(24) NOT NULL DEFAULT 'EXTRACTED',
    claim_key VARCHAR(128) NOT NULL,
    claim_value TEXT NOT NULL,
    claim_scope VARCHAR(32) NOT NULL DEFAULT 'PROFILE',
    field_tier VARCHAR(8) NOT NULL DEFAULT 'TIER_C',
    confidence_score NUMERIC(4,3) NOT NULL DEFAULT 0.500,
    rank_score NUMERIC(6,3) NOT NULL DEFAULT 0.500,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    linked_memory_id BIGINT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMPTZ,
    CONSTRAINT chk_memory_claims_source_type CHECK (source_type IN ('EXTRACTED','INFERRED','USER_STATED')),
    CONSTRAINT chk_memory_claims_scope CHECK (claim_scope IN ('PROFILE','EVENT','RELATION','PREFERENCE')),
    CONSTRAINT chk_memory_claims_status CHECK (status IN ('PENDING','CONFLICTING','CLARIFYING','CONFIRMED','REJECTED','MERGED')),
    CONSTRAINT chk_memory_claims_field_tier CHECK (field_tier IN ('TIER_A','TIER_B','TIER_C')),
    CONSTRAINT chk_memory_claims_valid_window CHECK (valid_until IS NULL OR valid_from IS NULL OR valid_until > valid_from)
);

CREATE INDEX IF NOT EXISTS idx_memory_claims_user_char_key
    ON memory_claims(user_id, character_id, claim_key, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_claims_status
    ON memory_claims(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_claims_tier_status
    ON memory_claims(field_tier, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_claims_temporal
    ON memory_claims(user_id, character_id, claim_key, valid_until, valid_from DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uq_memory_claims_dedup
    ON memory_claims(user_id, character_id, claim_key, client_message_id)
    WHERE client_message_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_memory_claims_client_message
    ON memory_claims(client_message_id)
    WHERE client_message_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS memory_claim_conflicts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    claim_key VARCHAR(128) NOT NULL,
    incumbent_claim_id BIGINT NOT NULL,
    incoming_claim_id BIGINT NOT NULL,
    conflict_type VARCHAR(32) NOT NULL DEFAULT 'VALUE_MISMATCH',
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    clarification_prompt TEXT,
    resolution_note TEXT,
    resolved_by_message_id BIGINT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_memory_claim_conflicts_type CHECK (conflict_type IN ('VALUE_MISMATCH','MUTEX_FIELD','AMBIGUOUS_APPEND')),
    CONSTRAINT chk_memory_claim_conflicts_status CHECK (status IN ('OPEN','CLARIFYING','RESOLVED_REPLACED','RESOLVED_APPENDED','RESOLVED_REJECTED','CANCELLED')),
    UNIQUE (incoming_claim_id)
);

CREATE INDEX IF NOT EXISTS idx_memory_claim_conflicts_user_time
    ON memory_claim_conflicts(user_id, character_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_claim_conflicts_status
    ON memory_claim_conflicts(status, updated_at DESC);

CREATE TABLE IF NOT EXISTS memory_audit_events (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT,
    memory_id BIGINT,
    event_type VARCHAR(64) NOT NULL,
    event_level VARCHAR(16) NOT NULL DEFAULT 'INFO',
    source_service VARCHAR(32) NOT NULL,
    request_id VARCHAR(64),
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_memory_audit_level CHECK (event_level IN ('INFO','WARN','ERROR')),
    CONSTRAINT chk_memory_audit_type CHECK (
        event_type IN (
            'ACCESS_DENIED',
            'CONFLICT_CLARIFICATION',
            'CONFLICT_STATE_TRANSITION',
            'CLAIM_PIPELINE_DLQ',
            'GRAPH_SYNC_FAILED',
            'MEMORY_EDIT',
            'MEMORY_DELETE'
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_memory_audit_user_time
    ON memory_audit_events(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_audit_type_time
    ON memory_audit_events(event_type, created_at DESC);

COMMIT;
