package analysis

import (
	"fmt"
)

const (
	ErrCodeProjectNotFound     = "PROJECT_NOT_FOUND"
	ErrCodeSchemaValidation    = "SCHEMA_VALIDATION_FAILED"
	ErrCodeNotFound            = "ANALYSIS_NOT_FOUND"
	ErrCodeInvalidRelationType = "INVALID_RELATIONSHIP_TYPE"
	ErrCodeInvalidConfidence   = "INVALID_CONFIDENCE_RANGE"
	ErrCodeDuplicate           = "DUPLICATE_RELATIONSHIP"
)

var (
	ErrProjectNotFound     = &Error{Code: ErrCodeProjectNotFound, Message: "Project not found"}
	ErrAnalysisNotFound    = &Error{Code: ErrCodeNotFound, Message: "Analysis not found"}
	ErrInvalidRelationType = &Error{Code: ErrCodeInvalidRelationType, Message: "Invalid relationship type"}
)

// Error represents an analysis-specific error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new analysis error
func NewError(code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// WrapError wraps an error with analysis context
func WrapError(code, message string, cause error) error {
	return fmt.Errorf("%s: %w", message, cause)
}
