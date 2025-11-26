# Build the supabase-operator binary
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Pre-download dependencies; only refetch when go.{mod,sum} change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Build the binary with a cached module/download layer
COPY . .
RUN --mount=type=cache,target=/go \
  CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
  go build -ldflags="-s -w" -o supabase-operator ./cmd/main.go

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source="https://github.com/STRRL/supabase-operator"
WORKDIR /
COPY --from=builder /workspace/supabase-operator /usr/bin/supabase-operator
# Use numeric UID/GID for Kubernetes runAsNonRoot verification
# 65532 is the UID of 'nonroot' user in distroless images
USER 65532:65532

ENTRYPOINT ["supabase-operator"]
