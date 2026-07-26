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

func TestExtractGitMetadata_NewSignals(t *testing.T) {
	dir := t.TempDir()
	createGitRepo(t, dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "first")

	// Second commit attributed to a different author -> 2 unique contributors.
	os.WriteFile(filepath.Join(dir, "other.go"), []byte("package main"), 0644)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "--author", "Other <other@test.com>", "-m", "second")

	mustRun(t, dir, "git", "tag", "v1.0.0")
	mustRun(t, dir, "git", "remote", "add", "origin", "https://github.com/foo/bar.git")

	result, err := metadata.ExtractGitMetadata(dir)
	if err != nil {
		t.Fatalf("ExtractGitMetadata failed: %v", err)
	}

	if result.CommitCount != 2 {
		t.Errorf("expected commit_count=2, got %d", result.CommitCount)
	}
	if result.FirstCommitAt == nil {
		t.Error("expected first_commit_at to be set")
	}
	if result.CommitVelocity90d != 2 {
		t.Errorf("expected commit_velocity_90d=2, got %d", result.CommitVelocity90d)
	}
	if result.ContributorCount != 2 {
		t.Errorf("expected contributor_count=2, got %d", result.ContributorCount)
	}
	if result.TagCount != 1 {
		t.Errorf("expected tag_count=1, got %d", result.TagCount)
	}
	if result.RemoteURL == nil || *result.RemoteURL == "" {
		t.Error("expected remote_url to be set")
	} else if *result.RemoteURL != "https://github.com/foo/bar.git" {
		t.Errorf("unexpected remote_url: %s", *result.RemoteURL)
	}
	if !result.IsPublished {
		t.Error("expected is_published=true for a github remote")
	}
}

func TestExtractGitMetadata_LocalRemoteNotPublished(t *testing.T) {
	dir := t.TempDir()
	createGitRepo(t, dir)

	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-m", "first")
	mustRun(t, dir, "git", "remote", "add", "origin", "/local/path/to/repo.git")

	result, err := metadata.ExtractGitMetadata(dir)
	if err != nil {
		t.Fatalf("ExtractGitMetadata failed: %v", err)
	}

	if result.RemoteURL == nil {
		t.Error("expected remote_url to be set")
	}
	if result.IsPublished {
		t.Error("expected is_published=false for a local remote")
	}
}
