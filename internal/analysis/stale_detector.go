package analysis

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// StaleDetector detects stale analyses
type StaleDetector struct {
	store  AnalysisStore
	logger *zap.Logger
}

// NewStaleDetector creates a new stale detector
func NewStaleDetector(store AnalysisStore) *StaleDetector {
	return &StaleDetector{
		store:  store,
		logger: zap.NewNop(),
	}
}

// IsOutdated checks if an analysis is outdated
func (d *StaleDetector) IsOutdated(ctx context.Context, analysis *Analysis) (bool, error) {
	gitHead, err := d.store.GetGitHeadForProject(ctx, uuid.MustParse(analysis.ProjectID))
	if err != nil {
		return false, WrapError("", "failed to get git head for project", err)
	}

	if gitHead == nil {
		// Git head is unknown, consider analysis as outdated
		return true, nil
	}

	return *gitHead != analysis.AnalyzedGitHead, nil
}

// ListNeedingAnalysis lists all projects that need analysis
func (d *StaleDetector) ListNeedingAnalysis(ctx context.Context) ([]NeedsAnalysisResult, error) {
	return d.store.ListProjectsNeedingAnalysis(ctx)
}
