# Software Catalog Service

A simple service for gaining visibility into software components across your organization. Track what you have, who maintains it, and how well it's performing.

## What It Solves

Software organizations struggle with visibility. Engineering teams lose track of services, libraries, and tools scattered across repositories. Questions like "Who maintains this service?", "How healthy is our deployment pipeline?", and "Which components need security attention?" become difficult to answer at scale.

This service provides a centralized catalog of your software components with real-time health and quality metrics.

## How It Works

### Component Discovery
Each repository defines its components through a simple manifest file. The service discovers and catalogs these components, extracting metadata like maintainers, teams, and basic configuration.

### Health Reporting
External reporters collect data about component health, code quality, deployment status, and security posture. These reporters push their findings to the service through a simple API.

### Visibility Through Scoreboards
The service exposes this data through APIs that power custom scoreboards. Create views focused on deployment health, code quality, security posture, or any combination of metrics that matter to your organization.

## Architecture

Single Go service built with Fiber and GORM for simplicity:

- **Component Registry**: Stores component metadata from repository manifests
- **Check Results API**: Receives health and quality reports from external collectors
- **Query API**: Serves component data and metrics to dashboards and tools

No complex message queues or orchestration engines. Just straightforward data collection and serving.

## Data Flow

```
[Repository Manifests] → [Component Registry]
[External Reporters] → [Check Results API] → [Database]
[Query API] → [Dashboards/Tools]
```

## Current Status

This project is in early development. The MVP focuses on:
- Component discovery from repository manifests
- Basic API for component data
- Foundation for check result reporting

Future iterations will add:
- Rich scoreboards and dashboards
- Advanced reporting capabilities
- Integration with common development tools
