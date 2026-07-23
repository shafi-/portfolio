package metadata_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"project-dash/internal/metadata"
)

func createGitRepo(t *testing.T, dir string) {
	t.Helper()
	mustRun(t, dir, "git", "init")
	mustRun(t, dir, "git", "config", "user.email", "test@test.com")
	mustRun(t, dir, "git", "config", "user.name", "Test")
}

func mustRun(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
	return string(out)
}

func TestExtractGitMetadata_Normal(t *testing.T) {
	dir := t.TempDir()
	createGitRepo(t, dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "initial commit")

	result, err := metadata.ExtractGitMetadata(dir)
	if err != nil {
		t.Fatalf("ExtractGitMetadata failed: %v", err)
	}

	if result.GitHead == nil || *result.GitHead == "" {
		t.Error("expected git_head to be set")
	}
	if result.CommitCount != 1 {
		t.Errorf("expected commit_count=1, got %d", result.CommitCount)
	}
	if result.LastCommitAt == nil {
		t.Error("expected last_commit_at to be set")
	}
	if result.LastModifiedAt == nil {
		t.Error("expected last_modified_at to be set")
	}
}

func TestExtractGitMetadata_EmptyRepo(t *testing.T) {
	dir := t.TempDir()
	createGitRepo(t, dir)

	result, err := metadata.ExtractGitMetadata(dir)
	if err != nil {
		t.Fatalf("ExtractGitMetadata failed: %v", err)
	}

	if result.GitHead != nil {
		t.Error("expected git_head to be nil for empty repo")
	}
	if result.CommitCount != 0 {
		t.Errorf("expected commit_count=0, got %d", result.CommitCount)
	}
	if result.LastCommitAt != nil {
		t.Error("expected last_commit_at to be nil for empty repo")
	}
}

func TestExtractGitMetadata_DetachedHead(t *testing.T) {
	dir := t.TempDir()
	createGitRepo(t, dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "initial commit")

	mustRun(t, dir, "git", "checkout", "--detach")

	result, err := metadata.ExtractGitMetadata(dir)
	if err != nil {
		t.Fatalf("ExtractGitMetadata failed: %v", err)
	}

	if result.GitHead == nil || *result.GitHead == "" {
		t.Error("expected git_head to be set in detached HEAD state")
	}
	if result.DefaultBranch == nil {
		t.Log("default_branch may be nil when no remote is configured")
	}
}
