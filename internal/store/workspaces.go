package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

// ErrWorkspaceNameTaken is returned when creating a workspace whose name is
// already in use. Workspace names are unique (UNIQUE constraint in the v4
// migration) and serve as the human-facing handle in the CLI and MCP tools.
var ErrWorkspaceNameTaken = errors.New("workspace name already in use")

// WorkspaceStore persists workspaces and their project membership
// (migration v4, ADR-023). Workspaces are grouping-only: nothing here ever
// mutates the projects themselves.
type WorkspaceStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewWorkspaceStore(db *sql.DB, logger *zap.Logger) *WorkspaceStore {
	return &WorkspaceStore{db: db, logger: logger}
}

const workspaceColumns = `id, name, description, created_at, updated_at`

func scanWorkspace(row interface{ Scan(...interface{}) error }) (*models.Workspace, error) {
	w := &models.Workspace{}
	if err := row.Scan(&w.ID, &w.Name, &w.Description, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WorkspaceStore) CreateWorkspace(w *models.Workspace) error {
	query := `
		INSERT INTO workspaces (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, w.ID, w.Name, w.Description, w.CreatedAt, w.UpdatedAt)
	if err != nil {
		if isUniqueConstraintError(err) {
			return fmt.Errorf("%w: %s", ErrWorkspaceNameTaken, w.Name)
		}
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	return nil
}

func (s *WorkspaceStore) GetWorkspace(id string) (*models.Workspace, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`SELECT %s FROM workspaces WHERE id = ?`, workspaceColumns), id,
	)
	w, err := scanWorkspace(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	return w, nil
}

func (s *WorkspaceStore) GetWorkspaceByName(name string) (*models.Workspace, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`SELECT %s FROM workspaces WHERE name = ?`, workspaceColumns), name,
	)
	w, err := scanWorkspace(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workspace by name: %w", err)
	}
	return w, nil
}

func (s *WorkspaceStore) ListWorkspaces() ([]*models.Workspace, error) {
	query := fmt.Sprintf(`SELECT %s FROM workspaces ORDER BY name`, workspaceColumns)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		w, err := scanWorkspace(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace row: %w", err)
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, rows.Err()
}

func (s *WorkspaceStore) DeleteWorkspace(id string) error {
	// Membership rows cascade via FK; this must not touch projects.
	_, err := s.db.Exec(`DELETE FROM workspaces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

func (s *WorkspaceStore) AddProjectToWorkspace(workspaceID, projectID string) error {
	query := `
		INSERT INTO workspace_projects (workspace_id, project_id)
		VALUES (?, ?)
	`
	_, err := s.db.Exec(query, workspaceID, projectID)
	if err != nil {
		if isUniqueConstraintError(err) {
			// Already a member — adding twice is a no-op, not an error.
			return nil
		}
		return fmt.Errorf("failed to add project to workspace: %w", err)
	}
	return nil
}

func (s *WorkspaceStore) RemoveProjectFromWorkspace(workspaceID, projectID string) error {
	_, err := s.db.Exec(
		`DELETE FROM workspace_projects WHERE workspace_id = ? AND project_id = ?`,
		workspaceID, projectID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove project from workspace: %w", err)
	}
	return nil
}

// IsProjectInWorkspace reports whether the membership row exists.
func (s *WorkspaceStore) IsProjectInWorkspace(workspaceID, projectID string) (bool, error) {
	var one int
	err := s.db.QueryRow(
		`SELECT 1 FROM workspace_projects WHERE workspace_id = ? AND project_id = ?`,
		workspaceID, projectID,
	).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check workspace membership: %w", err)
	}
	return true, nil
}

// ListWorkspaceProjects returns the member projects of a workspace, ordered by
// name. Projects deleted from the portfolio disappear from workspaces via the
// FK cascade.
func (s *WorkspaceStore) ListWorkspaceProjects(workspaceID string) ([]*models.Project, error) {
	query := `
		SELECT p.id, p.name, p.root_path, p.repository_type, p.discovered_at, p.updated_at
		FROM workspace_projects wp
		JOIN projects p ON p.id = wp.project_id
		WHERE wp.workspace_id = ?
		ORDER BY p.name
	`
	rows, err := s.db.Query(query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspace projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(
			&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workspace project row: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// CountWorkspaceProjects returns the number of members of a workspace.
func (s *WorkspaceStore) CountWorkspaceProjects(workspaceID string) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM workspace_projects WHERE workspace_id = ?`, workspaceID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count workspace projects: %w", err)
	}
	return count, nil
}

// isUniqueConstraintError reports whether err is a SQLite UNIQUE/PRIMARY KEY
// violation. modernc.org/sqlite surfaces this as "UNIQUE constraint failed"
// in the error string (it does not expose sqlite3.Error codes through the
// database/sql error path).
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "constraint failed: UNIQUE")
}
