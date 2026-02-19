## ADDED Requirements

### Requirement: Composite Memory Scoring
Memory retrieval MUST compute `final_score` using a composite model that includes base importance, time decay, reinforcement, novelty, and topology signals.

#### Scenario: Retrieval uses composite score
- **WHEN** retrieval receives candidate memories with `importance_score`, `decay_factor`, `memory_strength`, `surprise_score`, and `connectivity`
- **THEN** system MUST compute and rank candidates by a deterministic composite `final_score`

### Requirement: Advanced Scoring Fields Persistence
The system MUST persist advanced scoring fields `memory_strength`, `last_boosted_at`, `boost_history`, `surprise_score`, and `connectivity` for each eligible memory record.

#### Scenario: New memory initializes advanced scoring fields
- **WHEN** a new memory is created through write pipeline
- **THEN** system MUST initialize advanced scoring fields with safe defaults and keep record writable by maintenance jobs

### Requirement: Backward-Compatible Ranking Fallback
The scoring pipeline MUST remain backward-compatible with v1.0 records that do not yet have advanced scoring values.

#### Scenario: Missing advanced fields does not break ranking
- **WHEN** a candidate memory has null or default advanced scoring fields during migration window
- **THEN** system MUST fall back to v1.0-compatible ranking behavior and return valid ordered results
