# Multi-stage Dockerfile for Gatekeeper Backend
# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
# -ldflags="-w -s" strips debug info to reduce binary size
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -a -installsuffix cgo \
    -o gatekeeper \
    ./cmd/server

# Stage 2: Create minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl tzdata && \
    addgroup -g 1000 gatekeeper && \
    adduser -D -u 1000 -G gatekeeper gatekeeper

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gatekeeper .

# Copy migrations (needed for database initialization)
COPY --from=builder /build/deployments/migrations ./migrations

# Set ownership
RUN chown -R gatekeeper:gatekeeper /app

# Switch to non-root user
USER gatekeeper

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/app/gatekeeper"]
