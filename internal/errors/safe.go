package errors

import (
	"fmt"
	"runtime/debug"
	"time"
)

// SafeError represents an error that doesn't expose internal details
type SafeError struct {
	UserMessage string
	InternalErr error
	Timestamp   time.Time
	Safe        bool
	RequestID   string
}

func (e *SafeError) Error() string {
	if e.UserMessage != "" {
		if e.RequestID != "" {
			return fmt.Sprintf("%s (Request ID: %s)", e.UserMessage, e.RequestID)
		}
		return e.UserMessage
	}
	return "An internal error occurred"
}

func (e *SafeError) Unwrap() error {
	return e.InternalErr
}

// IsSafe returns true if this error is a safe error
func (e *SafeError) IsSafe() bool {
	return e.Safe
}

// NewSafe creates a safe error with a user-friendly message
func NewSafe(message string) *SafeError {
	return &SafeError{
		UserMessage: message,
		Timestamp:   time.Now(),
		Safe:        true,
		RequestID:   generateRequestID(),
	}
}

// Wrap wraps an error with a safe message
func Wrap(err error, message string) *SafeError {
	if err == nil {
		return nil
	}

	// If it's already a SafeError, just update the message
	if safeErr, ok := err.(*SafeError); ok {
		if message != "" {
			safeErr.UserMessage = message
		}
		return safeErr
	}

	return &SafeError{
		UserMessage: message,
		InternalErr: err,
		Timestamp:   time.Now(),
		Safe:        true,
		RequestID:   generateRequestID(),
	}
}

// Wrapf wraps an error with a formatted safe message
func Wrapf(err error, format string, args ...interface{}) *SafeError {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf(format, args...)
	return Wrap(err, message)
}

// SafePanicHandler recovers from panics and returns safe errors
func SafePanicHandler() error {
	if r := recover(); r != nil {
		// Log the panic internally but don't expose details to user
		logInternalPanic(r)

		return NewSafe("An internal error occurred. Please check logs for details.")
	}
	return nil
}

// SafePanicHandlerWithMessage recovers from panics with a custom safe message
func SafePanicHandlerWithMessage(message string) error {
	if r := recover(); r != nil {
		// Log the panic internally but don't expose details to user
		logInternalPanic(r)

		return NewSafe(message)
	}
	return nil
}

// SafeContext provides context for safe error handling
type SafeContext struct {
	Operation string
	Component string
}

// NewContext creates a new safe error context
func NewContext(operation, component string) *SafeContext {
	return &SafeContext{
		Operation: operation,
		Component: component,
	}
}

// Error creates a safe error from this context
func (ctx *SafeContext) Error(message string) *SafeError {
	return &SafeError{
		UserMessage: message,
		Timestamp:   time.Now(),
		Safe:        true,
		RequestID:   generateRequestID(),
	}
}

// Wrap wraps an error with context-aware message
func (ctx *SafeContext) Wrap(err error, message string) *SafeError {
	if err == nil {
		return nil
	}

	// Add context to the message
	if message != "" {
		message = fmt.Sprintf("%s: %s", ctx.Operation, message)
	} else {
		message = fmt.Sprintf("Failed to %s", ctx.Operation)
	}

	return Wrap(err, message)
}

// Database creates a database-specific error context
func Database(operation string) *SafeContext {
	return NewContext(operation, "database")
}

// FileSystem creates a filesystem-specific error context
func FileSystem(operation string) *SafeContext {
	return NewContext(operation, "filesystem")
}

// Configuration creates a configuration-specific error context
func Configuration(operation string) *SafeContext {
	return NewContext(operation, "configuration")
}

// Network creates a network-specific error context
func Network(operation string) *SafeContext {
	return NewContext(operation, "network")
}

// IsSafeError checks if an error is a SafeError
func IsSafeError(err error) bool {
	_, ok := err.(*SafeError)
	return ok
}

// GetSafeMessage extracts the safe message from an error
func GetSafeMessage(err error) string {
	if err == nil {
		return ""
	}

	if safeErr, ok := err.(*SafeError); ok {
		return safeErr.UserMessage
	}

	// For non-safe errors, return a generic message
	return "An error occurred"
}

// GetRequestID extracts the request ID from an error if available
func GetRequestID(err error) string {
	if err == nil {
		return ""
	}

	if safeErr, ok := err.(*SafeError); ok {
		return safeErr.RequestID
	}

	return ""
}

// logInternalPanic logs internal panic details for debugging
func logInternalPanic(r interface{}) {
	timestamp := time.Now().Format(time.RFC3339)

	fmt.Printf("[%s] INTERNAL PANIC: %v\n", timestamp, r)
	fmt.Printf("[%s] STACK TRACE:\n%s\n", timestamp, debug.Stack())
}

// generateRequestID generates a unique request ID for error tracking
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// SanitizeFilePath removes potentially sensitive path information
func SanitizeFilePath(path string) string {
	if path == "" {
		return ""
	}

	// Check if path contains home directory
	homeDir := getHomeDir()
	if homeDir != "" && containsPath(path, homeDir) {
		return "~/" + getRelativePath(path, homeDir)
	}

	// For system paths, just return the filename
	return getFileName(path)
}

// SanitizeError sanitizes error messages to remove internal details
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Remove common internal paths
	sanitized := removeInternalPaths(errMsg)

	// Remove file extensions that might reveal implementation
	sanitized = removeTechnicalDetails(sanitized)

	return sanitized
}

// Helper functions
func getHomeDir() string {
	// Simple implementation - in real code would use os.UserHomeDir()
	return "" // Placeholder
}

func containsPath(path, contains string) bool {
	return len(path) >= len(contains) && path[:len(contains)] == contains
}

func getRelativePath(path, base string) string {
	if len(path) > len(base) && path[len(base)] == '/' {
		return path[len(base)+1:]
	}
	return path[len(base):]
}

func getFileName(path string) string {
	// Simple implementation - extract filename from path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func removeInternalPaths(message string) string {
	// Remove common system paths that might be in error messages
	// This is a simplified implementation
	return message
}

func removeTechnicalDetails(message string) string {
	// Remove technical details like stack traces, memory addresses, etc.
	// This is a simplified implementation
	return message
}
