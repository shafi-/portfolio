package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/internal/fs"
)

func TestDetector_IsGitRepository_RegularRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with a .git subdirectory
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Use OS filesystem
	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	isRepo := detector.IsGitRepository(repoDir)
	if !isRepo {
		t.Error("expected directory with .git subdirectory to be detected as repository")
	}
}

func TestDetector_IsGitRepository_Worktree(t *testing.T) {
	// Create a mock filesystem with a .git file (worktree)
	// We need to create actual files for this test since it reads file content
	tmpDir := t.TempDir()

	worktreeDir := filepath.Join(tmpDir, "worktree")
	if err := os.Mkdir(worktreeDir, 0755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	// Create a .git file with worktree content
	gitFile := filepath.Join(worktreeDir, ".git")
	gitContent := "gitdir: /path/to/main/repo/.git/worktrees/worktree"
	if err := os.WriteFile(gitFile, []byte(gitContent), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	// Test with real filesystem
	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	isRepo := detector.IsGitRepository(worktreeDir)
	if !isRepo {
		t.Error("expected directory with .git worktree file to be detected as repository")
	}
}

func TestDetector_IsGitRepository_BareRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a bare repository structure
	bareRepo := filepath.Join(tmpDir, "repo.git")
	if err := os.MkdirAll(filepath.Join(bareRepo, "objects"), 0755); err != nil {
		t.Fatalf("failed to create objects directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(bareRepo, "refs"), 0755); err != nil {
		t.Fatalf("failed to create refs directory: %v", err)
	}

	// Create HEAD file
	headContent := "ref: refs/heads/main\n"
	if err := os.WriteFile(filepath.Join(bareRepo, "HEAD"), []byte(headContent), 0644); err != nil {
		t.Fatalf("failed to create HEAD file: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	isRepo := detector.IsGitRepository(bareRepo)
	if !isRepo {
		t.Error("expected bare repository to be detected")
	}
}

func TestDetector_IsGitRepository_NotARepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory without any Git indicators
	regularDir := filepath.Join(tmpDir, "regular")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatalf("failed to create regular directory: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	isRepo := detector.IsGitRepository(regularDir)
	if isRepo {
		t.Error("expected regular directory to not be detected as repository")
	}
}

func TestDetector_GetRepositoryType_Regular(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a regular repository
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	repoType := detector.GetRepositoryType(repoDir)
	if repoType != "regular" {
		t.Errorf("expected repository type 'regular', got '%s'", repoType)
	}
}

func TestDetector_GetRepositoryType_Worktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a worktree
	worktreeDir := filepath.Join(tmpDir, "worktree")
	if err := os.Mkdir(worktreeDir, 0755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	gitFile := filepath.Join(worktreeDir, ".git")
	gitContent := "gitdir: /path/to/main/repo/.git/worktrees/worktree"
	if err := os.WriteFile(gitFile, []byte(gitContent), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	repoType := detector.GetRepositoryType(worktreeDir)
	if repoType != "worktree" {
		t.Errorf("expected repository type 'worktree', got '%s'", repoType)
	}
}

func TestDetector_GetRepositoryType_Bare(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a bare repository
	bareRepo := filepath.Join(tmpDir, "repo.git")
	if err := os.MkdirAll(filepath.Join(bareRepo, "objects"), 0755); err != nil {
		t.Fatalf("failed to create objects directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(bareRepo, "refs"), 0755); err != nil {
		t.Fatalf("failed to create refs directory: %v", err)
	}

	headContent := "ref: refs/heads/main\n"
	if err := os.WriteFile(filepath.Join(bareRepo, "HEAD"), []byte(headContent), 0644); err != nil {
		t.Fatalf("failed to create HEAD file: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	repoType := detector.GetRepositoryType(bareRepo)
	if repoType != "bare" {
		t.Errorf("expected repository type 'bare', got '%s'", repoType)
	}
}

func TestDetector_GetRepositoryType_Unknown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory that's not a Git repository
	regularDir := filepath.Join(tmpDir, "regular")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatalf("failed to create regular directory: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	repoType := detector.GetRepositoryType(regularDir)
	if repoType != "unknown" {
		t.Errorf("expected repository type 'unknown', got '%s'", repoType)
	}
}

func TestDetector_IsWorktreeGitFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid worktree file
	gitFile := filepath.Join(tmpDir, ".git")
	gitContent := "gitdir: /some/path"
	if err := os.WriteFile(gitFile, []byte(gitContent), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	// This tests the internal method through IsGitRepository
	worktreeDir := filepath.Join(tmpDir, "worktree")
	if err := os.Mkdir(worktreeDir, 0755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	// Move the .git file to worktree directory
	worktreeGitFile := filepath.Join(worktreeDir, ".git")
	if err := os.WriteFile(worktreeGitFile, []byte(gitContent), 0644); err != nil {
		t.Fatalf("failed to create worktree .git file: %v", err)
	}

	isRepo := detector.IsGitRepository(worktreeDir)
	if !isRepo {
		t.Error("expected worktree with valid .git file to be detected as repository")
	}
}

func TestDetector_IsWorktreeGitFile_Invalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with a .git file that's not a worktree file
	worktreeDir := filepath.Join(tmpDir, "worktree")
	if err := os.Mkdir(worktreeDir, 0755); err != nil {
		t.Fatalf("failed to create worktree directory: %v", err)
	}

	gitFile := filepath.Join(worktreeDir, ".git")
	gitContent := "not a worktree file"
	if err := os.WriteFile(gitFile, []byte(gitContent), 0644); err != nil {
		t.Fatalf("failed to create .git file: %v", err)
	}

	fs := fs.NewOSFilesystem()
	detector := NewDetector(fs)

	isRepo := detector.IsGitRepository(worktreeDir)
	if isRepo {
		t.Error("expected directory with invalid .git file to not be detected as repository")
	}
}
