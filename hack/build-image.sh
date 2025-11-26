#!/usr/bin/env bash

set -exo pipefail

DIR=$(dirname $0)

COMMIT_HASH=$(bash "${DIR}"/commit-hash.sh)
IMAGE_REPO=${IMAGE_REPO:-ghcr.io/strrl/supabase-operator}

cd "${DIR}"/../

docker build \
    -t "${IMAGE_REPO}:${COMMIT_HASH}" \
    -f ./image/supabase-operator/Dockerfile \
    ./
