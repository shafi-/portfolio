package store

import (
	"errors"
	"testing"
	"time"

	"project-dash/pkg/models"
)

func createTestWorkspaceProject(t *testing.T, s *testStore, id, name string) *models.Project {
	t.Helper()
	p := &models.Project{
		ID:             id,
		Name:           name,
		RootPath:       "/test/" + id,
		RepositoryType: "git",
		DiscoveredAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:      "2024-01-01T00:00:00Z",
	}
	if err := s.projects.UpsertProject(p); err != nil {
		t.Fatalf("failed to create project %s: %v", id, err)
	}
	return p
}

func TestWorkspaceStore_CreateAndGet(t *testing.T) {
	s := setupTestStore(t)
	defer cleanupTestStore(t, s)

	w := &models.Workspace{
		ID:          "ws-1",
		Name:        "commerce",
		Description: "The commerce product services",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
	}
	if err := s.workspaces.CreateWorkspace(w); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	byID, err := s.workspaces.GetWorkspace("ws-1")
	if err != nil {
		t.Fatalf("GetWorkspace: %v", err)
	}
	if byID == nil || byID.Name != "commerce" {
		t.Fatalf("expected workspace by ID, got %+v", byID)
	}

	byName, err := s.workspaces.GetWorkspaceByName("commerce")
	if err != nil {
		t.Fatalf("GetWorkspaceByName: %v", err)
	}
	if byName == nil || byName.ID != "ws-1" {
		t.Fatalf("expected workspace by name, got %+v", byName)
	}

	missing, err := s.workspaces.GetWorkspaceByName("nope")
	if err != nil {
		t.Fatalf("GetWorkspaceByName(missing): %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing workspace, got %+v", missing)
	}
}

func TestWorkspaceStore_CreateDuplicateName(t *testing.T) {
	s := setupTestStore(t)
	defer cleanupTestStore(t, s)

	now := "2024-01-01T00:00:00Z"
	if err := s.workspaces.CreateWorkspace(&models.Workspace{ID: "ws-1", Name: "commerce", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("first CreateWorkspace: %v", err)
	}

	err := s.workspaces.CreateWorkspace(&models.Workspace{ID: "ws-2", Name: "commerce", CreatedAt: now, UpdatedAt: now})
	if !errors.Is(err, ErrWorkspaceNameTaken) {
		t.Fatalf("expected ErrWorkspaceNameTaken, got %v", err)
	}
}

func TestWorkspaceStore_Membership(t *testing.T) {
	s := setupTestStore(t)
	defer cleanupTestStore(t, s)

	createTestWorkspaceProject(t, s, "proj-1", "orders-service")
	createTestWorkspaceProject(t, s, "proj-2", "payments-service")
	createTestWorkspaceProject(t, s, "proj-3", "unrelated")

	now := "2024-01-01T00:00:00Z"
	if err := s.workspaces.CreateWorkspace(&models.Workspace{ID: "ws-1", Name: "commerce", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Adding a project twice is a no-op, not an error.
	for i := 0; i < 2; i++ {
		if err := s.workspaces.AddProjectToWorkspace("ws-1", "proj-1"); err != nil {
			t.Fatalf("AddProjectToWorkspace (pass %d): %v", i+1, err)
		}
	}
	if err := s.workspaces.AddProjectToWorkspace("ws-1", "proj-2"); err != nil {
		t.Fatalf("AddProjectToWorkspace: %v", err)
	}

	in, err := s.workspaces.IsProjectInWorkspace("ws-1", "proj-1")
	if err != nil || !in {
		t.Fatalf("IsProjectInWorkspace(proj-1) = %v, %v; want true, nil", in, err)
	}

	members, err := s.workspaces.ListWorkspaceProjects("ws-1")
	if err != nil {
		t.Fatalf("ListWorkspaceProjects: %v", err)
	}
	if len(members) != 2 || members[0].Name != "orders-service" || members[1].Name != "payments-service" {
		t.Fatalf("unexpected members: %+v", members)
	}

	count, err := s.workspaces.CountWorkspaceProjects("ws-1")
	if err != nil || count != 2 {
		t.Fatalf("CountWorkspaceProjects = %d, %v; want 2, nil", count, err)
	}

	if err := s.workspaces.RemoveProjectFromWorkspace("ws-1", "proj-2"); err != nil {
		t.Fatalf("RemoveProjectFromWorkspace: %v", err)
	}
	in, err = s.workspaces.IsProjectInWorkspace("ws-1", "proj-2")
	if err != nil || in {
		t.Fatalf("IsProjectInWorkspace(proj-2) after remove = %v, %v; want false, nil", in, err)
	}
}

func TestWorkspaceStore_DeleteCascadesMembershipNotProjects(t *testing.T) {
	s := setupTestStore(t)
	defer cleanupTestStore(t, s)

	createTestWorkspaceProject(t, s, "proj-1", "orders-service")

	now := "2024-01-01T00:00:00Z"
	if err := s.workspaces.CreateWorkspace(&models.Workspace{ID: "ws-1", Name: "commerce", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	if err := s.workspaces.AddProjectToWorkspace("ws-1", "proj-1"); err != nil {
		t.Fatalf("AddProjectToWorkspace: %v", err)
	}

	if err := s.workspaces.DeleteWorkspace("ws-1"); err != nil {
		t.Fatalf("DeleteWorkspace: %v", err)
	}

	// Membership rows cascade, the project itself must survive.
	project, err := s.projects.GetProject("proj-1")
	if err != nil {
		t.Fatalf("GetProject after workspace delete: %v", err)
	}
	if project == nil {
		t.Fatal("project was deleted along with its workspace; grouping must never mutate projects")
	}

	deleted, err := s.workspaces.GetWorkspace("ws-1")
	if err != nil || deleted != nil {
		t.Fatalf("GetWorkspace after delete = %+v, %v; want nil, nil", deleted, err)
	}
}

func TestWorkspaceStore_ListWorkspacesSorted(t *testing.T) {
	s := setupTestStore(t)
	defer cleanupTestStore(t, s)

	now := time.Now().UTC().Format(time.RFC3339)
	for _, name := range []string{"zeta", "alpha", "midway"} {
		if err := s.workspaces.CreateWorkspace(&models.Workspace{ID: "ws-" + name, Name: name, CreatedAt: now, UpdatedAt: now}); err != nil {
			t.Fatalf("CreateWorkspace(%s): %v", name, err)
		}
	}

	workspaces, err := s.workspaces.ListWorkspaces()
	if err != nil {
		t.Fatalf("ListWorkspaces: %v", err)
	}
	if len(workspaces) != 3 {
		t.Fatalf("expected 3 workspaces, got %d", len(workspaces))
	}
	if workspaces[0].Name != "alpha" || workspaces[1].Name != "midway" || workspaces[2].Name != "zeta" {
		t.Fatalf("workspaces not sorted by name: %v", []string{workspaces[0].Name, workspaces[1].Name, workspaces[2].Name})
	}
}
