package models

import (
	"os"
	"path/filepath"
)

// GetPortfolioDataDir returns the data directory for Portfolio.
// Uses os.UserConfigDir() which respects platform conventions:
//   - macOS:   ~/Library/Application Support
//   - Windows: %AppData%
//   - Linux:   $XDG_CONFIG_HOME or ~/.config
func GetPortfolioDataDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "com.portfolio.cli")
}

// GetPortfolioConfigDir returns the config directory for Portfolio.
// Same as GetPortfolioDataDir on all platforms.
func GetPortfolioConfigDir() string {
	return GetPortfolioDataDir()
}

// GetDefaultDatabasePath returns the secure database path
// Uses obscured filename to prevent discovery through guessing
func GetDefaultDatabasePath() string {
	return filepath.Join(GetPortfolioDataDir(), ".portfoliodata")
}

// GetDefaultConfigPath returns the secure config file path
func GetDefaultConfigPath() string {
	return filepath.Join(GetPortfolioConfigDir(), "config.toml")
}

// GetDefaultLogPath returns the secure log file path
func GetDefaultLogPath() string {
	return filepath.Join(GetPortfolioDataDir(), "portfolio.log")
}

// GetDefaultIntegrationsDir returns the default directory for integration files
func GetDefaultIntegrationsDir() string {
	return filepath.Join(GetPortfolioDataDir(), "integrations")
}

// GetDatabaseKeyPath returns the secure database key file path
func GetDatabaseKeyPath() string {
	return filepath.Join(GetPortfolioConfigDir(), "db_key")
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	return GetDefaultConfigPath()
}

// GetLegacyPortfolioDir returns the old ~/.portfolio path for migration
func GetLegacyPortfolioDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".portfolio")
}

// ShouldMigrateFromLegacy checks if migration from old ~/.portfolio is needed
func ShouldMigrateFromLegacy() bool {
	legacyDir := GetLegacyPortfolioDir()

	if _, err := os.Stat(legacyDir); os.IsNotExist(err) {
		return false
	}

	legacyDB := filepath.Join(legacyDir, "portfolio.db")
	if _, err := os.Stat(legacyDB); os.IsNotExist(err) {
		return false
	}

	newDB := GetDefaultDatabasePath()
	if _, err := os.Stat(newDB); err == nil {
		return false
	}

	return true
}
