## ADDED Requirements

### Requirement: Edge Explanation and Heat Metadata
Graph edges MUST persist explanation and heat metadata using `relation_description`, `mention_count`, and `last_mentioned_at`.

#### Scenario: New evidence updates edge heat
- **WHEN** pipeline confirms an existing relation through a new conversation mention
- **THEN** system MUST increment `mention_count`, update `last_mentioned_at`, and keep explanation metadata queryable

### Requirement: Temporal-Consistent Edge Evolution
Graph edge updates MUST preserve temporal semantics while extending advanced metadata.

#### Scenario: Relation changes across time
- **WHEN** relation type changes for the same entity pair
- **THEN** system MUST close the previous temporal edge validity and create a new edge version with independent heat metadata

### Requirement: Heat Decay Maintenance
The system MUST support scheduled decay of stale edge heat so outdated relations do not dominate retrieval ranking.

#### Scenario: Scheduled decay lowers stale relation heat
- **WHEN** graph maintenance job runs for edges with old `last_mentioned_at`
- **THEN** system MUST reduce effective heat contribution while retaining historical temporal records
