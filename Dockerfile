# Multi-stage Dockerfile for Gatekeeper Backend
# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
# -ldflags="-w -s" strips debug info to reduce binary size
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev') -extldflags '-static'" \
    -a -installsuffix cgo \
    -trimpath \
    -o gatekeeper \
    ./cmd/server

# Stage 2: Create minimal runtime image with distroless
FROM gcr.io/distroless/base-debian12:nonroot

# Copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gatekeeper .

# Copy migrations if they exist (needed for database initialization)
COPY --from=builder /build/deployments/migrations ./migrations

# Expose port
EXPOSE 8080

# Distroless doesn't support shell-based healthchecks
# Health checks will be handled by docker-compose/kubernetes

# Use non-root user (distroless base-debian12:nonroot uses UID 65532)
USER nonroot:nonroot

# Run the application
ENTRYPOINT ["/app/gatekeeper"]
