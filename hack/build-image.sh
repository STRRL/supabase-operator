#!/usr/bin/env bash

set -exo pipefail

DIR=$(dirname $0)

COMMIT_HASH=$(bash "${DIR}"/commit-hash.sh)

PLATFORM=linux/amd64
ARCH_TAG=linux-amd64

cd "${DIR}"/../ && \
    DOCKER_BUILDKIT=1 DOCKER_DEFAULT_PLATFORM=${PLATFORM} docker build -t ghcr.io/strrl/supabase-operator:"${COMMIT_HASH}" \
    -f ./image/supabase-operator/Dockerfile ./

docker tag ghcr.io/strrl/supabase-operator:"${COMMIT_HASH}" ghcr.io/strrl/supabase-operator:"${COMMIT_HASH}-${ARCH_TAG}"
docker tag ghcr.io/strrl/supabase-operator:"${COMMIT_HASH}" ghcr.io/strrl/supabase-operator:latest

if [ -n "${IMAGE_TAG:-}" ]; then
    docker tag ghcr.io/strrl/supabase-operator:"${COMMIT_HASH}" ghcr.io/strrl/supabase-operator:"${IMAGE_TAG}"
fi
