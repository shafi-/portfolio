// Package analysis provides freshness computation for stored analyses.
//
// Freshness answers one question: does this stored analysis still describe the
// repository as it exists right now? It compares the git HEAD the analysis was
// made against (analyzed_git_head) with the repository's live HEAD, read at
// query time. Freshness is always computed, never stored (ADR-023): a stored
// freshness flag would go stale itself.
package analysis

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"project-dash/pkg/models"
)

// Status is the outcome of a freshness check.
type Status string

const (
	// StatusFresh means the live repository HEAD matches the HEAD the
	// analysis was made against.
	StatusFresh Status = "fresh"
	// StatusStale means the repository has moved past the analyzed HEAD.
	StatusStale Status = "stale"
	// StatusUnknown means freshness cannot be determined — no analysis, no
	// HEAD anchor on the analysis, or the repository is unavailable. An
	// unknown result must be surfaced as unknown, never treated as fresh.
	StatusUnknown Status = "unknown"
)

// Freshness is the result of a staleness check. It is embedded in tool
// responses so agents see it before relying on an analysis.
type Freshness struct {
	Status          Status `json:"status"`
	AnalyzedGitHead string `json:"analyzed_git_head,omitempty"`
	CurrentGitHead  string `json:"current_git_head,omitempty"`
	AnalyzedAt      string `json:"analyzed_at,omitempty"`
	// CommitsBehind counts commits reachable from the live HEAD but not from
	// the analyzed HEAD (omitted on fresh/unknown results and when the count
	// cannot be determined — e.g. the analyzed HEAD was garbage-collected).
	CommitsBehind *int   `json:"commits_behind,omitempty"`
	Note          string `json:"note,omitempty"`
}

// Inputs carries everything a freshness check needs. StoredHead is the
// scan-time HEAD from project metadata (metadata.git_head); it is the
// fallback comparison point when the repository is unavailable.
type Inputs struct {
	RootPath   string
	StoredHead string
	Analysis   *models.Analysis // the latest analysis; nil means none stored
}

// Compute evaluates the freshness of a project's latest analysis.
func Compute(in Inputs) Freshness {
	if in.Analysis == nil {
		return Freshness{Status: StatusUnknown, Note: "no analysis stored"}
	}

	f := Freshness{
		AnalyzedGitHead: in.Analysis.AnalyzedGitHead,
		AnalyzedAt:      in.Analysis.AnalyzedAt,
	}

	if strings.TrimSpace(f.AnalyzedGitHead) == "" {
		f.Status = StatusUnknown
		f.Note = "analysis has no git HEAD anchor"
		return f
	}

	liveHead := gitOutput(in.RootPath, "rev-parse", "HEAD")
	if liveHead == "" {
		// Repository unavailable (missing, moved, or not a git repo) — fall
		// back to the last scan's HEAD so the comparison at least reflects
		// the most recent scan.
		if in.StoredHead == "" {
			return Freshness{
				Status:          StatusUnknown,
				AnalyzedGitHead: f.AnalyzedGitHead,
				AnalyzedAt:      f.AnalyzedAt,
				Note:            "repository unavailable and no scan metadata to compare against",
			}
		}
		f.Note = "repository unavailable — compared against last scan's HEAD"
		f.CurrentGitHead = in.StoredHead
		if in.StoredHead == f.AnalyzedGitHead {
			f.Status = StatusFresh
		} else {
			f.Status = StatusStale
		}
		return f
	}

	f.CurrentGitHead = liveHead
	if liveHead == f.AnalyzedGitHead {
		f.Status = StatusFresh
		return f
	}

	f.Status = StatusStale
	if n, err := strconv.Atoi(gitOutput(in.RootPath, "rev-list", "--count", f.AnalyzedGitHead+"..HEAD")); err == nil {
		f.CommitsBehind = &n
	}
	return f
}

// FreshnessForLatest is a convenience wrapper for the common shape of tool
// handlers: a project plus its analyses list (latest first, per
// AnalysisStore.ListAnalyses ordering).
func FreshnessForLatest(project *models.Project, storedHead string, analyses []*models.Analysis) Freshness {
	var latest *models.Analysis
	if len(analyses) > 0 {
		latest = analyses[0]
	}
	return Compute(Inputs{RootPath: project.RootPath, StoredHead: storedHead, Analysis: latest})
}

func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// String renders a short human-readable form used by CLI tables.
func (f Freshness) String() string {
	switch f.Status {
	case StatusFresh:
		return "✓ fresh"
	case StatusStale:
		if f.CommitsBehind != nil {
			return fmt.Sprintf("⚠ stale (+%d commits)", *f.CommitsBehind)
		}
		return "⚠ stale"
	default:
		return "? unknown"
	}
}
