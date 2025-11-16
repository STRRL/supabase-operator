# API Reference

## Overview

This document provides detailed API reference for the Supabase Operator Custom Resource Definitions (CRDs).

**API Group:** `supabase.strrl.dev`
**API Version:** `v1alpha1`
**Kind:** `SupabaseProject`

## SupabaseProject

`SupabaseProject` is the primary resource for deploying and managing a complete Supabase instance on Kubernetes.

### Resource Metadata

```yaml
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: my-supabase
  namespace: default
```

### Spec Fields

#### SupabaseProjectSpec

Top-level specification for a Supabase deployment.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `projectId` | string | Yes | - | Unique project identifier. Must be DNS-1123 compliant: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` |
| `database` | [DatabaseConfig](#databaseconfig) | Yes | - | PostgreSQL database configuration |
| `storage` | [StorageConfig](#storageconfig) | Yes | - | S3-compatible storage configuration |
| `kong` | [KongConfig](#kongconfig) | No | See defaults | Kong API Gateway configuration |
| `auth` | [AuthConfig](#authconfig) | No | See defaults | Auth/GoTrue service configuration |
| `realtime` | [RealtimeConfig](#realtimeconfig) | No | See defaults | Realtime service configuration |
| `postgrest` | [PostgRESTConfig](#postgrestconfig) | No | See defaults | PostgREST service configuration |
| `storageApi` | [StorageAPIConfig](#storageapiconfig) | No | See defaults | Storage API service configuration |
| `meta` | [MetaConfig](#metaconfg) | No | See defaults | Meta service configuration |
| `studio` | [StudioConfig](#studioconfig) | No | See defaults | Studio UI configuration |
| `ingress` | [IngressConfig](#ingressconfig) | No | - | Ingress configuration for external access |

#### DatabaseConfig

Configuration for external PostgreSQL database.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `secretRef` | [SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#secretreference-v1-core) | Yes | - | Reference to Secret containing database credentials. Must contain keys: `host`, `port`, `database`, `username`, `password` |
| `sslMode` | string | No | `"require"` | PostgreSQL SSL mode. Valid values: `disable`, `require`, `verify-ca`, `verify-full` |
| `maxConnections` | int | No | `20` | Maximum number of database connections per component. Range: 1-100 |

**Database Secret Requirements:**

The referenced secret must contain the following keys:

- `host`: PostgreSQL hostname (e.g., `postgres.default.svc.cluster.local`)
- `port`: PostgreSQL port (e.g., `5432`)
- `database`: Database name (must be `postgres` for supabase/postgres image)
- `username`: PostgreSQL user (must have SUPERUSER or be `supabase_admin`)
- `password`: PostgreSQL password

**Required Database Privileges:**

The database user must have privileges to:
- CREATE DATABASE (for `_supabase` database)
- CREATE ROLE (for service roles: `authenticator`, `anon`, `service_role`)
- CREATE EXTENSION (for `pg_net`, `pgcrypto`, `uuid-ossp`, etc.)
- CREATE EVENT TRIGGER (requires superuser)

**Recommended:** Use the `postgres` or `supabase_admin` user from the `supabase/postgres` image.

**Example Secret:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-config
type: Opaque
stringData:
  host: postgres.default.svc.cluster.local
  port: "5432"
  database: postgres
  username: postgres
  password: your-secure-password
```

#### StorageConfig

Configuration for S3-compatible object storage.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `secretRef` | [SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#secretreference-v1-core) | Yes | - | Reference to Secret containing S3 credentials. Must contain keys: `endpoint`, `region`, `bucket`, `accessKeyId`, `secretAccessKey` |
| `forcePathStyle` | bool | No | `true` | Use path-style URLs for S3 requests (required for MinIO) |

**Storage Secret Requirements:**

The referenced secret must contain the following keys:

- `endpoint`: S3-compatible endpoint URL (e.g., `https://minio.default.svc.cluster.local:9000`)
- `region`: Storage region (e.g., `us-east-1`)
- `bucket`: Bucket name for storing files (e.g., `supabase-storage`)
- `accessKeyId`: S3 access key ID (camelCase, not kebab-case)
- `secretAccessKey`: S3 secret access key (camelCase, not kebab-case)

**Important:** Use camelCase for `accessKeyId` and `secretAccessKey` keys.

**Example Secret:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: s3-config
type: Opaque
stringData:
  endpoint: https://minio.default.svc.cluster.local:9000
  region: us-east-1
  bucket: supabase-storage
  accessKeyId: minioadmin
  secretAccessKey: minioadmin
```

#### KongConfig

Configuration for Kong API Gateway.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `kong:2.8.1` | Container image for Kong |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 2.5Gi
    cpu: 500m
  requests:
    memory: 1Gi
    cpu: 250m
```

#### AuthConfig

Configuration for Auth/GoTrue authentication service.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `supabase/gotrue:v2.177.0` | Container image for Auth |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `smtpSecretRef` | [SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#secretreference-v1-core) | No | - | Reference to Secret containing SMTP configuration for email |
| `oauthSecretRef` | [SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#secretreference-v1-core) | No | - | Reference to Secret containing OAuth provider configuration |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 128Mi
    cpu: 100m
  requests:
    memory: 64Mi
    cpu: 50m
```

**SMTP Secret Keys (optional):**
- `host`: SMTP server hostname
- `port`: SMTP server port
- `username`: SMTP username
- `password`: SMTP password
- `from`: From email address

#### RealtimeConfig

Configuration for Realtime WebSocket service.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `supabase/realtime:v2.34.47` | Container image for Realtime |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 256Mi
    cpu: 200m
  requests:
    memory: 128Mi
    cpu: 100m
```

#### PostgRESTConfig

Configuration for PostgREST REST API service.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `postgrest/postgrest:v12.2.12` | Container image for PostgREST |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 256Mi
    cpu: 200m
  requests:
    memory: 128Mi
    cpu: 100m
```

#### StorageAPIConfig

Configuration for Storage API file storage service.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `supabase/storage-api:v1.25.7` | Container image for Storage API |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 128Mi
    cpu: 100m
  requests:
    memory: 64Mi
    cpu: 50m
```

#### MetaConfig

Configuration for Meta PostgreSQL metadata service.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `supabase/postgres-meta:v0.91.0` | Container image for Meta |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 128Mi
    cpu: 100m
  requests:
    memory: 64Mi
    cpu: 50m
```

#### StudioConfig

Configuration for Supabase Studio management UI.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `image` | string | No | `supabase/studio:2025.10.01-sha-8460121` | Container image for Studio |
| `resources` | [ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core) | No | See below | CPU and memory resource requirements |
| `replicas` | int32 | No | `1` | Number of replicas. Range: 0-10 |
| `extraEnv` | [][EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) | No | `[]` | Additional environment variables |
| `publicUrl` | string | No | - | Public URL where Studio will be accessible |
| `dashboardBasicAuthSecretRef` | [SecretReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#secretreference-v1-core) | No | - | Reference to Secret containing basic auth credentials for Studio dashboard. Must contain keys: `username`, `password` |

**Default Resources:**

```yaml
resources:
  limits:
    memory: 256Mi
    cpu: 100m
  requests:
    memory: 128Mi
    cpu: 50m
```

**Dashboard Basic Auth Secret:**

When provided, Kong will protect the Studio route with HTTP basic authentication:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: studio-dashboard-creds
type: Opaque
stringData:
  username: admin
  password: secure-password
```

#### IngressConfig

Configuration for Kubernetes Ingress resource.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | `false` | Enable Ingress creation |
| `className` | *string | No | - | Ingress class name (e.g., `nginx`, `traefik`) |
| `annotations` | map[string]string | No | `{}` | Ingress annotations |
| `host` | string | No | - | Hostname for Ingress rules |
| `tlsSecretName` | string | No | - | Name of Secret containing TLS certificate |

**Example:**

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  host: supabase.example.com
  tlsSecretName: supabase-tls
```

### Status Fields

#### SupabaseProjectStatus

Observed state of a Supabase deployment (read-only).

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Current lifecycle phase: `Pending`, `ValidatingDependencies`, `InitializingDatabase`, `DeployingSecrets`, `DeployingComponents`, `Running`, `Failed`, `Updating` |
| `message` | string | Human-readable message describing current state |
| `conditions` | [][Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) | Detailed condition information (see below) |
| `components` | [ComponentsStatus](#componentsstatus) | Per-component status information |
| `dependencies` | [DependenciesStatus](#dependenciesstatus) | External dependency connectivity status |
| `endpoints` | [EndpointsStatus](#endpointsstatus) | Service endpoints for accessing components |
| `observedGeneration` | int64 | Generation of spec that was last processed |
| `lastReconcileTime` | *[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#time-v1-meta) | Timestamp of last reconciliation |

#### Condition Types

The operator maintains the following condition types:

**Standard Conditions:**
- `Ready`: Overall readiness (`True` when all components are healthy)
- `Progressing`: Reconciliation in progress
- `Available`: Endpoints are accessible
- `Degraded`: Some components are unhealthy

**Component-Specific Conditions:**
- `KongReady`, `AuthReady`, `RealtimeReady`, `StorageReady`, `PostgRESTReady`, `MetaReady`, `StudioReady`

**Dependency Conditions:**
- `PostgreSQLConnected`: Database connectivity verified
- `S3Connected`: Storage connectivity verified

**Infrastructure Conditions:**
- `NetworkReady`: Services and networking configured
- `SecretsReady`: JWT secrets generated and available

#### ComponentsStatus

Status information for all Supabase components.

| Field | Type | Description |
|-------|------|-------------|
| `kong` | [ComponentStatus](#componentstatus) | Kong API Gateway status |
| `auth` | [ComponentStatus](#componentstatus) | Auth/GoTrue status |
| `realtime` | [ComponentStatus](#componentstatus) | Realtime status |
| `postgrest` | [ComponentStatus](#componentstatus) | PostgREST status |
| `storageApi` | [ComponentStatus](#componentstatus) | Storage API status |
| `meta` | [ComponentStatus](#componentstatus) | Meta status |
| `studio` | [ComponentStatus](#componentstatus) | Studio status |

#### ComponentStatus

Status for an individual component.

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | Component phase: `Pending`, `Deploying`, `Running`, `Failed` |
| `ready` | bool | Whether component is ready to serve traffic |
| `version` | string | Deployed container image version |
| `readyReplicas` | int32 | Number of ready replicas |
| `replicas` | int32 | Total number of replicas |
| `conditions` | [][Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) | Component-specific conditions |
| `lastUpdateTime` | *[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#time-v1-meta) | Last status update time |

#### DependenciesStatus

Status of external dependencies.

| Field | Type | Description |
|-------|------|-------------|
| `postgresql` | [DependencyStatus](#dependencystatus) | PostgreSQL database status |
| `s3` | [DependencyStatus](#dependencystatus) | S3 storage status |

#### DependencyStatus

Connectivity status for an external dependency.

| Field | Type | Description |
|-------|------|-------------|
| `connected` | bool | Whether connection is established |
| `lastConnectedTime` | *[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#time-v1-meta) | Last successful connection time |
| `error` | string | Error message if connection failed |
| `latencyMs` | int32 | Connection latency in milliseconds |

#### EndpointsStatus

Service endpoints for accessing deployed components.

| Field | Type | Description |
|-------|------|-------------|
| `api` | string | Kong API Gateway endpoint (main entry point) |
| `auth` | string | Auth service endpoint |
| `realtime` | string | Realtime WebSocket endpoint |
| `storage` | string | Storage API endpoint |
| `rest` | string | PostgREST endpoint |

## Complete Example

```yaml
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: production-supabase
  namespace: production
spec:
  projectId: prod-project

  database:
    secretRef:
      name: postgres-config
    sslMode: require
    maxConnections: 50

  storage:
    secretRef:
      name: s3-config
    forcePathStyle: true

  kong:
    replicas: 2
    resources:
      limits:
        memory: 4Gi
        cpu: 1000m
      requests:
        memory: 2Gi
        cpu: 500m
    extraEnv:
      - name: KONG_LOG_LEVEL
        value: info

  auth:
    replicas: 2
    smtpSecretRef:
      name: smtp-config
    oauthSecretRef:
      name: oauth-config

  realtime:
    replicas: 2

  postgrest:
    replicas: 3
    resources:
      limits:
        memory: 512Mi
        cpu: 400m

  storageApi:
    replicas: 2

  meta:
    replicas: 1

  studio:
    publicUrl: https://studio.example.com
    dashboardBasicAuthSecretRef:
      name: studio-creds

  ingress:
    enabled: true
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/proxy-body-size: 100m
    host: api.example.com
    tlsSecretName: api-tls
```

## Generated Secrets

The operator automatically generates a secret named `<project-name>-jwt` containing:

| Key | Description |
|-----|-------------|
| `jwt-secret` | Base64-encoded JWT signing secret (256-bit) |
| `anon-key` | JWT token with 'anon' role claim (public API key) |
| `service-role-key` | JWT token with 'service_role' role claim (admin API key) |
| `pg-meta-crypto-key` | Encryption key for Meta service |

**Retrieve Keys:**

```bash
kubectl get secret my-supabase-jwt -o jsonpath='{.data.anon-key}' | base64 -d
kubectl get secret my-supabase-jwt -o jsonpath='{.data.service-role-key}' | base64 -d
```

## Validation Rules

### Admission Webhook Validations

The operator enforces the following validations via admission webhook:

1. **Secret Existence:**
   - Referenced secrets must exist in the same namespace
   - Required secret keys must be present

2. **Field Constraints:**
   - `projectId` must match pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
   - `database.maxConnections` must be between 1 and 100
   - Component `replicas` must be between 0 and 10

3. **Resource Requirements:**
   - Resource limits must be greater than or equal to requests

4. **Image References:**
   - Container images must be valid references

### Database User Privileges

The database user specified in the database secret must have the following PostgreSQL privileges:

- `CREATEDB` or `SUPERUSER`
- Ability to create roles
- Ability to create extensions
- Ability to create event triggers (superuser required)

Recommended users from `supabase/postgres` image:
- `postgres` (superuser)
- `supabase_admin` (has required privileges)

## Monitoring

### Status Inspection

```bash
kubectl get supabaseproject my-supabase -o jsonpath='{.status.phase}'

kubectl get supabaseproject my-supabase -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}'

kubectl get supabaseproject my-supabase -o jsonpath='{.status.components.kong.ready}'
```

### Events

```bash
kubectl describe supabaseproject my-supabase
```

### Logs

```bash
kubectl logs -n supabase-operator-system -l control-plane=controller-manager --tail=100
```

## Migration Guide

### Updating Component Images

To update a component to a new image version:

```yaml
spec:
  kong:
    image: kong:3.0.0
```

The operator will perform a rolling update automatically.

### Scaling Components

To scale a component:

```yaml
spec:
  postgrest:
    replicas: 5
```

### Resource Tuning

To adjust resource limits:

```yaml
spec:
  kong:
    resources:
      limits:
        memory: 8Gi
        cpu: 2000m
      requests:
        memory: 4Gi
        cpu: 1000m
```

## Troubleshooting

### Status Checks

```bash
kubectl get supabaseproject my-supabase -o yaml | yq '.status'
```

### Common Issues

**Phase: ValidatingDependencies**
- Check that database and storage secrets exist
- Verify secret keys are correctly named
- Test database connectivity

**Phase: Failed**
- Check status message: `kubectl get supabaseproject my-supabase -o jsonpath='{.status.message}'`
- Review conditions: `kubectl get supabaseproject my-supabase -o jsonpath='{.status.conditions}'`
- Check operator logs

**Component Not Ready**
- Inspect component status: `kubectl get supabaseproject my-supabase -o jsonpath='{.status.components.<component>}'`
- Check pod logs: `kubectl logs <component-pod-name>`
- Verify resource availability in cluster

## References

- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Kubebuilder Markers](https://book.kubebuilder.io/reference/markers.html)
- [Supabase Documentation](https://supabase.com/docs)
