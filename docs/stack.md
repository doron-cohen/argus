# Argus Backend Tech Stack & Architecture

This document summarizes the key technology and architecture decisions for Argus Milestone 1. The focus is on a simple, maintainable, and future-proof backend, following a design-first approach.

---

## Language & Framework
- **Go**: Fast, simple, and idiomatic for backend services.
- **chi**: Lightweight, idiomatic HTTP router. Chosen because it is best supported by oapi-codegen for OpenAPI endpoints, provides type-safe handlers, and built-in request validation.

## API Design
- **OpenAPI (Design-First)**: All APIs are defined in an OpenAPI spec (YAML/JSON) before implementation. This spec is the single source of truth for endpoints, requests, and responses.
- **oapi-codegen**: Generates Go types and handler interfaces from the OpenAPI spec, ensuring code stays in sync with the contract.

## Database
- **PostgreSQL**: Reliable, open-source relational database.
- **GORM**: Popular Go ORM for easy model definition, migrations, and queries.
- **Alternatives**: sqlc (type-safe, SQL-first), ent (schema-based ORM).

## Configuration
- **YAML config**: Simple configuration loaded from a YAML file.
- **env**: Lightweight library for loading environment variables.

## Testing
- **testify/require**: For expressive, fail-fast test assertions. (We use `require` for immediate test failure on assertion errors.)

## Local Development
- **Docker Compose**: Runs the Go service and PostgreSQL together for easy local development and onboarding.

## Project Structure
- **cmd/**: Entry point for running the service.
- **internal/**: Private application code, including the repository (data access layer).
- **service/**: Main service logic, organized by module (e.g., API, repository). All modules run in one service for now, but are kept separate for future splitting.

This stack is chosen for simplicity, maintainability, and ease of onboarding. As needs evolve, we can revisit and adapt these choices.
