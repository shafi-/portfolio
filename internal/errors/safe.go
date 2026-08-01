package errors

import (
	"fmt"
	"os"
	"regexp"
	"strings"
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
	if homeDir != "" && strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
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

	// Replace home directory with ~
	homeDir := getHomeDir()
	if homeDir != "" {
		errMsg = strings.ReplaceAll(errMsg, homeDir, "~")
	}

	// Strip memory addresses (0x followed by hex digits)
	errMsg = stripMemoryAddresses(errMsg)

	return errMsg
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// getFileName extracts the filename from a path
func getFileName(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

// stripMemoryAddresses removes memory addresses like 0xc000123456 from error messages
var memoryAddrRegex = regexp.MustCompile(`0x[0-9a-fA-F]+`)

func stripMemoryAddresses(message string) string {
	return memoryAddrRegex.ReplaceAllString(message, "0x...")
}
