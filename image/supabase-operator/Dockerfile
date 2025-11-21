# Build the supabase-operator binary
FROM golang:1.25 AS builder

WORKDIR /workspace

# Pre-download dependencies; only refetch when go.{mod,sum} change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Build the binary with a cached module/download layer
COPY . .
RUN --mount=type=cache,target=/go \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -ldflags="-s -w" -o supabase-operator ./cmd/main.go

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source="https://github.com/STRRL/supabase-operator"
WORKDIR /
COPY --from=builder /workspace/supabase-operator /usr/bin/supabase-operator
USER nonroot:nonroot

ENTRYPOINT ["supabase-operator"]
