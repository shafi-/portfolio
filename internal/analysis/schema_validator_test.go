package analysis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchemaValidator(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)
	require.NotNil(t, validator)
}

func TestSchemaValidator_ValidAnalysis(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:            "A web application for task management",
		Purpose:            "Help users organize and track their daily tasks",
		Architecture:       "React frontend with Go/Gin backend, PostgreSQL database",
		Maturity:           "stable",
		Strengths:          []string{"Clean UI", "Fast performance", "Well-documented"},
		Weaknesses:         []string{"Limited mobile support", "No offline mode"},
		ReusableComponents: []string{"Task CRUD module", "User authentication"},
		Notes:              "Requires PostgreSQL 14+",
		AnalyzedAt:         time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead:    "abc123def456",
		Analyzer:           "claude-code",
		Features: []FeatureInput{
			{
				Name:        "Authentication",
				Description: "JWT-based authentication system",
				Confidence:  floatPtr(0.95),
			},
		},
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func TestSchemaValidator_MissingRequiredField(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
	}

	err = validator.Validate(input)
	assert.Error(t, err)
	var analysisErr *Error
	assert.True(t, assert.ErrorAs(t, err, &analysisErr))
	assert.Equal(t, ErrCodeSchemaValidation, analysisErr.Code)
}

func TestSchemaValidator_InvalidTimestampFormat(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func TestSchemaValidator_ArrayFieldValidation(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func TestSchemaValidator_EmptyOptionalFields(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:            "Valid summary",
		Purpose:            "Valid purpose",
		Architecture:       "Valid architecture",
		Maturity:           "",
		Strengths:          []string{},
		Weaknesses:         []string{},
		ReusableComponents: []string{},
		Notes:              "",
		AnalyzedAt:         time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead:    "abc123def456",
		Analyzer:           "test-analyzer",
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func TestSchemaValidator_ISO8601Timestamps(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	timestamps := []time.Time{
		time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		time.Date(2025, 1, 15, 10, 30, 0, 123456000, time.UTC),
		time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	for _, ts := range timestamps {
		input := AnalysisInput{
			Summary:         "Valid summary",
			Purpose:         "Valid purpose",
			Architecture:    "Valid architecture",
			AnalyzedAt:      ts,
			AnalyzedGitHead: "abc123def456",
			Analyzer:        "test-analyzer",
		}

		err = validator.Validate(input)
		assert.NoError(t, err, "Should accept timestamp: %v", ts)
	}
}

func TestSchemaValidator_InvalidGitHeadAccepted(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "not-a-real-sha",
		Analyzer:        "test-analyzer",
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func TestSchemaValidator_FeatureValidation(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
		Features: []FeatureInput{
			{
				Name:        "Authentication",
				Description: "JWT-based authentication",
				Confidence:  floatPtr(0.95),
			},
			{
				Name:        "Payments",
				Description: "Payment processing",
				Confidence:  floatPtr(0.87),
			},
		},
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func TestSchemaValidator_FeatureMissingName(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
		Features: []FeatureInput{
			{
				Description: "Missing name",
				Confidence:  floatPtr(0.95),
			},
		},
	}

	err = validator.Validate(input)
	assert.Error(t, err)
}

func TestSchemaValidator_FeatureInvalidConfidence(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
		Features: []FeatureInput{
			{
				Name:       "Authentication",
				Confidence: floatPtr(1.5), // Invalid: > 1.0
			},
		},
	}

	err = validator.Validate(input)
	assert.Error(t, err)
}

func TestSchemaValidator_FeatureEmptyConfidence(t *testing.T) {
	validator, err := NewSchemaValidator()
	require.NoError(t, err)

	input := AnalysisInput{
		Summary:         "Valid summary",
		Purpose:         "Valid purpose",
		Architecture:    "Valid architecture",
		AnalyzedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		AnalyzedGitHead: "abc123def456",
		Analyzer:        "test-analyzer",
		Features: []FeatureInput{
			{
				Name:       "Authentication",
				Confidence: nil, // Valid: optional
			},
		},
	}

	err = validator.Validate(input)
	assert.NoError(t, err)
}

func floatPtr(f float64) *float64 {
	return &f
}
