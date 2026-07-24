package models

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// AnalysisJSONSchema defines the JSON schema for validation
var AnalysisJSONSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"summary": map[string]interface{}{
			"type":        "string",
			"description": "Brief summary of the project",
		},
		"purpose": map[string]interface{}{
			"type":        "string",
			"description": "Purpose and goals of the project",
		},
		"architecture": map[string]interface{}{
			"type":        "string",
			"description": "Architecture description",
		},
		"maturity": map[string]interface{}{
			"type":        "string",
			"description": "Project maturity level",
		},
		"strengths": map[string]interface{}{
			"type":        "string",
			"description": "Project strengths",
		},
		"weaknesses": map[string]interface{}{
			"type":        "string",
			"description": "Project weaknesses or limitations",
		},
		"reusable_components": map[string]interface{}{
			"type":        "string",
			"description": "Reusable components or patterns",
		},
		"notes": map[string]interface{}{
			"type":        "string",
			"description": "Additional notes",
		},
	},
}

// ValidRelationshipTypes defines allowed relationship types
var ValidRelationshipTypes = []string{
	"Similar",
	"Evolution",
	"Shared Feature",
	"Shared Technology",
	"Reuses Component",
}

// ValidateAnalysis validates an analysis object
func ValidateAnalysis(a *Analysis) error {
	if a.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if a.Analyzer == "" {
		return fmt.Errorf("analyzer is required")
	}
	if a.AnalyzedAt == "" {
		return fmt.Errorf("analyzed_at is required")
	}

	return nil
}

// ValidateRelationship validates a relationship object
func ValidateRelationship(r *Relationship) error {
	if r.SourceProject == "" {
		return fmt.Errorf("source_project is required")
	}
	if r.TargetProject == "" {
		return fmt.Errorf("target_project is required")
	}
	if r.Type == "" {
		return fmt.Errorf("type is required")
	}

	valid := false
	for _, vt := range ValidRelationshipTypes {
		if r.Type == vt {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid relationship type: %s, must be one of: %v", r.Type, ValidRelationshipTypes)
	}

	if r.SourceProject == r.TargetProject {
		return fmt.Errorf("source_project and target_project cannot be the same")
	}

	if r.Confidence < 0 || r.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	return nil
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(uuid string) bool {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, uuid)
	return matched
}

// ValidateRawJSON validates that raw_json contains valid JSON
func ValidateRawJSON(rawJSON string) error {
	if rawJSON == "" {
		return nil
	}

	if !json.Valid([]byte(rawJSON)) {
		return fmt.Errorf("raw_json contains invalid JSON")
	}

	return nil
}

// IsAnalysisOutdated checks if an analysis is outdated compared to current git HEAD
func IsAnalysisOutdated(analysis *Analysis, currentGitHead string) bool {
	if analysis == nil {
		return true
	}
	if currentGitHead == "" {
		return false
	}
	return analysis.AnalyzedGitHead != currentGitHead
}
