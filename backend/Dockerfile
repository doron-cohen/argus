# Frontend build stage
FROM oven/bun:1-alpine AS frontend-builder

# Set working directory
WORKDIR /app

# Copy frontend files
COPY frontend/ ./frontend/

# Install dependencies and build
RUN bun install --cwd ./frontend
RUN bun run --cwd ./frontend build:prod

# Backend build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Copy both backend and frontend directories to preserve structure
COPY backend/ ./backend/
COPY frontend/ ./frontend/

# Copy built frontend assets
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Copy frontend Go files
COPY frontend/go.mod ./frontend/
COPY frontend/assets.go ./frontend/
COPY frontend/assets_test.go ./frontend/

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
COPY backend/api/client/go.mod backend/api/client/go.sum ./api/client/
COPY backend/sync/api/client/go.mod backend/sync/api/client/go.sum ./sync/api/client/
COPY backend/reports/api/client/go.mod backend/reports/api/client/go.sum ./reports/api/client/

# Download dependencies
WORKDIR /app/backend
RUN go mod download

# Copy source code
COPY backend/ .

# Build the application
RUN CGO_ENABLED=1 go build -ldflags="-w -s" -o /app/bin/argus ./cmd/main.go



# Runtime stage
FROM alpine:3.22.1

# Create non-root user
RUN addgroup -g 1001 -S argus && \
    adduser -u 1001 -S argus -G argus

# Set working directory first
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/argus /app/bin/argus

# Copy config file
COPY config.yaml /app/config.yaml

# Copy testdata directory for sync sources
COPY backend/tests/testdata /app/backend/tests/testdata

# Change ownership of files to non-root user
RUN chown -R argus:argus /app

# Switch to non-root user
USER argus

# Expose port
EXPOSE 8080

# Run the application
CMD ["./bin/argus"]
