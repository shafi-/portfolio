package models

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetPortfolioDataDir returns the data directory for Portfolio
func GetPortfolioDataDir() string {
	switch runtime.GOOS {
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Library", "Application Support", "com.portfolio.cli")
	case "windows":
		return filepath.Join(os.Getenv("AppData"), "portfolio")
	default:
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".config", "portfolio")
	}
}

// GetPortfolioConfigDir returns the config directory for Portfolio
func GetPortfolioConfigDir() string {
	switch runtime.GOOS {
	case "darwin":
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "Library", "Application Support", "com.portfolio.cli")
	case "windows":
		return filepath.Join(os.Getenv("AppData"), "portfolio")
	default:
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".config", "portfolio")
	}
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
