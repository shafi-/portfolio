package store

import (
	"testing"

	"project-dash/pkg/models"
)

// setupFeatureAnalysis inserts a project + analysis and returns the analysis
// ID, so feature tests have a valid analysis_id foreign key to attach to.
func setupFeatureAnalysis(t *testing.T, s *testStore, projectID, analysisID string) string {
	t.Helper()
	upsertTestProject(t, s, projectID)
	analysis := &models.Analysis{
		ID:              analysisID,
		ProjectID:       projectID,
		Analyzer:        "claude-code",
		AnalyzedGitHead: "abc123",
		AnalyzedAt:      "2024-01-01T12:00:00Z",
	}
	if err := s.analyses.CreateAnalysis(analysis); err != nil {
		t.Fatalf("create analysis: %v", err)
	}
	return analysisID
}

// TestFeatureStore_Tier3ColumnsRoundTrip asserts the three Tier-3 columns
// (implementation_status, feature_architecture, pattern) survive a
// create → read cycle — guards against the test schema drifting from the
// production migration that added them.
func TestFeatureStore_Tier3ColumnsRoundTrip(t *testing.T) {
	s := setupTestStore(t)
	defer s.db.Close()
	aid := setupFeatureAnalysis(t, s, "p1", "a1")

	in := &models.Feature{
		ID:                   "f1",
		AnalysisID:           aid,
		Name:                 "User Authentication",
		Description:          "JWT-based login",
		Confidence:           0.9,
		ImplementationStatus: "complete",
		FeatureArchitecture:  "JWT middleware + session store",
		Pattern:              "Middleware",
	}
	if err := s.features.CreateFeature(in); err != nil {
		t.Fatalf("create feature: %v", err)
	}

	out, err := s.features.GetFeature("f1")
	if err != nil {
		t.Fatalf("get feature: %v", err)
	}
	if out == nil {
		t.Fatal("expected feature, got nil")
	}
	if out.ImplementationStatus != "complete" {
		t.Errorf("implementation_status: got %q, want complete", out.ImplementationStatus)
	}
	if out.FeatureArchitecture != "JWT middleware + session store" {
		t.Errorf("feature_architecture: got %q", out.FeatureArchitecture)
	}
	if out.Pattern != "Middleware" {
		t.Errorf("pattern: got %q, want Middleware", out.Pattern)
	}
	if out.Description != "JWT-based login" || out.Confidence != 0.9 {
		t.Errorf("tier-2 fields drifted: description=%q confidence=%v", out.Description, out.Confidence)
	}
}

// TestFeatureStore_SearchFeatures exercises each SearchFeatures filter and the
// join to analyses (filtering by project). It also implicitly proves the
// store's plain-`?` positional binding binds args in the right order across a
// multi-condition query.
func TestFeatureStore_SearchFeatures(t *testing.T) {
	s := setupTestStore(t)
	defer s.db.Close()
	aid := setupFeatureAnalysis(t, s, "p1", "a1")

	features := []*models.Feature{
		{ID: "f1", AnalysisID: aid, Name: "User Authentication", Description: "JWT login", Confidence: 0.9, ImplementationStatus: "complete", Pattern: "Middleware"},
		{ID: "f2", AnalysisID: aid, Name: "Search", Description: "Full-text search", Confidence: 0.8, ImplementationStatus: "partial", Pattern: "Repository"},
		{ID: "f3", AnalysisID: aid, Name: "Billing", Description: "Stripe integration", Confidence: 0.7, ImplementationStatus: "planned", Pattern: "Service Layer"},
	}
	for _, f := range features {
		if err := s.features.CreateFeature(f); err != nil {
			t.Fatalf("create feature %s: %v", f.ID, err)
		}
	}

	// Filter by project (join).
	got, err := s.features.SearchFeatures(FeatureSearchOptions{ProjectID: "p1"})
	if err != nil {
		t.Fatalf("search by project: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("project filter: got %d, want 3", len(got))
	}
	// Project with no features.
	got, _ = s.features.SearchFeatures(FeatureSearchOptions{ProjectID: "nope"})
	if len(got) != 0 {
		t.Errorf("empty project: got %d, want 0", len(got))
	}

	// Filter by implementation_status.
	got, err = s.features.SearchFeatures(FeatureSearchOptions{ImplementationStatus: "partial"})
	if err != nil {
		t.Fatalf("search by status: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Search" {
		t.Errorf("status filter: got %v, want [Search]", names(got))
	}

	// Filter by pattern (LIKE).
	got, err = s.features.SearchFeatures(FeatureSearchOptions{Pattern: "Middleware"})
	if err != nil {
		t.Fatalf("search by pattern: %v", err)
	}
	if len(got) != 1 || got[0].Name != "User Authentication" {
		t.Errorf("pattern filter: got %v, want [User Authentication]", names(got))
	}

	// Free-text query across name/description.
	got, err = s.features.SearchFeatures(FeatureSearchOptions{Query: "stripe"})
	if err != nil {
		t.Fatalf("search by query: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Billing" {
		t.Errorf("query filter: got %v, want [Billing]", names(got))
	}

	// Combined filters narrow further.
	got, err = s.features.SearchFeatures(FeatureSearchOptions{ProjectID: "p1", ImplementationStatus: "complete", Pattern: "Middleware"})
	if err != nil {
		t.Fatalf("combined search: %v", err)
	}
	if len(got) != 1 || got[0].Name != "User Authentication" {
		t.Errorf("combined filter: got %v, want [User Authentication]", names(got))
	}
}

// TestFeatureStore_UpsertByNameMerge mirrors the storeFeature handler's
// upsert: a Tier-2 feature is created, then a Tier-3 deep-dive (same analysis +
// name) enriches implementation_status/architecture/pattern WITHOUT clobbering
// the stored description or confidence.
func TestFeatureStore_UpsertByNameMerge(t *testing.T) {
	s := setupTestStore(t)
	defer s.db.Close()
	aid := setupFeatureAnalysis(t, s, "p1", "a1")

	// Tier 2: create.
	tier2 := &models.Feature{
		ID: "f1", AnalysisID: aid, Name: "User Authentication",
		Description: "JWT-based login", Confidence: 0.9,
	}
	if err := s.features.CreateFeature(tier2); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Tier 3 deep-dive arrives as the same name with only Tier-3 fields. The
	// handler reads the existing row, overlays supplied fields, and updates.
	existing, err := s.features.GetByAnalysisAndName(aid, "User Authentication")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if existing == nil {
		t.Fatal("expected existing feature for upsert")
	}
	merged := *existing
	merged.ImplementationStatus = "complete"
	merged.FeatureArchitecture = "JWT middleware + session store"
	merged.Pattern = "Middleware"
	// Note: description/confidence intentionally NOT overwritten (merge semantics).
	if err := s.features.UpdateFeature(&merged); err != nil {
		t.Fatalf("update: %v", err)
	}

	out, _ := s.features.GetFeature("f1")
	if out == nil {
		t.Fatal("expected feature after upsert")
	}
	// Tier-3 fields enriched.
	if out.ImplementationStatus != "complete" || out.Pattern != "Middleware" {
		t.Errorf("tier-3 not enriched: status=%q pattern=%q", out.ImplementationStatus, out.Pattern)
	}
	// Tier-2 facts preserved.
	if out.Description != "JWT-based login" || out.Confidence != 0.9 {
		t.Errorf("tier-2 clobbered: description=%q confidence=%v", out.Description, out.Confidence)
	}
	// Still exactly one row (no duplicate from the "second call").
	list, _ := s.features.ListByAnalysis(aid)
	if len(list) != 1 {
		t.Errorf("upsert created duplicates: got %d rows, want 1", len(list))
	}
}

func names(fs []*models.Feature) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Name
	}
	return out
}
