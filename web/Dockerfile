# Multi-stage Dockerfile for Gatekeeper Frontend
# Stage 1: Build React application
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /build

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production --ignore-scripts

# Copy source code
COPY . .

# Build the application
# Vite will output to /build/dist by default
RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:1.25-alpine

# Install curl for health checks
RUN apk --no-cache add curl

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built application from builder stage
COPY --from=builder /build/dist /usr/share/nginx/html

# Create non-root user for nginx
RUN addgroup -g 1000 gatekeeper && \
    adduser -D -u 1000 -G gatekeeper gatekeeper && \
    chown -R gatekeeper:gatekeeper /usr/share/nginx/html && \
    chown -R gatekeeper:gatekeeper /var/cache/nginx && \
    chown -R gatekeeper:gatekeeper /var/log/nginx && \
    chown -R gatekeeper:gatekeeper /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R gatekeeper:gatekeeper /var/run/nginx.pid

# Switch to non-root user
USER gatekeeper

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
