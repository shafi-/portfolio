package mcp

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/analysis"
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
				mcp.WithString("type", mcp.Required(), mcp.Description("Relationship type")),
				mcp.WithString("description", mcp.Description("Relationship description")),
				mcp.WithNumber("confidence", mcp.Description("Confidence score (0-1)")),
			),
			Handler: s.handleStoreRelationship,
		},
		{
			Tool: mcp.NewTool("deleteRelationship",
				mcp.WithString("relationship_id", mcp.Required(), mcp.Description("Relationship ID")),
			),
			Handler: s.handleDeleteRelationship,
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

	return mcp.NewToolResultJSON(projects)
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

	rows, err := s.db.Query(
		"SELECT id, name, root_path, repository_type, discovered_at, updated_at FROM projects WHERE name LIKE ? ORDER BY name LIMIT 50",
		"%"+query+"%",
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

	now := time.Now().UTC()

	// Validate analysis schema using JSON schema
	validator, err := analysis.NewSchemaValidator()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to create validator", err), nil
	}
	
	// Convert args to AnalysisInput for validation
	analysisInput := analysis.AnalysisInput{
		Summary:         getStringArg(args, "summary"),
		Purpose:         getStringArg(args, "purpose"),
		Architecture:    getStringArg(args, "architecture"),
		Maturity:        getStringArg(args, "maturity"),
		Notes:           getStringArg(args, "notes"),
		Analyzer:        analyzer,
		AnalyzedGitHead: gitHead,
		AnalyzedAt:      now,
	}
	
	if err := validator.Validate(analysisInput); err != nil {
		return mcp.NewToolResultErrorFromErr("analysis validation failed", err), nil
	}

	analysis := &models.Analysis{
		ID:              uuid.New().String(),
		ProjectID:       projectID,
		Analyzer:        analyzer,
		AnalyzedGitHead: gitHead,
		AnalyzedAt:      now,
		Summary:         getStringArg(args, "summary"),
		Purpose:         getStringArg(args, "purpose"),
		Architecture:    getStringArg(args, "architecture"),
		Notes:           getStringArg(args, "notes"),
		RawJSON:         []byte(getStringArg(args, "raw_json")),
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

	var needing []*models.Project
	for _, p := range projects {
		analyses, err := s.analyses.ListAnalyses(p.ID)
		if err != nil || len(analyses) == 0 {
			needing = append(needing, p)
			continue
		}

		meta, err := s.metadata.GetMetadata(p.ID)
		if err != nil || meta == nil {
			needing = append(needing, p)
			continue
		}

		latest := analyses[0]
		if meta.GitHead != "" && latest.AnalyzedGitHead != meta.GitHead {
			needing = append(needing, p)
		}
	}

	result := map[string]interface{}{
		"projects": needing,
		"count":    len(needing),
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

	return mcp.NewToolResultJSON(relationships)
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

func (s *Server) handleStoreRelationship(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	sourceID := getStringArg(args, "source_project")
	targetID := getStringArg(args, "target_project")
	relType := getStringArg(args, "type")
	description := getStringArg(args, "description")

	if sourceID == "" || targetID == "" || relType == "" {
		return mcp.NewToolResultError("source_project, target_project, and type are required"), nil
	}

	// Validate relationship type
	allowedTypes := []string{"Similar", "Evolution", "Shared Feature", "Shared Technology", "Reuses Component"}
	typeValid := false
	for _, t := range allowedTypes {
		if relType == t {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return mcp.NewToolResultError("invalid relationship type"), nil
	}

	var confidence *float64
	if confNum, ok := args["confidence"].(float64); ok {
		conf := float64(confNum)
		if conf < 0 || conf > 1 {
			return mcp.NewToolResultError("confidence must be between 0 and 1"), nil
		}
		confidence = &conf
	}

	rel := &models.Relationship{
		ID:            uuid.New().String(),
		SourceProject: sourceID,
		TargetProject: targetID,
		Type:          relType,
		Description:   description,
		Confidence:    confidence,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.relationships.CreateRelationship(rel); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to store relationship", err), nil
	}

	result := map[string]interface{}{
		"id":             rel.ID,
		"source_project": rel.SourceProject,
		"target_project": rel.TargetProject,
		"type":           rel.Type,
		"created_at":     rel.CreatedAt,
	}
	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleDeleteRelationship(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	relID := getStringArg(args, "relationship_id")
	if relID == "" {
		return mcp.NewToolResultError("relationship_id is required"), nil
	}

	if err := s.relationships.DeleteRelationship(relID); err != nil {
		return mcp.NewToolResultErrorFromErr("failed to delete relationship", err), nil
	}

	result := map[string]interface{}{
		"deleted": true,
		"id":      relID,
	}
	return mcp.NewToolResultJSON(result)
}
