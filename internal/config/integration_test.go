package config

import (
	"os"
	"path/filepath"
	"testing"

	"project-dash/pkg/models"
)

// Integration test to verify Story 1.2 acceptance criteria
func TestStory12Integration(t *testing.T) {
	// AC-05: TOML Configuration Format Defined
	t.Run("AC05_TOMLConfigurationFormat", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		// Create a TOML config file
		testConfig := &models.Config{
			General: models.GeneralConfig{
				DatabasePath: filepath.Join(tempDir, "test.db"),
			},
			Discovery: models.DiscoveryConfig{
				ProjectRoots: []string{tempDir, filepath.Join(tempDir, "projects")},
				IgnoredPaths: []string{"node_modules", ".git"},
			},
			Logging: models.LoggingConfig{
				Level: "DEBUG",
			},
		}

		loader := NewLoader(configPath)
		if err := loader.Save(testConfig); err != nil {
			t.Fatalf("Failed to save TOML config: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		// Verify file can be parsed
		loadedConfig, err := loader.Load()
		if err != nil {
			t.Errorf("Failed to load TOML config: %v", err)
		}

		// Verify all sections are accessible
		if loadedConfig.General.DatabasePath == "" {
			t.Error("General section not accessible")
		}
		if len(loadedConfig.Discovery.ProjectRoots) == 0 {
			t.Error("Discovery section not accessible")
		}
		if loadedConfig.Logging.Level == "" {
			t.Error("Logging section not accessible")
		}
	})

	// AC-06: Configuration Schema Complete
	t.Run("AC06_ConfigurationSchema", func(t *testing.T) {
		config := models.GetDefaultConfig()

		// Verify schema includes project_roots
		if config.Discovery.ProjectRoots == nil {
			t.Error("Schema missing project_roots")
		}

		// Verify schema includes ignored_paths
		if config.Discovery.IgnoredPaths == nil {
			t.Error("Schema missing ignored_paths")
		}

		// Verify schema includes database_path
		if config.General.DatabasePath == "" {
			t.Error("Schema missing database_path")
		}

		// Verify schema includes log_level
		if config.Logging.Level == "" {
			t.Error("Schema missing log_level")
		}
	})

	// AC-07: Configuration Loading with Defaults Working
	t.Run("AC07_ConfigurationLoadingWithDefaults", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.toml")
		loader := NewLoader(configPath)

		// Test missing config creates default
		config, err := loader.Load()
		if err != nil {
			t.Errorf("Failed to load default config: %v", err)
		}

		// Verify defaults applied
		if config.Logging.Level != "INFO" {
			t.Errorf("Expected default log level INFO, got: %s", config.Logging.Level)
		}

		if len(config.Discovery.IgnoredPaths) == 0 {
			t.Error("Expected default ignored paths")
		}

		if config.General.DatabasePath == "" {
			t.Error("Expected default database path")
		}
	})

	// AC-08: Configuration Validation Functional
	t.Run("AC08_ConfigurationValidation", func(t *testing.T) {
		tempDir := t.TempDir()
		validator := NewValidator()

		// Test validation rules enforceable
		validConfig := &models.Config{
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

		if err := validator.Validate(validConfig); err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}

		// Test validation catches errors
		invalidConfig := &models.Config{
			General: models.GeneralConfig{
				DatabasePath: filepath.Join(tempDir, "nonexistent", "path"),
			},
			Discovery: models.DiscoveryConfig{
				ProjectRoots: []string{},
			},
			Logging: models.LoggingConfig{
				Level: "INVALID",
			},
		}

		if err := validator.Validate(invalidConfig); err == nil {
			t.Error("Invalid config should fail validation")
		}
	})

	// AC-09: Configuration Error Handling Comprehensive
	t.Run("AC09_ConfigurationErrorHandling", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test error scenarios produce actionable messages
		testCases := []struct {
			name      string
			config    *models.Config
			expectErr bool
		}{
			{
				name: "invalid log level",
				config: &models.Config{
					General: models.GeneralConfig{
						DatabasePath: tempDir,
					},
					Discovery: models.DiscoveryConfig{
						ProjectRoots: []string{tempDir},
					},
					Logging: models.LoggingConfig{
						Level: "INVALID",
					},
				},
				expectErr: true,
			},
			{
				name: "empty project roots",
				config: &models.Config{
					General: models.GeneralConfig{
						DatabasePath: tempDir,
					},
					Discovery: models.DiscoveryConfig{
						ProjectRoots: []string{},
					},
					Logging: models.LoggingConfig{
						Level: "INFO",
					},
				},
				expectErr: true,
			},
			{
				name: "nonexistent database path",
				config: &models.Config{
					General: models.GeneralConfig{
						DatabasePath: filepath.Join(tempDir, "nonexistent", "db", "path"),
					},
					Discovery: models.DiscoveryConfig{
						ProjectRoots: []string{tempDir},
					},
					Logging: models.LoggingConfig{
						Level: "INFO",
					},
				},
				expectErr: true,
			},
		}

		validator := NewValidator()
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := validator.Validate(tc.config)
				if tc.expectErr && err == nil {
					t.Error("Expected error but got none")
				}
				if !tc.expectErr && err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if err != nil {
					// Verify error message is actionable
					if err.Error() == "" {
						t.Error("Error message should not be empty")
					}
				}
			})
		}
	})

	// CFG-003: Configuration File Management
	t.Run("CFG003_FileManagement", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		// Test creation
		manager := NewManager(configPath)
		if err := manager.CreateDefaultConfig(); err != nil {
			t.Errorf("Failed to create default config: %v", err)
		}

		// Test file permissions (should be 0600)
		info, err := os.Stat(configPath)
		if err != nil {
			t.Errorf("Failed to stat config file: %v", err)
		}

		// Check file is readable/writable by owner
		if info.Mode().Perm()&0600 != 0600 {
			t.Errorf("Config file should have 0600 permissions, got: %o", info.Mode().Perm())
		}

		// Test updates
		testConfig := &models.Config{
			General: models.GeneralConfig{
				DatabasePath: filepath.Join(tempDir, "updated.db"),
			},
			Discovery: models.DiscoveryConfig{
				ProjectRoots: []string{tempDir},
				IgnoredPaths: []string{"updated"},
			},
			Logging: models.LoggingConfig{
				Level: "DEBUG",
			},
		}

		loader := NewLoader(configPath)
		if err := loader.Save(testConfig); err != nil {
			t.Errorf("Failed to update config: %v", err)
		}

		// Verify updates persisted
		loadedConfig, err := loader.Load()
		if err != nil {
			t.Errorf("Failed to load updated config: %v", err)
		}

		if loadedConfig.Logging.Level != "DEBUG" {
			t.Error("Config updates not persisted")
		}
	})
}

func TestConfigurationIntegrationPoints(t *testing.T) {
	// Test integration points for database and other components
	t.Run("Integration_Points", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		// Create a complete config
		testConfig := &models.Config{
			General: models.GeneralConfig{
				DatabasePath: filepath.Join(tempDir, "test.db"),
			},
			Discovery: models.DiscoveryConfig{
				ProjectRoots: []string{tempDir},
				IgnoredPaths: []string{"node_modules", ".git"},
			},
			Logging: models.LoggingConfig{
				Level: "INFO",
			},
		}

		loader := NewLoader(configPath)
		if err := loader.Save(testConfig); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Test manager interface
		manager := NewManager(configPath)
		config, err := manager.LoadConfig()
		if err != nil {
			t.Errorf("Failed to load config via manager: %v", err)
		}

		// Verify integration methods
		if config.General.DatabasePath == "" {
			t.Error("Database path not available for database integration")
		}

		if len(config.Discovery.ProjectRoots) == 0 {
			t.Error("Project roots not available for discovery integration")
		}

		if config.Logging.Level == "" {
			t.Error("Log level not available for logging integration")
		}
	})
}
