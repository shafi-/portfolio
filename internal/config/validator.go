package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"project-dash/pkg/models"
)

// Validate checks configuration for errors
func Validate(config *models.Config) error {
	var errors []string

	// Validate general section
	if err := validateGeneralConfig(&config.General); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate discovery section
	if err := validateDiscoveryConfig(&config.Discovery); err != nil {
		errors = append(errors, err.Error())
	}

	// Validate logging section
	if err := validateLoggingConfig(&config.Logging); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return &ConfigError{
			Message: fmt.Sprintf("Configuration validation failed:\n- %s", strings.Join(errors, "\n- ")),
		}
	}

	return nil
}

// validateGeneralConfig validates the general configuration section
func validateGeneralConfig(config *models.GeneralConfig) error {
	if config.DatabasePath == "" {
		return fmt.Errorf("general.database_path is required")
	}

	// Check if parent directory exists and is writable
	parentDir := filepath.Dir(config.DatabasePath)
	if info, err := os.Stat(parentDir); os.IsNotExist(err) {
		return fmt.Errorf("Database parent directory does not exist: %s", parentDir)
	} else if err != nil {
		return fmt.Errorf("Cannot access database parent directory: %s", parentDir)
	} else if !info.IsDir() {
		return fmt.Errorf("Database path parent is not a directory: %s", parentDir)
	}

	// Check if directory is writable
	testFile := filepath.Join(parentDir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("Database parent directory is not writable: %s", parentDir)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// validateDiscoveryConfig validates the discovery configuration section
func validateDiscoveryConfig(config *models.DiscoveryConfig) error {
	if len(config.ProjectRoots) == 0 {
		return fmt.Errorf("discovery.project_roots cannot be empty - run 'portfolio init' to configure")
	}

	// Validate each project root
	for i, root := range config.ProjectRoots {
		if err := validatePath(root, i, "project_roots"); err != nil {
			return err
		}
	}

	// Validate ignored paths (just check they're valid patterns)
	for i, pattern := range config.IgnoredPaths {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("discovery.ignored_paths[%d] is invalid glob pattern: %s", i, pattern)
		}
	}

	return nil
}

// validateLoggingConfig validates the logging configuration section
func validateLoggingConfig(config *models.LoggingConfig) error {
	validLevels := models.ValidLogLevels()

	if config.Level == "" {
		return fmt.Errorf("logging.level is required")
	}

	config.Level = strings.ToUpper(config.Level)

	if !contains(validLevels, config.Level) {
		return fmt.Errorf("logging.level must be one of: %s (got: %s)",
			strings.Join(validLevels, ", "), config.Level)
	}

	return nil
}

// validatePath validates a filesystem path
func validatePath(path string, index int, field string) error {
	if path == "" {
		return fmt.Errorf("discovery.%s[%d] is empty", field, index)
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("discovery.%s[%d] does not exist: %s", field, index, path)
	} else if err != nil {
		return fmt.Errorf("discovery.%s[%d] is not accessible: %s", field, index, path)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("discovery.%s[%d] is not a directory: %s", field, index, path)
	}

	// Check if readable
	if _, err := os.ReadDir(path); err != nil {
		return fmt.Errorf("discovery.%s[%d] is not readable: %s", field, index, path)
	}

	return nil
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
