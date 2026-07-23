package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"project-dash/pkg/models"
)

// Loader handles configuration loading and saving
type Loader struct {
	configPath string
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	if configPath == "" {
		configPath = models.GetConfigPath()
	}
	return &Loader{configPath: configPath}
}

// Load reads and parses the configuration file
func (l *Loader) Load() (*models.Config, error) {
	// Check if config file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		// Create default config if missing
		defaultConfig := models.GetDefaultConfig()
		if err := l.Save(defaultConfig); err != nil {
			return nil, &ConfigError{
				Message: fmt.Sprintf("Cannot create default config: %s", l.configPath),
				Cause:   err,
			}
		}
		return defaultConfig, nil
	}

	// Read existing config
	data, err := os.ReadFile(l.configPath)
	if err != nil {
		return nil, &ConfigError{
			Message: fmt.Sprintf("Cannot read config file: %s", l.configPath),
			Cause:   err,
		}
	}

	// Parse TOML
	config := models.GetDefaultConfig()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, &ConfigError{
			Message: fmt.Sprintf("Invalid TOML syntax in config file: %s", l.configPath),
			Cause:   err,
		}
	}

	return config, nil
}

// Save writes configuration to file
func (l *Loader) Save(config *models.Config) error {
	// Ensure directory exists
	dir := filepath.Dir(l.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with secure permissions
	if err := os.WriteFile(l.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ConfigError represents a configuration error with context
type ConfigError struct {
	Message string
	Cause   error
}

func (e *ConfigError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *ConfigError) Unwrap() error {
	return e.Cause
}

// EnsureConfigDir ensures the configuration directory exists
func EnsureConfigDir() error {
	configPath := models.GetConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}
