# Argus

The all seeing software catalog is an early-stage tool designed to provide clear visibility into the state of all software components managed within an organization.

## Goals

- Maintain a catalog of software components, including essential metadata such as ownership and maintenance location.
- Track the health and maintenance status of each component through automated checks and reports.
- Highlight potential issues in component health or maintainability, helping teams identify and address risks early.

## Planned Features

1. **Component Registry**  
   A comprehensive registry of all software components, including:
   - Standard metadata (type, ownership, etc.)
   - Custom metadata fields
   - Inter-maintainer relationships

2. **Aggregated Reports**  
   Views that aggregate all reports related to a software component, with flexible grouping and filtering.

3. **Configurable Scorecards**  
   Define scorecards that combine the results of various checks into meaningful scores.  
   Example scorecards: Security, Maintainability, Health, or any custom-defined metric.

## How It Works

- The backend syncs component manifests into persistent storage and receives check reports, storing them for later analysis.
- Both built-in reporters and external integrations using a CLI or SDKs will be provided to report various information.
- A user interface presents all this information in a clear, actionable way.

This project is in its initial stages and under active development. Expect rapid changes and improvements as we define and build out the core features.

## Quick Start with Docker

The easiest way to get started with Argus is using Docker:

```bash
# Build and start all services
make docker/up-build

# Or test the complete setup
make docker/test
```

This will start:
- **PostgreSQL** database on port 5432
- **Backend** API (serving frontend) on port 8080

For development with file watching:
```bash
make docker/develop
```

See [DOCKER.md](DOCKER.md) for detailed Docker documentation.

## Development

### CI/CD

The project uses GitHub Actions for continuous integration:

- **Linting**: Uses golangci-lint with comprehensive rules for code quality
- **Testing**: Runs all tests with race detection and coverage reporting
- **Building**: Creates optimized binaries for deployment
- **Triggers**: Runs on pushes to main/master and all pull requests

The workflow is designed to be fast and efficient, focusing on essential quality checks without unnecessary deployment steps.
