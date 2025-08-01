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

func TestLoadConfig_FromEnvironment(t *testing.T) {
	// Copy testdata file to current directory
	srcFile := "testdata/environment.yaml"
	dstFile := "test-config.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer os.Remove(dstFile)

	// Set environment variable
	os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	defer os.Unsetenv("ARGUS_CONFIG_PATH")

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
	if fsConfig != nil {
		assert.Equal(t, "filesystem", fsConfig.GetSourceType())
		assert.Equal(t, "/test/path", fsConfig.(*sync.FilesystemSourceConfig).Path)
		assert.Equal(t, 30*time.Second, fsConfig.GetInterval())
	}

	// Check git source
	gitSource := cfg.Sync.Sources[1]
	gitConfig := gitSource.GetConfig()
	if gitConfig != nil {
		assert.Equal(t, "git", gitConfig.GetSourceType())
		assert.Equal(t, "https://github.com/test/repo", gitConfig.(*sync.GitSourceConfig).URL)
		assert.Equal(t, "main", gitConfig.(*sync.GitSourceConfig).Branch)
		assert.Equal(t, 5*time.Minute, gitConfig.GetInterval())
	}
}

func TestLoadConfig_DefaultBehavior(t *testing.T) {
	// Copy testdata file to current directory as config.yaml
	srcFile := "testdata/default.yaml"
	dstFile := "config.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer os.Remove(dstFile)

	// Clear environment variable to test default behavior
	os.Unsetenv("ARGUS_CONFIG_PATH")

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
	if fsConfig != nil {
		assert.Equal(t, "filesystem", fsConfig.GetSourceType())
		assert.Equal(t, "./local/path", fsConfig.(*sync.FilesystemSourceConfig).Path)
		assert.Equal(t, 1*time.Minute, fsConfig.GetInterval())
	}
}

func TestLoadConfig_EmptySyncSources(t *testing.T) {
	// Copy testdata file to current directory
	srcFile := "testdata/empty-sync.yaml"
	dstFile := "test-config-empty.yaml"

	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)
	defer os.Remove(dstFile)

	os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	defer os.Unsetenv("ARGUS_CONFIG_PATH")

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
	defer os.Remove(dstFile)

	os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	defer os.Unsetenv("ARGUS_CONFIG_PATH")

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
	defer os.Remove(dstFile)

	os.Setenv("ARGUS_CONFIG_PATH", dstFile)
	defer os.Unsetenv("ARGUS_CONFIG_PATH")

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
	if fsConfig != nil {
		assert.Equal(t, "filesystem", fsConfig.GetSourceType())
		assert.Equal(t, "/path/with/spaces", fsConfig.(*sync.FilesystemSourceConfig).Path)
		assert.Equal(t, 2*time.Minute, fsConfig.GetInterval())
	}

	// Check first git source
	gitSource1 := cfg.Sync.Sources[1]
	gitConfig1 := gitSource1.GetConfig()
	if gitConfig1 != nil {
		assert.Equal(t, "git", gitConfig1.GetSourceType())
		assert.Equal(t, "https://github.com/org/monorepo", gitConfig1.(*sync.GitSourceConfig).URL)
		assert.Equal(t, "develop", gitConfig1.(*sync.GitSourceConfig).Branch)
		assert.Equal(t, "services/backend", gitConfig1.(*sync.GitSourceConfig).BasePath)
		assert.Equal(t, 10*time.Minute, gitConfig1.GetInterval())
	}

	// Check second git source
	gitSource2 := cfg.Sync.Sources[2]
	gitConfig2 := gitSource2.GetConfig()
	if gitConfig2 != nil {
		assert.Equal(t, "git", gitConfig2.GetSourceType())
		assert.Equal(t, "https://github.com/org/infrastructure", gitConfig2.(*sync.GitSourceConfig).URL)
		assert.Equal(t, "main", gitConfig2.(*sync.GitSourceConfig).Branch)
		assert.Equal(t, "k8s", gitConfig2.(*sync.GitSourceConfig).BasePath)
		assert.Equal(t, 30*time.Minute, gitConfig2.GetInterval())
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	// Set environment variable to non-existent file
	os.Setenv("ARGUS_CONFIG_PATH", "/non/existent/config.yaml")
	defer os.Unsetenv("ARGUS_CONFIG_PATH")

	// Load config should fail
	_, err := LoadConfig()
	assert.Error(t, err)
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
