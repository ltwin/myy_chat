-- 012_schema_consistency_hardening.sql
-- Close remaining schema gaps found by audit:
-- 1) users/characters NOT NULL + active-unique username
-- 2) memory tables missing constraints/indexes
-- 3) memory_graph_edges uniqueness for NULL valid_from

BEGIN;

-- ================================================================
-- 1. users / characters hardening
-- ================================================================
UPDATE users SET email_verified = FALSE WHERE email_verified IS NULL;
UPDATE users SET is_deleted = FALSE WHERE is_deleted IS NULL;

ALTER TABLE users
    ALTER COLUMN email_verified SET DEFAULT FALSE,
    ALTER COLUMN email_verified SET NOT NULL,
    ALTER COLUMN is_deleted SET DEFAULT FALSE,
    ALTER COLUMN is_deleted SET NOT NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_username_key'
          AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users DROP CONSTRAINT users_username_key;
    END IF;
END
$$;

DROP INDEX IF EXISTS idx_users_username;
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_username_active
    ON users(username)
    WHERE NOT is_deleted;

UPDATE characters SET is_public = FALSE WHERE is_public IS NULL;
UPDATE characters SET is_preset = FALSE WHERE is_preset IS NULL;
UPDATE characters SET version = 1 WHERE version IS NULL;
UPDATE characters SET is_deleted = FALSE WHERE is_deleted IS NULL;
UPDATE characters SET world_view = '{}'::jsonb WHERE world_view IS NULL;

ALTER TABLE characters
    ALTER COLUMN world_view SET DEFAULT '{}'::jsonb,
    ALTER COLUMN world_view SET NOT NULL,
    ALTER COLUMN is_public SET DEFAULT FALSE,
    ALTER COLUMN is_public SET NOT NULL,
    ALTER COLUMN is_preset SET DEFAULT FALSE,
    ALTER COLUMN is_preset SET NOT NULL,
    ALTER COLUMN version SET DEFAULT 1,
    ALTER COLUMN version SET NOT NULL,
    ALTER COLUMN is_deleted SET DEFAULT FALSE,
    ALTER COLUMN is_deleted SET NOT NULL;

-- ================================================================
-- 2. memory schema hardening
-- ================================================================
CREATE INDEX IF NOT EXISTS idx_memories_conversation_time
    ON memories(conversation_id, created_at DESC)
    WHERE conversation_id IS NOT NULL AND NOT is_deleted;

CREATE INDEX IF NOT EXISTS idx_conversation_summaries_user_char_time
    ON conversation_summaries(user_id, character_id, created_at DESC);

UPDATE user_portraits
SET confidence_score = 0.50
WHERE confidence_score < 0 OR confidence_score > 1 OR confidence_score IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_user_portraits_confidence'
          AND conrelid = 'user_portraits'::regclass
    ) THEN
        ALTER TABLE user_portraits
            ADD CONSTRAINT chk_user_portraits_confidence
            CHECK (confidence_score >= 0 AND confidence_score <= 1);
    END IF;
END
$$;

UPDATE emotional_states
SET intensity = NULL
WHERE intensity IS NOT NULL AND (intensity < 0 OR intensity > 1);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_emotional_intensity'
          AND conrelid = 'emotional_states'::regclass
    ) THEN
        ALTER TABLE emotional_states
            ADD CONSTRAINT chk_emotional_intensity
            CHECK (intensity IS NULL OR (intensity >= 0 AND intensity <= 1));
    END IF;
END
$$;

UPDATE important_events
SET event_type = 'OTHER'
WHERE event_type IS NOT NULL
  AND event_type NOT IN (
      'MILESTONE',
      'RELATIONSHIP',
      'PREFERENCE',
      'GOAL',
      'PROMISE',
      'SCHEDULE',
      'HEALTH',
      'WORK',
      'OTHER'
  );

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_important_events_type'
          AND conrelid = 'important_events'::regclass
    ) THEN
        ALTER TABLE important_events
            ADD CONSTRAINT chk_important_events_type
            CHECK (
                event_type IS NULL OR event_type IN (
                    'MILESTONE',
                    'RELATIONSHIP',
                    'PREFERENCE',
                    'GOAL',
                    'PROMISE',
                    'SCHEDULE',
                    'HEALTH',
                    'WORK',
                    'OTHER'
                )
            );
    END IF;
END
$$;

WITH ranked_null_edges AS (
    SELECT
        ctid,
        row_number() OVER (
            PARTITION BY user_id, character_id, from_node_id, to_node_id, relation_type
            ORDER BY created_at DESC, id DESC
        ) AS rn
    FROM memory_graph_edges
    WHERE valid_from IS NULL
)
DELETE FROM memory_graph_edges e
USING ranked_null_edges r
WHERE e.ctid = r.ctid
  AND r.rn > 1;

DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT c.conname
    INTO constraint_name
    FROM pg_constraint c
    WHERE c.conrelid = 'memory_graph_edges'::regclass
      AND c.contype = 'u'
      AND pg_get_constraintdef(c.oid) ILIKE '%(user_id, character_id, from_node_id, to_node_id, relation_type, valid_from)%'
    LIMIT 1;

    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE memory_graph_edges DROP CONSTRAINT %I', constraint_name);
    END IF;
END
$$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_memory_graph_edges_nonnull_valid_from
    ON memory_graph_edges(user_id, character_id, from_node_id, to_node_id, relation_type, valid_from)
    WHERE valid_from IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_memory_graph_edges_null_valid_from
    ON memory_graph_edges(user_id, character_id, from_node_id, to_node_id, relation_type)
    WHERE valid_from IS NULL;

COMMIT;
