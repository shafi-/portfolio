package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/pkg/models"
)

func (s *Server) featureTools() []serverTool {
	return []serverTool{
		{
			Tool: mcp.NewTool("storeFeature",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
				mcp.WithString("analyzer", mcp.Required(), mcp.Description("Analyzer name")),
				mcp.WithString("name", mcp.Required(), mcp.Description("Feature name")),
				mcp.WithString("description", mcp.Description("Feature description")),
				mcp.WithNumber("confidence", mcp.Description("Confidence score (0-1)")),
			),
			Handler: s.handleStoreFeature,
		},
		{
			Tool: mcp.NewTool("listFeatures",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleListFeatures,
		},
	}
}

func (s *Server) handleStoreFeature(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	analyzer, _ := args["analyzer"].(string)
	name, _ := args["name"].(string)

	if projectID == "" || analyzer == "" || name == "" {
		return mcp.NewToolResultError("project_id, analyzer, and name are required"), nil
	}

	project, err := s.projects.GetProject(projectID)
	if err != nil || project == nil {
		return mcp.NewToolResultError("project not found"), nil
	}

	analyses, err := s.analyses.ListAnalyses(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get analyses: %v", err)), nil
	}

	var analysisID string
	for _, a := range analyses {
		if a.Analyzer == analyzer {
			analysisID = a.ID
			break
		}
	}
	if analysisID == "" {
		return mcp.NewToolResultError(fmt.Sprintf("no analysis found for analyzer: %s", analyzer)), nil
	}

	confidence := 0.8
	if cf, ok := args["confidence"].(float64); ok {
		confidence = cf
	}

	desc, _ := args["description"].(string)

	feature := &models.Feature{
		ID:          uuid.New().String(),
		AnalysisID:  analysisID,
		Name:        name,
		Description: desc,
		Confidence:  confidence,
	}

	if err := s.features.CreateFeature(feature); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to store feature: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"id":          feature.ID,
		"analysis_id": analysisID,
		"name":        feature.Name,
	})
}

func (s *Server) handleListFeatures(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	analyses, err := s.analyses.ListAnalyses(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get analyses: %v", err)), nil
	}

	var allFeatures []*models.Feature
	for _, a := range analyses {
		features, err := s.features.ListByAnalysis(a.ID)
		if err != nil {
			continue
		}
		allFeatures = append(allFeatures, features...)
	}

	if allFeatures == nil {
		allFeatures = []*models.Feature{}
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"project_id": projectID,
		"features":   allFeatures,
		"count":      len(allFeatures),
	})
}
