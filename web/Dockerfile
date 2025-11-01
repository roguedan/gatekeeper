# Multi-stage Dockerfile for Gatekeeper Frontend
# Stage 1: Build React application
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /build

# Install dependencies for node-gyp (if needed)
RUN apk add --no-cache python3 make g++

# Copy package files
COPY package*.json ./

# Install all dependencies (including devDependencies for build)
RUN npm ci --ignore-scripts

# Copy source code and configuration files
COPY . .

# Build the application
# Vite will output to /build/dist by default
RUN npm run build

# Verify build output exists
RUN ls -la /build/dist && test -f /build/dist/index.html

# Stage 2: Serve with nginx
FROM nginx:alpine

# Install curl for health checks
RUN apk --no-cache add curl

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built application from builder stage
COPY --from=builder /build/dist /usr/share/nginx/html

# Create non-root user for nginx
RUN addgroup -g 101 nginx-gatekeeper && \
    adduser -D -u 101 -G nginx-gatekeeper nginx-gatekeeper && \
    chown -R nginx-gatekeeper:nginx-gatekeeper /usr/share/nginx/html && \
    chown -R nginx-gatekeeper:nginx-gatekeeper /var/cache/nginx && \
    chown -R nginx-gatekeeper:nginx-gatekeeper /var/log/nginx && \
    chown -R nginx-gatekeeper:nginx-gatekeeper /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R nginx-gatekeeper:nginx-gatekeeper /var/run/nginx.pid

# Expose port (nginx will listen on port 80 by default, mapped to 3000 externally)
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:80/health || exit 1

# Start nginx (run as root to bind to port 80, then drop privileges)
CMD ["nginx", "-g", "daemon off;"]
