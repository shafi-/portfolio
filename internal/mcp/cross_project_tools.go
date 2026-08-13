package mcp

import (
	"context"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// crossProjectTools returns tools for analyzing portfolio-level patterns across multiple projects
func (s *Server) crossProjectTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("searchAcrossProjects",
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query for cross-project search")),
				mcp.WithString("project_ids", mcp.Description("Comma-separated list of project IDs to search (optional, searches all if not provided)")),
			),
			Handler: s.handleSearchAcrossProjects,
		},
		{
			Tool: mcp.NewTool("getPortfolioOverview",
				mcp.WithString("group_by", mcp.Description("Group projects by: technology, language, framework, maturity (default: none)")),
			),
			Handler: s.handleGetPortfolioOverview,
		},
		{
			Tool: mcp.NewTool("analyzeTechnologySpread",
				mcp.WithString("technology_name", mcp.Description("Specific technology to analyze (optional)")),
				mcp.WithString("category", mcp.Description("Filter by technology category (optional)")),
			),
			Handler: s.handleAnalyzeTechnologySpread,
		},
		{
			Tool: mcp.NewTool("findProjectDependencies",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID to find dependencies for")),
				mcp.WithString("direction", mcp.Description("Direction: 'incoming' (projects that depend on this), 'outgoing' (dependencies of this), or 'both' (default: 'both')")),
			),
			Handler: s.handleFindProjectDependencies,
		},
		{
			Tool: mcp.NewTool("compareProjects",
				mcp.WithString("project_ids", mcp.Required(), mcp.Description("Comma-separated list of project IDs to compare")),
				mcp.WithString("aspect", mcp.Description("Aspect to compare: technologies, dependencies, complexity, size, maturity (default: all)")),
			),
			Handler: s.handleCompareProjects,
		},
	}
}

// handleSearchAcrossProjects performs cross-project search
func (s *Server) handleSearchAcrossProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	query, _ := args["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	// Get project IDs to search, or search all
	projectIDs := []string{}
	if ids, ok := args["project_ids"].(string); ok && ids != "" {
		projectIDs = strings.Split(ids, ",")
		// Trim whitespace from each ID
		for i, id := range projectIDs {
			projectIDs[i] = strings.TrimSpace(id)
		}
	} else {
		// Get all projects
		projects, err := s.projects.ListProjects()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to get projects", err), nil
		}
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}
	}

	like := "%" + query + "%"
	results := make([]map[string]interface{}, 0)

	// Build query to search across multiple projects
	queryStr := `
		SELECT DISTINCT p.id, p.name, p.root_path, p.repository_type,
		       a.summary, a.purpose, a.architecture,
		       f.name as feature_name, f.description as feature_description,
		       t.name as technology_name, t.category as technology_category
		FROM projects p
		LEFT JOIN analyses a ON a.project_id = p.id
		LEFT JOIN features f ON f.analysis_id = a.id
		LEFT JOIN project_technologies pt ON pt.project_id = p.id
		LEFT JOIN technologies t ON t.id = pt.technology_id
		WHERE p.id IN (`

	// Add placeholders for project IDs
	placeholders := make([]string, len(projectIDs))
	for i := range projectIDs {
		placeholders[i] = "?"
	}
	queryStr += strings.Join(placeholders, ",") + ") AND ("

	// Add search conditions
	conditions := []string{
		"p.name LIKE ?", "a.summary LIKE ?", "a.purpose LIKE ?",
		"a.architecture LIKE ?", "a.notes LIKE ?",
		"f.name LIKE ?", "f.description LIKE ?", "t.name LIKE ?",
	}
	queryStr += strings.Join(conditions, " OR ") + ") ORDER BY p.name"

	// Build arguments
	queryArgs := make([]interface{}, 0)
	queryArgs = append(queryArgs, likeStringsToInterfaces(projectIDs)...)
	for i := 0; i < 8; i++ {
		queryArgs = append(queryArgs, like)
	}

	rows, err := s.db.Query(queryStr, queryArgs...)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("cross-project search failed", err), nil
	}
	defer rows.Close()

	for rows.Next() {
		var pID, pName, pRoot, pRepoType string
		var summary, purpose, architecture, featName, featDesc, techName, techCategory interface{}

		err := rows.Scan(&pID, &pName, &pRoot, &pRepoType, &summary, &purpose, &architecture, &featName, &featDesc, &techName, &techCategory)
		if err != nil {
			continue
		}

		result := map[string]interface{}{
			"project_id":      pID,
			"project_name":    pName,
			"root_path":       pRoot,
			"repository_type": pRepoType,
		}

		if summary != nil {
			result["summary"] = summary
		}
		if purpose != nil {
			result["purpose"] = purpose
		}
		if architecture != nil {
			result["architecture"] = architecture
		}
		if featName != nil {
			result["feature_name"] = featName
		}
		if featDesc != nil {
			result["feature_description"] = featDesc
		}
		if techName != nil {
			result["technology_name"] = techName
		}
		if techCategory != nil {
			result["technology_category"] = techCategory
		}

		results = append(results, result)
	}

	result := map[string]interface{}{
		"results":           results,
		"query":             query,
		"projects_searched": len(projectIDs),
		"count":             len(results),
	}
	return mcp.NewToolResultJSON(result)
}

// handleGetPortfolioOverview provides portfolio-level overview
func (s *Server) handleGetPortfolioOverview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	groupBy, _ := args["group_by"].(string)

	projects, err := s.projects.ListProjects()
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to get projects", err), nil
	}

	overview := map[string]interface{}{
		"total_projects": len(projects),
		"projects":       projects,
	}

	// Group by specified criteria
	switch groupBy {
	case "technology":
		groups, err := s.groupProjectsByTechnology()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to group by technology", err), nil
		}
		overview["technology_groups"] = groups

	case "language":
		groups, err := s.groupProjectsByLanguage()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to group by language", err), nil
		}
		overview["language_groups"] = groups

	case "framework":
		groups, err := s.groupProjectsByFramework()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to group by framework", err), nil
		}
		overview["framework_groups"] = groups

	case "maturity":
		groups, err := s.groupProjectsByMaturity()
		if err != nil {
			return mcp.NewToolResultErrorFromErr("failed to group by maturity", err), nil
		}
		overview["maturity_groups"] = groups
	}

	return mcp.NewToolResultJSON(overview)
}

// handleAnalyzeTechnologySpread analyzes technology usage across portfolio
func (s *Server) handleAnalyzeTechnologySpread(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	techName, _ := args["technology_name"].(string)
	category, _ := args["category"].(string)

	query := `
		SELECT t.name, t.category, COUNT(DISTINCT pt.project_id) as project_count,
		       GROUP_CONCAT(p.name) as project_names
		FROM technologies t
		LEFT JOIN project_technologies pt ON pt.technology_id = t.id
		LEFT JOIN projects p ON p.id = pt.project_id
		WHERE 1=1`

	argsList := []interface{}{}

	if techName != "" {
		query += " AND t.name LIKE ?"
		argsList = append(argsList, "%"+techName+"%")
	}
	if category != "" {
		query += " AND t.category LIKE ?"
		argsList = append(argsList, "%"+category+"%")
	}

	query += " GROUP BY t.id ORDER BY project_count DESC"

	rows, err := s.db.Query(query, argsList...)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to analyze technology spread", err), nil
	}
	defer rows.Close()

	technologies := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, category string
		var projectCount int
		var projectNames string

		err := rows.Scan(&name, &category, &projectCount, &projectNames)
		if err != nil {
			continue
		}

		tech := map[string]interface{}{
			"name":          name,
			"category":      category,
			"project_count": projectCount,
		}

		if projectNames != "" {
			tech["project_names"] = strings.Split(projectNames, ",")
		}

		technologies = append(technologies, tech)
	}

	result := map[string]interface{}{
		"technologies": technologies,
		"count":        len(technologies),
	}

	if techName != "" {
		result["filter"] = map[string]interface{}{
			"technology_name": techName,
		}
	}
	if category != "" {
		if result["filter"] == nil {
			result["filter"] = map[string]interface{}{}
		}
		result["filter"].(map[string]interface{})["category"] = category
	}

	return mcp.NewToolResultJSON(result)
}

// handleFindProjectDependencies finds dependencies between projects
func (s *Server) handleFindProjectDependencies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	direction, _ := args["direction"].(string)
	if direction == "" {
		direction = "both"
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	result := map[string]interface{}{
		"project_id":   projectID,
		"project_name": project.Name,
		"incoming":     []map[string]interface{}{},
		"outgoing":     []map[string]interface{}{},
	}

	// Get existing relationships
	relationships, err := s.relationships.ListRelationships(projectID)
	if err == nil {
		for _, rel := range relationships {
			if direction == "both" || direction == "incoming" {
				if rel.TargetProject == projectID {
					sourceProj, _ := s.projects.GetProject(rel.SourceProject)
					if sourceProj != nil {
						incoming := map[string]interface{}{
							"source_project_id":   rel.SourceProject,
							"source_project_name": sourceProj.Name,
							"relationship_type":   rel.Type,
							"description":         rel.Description,
							"confidence":          rel.Confidence,
						}
						result["incoming"] = append(result["incoming"].([]map[string]interface{}), incoming)
					}
				}
			}
			if direction == "both" || direction == "outgoing" {
				if rel.SourceProject == projectID {
					targetProj, _ := s.projects.GetProject(rel.TargetProject)
					if targetProj != nil {
						outgoing := map[string]interface{}{
							"target_project_id":   rel.TargetProject,
							"target_project_name": targetProj.Name,
							"relationship_type":   rel.Type,
							"description":         rel.Description,
							"confidence":          rel.Confidence,
						}
						result["outgoing"] = append(result["outgoing"].([]map[string]interface{}), outgoing)
					}
				}
			}
		}
	}

	// Analyze shared dependencies for indirect relationships
	if direction == "both" || direction == "outgoing" {
		sharedDeps, err := s.findProjectsBySharedDependencies(projectID)
		if err == nil && len(sharedDeps) > 0 {
			result["potential_dependencies"] = sharedDeps
		}
	}

	return mcp.NewToolResultJSON(result)
}

// handleCompareProjects compares multiple projects
func (s *Server) handleCompareProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectIDsStr, _ := args["project_ids"].(string)
	if projectIDsStr == "" {
		return mcp.NewToolResultError("project_ids is required"), nil
	}

	aspect, _ := args["aspect"].(string)
	if aspect == "" {
		aspect = "all"
	}

	projectIDs := strings.Split(projectIDsStr, ",")
	for i, id := range projectIDs {
		projectIDs[i] = strings.TrimSpace(id)
	}

	comparison := map[string]interface{}{
		"project_ids": projectIDs,
		"aspect":      aspect,
		"projects":    make([]map[string]interface{}, 0),
	}

	for _, projectID := range projectIDs {
		project, err := s.projects.GetProject(projectID)
		if err != nil || project == nil {
			continue
		}

		projectData := map[string]interface{}{
			"id":              project.ID,
			"name":            project.Name,
			"root_path":       project.RootPath,
			"repository_type": project.RepositoryType,
		}

		// Add metadata based on aspect
		if aspect == "all" || aspect == "technologies" {
			technologies, _ := s.technologies.ListProjectTechnologies(projectID)
			projectData["technologies"] = technologies
		}

		if aspect == "all" || aspect == "dependencies" {
			dependencies, _ := s.dependencies.ListDependencies(projectID)
			projectData["dependencies"] = dependencies
		}

		if aspect == "all" || aspect == "size" {
			metadata, _ := s.metadata.GetMetadata(projectID)
			if metadata != nil {
				projectData["file_count"] = metadata.CommitCount // Using commit count as size indicator
				projectData["language_summary"] = metadata.LanguageSummary
			}
		}

		if aspect == "all" || aspect == "maturity" {
			metadata, _ := s.metadata.GetMetadata(projectID)
			if metadata != nil {
				projectData["maturity_score"] = metadata.MaturityScore
				projectData["maturity_indicators"] = metadata.MaturityIndicators
			}
			analyses, _ := s.analyses.ListAnalyses(projectID)
			if len(analyses) > 0 {
				projectData["analysis_maturity"] = analyses[0].Maturity
			}
		}

		comparison["projects"] = append(comparison["projects"].([]map[string]interface{}), projectData)
	}

	return mcp.NewToolResultJSON(comparison)
}

// Helper functions for cross-project analysis

func (s *Server) groupProjectsByTechnology() (map[string][]map[string]interface{}, error) {
	groups := make(map[string][]map[string]interface{})

	rows, err := s.db.Query(`
		SELECT t.name, p.id, p.name, p.root_path
		FROM technologies t
		JOIN project_technologies pt ON pt.technology_id = t.id
		JOIN projects p ON p.id = pt.project_id
		ORDER BY t.name, p.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var techName, pID, pName, pRoot string
		if err := rows.Scan(&techName, &pID, &pName, &pRoot); err != nil {
			continue
		}

		groups[techName] = append(groups[techName], map[string]interface{}{
			"id":        pID,
			"name":      pName,
			"root_path": pRoot,
		})
	}

	return groups, nil
}

func (s *Server) groupProjectsByLanguage() (map[string][]map[string]interface{}, error) {
	groups := make(map[string][]map[string]interface{})

	projects, err := s.projects.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		metadata, err := s.metadata.GetMetadata(project.ID)
		if err != nil || metadata == nil || metadata.LanguageSummary == "" {
			groups["Unknown"] = append(groups["Unknown"], map[string]interface{}{
				"id":        project.ID,
				"name":      project.Name,
				"root_path": project.RootPath,
			})
			continue
		}

		// Parse language summary (assumes comma-separated)
		languages := strings.Split(metadata.LanguageSummary, ",")
		for _, lang := range languages {
			lang = strings.TrimSpace(lang)
			groups[lang] = append(groups[lang], map[string]interface{}{
				"id":        project.ID,
				"name":      project.Name,
				"root_path": project.RootPath,
			})
		}
	}

	return groups, nil
}

func (s *Server) groupProjectsByFramework() (map[string][]map[string]interface{}, error) {
	groups := make(map[string][]map[string]interface{})

	projects, err := s.projects.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		metadata, err := s.metadata.GetMetadata(project.ID)
		if err != nil || metadata == nil || metadata.FrameworkSummary == "" {
			groups["Unknown"] = append(groups["Unknown"], map[string]interface{}{
				"id":        project.ID,
				"name":      project.Name,
				"root_path": project.RootPath,
			})
			continue
		}

		// Parse framework summary (assumes comma-separated)
		frameworks := strings.Split(metadata.FrameworkSummary, ",")
		for _, fw := range frameworks {
			fw = strings.TrimSpace(fw)
			groups[fw] = append(groups[fw], map[string]interface{}{
				"id":        project.ID,
				"name":      project.Name,
				"root_path": project.RootPath,
			})
		}
	}

	return groups, nil
}

func (s *Server) groupProjectsByMaturity() (map[string][]map[string]interface{}, error) {
	groups := make(map[string][]map[string]interface{})

	projects, err := s.projects.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		metadata, err := s.metadata.GetMetadata(project.ID)
		if err != nil || metadata == nil {
			groups["Unknown"] = append(groups["Unknown"], map[string]interface{}{
				"id":        project.ID,
				"name":      project.Name,
				"root_path": project.RootPath,
			})
			continue
		}

		// Use maturity score to group
		maturityLevel := "Early"
		if metadata.MaturityScore >= 80 {
			maturityLevel = "Mature"
		} else if metadata.MaturityScore >= 50 {
			maturityLevel = "Developing"
		}

		groups[maturityLevel] = append(groups[maturityLevel], map[string]interface{}{
			"id":             project.ID,
			"name":           project.Name,
			"root_path":      project.RootPath,
			"maturity_score": metadata.MaturityScore,
		})
	}

	return groups, nil
}

func (s *Server) findProjectsBySharedDependencies(projectID string) ([]map[string]interface{}, error) {
	// Get dependencies of the specified project
	deps, err := s.dependencies.ListDependencies(projectID)
	if err != nil || len(deps) == 0 {
		return nil, nil
	}

	// Build a map of this project's dependencies
	depMap := make(map[string]bool)
	for _, dep := range deps {
		key := dep.Name + ":" + dep.Manager
		depMap[key] = true
	}

	// Find other projects that share dependencies
	projects, err := s.projects.ListProjects()
	if err != nil {
		return nil, err
	}

	shares := make([]map[string]interface{}, 0)

	for _, project := range projects {
		if project.ID == projectID {
			continue
		}

		otherDeps, err := s.dependencies.ListDependencies(project.ID)
		if err != nil {
			continue
		}

		sharedCount := 0
		sharedDeps := []string{}

		for _, otherDep := range otherDeps {
			key := otherDep.Name + ":" + otherDep.Manager
			if depMap[key] {
				sharedCount++
				sharedDeps = append(sharedDeps, otherDep.Name)
			}
		}

		if sharedCount > 0 {
			shares = append(shares, map[string]interface{}{
				"project_id":              project.ID,
				"project_name":            project.Name,
				"shared_dependency_count": sharedCount,
				"shared_dependencies":     sharedDeps,
			})
		}
	}

	return shares, nil
}

// Helper function to convert []string to []interface{}
func likeStringsToInterfaces(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}
