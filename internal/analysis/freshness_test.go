package analysis

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"project-dash/pkg/models"
)

// osDevNull returns the platform null device for `git hash-object`.
func osDevNull() string {
	return os.DevNull
}

var gitEnv = []string{
	"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
	"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
}

// gitCmd runs a git command in dir and fails the test on error.
func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(), gitEnv...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v (%s)", args, err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out))
}

// initRepo creates a git repository in a temp dir with one commit on main and
// returns its path and HEAD.
func initRepo(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	gitCmd(t, dir, "init", "-q")
	gitCmd(t, dir, "checkout", "-q", "-b", "main")
	gitCmd(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	return dir, gitCmd(t, dir, "rev-parse", "HEAD")
}

// commit adds an empty commit and returns the new HEAD.
func commit(t *testing.T, dir string) string {
	t.Helper()
	gitCmd(t, dir, "commit", "-q", "--allow-empty", "-m", "more")
	return gitCmd(t, dir, "rev-parse", "HEAD")
}

func TestFreshness_NoAnalysis(t *testing.T) {
	dir, _ := initRepo(t)
	f := Compute(Inputs{RootPath: dir})
	if f.Status != StatusUnknown {
		t.Fatalf("expected unknown with no analysis, got %s", f.Status)
	}
	if f.Note == "" {
		t.Fatal("expected an explanatory note")
	}
}

func TestFreshness_NoAnchor(t *testing.T) {
	dir, _ := initRepo(t)
	f := Compute(Inputs{RootPath: dir, Analysis: &models.Analysis{AnalyzedGitHead: ""}})
	if f.Status != StatusUnknown {
		t.Fatalf("expected unknown with empty anchor, got %s", f.Status)
	}
}

func TestFreshness_Fresh(t *testing.T) {
	dir, head := initRepo(t)
	f := Compute(Inputs{RootPath: dir, Analysis: &models.Analysis{AnalyzedGitHead: head, AnalyzedAt: "2024-01-01T00:00:00Z"}})
	if f.Status != StatusFresh {
		t.Fatalf("expected fresh, got %s (%s)", f.Status, f.Note)
	}
	if f.CurrentGitHead != head || f.AnalyzedGitHead != head {
		t.Fatalf("heads not populated: %+v", f)
	}
}

func TestFreshness_StaleWithCommitsBehind(t *testing.T) {
	dir, analyzed := initRepo(t)
	commit(t, dir)
	commit(t, dir)

	f := Compute(Inputs{RootPath: dir, Analysis: &models.Analysis{AnalyzedGitHead: analyzed}})
	if f.Status != StatusStale {
		t.Fatalf("expected stale, got %s", f.Status)
	}
	if f.CommitsBehind == nil || *f.CommitsBehind != 2 {
		t.Fatalf("expected 2 commits behind, got %+v", f.CommitsBehind)
	}
}

func TestFreshness_NonAncestorStillCounts(t *testing.T) {
	dir, analyzed := initRepo(t)

	// Force-push simulation: reset to an orphan root commit so the analyzed
	// HEAD is no longer an ancestor of the current one. The repo still has
	// one commit the analysis has not seen, so commits_behind reports 1.
	emptyTree := gitCmd(t, dir, "hash-object", "-t", "tree", osDevNull())
	orphan := gitCmd(t, dir, "commit-tree", emptyTree, "-m", "orphan root")
	gitCmd(t, dir, "reset", "--hard", orphan)

	f := Compute(Inputs{RootPath: dir, Analysis: &models.Analysis{AnalyzedGitHead: analyzed}})
	if f.Status != StatusStale {
		t.Fatalf("expected stale, got %s", f.Status)
	}
	if f.CommitsBehind == nil || *f.CommitsBehind != 1 {
		t.Fatalf("expected 1 commits behind, got %+v", f.CommitsBehind)
	}
}

func TestFreshness_RepoUnavailableFallsBackToStoredHead(t *testing.T) {
	analyzed := "aaa111"
	// Stored scan HEAD matches the analysis → fresh as of last scan.
	f := Compute(Inputs{RootPath: t.TempDir(), StoredHead: analyzed, Analysis: &models.Analysis{AnalyzedGitHead: analyzed}})
	if f.Status != StatusFresh {
		t.Fatalf("expected fresh fallback, got %s (%s)", f.Status, f.Note)
	}

	// Stored scan HEAD moved past the analysis → stale, even without the repo.
	f = Compute(Inputs{RootPath: t.TempDir(), StoredHead: "bbb222", Analysis: &models.Analysis{AnalyzedGitHead: analyzed}})
	if f.Status != StatusStale {
		t.Fatalf("expected stale fallback, got %s", f.Status)
	}
	if f.CommitsBehind != nil {
		t.Fatal("commits_behind must be omitted in fallback mode")
	}

	// No fallback available → unknown.
	f = Compute(Inputs{RootPath: t.TempDir(), Analysis: &models.Analysis{AnalyzedGitHead: analyzed}})
	if f.Status != StatusUnknown {
		t.Fatalf("expected unknown, got %s", f.Status)
	}
}

func TestFreshnessForLatest_PicksLatest(t *testing.T) {
	dir, _ := initRepo(t)
	analyses := []*models.Analysis{
		{AnalyzedGitHead: "newest", AnalyzedAt: "2024-03-01T00:00:00Z"},
		{AnalyzedGitHead: "oldest", AnalyzedAt: "2024-01-01T00:00:00Z"},
	}
	f := FreshnessForLatest(&models.Project{RootPath: dir}, "", analyses)
	if f.AnalyzedGitHead != "newest" {
		t.Fatalf("expected latest analysis to be used, got %s", f.AnalyzedGitHead)
	}

	none := FreshnessForLatest(&models.Project{RootPath: dir}, "", nil)
	if none.Status != StatusUnknown {
		t.Fatalf("expected unknown for empty analyses, got %s", none.Status)
	}
}
