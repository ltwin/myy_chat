## ADDED Requirements

### Requirement: Rich Event Structure Storage
Important events MUST support structured attributes including `title`, `is_recurring`, `participants`, `related_memory_ids`, and `source_message_ids`.

#### Scenario: Event extraction persists rich attributes
- **WHEN** memory pipeline identifies an important event from one or more messages
- **THEN** system MUST persist event core summary and all available rich attributes in event storage

### Requirement: Multi-Source Event Traceability
Important events MUST preserve traceability to multiple source messages.

#### Scenario: Event links to multiple source messages
- **WHEN** an event is synthesized from several user-assistant turns
- **THEN** system MUST store all contributing `source_message_ids` so audit and explanation can reconstruct provenance

### Requirement: Recurring Event Upsert Semantics
Recurring events MUST support update semantics that preserve continuity instead of creating uncontrolled duplicates.

#### Scenario: Recurring event appears again
- **WHEN** a newly detected event matches an existing recurring event identity
- **THEN** system MUST update the existing recurring event record and append new source trace rather than inserting an unrelated duplicate
