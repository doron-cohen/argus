package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/doron-cohen/argus/backend/internal/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitSourceConfig holds git-specific configuration
type GitSourceConfig struct {
	Type     string        `yaml:"type"`
	Interval time.Duration `yaml:"interval"`
	URL      string        `yaml:"url"`
	Branch   string        `yaml:"branch,omitempty"`
	BasePath string        `yaml:"base_path,omitempty"`
}

// Validate ensures the git configuration is valid
func (g *GitSourceConfig) Validate() error {
	if g.Type != sourceTypeGit {
		return fmt.Errorf("expected type '%s', got '%s'", sourceTypeGit, g.Type)
	}
	if g.URL == "" {
		return fmt.Errorf("git source requires url field")
	}

	interval := g.GetInterval()
	if interval < MinGitInterval {
		return fmt.Errorf("git source interval must be at least %v, got %v", MinGitInterval, interval)
	}

	// Set default values if not provided
	if g.Type == "" {
		g.Type = sourceTypeGit
	}
	if g.Branch == "" {
		g.Branch = "main"
	}

	return nil
}

// GetInterval returns the sync interval for this source
func (g *GitSourceConfig) GetInterval() time.Duration {
	if g.Interval == 0 {
		return 5 * time.Minute // default
	}
	return g.Interval
}

// GetBasePath returns the base path for this source
func (g *GitSourceConfig) GetBasePath() string {
	return g.BasePath
}

// GetSourceType returns the source type
func (g *GitSourceConfig) GetSourceType() string {
	return sourceTypeGit
}

// GitFetcher implements ComponentsFetcher for git repositories
type GitFetcher struct {
	tempDir string
}

// NewGitFetcher creates a new git fetcher
func NewGitFetcher() *GitFetcher {
	return &GitFetcher{
		tempDir: os.TempDir(),
	}
}

// Fetch retrieves all components from a git repository
func (g *GitFetcher) Fetch(ctx context.Context, source SourceConfig) ([]models.Component, error) {
	cfg := source.GetConfig()
	gitConfig, ok := cfg.(*GitSourceConfig)
	if !ok {
		return nil, fmt.Errorf("source is not a git config")
	}

	// Get repository directory
	repoDir, err := g.ensureRepository(ctx, *gitConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Determine search directory based on base path
	searchDir := repoDir
	if gitConfig.BasePath != "" {
		searchDir = filepath.Join(repoDir, gitConfig.BasePath)
		// Check if base path exists
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("base path %s does not exist in repository", gitConfig.BasePath)
		}
	}

	// Load all manifests directly
	manifests, err := LoadManifests(ctx, searchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifests: %w", err)
	}

	slog.Debug("Found manifest files", "count", len(manifests), "source", gitConfig.URL)

	var components []models.Component
	for _, manifest := range manifests {
		component := manifest.Content.ToComponent()
		components = append(components, component)
	}

	return components, nil
}

// ensureRepository clones or updates the repository and returns the local path
func (g *GitFetcher) ensureRepository(ctx context.Context, gitConfig GitSourceConfig) (string, error) {
	// Create a safe directory name from the URL
	dirName := g.sanitizeURL(gitConfig.URL)
	repoDir := filepath.Join(g.tempDir, "argus-sync", dirName)

	// Check if directory exists and has a .git folder
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone the repository
		if err := g.cloneRepository(ctx, gitConfig, repoDir); err != nil {
			return "", err
		}
	} else {
		// Update existing repository
		if err := g.updateRepository(ctx, gitConfig, repoDir); err != nil {
			return "", err
		}
	}

	return repoDir, nil
}

// cloneRepository clones the repository using go-git with optional sparse checkout
func (g *GitFetcher) cloneRepository(ctx context.Context, gitConfig GitSourceConfig, repoDir string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(repoDir), 0750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone options
	cloneOptions := &git.CloneOptions{
		URL:           gitConfig.URL,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", gitConfig.Branch)),
		SingleBranch:  true,
		Depth:         1,
	}

	// Clone the repository
	repo, err := git.PlainClone(repoDir, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", gitConfig.URL, err)
	}

	// Set up sparse checkout if BasePath is specified
	if gitConfig.BasePath != "" {
		if err := g.setupSparseCheckout(repo, gitConfig.BasePath); err != nil {
			return fmt.Errorf("failed to setup sparse checkout: %w", err)
		}
	}

	return nil
}

// updateRepository pulls the latest changes using go-git
func (g *GitFetcher) updateRepository(ctx context.Context, gitConfig GitSourceConfig, repoDir string) error {
	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Fetch options
	fetchOptions := &git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", gitConfig.Branch, gitConfig.Branch)),
		},
	}

	// Fetch latest changes
	err = repo.Fetch(fetchOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch from repository: %w", err)
	}

	// Get the latest commit from the remote branch
	remoteRef, err := repo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", gitConfig.Branch)), true)
	if err != nil {
		return fmt.Errorf("failed to get remote reference: %w", err)
	}

	// Reset to the latest commit
	resetOptions := &git.ResetOptions{
		Commit: remoteRef.Hash(),
		Mode:   git.HardReset,
	}

	err = worktree.Reset(resetOptions)
	if err != nil {
		return fmt.Errorf("failed to reset repository: %w", err)
	}

	// Ensure sparse checkout is still configured if BasePath is specified
	if gitConfig.BasePath != "" {
		if err := g.setupSparseCheckout(repo, gitConfig.BasePath); err != nil {
			return fmt.Errorf("failed to maintain sparse checkout: %w", err)
		}
	}

	return nil
}

// setupSparseCheckout configures sparse checkout for the specified base path
func (g *GitFetcher) setupSparseCheckout(repo *git.Repository, basePath string) error {
	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get repository root
	repoRoot := worktree.Filesystem.Root()

	// Create .git/info/sparse-checkout file
	sparseCheckoutPath := filepath.Join(repoRoot, ".git", "info", "sparse-checkout")

	// Ensure the info directory exists
	if err := os.MkdirAll(filepath.Dir(sparseCheckoutPath), 0750); err != nil {
		return fmt.Errorf("failed to create sparse-checkout directory: %w", err)
	}

	// Write sparse checkout configuration
	// Format: the base path and everything under it
	sparseContent := fmt.Sprintf("%s/*\n", strings.TrimPrefix(basePath, "/"))
	if err := os.WriteFile(sparseCheckoutPath, []byte(sparseContent), 0600); err != nil {
		return fmt.Errorf("failed to write sparse-checkout file: %w", err)
	}

	// Configure git to use sparse checkout
	gitConfigPath := filepath.Join(repoRoot, ".git", "config")

	// Read existing config
	cleanConfigPath := filepath.Clean(gitConfigPath)
	configContent, err := os.ReadFile(cleanConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}

	// Add sparse checkout configuration if not present
	configStr := string(configContent)
	if !strings.Contains(configStr, "core.sparseCheckout") {
		configStr += "\n[core]\n\tsparseCheckout = true\n"
		if err := os.WriteFile(cleanConfigPath, []byte(configStr), 0600); err != nil {
			return fmt.Errorf("failed to update git config: %w", err)
		}
	}

	// Apply sparse checkout by re-reading the index
	// This will remove files not matching the sparse checkout pattern
	// Some errors are expected with sparse checkout, ignore them
	_, _ = worktree.Add(".")

	return nil
}

// sanitizeURL creates a safe directory name from a URL
func (g *GitFetcher) sanitizeURL(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "git@")

	// Replace special characters with underscores
	url = strings.ReplaceAll(url, "/", "_")
	url = strings.ReplaceAll(url, ":", "_")
	url = strings.ReplaceAll(url, ".", "_")

	return url
}
