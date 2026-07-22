package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nerddevsltd/portfolio/pkg/models"
)

// Manager handles configuration loading, validation, and defaults
type Manager struct {
	loader    *Loader
	validator *Validator
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		loader:    NewLoader(configPath),
		validator: NewValidator(),
	}
}

// LoadConfig loads, validates, and returns configuration
func (m *Manager) LoadConfig() (*models.Config, error) {
	// Load configuration
	config, err := m.loader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := m.validator.Validate(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// CreateDefaultConfig creates a default configuration file
func (m *Manager) CreateDefaultConfig() error {
	defaultConfig := models.GetDefaultConfig()

	if err := m.loader.Save(defaultConfig); err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	fmt.Printf("Created default configuration at: %s\n", m.loader.configPath)
	return nil
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

// GetConfigStatus returns the current configuration status for diagnostics
func GetConfigStatus() (string, error) {
	configPath := models.GetConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "missing", nil
	}

	loader := NewLoader(configPath)
	config, err := loader.Load()
	if err != nil {
		return "invalid", err
	}

	if len(config.Discovery.ProjectRoots) == 0 {
		return "incomplete", nil
	}

	return "valid", nil
}