package discovery

import (
	"os"
	"path/filepath"
	"strings"
)

// GitDetector detects Git repositories by checking for .git directory or file
type Detector struct {
	fs DetectorFS
}

// DetectorFS interface for detector operations
type DetectorFS interface {
	Lstat(path string) (os.FileInfo, error)
	Stat(path string) (os.FileInfo, error)
}

// NewDetector creates a new GitDetector
func NewDetector(fs DetectorFS) *Detector {
	return &Detector{fs: fs}
}

// IsGitRepository checks if a directory is a Git repository
// It returns true if:
// - The directory itself contains a .git subdirectory (regular repo)
// - The directory itself contains a .git file (worktree)
// - The directory IS the .git directory (bare repo)
func (d *Detector) IsGitRepository(dirPath string) bool {
	// Check for .git subdirectory (regular repo)
	gitDir := filepath.Join(dirPath, ".git")
	if info, err := d.fs.Lstat(gitDir); err == nil {
		if info.IsDir() {
			return true // Regular repository
		}
		// Check if it's a .git file (worktree)
		// A regular file has no mode bits set for special file types
		if !info.IsDir() && info.Mode()&os.ModeType == 0 {
			return d.isWorktreeGitFile(gitDir)
		}
	}

	// Check if this is a bare repository (directory name ends with .git)
	if strings.HasSuffix(filepath.Base(dirPath), ".git") {
		// A bare repo has specific structure: HEAD, objects, refs directories
		if d.hasBareRepoStructure(dirPath) {
			return true
		}
	}

	return false
}

// isWorktreeGitFile checks if a .git file is a Git worktree file
// Worktree files contain: "gitdir: <path>"
func (d *Detector) isWorktreeGitFile(gitFilePath string) bool {
	// Read the file content to verify it's a Git worktree file
	content, err := os.ReadFile(gitFilePath)
	if err != nil {
		return false
	}

	// Check if it starts with "gitdir:"
	return strings.HasPrefix(strings.TrimSpace(string(content)), "gitdir:")
}

// hasBareRepoStructure checks if a directory has the structure of a bare Git repository
func (d *Detector) hasBareRepoStructure(dirPath string) bool {
	// Check for HEAD file or directory
	headPath := filepath.Join(dirPath, "HEAD")
	if _, err := d.fs.Stat(headPath); err != nil {
		return false
	}

	// Check for objects directory
	objectsPath := filepath.Join(dirPath, "objects")
	if info, err := d.fs.Stat(objectsPath); err != nil || !info.IsDir() {
		return false
	}

	// Check for refs directory
	refsPath := filepath.Join(dirPath, "refs")
	if info, err := d.fs.Stat(refsPath); err != nil || !info.IsDir() {
		return false
	}

	return true
}

// GetRepositoryType determines the repository type
func (d *Detector) GetRepositoryType(dirPath string) string {
	// Check for .git subdirectory
	gitDir := filepath.Join(dirPath, ".git")
	if info, err := d.fs.Lstat(gitDir); err == nil {
		if info.IsDir() {
			return "regular"
		}
		// Check if it's a worktree .git file
		// A regular file has no mode bits set for special file types
		if !info.IsDir() && info.Mode()&os.ModeType == 0 {
			if d.isWorktreeGitFile(gitDir) {
				return "worktree"
			}
		}
	}

	// Check for bare repo
	if strings.HasSuffix(filepath.Base(dirPath), ".git") {
		if d.hasBareRepoStructure(dirPath) {
			return "bare"
		}
	}

	return "unknown"
}
