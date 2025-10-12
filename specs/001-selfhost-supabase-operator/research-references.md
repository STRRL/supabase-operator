# Research References for Supabase Operator

## Rook - Cloud-Native Storage Orchestrator
**Repository**: https://github.com/rook/rook
**Relevance**: Excellent example of Kubernetes operator patterns, especially for status management and CRD design

### Key Patterns to Study from Rook

#### 1. Status Management Design
Rook implements sophisticated status reporting with multiple layers:
- **Cluster-level conditions** (similar to our SupabaseProject)
- **Component-level status** tracking individual services
- **Phase progression** with detailed state machines
- **Health monitoring** with granular condition types

#### 2. CRD Structure Patterns
```yaml
# Example from Rook CephCluster
status:
  phase: Ready  # Overall phase
  message: "Cluster is healthy"
  conditions:
    - type: Progressing
      status: "False"
      reason: ClusterCompleted
      message: "Cluster is operating normally"
    - type: Ready
      status: "True"
      reason: ClusterReady
    - type: Upgrading
      status: "False"
  ceph:
    health: HEALTH_OK
    details: {...}
  storage:
    deviceClasses: [...]
```

#### 3. Reconciliation Patterns
- **Ordered component deployment** - Dependencies respected
- **Health check integration** before marking components ready
- **Rolling update strategies** with version tracking
- **Finalizer patterns** for cleanup

#### 4. Observability Features
- **Detailed events** for user feedback
- **Prometheus metrics** exposure
- **Structured logging** with correlation IDs
- **Status conditions** following K8s conventions

### Specific Implementation Ideas for Supabase Operator

#### Status Structure (inspired by Rook)
```go
type SupabaseProjectStatus struct {
    // Overall phase (like Rook's cluster phase)
    Phase string `json:"phase,omitempty"`
    Message string `json:"message,omitempty"`

    // Granular conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Component-specific status (like Rook's ceph status)
    Components ComponentsStatus `json:"components,omitempty"`

    // Connection status (like Rook's external cluster info)
    Dependencies DependenciesStatus `json:"dependencies,omitempty"`

    // Operational info
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`
    LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

type ComponentsStatus struct {
    Kong      ComponentStatus `json:"kong,omitempty"`
    Auth      ComponentStatus `json:"auth,omitempty"`
    Realtime  ComponentStatus `json:"realtime,omitempty"`
    Storage   ComponentStatus `json:"storage,omitempty"`
    PostgREST ComponentStatus `json:"postgrest,omitempty"`
    Meta      ComponentStatus `json:"meta,omitempty"`
}

type ComponentStatus struct {
    Phase      string            `json:"phase,omitempty"`
    Ready      bool              `json:"ready,omitempty"`
    Version    string            `json:"version,omitempty"`
    Endpoint   string            `json:"endpoint,omitempty"`
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

#### Condition Types (following Rook patterns)
- **Standard Conditions**: Ready, Progressing, Available, Degraded
- **Component Conditions**: KongReady, AuthReady, RealtimeReady, etc.
- **Dependency Conditions**: PostgreSQLConnected, S3Connected
- **Operational Conditions**: NetworkReady, SecretsReady, CertificatesReady

#### Phase Progression (like Rook)
```
Pending → ValidatingDependencies → CreatingSecrets → DeployingInfrastructure →
DeployingComponents → Configuring → Running → Updating → Terminating
```

### Reconciliation Loop Structure (from Rook)
```go
func (r *SupabaseProjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Fetch the SupabaseProject instance
    // 2. Check deletion timestamp (handle finalizers)
    // 3. Validate external dependencies
    // 4. Reconcile secrets
    // 5. Reconcile network resources
    // 6. Reconcile each component in order
    // 7. Update status with detailed conditions
    // 8. Emit events for significant state changes
}
```

### Testing Patterns from Rook
- **Envtest** for integration testing
- **Fake client** for unit testing controllers
- **Status assertion helpers** for validating conditions
- **Timeout handling** in tests

### References for Implementation
1. [Rook CephCluster Controller](https://github.com/rook/rook/blob/master/pkg/operator/ceph/cluster/controller.go)
2. [Rook Status Types](https://github.com/rook/rook/blob/master/pkg/apis/ceph.rook.io/v1/types.go)
3. [Rook Conditions](https://github.com/rook/rook/blob/master/pkg/operator/ceph/cluster/status.go)
4. [Rook Reconcile Pattern](https://github.com/rook/rook/blob/master/pkg/operator/ceph/cluster/cluster.go)

### Action Items
- [ ] Study Rook's condition update patterns
- [ ] Implement similar phase progression logic
- [ ] Adopt Rook's component status tracking approach
- [ ] Use Rook's event emission patterns
- [ ] Follow Rook's testing strategies with envtest