# Quick Start: Supabase Operator

This guide walks you through installing the Supabase Operator on Kubernetes and deploying your first Supabase project. It assumes you already run a Kubernetes cluster and have administrative access to it.

## Prerequisites
- Kubernetes cluster running v1.33 or later
- `kubectl` configured to talk to the target cluster
- External PostgreSQL database reachable from the cluster
- S3-compatible object storage (for example, MinIO or Amazon S3)
- Administrative access for creating Kubernetes secrets and applying manifests

## Step 1 – Install the Operator
### Option A: Use the Published Manifests (Recommended)
```bash
kubectl apply -f https://raw.githubusercontent.com/strrl/supabase-operator/main/config/install.yaml
```

### Option B: Install from Source
```bash
git clone https://github.com/strrl/supabase-operator
cd supabase-operator
make install
make deploy
```

Verify that the controller manager pods are running:
```bash
kubectl get pods -n supabase-operator-system
```

## Step 2 – Provide Database Credentials
Create a Kubernetes secret that stores your PostgreSQL connection details. Replace the placeholder values with your own database credentials before running the command.
```bash
kubectl create secret generic postgres-config \
  --from-literal=host=postgres.example.com \
  --from-literal=port=5432 \
  --from-literal=database=supabase \
  --from-literal=username=postgres \
  --from-literal=password=your-secure-password
```

## Step 3 – Provide Storage Credentials
Create a Kubernetes secret for your S3-compatible object storage. The keys must use camelCase names (`accessKeyId`, `secretAccessKey`).
```bash
kubectl create secret generic s3-config \
  --from-literal=endpoint=https://s3.example.com \
  --from-literal=region=us-east-1 \
  --from-literal=bucket=supabase-storage \
  --from-literal=accessKeyId=your-access-key \
  --from-literal=secretAccessKey=your-secret-key
```

If you use a self-hosted endpoint such as MinIO, set `--from-literal=forcePathStyle=true` to avoid virtual-hosted-style bucket addressing.

## Step 4 – Configure Studio Basic Auth (Recommended)
Kong serves Supabase Studio behind HTTP basic authentication. Create a secret that stores the credentials you want to require when accessing the dashboard.

```bash
kubectl create secret generic studio-dashboard-creds \
  --from-literal=username=supabase \
  --from-literal=password='choose-a-strong-password'
```

You can keep the sample username for local testing, but **always** change the password before exposing Kong outside the cluster.

## Step 5 – Deploy a Supabase Project
Create a file named `my-supabase.yaml` with the following content. Adjust the values to match your project requirements.
```yaml
apiVersion: supabase.strrl.dev/v1alpha1
kind: SupabaseProject
metadata:
  name: my-supabase
  namespace: default
spec:
  projectId: my-supabase-project
  database:
    secretRef:
      name: postgres-config
    sslMode: require
    maxConnections: 50
  storage:
    secretRef:
      name: s3-config
    forcePathStyle: true
  studio:
    dashboardBasicAuthSecretRef:
      name: studio-dashboard-creds
```
Apply the manifest:
```bash
kubectl apply -f my-supabase.yaml
```

## Step 6 – Confirm the Deployment
Check that the custom resource has been created and is progressing.
```bash
kubectl get supabaseproject my-supabase -o yaml
```
Inspect component-level status when the operator reports ready state:
```bash
kubectl get supabaseproject my-supabase -o jsonpath='{.status.components}'
```

List the services belonging to the Supabase stack:
```bash
kubectl get services -l app.kubernetes.io/part-of=supabase
```

## Step 7 – Access the Supabase APIs and Studio
Expose services locally using `kubectl port-forward`. The example below forwards the Kong API Gateway.
```bash
kubectl port-forward svc/my-supabase-kong 8000:8000
```
You can now interact with the Supabase REST endpoints through `http://localhost:8000`. Requests to the root path (or `/` in a browser) respond with `401 Unauthorized` until you provide the dashboard credentials from `studio-dashboard-creds`.

Port-forward the Studio service to open the management UI:
```bash
kubectl port-forward svc/my-supabase-studio 3000:3000
```
Then browse to `http://localhost:3000` to manage your project through Studio. The operator injects the generated Supabase keys so Studio can talk to your stack without additional configuration. Remember that port-forwarding the Studio service bypasses Kong, so expose it only on trusted networks.

To access other components, port-forward the corresponding services (for example, `my-supabase-gotrue`) or configure an Ingress of your choice.

## Step 8 – Retrieve Supabase API Keys
The operator stores generated API keys in a secret named `<project>-jwt` in the same namespace as your `SupabaseProject`.
```bash
# Retrieve the public ANON key
ANON_KEY=$(kubectl get secret my-supabase-jwt \
  -o jsonpath='{.data.anon-key}' | base64 -d)

# Retrieve the Service Role key
SERVICE_ROLE_KEY=$(kubectl get secret my-supabase-jwt \
  -o jsonpath='{.data.service-role-key}' | base64 -d)

# Optional: read the API endpoint published in status
API_URL=$(kubectl get supabaseproject my-supabase \
  -o jsonpath='{.status.endpoints.api}')
```
Use `ANON_KEY` for client-side requests and reserve `SERVICE_ROLE_KEY` for backend jobs that require elevated privileges.

## Step 9 – Connect to Your Database
Supabase relies on the external PostgreSQL database you referenced in `postgres-config`. You can reuse the stored credentials to build a connection string for `psql` or other tools.
```bash
POSTGRES_HOST=$(kubectl get secret postgres-config -o jsonpath='{.data.host}' | base64 -d)
POSTGRES_PORT=$(kubectl get secret postgres-config -o jsonpath='{.data.port}' | base64 -d)
POSTGRES_DB=$(kubectl get secret postgres-config -o jsonpath='{.data.database}' | base64 -d)
POSTGRES_USER=$(kubectl get secret postgres-config -o jsonpath='{.data.username}' | base64 -d)
POSTGRES_PASSWORD=$(kubectl get secret postgres-config -o jsonpath='{.data.password}' | base64 -d)

psql "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"
```
If the database is only reachable from within the cluster, create a temporary pod (for example, `kubectl run pg-client --rm -it --image=postgres:17 -- bash`) and connect from there, or establish the necessary network tunnel from your workstation.

## Step 10 – Clean Up (Optional)
Remove the sample project when you are done testing:
```bash
kubectl delete supabaseproject my-supabase
```
If you installed the operator only for evaluation, remove it as well:
```bash
kubectl delete -f https://raw.githubusercontent.com/strrl/supabase-operator/main/config/install.yaml
```

## Next Steps
- Review the CRD specification under `config/crd/bases` to customize per-component settings such as replica counts and resource limits.
- Explore `docs/database-initialization.md` for bootstrap guidance on preparing your PostgreSQL database.
- Add cluster ingress or load balancer resources to expose Supabase services publicly.
