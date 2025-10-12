
# Implementation Plan: Supabase Operator for Self-Hosted Deployments

**Branch**: `001-selfhost-supabase-operator` | **Date**: 2025-10-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-selfhost-supabase-operator/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Build a Kubernetes operator that deploys and manages complete Supabase instances via CRD. The operator handles all Supabase native components (Kong, Auth, PostgREST, Realtime, Storage API, Meta) while integrating with external PostgreSQL and S3 dependencies. Uses Kubebuilder framework with granular status reporting inspired by Rook operator patterns.

## Technical Context
**Language/Version**: Go 1.22+ (Kubebuilder requirement)
**Primary Dependencies**: controller-runtime v0.22.1, k8s.io/client-go v0.34.0, Kubebuilder v4.0+
**Storage**: External PostgreSQL (user-provided), External S3-compatible storage (user-provided)
**Testing**: Go test with envtest (controller-runtime test environment), Ginkgo/Gomega
**Target Platform**: Kubernetes 1.25+ clusters
**Project Type**: single (Kubernetes operator)
**Performance Goals**: Handle 100+ SupabaseProject instances, reconciliation <5s per resource
**Constraints**: Memory limits per spec (Kong 2.5GB, others 128-256MB), rolling updates required
**Scale/Scope**: Multi-tenant operator managing multiple Supabase instances across namespaces

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **I. Controller Reconciliation Pattern**: ✅ Using Kubebuilder with reconciliation loops
- [x] **II. Custom Resource Definitions First**: ✅ SupabaseProject CRD with comprehensive validation
- [x] **III. Test-First Development**: ✅ TDD with envtest planned
- [x] **IV. Structured Status Reporting**: ✅ Granular conditions and component status
- [x] **V. Dependency Integration via Composition**: ✅ External PostgreSQL and S3 only
- [x] **VI. Observability and Operations**: ✅ Logs, events, metrics planned
- [x] **VII. Security and RBAC**: ✅ JWT generation, secrets management planned

## Project Structure

### Documentation (this feature)
```
specs/001-selfhost-supabase-operator/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── research-references.md # Rook patterns reference
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Kubebuilder standard layout
api/
├── v1alpha1/
│   ├── supabaseproject_types.go      # CRD type definitions
│   ├── supabaseproject_webhook.go    # Admission webhooks
│   ├── groupversion_info.go          # API group registration
│   └── zz_generated.deepcopy.go      # Generated code

internal/
├── controller/
│   ├── supabaseproject_controller.go # Main reconciliation logic
│   ├── supabaseproject_controller_test.go # Controller tests
│   └── suite_test.go                 # Test suite setup
├── resources/
│   ├── kong.go                       # Kong deployment logic
│   ├── auth.go                       # Auth/GoTrue deployment
│   ├── postgrest.go                  # PostgREST deployment
│   ├── realtime.go                   # Realtime deployment
│   ├── storage.go                    # Storage API deployment
│   └── meta.go                       # Meta deployment
├── status/
│   ├── conditions.go                 # Condition management
│   └── phase.go                      # Phase tracking
└── secrets/
    └── jwt.go                         # JWT generation logic

config/
├── crd/                               # CRD manifests
├── manager/                           # Operator deployment
├── rbac/                              # RBAC permissions
├── samples/                           # Sample CRs
└── webhook/                           # Webhook configs
```

**Structure Decision**: Using Kubebuilder standard layout for Kubernetes operators. This follows the established pattern with clear separation between API definitions (`api/v1alpha1/`), controller logic (`internal/controller/`), and resource management (`internal/resources/`). The structure aligns with constitution requirements for code organization and testability.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- CRD definition and validation tasks
- Controller scaffolding and reconciliation logic tasks
- Component deployment tasks (one per Supabase service)
- Status management and condition tracking tasks
- Integration and e2e test tasks

**Ordering Strategy**:
- TDD order: Tests before implementation
- API types → Controller logic → Component resources → Status updates
- Mark [P] for parallel execution (independent files)

**Task Categories**:
1. **Setup**: Project structure, dependencies
2. **API Definition**: CRD types, validation, webhooks
3. **Controller Core**: Reconciliation loop, finalizers
4. **Component Resources**: Deployment logic for each service (Kong, Auth, etc.)
5. **Status Management**: Conditions, phases, component tracking
6. **Testing**: Unit, integration with envtest, e2e

**Estimated Output**: 40-50 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 1.5: Design review complete (user review - 2025-10-03)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Design conflicts resolved (see design-decisions.md)
- [x] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
