# Roadmap

## Milestone 1: Simple Component Registry ✅

- Sync component manifests from a public git repository.
- Update the component registry in the persistent storage as changes occur.
- Expose the list of components through the API (unauthenticated).
- No checks, reports, or advanced features at this stage.
- The manifest format will be minimal and straightforward.

This milestone establishes the foundation for Argus as a basic software catalog.

## Milestone 2: Quality Checks & Reports

- Define and run quality checks against components (code coverage, security scans, test results).
- Store check results with status tracking (pass, fail, disabled, skipped).
- Track historical check data with timestamps and metadata for analytics.
- Group and filter results by metadata (team, environment, version) for insights.
- Monitor check performance trends and success rates over time.

This milestone adds quality monitoring capabilities to the software catalog.
