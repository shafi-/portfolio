package models

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetPortfolioDataDir returns the secure data directory for Portfolio
// Uses system-wide locations away from home folder for better security
func GetPortfolioDataDir() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: System-wide Application Support
		// Requires proper permissions, more secure than user home
		return "/Library/Application Support/com.portfolio.cli"
	case "windows":
		// Windows: ProgramData (system-wide, not user-specific)
		// More secure than user AppData
		return filepath.Join(os.Getenv("ProgramData"), "portfolio")
	default:
		// Linux: System-wide application data
		// More secure than user home directories
		return "/var/lib/portfolio"
	}
}

// GetPortfolioConfigDir returns the secure config directory for Portfolio
// Uses system-wide locations for better security
func GetPortfolioConfigDir() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: System config location
		return "/Library/Preferences/com.portfolio.cli"
	case "windows":
		// Windows: System program data
		return filepath.Join(os.Getenv("ProgramData"), "portfolio", "config")
	default:
		// Linux: System-wide config
		return "/etc/portfolio"
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
