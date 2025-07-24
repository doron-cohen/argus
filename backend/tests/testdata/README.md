# Example Manifests for Sync Testing

This directory contains example component manifests for testing the Argus VCS sync functionality.

## Directory Structure

```
examples/
├── services/           # Backend microservices
│   ├── auth/
│   │   └── manifest.yaml
│   ├── api/
│   │   └── manifest.yaml
│   └── user/
│       └── manifest.yml
└── platform/          # Platform/infrastructure components
    └── infrastructure/
        └── manifest.yml
```

## Components

### Services (examples/services/)
- **auth-service** - Authentication and authorization service
- **api-gateway** - Main API gateway routing requests
- **user-service** - User management and profiles

### Platform (examples/platform/)
- **platform-infrastructure** - Core infrastructure and monitoring

## Testing Sync Configuration

You can test the sync functionality with different BasePath configurations:

### 1. Sync All Examples
```yaml
sync:
  sources:
    - type: git
      url: "https://github.com/your-username/argus"
      branch: "main"
      base_path: "examples"
      interval: "1m"
```

### 2. Sync Only Services
```yaml
sync:
  sources:
    - type: git
      url: "https://github.com/your-username/argus"
      branch: "main"
      base_path: "examples/services"
      interval: "1m"
```

### 3. Sync Only Platform Components
```yaml
sync:
  sources:
    - type: git
      url: "https://github.com/your-username/argus"
      branch: "main"
      base_path: "examples/platform"
      interval: "1m"
```

### 4. Multiple Sources (Test BasePath Optimization)
```yaml
sync:
  sources:
    # Services only
    - type: git
      url: "https://github.com/your-username/argus"
      branch: "main"
      base_path: "examples/services"
      interval: "1m"
    
    # Platform only
    - type: git
      url: "https://github.com/your-username/argus"
      branch: "main"
      base_path: "examples/platform"
      interval: "2m"
```

## Expected Sync Results

When syncing successfully, Argus should discover and create these components:
- `auth-service`
- `api-gateway` 
- `user-service`
- `platform-infrastructure`

The sync service will log the discovery and creation of each component, allowing you to verify the BasePath optimization is working correctly. 