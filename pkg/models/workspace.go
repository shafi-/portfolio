package models

// Workspace represents a named group of projects — for example the
// microservices that together form one product. Workspaces are grouping-only:
// they scope queries and reporting but never change scan or discovery
// behavior (ADR-023).
type Workspace struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// WorkspaceMembership is the (workspace, project) pair stored in
// workspace_projects. It exists so stores can scan the join table directly.
type WorkspaceMembership struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id"`
}
