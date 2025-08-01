# Docker Setup for Argus

This document describes the Docker setup for the Argus project, including PostgreSQL database, health checks, and development features.

## Overview

The Docker setup includes:
- **PostgreSQL 16** database for persistent storage
- **Backend** Go service that serves the frontend
- **Development mode** with file watching and hot reloading
- **Health checks** for all services
- **Service dependencies** ensuring proper startup order

## Quick Start

### Prerequisites
- Docker and Docker Compose installed
- Docker Compose version 2.22.0+ (for develop feature)

### Start the entire stack
```bash
# Build and start all services
make docker/up-build

# Or use docker compose directly
docker compose up -d --build
```

### Development mode with file watching
```bash
# Start with file watching enabled
make docker/develop

# Or use docker compose directly
docker compose watch
```

## Services

### PostgreSQL
- **Image**: `postgres:16-alpine`
- **Port**: 5432
- **Database**: argus
- **Credentials**: postgres/postgres
- **Health Check**: Uses `pg_isready` to ensure database is ready

### Backend
- **Port**: 8080
- **Health Check**: HTTP GET to `/health` endpoint
- **Dependencies**: Waits for PostgreSQL to be healthy
- **Development**: Watches backend directory for changes and rebuilds



## Configuration

The application uses `config.yaml` which includes:
- PostgreSQL connection settings
- Filesystem sync pointing to test data directories
- Development-friendly sync intervals (30s)

## Makefile Commands

### Docker Build Commands
```bash
make docker/build          # Build backend
make docker/build-backend  # Build backend
```

### Docker Compose Commands
```bash
make docker/up            # Start services
make docker/up-build      # Build and start services
make docker/down          # Stop services
make docker/restart       # Restart all services
make docker/clean         # Stop and remove volumes
```

### Logging Commands
```bash
make docker/logs          # Follow all logs
make docker/logs-backend  # Follow backend logs
make docker/logs-postgres # Follow PostgreSQL logs
```

### Development Commands
```bash
make docker/develop       # Start with file watching
make docker/status        # Show service status
```

## Development Features

### File Watching
The Docker Compose setup includes the `develop` feature that watches for file changes:

- **Backend**: Watches the entire backend directory and rebuilds on changes
- **Frontend**: Watches the frontend directory and rebuilds on changes

### Health Checks
All services include health checks:
- **PostgreSQL**: Uses `pg_isready` to verify database connectivity
- **Backend**: HTTP health check at `/healthz` endpoint

### Service Dependencies
Services start in the correct order:
1. PostgreSQL starts first
2. Backend waits for PostgreSQL to be healthy

## Environment Variables

The backend service uses:
- `ARGUS_CONFIG_PATH`: Path to configuration file (default: `/app/config.yaml`)

## Volumes

- `postgres_data`: Persistent PostgreSQL data
- Backend source code is mounted for development
- Frontend source code is mounted for development

## Networks

All services use the `argus-network` bridge network for internal communication.

## Troubleshooting

### Check service status
```bash
make docker/status
```

### View logs
```bash
make docker/logs
```

### Restart services
```bash
make docker/restart
```

### Clean start
```bash
make docker/clean
make docker/up-build
```

### Health check issues
If health checks are failing:
1. Check if PostgreSQL is running: `docker compose logs postgres`
2. Check if backend can connect to database: `docker compose logs backend`

## Production Considerations

For production deployment:
1. Use environment variables for database credentials
2. Configure proper SSL/TLS for database connections
3. Set up proper logging and monitoring
4. Consider using Docker secrets for sensitive data
5. Configure proper resource limits
6. Set up backup strategies for PostgreSQL data 