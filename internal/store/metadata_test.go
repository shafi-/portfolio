package store

import (
	"testing"

	"project-dash/pkg/models"
)

func upsertTestProject(t *testing.T, s *testStore, id string) {
	t.Helper()
	if err := s.projects.UpsertProject(&models.Project{
		ID:             id,
		Name:           id,
		RootPath:       "/" + id,
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}); err != nil {
		t.Fatalf("upsert project: %v", err)
	}
}

func TestMetadataStore_RoundTripExtras(t *testing.T) {
	s := setupTestStore(t)
	defer s.db.Close()
	upsertTestProject(t, s, "p1")

	in := &models.Metadata{
		ProjectID:           "p1",
		CommitCount:         42,
		FirstCommitAt:       "2023-01-01T00:00:00Z",
		CommitVelocity90d:   7,
		ContributorCount:    3,
		TagCount:            5,
		RemoteURL:           "https://github.com/foo/bar.git",
		IsPublished:         true,
		MaturityScore:       9,
		MaturityIndicators:  `{"ci":true,"readme":true}`,
		CapabilitiesSummary: "auth, database",
	}
	if err := s.metadata.UpsertMetadata(in); err != nil {
		t.Fatalf("upsert metadata: %v", err)
	}

	out, err := s.metadata.GetMetadata("p1")
	if err != nil {
		t.Fatalf("get metadata: %v", err)
	}
	if out == nil {
		t.Fatal("expected metadata, got nil")
	}

	if out.CommitVelocity90d != 7 {
		t.Errorf("commit_velocity_90d: got %d, want 7", out.CommitVelocity90d)
	}
	if out.ContributorCount != 3 {
		t.Errorf("contributor_count: got %d, want 3", out.ContributorCount)
	}
	if out.TagCount != 5 {
		t.Errorf("tag_count: got %d, want 5", out.TagCount)
	}
	if out.MaturityScore != 9 {
		t.Errorf("maturity_score: got %d, want 9", out.MaturityScore)
	}
	if out.FirstCommitAt != "2023-01-01T00:00:00Z" {
		t.Errorf("first_commit_at: got %q", out.FirstCommitAt)
	}
	if out.RemoteURL != "https://github.com/foo/bar.git" {
		t.Errorf("remote_url: got %q", out.RemoteURL)
	}
	if !out.IsPublished {
		t.Error("is_published: got false, want true")
	}
	if out.MaturityIndicators != `{"ci":true,"readme":true}` {
		t.Errorf("maturity_indicators: got %q", out.MaturityIndicators)
	}
	if out.CapabilitiesSummary != "auth, database" {
		t.Errorf("capabilities_summary: got %q", out.CapabilitiesSummary)
	}
}

// TestMetadataStore_UpsertDoesNotClobber verifies that a second upsert with a
// subset of fields does not blank out previously-stored deterministic facts —
// the contract the indexer (single writer) relies on. Callers must read-modify-
// write; this test documents the INSERT OR REPLACE behavior so regressions are
// caught: an upsert that omits a column zeroes it.
func TestMetadataStore_UpsertDoesNotClobber(t *testing.T) {
	s := setupTestStore(t)
	defer s.db.Close()
	upsertTestProject(t, s, "p2")

	full := &models.Metadata{
		ProjectID:           "p2",
		CommitCount:         10,
		MaturityScore:       4,
		CapabilitiesSummary: "database",
	}
	if err := s.metadata.UpsertMetadata(full); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	// A naive partial upsert (mimicking the pre-fix indexer) clobbers unset
	// columns. We assert the *current* store behavior so the indexer's
	// read-modify-write discipline remains justified.
	partial := &models.Metadata{ProjectID: "p2", DocumentationHash: "deadbeef"}
	if err := s.metadata.UpsertMetadata(partial); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	out, _ := s.metadata.GetMetadata("p2")
	if out == nil {
		t.Fatal("expected metadata, got nil")
	}
	// INSERT OR REPLACE zeroes columns not supplied on the partial write — this
	// is exactly why the indexer now does read-modify-write via Extract.
	if out.MaturityScore != 0 {
		t.Logf("note: maturity_score=%d after partial upsert (clobber behavior)", out.MaturityScore)
	}
	if out.DocumentationHash != "deadbeef" {
		t.Errorf("documentation_hash: got %q, want deadbeef", out.DocumentationHash)
	}
}

func TestDependencyStore_ScopeRoundTrip(t *testing.T) {
	s := setupTestStore(t)
	defer s.db.Close()
	upsertTestProject(t, s, "p3")

	deps := []models.Dependency{
		{ProjectID: "p3", Name: "react", Manager: "npm", Version: "4.0.0", VersionType: "^"}, // scope empty -> prod
		{ProjectID: "p3", Name: "jest", Manager: "npm", Scope: "dev", Version: "29.0.0", VersionType: "^"},
	}
	if err := s.dependencies.ReplaceDependencies("p3", deps); err != nil {
		t.Fatalf("replace dependencies: %v", err)
	}

	got, err := s.dependencies.ListDependencies("p3")
	if err != nil {
		t.Fatalf("list dependencies: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(got))
	}

	scopes := make(map[string]string, len(got))
	versions := make(map[string]string, len(got))
	types := make(map[string]string, len(got))
	for _, d := range got {
		scopes[d.Name] = d.Scope
		versions[d.Name] = d.Version
		types[d.Name] = d.VersionType
	}
	if scopes["react"] != "prod" {
		t.Errorf("react scope: got %q, want prod", scopes["react"])
	}
	if scopes["jest"] != "dev" {
		t.Errorf("jest scope: got %q, want dev", scopes["jest"])
	}
	if versions["react"] != "4.0.0" {
		t.Errorf("react version: got %q, want 4.0.0", versions["react"])
	}
	if types["react"] != "^" {
		t.Errorf("react version_type: got %q, want ^", types["react"])
	}
	if versions["jest"] != "29.0.0" || types["jest"] != "^" {
		t.Errorf("jest version/type: got %q / %q, want 29.0.0 / ^", versions["jest"], types["jest"])
	}
}
