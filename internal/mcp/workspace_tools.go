package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/analysis"
)

// workspaceTools exposes read access to workspaces (ADR-023). Management
// stays in the CLI; agents get enough to check which services' analyses are
// current before relying on them.
func (s *Server) workspaceTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("listWorkspaces",
				mcp.WithDescription("List all workspaces (named groups of projects, e.g. the microservices of one product) with member counts."),
			),
			Handler: s.handleListWorkspaces,
		},
		{
			Tool: mcp.NewTool("getWorkspace",
				mcp.WithDescription("Show one workspace and its member projects, each with analysis freshness (fresh/stale/none) computed against the repository's live HEAD. Use this to check which services' analyses are up to date before relying on them."),
				mcp.WithString("workspace", mcp.Required(), mcp.Description("Workspace name or ID")),
			),
			Handler: s.handleGetWorkspace,
		},
	}
}

func (s *Server) handleListWorkspaces(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workspaces, err := s.workspaces.ListWorkspaces()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list workspaces", err), nil
	}

	items := make([]map[string]interface{}, 0, len(workspaces))
	for _, w := range workspaces {
		count, err := s.workspaces.CountWorkspaceProjects(w.ID)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to count workspace projects", err), nil
		}
		items = append(items, map[string]interface{}{
			"id":           w.ID,
			"name":         w.Name,
			"description":  w.Description,
			"member_count": count,
			"created_at":   w.CreatedAt,
		})
	}

	result := map[string]interface{}{
		"workspaces": items,
		"count":      len(items),
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleGetWorkspace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	workspaceRef, _ := args["workspace"].(string)
	if workspaceRef == "" {
		return mcp.NewToolResultError("workspace is required"), nil
	}

	workspace, err := s.resolveWorkspace(workspaceRef)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to resolve workspace", err), nil
	}
	if workspace == nil {
		return mcp.NewToolResultError("workspace not found: " + workspaceRef), nil
	}

	members, err := s.workspaces.ListWorkspaceProjects(workspace.ID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list workspace projects", err), nil
	}

	memberViews := make([]map[string]interface{}, 0, len(members))
	freshCount, staleCount, noAnalysisCount := 0, 0, 0
	for _, p := range members {
		analyses, err := s.analyses.ListAnalyses(p.ID)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to get analyses", err), nil
		}
		freshness := s.freshnessForProject(p.ID, analyses)

		switch {
		case freshness.Status == analysis.StatusFresh:
			freshCount++
		case freshness.Status == analysis.StatusStale:
			staleCount++
		default:
			noAnalysisCount++
		}

		memberViews = append(memberViews, map[string]interface{}{
			"id":              p.ID,
			"name":            p.Name,
			"root_path":       p.RootPath,
			"repository_type": p.RepositoryType,
			"analysis":        freshness,
		})
	}

	result := map[string]interface{}{
		"id":          workspace.ID,
		"name":        workspace.Name,
		"description": workspace.Description,
		"created_at":  workspace.CreatedAt,
		"updated_at":  workspace.UpdatedAt,
		"members":     memberViews,
		"counts": map[string]interface{}{
			"projects":    len(memberViews),
			"fresh":       freshCount,
			"stale":       staleCount,
			"no_analysis": noAnalysisCount,
		},
	}
	return mcp.NewToolResultJSON(result)
}
