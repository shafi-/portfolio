package mcp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"project-dash/internal/store"
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
				mcp.WithString("implementation_status", mcp.Description("planned|partial|complete|mature|deprecated")),
				mcp.WithString("feature_architecture", mcp.Description("How the feature is implemented")),
				mcp.WithString("pattern", mcp.Description("Architectural pattern(s) used")),
			),
			Handler: s.handleStoreFeature,
		},
		{
			Tool: mcp.NewTool("listFeatures",
				mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID")),
			),
			Handler: s.handleListFeatures,
		},
		{
			Tool: mcp.NewTool("searchFeatures",
				mcp.WithString("project_id", mcp.Description("Filter by project ID")),
				mcp.WithString("query", mcp.Description("Search across name, description, architecture")),
				mcp.WithString("implementation_status", mcp.Description("planned|partial|complete|mature|deprecated")),
				mcp.WithString("pattern", mcp.Description("Filter by architectural pattern")),
			),
			Handler: s.handleSearchFeatures,
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
	confProvided := false
	if cf, ok := args["confidence"].(float64); ok {
		confidence = cf
		confProvided = true
	}

	desc, _ := args["description"].(string)
	implStatus, _ := args["implementation_status"].(string)
	featArch, _ := args["feature_architecture"].(string)
	pattern, _ := args["pattern"].(string)

	// Upsert by (analysis, name). A Tier-3 deep-dive calls storeFeature again
	// with the same project/analyzer/name plus the Tier-3 fields. We look up the
	// existing feature and overlay ONLY the fields the caller supplied, so an
	// enrich call never blanks the Tier-2 facts already stored (empty string =
	// "leave as-is"; confidence is only applied when explicitly passed). Absent a
	// matching feature, we create one.
	existing, err := s.features.GetByAnalysisAndName(analysisID, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to look up feature: %v", err)), nil
	}

	if existing != nil {
		merged := *existing
		if desc != "" {
			merged.Description = desc
		}
		if confProvided {
			merged.Confidence = confidence
		}
		if implStatus != "" {
			merged.ImplementationStatus = implStatus
		}
		if featArch != "" {
			merged.FeatureArchitecture = featArch
		}
		if pattern != "" {
			merged.Pattern = pattern
		}
		if err := s.features.UpdateFeature(&merged); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update feature: %v", err)), nil
		}
		return mcp.NewToolResultJSON(map[string]interface{}{
			"id":          merged.ID,
			"analysis_id": analysisID,
			"name":        merged.Name,
			"updated":     true,
		})
	}

	feature := &models.Feature{
		ID:                   uuid.New().String(),
		AnalysisID:           analysisID,
		Name:                 name,
		Description:          desc,
		Confidence:           confidence,
		ImplementationStatus: implStatus,
		FeatureArchitecture:  featArch,
		Pattern:              pattern,
	}

	if err := s.features.CreateFeature(feature); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to store feature: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"id":          feature.ID,
		"analysis_id": analysisID,
		"name":        feature.Name,
		"created":     true,
	})
}

func (s *Server) handleListFeatures(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	projectID, _ := args["project_id"].(string)
	if projectID == "" {
		return mcp.NewToolResultError("project_id is required"), nil
	}

	features, err := s.features.ListByProject(projectID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list features: %v", err)), nil
	}

	if features == nil {
		features = []*models.Feature{}
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"project_id": projectID,
		"features":   features,
		"count":      len(features),
	})
}

func (s *Server) handleSearchFeatures(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	opts := store.FeatureSearchOptions{}

	if pid, _ := args["project_id"].(string); pid != "" {
		opts.ProjectID = pid
	}
	if st, _ := args["implementation_status"].(string); st != "" {
		opts.ImplementationStatus = st
	}
	if pat, _ := args["pattern"].(string); pat != "" {
		opts.Pattern = pat
	}
	if q, _ := args["query"].(string); q != "" {
		opts.Query = q
	}

	features, err := s.features.SearchFeatures(opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	if features == nil {
		features = []*models.Feature{}
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"features": features,
		"count":    len(features),
	})
}
