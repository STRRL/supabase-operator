# Quickstart validation

This file records a successful run on a fresh kind cluster.

## Tested versions

```text
kind v0.32.0
Kubernetes v1.36.1
CloudNativePG chart 0.29.0
CloudNativePG 1.30.0
MinIO chart 5.4.0
MinIO RELEASE.2024-12-18T13-15-44Z
cert-manager v1.21.0
Supabase PostgreSQL 15.14.1.021
Supabase Operator image ghcr.io/strrl/supabase-operator:358f551-dirty
Supabase Operator chart 0.1.0
```

## Commands

Create the cluster and add the chart sources.

```sh
kind create cluster --name quickstart-claude --wait 120s

helm repo add cnpg https://cloudnative-pg.github.io/charts
helm repo add minio https://charts.min.io/
helm repo add jetstack https://charts.jetstack.io
helm repo update cnpg minio jetstack
```

Install CloudNativePG and create PostgreSQL.

```sh
helm upgrade --install cnpg cnpg/cloudnative-pg \
  --kube-context kind-quickstart-claude \
  --namespace cnpg-system \
  --create-namespace \
  --version 0.29.0 \
  --wait \
  --timeout 5m

kubectl --context kind-quickstart-claude \
  apply -f docs/quickstart/postgres-cluster.yaml
kubectl --context kind-quickstart-claude \
  wait --namespace supabase-quickstart \
  --for=condition=Ready \
  cluster/supabase-postgres \
  --timeout 10m
kubectl --context kind-quickstart-claude \
  get cluster supabase-postgres \
  --namespace supabase-quickstart \
  -o custom-columns='NAME:.metadata.name,READY:.status.conditions[?(@.type=="Ready")].status,IMAGE:.spec.imageName,UID:.spec.postgresUID,GID:.spec.postgresGID'
```

Confirm that PostgreSQL accepts a TLS connection.

```sh
kubectl --context kind-quickstart-claude \
  run postgres-tls-check \
  --namespace supabase-quickstart \
  --image=postgres:15-alpine \
  --restart=Never \
  --command \
  -- sh -c 'PGPASSWORD=quickstart-postgres-password psql "host=supabase-postgres-rw port=5432 dbname=postgres user=postgres sslmode=require" -Atc "select ssl from pg_stat_ssl where pid = pg_backend_pid()"'
kubectl --context kind-quickstart-claude \
  wait --namespace supabase-quickstart \
  --for=jsonpath='{.status.phase}'=Succeeded \
  pod/postgres-tls-check \
  --timeout 2m
kubectl --context kind-quickstart-claude \
  logs postgres-tls-check \
  --namespace supabase-quickstart
```

Install MinIO and confirm the declared bucket.

```sh
helm upgrade --install minio minio/minio \
  --kube-context kind-quickstart-claude \
  --namespace supabase-quickstart \
  --create-namespace \
  --version 5.4.0 \
  --values docs/quickstart/minio/values.yaml \
  --wait \
  --timeout 5m
kubectl --context kind-quickstart-claude \
  rollout status deployment/minio \
  --namespace supabase-quickstart \
  --timeout 5m
kubectl --context kind-quickstart-claude \
  run minio-check \
  --namespace supabase-quickstart \
  --image=quay.io/minio/mc:RELEASE.2024-11-21T17-21-54Z \
  --restart=Never \
  --command \
  -- sh -c 'mc alias set quickstart http://minio:9000 storage-admin quickstart-minio-password >/dev/null && mc stat quickstart/supabase-storage'
kubectl --context kind-quickstart-claude \
  wait --namespace supabase-quickstart \
  --for=jsonpath='{.status.phase}'=Succeeded \
  pod/minio-check \
  --timeout 2m
kubectl --context kind-quickstart-claude \
  logs minio-check \
  --namespace supabase-quickstart
```

Install cert-manager, apply the Secrets, build the operator, and install its
chart.

```sh
helm upgrade --install cert-manager jetstack/cert-manager \
  --kube-context kind-quickstart-claude \
  --namespace cert-manager \
  --create-namespace \
  --version v1.21.0 \
  --set crds.enabled=true \
  --wait \
  --timeout 5m

kubectl --context kind-quickstart-claude \
  apply -f docs/quickstart/database-secret.yaml
kubectl --context kind-quickstart-claude \
  apply -f docs/quickstart/storage-secret.yaml
kubectl --context kind-quickstart-claude \
  apply -f docs/quickstart/studio-basic-auth-secret.yaml

OPERATOR_IMAGE_TAG="$(bash hack/commit-hash.sh)"
make image
kind load docker-image \
  "ghcr.io/strrl/supabase-operator:${OPERATOR_IMAGE_TAG}" \
  --name quickstart-claude

helm upgrade --install supabase-operator ./helm/supabase-operator \
  --kube-context kind-quickstart-claude \
  --namespace supabase-operator-system \
  --create-namespace \
  --set image.repository=ghcr.io/strrl/supabase-operator \
  --set "image.tag=${OPERATOR_IMAGE_TAG}" \
  --set image.pullPolicy=IfNotPresent \
  --wait \
  --timeout 5m
```

Apply the project and check every component.

```sh
kubectl --context kind-quickstart-claude \
  apply -f docs/quickstart/supabase-project.yaml
kubectl --context kind-quickstart-claude \
  wait --namespace supabase-quickstart \
  --for=jsonpath='{.status.phase}'=Running \
  supabaseproject/quickstart \
  --timeout 10m
kubectl --context kind-quickstart-claude \
  get supabaseproject quickstart \
  --namespace supabase-quickstart \
  -o custom-columns='NAME:.metadata.name,PHASE:.status.phase'
kubectl --context kind-quickstart-claude \
  get deployment \
  --namespace supabase-quickstart \
  -o custom-columns='NAME:.metadata.name,READY:.status.readyReplicas,DESIRED:.spec.replicas'
```

In one terminal, forward Kong.

```sh
kubectl --context kind-quickstart-claude \
  port-forward service/quickstart-kong 18000:8000 \
  --namespace supabase-quickstart
```

In another terminal, check the unauthenticated and authenticated responses.

```sh
curl \
  --output /dev/null \
  --silent \
  --write-out 'HTTP %{http_code}\n' \
  http://127.0.0.1:18000/
curl \
  --user supabase:quickstart-studio-password \
  --location \
  --output /dev/null \
  --silent \
  --show-error \
  --write-out 'HTTP %{http_code}\n' \
  http://127.0.0.1:18000/
```

## Selected outputs

```text
CloudNativePG:
NAME                READY   IMAGE                           UID   GID
supabase-postgres   True    supabase/postgres:15.14.1.021   101   102

Database TLS check:
t

MinIO:
Name      : supabase-storage
Type      : folder
Location  : us-east-1
Objects count: 0

SupabaseProject:
NAME         PHASE
quickstart   Running

Deployments:
minio                  1   1
quickstart-auth        1   1
quickstart-kong        1   1
quickstart-meta        1   1
quickstart-postgrest   1   1
quickstart-realtime    1   1
quickstart-storage     1   1
quickstart-studio      1   1

Studio through Kong:
HTTP 401
HTTP 200
```

## Static checks

```sh
kubectl --context kind-quickstart-claude \
  apply --dry-run=client -f docs/quickstart/
helm template minio minio/minio \
  --kube-context kind-quickstart-claude \
  --namespace supabase-quickstart \
  --version 5.4.0 \
  --values docs/quickstart/minio/values.yaml
helm lint ./helm/supabase-operator
git diff --check
```
