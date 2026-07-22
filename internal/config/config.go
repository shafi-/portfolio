package config

import (
	"fmt"
	"os"

	"github.com/nerddevsltd/portfolio/pkg/models"
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
				Code:    "CONFIG_CREATE_FAILED",
				Message: fmt.Sprintf("Cannot create default config: %s", l.configPath),
				Cause:   err,
			}
		}
		return defaultConfig, nil
	}

	// Read existing config
	configData, err := parseTOMLFile(l.configPath)
	if err != nil {
		return nil, &ConfigError{
			Code:    "CONFIG_NOT_READABLE",
			Message: fmt.Sprintf("Cannot read config file: %s", l.configPath),
			Cause:   err,
		}
	}

	// Convert to models.Config
	config := &models.Config{
		General: models.GeneralConfig{
			DatabasePath: configData.GeneralConfig.DatabasePath,
		},
		Discovery: models.DiscoveryConfig{
			ProjectRoots: configData.DiscoveryConfig.ProjectRoots,
			IgnoredPaths:  configData.DiscoveryConfig.IgnoredPaths,
		},
		Logging: models.LoggingConfig{
			Level: configData.LoggingConfig.Level,
		},
	}

	return config, nil
}

// Save writes configuration to file
func (l *Loader) Save(config *models.Config) error {
	// Convert to ConfigData
	configData := &ConfigData{
		GeneralConfig: GeneralConfigData{
			DatabasePath: config.General.DatabasePath,
		},
		DiscoveryConfig: DiscoveryConfigData{
			ProjectRoots: config.Discovery.ProjectRoots,
			IgnoredPaths:  config.Discovery.IgnoredPaths,
		},
		LoggingConfig: LoggingConfigData{
			Level: config.Logging.Level,
		},
	}

	return writeTOMLFile(l.configPath, configData)
}

// ConfigError represents a configuration error with context
type ConfigError struct {
	Code    string
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