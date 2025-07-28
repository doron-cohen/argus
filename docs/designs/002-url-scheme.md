# Design: API and Web UI URL Scheme

## Context
**Problem**: Current API structure has potential conflicts between modules and between API routes and future UI routes
**Solution**: Establish a clear URL scheme that separates API routes by module and prevents conflicts with UI routes

## Goals
- Establish clear separation between API routes and UI routes
- Prevent conflicts between different API modules
- Enable independent versioning per module
- Provide clean, semantic URL structure

## Constraints
- Must work with existing Go backend structure
- Should support future UI development
- Must maintain backward compatibility where possible

## Design

### API Path Structure (Updated)
**API Paths (per module with versioning):**
- `/api/catalog/v1/*` (API module - component catalog)
- `/api/reports/v1/*` (Reports module)
- `/api/sync/v1/*` (Sync module)

**UI Paths:**
- `/` (components list)
- `/component/{id}` (component details)
- `/settings`
- `/reports` (reports UI page)
- `/sync` (sync UI page)

### Rationale for API Structure
- **Module ownership**: Each module owns its namespace completely (catalog, reports, sync)
- **Independent versioning**: Each module can evolve independently (catalog v2 while reports stays v1)
- **No conflicts**: UI routes are clean and semantic, no interference with API paths
- **Clear separation**: Management API (catalog) vs. service APIs (reports, sync)

## Tradeoffs

### API Structure
- **Chosen**: Versioned, namespaced APIs per module
- **Benefit**: No conflicts, independent evolution
- **Cost**: Slightly more verbose paths

This URL scheme provides clean separation between UI and API routes while enabling independent module evolution.
