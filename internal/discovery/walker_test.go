package discovery

import (
	"context"
	"os"
	"path/filepath"
	"project-dash/internal/fs"
	"testing"
)

func TestWalker_Walk_RegularRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory structure with a Git repository
	repoPath := filepath.Join(tmpDir, "myproject", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 0)

	ctx := context.Background()
	foundRepo := false
	var foundPath string

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		switch event {
		case EventFoundRepo:
			foundRepo = true
			foundPath = path
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if !foundRepo {
		t.Error("expected to find repository, but found none")
	}

	expectedPath := filepath.Join(tmpDir, "myproject")
	if foundPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, foundPath)
	}
}

func TestWalker_Walk_IgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure with node_modules containing a .git directory
	nodeModulesRepo := filepath.Join(tmpDir, "project", "node_modules", "dependency", ".git")
	if err := os.MkdirAll(nodeModulesRepo, 0755); err != nil {
		t.Fatalf("failed to create node_modules repo: %v", err)
	}

	// Create another repo outside node_modules
	regularRepo := filepath.Join(tmpDir, "project", ".git")
	if err := os.MkdirAll(regularRepo, 0755); err != nil {
		t.Fatalf("failed to create regular repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{"node_modules"}, 0)

	ctx := context.Background()
	foundRepos := []string{}

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		if event == EventFoundRepo {
			foundRepos = append(foundRepos, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should only find the regular repo, not the one in node_modules
	if len(foundRepos) != 1 {
		t.Errorf("expected 1 repository, found %d", len(foundRepos))
	}

	expectedPath := filepath.Join(tmpDir, "project")
	if len(foundRepos) == 1 && foundRepos[0] != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, foundRepos[0])
	}
}

func TestWalker_Walk_MaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deep directory structure
	deepRepo := filepath.Join(tmpDir, "a", "b", "c", "d", "e", ".git")
	if err := os.MkdirAll(deepRepo, 0755); err != nil {
		t.Fatalf("failed to create deep repo: %v", err)
	}

	// Create a shallow repo
	shallowRepo := filepath.Join(tmpDir, "shallow", ".git")
	if err := os.MkdirAll(shallowRepo, 0755); err != nil {
		t.Fatalf("failed to create shallow repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 2) // max depth 2

	ctx := context.Background()
	foundRepos := []string{}

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		if event == EventFoundRepo {
			foundRepos = append(foundRepos, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should only find the shallow repo
	if len(foundRepos) != 1 {
		t.Errorf("expected 1 repository with max depth 2, found %d", len(foundRepos))
	}

	if len(foundRepos) == 1 && filepath.Base(foundRepos[0]) != "shallow" {
		t.Errorf("expected to find shallow repo, found %s", foundRepos[0])
	}
}

func TestWalker_Walk_PermissionError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping permission test in short mode")
	}

	tmpDir := t.TempDir()

	// Create a directory structure with a repo
	repoPath := filepath.Join(tmpDir, "repo", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// Create a subdirectory with no permissions
	noPermDir := filepath.Join(tmpDir, "noperm")
	if err := os.Mkdir(noPermDir, 0000); err != nil {
		t.Fatalf("failed to create no-permission directory: %v", err)
	}
	defer os.Chmod(noPermDir, 0755) // Cleanup permissions

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 0)

	ctx := context.Background()
	foundRepo := false
	hadError := false

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		switch event {
		case EventFoundRepo:
			foundRepo = true
		case EventError:
			if err != nil {
				hadError = true
			}
		}
		return nil
	})

	// Walk should complete successfully despite permission error
	if err != nil {
		t.Fatalf("Walk should not fail: %v", err)
	}

	// Should still find the repository
	if !foundRepo {
		t.Error("expected to find repository despite permission errors")
	}

	// Should have encountered permission error
	if !hadError {
		t.Error("expected to encounter permission error")
	}
}

func TestWalker_Walk_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deep directory structure to give us time to cancel
	for i := 0; i < 100; i++ {
		deepPath := filepath.Join(tmpDir, "deep", "level"+string(rune('0'+i)))
		if err := os.MkdirAll(deepPath, 0755); err != nil {
			t.Fatalf("failed to create deep directory: %v", err)
		}
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 0)

	// Create a context that gets cancelled immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		return nil
	})

	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestWalker_Walk_MultipleRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple repositories in different locations
	repos := []string{"project1", "project2", "nested/project3"}
	for _, repo := range repos {
		repoPath := filepath.Join(tmpDir, repo, ".git")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo %s: %v", repo, err)
		}
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 0)

	ctx := context.Background()
	foundRepos := []string{}

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		if event == EventFoundRepo {
			foundRepos = append(foundRepos, filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find all 3 repositories
	if len(foundRepos) != 3 {
		t.Errorf("expected 3 repositories, found %d", len(foundRepos))
	}

	// Check that we found the expected repos
	foundNames := make(map[string]bool)
	for _, name := range foundRepos {
		foundNames[name] = true
	}

	expectedNames := []string{"project1", "project2", "project3"}
	for _, name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("expected to find repository %s", name)
		}
	}
}

func TestWalker_Walk_EmptyRoot(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create any subdirectories

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 0)

	ctx := context.Background()
	enterDir := false

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		if event == EventEnterDir {
			enterDir = true
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should at least enter the root directory
	if !enterDir {
		t.Error("expected to enter root directory")
	}
}

func TestWalker_ResetInodeTracking(t *testing.T) {
	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	walker := NewWalker(filesystem, detector, []string{}, 0)

	// This tests the reset functionality
	walker.ResetInodeTracking()

	// We can't easily test inode tracking without creating symlinks,
	// but we can at least verify the method exists and doesn't panic
	if walker == nil {
		t.Error("walker should not be nil after reset")
	}
}
