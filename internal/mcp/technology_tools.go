package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/pkg/models"
)

func (s *Server) technologyTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("storeTechnology",
				mcp.WithString("name", mcp.Required(), mcp.Description("Technology name (e.g. Flutter, Supabase)")),
				mcp.WithString("category", mcp.Description("Category (e.g. framework, database, language)")),
			),
			Handler: s.handleStoreTechnology,
		},
		{
			Tool: mcp.NewTool("tagProjectWithTechnology",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("technology_name", mcp.Required(), mcp.Description("Technology name")),
				mcp.WithString("category", mcp.Description("Technology category (created if tech doesn't exist)")),
			),
			Handler: s.handleTagProjectWithTechnology,
		},
		{
			Tool:    mcp.NewTool("listTechnologies"),
			Handler: s.handleListTechnologies,
		},
		{
			Tool: mcp.NewTool("listProjectTechnologies",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleListProjectTechnologies,
		},
		{
			Tool: mcp.NewTool("searchByTechnology",
				mcp.WithString("technology_name", mcp.Required(), mcp.Description("Technology name to search for")),
			),
			Handler: s.handleSearchByTechnology,
		},
	}
}

func (s *Server) handleStoreTechnology(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	name, _ := args["name"].(string)
	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}
	category, _ := args["category"].(string)

	existing, err := s.technologies.GetByName(name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to check technology: %v", err)), nil
	}
	if existing != nil {
		return mcp.NewToolResultJSON(map[string]interface{}{
			"id":       existing.ID,
			"name":     existing.Name,
			"category": existing.Category,
			"existing": true,
		})
	}

	tech := &models.Technology{
		ID:       uuid.New().String(),
		Name:     name,
		Category: category,
	}
	if err := s.technologies.CreateTechnology(tech); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to store technology: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"id":       tech.ID,
		"name":     tech.Name,
		"category": tech.Category,
		"existing": false,
	})
}

func (s *Server) handleTagProjectWithTechnology(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	techName, _ := args["technology_name"].(string)
	if projectID == "" || techName == "" {
		return mcp.NewToolResultError("project_id and technology_name are required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	tech, err := s.technologies.GetByName(techName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to look up technology: %v", err)), nil
	}
	if tech == nil {
		category, _ := args["category"].(string)
		tech = &models.Technology{
			ID:       uuid.New().String(),
			Name:     techName,
			Category: category,
		}
		if err := s.technologies.CreateTechnology(tech); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to create technology: %v", err)), nil
		}
	}

	pt := models.ProjectTechnology{
		ProjectID:    projectID,
		TechnologyID: tech.ID,
	}
	if err := s.technologies.AddProjectTechnology(pt); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to tag project: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"project_id": projectID,
		"technology": tech,
		"status":     "tagged",
	})
}

func (s *Server) handleListTechnologies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	technologies, err := s.technologies.ListTechnologies()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list technologies: %v", err)), nil
	}

	if technologies == nil {
		technologies = []*models.Technology{}
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"technologies": technologies,
		"count":        len(technologies),
	})
}

func (s *Server) handleListProjectTechnologies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	pts, err := s.technologies.ListProjectTechnologies(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list project technologies: %v", err)), nil
	}

	var result []map[string]interface{}
	for _, pt := range pts {
		tech, err := s.technologies.GetTechnology(pt.TechnologyID)
		if err != nil || tech == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":       tech.ID,
			"name":     tech.Name,
			"category": tech.Category,
		})
	}

	if result == nil {
		result = []map[string]interface{}{}
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"project_id":   projectID,
		"technologies": result,
		"count":        len(result),
	})
}

func (s *Server) handleSearchByTechnology(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	techName, _ := args["technology_name"].(string)
	if techName == "" {
		return mcp.NewToolResultError("technology_name is required"), nil
	}

	tech, err := s.technologies.GetByName(techName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to look up technology: %v", err)), nil
	}
	if tech == nil {
		return mcp.NewToolResultJSON(map[string]interface{}{
			"query":   techName,
			"count":   0,
			"results": []interface{}{},
		})
	}

	rows, err := s.db.Query(
		`SELECT p.id, p.name, p.root_path, p.repository_type
		 FROM projects p
		 JOIN project_technologies pt ON pt.project_id = p.id
		 WHERE pt.technology_id = ?
		 ORDER BY p.name`, tech.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}
	defer rows.Close()

	var results []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.RootPath, &p.RepositoryType); err != nil {
			continue
		}
		results = append(results, p)
	}

	if results == nil {
		results = []*models.Project{}
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"query":      techName,
		"technology": tech,
		"results":    results,
		"count":      len(results),
	})
}
