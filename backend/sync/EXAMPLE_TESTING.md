# Example Testing Guide

This document explains how to run example tests for the Argus sync module against a real Git repository with the pushed example manifests.

## Prerequisites

1. **Repository Access**: You need a Git repository (public or accessible) that contains the example manifests
2. **Examples Directory**: The repository should have the `examples/` directory structure with test manifests
3. **Network Access**: Tests require internet connectivity to clone repositories

## Test Structure

The example tests validate:
- **Real Git Operations**: Actual cloning and fetching from repositories
- **Manifest Discovery**: Finding `.yaml` and `.yml` files in examples
- **BasePath Optimization**: Testing sparse checkout functionality  
- **End-to-End Flow**: Complete sync pipeline from Git to component creation

## Running Tests

### 1. Using Your Repository URL

Set the `ARGUS_TEST_REPO_URL` environment variable to your repository:

```bash
# Replace with your repository URL
export ARGUS_TEST_REPO_URL="https://github.com/YOUR-USERNAME/argus"

# Run example tests
cd backend
go test ./sync -run TestExample -v
```

### 2. Quick Test (Skip if Repository Not Accessible)

The tests will automatically skip if the repository is not accessible:

```bash
cd backend
go test ./sync -run TestExample -v
```

**Output if repository not accessible:**
```
=== RUN   TestExample_SyncFromRealRepository
--- SKIP: TestExample_SyncFromRealRepository (0.15s)
    example_test.go:XX: Repository https://github.com/doron-cohen/argus not accessible 
    (you may need to set ARGUS_TEST_REPO_URL env var to your fork): authentication required
```

### 3. Run Specific Test

```bash
# Test only manifest discovery
go test ./sync -run TestExample_GitClient_RealRepository -v

# Test only BasePath optimization  
go test ./sync -run TestExample_SyncWithBasePath -v

# Test complete end-to-end flow
go test ./sync -run TestExample_FullEndToEnd -v
```

## Test Scenarios

### 1. **Full Repository Sync**
- **Test**: `TestExample_SyncFromRealRepository`
- **BasePath**: `examples`
- **Expected**: Finds all 4 components (auth, api, user, infrastructure)

### 2. **Services-Only Sync** 
- **Test**: `TestExample_SyncWithBasePath_ServicesOnly`
- **BasePath**: `examples/services`
- **Expected**: Finds only 3 service components (excludes infrastructure)

### 3. **Platform-Only Sync**
- **Test**: `TestExample_SyncWithBasePath_PlatformOnly` 
- **BasePath**: `examples/platform`
- **Expected**: Finds only 1 platform component (infrastructure)

### 4. **Git Client Operations**
- **Test**: `TestExample_GitClient_RealRepository`
- **Validates**: Repository access, manifest discovery, file reading, commit retrieval

### 5. **Manifest Validation**
- **Test**: `TestExample_ManifestValidation`
- **Validates**: All example manifests parse correctly and contain expected component names

### 6. **End-to-End Pipeline**
- **Test**: `TestExample_FullEndToEnd`
- **Validates**: Complete flow from Git fetch to component creation

## Expected Repository Structure

Your test repository should contain:

```
examples/
├── services/
│   ├── auth/
│   │   └── manifest.yaml       # name: "auth-service"
│   ├── api/
│   │   └── manifest.yaml       # name: "api-gateway" 
│   └── user/
│       └── manifest.yml        # name: "user-service"
└── platform/
    └── infrastructure/
        └── manifest.yml        # name: "platform-infrastructure"
```

## Troubleshooting

### Repository Not Found
```
Error: authentication required: Repository not found
```
**Solution**: 
- Ensure the repository URL is correct
- Make sure it's publicly accessible or you have access
- Set `ARGUS_TEST_REPO_URL` to your repository

### No Examples Directory  
```
Error: base path examples does not exist in repository
```
**Solution**: 
- Push the examples directory to your repository
- Ensure the directory structure matches the expected layout

### Network Issues
```
Error: failed to clone repository: context deadline exceeded
```
**Solution**:
- Check internet connectivity
- Try running tests with longer timeout: `go test -timeout 5m`

## Performance Testing

Example tests also validate BasePath optimization:

```bash
# Time the difference between full repo vs BasePath
time go test ./sync -run TestExample_SyncFromRealRepository -v
time go test ./sync -run TestExample_SyncWithBasePath_ServicesOnly -v
```

**Expected Results**: BasePath tests should be significantly faster for large repositories due to sparse checkout optimization.

## Continuous Integration

For CI environments, set the repository URL and run example tests:

```bash
#!/bin/bash
export ARGUS_TEST_REPO_URL="https://github.com/your-org/argus"
cd backend
go test ./sync -run TestExample -v -timeout 5m
```

The tests will automatically skip if the repository is not accessible, preventing CI failures due to network issues.

## Test Data Validation

All example tests use mock repositories for database operations, so:
- ✅ **No real database required**
- ✅ **Tests are isolated and repeatable**  
- ✅ **Only Git operations use real network resources**
- ✅ **Component creation is mocked and verified**

This ensures tests are fast, reliable, and don't require complex setup while still validating the complete sync pipeline against real Git repositories with the pushed example manifests. 