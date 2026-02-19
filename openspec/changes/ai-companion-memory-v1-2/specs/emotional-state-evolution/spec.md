## ADDED Requirements

### Requirement: Secondary Emotion Capture
Emotional state records MUST support `secondary_emotion` in addition to primary PAD-aligned signals.

#### Scenario: Mixed emotion is recorded
- **WHEN** emotional analysis detects both primary and secondary emotional signals
- **THEN** system MUST persist both signals in the same emotional state record

### Requirement: Trigger-Cause Annotation
Emotional state records MUST support trigger-cause annotation via `trigger_type` and `trigger_content`.

#### Scenario: Emotion trigger is identified
- **WHEN** emotional update is caused by an identifiable conversation trigger
- **THEN** system MUST store trigger type and a trigger content representation usable for later explanation

### Requirement: Trigger Privacy Protection
Trigger-cause annotation MUST enforce privacy-safe storage rules.

#### Scenario: Sensitive trigger content appears
- **WHEN** trigger content includes sensitive raw text
- **THEN** system MUST store redacted or summarized trigger content according to configured privacy policy
