#!/bin/bash

# Test script for Docker setup
set -e

echo "🐳 Testing Docker setup for Argus..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check Docker Compose version
COMPOSE_VERSION=$(docker compose version --short)
echo "📋 Docker Compose version: $COMPOSE_VERSION"

# Build images
echo "🔨 Building Docker images..."
make docker/build

# Start services
echo "🚀 Starting services..."
make docker/up-build

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check service status
echo "📊 Service status:"
make docker/status

# Test health endpoints
echo "🏥 Testing health endpoints..."

# Test PostgreSQL
echo "Testing PostgreSQL..."
if docker compose exec postgres pg_isready -U postgres -d argus; then
    echo "✅ PostgreSQL is healthy"
else
    echo "❌ PostgreSQL health check failed"
fi

# Test Backend
echo "Testing Backend..."
if curl -f http://localhost:8080/healthz > /dev/null 2>&1; then
    echo "✅ Backend is healthy"
else
    echo "❌ Backend health check failed"
fi



echo "🎉 Docker setup test completed!"
echo ""
echo "Services are running at:"
echo "  - Backend (with frontend): http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"
echo ""
echo "To view logs: make docker/logs"
echo "To stop services: make docker/down"
echo "To start with file watching: make docker/develop" 