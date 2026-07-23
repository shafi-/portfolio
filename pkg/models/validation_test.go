package models

import (
	"testing"
)

func TestValidateAnalysis(t *testing.T) {
	tests := []struct {
		name        string
		analysis    *Analysis
		expectError bool
	}{
		{
			name: "valid analysis",
			analysis: &Analysis{
				ProjectID:  "proj-1",
				Analyzer:   "test-analyzer",
				AnalyzedAt: "2024-01-01T12:00:00Z",
			},
			expectError: false,
		},
		{
			name: "missing project_id",
			analysis: &Analysis{
				Analyzer:   "test-analyzer",
				AnalyzedAt: "2024-01-01T12:00:00Z",
			},
			expectError: true,
		},
		{
			name: "missing analyzer",
			analysis: &Analysis{
				ProjectID:  "proj-1",
				AnalyzedAt: "2024-01-01T12:00:00Z",
			},
			expectError: true,
		},
		{
			name: "missing analyzed_at",
			analysis: &Analysis{
				ProjectID: "proj-1",
				Analyzer:  "test-analyzer",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnalysis(tt.analysis)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateAnalysis() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestValidateRelationship(t *testing.T) {
	tests := []struct {
		name         string
		relationship *Relationship
		expectError  bool
	}{
		{
			name: "valid relationship",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Similar",
				Confidence:    0.8,
			},
			expectError: false,
		},
		{
			name: "missing source_project",
			relationship: &Relationship{
				TargetProject: "proj-2",
				Type:          "Similar",
				Confidence:    0.8,
			},
			expectError: true,
		},
		{
			name: "missing target_project",
			relationship: &Relationship{
				SourceProject: "proj-1",
				Type:          "Similar",
				Confidence:    0.8,
			},
			expectError: true,
		},
		{
			name: "missing type",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Confidence:    0.8,
			},
			expectError: true,
		},
		{
			name: "invalid relationship type",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Invalid Type",
				Confidence:    0.8,
			},
			expectError: true,
		},
		{
			name: "same source and target project",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-1",
				Type:          "Similar",
				Confidence:    0.8,
			},
			expectError: true,
		},
		{
			name: "confidence out of range - negative",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Similar",
				Confidence:    -0.1,
			},
			expectError: true,
		},
		{
			name: "confidence out of range - greater than 1",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Similar",
				Confidence:    1.1,
			},
			expectError: true,
		},
		{
			name: "confidence exactly 0",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Similar",
				Confidence:    0,
			},
			expectError: false,
		},
		{
			name: "confidence exactly 1",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Similar",
				Confidence:    1,
			},
			expectError: false,
		},
		{
			name: "valid relationship type - Evolution",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Evolution",
				Confidence:    0.8,
			},
			expectError: false,
		},
		{
			name: "valid relationship type - Shared Feature",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Shared Feature",
				Confidence:    0.8,
			},
			expectError: false,
		},
		{
			name: "valid relationship type - Shared Technology",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Shared Technology",
				Confidence:    0.8,
			},
			expectError: false,
		},
		{
			name: "valid relationship type - Reuses Component",
			relationship: &Relationship{
				SourceProject: "proj-1",
				TargetProject: "proj-2",
				Type:          "Reuses Component",
				Confidence:    0.8,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRelationship(tt.relationship)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateRelationship() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestValidateRawJSON(t *testing.T) {
	tests := []struct {
		name        string
		rawJSON     string
		expectError bool
	}{
		{
			name:        "empty raw_json",
			rawJSON:     "",
			expectError: false,
		},
		{
			name:        "valid json",
			rawJSON:     `{"key": "value"}`,
			expectError: false,
		},
		{
			name:        "valid json array",
			rawJSON:     `[{"key": "value"}]`,
			expectError: false,
		},
		{
			name:        "valid json nested",
			rawJSON:     `{"key": {"nested": "value"}}`,
			expectError: false,
		},
		{
			name:        "invalid json",
			rawJSON:     `{"key": "value"`,
			expectError: true,
		},
		{
			name:        "invalid json - trailing comma",
			rawJSON:     `{"key": "value",}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRawJSON(tt.rawJSON)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateRawJSON() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		uuid  string
		valid bool
	}{
		{
			name:  "valid uuid",
			uuid:  "550e8400-e29b-41d4-a716-446655440000",
			valid: true,
		},
		{
			name:  "invalid uuid - missing dashes",
			uuid:  "550e8400e29b41d4a716446655440000",
			valid: false,
		},
		{
			name:  "invalid uuid - wrong format",
			uuid:  "not-a-uuid",
			valid: false,
		},
		{
			name:  "empty string",
			uuid:  "",
			valid: false,
		},
		{
			name:  "valid uuid with different values",
			uuid:  "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.uuid)
			if result != tt.valid {
				t.Errorf("IsValidUUID() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func TestIsAnalysisOutdated(t *testing.T) {
	tests := []struct {
		name            string
		analysis        *Analysis
		currentGitHead  string
		expectOutdated  bool
	}{
		{
			name:           "analysis is nil",
			analysis:       nil,
			currentGitHead: "abc123",
			expectOutdated: true,
		},
		{
			name:           "current git head is empty",
			analysis:       &Analysis{AnalyzedGitHead: "abc123"},
			currentGitHead: "",
			expectOutdated: false,
		},
		{
			name:           "git heads match",
			analysis:       &Analysis{AnalyzedGitHead: "abc123"},
			currentGitHead: "abc123",
			expectOutdated: false,
		},
		{
			name:           "git heads differ",
			analysis:       &Analysis{AnalyzedGitHead: "abc123"},
			currentGitHead: "def456",
			expectOutdated: true,
		},
		{
			name:           "analysis git head is empty",
			analysis:       &Analysis{AnalyzedGitHead: ""},
			currentGitHead: "abc123",
			expectOutdated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAnalysisOutdated(tt.analysis, tt.currentGitHead)
			if result != tt.expectOutdated {
				t.Errorf("IsAnalysisOutdated() = %v, want %v", result, tt.expectOutdated)
			}
		})
	}
}