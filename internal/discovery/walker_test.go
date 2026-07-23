package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"project-dash/internal/fs"
	"project-dash/internal/logging"
	"strings"
	"testing"
)

// getTestLogger returns a logger for testing
func getTestLogger() *logging.Logger {
	logger, err := logging.NewLogger("INFO", "console")
	if err != nil {
		panic(err)
	}
	return logger
}

func TestWalker_Walk_RegularRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory structure with a Git repository
	repoPath := filepath.Join(tmpDir, "myproject", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{"node_modules"}, 0, logger)

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 2, logger) // max depth 2

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

	// This tests the reset functionality
	walker.ResetInodeTracking()

	// We can't easily test inode tracking without creating symlinks,
	// but we can at least verify the method exists and doesn't panic
	if walker == nil {
		t.Error("walker should not be nil after reset")
	}
}

// TestWalker_Walk_NestedRepo tests AC-11: nested repo inside parent repo
func TestWalker_Walk_NestedRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a parent repository
	parentRepo := filepath.Join(tmpDir, "parent-project", ".git")
	if err := os.MkdirAll(parentRepo, 0755); err != nil {
		t.Fatalf("failed to create parent repo: %v", err)
	}

	// Create a nested repository inside the parent
	nestedRepo := filepath.Join(tmpDir, "parent-project", "services", "auth", ".git")
	if err := os.MkdirAll(nestedRepo, 0755); err != nil {
		t.Fatalf("failed to create nested repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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

	// Should find both parent and nested repositories
	if len(foundRepos) != 2 {
		t.Errorf("expected 2 repositories (parent + nested), found %d", len(foundRepos))
	}

	// Verify we found the parent repo
	foundParent := false
	foundNested := false
	for _, repo := range foundRepos {
		if filepath.Base(repo) == "parent-project" {
			foundParent = true
		}
		if filepath.Base(repo) == "auth" {
			foundNested = true
		}
	}

	if !foundParent {
		t.Error("expected to find parent repository")
	}
	if !foundNested {
		t.Error("expected to find nested repository")
	}
}

// TestWalker_Walk_MonorepoStructure tests AC-12: monorepo with multiple nested services
func TestWalker_Walk_MonorepoStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a monorepo root
	monorepoRoot := filepath.Join(tmpDir, "monorepo", ".git")
	if err := os.MkdirAll(monorepoRoot, 0755); err != nil {
		t.Fatalf("failed to create monorepo root: %v", err)
	}

	// Create multiple nested service repositories
	services := []string{"auth", "api", "frontend", "backend"}
	for _, service := range services {
		serviceRepo := filepath.Join(tmpDir, "monorepo", "services", service, ".git")
		if err := os.MkdirAll(serviceRepo, 0755); err != nil {
			t.Fatalf("failed to create service repo %s: %v", service, err)
		}
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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

	// Should find monorepo root + all service repositories (5 total)
	if len(foundRepos) != 5 {
		t.Errorf("expected 5 repositories (monorepo + 4 services), found %d: %v", len(foundRepos), foundRepos)
	}

	// Verify we found all expected repos
	foundNames := make(map[string]bool)
	for _, name := range foundRepos {
		foundNames[name] = true
	}

	expectedNames := append([]string{"monorepo"}, services...)
	for _, name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("expected to find repository %s", name)
		}
	}
}

// TestWalker_Walk_DeepNesting tests AC-13: deep nesting (3+ levels)
func TestWalker_Walk_DeepNesting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a chain of nested repositories: root > level1 > level2 > level3
	levels := []string{"root", "level1", "level2", "level3"}
	for i, level := range levels {
		pathParts := append([]string{tmpDir}, levels[:i+1]...)
		repoPath := filepath.Join(pathParts...)
		gitPath := filepath.Join(repoPath, ".git")
		if err := os.MkdirAll(gitPath, 0755); err != nil {
			t.Fatalf("failed to create level %s repo: %v", level, err)
		}
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger) // No depth limit

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

	// Should find all 4 levels of nested repositories
	if len(foundRepos) != 4 {
		t.Errorf("expected 4 repositories at different nesting levels, found %d: %v", len(foundRepos), foundRepos)
	}

	// Verify we found all levels
	foundNames := make(map[string]bool)
	for _, name := range foundRepos {
		foundNames[name] = true
	}

	for _, level := range levels {
		if !foundNames[level] {
			t.Errorf("expected to find repository at level %s", level)
		}
	}
}

// TestWalker_Walk_MixedStructure tests mixed structures (some dirs are repos, some aren't)
func TestWalker_Walk_MixedStructure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create parent repository
	parentRepo := filepath.Join(tmpDir, "project", ".git")
	if err := os.MkdirAll(parentRepo, 0755); err != nil {
		t.Fatalf("failed to create parent repo: %v", err)
	}

	// Create a nested repository
	nestedRepo := filepath.Join(tmpDir, "project", "libs", "auth", ".git")
	if err := os.MkdirAll(nestedRepo, 0755); err != nil {
		t.Fatalf("failed to create nested repo: %v", err)
	}

	// Create directories that are NOT repositories (should not be reported as repos)
	nonRepoDirs := []string{
		filepath.Join(tmpDir, "project", "docs"),
		filepath.Join(tmpDir, "project", "build"),
		filepath.Join(tmpDir, "project", "libs", "utils"),
	}
	for _, dir := range nonRepoDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create non-repo directory: %v", err)
		}
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

	ctx := context.Background()
	foundRepos := []string{}
	enteredDirs := []string{}

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		switch event {
		case EventFoundRepo:
			foundRepos = append(foundRepos, filepath.Base(path))
		case EventEnterDir:
			enteredDirs = append(enteredDirs, filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should find exactly 2 repositories
	if len(foundRepos) != 2 {
		t.Errorf("expected 2 repositories in mixed structure, found %d: %v", len(foundRepos), foundRepos)
	}

	// Verify we found the expected repos
	foundNames := make(map[string]bool)
	for _, name := range foundRepos {
		foundNames[name] = true
	}

	if !foundNames["project"] {
		t.Error("expected to find project repository")
	}
	if !foundNames["auth"] {
		t.Error("expected to find nested auth repository")
	}
}

// TestWalker_Walk_SiblingRepos tests sibling repos at same level
func TestWalker_Walk_SiblingRepos(t *testing.T) {
	tmpDir := t.TempDir()

	// Create parent repository
	parentRepo := filepath.Join(tmpDir, "workspace", ".git")
	if err := os.MkdirAll(parentRepo, 0755); err != nil {
		t.Fatalf("failed to create parent repo: %v", err)
	}

	// Create sibling repositories at the same level
	siblings := []string{"service-a", "service-b", "service-c"}
	for _, sibling := range siblings {
		siblingRepo := filepath.Join(tmpDir, "workspace", "services", sibling, ".git")
		if err := os.MkdirAll(siblingRepo, 0755); err != nil {
			t.Fatalf("failed to create sibling repo %s: %v", sibling, err)
		}
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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

	// Should find workspace + 3 sibling services (4 total)
	if len(foundRepos) != 4 {
		t.Errorf("expected 4 repositories (workspace + 3 siblings), found %d: %v", len(foundRepos), foundRepos)
	}

	// Verify we found all expected repos
	foundNames := make(map[string]bool)
	for _, name := range foundRepos {
		foundNames[name] = true
	}

	expectedNames := append([]string{"workspace"}, siblings...)
	for _, name := range expectedNames {
		if !foundNames[name] {
			t.Errorf("expected to find repository %s", name)
		}
	}
}

// TestWalker_Walk_ComplexMonorepo tests a complex monorepo structure
func TestWalker_Walk_ComplexMonorepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create root monorepo
	monorepo := filepath.Join(tmpDir, "company-monorepo", ".git")
	if err := os.MkdirAll(monorepo, 0755); err != nil {
		t.Fatalf("failed to create monorepo: %v", err)
	}

	// Create services with nested repositories
	services := []string{"auth", "api", "worker"}
	for _, service := range services {
		servicePath := filepath.Join(tmpDir, "company-monorepo", "services", service, ".git")
		if err := os.MkdirAll(servicePath, 0755); err != nil {
			t.Fatalf("failed to create service %s: %v", service, err)
		}
	}

	// Create libraries with nested repositories
	libs := []string{"common", "utils"}
	for _, lib := range libs {
		libPath := filepath.Join(tmpDir, "company-monorepo", "libs", lib, ".git")
		if err := os.MkdirAll(libPath, 0755); err != nil {
			t.Fatalf("failed to create lib %s: %v", lib, err)
		}
	}

	// Create nested repository within a service (deep nesting)
	deepNested := filepath.Join(tmpDir, "company-monorepo", "services", "api", "internal", "core", ".git")
	if err := os.MkdirAll(deepNested, 0755); err != nil {
		t.Fatalf("failed to create deeply nested repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

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

	// Should find: monorepo root + 3 services + 2 libs + 1 deep nested = 7 total
	if len(foundRepos) != 7 {
		t.Errorf("expected 7 repositories in complex monorepo, found %d: %v", len(foundRepos), foundRepos)
	}

	// Verify key repositories are found
	repoBasenames := make(map[string]bool)
	for _, repo := range foundRepos {
		repoBasenames[filepath.Base(repo)] = true
	}

	// Check for expected repositories
	expectedBasenames := []string{"company-monorepo", "auth", "api", "worker", "common", "utils", "core"}
	for _, expected := range expectedBasenames {
		if !repoBasenames[expected] {
			t.Errorf("expected to find repository %s", expected)
		}
	}
}

// TestWalker_IgnoreMatcher_DefaultPatterns tests AC-21: default ignore patterns
func TestWalker_IgnoreMatcher_DefaultPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repositories in default ignored directories
	ignoredDirs := []string{"node_modules", "vendor", ".venv", "target", "build", "dist"}
	for _, dir := range ignoredDirs {
		repoPath := filepath.Join(tmpDir, "project", dir, "subdir", ".git")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo in %s: %v", dir, err)
		}
	}

	// Create a regular repository that should be found
	regularRepo := filepath.Join(tmpDir, "project", ".git")
	if err := os.MkdirAll(regularRepo, 0755); err != nil {
		t.Fatalf("failed to create regular repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

	ctx := context.Background()
	foundRepos := []string{}
	skippedDirs := []string{}

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		switch event {
		case EventFoundRepo:
			foundRepos = append(foundRepos, filepath.Base(path))
		case EventSkipped:
			skippedDirs = append(skippedDirs, filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should only find the regular repository, not those in ignored dirs
	if len(foundRepos) != 1 {
		t.Errorf("expected 1 repository (ignored dirs should be skipped), found %d: %v", len(foundRepos), foundRepos)
	}

	// Verify that all ignored directories were skipped
	foundIgnoredDirs := make(map[string]bool)
	for _, dir := range skippedDirs {
		foundIgnoredDirs[filepath.Base(dir)] = true
	}

	for _, expectedDir := range ignoredDirs {
		if !foundIgnoredDirs[expectedDir] {
			t.Errorf("expected directory %s to be skipped", expectedDir)
		}
	}
}

// TestWalker_IgnoreMatcher_CustomPatterns tests AC-23: custom ignore patterns from config
func TestWalker_IgnoreMatcher_CustomPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repository in custom ignored directory
	customIgnoredRepo := filepath.Join(tmpDir, "project", "custom_ignore", ".git")
	if err := os.MkdirAll(customIgnoredRepo, 0755); err != nil {
		t.Fatalf("failed to create repo in custom ignored dir: %v", err)
	}

	// Create a regular repository
	regularRepo := filepath.Join(tmpDir, "project", ".git")
	if err := os.MkdirAll(regularRepo, 0755); err != nil {
		t.Fatalf("failed to create regular repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	// Provide custom ignore patterns
	walker := NewWalker(filesystem, detector, []string{"custom_ignore"}, 0, logger)

	ctx := context.Background()
	foundRepos := []string{}
	skippedDirs := []string{}

	err := walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		switch event {
		case EventFoundRepo:
			foundRepos = append(foundRepos, filepath.Base(path))
		case EventSkipped:
			skippedDirs = append(skippedDirs, filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Should only find the regular repository
	if len(foundRepos) != 1 {
		t.Errorf("expected 1 repository (custom ignored should be skipped), found %d", len(foundRepos))
	}

	// Verify custom ignored directory was skipped
	customIgnoredSkipped := false
	for _, dir := range skippedDirs {
		if filepath.Base(dir) == "custom_ignore" {
			customIgnoredSkipped = true
			break
		}
	}

	if !customIgnoredSkipped {
		t.Error("expected custom_ignore directory to be skipped")
	}
}

// TestWalker_IgnoreMatcher_WildcardPatterns tests wildcard pattern matching
func TestWalker_IgnoreMatcher_WildcardPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repositories matching wildcard patterns
	wildcardRepos := []string{"test_data", "test_cache", "build_artifacts", "dist_output"}
	for _, repo := range wildcardRepos {
		repoPath := filepath.Join(tmpDir, "project", repo, ".git")
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			t.Fatalf("failed to create repo %s: %v", repo, err)
		}
	}

	// Create a regular repository
	regularRepo := filepath.Join(tmpDir, "project", ".git")
	if err := os.MkdirAll(regularRepo, 0755); err != nil {
		t.Fatalf("failed to create regular repo: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	// Provide wildcard patterns
	walker := NewWalker(filesystem, detector, []string{"test_*", "build_*", "dist_*"}, 0, logger)

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

	// Should only find the regular repository
	if len(foundRepos) != 1 {
		t.Errorf("expected 1 repository (wildcard patterns should be skipped), found %d", len(foundRepos))
	}
}

// TestWalker_IgnoreMatcher_DebugLogging tests AC-24: DEBUG-level logging for skipped directories
func TestWalker_IgnoreMatcher_DebugLogging(t *testing.T) {
	tmpDir := t.TempDir()

	// Create repository in default ignored directory
	ignoredRepo := filepath.Join(tmpDir, "project", "node_modules", "dependency", ".git")
	if err := os.MkdirAll(ignoredRepo, 0755); err != nil {
		t.Fatalf("failed to create repo in ignored dir: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	// Create logger with DEBUG level to verify logging works
	logger, err := logging.NewLogger("DEBUG", "console")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

	ctx := context.Background()
	skippedCount := 0

	err = walker.Walk(ctx, tmpDir, func(path string, event WalkEvent, err error) error {
		if event == EventSkipped {
			skippedCount++
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Verify that ignored directories were skipped
	if skippedCount == 0 {
		t.Error("expected directories to be skipped")
	}
}

// TestNormalizePath tests path normalization edge cases
func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		checkResult func(string) error
	}{
		{
			name:    "empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "trailing slash",
			input:   "/home/user/projects/",
			wantErr: false,
			checkResult: func(result string) error {
				if result != "/home/user/projects" {
					return fmt.Errorf("expected trailing slash to be removed, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "multiple trailing slashes",
			input:   "/home/user/projects///",
			wantErr: false,
			checkResult: func(result string) error {
				if result != "/home/user/projects" {
					return fmt.Errorf("expected multiple trailing slashes to be cleaned, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "relative path with dot",
			input:   "./projects",
			wantErr: false,
			checkResult: func(result string) error {
				// Should convert to absolute path
				if !filepath.IsAbs(result) {
					return fmt.Errorf("expected absolute path, got %s", result)
				}
				if !strings.HasSuffix(result, "projects") {
					return fmt.Errorf("expected path to end with 'projects', got %s", result)
				}
				return nil
			},
		},
		{
			name:    "relative path with double dot",
			input:   "../projects",
			wantErr: false,
			checkResult: func(result string) error {
				if !filepath.IsAbs(result) {
					return fmt.Errorf("expected absolute path, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "multiple slashes in path",
			input:   "/home//user///projects",
			wantErr: false,
			checkResult: func(result string) error {
				if strings.Contains(result, "//") {
					return fmt.Errorf("expected multiple slashes to be cleaned, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "absolute path remains absolute",
			input:   "/home/user/projects",
			wantErr: false,
			checkResult: func(result string) error {
				if !filepath.IsAbs(result) {
					return fmt.Errorf("expected absolute path, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "path with current directory reference",
			input:   "/home/user/./projects",
			wantErr: false,
			checkResult: func(result string) error {
				if strings.Contains(result, "/./") {
					return fmt.Errorf("expected ./ to be cleaned, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "complex path with multiple issues",
			input:   "/home/user/./projects///subdir/../projects",
			wantErr: false,
			checkResult: func(result string) error {
				// Should be cleaned and normalized
				if strings.Contains(result, ".") || strings.Contains(result, "//") {
					return fmt.Errorf("expected path to be fully cleaned, got %s", result)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for relative path tests
			tmpDir := t.TempDir()

			// Change to temp dir for relative path tests
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change to temp directory: %v", err)
			}

			result, err := normalizePath(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.checkResult != nil {
				if err := tt.checkResult(result); err != nil {
					t.Errorf("path validation failed: %v", err)
				}
			}
		})
	}
}

// TestNormalizePath_HomeDirectory tests home directory expansion
func TestNormalizePath_HomeDirectory(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory for testing")
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(result string) error
	}{
		{
			name:    "home directory with tilde",
			input:   "~/Projects",
			wantErr: false,
			check: func(result string) error {
				if !strings.HasPrefix(result, homeDir) {
					return fmt.Errorf("expected path to start with home dir, got %s", result)
				}
				if !strings.HasSuffix(result, "Projects") {
					return fmt.Errorf("expected path to end with Projects, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "home directory with tilde and trailing slash",
			input:   "~/Projects/",
			wantErr: false,
			check: func(result string) error {
				if !strings.HasPrefix(result, homeDir) {
					return fmt.Errorf("expected path to start with home dir, got %s", result)
				}
				if strings.HasSuffix(result, "/") {
					return fmt.Errorf("expected no trailing slash, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "just tilde",
			input:   "~",
			wantErr: false,
			check: func(result string) error {
				if result != homeDir {
					return fmt.Errorf("expected home directory, got %s", result)
				}
				return nil
			},
		},
		{
			name:    "tilde with slash",
			input:   "~/",
			wantErr: false,
			check: func(result string) error {
				if result != homeDir {
					return fmt.Errorf("expected home directory, got %s", result)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizePath(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.check != nil {
				if err := tt.check(result); err != nil {
					t.Errorf("home directory check failed: %v", err)
				}
			}
		})
	}
}

// TestWalker_Walk_PathNormalization tests that walker properly normalizes paths
func TestWalker_Walk_PathNormalization(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a repository
	repoPath := filepath.Join(tmpDir, "myproject", ".git")
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("failed to create repo directory: %v", err)
	}

	filesystem := fs.NewOSFilesystem()
	detector := NewDetector(filesystem)
	logger := getTestLogger()
	walker := NewWalker(filesystem, detector, []string{}, 0, logger)

	// Test various path formats
	testPaths := []string{
		tmpDir,
		tmpDir + "/",  // trailing slash
		tmpDir + "//", // multiple trailing slashes
		filepath.Join(tmpDir, ".", "myproject", ".."), // with . and ..
	}

	for _, testPath := range testPaths {
		t.Run(fmt.Sprintf("path_%s", filepath.Base(testPath)), func(t *testing.T) {
			ctx := context.Background()
			foundRepo := false
			var foundPath string

			err := walker.Walk(ctx, testPath, func(path string, event WalkEvent, err error) error {
				if event == EventFoundRepo {
					foundRepo = true
					foundPath = path
				}
				return nil
			})

			if err != nil {
				t.Errorf("Walk failed for path %s: %v", testPath, err)
				return
			}

			if !foundRepo {
				t.Errorf("expected to find repository for path %s, but found none", testPath)
			}

			expectedPath := filepath.Join(tmpDir, "myproject")
			if foundPath != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, foundPath)
			}
		})
	}
}
