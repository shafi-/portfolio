package cli

import (
	"project-dash/internal/analysis"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

// analysisFreshnessLabel renders the freshness of a project's latest analysis
// for CLI tables. Errors degrade to the unknown label — freshness reporting
// must never fail a listing.
func analysisFreshnessLabel(as *store.AnalysisStore, ms *store.MetadataStore, project *models.Project) string {
	analyses, err := as.ListAnalyses(project.ID)
	if err != nil || len(analyses) == 0 {
		return "– none"
	}

	var storedHead string
	if meta, err := ms.GetMetadata(project.ID); err == nil && meta != nil {
		storedHead = meta.GitHead
	}

	f := analysis.FreshnessForLatest(project, storedHead, analyses)
	return f.String()
}
