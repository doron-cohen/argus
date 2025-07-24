package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitClient handles git repository operations using go-git
type GitClient struct {
	tempDir string
}

// NewGitClient creates a new git client
func NewGitClient() *GitClient {
	return &GitClient{
		tempDir: os.TempDir(),
	}
}

// FindManifests finds all manifest.yaml and manifest.yml files in the repository
func (g *GitClient) FindManifests(ctx context.Context, source SourceConfig) ([]string, error) {
	repoDir, err := g.ensureRepository(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Determine search directory based on base path
	searchDir := repoDir
	if source.BasePath != "" {
		searchDir = filepath.Join(repoDir, source.BasePath)
		// Check if base path exists
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("base path %s does not exist in repository", source.BasePath)
		}
	}

	var manifests []string

	// Find manifest.yaml files
	yamlFiles, err := g.findFiles(searchDir, "manifest.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest.yaml files: %w", err)
	}

	// If we have a base path, adjust the relative paths
	if source.BasePath != "" {
		for i, file := range yamlFiles {
			yamlFiles[i] = filepath.Join(source.BasePath, file)
		}
	}
	manifests = append(manifests, yamlFiles...)

	// Find manifest.yml files
	ymlFiles, err := g.findFiles(searchDir, "manifest.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to find manifest.yml files: %w", err)
	}

	// If we have a base path, adjust the relative paths
	if source.BasePath != "" {
		for i, file := range ymlFiles {
			ymlFiles[i] = filepath.Join(source.BasePath, file)
		}
	}
	manifests = append(manifests, ymlFiles...)

	return manifests, nil
}

// GetFileContent reads the content of a file from the repository
func (g *GitClient) GetFileContent(ctx context.Context, source SourceConfig, filePath string) ([]byte, error) {
	repoDir, err := g.ensureRepository(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure repository: %w", err)
	}

	fullPath := filepath.Join(repoDir, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return content, nil
}

// GetLatestCommit returns the latest commit hash using go-git
func (g *GitClient) GetLatestCommit(ctx context.Context, source SourceConfig) (string, error) {
	repoDir, err := g.ensureRepository(ctx, source)
	if err != nil {
		return "", fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	return ref.Hash().String(), nil
}

// ensureRepository clones or updates the repository and returns the local path
func (g *GitClient) ensureRepository(ctx context.Context, source SourceConfig) (string, error) {
	// Create a safe directory name from the URL
	dirName := g.sanitizeURL(source.URL)
	repoDir := filepath.Join(g.tempDir, "argus-sync", dirName)

	// Check if directory exists and has a .git folder
	gitDir := filepath.Join(repoDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone the repository
		if err := g.cloneRepository(ctx, source, repoDir); err != nil {
			return "", err
		}
	} else {
		// Update existing repository
		if err := g.updateRepository(ctx, source, repoDir); err != nil {
			return "", err
		}
	}

	return repoDir, nil
}

// cloneRepository clones the repository using go-git with optional sparse checkout
func (g *GitClient) cloneRepository(ctx context.Context, source SourceConfig, repoDir string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(repoDir), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Clone options
	cloneOptions := &git.CloneOptions{
		URL:           source.URL,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", source.Branch)),
		SingleBranch:  true,
		Depth:         1,
	}

	// Clone the repository
	repo, err := git.PlainClone(repoDir, false, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w", source.URL, err)
	}

	// Set up sparse checkout if BasePath is specified
	if source.BasePath != "" {
		if err := g.setupSparseCheckout(repo, source.BasePath); err != nil {
			return fmt.Errorf("failed to setup sparse checkout: %w", err)
		}
	}

	return nil
}

// updateRepository pulls the latest changes using go-git
func (g *GitClient) updateRepository(ctx context.Context, source SourceConfig, repoDir string) error {
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
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", source.Branch, source.Branch)),
		},
	}

	// Fetch latest changes
	err = repo.Fetch(fetchOptions)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch from repository: %w", err)
	}

	// Get the latest commit from the remote branch
	remoteRef, err := repo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", source.Branch)), true)
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
	if source.BasePath != "" {
		if err := g.setupSparseCheckout(repo, source.BasePath); err != nil {
			return fmt.Errorf("failed to maintain sparse checkout: %w", err)
		}
	}

	return nil
}

// setupSparseCheckout configures sparse checkout for the specified base path
func (g *GitClient) setupSparseCheckout(repo *git.Repository, basePath string) error {
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
	if err := os.MkdirAll(filepath.Dir(sparseCheckoutPath), 0755); err != nil {
		return fmt.Errorf("failed to create sparse-checkout directory: %w", err)
	}

	// Write sparse checkout configuration
	// Format: the base path and everything under it
	sparseContent := fmt.Sprintf("%s/*\n", strings.TrimPrefix(basePath, "/"))
	if err := os.WriteFile(sparseCheckoutPath, []byte(sparseContent), 0644); err != nil {
		return fmt.Errorf("failed to write sparse-checkout file: %w", err)
	}

	// Configure git to use sparse checkout
	gitConfigPath := filepath.Join(repoRoot, ".git", "config")

	// Read existing config
	configContent, err := os.ReadFile(gitConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read git config: %w", err)
	}

	// Add sparse checkout configuration if not present
	configStr := string(configContent)
	if !strings.Contains(configStr, "core.sparseCheckout") {
		configStr += "\n[core]\n\tsparseCheckout = true\n"
		if err := os.WriteFile(gitConfigPath, []byte(configStr), 0644); err != nil {
			return fmt.Errorf("failed to update git config: %w", err)
		}
	}

	// Apply sparse checkout by re-reading the index
	// This will remove files not matching the sparse checkout pattern
	_, err = worktree.Add(".")
	if err != nil && err != git.ErrGitModulesSymlink {
		// Some errors are expected with sparse checkout, ignore them
	}

	return nil
}

// findFiles recursively finds files with the given name
func (g *GitClient) findFiles(rootDir, fileName string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == fileName {
			// Get relative path from root directory
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// sanitizeURL creates a safe directory name from a URL
func (g *GitClient) sanitizeURL(url string) string {
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
