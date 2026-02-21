-- 009_long_term_memory_core.sql
-- Long-term memory core schema.

BEGIN;

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS memories (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    conversation_id BIGINT,
    source_message_id BIGINT,
    client_message_id VARCHAR(64),
    source_type VARCHAR(24) NOT NULL DEFAULT 'EXTRACTED',
    target_table VARCHAR(64),
    target_id BIGINT,
    memory_type VARCHAR(32) NOT NULL,
    content TEXT NOT NULL,
    embedding VECTOR(1536),
    importance_score NUMERIC(5,2) NOT NULL DEFAULT 50.00,
    decay_factor NUMERIC(5,2) NOT NULL DEFAULT 1.00,
    boost_count INT NOT NULL DEFAULT 0,
    access_count INT NOT NULL DEFAULT 0,
    visibility_scope VARCHAR(32) NOT NULL DEFAULT 'OWNER_PRIVATE',
    actor_user_id BIGINT,
    is_mutable BOOLEAN NOT NULL DEFAULT TRUE,
    last_accessed_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT chk_memories_type CHECK (memory_type IN ('PROFILE','EMOTION','EVENT','SUMMARY','ENTITY','RELATION','CORE')),
    CONSTRAINT chk_memories_source_type CHECK (source_type IN ('EXTRACTED','INFERRED','USER_STATED')),
    CONSTRAINT chk_memories_visibility CHECK (visibility_scope IN ('OWNER_PRIVATE','PUBLIC','VISITOR_PRIVATE')),
    CONSTRAINT chk_memories_target_ref CHECK (
        (target_table IS NULL AND target_id IS NULL)
        OR (target_table IS NOT NULL AND target_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_memories_user_character
    ON memories(user_id, character_id, created_at DESC)
    WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_memories_visibility_priority
    ON memories(user_id, character_id, visibility_scope, importance_score DESC, updated_at DESC)
    WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_memories_source_message
    ON memories(source_message_id)
    WHERE source_message_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_memories_client_message
    ON memories(client_message_id)
    WHERE client_message_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_memories_target_ref
    ON memories(target_table, target_id)
    WHERE target_table IS NOT NULL AND target_id IS NOT NULL;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN
        BEGIN
            CREATE INDEX IF NOT EXISTS idx_memories_embedding_hnsw
                ON memories USING hnsw (embedding vector_cosine_ops)
                WITH (m = 16, ef_construction = 64);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'skip hnsw index for memories: %', SQLERRM;
        END;
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS conversation_summaries (
    id BIGINT PRIMARY KEY,
    conversation_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    start_message_id BIGINT,
    end_message_id BIGINT,
    summary TEXT NOT NULL,
    summary_tokens INT,
    compression_ratio NUMERIC(5,2),
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversation_summaries_conv
    ON conversation_summaries(conversation_id, created_at DESC);

CREATE TABLE IF NOT EXISTS user_portraits (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    core_summary TEXT,
    demographics JSONB,
    preferences JSONB,
    persona_traits JSONB,
    current_occupation VARCHAR(100),
    career_history JSONB NOT NULL DEFAULT '[]'::jsonb,
    confidence_score NUMERIC(3,2) NOT NULL DEFAULT 0.50,
    source_memory_ids BIGINT[],
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_user_portraits_demographics CHECK (
        demographics IS NULL OR jsonb_typeof(demographics) = 'object'
    ),
    CONSTRAINT chk_user_portraits_career_history CHECK (
        jsonb_typeof(career_history) = 'array'
    ),
    UNIQUE (user_id, character_id)
);

CREATE TABLE IF NOT EXISTS emotional_states (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    conversation_id BIGINT,
    message_id BIGINT,
    valence NUMERIC(4,3),
    arousal NUMERIC(4,3),
    dominance NUMERIC(4,3),
    primary_emotion VARCHAR(50),
    intensity NUMERIC(4,3),
    confidence_score NUMERIC(4,3) NOT NULL DEFAULT 0.500,
    metadata JSONB,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_emotional_valence CHECK (valence IS NULL OR (valence >= -1 AND valence <= 1)),
    CONSTRAINT chk_emotional_arousal CHECK (arousal IS NULL OR (arousal >= 0 AND arousal <= 1)),
    CONSTRAINT chk_emotional_dominance CHECK (dominance IS NULL OR (dominance >= 0 AND dominance <= 1)),
    CONSTRAINT chk_emotional_primary_emotion CHECK (
        primary_emotion IS NULL OR primary_emotion IN ('JOY','TRUST','FEAR','SURPRISE','SADNESS','DISGUST','ANGER','ANTICIPATION','NEUTRAL')
    ),
    CONSTRAINT chk_emotional_signal_exists CHECK (primary_emotion IS NOT NULL OR valence IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_emotional_states_user_char_time
    ON emotional_states(user_id, character_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_emotional_states_conv_time
    ON emotional_states(conversation_id, recorded_at DESC)
    WHERE conversation_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS important_events (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    message_id BIGINT,
    event_type VARCHAR(32),
    event_summary TEXT NOT NULL,
    temporal_context VARCHAR(255),
    importance NUMERIC(5,2) NOT NULL DEFAULT 50.00,
    occurred_at TIMESTAMPTZ NOT NULL,
    expiry_at TIMESTAMPTZ,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_important_events_status CHECK (status IN ('ACTIVE','ARCHIVED','DELETED'))
);

CREATE INDEX IF NOT EXISTS idx_important_events_user_char_time
    ON important_events(user_id, character_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_important_events_importance
    ON important_events(user_id, character_id, importance DESC);

CREATE TABLE IF NOT EXISTS memory_graph_nodes (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    node_type VARCHAR(32) NOT NULL,
    node_name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL,
    aliases TEXT[] NOT NULL DEFAULT '{}',
    node_properties JSONB,
    source_memory_id BIGINT,
    sync_status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    sync_version INT NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_memory_graph_nodes_sync CHECK (sync_status IN ('PENDING','SYNCED','FAILED')),
    UNIQUE (user_id, character_id, node_type, normalized_name)
);

CREATE INDEX IF NOT EXISTS idx_memory_graph_nodes_normalized_name
    ON memory_graph_nodes(user_id, character_id, normalized_name);
CREATE INDEX IF NOT EXISTS idx_memory_graph_nodes_aliases_gin
    ON memory_graph_nodes USING gin (aliases);

CREATE TABLE IF NOT EXISTS memory_graph_edges (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    character_id BIGINT NOT NULL,
    from_node_id BIGINT NOT NULL,
    to_node_id BIGINT NOT NULL,
    relation_type VARCHAR(64) NOT NULL,
    weight NUMERIC(5,2) NOT NULL DEFAULT 1.00,
    confidence NUMERIC(4,3) NOT NULL DEFAULT 0.500,
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    source_memory_id BIGINT,
    sync_status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    sync_version INT NOT NULL DEFAULT 0,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_memory_graph_edges_sync CHECK (sync_status IN ('PENDING','SYNCED','FAILED')),
    CONSTRAINT chk_memory_graph_edges_confidence CHECK (confidence >= 0 AND confidence <= 1),
    CONSTRAINT chk_memory_graph_edges_valid_time CHECK (valid_until IS NULL OR valid_from IS NULL OR valid_until > valid_from),
    UNIQUE (user_id, character_id, from_node_id, to_node_id, relation_type, valid_from)
);

CREATE INDEX IF NOT EXISTS idx_memory_graph_edges_from ON memory_graph_edges(from_node_id);
CREATE INDEX IF NOT EXISTS idx_memory_graph_edges_to ON memory_graph_edges(to_node_id);
CREATE INDEX IF NOT EXISTS idx_memory_graph_edges_active
    ON memory_graph_edges(user_id, character_id, relation_type, valid_until);

COMMIT;
