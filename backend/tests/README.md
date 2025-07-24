# Integration Tests

This directory contains end-to-end integration tests for the Argus backend, using real PostgreSQL databases and testing the complete sync + API pipeline.

## Structure

```
tests/
├── README.md                 # This file
├── main_test.go             # Test setup with testcontainers PostgreSQL
├── health_test.go           # Health endpoint tests
├── components_test.go       # Basic components API tests
├── sync_integration_test.go # Full sync + API integration tests
└── testdata/                # Test manifest data (moved from examples/)
    ├── README.md
    ├── services/
    │   ├── auth/manifest.yaml
    │   ├── api/manifest.yaml
    │   └── user/manifest.yml
    └── platform/
        └── infrastructure/manifest.yml
```

## Test Categories

### 🔧 **Basic Integration Tests**
- **Health Tests**: API health endpoint validation
- **Components API**: Basic CRUD operations via API
- **Database**: PostgreSQL integration with testcontainers

### 🚀 **Sync Integration Tests** 
**Complete end-to-end sync testing with real PostgreSQL + API validation:**

#### **Filesystem Source Tests**
- `TestFilesystemSyncIntegration`: Full filesystem sync of all testdata
- `TestFilesystemSyncWithBasePath`: BasePath filtering (services only)
- `TestSyncPerformance`: High-frequency sync performance testing

#### **Git Source Tests**  
- `TestGitSyncIntegration`: Full Git repository sync with sparse checkout
- Git tests use the actual Argus repository with `backend/tests/testdata` BasePath

#### **Mixed Source Tests**
- `TestMixedSourcesIntegration`: Filesystem + Git sources in same configuration
- Tests services from filesystem, platform from Git repository

#### **Error Handling Tests**
- `TestSyncWithNoSources`: Graceful handling of empty source configuration
- `TestSyncErrorHandling`: Invalid paths and network failures

## Running Tests

### **All Integration Tests**
```bash
go test ./tests -v
```

### **Short Mode (Skip Integration)**
```bash
go test ./tests -short
```

### **Specific Test Categories**
```bash
# Filesystem sync only
go test ./tests -run TestFilesystemSync -v

# Git sync only (requires network)
go test ./tests -run TestGitSync -v

# Mixed sources (requires network)
go test ./tests -run TestMixedSources -v

# Error handling
go test ./tests -run TestSyncError -v
```

### **With Custom Git Repository**
```bash
export ARGUS_TEST_REPO_URL="https://github.com/your-fork/argus"
go test ./tests -run TestGitSync -v
```

## Test Infrastructure

### **TestContainers PostgreSQL**
- Each test gets a fresh PostgreSQL container
- Automatic setup/teardown with proper cleanup
- No external dependencies or shared state

### **Test Data Structure**
```
testdata/
├── services/          # 3 service components
│   ├── auth/         # auth-service
│   ├── api/          # api-gateway  
│   └── user/         # user-service
└── platform/         # 1 platform component
    └── infrastructure/ # platform-infrastructure
```

### **API Validation**
- Server starts with real sync configuration
- Tests wait for sync completion (1-10 seconds)
- API client validates components via GET /api/components
- Component names and counts verified

## Test Scenarios

### **1. Filesystem Source (Fast)**
```yaml
sync:
  sources:
    - type: filesystem
      path: "./testdata"
      interval: "1s"
```
**Expected**: 4 components synced instantly, API returns all components

### **2. Filesystem with BasePath**
```yaml
sync:
  sources:
    - type: filesystem
      path: "./testdata"  
      base_path: "services"
      interval: "1s"
```
**Expected**: 3 service components only, no platform components

### **3. Git Source (Network)**
```yaml
sync:
  sources:
    - type: git
      url: "https://github.com/doron-cohen/argus"
      branch: "main"
      base_path: "backend/tests/testdata"
      interval: "1s"
```
**Expected**: 4 components from Git repository, sparse checkout optimization

### **4. Mixed Sources**
```yaml
sync:
  sources:
    # Services from filesystem
    - type: filesystem
      path: "./testdata"
      base_path: "services"
      interval: "1s"
    # Platform from Git  
    - type: git
      url: "https://github.com/doron-cohen/argus"
      base_path: "backend/tests/testdata/platform"
      interval: "1s"
```
**Expected**: 4 total components (3 from filesystem, 1 from Git)

## Performance Expectations

### **Filesystem Sync**
- **Initial sync**: <1 second (4 components)
- **Subsequent syncs**: <100ms (cached)
- **High frequency**: 100ms intervals sustainable
- **Memory**: Minimal overhead

### **Git Sync**  
- **Initial clone**: 3-5 seconds (with BasePath optimization)
- **Subsequent fetches**: 1-2 seconds (sparse checkout)
- **Network usage**: 80-90% reduction with BasePath
- **Storage**: Only relevant files downloaded

### **Mixed Sources**
- **Combined sync**: 5-8 seconds total
- **Parallel execution**: Sources sync independently  
- **Optimal configuration**: Fast filesystem + slower Git intervals

## Troubleshooting

### **Git Tests Failing**
```bash
# Check repository access
export ARGUS_TEST_REPO_URL="https://github.com/your-fork/argus"

# Skip Git tests
go test ./tests -run "Test.*Filesystem" -v
```

### **Test Timeouts**
```bash
# Increase timeout for slow networks
go test ./tests -timeout 60s -v
```

### **Container Issues**
```bash
# Check Docker daemon
docker ps

# Clean up orphaned containers
docker system prune -f
```

### **Database Connection Issues**
- Tests automatically handle PostgreSQL startup
- Each test gets isolated database
- No manual setup required

## Integration with CI/CD

### **Fast Tests** (Filesystem only)
```bash
go test ./tests -run "Test.*Filesystem" -timeout 30s
```

### **Full Tests** (Including Git)
```bash
export ARGUS_TEST_REPO_URL="https://github.com/company/argus-fork"  
go test ./tests -timeout 120s
```

### **Parallel Execution**
```bash
go test ./tests -parallel 4 -timeout 180s
```

This comprehensive integration test suite ensures the entire sync + API pipeline works correctly across all source types and configurations! 🚀 