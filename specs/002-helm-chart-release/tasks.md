---

description: "Task list for implementing automated Helm chart release"
---

# Tasks: Automated Helm Chart Release

**Input**: Design documents from `/specs/002-helm-chart-release/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: No test tasks included - specification does not request TDD approach for this CI/CD feature

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Workflow**: `.github/workflows/` at repository root
- **Charts**: `charts/supabase-operator/` at repository root
- **Documentation**: Root level `README.md` and `docs/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: GitHub repository configuration and prerequisite setup

**Note**: Assumes PAT already exists and will be manually added to repository settings

- [ ] T001 Verify PAT is added as secret named HELM_TOKEN in repository settings
- [X] T002 [P] Verify target repository STRRL/helm.strrl.dev exists
- [X] T003 [P] Initialize gh-pages branch in STRRL/helm.strrl.dev if not present
- [X] T004 Enable GitHub Pages for STRRL/helm.strrl.dev from gh-pages branch

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: No foundational blocking tasks for this feature - workflow is self-contained

**Note**: This phase is empty as the GitHub Actions workflow does not depend on any operator code changes

**Checkpoint**: Setup complete - user story implementation can now begin

---

## Phase 3: User Story 1 - Automated Chart Publishing on Version Tag (Priority: P1) 🎯 MVP

**Goal**: Automatically package and publish Helm charts when semantic version tags are pushed

**Independent Test**: Push a version tag v1.0.0 and verify chart appears at https://helm.strrl.dev with correct version

### Implementation for User Story 1

- [X] T005 [US1] Create GitHub Actions workflow file at .github/workflows/release-helm.yaml
- [X] T006 [US1] Configure workflow trigger for semantic version tags (v*.*.*)
- [X] T007 [US1] Add checkout step with full git history in workflow
- [X] T008 [US1] Add version extraction step to parse tag into version number
- [X] T009 [US1] Add Chart.yaml update step to apply extracted version
- [X] T010 [US1] Add helm lint validation step in workflow
- [X] T011 [US1] Configure stefanprodan/helm-gh-pages action with required parameters
- [X] T012 [US1] Add concurrency control to prevent race conditions
- [X] T013 [US1] Configure git user for automated commits in workflow
- [ ] T014 [US1] Test workflow by pushing test tag v0.0.1-test

**Checkpoint**: At this point, charts should automatically publish when version tags are pushed

---

## Phase 4: User Story 2 - Chart Installation and Upgrade (Priority: P2)

**Goal**: Ensure published charts are properly installable and upgradeable via standard Helm commands

**Independent Test**: Run `helm repo add` and `helm install` commands to verify chart installation works

### Implementation for User Story 2

- [X] T015 [P] [US2] Update charts/supabase-operator/Chart.yaml with proper metadata
- [X] T016 [P] [US2] Add maintainers section to Chart.yaml
- [X] T017 [P] [US2] Add home and sources URLs to Chart.yaml
- [X] T018 [P] [US2] Add keywords for chart discovery in Chart.yaml
- [X] T019 [US2] Verify charts/supabase-operator/values.yaml has sensible defaults
- [ ] T020 [US2] Test chart installation with helm install command locally
- [ ] T021 [US2] Verify chart upgrade path with helm upgrade command

**Checkpoint**: Charts can be installed and upgraded successfully via Helm commands

---

## Phase 5: User Story 3 - Chart Discovery and Documentation (Priority: P3)

**Goal**: Improve chart discoverability and provide comprehensive documentation for users

**Independent Test**: Access https://helm.strrl.dev/index.yaml and verify metadata is complete and accurate

### Implementation for User Story 3

- [X] T022 [P] [US3] Create charts/supabase-operator/README.md with installation instructions
- [X] T023 [P] [US3] Add description field to Chart.yaml (max 140 chars)
- [X] T024 [P] [US3] Add icon URL to Chart.yaml for visual identification
- [X] T025 [US3] Update root README.md with Helm installation section
- [X] T026 [US3] Add usage examples to charts/supabase-operator/README.md
- [X] T027 [US3] Add configuration options documentation to chart README
- [X] T028 [US3] Create RELEASE.md with instructions for maintainers on creating releases
- [X] T029 [P] [US3] Add annotations for ArtifactHub compatibility in Chart.yaml

**Checkpoint**: Charts are well-documented and easily discoverable

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that enhance the overall release process

- [X] T030 [P] Add troubleshooting section to documentation
- [X] T031 Create GitHub release template for version tags
- [X] T032 [P] Document PAT rotation process for security
- [ ] T033 Test concurrent release scenario with multiple tags
- [X] T034 [P] Add workflow status badge to README.md
- [ ] T035 Validate quickstart.md instructions end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - must complete first for PAT and repository configuration
- **Foundational (Phase 2)**: Empty phase - no blocking tasks
- **User Stories (Phase 3+)**:
  - User Story 1 depends on Setup completion
  - User Stories 2 and 3 can proceed after Setup but benefit from Story 1 completion
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Setup - Core workflow implementation
- **User Story 2 (P2)**: Can start after Setup - Chart metadata improvements (benefits from US1 for testing)
- **User Story 3 (P3)**: Can start after Setup - Documentation (benefits from US1 & US2 for examples)

### Within Each User Story

- User Story 1: Sequential workflow construction (each step builds on previous)
- User Story 2: Chart.yaml updates can be parallel, testing is sequential
- User Story 3: Documentation tasks mostly parallel

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel after T001
- User Story 2 Chart.yaml updates (T015-T018) can run in parallel
- User Story 3 documentation tasks (T022-T024, T029) can run in parallel
- Polish phase tasks marked [P] can run in parallel

---

## Parallel Example: User Story 2

```bash
# Launch all Chart.yaml updates together:
Task: "Update charts/supabase-operator/Chart.yaml with proper metadata"
Task: "Add maintainers section to Chart.yaml"
Task: "Add home and sources URLs to Chart.yaml"
Task: "Add keywords for chart discovery in Chart.yaml"
```

---

## Parallel Example: User Story 3

```bash
# Launch all documentation tasks together:
Task: "Create charts/supabase-operator/README.md with installation instructions"
Task: "Add description field to Chart.yaml (max 140 chars)"
Task: "Add icon URL to Chart.yaml for visual identification"
Task: "Add annotations for ArtifactHub compatibility in Chart.yaml"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T004)
2. Complete Phase 3: User Story 1 (T005-T014)
3. **STOP and VALIDATE**: Test by pushing a version tag
4. Verify chart publishes to https://helm.strrl.dev
5. MVP complete - automated releases working!

**Total MVP Tasks**: 14 tasks (4 setup + 10 implementation)

### Incremental Delivery

1. Complete Setup → Prerequisites ready
2. Add User Story 1 → Test tag push → Automated publishing works (MVP!)
3. Add User Story 2 → Test installation → Charts fully installable
4. Add User Story 3 → Documentation complete → Full user experience
5. Each story adds value without breaking previous functionality

### Parallel Team Strategy

With multiple developers:

1. Developer A: Complete Setup (T001-T004)
2. Once Setup done:
   - Developer A: User Story 1 (workflow implementation)
   - Developer B: User Story 2 (chart metadata)
   - Developer C: User Story 3 (documentation)
3. Stories complete independently and integrate seamlessly

---

## Task Summary

- **Total Tasks**: 35
- **Setup Tasks**: 4
- **User Story 1 Tasks**: 10 (Core workflow - MVP)
- **User Story 2 Tasks**: 7 (Chart quality)
- **User Story 3 Tasks**: 8 (Documentation)
- **Polish Tasks**: 6
- **Parallel Opportunities**: 15 tasks marked [P]

## Notes

- [P] tasks = different files or independent GitHub configurations
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable
- Commit after each task completion
- Test workflow with incremental tags (v0.0.1-test, v0.0.2-test)
- No operator code changes required - purely CI/CD configuration