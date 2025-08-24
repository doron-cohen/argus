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
- Track historical check data with timestamps and metadata.
- Expose check results through the API for component details.
- Support basic check status filtering and querying.

This milestone adds quality monitoring capabilities to the software catalog.

## Milestone 3: Basic Web UI

- **Components List Page**: Display all components in a simple list view with basic metadata.
- **Component Details Page**: Show individual component information including:
  - All component metadata
  - Latest check results with status indicators
  - Complete list of all checks with their current statuses
  - Historical timeline view of all submitted checks
- **Settings Page**: Display repository configuration and sync settings.
- **Navigation**: Basic navigation between pages starting from the root path.
- **Simple Design**: Minimal, functional UI focused on usability over aesthetics.

This milestone provides a basic web interface for interacting with the component registry and quality check data.

### Design Notes
- See `docs/designs/006-css-tokens-and-theme.md` for the CSS tokens and theme groundwork used by UI components (initially `ui-badge`).
