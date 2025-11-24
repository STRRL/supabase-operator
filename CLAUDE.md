# Claude Code Context: Supabase Operator

## Project Overview
Building a Kubernetes operator for deploying and managing self-hosted Supabase instances using Kubebuilder framework.

## Current Technologies
- **Language**: Go 1.22+
- **Framework**: Kubebuilder v4.0+
- **Testing**: Ginkgo/Gomega with envtest
- **Container Runtime**: Kubernetes 1.25+
- **Key Dependencies**: controller-runtime v0.22.1, k8s.io/client-go v0.34.0

## Architecture Patterns
- **Controller Pattern**: Idempotent reconciliation loops
- **Status Management**: Granular conditions inspired by Rook operator
- **Component Deployment**: Ordered deployment with health checks
- **External Dependencies**: PostgreSQL and S3 (user-provided only)

## Key Components
1. **CRD**: SupabaseProject (namespace-scoped)
2. **Managed Services**: Kong, Auth/GoTrue, PostgREST, Realtime, Storage API, Meta
3. **Status Tracking**: Per-component phases and conditions
4. **Secret Management**: JWT generation and API key management

## Development Guidelines
- Follow Kubebuilder project structure
- Use TDD with envtest for controller testing
- Implement comprehensive status reporting
- Use structured logging with controller-runtime
- Follow Kubernetes API conventions

## Recent Changes
- 002-helm-chart-release: Added Go 1.22+ (operator code), YAML (Helm charts/workflows) + GitHub Actions, helm-gh-pages action, Helm 3.x
- Initialized Kubebuilder project with domain strrl.dev
- Created SupabaseProject API scaffold

## Testing Requirements
- Unit tests for reconciliation logic
- Integration tests with envtest
- Contract tests for CRD validation
- E2E tests for deployment scenarios

## Current Focus
Implementing the SupabaseProject controller with granular status management and component deployment logic.

## Active Technologies
- Go 1.22+ (operator code), YAML (Helm charts/workflows) + GitHub Actions, helm-gh-pages action, Helm 3.x (002-helm-chart-release)
- GitHub Pages repository for chart hosting (002-helm-chart-release)
