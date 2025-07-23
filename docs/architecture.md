# Architecture Overview

Argus is designed with simplicity and modularity in mind. The architecture consists of the following main parts:

## Server

The server is organized into several modules:

- **Repository**: Handles access to the persistent storage layer.
- **Ingest API**: Receives and processes check reports from external sources.
- **API**: Exposes data and functionality to clients (UI, SDK, CLI, etc.).
- **Sync Module**: Responsible for syncing component manifests from version control systems.

These modules are kept separate for clarity and maintainability. In the future, they may be deployed independently, but for now, they are part of a single service.

## UI

The user interface is served from the root path of the service by default, making it easy to access. Alternative hosting options can be supported if needed.

## Reporters

The reporters module includes:
- **Client SDK**: For integrating with other tools and services.
- **CLI**: Command-line tool for reporting and interacting with the server.
- **Built-in Reporters**: Deployable components that automatically report on application state or health.

This structure aims to keep the project organized, testable, and easy to extend.
