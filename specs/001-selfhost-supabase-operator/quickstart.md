# Quickstart: Supabase Operator

## Prerequisites

1. **Kubernetes Cluster** (v1.33+)
   ```bash
   # Verify cluster version
   kubectl version --short
   ```

2. **External PostgreSQL Database**
   - PostgreSQL 14+ recommended
   - Database and user created
   - Connection details available

3. **S3-Compatible Object Storage**
   - MinIO, AWS S3, or compatible
   - Bucket created
   - Access credentials available

4. **cert-manager** (optional, for TLS)
   ```bash
   kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
   ```

## Installation

### 1. Install the Supabase Operator

```bash
helm upgrade --install supabase-operator ./helm/supabase-operator \
  --namespace supabase-operator-system \
  --create-namespace \
  --wait
```

### 2. Verify Installation

```bash
# Check operator is running
kubectl get pods -n supabase-operator-system -l app.kubernetes.io/name=supabase-operator

# Check CRD is installed
kubectl get crd supabaseprojects.supabase.strrl.dev
```

## Creating Your First Supabase Instance

### 1. Create Secrets for External Dependencies

```bash
# Create PostgreSQL configuration secret
kubectl create secret generic postgres-config \
  --from-literal=host=postgres.example.com \
  --from-literal=port=5432 \
  --from-literal=database=supabase_db \
  --from-literal=username=supabase_user \
  --from-literal=password=your-postgres-password \
  -n default

# Create S3 storage configuration secret
kubectl create secret generic s3-config \
  --from-literal=endpoint=https://s3.amazonaws.com \
  --from-literal=region=us-east-1 \
  --from-literal=bucket=my-supabase-bucket \
  --from-literal=accessKeyId=your-access-key \
  --from-literal=secretAccessKey=your-secret-key \
  -n default
```

### 2. Create a SupabaseProject

```yaml
# supabase-project.yaml
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: my-project
  namespace: default
spec:
  projectId: my-project

  database:
    secretRef:
      name: postgres-config
    sslMode: require
    maxConnections: 20

  storage:
    secretRef:
      name: s3-config
    forcePathStyle: false  # Set to true for MinIO

  # Optional: Configure specific component versions and resources
  kong:
    image: kong:2.8.1  # Optional, uses default if not specified
    resources:
      limits:
        memory: "2Gi"
        cpu: "500m"
      requests:
        memory: "1Gi"
        cpu: "250m"

  auth:
    image: supabase/gotrue:v2.177.0
    resources:
      limits:
        memory: "128Mi"
        cpu: "100m"

  # Optional: Enable ingress
  ingress:
    enabled: true
    host: my-project.example.com
    tls:
      enabled: true
      secretName: my-project-tls
```

Apply the configuration:

```bash
kubectl apply -f supabase-project.yaml
```

### 3. Monitor Deployment Progress

```bash
# Watch the deployment status
kubectl get supabaseproject my-project -w

# Check detailed status
kubectl describe supabaseproject my-project

# View component status
kubectl get pods -l app.kubernetes.io/instance=my-project

# Check conditions
kubectl get supabaseproject my-project -o jsonpath='{.status.conditions[*].type}: {.status.conditions[*].status}'
```

### 4. Verify All Components Are Running

```bash
# Check all pods are ready
kubectl wait --for=condition=Ready pods -l app.kubernetes.io/instance=my-project --timeout=300s

# Verify services are created
kubectl get services -l app.kubernetes.io/instance=my-project
```

## Accessing Your Supabase Instance

### Get Service Endpoints

```bash
# Get all endpoints
kubectl get supabaseproject my-project -o jsonpath='{.status.endpoints}'

# Get API URL
API_URL=$(kubectl get supabaseproject my-project -o jsonpath='{.status.endpoints.api}')

# Get Anon Key
ANON_KEY=$(kubectl get secret my-project-jwt -o jsonpath='{.data.ANON_KEY}' | base64 -d)

# Get Service Role Key
SERVICE_KEY=$(kubectl get secret my-project-jwt -o jsonpath='{.data.SERVICE_ROLE_KEY}' | base64 -d)
```

### Test the API

```bash
# Test health endpoint
curl -H "apikey: $ANON_KEY" \
     -H "Authorization: Bearer $ANON_KEY" \
     $API_URL/rest/v1/

# Test auth endpoint
curl -H "apikey: $ANON_KEY" \
     $API_URL/auth/v1/health
```

## Testing Scenarios

### Scenario 1: Basic Deployment
**Given**: A Kubernetes cluster with external PostgreSQL and S3
**When**: Create a SupabaseProject with minimal configuration
**Then**: All components deploy successfully and endpoints are accessible

```bash
# Create minimal project
cat <<EOF | kubectl apply -f -
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: test-basic
spec:
  projectId: test-basic
  database:
    secretRef:
      name: postgres-config
  storage:
    secretRef:
      name: s3-config
EOF

# Verify deployment
kubectl wait --for=jsonpath='{.status.phase}'=Running supabaseproject/test-basic --timeout=300s
```

### Scenario 2: Configuration Update
**Given**: A running SupabaseProject
**When**: Update resource limits
**Then**: Components undergo rolling update

```bash
# Update resources
kubectl patch supabaseproject my-project --type merge -p '
{
  "spec": {
    "kong": {
      "resources": {
        "limits": {
          "memory": "3Gi"
        }
      }
    }
  }
}'

# Watch rolling update
kubectl get pods -l app.kubernetes.io/instance=my-project -w
```

### Scenario 3: Dependency Validation
**Given**: Invalid PostgreSQL credentials
**When**: Create a SupabaseProject
**Then**: Status shows dependency validation failure

```bash
# Create with invalid credentials
kubectl create secret generic bad-postgres \
  --from-literal=host=postgres.example.com \
  --from-literal=port=5432 \
  --from-literal=database=test_db \
  --from-literal=username=wrong \
  --from-literal=password=wrong

cat <<EOF | kubectl apply -f -
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: test-invalid
spec:
  projectId: test-invalid
  database:
    secretRef:
      name: bad-postgres
  storage:
    secretRef:
      name: s3-config
EOF

# Check status
kubectl get supabaseproject test-invalid -o jsonpath='{.status.dependencies.postgresql}'
```

### Scenario 4: Component Health Monitoring
**Given**: A running SupabaseProject
**When**: A component becomes unhealthy
**Then**: Degraded condition is set

```bash
# Simulate component failure
kubectl delete pod -l app.kubernetes.io/instance=my-project,app.kubernetes.io/component=realtime

# Check degraded condition
kubectl get supabaseproject my-project -o jsonpath='{.status.conditions[?(@.type=="Degraded")].status}'

# Watch recovery
kubectl get supabaseproject my-project -o jsonpath='{.status.components.realtime}' -w
```

## Cleanup

```bash
# Delete SupabaseProject
kubectl delete supabaseproject my-project

# Verify cleanup
kubectl get pods -l app.kubernetes.io/instance=my-project

# Delete secrets
kubectl delete secret postgres-config s3-config my-project-jwt
```

## Troubleshooting

### Check Operator Logs
```bash
kubectl logs -n supabase-operator-system deployment/supabase-operator
```

### Check Component Logs
```bash
# Kong logs
kubectl logs -l app.kubernetes.io/instance=my-project,app.kubernetes.io/component=kong

# Auth logs
kubectl logs -l app.kubernetes.io/instance=my-project,app.kubernetes.io/component=auth
```

### Common Issues

1. **Dependencies Not Connected**
   - Verify PostgreSQL is accessible from cluster
   - Check S3 endpoint and credentials
   - Review firewall/network policies

2. **Components Not Ready**
   - Check pod events: `kubectl describe pod <pod-name>`
   - Verify resource limits are sufficient
   - Check for image pull errors

3. **Secrets Not Found**
   - Ensure secrets exist in the same namespace
   - Verify secret key names match configuration

## Next Steps

- Configure SMTP for email authentication
- Set up OAuth providers
- Enable monitoring with Prometheus
- Configure backup and restore procedures
- Implement network policies for security
