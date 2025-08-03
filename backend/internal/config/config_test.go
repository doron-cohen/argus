package config

import (
	"os"
	"testing"
	"time"

	"github.com/doron-cohen/argus/backend/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify storage defaults
	assert.Equal(t, "localhost", cfg.Storage.Host)
	assert.Equal(t, 5432, cfg.Storage.Port)
	assert.Equal(t, "postgres", cfg.Storage.User)
	assert.Equal(t, "postgres", cfg.Storage.Password)
	assert.Equal(t, "argus", cfg.Storage.DBName)
	assert.Equal(t, "disable", cfg.Storage.SSLMode)

	// Verify sync defaults
	assert.Len(t, cfg.Sync.Sources, 0)
}

func TestLoadConfig_NoConfigFile(t *testing.T) {
	// Clear environment variable to test default behavior
	err := os.Unsetenv("ARGUS_CONFIG_PATH")
	require.NoError(t, err)

	// Remove any existing config.yaml
	if err := os.Remove("config.yaml"); err != nil && !os.IsNotExist(err) {
		t.Logf("Failed to remove config.yaml: %v", err)
	}

	// Load config should succeed with defaults
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify default values are used
	assert.Equal(t, "localhost", cfg.Storage.Host)
	assert.Equal(t, 5432, cfg.Storage.Port)
	assert.Equal(t, "postgres", cfg.Storage.User)
	assert.Equal(t, "postgres", cfg.Storage.Password)
	assert.Equal(t, "argus", cfg.Storage.DBName)
	assert.Equal(t, "disable", cfg.Storage.SSLMode)
	assert.Len(t, cfg.Sync.Sources, 0)
}

func TestLoadConfig_EnvironmentVariableOverrides(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"ARGUS_STORAGE_HOST":     "env-host",
		"ARGUS_STORAGE_PORT":     "5433",
		"ARGUS_STORAGE_USER":     "env-user",
		"ARGUS_STORAGE_PASSWORD": "env-pass",
		"ARGUS_STORAGE_DBNAME":   "env-db",
		"ARGUS_STORAGE_SSLMODE":  "require",
	}

	for key, value := range envVars {
		err := os.Setenv(key, value)
		require.NoError(t, err)
		defer func(key string) {
			if err := os.Unsetenv(key); err != nil {
				t.Logf("Failed to unset environment variable %s: %v", key, err)
			}
		}(key)
	}

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify environment variables override defaults
	assert.Equal(t, "env-host", cfg.Storage.Host)
	assert.Equal(t, 5433, cfg.Storage.Port)
	assert.Equal(t, "env-user", cfg.Storage.User)
	assert.Equal(t, "env-pass", cfg.Storage.Password)
	assert.Equal(t, "env-db", cfg.Storage.DBName)
	assert.Equal(t, "require", cfg.Storage.SSLMode)
}

func TestLoadConfig_EnvironmentVariableOverridesConfigFile(t *testing.T) {
	// Create a config file with some values
	configContent := `
storage:
  host: file-host
  port: 5434
  user: file-user
  password: file-pass
  dbname: file-db
  sslmode: disable
`
	err := os.WriteFile("test-config-env-override.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove("test-config-env-override.yaml"); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	// Set environment variables to override config file
	envVars := map[string]string{
		"ARGUS_CONFIG_PATH":      "test-config-env-override.yaml",
		"ARGUS_STORAGE_HOST":     "env-host",
		"ARGUS_STORAGE_PORT":     "5435",
		"ARGUS_STORAGE_USER":     "env-user",
		"ARGUS_STORAGE_PASSWORD": "env-pass",
		"ARGUS_STORAGE_DBNAME":   "env-db",
		"ARGUS_STORAGE_SSLMODE":  "require",
	}

	for key, value := range envVars {
		err := os.Setenv(key, value)
		require.NoError(t, err)
		defer func(key string) {
			if err := os.Unsetenv(key); err != nil {
				t.Logf("Failed to unset environment variable %s: %v", key, err)
			}
		}(key)
	}

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify environment variables override config file values
	assert.Equal(t, "env-host", cfg.Storage.Host)
	assert.Equal(t, 5435, cfg.Storage.Port)
	assert.Equal(t, "env-user", cfg.Storage.User)
	assert.Equal(t, "env-pass", cfg.Storage.Password)
	assert.Equal(t, "env-db", cfg.Storage.DBName)
	assert.Equal(t, "require", cfg.Storage.SSLMode)
}

func TestLoadConfig_InvalidPortEnvironmentVariable(t *testing.T) {
	// Set invalid port environment variable
	err := os.Setenv("ARGUS_STORAGE_PORT", "invalid-port")
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_STORAGE_PORT"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Load config should succeed with default port
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify default port is used when environment variable is invalid
	assert.Equal(t, 5432, cfg.Storage.Port)
}

func TestGetEnvironmentVariables(t *testing.T) {
	// Set some ARGUS_ environment variables
	envVars := map[string]string{
		"ARGUS_STORAGE_HOST": "test-host",
		"ARGUS_STORAGE_PORT": "5433",
		"ARGUS_CUSTOM_VAR":   "custom-value",
	}

	for key, value := range envVars {
		err := os.Setenv(key, value)
		require.NoError(t, err)
		defer func(key string) {
			if err := os.Unsetenv(key); err != nil {
				t.Logf("Failed to unset environment variable %s: %v", key, err)
			}
		}(key)
	}

	// Get environment variables
	result := GetEnvironmentVariables()

	// Verify all ARGUS_ variables are returned
	assert.Equal(t, "test-host", result["ARGUS_STORAGE_HOST"])
	assert.Equal(t, "5433", result["ARGUS_STORAGE_PORT"])
	assert.Equal(t, "custom-value", result["ARGUS_CUSTOM_VAR"])
	assert.Len(t, result, 3)
}

func TestLoadConfig_FromEnvironment(t *testing.T) {
	// Copy testdata file to current directory
	srcFile := "testdata/environment.yaml"
	dstFile := "test-config.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(dstFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	// Set environment variable
	err = os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_CONFIG_PATH"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Debug: Print the environment variable
	t.Logf("ARGUS_CONFIG_PATH set to: %s", os.Getenv("ARGUS_CONFIG_PATH"))

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify storage config
	assert.Equal(t, "test-host", cfg.Storage.Host)
	assert.Equal(t, 5433, cfg.Storage.Port)
	assert.Equal(t, "test-user", cfg.Storage.User)
	assert.Equal(t, "test-pass", cfg.Storage.Password)
	assert.Equal(t, "test-db", cfg.Storage.DBName)
	assert.Equal(t, "require", cfg.Storage.SSLMode)

	// Verify sync config
	assert.Len(t, cfg.Sync.Sources, 2)

	// Check filesystem source
	fsSource := cfg.Sync.Sources[0]
	fsConfig := fsSource.GetConfig()
	assert.NotNil(t, fsConfig, "filesystem config should not be nil")
	assert.Equal(t, "filesystem", fsConfig.GetSourceType())
	assert.Equal(t, "/test/path", fsConfig.(*sync.FilesystemSourceConfig).Path)
	assert.Equal(t, 30*time.Second, fsConfig.GetInterval())

	// Check git source
	gitSource := cfg.Sync.Sources[1]
	gitConfig := gitSource.GetConfig()
	assert.NotNil(t, gitConfig, "git config should not be nil")
	assert.Equal(t, "git", gitConfig.GetSourceType())
	assert.Equal(t, "https://github.com/test/repo", gitConfig.(*sync.GitSourceConfig).URL)
	assert.Equal(t, "main", gitConfig.(*sync.GitSourceConfig).Branch)
	assert.Equal(t, 5*time.Minute, gitConfig.GetInterval())
}

func TestLoadConfig_DefaultBehavior(t *testing.T) {
	// Copy testdata file to current directory as config.yaml
	srcFile := "testdata/default.yaml"
	dstFile := "config.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(dstFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	// Clear environment variable to test default behavior
	err = os.Unsetenv("ARGUS_CONFIG_PATH")
	require.NoError(t, err)

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify storage config
	assert.Equal(t, "localhost", cfg.Storage.Host)
	assert.Equal(t, 5432, cfg.Storage.Port)
	assert.Equal(t, "default-user", cfg.Storage.User)
	assert.Equal(t, "default-pass", cfg.Storage.Password)
	assert.Equal(t, "default-db", cfg.Storage.DBName)
	assert.Equal(t, "disable", cfg.Storage.SSLMode)

	// Verify sync config
	assert.Len(t, cfg.Sync.Sources, 1)

	fsSource := cfg.Sync.Sources[0]
	fsConfig := fsSource.GetConfig()
	assert.NotNil(t, fsConfig, "filesystem config should not be nil")
	assert.Equal(t, "filesystem", fsConfig.GetSourceType())
	assert.Equal(t, "./local/path", fsConfig.(*sync.FilesystemSourceConfig).Path)
	assert.Equal(t, 1*time.Minute, fsConfig.GetInterval())
}

func TestLoadConfig_EmptySyncSources(t *testing.T) {
	// Copy testdata file to current directory
	srcFile := "testdata/empty-sync.yaml"
	dstFile := "test-config-empty.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(dstFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	err = os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_CONFIG_PATH"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify sync config is empty
	assert.Len(t, cfg.Sync.Sources, 0)
}

func TestLoadConfig_MissingSyncSection(t *testing.T) {
	// Copy testdata file to current directory
	srcFile := "testdata/no-sync.yaml"
	dstFile := "test-config-no-sync.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(dstFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	err = os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_CONFIG_PATH"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify sync config is empty (default behavior)
	assert.Len(t, cfg.Sync.Sources, 0)
}

func TestLoadConfig_ComplexSyncSources(t *testing.T) {
	// Copy testdata file to current directory
	srcFile := "testdata/complex-sync.yaml"
	dstFile := "test-config-complex.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(dstFile); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	err = os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_CONFIG_PATH"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Load config
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify storage config
	assert.Equal(t, "complex-host", cfg.Storage.Host)
	assert.Equal(t, "complex-user", cfg.Storage.User)

	// Verify sync config has 3 sources
	assert.Len(t, cfg.Sync.Sources, 3)

	// Check filesystem source
	fsSource := cfg.Sync.Sources[0]
	fsConfig := fsSource.GetConfig()
	assert.NotNil(t, fsConfig, "filesystem config should not be nil")
	assert.Equal(t, "filesystem", fsConfig.GetSourceType())
	assert.Equal(t, "/path/with/spaces", fsConfig.(*sync.FilesystemSourceConfig).Path)
	assert.Equal(t, 2*time.Minute, fsConfig.GetInterval())

	// Check first git source
	gitSource1 := cfg.Sync.Sources[1]
	gitConfig1 := gitSource1.GetConfig()
	assert.NotNil(t, gitConfig1, "first git config should not be nil")
	assert.Equal(t, "git", gitConfig1.GetSourceType())
	assert.Equal(t, "https://github.com/org/monorepo", gitConfig1.(*sync.GitSourceConfig).URL)
	assert.Equal(t, "develop", gitConfig1.(*sync.GitSourceConfig).Branch)
	assert.Equal(t, "services/backend", gitConfig1.(*sync.GitSourceConfig).BasePath)
	assert.Equal(t, 10*time.Minute, gitConfig1.GetInterval())

	// Check second git source
	gitSource2 := cfg.Sync.Sources[2]
	gitConfig2 := gitSource2.GetConfig()
	assert.NotNil(t, gitConfig2, "second git config should not be nil")
	assert.Equal(t, "git", gitConfig2.GetSourceType())
	assert.Equal(t, "https://github.com/org/infrastructure", gitConfig2.(*sync.GitSourceConfig).URL)
	assert.Equal(t, "main", gitConfig2.(*sync.GitSourceConfig).Branch)
	assert.Equal(t, "k8s", gitConfig2.(*sync.GitSourceConfig).BasePath)
	assert.Equal(t, 30*time.Minute, gitConfig2.GetInterval())
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Set environment variable to non-existent file
	err := os.Setenv("ARGUS_CONFIG_PATH", "/non/existent/config.yaml")
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_CONFIG_PATH"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Load config should succeed with defaults
	cfg, err := LoadConfig()
	require.NoError(t, err)

	// Verify default values are used
	assert.Equal(t, "localhost", cfg.Storage.Host)
	assert.Equal(t, 5432, cfg.Storage.Port)
	assert.Equal(t, "postgres", cfg.Storage.User)
	assert.Equal(t, "postgres", cfg.Storage.Password)
	assert.Equal(t, "argus", cfg.Storage.DBName)
	assert.Equal(t, "disable", cfg.Storage.SSLMode)
	assert.Len(t, cfg.Sync.Sources, 0)
}

func TestLoadConfig_InvalidConfigFile(t *testing.T) {
	// Create an invalid config file
	invalidConfig := `invalid: yaml: content`
	err := os.WriteFile("test-invalid-config.yaml", []byte(invalidConfig), 0644)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove("test-invalid-config.yaml"); err != nil {
			t.Logf("Failed to remove test file: %v", err)
		}
	}()

	// Set environment variable to invalid config file
	err = os.Setenv("ARGUS_CONFIG_PATH", "test-invalid-config.yaml")
	require.NoError(t, err)
	defer func() {
		if err := os.Unsetenv("ARGUS_CONFIG_PATH"); err != nil {
			t.Logf("Failed to unset environment variable: %v", err)
		}
	}()

	// Load config should fail
	_, err = LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestSourceConfig_UnmarshalYAML_InvalidType(t *testing.T) {
	// Test that the custom UnmarshalYAML method properly rejects invalid source types
	invalidYAML := `
- type: invalid-type
  path: "/test/path"
  interval: "30s"
`

	var sources []sync.SourceConfig
	err := yaml.Unmarshal([]byte(invalidYAML), &sources)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown source type: invalid-type")
}

func TestSourceConfig_UnmarshalYAML_ValidTypes(t *testing.T) {
	// Test that the custom UnmarshalYAML method accepts valid source types
	validYAML := `
- type: filesystem
  path: "/test/path"
  interval: "30s"
- type: git
  url: "https://github.com/test/repo"
  branch: "main"
  interval: "5m"
`

	var sources []sync.SourceConfig
	err := yaml.Unmarshal([]byte(validYAML), &sources)
	assert.NoError(t, err)
	assert.Len(t, sources, 2)

	// Check filesystem source
	fsConfig := sources[0].GetConfig()
	assert.NotNil(t, fsConfig)
	assert.Equal(t, "filesystem", fsConfig.GetSourceType())
	assert.Equal(t, "/test/path", fsConfig.(*sync.FilesystemSourceConfig).Path)
	assert.Equal(t, 30*time.Second, fsConfig.GetInterval())

	// Check git source
	gitConfig := sources[1].GetConfig()
	assert.NotNil(t, gitConfig)
	assert.Equal(t, "git", gitConfig.GetSourceType())
	assert.Equal(t, "https://github.com/test/repo", gitConfig.(*sync.GitSourceConfig).URL)
	assert.Equal(t, "main", gitConfig.(*sync.GitSourceConfig).Branch)
	assert.Equal(t, 5*time.Minute, gitConfig.GetInterval())
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
