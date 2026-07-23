package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"project-dash/pkg/models"
)

// Validator handles configuration validation
type Validator struct{}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks configuration for errors
func (v *Validator) Validate(config *models.Config) error {
	var errors []string

	// Validate general section
	if err := v.validateGeneralConfig(&config.General); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate discovery section
	if err := v.validateDiscoveryConfig(&config.Discovery); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate logging section
	if err := v.validateLoggingConfig(&config.Logging); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return &ValidationError{
			Message: "Configuration validation failed",
			Errors:  errors,
		}
	}

	return nil
}

// validateGeneralConfig validates the general configuration section
func (v *Validator) validateGeneralConfig(config *models.GeneralConfig) error {
	if config.DatabasePath == "" {
		return &ValidationError{
			Message: "general.database_path is required",
		}
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(config.DatabasePath)
	if info, err := os.Stat(parentDir); os.IsNotExist(err) {
		return &ValidationError{
			Message: fmt.Sprintf("Database parent directory does not exist: %s", parentDir),
			Action:  fmt.Sprintf("Create the directory: mkdir -p %s", parentDir),
		}
	} else if err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("Cannot access database parent directory: %s", parentDir),
		}
	} else if !info.IsDir() {
		return &ValidationError{
			Message: fmt.Sprintf("Database path parent is not a directory: %s", parentDir),
		}
	}

	// Check if directory is writable
	testFile := filepath.Join(parentDir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("Database parent directory is not writable: %s", parentDir),
			Action:  "Check directory permissions",
		}
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// validateDiscoveryConfig validates the discovery configuration section
func (v *Validator) validateDiscoveryConfig(config *models.DiscoveryConfig) error {
	if len(config.ProjectRoots) == 0 {
		return &ValidationError{
			Message: "discovery.project_roots cannot be empty - run 'portfolio init' to configure",
			Action:  "Run 'portfolio init' to set up project roots",
		}
	}

	// Validate each project root
	for i, root := range config.ProjectRoots {
		if err := v.validatePath(root, i, "project_roots"); err != nil {
			return err
		}
	}

	// Validate ignored paths (just check they're valid patterns)
	for i, pattern := range config.IgnoredPaths {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return &ValidationError{
				Message: fmt.Sprintf("discovery.ignored_paths[%d] is invalid glob pattern: %s", i, pattern),
			}
		}
	}

	return nil
}

// validateLoggingConfig validates the logging configuration section
func (v *Validator) validateLoggingConfig(config *models.LoggingConfig) error {
	validLevels := models.ValidLogLevels()

	if config.Level == "" {
		return &ValidationError{
			Message: "logging.level is required",
		}
	}

	config.Level = strings.ToUpper(config.Level)

	if !contains(validLevels, config.Level) {
		return &ValidationError{
			Message: fmt.Sprintf("logging.level must be one of: %s (got: %s)",
				strings.Join(validLevels, ", "), config.Level),
			Action: "Use one of: DEBUG, INFO, WARN, ERROR",
		}
	}

	return nil
}

// validatePath validates a filesystem path
func (v *Validator) validatePath(path string, index int, field string) error {
	if path == "" {
		return &ValidationError{
			Message: fmt.Sprintf("discovery.%s[%d] is empty", field, index),
		}
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &ValidationError{
			Message: fmt.Sprintf("discovery.%s[%d] does not exist: %s", field, index, path),
			Action:  "Create the directory or provide a valid path",
		}
	} else if err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("discovery.%s[%d] is not accessible: %s", field, index, path),
		}
	}

	// Check if it's a directory
	if !info.IsDir() {
		return &ValidationError{
			Message: fmt.Sprintf("discovery.%s[%d] is not a directory: %s", field, index, path),
		}
	}

	// Check if readable
	if _, err := os.ReadDir(path); err != nil {
		return &ValidationError{
			Message: fmt.Sprintf("discovery.%s[%d] is not readable: %s", field, index, path),
			Action:  "Check directory permissions",
		}
	}

	return nil
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Message string
	Errors  []string
	Action  string
}

func (e *ValidationError) Error() string {
	if e.Action != "" {
		return fmt.Sprintf("%s\nAction: %s", e.Message, e.Action)
	}
	if len(e.Errors) > 0 {
		return fmt.Sprintf("%s:\n- %s", e.Message, strings.Join(e.Errors, "\n- "))
	}
	return e.Message
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
