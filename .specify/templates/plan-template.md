# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: [e.g., Python 3.11, Swift 5.9, Rust 1.75 or NEEDS CLARIFICATION]  
**Primary Dependencies**: [e.g., FastAPI, UIKit, LLVM or NEEDS CLARIFICATION]  
**Storage**: [if applicable, e.g., PostgreSQL, CoreData, files or N/A]  
**Testing**: [e.g., pytest, XCTest, cargo test or NEEDS CLARIFICATION]  
**Target Platform**: [e.g., Linux server, iOS 15+, WASM or NEEDS CLARIFICATION]
**Project Type**: [single/web/mobile - determines source structure]  
**Performance Goals**: [domain-specific, e.g., 1000 req/s, 10k lines/sec, 60 fps or NEEDS CLARIFICATION]  
**Constraints**: [domain-specific, e.g., <200ms p95, <100MB memory, offline-capable or NEEDS CLARIFICATION]  
**Scale/Scope**: [domain-specific, e.g., 10k users, 1M LOC, 50 screens or NEEDS CLARIFICATION]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### MyY Chat Constitution Gates

**代码质量原则 Gates:**
- [ ] Testing strategy ensures ≥70% test coverage per microservice
- [ ] Security risks, concurrency, and idempotency addressed in design
- [ ] Static code analysis tools configured

**API规范原则 Gates:**
- [ ] Service communication uses gRPC + Protobuf (Kratos framework)
- [ ] External APIs follow RESTful standards
- [ ] API documentation plan included

**部署原则 Gates:**
- [ ] Docker Compose configuration for development/single-machine
- [ ] Kubernetes manifests for cloud-native deployment
- [ ] Configuration switching mechanism designed

**设计原则 Gates:**
- [ ] Object-oriented design principles followed (LSP, DIP, etc.)
- [ ] Design patterns justified, not over-engineered
- [ ] Single responsibility principle applied

**项目结构原则 Gates:**
- [ ] backend/golang and backend/python separation maintained
- [ ] frontend directory properly structured
- [ ] Each service has independent config, tests, deployment files

**架构原则 Gates:**
- [ ] High readability, scalability, maintainability demonstrated
- [ ] Horizontal expansion capability designed
- [ ] Service coupling minimized

**开发规范原则 Gates:**
- [ ] Chinese comments planned for code
- [ ] English logging configured
- [ ] Documentation plan aligned with requirements
- [ ] GitFlow branching strategy adopted (main, develop, feature/*, release/*, hotfix/*)
- [ ] Context7 MCP tool used for third-party library documentation lookup
- [ ] Code review checkpoints planned after each complete iteration/feature/bugfix
- [ ] Conventional Commits format adopted for commit messages

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── golang/              # Go services for microservices architecture
│   ├── src/
│   │   ├── models/
│   │   ├── services/
│   │   └── api/
│   └── tests/
├── python/              # Python services for microservices architecture
│   ├── src/
│   │   ├── models/
│   │   ├── services/
│   │   └── api/
│   └── tests/
└── tests/               # Integration tests across services

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
