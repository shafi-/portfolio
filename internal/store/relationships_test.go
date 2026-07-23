package store

import (
	"testing"

	"project-dash/pkg/models"
)

func TestRelationshipStore_CreateRelationship(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project1 := &models.Project{
		ID:             "proj-1",
		Name:           "Project 1",
		RootPath:       "/test/path1",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	project2 := &models.Project{
		ID:             "proj-2",
		Name:           "Project 2",
		RootPath:       "/test/path2",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	if err := store.projects.UpsertProject(project1); err != nil {
		t.Fatalf("failed to create project 1: %v", err)
	}

	if err := store.projects.UpsertProject(project2); err != nil {
		t.Fatalf("failed to create project 2: %v", err)
	}

	relationship := &models.Relationship{
		ID:            "rel-1",
		SourceProject: "proj-1",
		TargetProject: "proj-2",
		Type:          "Similar",
		Description:   "These projects are similar",
		Confidence:    0.9,
	}

	if err := store.relationships.CreateRelationship(relationship); err != nil {
		t.Fatalf("failed to create relationship: %v", err)
	}

	retrieved, err := store.relationships.GetRelationship("rel-1")
	if err != nil {
		t.Fatalf("failed to get relationship: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected relationship to be retrieved, got nil")
	}

	if retrieved.SourceProject != "proj-1" {
		t.Errorf("expected source_project 'proj-1', got '%s'", retrieved.SourceProject)
	}

	if retrieved.TargetProject != "proj-2" {
		t.Errorf("expected target_project 'proj-2', got '%s'", retrieved.TargetProject)
	}

	if retrieved.Type != "Similar" {
		t.Errorf("expected type 'Similar', got '%s'", retrieved.Type)
	}

	if retrieved.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %f", retrieved.Confidence)
	}
}

func TestRelationshipStore_ListRelationships(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project1 := &models.Project{
		ID:             "proj-1",
		Name:           "Project 1",
		RootPath:       "/test/path1",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	project2 := &models.Project{
		ID:             "proj-2",
		Name:           "Project 2",
		RootPath:       "/test/path2",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	project3 := &models.Project{
		ID:             "proj-3",
		Name:           "Project 3",
		RootPath:       "/test/path3",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	for _, project := range []*models.Project{project1, project2, project3} {
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	}

	relationships := []*models.Relationship{
		{
			ID:            "rel-1",
			SourceProject: "proj-1",
			TargetProject: "proj-2",
			Type:          "Similar",
			Description:   "Projects 1 and 2 are similar",
			Confidence:    0.9,
		},
		{
			ID:            "rel-2",
			SourceProject: "proj-1",
			TargetProject: "proj-3",
			Type:          "Shared Technology",
			Description:   "Both use Go",
			Confidence:    0.8,
		},
		{
			ID:            "rel-3",
			SourceProject: "proj-2",
			TargetProject: "proj-3",
			Type:          "Evolution",
			Description:   "Project 3 evolved from 2",
			Confidence:    0.7,
		},
	}

	for _, rel := range relationships {
		if err := store.relationships.CreateRelationship(rel); err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
	}

	project1Rels, err := store.relationships.ListRelationships("proj-1")
	if err != nil {
		t.Fatalf("failed to list relationships for project 1: %v", err)
	}

	if len(project1Rels) != 2 {
		t.Errorf("expected 2 relationships for project 1, got %d", len(project1Rels))
	}

	project2Rels, err := store.relationships.ListRelationships("proj-2")
	if err != nil {
		t.Fatalf("failed to list relationships for project 2: %v", err)
	}

	if len(project2Rels) != 2 {
		t.Errorf("expected 2 relationships for project 2, got %d", len(project2Rels))
	}

	project3Rels, err := store.relationships.ListRelationships("proj-3")
	if err != nil {
		t.Fatalf("failed to list relationships for project 3: %v", err)
	}

	if len(project3Rels) != 2 {
		t.Errorf("expected 2 relationships for project 3, got %d", len(project3Rels))
	}
}

func TestRelationshipStore_DeleteRelationship(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project1 := &models.Project{
		ID:             "proj-1",
		Name:           "Project 1",
		RootPath:       "/test/path1",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	project2 := &models.Project{
		ID:             "proj-2",
		Name:           "Project 2",
		RootPath:       "/test/path2",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	if err := store.projects.UpsertProject(project1); err != nil {
		t.Fatalf("failed to create project 1: %v", err)
	}

	if err := store.projects.UpsertProject(project2); err != nil {
		t.Fatalf("failed to create project 2: %v", err)
	}

	relationship := &models.Relationship{
		ID:            "rel-1",
		SourceProject: "proj-1",
		TargetProject: "proj-2",
		Type:          "Similar",
		Description:   "These projects are similar",
		Confidence:    0.9,
	}

	if err := store.relationships.CreateRelationship(relationship); err != nil {
		t.Fatalf("failed to create relationship: %v", err)
	}

	if err := store.relationships.DeleteRelationship("rel-1"); err != nil {
		t.Fatalf("failed to delete relationship: %v", err)
	}

	retrieved, err := store.relationships.GetRelationship("rel-1")
	if err != nil {
		t.Fatalf("failed to get relationship after deletion: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil after deletion, got value")
	}
}

func TestRelationshipStore_DeleteAllForProject(t *testing.T) {
	store := setupTestStore(t)
	defer store.db.Close()

	project1 := &models.Project{
		ID:             "proj-1",
		Name:           "Project 1",
		RootPath:       "/test/path1",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	project2 := &models.Project{
		ID:             "proj-2",
		Name:           "Project 2",
		RootPath:       "/test/path2",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	project3 := &models.Project{
		ID:             "proj-3",
		Name:           "Project 3",
		RootPath:       "/test/path3",
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}

	for _, project := range []*models.Project{project1, project2, project3} {
	if err := store.projects.UpsertProject(project); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	}

	relationships := []*models.Relationship{
		{
			ID:            "rel-1",
			SourceProject: "proj-1",
			TargetProject: "proj-2",
			Type:          "Similar",
			Description:   "Projects 1 and 2 are similar",
			Confidence:    0.9,
		},
		{
			ID:            "rel-2",
			SourceProject: "proj-1",
			TargetProject: "proj-3",
			Type:          "Shared Technology",
			Description:   "Both use Go",
			Confidence:    0.8,
		},
		{
			ID:            "rel-3",
			SourceProject: "proj-2",
			TargetProject: "proj-3",
			Type:          "Evolution",
			Description:   "Project 3 evolved from 2",
			Confidence:    0.7,
		},
	}

	for _, rel := range relationships {
		if err := store.relationships.CreateRelationship(rel); err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
	}

	if err := store.relationships.DeleteAllForProject("proj-1"); err != nil {
		t.Fatalf("failed to delete relationships for project 1: %v", err)
	}

	project1Rels, err := store.relationships.ListRelationships("proj-1")
	if err != nil {
		t.Fatalf("failed to list relationships for project 1 after deletion: %v", err)
	}

	if len(project1Rels) != 0 {
		t.Errorf("expected 0 relationships for project 1 after deletion, got %d", len(project1Rels))
	}

	project2Rels, err := store.relationships.ListRelationships("proj-2")
	if err != nil {
		t.Fatalf("failed to list relationships for project 2 after deletion: %v", err)
	}

	if len(project2Rels) != 1 {
		t.Errorf("expected 1 relationship for project 2 after deletion of project 1, got %d", len(project2Rels))
	}
}