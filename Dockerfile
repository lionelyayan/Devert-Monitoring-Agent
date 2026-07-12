# ─────────────────────────────────────────────────────────────
# Stage 1: Build
# ─────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Download modules first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build args injected by GitHub Actions
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown

# Build static binary with version info embedded
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build \
    -ldflags="-w -s \
      -X main.version=${VERSION} \
      -X main.buildDate=${BUILD_DATE} \
      -X main.gitCommit=${GIT_COMMIT}" \
    -o /bin/monitor-agent \
    ./cmd/agent

# ─────────────────────────────────────────────────────────────
# Stage 2: Runtime (minimal)
# ─────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S agent && adduser -S agent -G agent

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bin/monitor-agent .

# Migrations directory
COPY migrations/ ./migrations/

USER agent

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./monitor-agent"]
