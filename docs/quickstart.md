# Quickstart

This guide installs Supabase Operator and its PostgreSQL and object storage
dependencies on a fresh Kubernetes cluster.

The dependency manifests use local credentials for evaluation. Change every
password before exposing any service outside the cluster.

## Prerequisites

1. Kubernetes 1.33 or later
2. `kubectl`
3. Helm
4. A local copy of this repository

Add the chart repositories:

```sh
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm repo add minio https://charts.min.io/
helm repo add jetstack https://charts.jetstack.io
helm repo add strrl https://helm.strrl.dev
helm repo update cnpg minio jetstack strrl
```

## 1. Install CloudNativePG

```sh
helm upgrade --install cnpg cnpg/cloudnative-pg \
  --namespace cnpg-system \
  --create-namespace \
  --version 0.29.0 \
  --wait \
  --timeout 5m

kubectl apply -f docs/quickstart/postgres-cluster.yaml
kubectl wait \
  --namespace supabase-quickstart \
  --for=condition=Ready \
  cluster/supabase-postgres \
  --timeout 10m
```

Expected result:

```text
cluster.postgresql.cnpg.io/supabase-postgres condition met
```

The checked in Cluster manifest uses `supabase/postgres`, UID `101`, GID
`102`, and superuser access. These settings are required by the database
initialization job.

## 2. Install MinIO

```sh
helm upgrade --install minio minio/minio \
  --namespace supabase-quickstart \
  --create-namespace \
  --version 5.4.0 \
  --values docs/quickstart/minio/values.yaml \
  --wait \
  --timeout 5m

kubectl rollout status deployment/minio \
  --namespace supabase-quickstart \
  --timeout 5m
```

Expected result:

```text
deployment "minio" successfully rolled out
```

The chart values create the `supabase-storage` bucket. Its name and
credentials match `docs/quickstart/storage-secret.yaml`.

## 3. Install cert-manager and Supabase Operator

The operator webhook uses cert-manager.

```sh
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.21.0 \
  --set crds.enabled=true \
  --wait \
  --timeout 5m

helm upgrade --install supabase-operator strrl/supabase-operator \
  --namespace supabase-operator-system \
  --create-namespace \
  --version 2026.7.25 \
  --wait \
  --timeout 5m

kubectl rollout status deployment/supabase-operator \
  --namespace supabase-operator-system \
  --timeout 5m
```

Expected result:

```text
deployment "supabase-operator" successfully rolled out
```

Version `2026.7.25` is the verified Phase A release. It does not include the
Dashboard. Use a later release that includes the Dashboard before following
step 6.

## 4. Create the credential Secrets

```sh
kubectl apply -f docs/quickstart/database-secret.yaml
kubectl apply -f docs/quickstart/storage-secret.yaml
kubectl apply -f docs/quickstart/studio-basic-auth-secret.yaml
```

Expected Secrets:

```text
supabase-db-credentials
supabase-storage-credentials
supabase-studio-basic-auth
```

Secret key names are part of the operator API contract. Do not rename
`host`, `port`, `database`, `username`, `password`, `endpoint`, `region`,
`bucket`, `accessKeyId`, or `secretAccessKey`.

## 5. Create the SupabaseProject

```sh
kubectl apply -f docs/quickstart/supabase-project.yaml
kubectl wait \
  --namespace supabase-quickstart \
  --for=jsonpath='{.status.phase}'=Running \
  supabaseproject/quickstart \
  --timeout 10m

kubectl get supabaseproject quickstart \
  --namespace supabase-quickstart
```

Expected result:

```text
NAME         PHASE
quickstart   Running
```

The project uses `sslMode: require` for PostgreSQL and path style S3
requests for MinIO.

## 6. Open the Dashboard

The operator creates the Dashboard Service by default. No Helm value is
required.

Check the Service:

```sh
kubectl get service supabase-operator-dashboard \
  --namespace supabase-operator-system
```

Expected result:

```text
NAME                          TYPE        PORT(S)
supabase-operator-dashboard   ClusterIP   8080/TCP
```

Forward the Dashboard in one terminal:

```sh
kubectl port-forward \
  --namespace supabase-operator-system \
  service/supabase-operator-dashboard \
  18080:8080
```

Open <http://127.0.0.1:18080/>. The page lists each `SupabaseProject`, its
phase, and component readiness.

You can also check the Dashboard API:

```sh
curl \
  --fail \
  --silent \
  --show-error \
  http://127.0.0.1:18080/api/projects
```

Expected response:

```json
{"projects":[...]}
```

The Dashboard has no built in authentication. Its Service is `ClusterIP` by
default. Keep the port forward on localhost and do not expose the Service
outside the cluster.

## 7. Open Studio

Forward Kong in one terminal:

```sh
kubectl port-forward \
  --namespace supabase-quickstart \
  service/quickstart-kong \
  18000:8000
```

Open <http://127.0.0.1:18000/> and enter:

```text
Username: supabase
Password: quickstart-studio-password
```

The browser should load Supabase Studio. You can also check the response:

```sh
curl \
  --user supabase:quickstart-studio-password \
  --location \
  --output /dev/null \
  --silent \
  --show-error \
  --write-out 'HTTP %{http_code}\n' \
  http://127.0.0.1:18000/
```

Expected result:

```text
HTTP 200
```

## Troubleshooting

Check project status and operator logs:

```sh
kubectl get supabaseproject quickstart \
  --namespace supabase-quickstart \
  -o yaml

kubectl logs deployment/supabase-operator \
  --namespace supabase-operator-system
```

The complete validation record is in
[`docs/quickstart/VALIDATION.md`](quickstart/VALIDATION.md).
