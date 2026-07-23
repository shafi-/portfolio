package config

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/pkg/models"
)

func TestLoadConfig(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	loader := NewLoader(configPath)

	// Test loading non-existent config (should create default)
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Expected no error for missing config, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be created")
	}

	// Verify defaults
	if config.Logging.Level != "INFO" {
		t.Errorf("Expected default log level INFO, got: %s", config.Logging.Level)
	}

	if config.General.DatabasePath == "" {
		t.Error("Expected default database path")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}
}

func TestConfigValidation(t *testing.T) {
	validator := NewValidator()

	// Test valid config
	validConfig := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: filepath.Join(os.TempDir(), "test.db"),
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: []string{os.TempDir()},
		},
		Logging: models.LoggingConfig{
			Level: "INFO",
		},
	}

	if err := validator.Validate(validConfig); err != nil {
		t.Errorf("Expected valid config to pass validation, got: %v", err)
	}

	// Test invalid log level
	invalidConfig := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: filepath.Join(os.TempDir(), "test.db"),
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: []string{os.TempDir()},
		},
		Logging: models.LoggingConfig{
			Level: "INVALID",
		},
	}

	if err := validator.Validate(invalidConfig); err == nil {
		t.Error("Expected validation error for invalid log level")
	}

	// Test empty project roots
	emptyRootsConfig := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: filepath.Join(os.TempDir(), "test.db"),
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: []string{},
		},
		Logging: models.LoggingConfig{
			Level: "INFO",
		},
	}

	if err := validator.Validate(emptyRootsConfig); err == nil {
		t.Error("Expected validation error for empty project roots")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")
	loader := NewLoader(configPath)

	// Create test config
	testConfig := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: []string{tempDir},
			IgnoredPaths: []string{"node_modules", ".git"},
		},
		Logging: models.LoggingConfig{
			Level: "DEBUG",
		},
	}

	// Save config
	if err := loader.Save(testConfig); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedConfig, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if loadedConfig.Logging.Level != "DEBUG" {
		t.Errorf("Expected log level DEBUG, got: %s", loadedConfig.Logging.Level)
	}

	if loadedConfig.General.DatabasePath != testConfig.General.DatabasePath {
		t.Error("Database path mismatch")
	}

	if len(loadedConfig.Discovery.ProjectRoots) != 1 {
		t.Error("Expected 1 project root")
	}
}

func TestManagerLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Create a valid config file first
	loader := NewLoader(configPath)
	testConfig := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: []string{tempDir},
		},
		Logging: models.LoggingConfig{
			Level: "INFO",
		},
	}

	if err := loader.Save(testConfig); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Test manager load
	manager := NewManager(configPath)
	config, err := manager.LoadConfig()
	if err != nil {
		t.Errorf("Expected successful load, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be loaded")
	}
}

func TestConfigError(t *testing.T) {
	// Test ConfigError structure
	err := &ConfigError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
	}

	if err.Error() != "Test error message" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestValidationError(t *testing.T) {
	// Test ValidationError with action
	err := &ValidationError{
		Message: "Validation failed",
		Action:  "Fix the configuration",
	}

	expectedMsg := "Validation failed\nAction: Fix the configuration"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message: %s, got: %s", expectedMsg, err.Error())
	}

	// Test ValidationError with multiple errors
	err2 := &ValidationError{
		Message: "Multiple errors",
		Errors:  []string{"Error 1", "Error 2"},
	}

	expectedMsg2 := "Multiple errors:\n- Error 1\n- Error 2"
	if err2.Error() != expectedMsg2 {
		t.Errorf("Expected error message: %s, got: %s", expectedMsg2, err2.Error())
	}
}

func TestDefaultIgnoredPaths(t *testing.T) {
	paths := models.DefaultIgnoredPaths()

	expectedPaths := []string{
		"node_modules",
		".git",
		"vendor",
		"build",
		"dist",
		"target",
		"bin",
	}

	if len(paths) != len(expectedPaths) {
		t.Errorf("Expected %d ignored paths, got: %d", len(expectedPaths), len(paths))
	}

	for i, path := range paths {
		if path != expectedPaths[i] {
			t.Errorf("Expected path %s at index %d, got: %s", expectedPaths[i], i, path)
		}
	}
}

func TestValidLogLevels(t *testing.T) {
	levels := models.ValidLogLevels()

	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	if len(levels) != len(expectedLevels) {
		t.Errorf("Expected %d log levels, got: %d", len(expectedLevels), len(levels))
	}

	for i, level := range levels {
		if level != expectedLevels[i] {
			t.Errorf("Expected level %s at index %d, got: %s", expectedLevels[i], i, level)
		}
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "config.toml")

	// Create a temporary config path
	configDir := filepath.Dir(configPath)

	// Ensure directory doesn't exist
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("Expected config directory to not exist")
	}

	// Create a custom GetConfigPath function for testing
	// Note: This would require modifying the function to accept a path parameter
	// For now, we'll test the EnsureConfigDir function with the default path
	err := EnsureConfigDir()
	if err != nil {
		t.Errorf("Expected successful directory creation, got: %v", err)
	}
}

func TestGetConfigStatus(t *testing.T) {
	// Test with non-existent config
	// Note: This test uses the default config path from GetConfigPath()
	status, err := GetConfigStatus()
	if err != nil {
		t.Errorf("Expected no error for status check, got: %v", err)
	}

	// Status should be one of: missing, invalid, incomplete, valid
	if status != "missing" && status != "invalid" && status != "incomplete" && status != "valid" {
		t.Errorf("Expected valid status value, got: %s", status)
	}
}
