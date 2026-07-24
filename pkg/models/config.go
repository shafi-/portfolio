package models

import (
	"os"
	"path/filepath"
)

// Config represents the complete Portfolio configuration
type Config struct {
	General   GeneralConfig   `toml:"general"`
	Discovery DiscoveryConfig `toml:"discovery"`
	Logging   LoggingConfig   `toml:"logging"`
	Dashboard DashboardConfig `toml:"dashboard"`
}

// GeneralConfig contains system-wide configuration
type GeneralConfig struct {
	DatabasePath string `toml:"database_path"`
}

// DiscoveryConfig contains project discovery settings
type DiscoveryConfig struct {
	ProjectRoots []string `toml:"project_roots"`
	IgnoredPaths []string `toml:"ignored_paths"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level string `toml:"level"`
}

// DashboardConfig contains dashboard server settings
type DashboardConfig struct {
	Host           string   `toml:"host"`
	Port           int      `toml:"port"`
	AssetPath      string   `toml:"asset_path"`
	AllowedOrigins []string `toml:"allowed_origins"`
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		General: GeneralConfig{
			DatabasePath: filepath.Join(homeDir, ".portfolio", "portfolio.db"),
		},
		Discovery: DiscoveryConfig{
			ProjectRoots: []string{},
			IgnoredPaths: []string{
				"node_modules",
				".git",
				"vendor",
				"build",
				"dist",
				"target",
				"bin",
			},
		},
		Logging: LoggingConfig{
			Level: "INFO",
		},
		Dashboard: DashboardConfig{
			Host:           "localhost",
			Port:           8090,
			AssetPath:      "",
			AllowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		},
	}
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".portfolio", "config.toml")
}

// ValidLogLevels returns the list of valid log levels
func ValidLogLevels() []string {
	return []string{"DEBUG", "INFO", "WARN", "ERROR"}
}

// DefaultIgnoredPaths returns the default ignored paths
func DefaultIgnoredPaths() []string {
	return []string{
		"node_modules",
		".git",
		"vendor",
		"build",
		"dist",
		"target",
		"bin",
	}
}
