package mcp

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/analysis"
	"project-dash/internal/discovery"
	"project-dash/internal/indexer"
	"project-dash/internal/store"
	"project-dash/pkg/models"
)

//go:embed prompts/analysis.md
var analysisPrompt string

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
			Tool: mcp.NewTool("listProjectsNeedingAnalysis",
				mcp.WithString("workspace",
					mcp.Description("Optional workspace name or ID. When given, only the workspace's member projects are checked."),
				),
			),
			Handler: s.handleListProjectsNeedingAnalysis,
		},
		{
			Tool: mcp.NewTool("getProjectAnalyzerPrompt",
				mcp.WithString("project_id",
					mcp.Description("Optional project ID. When given, the prompt includes the project's analysis freshness, the previous analysis as an incremental seed, and instructions to inform the user about staleness before using or refreshing it."),
				),
			),
			Handler: s.handleGetProjectAnalyzerPrompt,
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
	metadataCount := 0
	if dbOK {
		s.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
		s.db.QueryRow("SELECT COUNT(*) FROM metadata").Scan(&metadataCount)
	}

	status := "healthy"
	if !dbOK {
		status = "unhealthy"
	}

	needsScan := dbOK && projectCount > 0 && metadataCount == 0

	result := map[string]interface{}{
		"status":             status,
		"database_connected": dbOK,
		"project_count":      projectCount,
		"metadata_count":     metadataCount,
		"needs_scan":         needsScan,
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

	// Scan all projects after discovery to populate metadata.
	idx := indexer.NewIndexer(s.db, s.logger.Zap()).WithProjectLister(s.projects)
	scanResults, scanErr := idx.IndexAll(ctx)
	scanCount := len(scanResults)
	if scanErr != nil {
		discLogger.Warn("scan after discovery failed", models.Field{Key: "error", Value: scanErr})
	}

	resultMap := map[string]interface{}{
		"discovered":    result.Discovered,
		"error_count":   len(result.Errors),
		"roots_checked": len(s.roots),
		"scanned":       scanCount,
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

	// Freshness gate (ADR-023): the response always states how current the
	// analysis is relative to the repository's live HEAD, so the agent sees
	// staleness BEFORE relying on the analysis.
	freshness := s.freshnessForProject(projectID, analyses)

	result := map[string]interface{}{
		"analyses":  analyses,
		"count":     len(analyses),
		"freshness": freshness,
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
	args := req.GetArguments()
	workspaceRef, _ := args["workspace"].(string)

	projects, err := s.projects.ListProjects()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list projects", err), nil
	}

	// Optional workspace scope (ADR-023) — restrict to the members of one
	// workspace, resolved by name or ID.
	workspaceName := ""
	if workspaceRef != "" {
		workspace, err := s.resolveWorkspace(workspaceRef)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to resolve workspace", err), nil
		}
		if workspace == nil {
			return mcp.NewToolResultError("workspace not found: " + workspaceRef), nil
		}
		workspaceName = workspace.Name

		members, err := s.workspaces.ListWorkspaceProjects(workspace.ID)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to list workspace projects", err), nil
		}
		projects = members
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

		// Live-HEAD freshness (ADR-023): compare the analysis against the
		// repository's current HEAD rather than the last scan's snapshot.
		var storedHead string
		if meta, err := s.metadata.GetMetadata(p.ID); err == nil && meta != nil {
			storedHead = meta.GitHead
		}
		freshness := analysis.FreshnessForLatest(p, storedHead, analyses)
		if freshness.Status != analysis.StatusStale {
			// Fresh, or freshness cannot be proven — not actionable work.
			continue
		}

		staleAnalysis = append(staleAnalysis, map[string]interface{}{
			"id":                p.ID,
			"name":              p.Name,
			"path":              p.RootPath,
			"analyzed_at":       freshness.AnalyzedAt,
			"analyzed_git_head": freshness.AnalyzedGitHead,
			"current_git_head":  freshness.CurrentGitHead,
			"commits_behind":    freshness.CommitsBehind,
		})
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
	if workspaceName != "" {
		result["workspace"] = workspaceName
	}
	return mcp.NewToolResultJSON(result)
}

// resolveWorkspace finds a workspace by name, falling back to ID.
func (s *Server) resolveWorkspace(nameOrID string) (*models.Workspace, error) {
	w, err := s.workspaces.GetWorkspaceByName(nameOrID)
	if err != nil {
		return nil, err
	}
	if w != nil {
		return w, nil
	}
	return s.workspaces.GetWorkspace(nameOrID)
}

func (s *Server) handleGetProjectAnalyzerPrompt(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultText(analysisPrompt), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get project", err), nil
	}
	if project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	analyses, err := s.analyses.ListAnalyses(projectID)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get analyses", err), nil
	}
	freshness := s.freshnessForProject(projectID, analyses)

	return mcp.NewToolResultText(s.buildAnalyzerPrompt(project, analyses, freshness)), nil
}

// buildAnalyzerPrompt assembles the per-project analyzer prompt: a freshness
// banner with the user-consent gate, the previous analysis as an incremental
// seed, and the static analysis instructions. The gate wording is deliberate:
// the agent must tell the user about staleness and only refresh if the user
// asks — otherwise it proceeds with the existing analysis (ADR-023).
func (s *Server) buildAnalyzerPrompt(project *models.Project, analyses []*models.Analysis, freshness analysis.Freshness) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Analysis context for project %q (id: %s)\n\n", project.Name, project.ID)

	b.WriteString("FRESHNESS: ")
	switch freshness.Status {
	case analysis.StatusFresh:
		fmt.Fprintf(&b, "the stored analysis matches the repository HEAD (%s). No refresh needed — reuse it.\n",
			freshness.CurrentGitHead)
	case analysis.StatusStale:
		b.WriteString("STALE — the repository has moved past the stored analysis")
		if freshness.CommitsBehind != nil {
			fmt.Fprintf(&b, " (%d commits ahead of the analyzed HEAD)", *freshness.CommitsBehind)
		}
		fmt.Fprintf(&b, ".\n  Tell the user the stored analysis is stale BEFORE using it, and ask whether to refresh it.\n")
		b.WriteString("  If the user asks for a refresh, produce an updated analysis (incrementally — keep what is still accurate) and store it with `storeAnalysis`.\n")
		b.WriteString("  If the user declines, proceed with the existing analysis and note its age in your answer.\n")
	default:
		fmt.Fprintf(&b, "UNKNOWN — %s.\n  Mention this to the user before relying on the analysis below.\n", freshness.Note)
	}

	b.WriteString("\n")

	if len(analyses) > 0 {
		latest := analyses[0]
		fmt.Fprintf(&b, "PREVIOUS ANALYSIS (analyzer: %s, made: %s, git head: %s)\n",
			latest.Analyzer, latest.AnalyzedAt, latest.AnalyzedGitHead)
		b.WriteString("Update it incrementally — keep sections that are still accurate, change what moved on:\n\n")
		writeAnalysisSection(&b, "SUMMARY", latest.Summary)
		writeAnalysisSection(&b, "PURPOSE", latest.Purpose)
		writeAnalysisSection(&b, "ARCHITECTURE", latest.Architecture)
		writeAnalysisSection(&b, "MATURITY", latest.Maturity)
		writeAnalysisSection(&b, "STRENGTHS", latest.Strengths)
		writeAnalysisSection(&b, "WEAKNESSES", latest.Weaknesses)
		writeAnalysisSection(&b, "REUSABLE COMPONENTS", latest.ReusableComponents)
		writeAnalysisSection(&b, "NOTES", latest.Notes)
	} else {
		b.WriteString("NO PREVIOUS ANALYSIS — produce a full analysis and store it with `storeAnalysis`.\n")
	}

	b.WriteString("\n---\n\n")
	b.WriteString(analysisPrompt)
	return b.String()
}

func writeAnalysisSection(b *strings.Builder, title, content string) {
	if strings.TrimSpace(content) == "" {
		return
	}
	fmt.Fprintf(b, "\n%s:\n%s\n", title, content)
}

// freshnessForProject computes analysis freshness for a project, degrading
// gracefully when the project or its metadata is unavailable.
func (s *Server) freshnessForProject(projectID string, analyses []*models.Analysis) analysis.Freshness {
	var latest *models.Analysis
	if len(analyses) > 0 {
		latest = analyses[0]
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil {
		f := analysis.Compute(analysis.Inputs{Analysis: latest})
		return f
	}
	if project == nil {
		f := analysis.Compute(analysis.Inputs{Analysis: latest})
		f.Note = "project not found"
		return f
	}

	var storedHead string
	if meta, err := s.metadata.GetMetadata(projectID); err == nil && meta != nil {
		storedHead = meta.GitHead
	}
	return analysis.FreshnessForLatest(project, storedHead, analyses)
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
