package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/discovery"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

func (s *Server) discoveryTools() []serverTool {
	return []serverTool{
		{
			Tool:    mcp.NewTool("health"),
			Handler: s.handleHealth,
		},
		{
			Tool:    mcp.NewTool("discoverProjects"),
			Handler: s.handleDiscoverProjects,
		},
		{
			Tool:    mcp.NewTool("listProjects"),
			Handler: s.handleListProjects,
		},
		{
			Tool: mcp.NewTool("getProject",
				mcp.WithString("id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleGetProject,
		},
	}
}

func (s *Server) searchTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("searchProjects",
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			),
			Handler: s.handleSearchProjects,
		},
		{
			Tool: mcp.NewTool("searchDocumentation",
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			),
			Handler: s.handleSearchDocumentation,
		},
	}
}

func (s *Server) analysisTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("getAnalysis",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleGetAnalysis,
		},
		{
			Tool: mcp.NewTool("storeAnalysis",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("analyzer", mcp.Required(), mcp.Description("Analyzer name")),
				mcp.WithString("analyzed_git_head", mcp.Description("Git HEAD at analysis time")),
				mcp.WithString("summary", mcp.Description("Analysis summary")),
				mcp.WithString("purpose", mcp.Description("Project purpose")),
				mcp.WithString("architecture", mcp.Description("Architecture description")),
				mcp.WithString("maturity", mcp.Description("Project maturity level")),
				mcp.WithString("strengths", mcp.Description("Project strengths")),
				mcp.WithString("weaknesses", mcp.Description("Project weaknesses")),
				mcp.WithString("reusable_components", mcp.Description("Reusable components")),
				mcp.WithString("notes", mcp.Description("Additional notes")),
				mcp.WithString("raw_json", mcp.Description("Raw analysis JSON")),
			),
			Handler: s.handleStoreAnalysis,
		},
		{
			Tool:    mcp.NewTool("listProjectsNeedingAnalysis"),
			Handler: s.handleListProjectsNeedingAnalysis,
		},
	}
}

func (s *Server) configTools() []serverTool {
	return []serverTool{
		{
			Tool:    mcp.NewTool("getConfiguration"),
			Handler: s.handleGetConfiguration,
		},
		{
			Tool: mcp.NewTool("updateConfiguration",
				mcp.WithString("key", mcp.Required(), mcp.Description("Configuration key")),
				mcp.WithString("value", mcp.Required(), mcp.Description("Configuration value")),
			),
			Handler: s.handleUpdateConfiguration,
		},
	}
}

func (s *Server) relationshipTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("listRelationships",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleListRelationships,
		},
		{
			Tool: mcp.NewTool("storeRelationship",
				mcp.WithString("source_project", mcp.Required(), mcp.Description("Source project ID")),
				mcp.WithString("target_project", mcp.Required(), mcp.Description("Target project ID")),
				mcp.WithString("type", mcp.Required(), mcp.Description("Relationship type: Similar, Evolution, Shared Feature, Shared Technology, Reuses Component")),
				mcp.WithString("description", mcp.Description("Description of the relationship")),
				mcp.WithNumber("confidence", mcp.Description("Confidence score (0-1)")),
			),
			Handler: s.handleStoreRelationship,
		},
	}
}

func (s *Server) handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbOK := true
	if err := s.db.Ping(); err != nil {
		dbOK = false
	}

	projectCount := 0
	if dbOK {
		s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
	}

	status := "healthy"
	if !dbOK {
		status = "unhealthy"
	}

	result := map[string]interface{}{
		"status":             status,
		"database_connected": dbOK,
		"project_count":      projectCount,
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleDiscoverProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if len(s.roots) == 0 {
		return mcp.NewToolResultError("no project roots configured"), nil
	}

	discLogger := s.logger.With("mcp-discovery")
	discoverer := discovery.NewDiscoverer(
		s.osFS,
		&rootsConfigProvider{roots: s.roots},
		&discoveryStoreAdapter{store: s.projects},
		discLogger,
		10,
	)

	result, err := discoverer.DiscoverProjects(ctx)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("discovery failed", err), nil
	}

	resultMap := map[string]interface{}{
		"discovered":    result.Discovered,
		"error_count":   len(result.Errors),
		"roots_checked": len(s.roots),
	}
	return mcp.NewToolResultJSON(resultMap)
}

func (s *Server) handleListProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projects, err := s.projects.ListProjects()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list projects", err), nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d projects:\n\n", len(projects)))
	for _, p := range projects {
		output.WriteString(fmt.Sprintf("- %s (%s): %s [%s]\n", p.Name, p.ID, p.RootPath, p.RepositoryType))
	}
	return mcp.NewToolResultText(output.String()), nil
}

func (s *Server) handleGetProject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	project, err := s.projects.GetProject(id)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get project", err), nil
	}
	if project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	meta, err := s.metadata.GetMetadata(id)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get metadata", err), nil
	}

	result := map[string]interface{}{
		"project":  project,
		"metadata": meta,
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleSearchProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	query, _ := args["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	like := "%" + query + "%"

	// Search across projects, analyses, features, and technologies
	rows, err := s.db.Query(`
		SELECT DISTINCT p.id, p.name, p.root_path, p.repository_type, p.discovered_at, p.updated_at
		FROM projects p
		LEFT JOIN analyses a ON a.project_id = p.id
		LEFT JOIN features f ON f.analysis_id = a.id
		LEFT JOIN project_technologies pt ON pt.project_id = p.id
		LEFT JOIN technologies t ON t.id = pt.technology_id
		WHERE p.name LIKE ?
		   OR a.summary LIKE ?
		   OR a.purpose LIKE ?
		   OR a.architecture LIKE ?
		   OR a.notes LIKE ?
		   OR f.name LIKE ?
		   OR f.description LIKE ?
		   OR t.name LIKE ?
		ORDER BY p.name LIMIT 50`,
		like, like, like, like, like, like, like, like,
	)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("search failed", err), nil
	}
	defer rows.Close()

	var results []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.RepositoryType, &p.DiscoveredAt, &p.UpdatedAt); err != nil {
			continue
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return mcp.NewToolResultErrorFromErr("search error", err), nil
	}

	result := map[string]interface{}{
		"results": results,
		"query":   query,
		"count":   len(results),
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleSearchDocumentation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	query, _ := args["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	var results []map[string]interface{}

	docRows, err := s.db.Query(
		`SELECT d.id, d.project_id, d.path, d.kind, SUBSTR(d.content, 1, 500) as content_preview, p.name
		 FROM documents d JOIN projects p ON p.id = d.project_id
		 WHERE d.content LIKE ? ORDER BY d.kind LIMIT 50`,
		"%"+query+"%",
	)
	if err == nil {
		defer docRows.Close()
		for docRows.Next() {
			var id, projectID, path, kind, content, projName string
			if err := docRows.Scan(&id, &projectID, &path, &kind, &content, &projName); err != nil {
				continue
			}
			results = append(results, map[string]interface{}{
				"id":         id,
				"project_id": projectID,
				"project":    projName,
				"path":       path,
				"kind":       kind,
				"content":    content,
			})
		}
	}

	result := map[string]interface{}{
		"results": results,
		"query":   query,
		"count":   len(results),
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleGetAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	analyses, err := s.analyses.ListAnalyses(projectID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get analyses", err), nil
	}

	result := map[string]interface{}{
		"analyses": analyses,
		"count":    len(analyses),
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleStoreAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	projectID, _ := args["project_id"].(string)
	analyzer, _ := args["analyzer"].(string)

	if projectID == "" || analyzer == "" {
		return mcp.NewToolResultError("project_id and analyzer are required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get project", err), nil
	}
	if project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	gitHead, _ := args["analyzed_git_head"].(string)
	if gitHead == "" {
		meta, err := s.metadata.GetMetadata(projectID)
		if err == nil && meta != nil {
			gitHead = meta.GitHead
		}
	}

	rawJSON := getStringArg(args, "raw_json")
	if rawJSON != "" {
		if err := models.ValidateRawJSON(rawJSON); err != nil {
			return mcp.NewToolResultErrorFromErr("invalid raw_json", err), nil
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	analysis := &models.Analysis{
		ID:                 uuid.New().String(),
		ProjectID:          projectID,
		Analyzer:           analyzer,
		AnalyzedGitHead:    gitHead,
		AnalyzedAt:         now,
		Summary:            getStringArg(args, "summary"),
		Purpose:            getStringArg(args, "purpose"),
		Architecture:       getStringArg(args, "architecture"),
		Maturity:           getStringArg(args, "maturity"),
		Strengths:          getStringArg(args, "strengths"),
		Weaknesses:         getStringArg(args, "weaknesses"),
		ReusableComponents: getStringArg(args, "reusable_components"),
		Notes:              getStringArg(args, "notes"),
		RawJSON:            rawJSON,
	}

	if err := models.ValidateAnalysis(analysis); err != nil {
		return mcp.NewToolResultErrorFromErr("validation failed", err), nil
	}

	if err := s.analyses.CreateAnalysis(analysis); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to store analysis", err), nil
	}

	result := map[string]interface{}{
		"id":          analysis.ID,
		"project_id":  analysis.ProjectID,
		"analyzed_at": analysis.AnalyzedAt,
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleListProjectsNeedingAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projects, err := s.projects.ListProjects()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list projects", err), nil
	}

	noAnalysis := make([]map[string]interface{}, 0)
	staleAnalysis := make([]map[string]interface{}, 0)

	for _, p := range projects {
		analyses, err := s.analyses.ListAnalyses(p.ID)
		if err != nil || len(analyses) == 0 {
			noAnalysis = append(noAnalysis, map[string]interface{}{
				"id":   p.ID,
				"name": p.Name,
				"path": p.RootPath,
			})
			continue
		}

		meta, err := s.metadata.GetMetadata(p.ID)
		if err != nil || meta == nil {
			noAnalysis = append(noAnalysis, map[string]interface{}{
				"id":   p.ID,
				"name": p.Name,
				"path": p.RootPath,
			})
			continue
		}

		latest := analyses[0]
		if meta.GitHead != "" && latest.AnalyzedGitHead != meta.GitHead {
			staleAnalysis = append(staleAnalysis, map[string]interface{}{
				"id":                p.ID,
				"name":              p.Name,
				"path":              p.RootPath,
				"analyzed_at":       latest.AnalyzedAt,
				"analyzed_git_head": latest.AnalyzedGitHead,
				"current_git_head":  meta.GitHead,
			})
		}
	}

	result := map[string]interface{}{
		"no_analysis":    noAnalysis,
		"stale_analysis": staleAnalysis,
		"counts": map[string]interface{}{
			"no_analysis":    len(noAnalysis),
			"stale_analysis": len(staleAnalysis),
			"total":          len(noAnalysis) + len(staleAnalysis),
		},
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleGetConfiguration(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	configs, err := s.configuration.List()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get configuration", err), nil
	}

	cfg := make(map[string]string)
	for _, c := range configs {
		cfg[c.Key] = c.Value
	}

	return mcp.NewToolResultJSON(cfg)
}

func (s *Server) handleUpdateConfiguration(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)

	if key == "" {
		return mcp.NewToolResultError("key is required"), nil
	}

	if err := s.configuration.Set(key, value); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to update configuration", err), nil
	}

	result := map[string]interface{}{
		"key":    key,
		"value":  value,
		"status": "updated",
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleListRelationships(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	relationships, err := s.relationships.ListRelationships(projectID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list relationships", err), nil
	}

	// Ensure empty slice is serialized as [], not null
	if relationships == nil {
		relationships = []*models.Relationship{}
	}

	result := map[string]interface{}{
		"relationships": relationships,
		"count":         len(relationships),
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleStoreRelationship(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	sourceProject, _ := args["source_project"].(string)
	targetProject, _ := args["target_project"].(string)
	relType, _ := args["type"].(string)

	if sourceProject == "" || targetProject == "" || relType == "" {
		return mcp.NewToolResultError("source_project, target_project, and type are required"), nil
	}

	sourceProj, err := s.projects.GetProject(sourceProject)
	if err != nil || sourceProj == nil {
		return mcp.NewToolResultError("source_project not found"), nil
	}

	targetProj, err := s.projects.GetProject(targetProject)
	if err != nil || targetProj == nil {
		return mcp.NewToolResultError("target_project not found"), nil
	}

	confidence := 0.5
	if cf, ok := args["confidence"].(float64); ok {
		confidence = cf
	}

	relationship := &models.Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceProject,
		TargetProject: targetProject,
		Type:          relType,
		Description:   getStringArg(args, "description"),
		Confidence:    confidence,
	}

	if err := models.ValidateRelationship(relationship); err != nil {
		return mcp.NewToolResultErrorFromErr("validation failed", err), nil
	}

	if err := s.relationships.CreateRelationship(relationship); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to store relationship", err), nil
	}

	result := map[string]interface{}{
		"id":             relationship.ID,
		"source_project": relationship.SourceProject,
		"target_project": relationship.TargetProject,
		"type":           relationship.Type,
	}
	return mcp.NewToolResultJSON(result)
}

func getStringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

type discoveryStoreAdapter struct {
	store *store.ProjectStore
}

func (a *discoveryStoreAdapter) UpsertProject(p *discovery.Project) error {
	return a.store.UpsertProject(&models.Project{
		ID:             p.ID,
		Name:           p.Name,
		RootPath:       p.RootPath,
		RepositoryType: p.RepositoryType,
		DiscoveredAt:   p.DiscoveredAt.Format(time.RFC3339),
		UpdatedAt:      p.DiscoveredAt.Format(time.RFC3339),
	})
}

type rootsConfigProvider struct {
	roots []string
}

func (r *rootsConfigProvider) GetProjectRoots() ([]string, error) {
	return r.roots, nil
}

func (r *rootsConfigProvider) GetIgnoredPaths() []string {
	return []string{
		"node_modules",
		".git",
		"vendor",
		"build",
		"dist",
		"target",
		"bin",
	}
}
