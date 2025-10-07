# Tasks: Supabase Operator for Self-Hosted Deployments

**Input**: Design documents from `/specs/001-selfhost-supabase-operator/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
Kubernetes operator using Kubebuilder standard layout:
- **API types**: `api/v1alpha1/`
- **Controllers**: `internal/controller/`
- **Resources**: `internal/resources/`
- **Status**: `internal/status/`
- **Secrets**: `internal/secrets/`
- **Tests**: `internal/controller/` and root-level `test/e2e/`

## Phase 3.1: Project Setup & Dependencies
- [x] T001 Verify Kubebuilder initialization and Go module structure
- [x] T002 Add controller-runtime dependencies to go.mod (v0.22.1, k8s.io/client-go v0.34.0)
- [x] T003 [P] Configure golangci-lint with Kubernetes operator best practices
- [x] T004 [P] Set up GitHub Actions CI workflow for testing and linting
- [x] T005 [P] Create .editorconfig and .gitignore for Go/Kubernetes projects

## Phase 3.2: API Definition (TDD - Tests First) ⚠️ MUST COMPLETE BEFORE 3.3

### CRD Type Definition Tests
- [x] T006 [P] Unit test for SupabaseProjectSpec validation in api/v1alpha1/supabaseproject_types_test.go
- [x] T007 [P] Unit test for DatabaseConfig secret reference validation in api/v1alpha1/supabaseproject_types_test.go
- [x] T008 [P] Unit test for StorageConfig secret reference validation in api/v1alpha1/supabaseproject_types_test.go
- [x] T009 [P] Unit test for component config defaults (Kong, Auth, etc.) in api/v1alpha1/supabaseproject_types_test.go
- [x] T010 [P] Unit test for SupabaseProjectStatus structure in api/v1alpha1/supabaseproject_types_test.go

### Webhook Validation Tests
- [x] T011 [P] Validating webhook test for secret existence check in api/v1alpha1/supabaseproject_webhook_test.go
- [x] T012 [P] Validating webhook test for required secret keys in api/v1alpha1/supabaseproject_webhook_test.go
- [x] T013 [P] Validating webhook test for image reference validation in api/v1alpha1/supabaseproject_webhook_test.go
- [x] T014 [P] Mutating webhook test for resource defaults in api/v1alpha1/supabaseproject_webhook_test.go

## Phase 3.3: API Implementation (ONLY after tests are failing)

### CRD Types
- [x] T015 Define SupabaseProjectSpec struct in api/v1alpha1/supabaseproject_types.go
- [x] T016 Define DatabaseConfig and StorageConfig in api/v1alpha1/supabaseproject_types.go
- [x] T017 [P] Define KongConfig with image default in api/v1alpha1/supabaseproject_types.go
- [x] T018 [P] Define AuthConfig with image default in api/v1alpha1/supabaseproject_types.go
- [x] T019 [P] Define RealtimeConfig with image default in api/v1alpha1/supabaseproject_types.go
- [x] T020 [P] Define PostgRESTConfig with image default in api/v1alpha1/supabaseproject_types.go
- [x] T021 [P] Define StorageAPIConfig with image default in api/v1alpha1/supabaseproject_types.go
- [x] T022 [P] Define MetaConfig with image default in api/v1alpha1/supabaseproject_types.go
- [x] T023 Define SupabaseProjectStatus with conditions and components in api/v1alpha1/supabaseproject_types.go
- [x] T024 Define ComponentsStatus and ComponentStatus in api/v1alpha1/supabaseproject_types.go
- [x] T025 Define DependenciesStatus and EndpointsStatus in api/v1alpha1/supabaseproject_types.go
- [x] T026 Add kubebuilder markers for CRD generation in api/v1alpha1/supabaseproject_types.go
- [x] T027 Run make manifests to generate CRD YAML in config/crd/bases/

### Webhooks
- [x] T028 Implement ValidateCreate for secret validation in api/v1alpha1/supabaseproject_webhook.go
- [x] T029 Implement ValidateUpdate for spec changes in api/v1alpha1/supabaseproject_webhook.go
- [x] T030 Implement Default method for resource defaults in api/v1alpha1/supabaseproject_webhook.go
- [x] T031 Add webhook configuration in config/webhook/

## Phase 3.4: Secret Management (TDD)

### Tests
- [x] T032 [P] Unit test for JWT secret generation in internal/secrets/jwt_test.go
- [x] T033 [P] Unit test for ANON_KEY generation in internal/secrets/jwt_test.go
- [x] T034 [P] Unit test for SERVICE_ROLE_KEY generation in internal/secrets/jwt_test.go
- [x] T035 [P] Unit test for secret validation (database keys) in internal/secrets/validation_test.go
- [x] T036 [P] Unit test for secret validation (storage keys) in internal/secrets/validation_test.go

### Implementation
- [x] T037 Implement JWT secret generation (crypto/rand, 32 bytes) in internal/secrets/jwt.go
- [x] T038 Implement ANON_KEY JWT generation in internal/secrets/jwt.go
- [x] T039 Implement SERVICE_ROLE_KEY JWT generation in internal/secrets/jwt.go
- [x] T040 [P] Implement database secret validation in internal/secrets/validation.go
- [x] T041 [P] Implement storage secret validation in internal/secrets/validation.go

## Phase 3.5: Status Management (TDD)

### Tests
- [ ] T042 [P] Unit test for condition creation (Ready, Progressing, etc.) in internal/status/conditions_test.go
- [ ] T043 [P] Unit test for condition updates and transitions in internal/status/conditions_test.go
- [ ] T044 [P] Unit test for phase progression logic in internal/status/phase_test.go
- [ ] T045 [P] Unit test for component status tracking in internal/status/component_test.go

### Implementation
- [ ] T046 Implement standard condition helpers (Ready, Progressing, Available, Degraded) in internal/status/conditions.go
- [ ] T047 Implement component-specific condition helpers (KongReady, AuthReady, RealtimeReady, StorageReady, PostgRESTReady, MetaReady) in internal/status/conditions.go
- [ ] T048 Implement dependency condition helpers (PostgreSQLConnected, S3Connected) and infrastructure conditions (NetworkReady, SecretsReady) in internal/status/conditions.go
- [ ] T049 Implement phase progression state machine in internal/status/phase.go
- [ ] T050 Implement component status builder in internal/status/component.go

## Phase 3.6: Component Resource Builders (TDD)

### Tests for Kong
- [ ] T051 [P] Unit test for Kong Deployment creation in internal/resources/kong_test.go
- [ ] T052 [P] Unit test for Kong Service creation in internal/resources/kong_test.go
- [ ] T053 [P] Unit test for Kong resource defaults application in internal/resources/kong_test.go

### Tests for Auth
- [ ] T054 [P] Unit test for Auth Deployment creation in internal/resources/auth_test.go
- [ ] T055 [P] Unit test for Auth Service creation in internal/resources/auth_test.go
- [ ] T056 [P] Unit test for Auth environment variable injection in internal/resources/auth_test.go

### Tests for PostgREST
- [ ] T057 [P] Unit test for PostgREST Deployment creation in internal/resources/postgrest_test.go
- [ ] T058 [P] Unit test for PostgREST Service creation in internal/resources/postgrest_test.go

### Tests for Realtime
- [ ] T059 [P] Unit test for Realtime Deployment creation in internal/resources/realtime_test.go
- [ ] T060 [P] Unit test for Realtime Service creation in internal/resources/realtime_test.go

### Tests for Storage API
- [ ] T061 [P] Unit test for Storage API Deployment creation in internal/resources/storage_test.go
- [ ] T062 [P] Unit test for Storage API Service creation in internal/resources/storage_test.go
- [ ] T063 [P] Unit test for Storage API S3 config injection in internal/resources/storage_test.go

### Tests for Meta
- [ ] T064 [P] Unit test for Meta Deployment creation in internal/resources/meta_test.go
- [ ] T065 [P] Unit test for Meta Service creation in internal/resources/meta_test.go

### Implementation
- [ ] T066 Implement Kong Deployment builder with defaults (2.5GB, 500m CPU) in internal/resources/kong.go
- [ ] T067 Implement Kong Service builder in internal/resources/kong.go
- [ ] T068 Implement Auth Deployment builder with defaults (128MB, 100m CPU) in internal/resources/auth.go
- [ ] T069 Implement Auth Service builder in internal/resources/auth.go
- [ ] T070 Implement PostgREST Deployment builder with defaults (256MB, 200m CPU) in internal/resources/postgrest.go
- [ ] T071 Implement PostgREST Service builder in internal/resources/postgrest.go
- [ ] T072 Implement Realtime Deployment builder with defaults (256MB, 200m CPU) in internal/resources/realtime.go
- [ ] T073 Implement Realtime Service builder in internal/resources/realtime.go
- [ ] T074 Implement Storage API Deployment builder with defaults (128MB, 100m CPU) in internal/resources/storage.go
- [ ] T075 Implement Storage API Service builder in internal/resources/storage.go
- [ ] T076 Implement Meta Deployment builder with defaults (128MB, 100m CPU) in internal/resources/meta.go
- [ ] T077 Implement Meta Service builder in internal/resources/meta.go

## Phase 3.7: Controller Core (TDD)

### Controller Tests
- [ ] T078 Integration test for SupabaseProject creation with envtest in internal/controller/supabaseproject_controller_test.go
- [ ] T079 Integration test for dependency validation (PostgreSQL, S3) in internal/controller/supabaseproject_controller_test.go
- [ ] T079a Integration test for PostgreSQL database initialization (schemas, roles, extensions) in internal/controller/supabaseproject_controller_test.go
- [ ] T080 Integration test for secret generation in internal/controller/supabaseproject_controller_test.go
- [ ] T081 Integration test for component deployment order in internal/controller/supabaseproject_controller_test.go
- [ ] T082 Integration test for status updates during reconciliation in internal/controller/supabaseproject_controller_test.go
- [ ] T083 Integration test for finalizer cleanup in internal/controller/supabaseproject_controller_test.go
- [ ] T084 Integration test for update reconciliation (rolling updates) in internal/controller/supabaseproject_controller_test.go
- [ ] T085 Integration test for failure handling (no rollback) in internal/controller/supabaseproject_controller_test.go

### Controller Implementation
- [ ] T086 Implement Reconcile skeleton with finalizer check in internal/controller/supabaseproject_controller.go
- [ ] T087 Implement dependency validation phase in internal/controller/supabaseproject_controller.go
- [ ] T087a Implement PostgreSQL schema creation (auth, storage, realtime) in internal/controller/supabaseproject_controller.go
- [ ] T087b Implement PostgreSQL role creation (authenticator, anon, service_role) in internal/controller/supabaseproject_controller.go
- [ ] T087c Implement PostgreSQL extension setup (pgcrypto, pgjwt, uuid-ossp, pg_stat_statements) in internal/controller/supabaseproject_controller.go
- [ ] T088 Implement secret generation and reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T089 Implement component deployment orchestration (ordered) in internal/controller/supabaseproject_controller.go
- [ ] T090 Implement Kong deployment reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T091 Implement Auth deployment reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T092 Implement PostgREST deployment reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T093 Implement Realtime deployment reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T094 Implement Storage API deployment reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T095 Implement Meta deployment reconciliation in internal/controller/supabaseproject_controller.go
- [ ] T096 Implement status update logic for all conditions in internal/controller/supabaseproject_controller.go
- [ ] T097 Implement finalizer cleanup logic in internal/controller/supabaseproject_controller.go
- [ ] T098 Add controller to manager in cmd/main.go
- [ ] T099 Configure controller watches and predicates in internal/controller/supabaseproject_controller.go

## Phase 3.8: E2E Testing
- [ ] T100 [P] E2E test: Deploy SupabaseProject with minimal config in test/e2e/basic_test.go
- [ ] T101 [P] E2E test: Verify all components running in test/e2e/basic_test.go
- [ ] T102 [P] E2E test: Update component resources in test/e2e/update_test.go
- [ ] T103 [P] E2E test: Verify rolling update behavior in test/e2e/update_test.go
- [ ] T104 [P] E2E test: Delete SupabaseProject and verify cleanup in test/e2e/cleanup_test.go
- [ ] T105 [P] E2E test: Invalid dependency credentials in test/e2e/validation_test.go
- [ ] T106 [P] E2E test: Component failure recovery in test/e2e/failure_test.go
- [ ] T106a [P] E2E test: Deploy multiple SupabaseProjects across different namespaces in test/e2e/multinamespace_test.go

## Phase 3.9: Integration & Configuration
- [ ] T107 Create sample SupabaseProject CR in config/samples/supabase_v1alpha1_supabaseproject.yaml
- [ ] T108 [P] Configure RBAC permissions in config/rbac/
- [ ] T109 [P] Configure manager deployment in config/manager/manager.yaml
- [ ] T110 [P] Add Prometheus metrics endpoints (reconciliation_duration, queue_depth, error_rate, api_call_latency) in internal/controller/supabaseproject_controller.go
- [ ] T111 [P] Implement structured logging with correlation IDs and trace IDs in internal/controller/supabaseproject_controller.go
- [ ] T112 [P] Add event recording for state transitions in internal/controller/supabaseproject_controller.go

## Phase 3.10: Documentation & Polish
- [ ] T113 [P] Update README.md with installation instructions
- [ ] T114 [P] Create docs/architecture.md documenting design decisions
- [ ] T115 [P] Create docs/api-reference.md from CRD schema
- [ ] T116 [P] Add code comments and package documentation
- [ ] T117 [P] Run quickstart.md scenarios manually and verify
- [ ] T118 Verify all integration tests pass with `make test`
- [ ] T119 Build operator image with `make docker-build`
- [ ] T120 Deploy to test cluster and verify end-to-end

## Dependencies

### Setup Phase (T001-T005)
All setup tasks must complete before any other work

### API Definition (T006-T031)
- Tests (T006-T014) before implementation (T015-T031)
- T015-T026 (types) before T027 (CRD generation)
- T027 before webhooks (T028-T031)

### Secret Management (T032-T041)
- Tests (T032-T036) before implementation (T037-T041)
- Independent of other phases (can run in parallel with status)

### Status Management (T042-T050)
- Tests (T042-T045) before implementation (T046-T050)
- Independent of other phases

### Component Resources (T051-T077)
- All tests (T051-T065) before all implementations (T066-T077)
- Resource builders are parallel [P] after tests pass

### Controller Core (T078-T099)
- Integration tests (T078-T085) before controller implementation
- T086 (skeleton) before T087-T097
- Requires: API types (T015-T027), secrets (T037-T041), status (T046-T050), resources (T066-T077)
- T098-T099 must be last in controller phase

### E2E Testing (T100-T106)
- Requires complete controller implementation (T086-T099)
- All E2E tests can run in parallel [P]

### Integration & Config (T107-T112)
- Can run in parallel after controller core
- T107-T109 can be parallel [P]
- T110-T112 can be parallel [P]

### Documentation (T113-T120)
- Documentation tasks (T113-T116) can be parallel [P]
- T117-T120 must be sequential (manual testing → tests → build → deploy)

## Parallel Execution Examples

### After API tests pass (T006-T014 complete):
```bash
# Launch T017-T022 together (component configs)
Task: "Define KongConfig with image default in api/v1alpha1/supabaseproject_types.go"
Task: "Define AuthConfig with image default in api/v1alpha1/supabaseproject_types.go"
Task: "Define RealtimeConfig with image default in api/v1alpha1/supabaseproject_types.go"
Task: "Define PostgRESTConfig with image default in api/v1alpha1/supabaseproject_types.go"
Task: "Define StorageAPIConfig with image default in api/v1alpha1/supabaseproject_types.go"
Task: "Define MetaConfig with image default in api/v1alpha1/supabaseproject_types.go"
```

### Secret and Status management (independent):
```bash
# Launch T032-T036 and T042-T045 together
Task: "Unit test for JWT secret generation in internal/secrets/jwt_test.go"
Task: "Unit test for condition creation in internal/status/conditions_test.go"
# ... (all secret and status tests in parallel)
```

### Component resource tests:
```bash
# Launch T051-T065 together (all component tests)
Task: "Unit test for Kong Deployment creation in internal/resources/kong_test.go"
Task: "Unit test for Auth Deployment creation in internal/resources/auth_test.go"
Task: "Unit test for PostgREST Deployment creation in internal/resources/postgrest_test.go"
# ... (all resource tests in parallel)
```

### Component implementations after tests pass:
```bash
# Launch T066-T077 together (all resource builders)
Task: "Implement Kong Deployment builder in internal/resources/kong.go"
Task: "Implement Auth Deployment builder in internal/resources/auth.go"
Task: "Implement PostgREST Deployment builder in internal/resources/postgrest.go"
# ... (all builders in parallel)
```

### E2E tests after controller complete:
```bash
# Launch T100-T106 together
Task: "E2E test: Deploy SupabaseProject with minimal config"
Task: "E2E test: Update component resources"
Task: "E2E test: Delete and verify cleanup"
# ... (all E2E tests in parallel)
```

## Notes
- [P] tasks target different files and have no dependencies
- Verify tests fail before implementing (TDD)
- Commit after completing each task phase
- Run `make manifests` after modifying API types
- Run `make test` frequently to catch regressions
- Use envtest for integration tests (no real cluster needed)
- Keep failed state for investigation (no automatic rollback)

## Critical Path
T001-T005 → T006-T014 → T015-T027 → T028-T031 → T078-T085 → T086-T099 → T117-T120

Total estimated tasks: 120 tasks across 10 phases