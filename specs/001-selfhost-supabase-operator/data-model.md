# Data Model: Supabase Operator

## Core CRD: SupabaseProject

### SupabaseProjectSpec
```go
type SupabaseProjectSpec struct {
    // Project identity
    ProjectID string `json:"projectId" kubebuilder:validation:Required`

    // External dependencies configuration
    Database  DatabaseConfig  `json:"database" kubebuilder:validation:Required`
    Storage   StorageConfig   `json:"storage" kubebuilder:validation:Required`

    // Component configurations (optional overrides)
    Kong      *KongConfig      `json:"kong,omitempty"`
    Auth      *AuthConfig      `json:"auth,omitempty"`
    Realtime  *RealtimeConfig  `json:"realtime,omitempty"`
    PostgREST *PostgRESTConfig `json:"postgrest,omitempty"`
    StorageAPI *StorageAPIConfig `json:"storageApi,omitempty"`
    Meta      *MetaConfig      `json:"meta,omitempty"`

    // Network configuration
    Ingress *IngressConfig `json:"ingress,omitempty"`
}
```

### SupabaseProjectStatus
```go
type SupabaseProjectStatus struct {
    // Overall phase
    Phase string `json:"phase,omitempty"`

    // Human-readable message
    Message string `json:"message,omitempty"`

    // Kubernetes standard conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Component-specific status
    Components ComponentsStatus `json:"components,omitempty"`

    // External dependency status
    Dependencies DependenciesStatus `json:"dependencies,omitempty"`

    // Service endpoints
    Endpoints EndpointsStatus `json:"endpoints,omitempty"`

    // Reconciliation metadata
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`
    LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}
```

## Configuration Types

### DatabaseConfig
```go
type DatabaseConfig struct {
    // Connection details (stored in secret)
    // Secret must contain: host, port, database, username, password
    SecretRef corev1.SecretReference `json:"secretRef" kubebuilder:validation:Required`

    // SSL configuration
    SSLMode string `json:"sslMode,omitempty" kubebuilder:default:"require"`

    // Connection pool settings
    MaxConnections int `json:"maxConnections,omitempty" kubebuilder:default:20`
}
```

### StorageConfig
```go
type StorageConfig struct {
    // S3 configuration (stored in secret)
    // Secret must contain: endpoint, region, bucket, accessKeyId, secretAccessKey
    SecretRef corev1.SecretReference `json:"secretRef" kubebuilder:validation:Required`

    // Force path style (for MinIO)
    ForcePathStyle bool `json:"forcePathStyle,omitempty" kubebuilder:default:true`
}
```


## Component Configuration Types

### KongConfig
```go
type KongConfig struct {
    // Container image
    Image string `json:"image,omitempty" kubebuilder:default:"kong:2.8.1"`

    // Resource requirements
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // Number of replicas
    Replicas int32 `json:"replicas,omitempty" kubebuilder:default:1`

    // Additional environment variables
    ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}
```

### AuthConfig
```go
type AuthConfig struct {
    // Container image
    Image string `json:"image,omitempty" kubebuilder:default:"supabase/gotrue:v2.177.0"`

    // Resource requirements
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // Number of replicas
    Replicas int32 `json:"replicas,omitempty" kubebuilder:default:1`

    // SMTP configuration for email (via secret reference)
    SMTPSecretRef *corev1.SecretReference `json:"smtpSecretRef,omitempty"`

    // OAuth providers configuration (via secret reference)
    OAuthSecretRef *corev1.SecretReference `json:"oauthSecretRef,omitempty"`

    // Additional environment variables
    ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}
```

### RealtimeConfig
```go
type RealtimeConfig struct {
    // Container image
    Image string `json:"image,omitempty" kubebuilder:default:"supabase/realtime:v2.34.47"`

    // Resource requirements
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // Number of replicas
    Replicas int32 `json:"replicas,omitempty" kubebuilder:default:1`

    // Additional environment variables
    ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}
```

### PostgRESTConfig
```go
type PostgRESTConfig struct {
    // Container image
    Image string `json:"image,omitempty" kubebuilder:default:"postgrest/postgrest:v12.2.12"`

    // Resource requirements
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // Number of replicas
    Replicas int32 `json:"replicas,omitempty" kubebuilder:default:1`

    // Additional environment variables
    ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}
```

### StorageAPIConfig
```go
type StorageAPIConfig struct {
    // Container image
    Image string `json:"image,omitempty" kubebuilder:default:"supabase/storage-api:v1.25.7"`

    // Resource requirements
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // Number of replicas
    Replicas int32 `json:"replicas,omitempty" kubebuilder:default:1`

    // Additional environment variables
    ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}
```

### MetaConfig
```go
type MetaConfig struct {
    // Container image
    Image string `json:"image,omitempty" kubebuilder:default:"supabase/postgres-meta:v0.91.0"`

    // Resource requirements
    Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

    // Number of replicas
    Replicas int32 `json:"replicas,omitempty" kubebuilder:default:1`

    // Additional environment variables
    ExtraEnv []corev1.EnvVar `json:"extraEnv,omitempty"`
}
```

## Status Types

### ComponentsStatus
```go
type ComponentsStatus struct {
    Kong      ComponentStatus `json:"kong,omitempty"`
    Auth      ComponentStatus `json:"auth,omitempty"`
    Realtime  ComponentStatus `json:"realtime,omitempty"`
    PostgREST ComponentStatus `json:"postgrest,omitempty"`
    StorageAPI ComponentStatus `json:"storageApi,omitempty"`
    Meta      ComponentStatus `json:"meta,omitempty"`
}
```

### ComponentStatus
```go
type ComponentStatus struct {
    // Component phase
    Phase string `json:"phase,omitempty"`

    // Ready status
    Ready bool `json:"ready,omitempty"`

    // Deployed version
    Version string `json:"version,omitempty"`

    // Number of ready replicas
    ReadyReplicas int32 `json:"readyReplicas,omitempty"`

    // Total replicas
    Replicas int32 `json:"replicas,omitempty"`

    // Component-specific conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Last update time
    LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}
```

### DependenciesStatus
```go
type DependenciesStatus struct {
    PostgreSQL DependencyStatus `json:"postgresql,omitempty"`
    S3         DependencyStatus `json:"s3,omitempty"`
}

type DependencyStatus struct {
    // Connection status
    Connected bool `json:"connected"`

    // Last successful connection
    LastConnectedTime *metav1.Time `json:"lastConnectedTime,omitempty"`

    // Error message if not connected
    Error string `json:"error,omitempty"`

    // Latency in milliseconds
    LatencyMs int32 `json:"latencyMs,omitempty"`
}
```

### EndpointsStatus
```go
type EndpointsStatus struct {
    // Kong API gateway endpoint
    API string `json:"api,omitempty"`

    // Auth service endpoint
    Auth string `json:"auth,omitempty"`

    // Realtime websocket endpoint
    Realtime string `json:"realtime,omitempty"`

    // Storage API endpoint
    Storage string `json:"storage,omitempty"`

    // PostgreSQL REST endpoint
    REST string `json:"rest,omitempty"`
}
```

## Validation Rules

### Field Validations
- `projectId`: Required, must be DNS-1123 compliant
- `database.secretRef`: Required, must reference existing secret
- `storage.secretRef`: Required, must reference existing secret
- Component images: Must be valid container image references
- Resource requirements: Must be valid Kubernetes resource quantities
- Replicas: Minimum 0, maximum 10

### Secret Key Conventions (validated in controller/webhook)
**Database Secret Keys**:
- `host`: PostgreSQL host (required)
- `port`: PostgreSQL port (required)
- `database`: Database name (required)
- `username`: Database username (required)
- `password`: Database password (required)

**Storage Secret Keys**:
- `endpoint`: S3-compatible endpoint URL (required)
- `region`: Storage region (required)
- `bucket`: Bucket name (required)
- `accessKeyId`: Access key ID (required)
- `secretAccessKey`: Secret access key (required)

**SMTP Secret Keys** (if configured):
- `host`: SMTP host
- `port`: SMTP port
- `username`: SMTP username
- `password`: SMTP password
- `from`: From address

### Cross-field Validations (via webhook)
- Secrets must exist in the same namespace
- Required secret keys must be present
- Resource limits must be greater than or equal to requests
- Component replicas must be 1 if not explicitly configured for HA

## State Transitions

### Phase Transitions
```
Pending → ValidatingDependencies → DeployingSecrets → DeployingNetwork →
DeployingComponents → Configuring → Running

From any state → Updating (on spec change)
From any state → Failed (on error)
From Running/Failed → Terminating → Terminated
```

### Condition Transitions
- `Progressing`: True when any reconciliation is active
- `Ready`: True when all components are running
- `Available`: True when endpoints are accessible
- `Degraded`: True when some components are unhealthy

## Default Values

### Resource Defaults
```yaml
kong:
  resources:
    limits:
      memory: "2.5Gi"
      cpu: "500m"
    requests:
      memory: "1Gi"
      cpu: "250m"

auth:
  resources:
    limits:
      memory: "128Mi"
      cpu: "100m"
    requests:
      memory: "64Mi"
      cpu: "50m"

realtime:
  resources:
    limits:
      memory: "256Mi"
      cpu: "200m"
    requests:
      memory: "128Mi"
      cpu: "100m"

postgrest:
  resources:
    limits:
      memory: "256Mi"
      cpu: "200m"
    requests:
      memory: "128Mi"
      cpu: "100m"

storageApi:
  resources:
    limits:
      memory: "128Mi"
      cpu: "100m"
    requests:
      memory: "64Mi"
      cpu: "50m"

meta:
  resources:
    limits:
      memory: "128Mi"
      cpu: "100m"
    requests:
      memory: "64Mi"
      cpu: "50m"
```