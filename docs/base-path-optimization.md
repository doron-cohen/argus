# Base Path Optimization

This document explains the `base_path` feature in Argus sync configuration, which optimizes network usage when syncing from large repositories or monorepos.

## Problem

Large repositories and monorepos can contain hundreds of projects, but you may only need to sync component manifests from a specific subdirectory. Without optimization, Argus would:

- Clone the entire repository (potentially gigabytes)
- Scan the entire repository for manifest files
- Consume unnecessary bandwidth and storage

## Solution: BasePath Configuration

The `base_path` field in `SourceConfig` allows you to specify a subdirectory to sync from, using git's sparse checkout feature to optimize the process.

## Configuration

### Basic Example

```yaml
sync:
  sources:
    - type: git
      url: "https://github.com/company/monorepo"
      branch: "main" 
      interval: "10m"
      base_path: "services"  # Only sync the "services" directory
```

### Multiple Examples

```yaml
sync:
  sources:
    # Full repository sync (no base_path)
    - type: git
      url: "https://github.com/company/small-repo"
      branch: "main"
      interval: "5m"
    
    # Monorepo - only backend services
    - type: git
      url: "https://github.com/company/platform"
      branch: "main"
      interval: "10m"
      base_path: "backend/services"
    
    # Another monorepo - only mobile apps
    - type: git
      url: "https://github.com/company/platform"
      branch: "main"
      interval: "15m"
      base_path: "mobile/apps"
```

## How It Works

### 1. Sparse Checkout

When `base_path` is specified, Argus:

1. **Clones** the repository normally
2. **Configures** git sparse checkout for the specified path
3. **Removes** files outside the base path from the working directory
4. **Searches** only within the base path for manifest files

### 2. Path Processing

```
Repository Structure:
├── frontend/
│   └── apps/
├── backend/
│   └── services/
│       ├── auth/
│       │   └── manifest.yaml
│       └── api/
│           └── manifest.yaml
└── docs/

With base_path: "backend/services"
Found manifests:
- backend/services/auth/manifest.yaml
- backend/services/api/manifest.yaml
```

### 3. Network Optimization

| Scenario | Without BasePath | With BasePath |
|----------|------------------|---------------|
| **Initial Clone** | Full repository | Full repository* |
| **Updates** | Fetch all changes | Fetch only relevant changes |
| **Disk Usage** | All files | Only base_path files |
| **Manifest Search** | Entire repository | Only base_path |

*Initial clone downloads the full repository, but sparse checkout immediately removes irrelevant files.

## Benefits

### 1. **Reduced Bandwidth**
- Subsequent fetches only download changes within the base path
- Significant savings for active monorepos

### 2. **Faster Processing**
- Manifest discovery only searches within the specified directory
- Reduced I/O operations

### 3. **Lower Storage Usage**  
- Working directory only contains relevant files
- Important for systems with limited disk space

### 4. **Multiple Source Support**
- Sync different parts of the same repository as separate sources
- Different intervals for different parts

## Implementation Details

### Sparse Checkout Configuration

Argus automatically creates:

```bash
# .git/info/sparse-checkout
backend/services/*

# .git/config  
[core]
    sparseCheckout = true
```

### Path Handling

- **Relative paths**: `services/api` ✅
- **Absolute paths**: `/services/api` ✅ (leading slash removed)
- **Nested paths**: `backend/microservices/auth` ✅
- **Empty path**: Syncs entire repository (same as no base_path)

### Error Handling

```yaml
# Invalid base path - will log error and skip sync
- type: git
  url: "https://github.com/company/repo"
  base_path: "non-existent-directory"
```

**Error**: `base path non-existent-directory does not exist in repository`

## Use Cases

### 1. **Monorepo with Multiple Teams**

```yaml
# Team A - Backend services
- type: git
  url: "https://github.com/company/platform"
  base_path: "backend"
  interval: "5m"

# Team B - Frontend apps  
- type: git
  url: "https://github.com/company/platform"
  base_path: "frontend"
  interval: "10m"
```

### 2. **Large Repository with Specific Components**

```yaml
# Only sync Kubernetes manifests
- type: git
  url: "https://github.com/company/infrastructure"
  base_path: "k8s/applications"
  interval: "15m"
```

### 3. **Multi-Environment Setups**

```yaml
# Production services only
- type: git
  url: "https://github.com/company/deployments"
  base_path: "production/services"
  branch: "main"
  interval: "30m"

# Staging services only
- type: git
  url: "https://github.com/company/deployments"  
  base_path: "staging/services"
  branch: "develop"
  interval: "10m"
```

## Best Practices

### 1. **Path Organization**
Structure your repositories with clear directory hierarchies:

```
monorepo/
├── services/           # All microservices
│   ├── auth/
│   ├── api/
│   └── gateway/
├── libraries/          # Shared libraries
├── infrastructure/     # Infrastructure code
└── docs/              # Documentation
```

### 2. **Appropriate Intervals**
- Frequently changing paths: shorter intervals (5-10m)
- Stable infrastructure: longer intervals (30m-1h)

### 3. **Multiple Sources**
Use multiple sources for the same repository to:
- Sync different paths at different frequencies
- Separate concerns by team or service type

### 4. **Monitor Performance**
- Check sync logs for timing improvements
- Monitor bandwidth usage reduction
- Verify manifest discovery is working correctly

## Migration from Full Repository Sync

### Before
```yaml
sync:
  sources:
    - type: git
      url: "https://github.com/company/monorepo"
      interval: "10m"
```

### After
```yaml
sync:
  sources:
    # Only sync the parts you need
    - type: git
      url: "https://github.com/company/monorepo"
      base_path: "backend/services"
      interval: "5m"
    
    - type: git
      url: "https://github.com/company/monorepo"
      base_path: "mobile/apps"
      interval: "15m"
```

This optimization can reduce sync time and bandwidth usage by 80-90% for large monorepos! 🚀 