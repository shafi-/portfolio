package logging

import (
	"fmt"
	"os"
	"strings"

	"project-dash/pkg/models"
)

// Config holds logging configuration
type Config struct {
	Level  string
	Format string // "json" or "console"
	File   string // Path to log file; empty disables file logging
}

// LoadConfigFromEnv loads logging configuration from environment
func LoadConfigFromEnv() *Config {
	return &Config{
		Level:  GetLogLevelFromEnv(),
		Format: GetLogFormatFromEnv(),
		File:   GetLogFileFromEnv(),
	}
}

// LoadConfigFromSettings loads logging configuration from settings
func LoadConfigFromSettings(level string, format string) *Config {
	return &Config{
		Level:  level,
		Format: format,
	}
}

// Validate checks if the logging configuration is valid
func (c *Config) Validate() error {
	// Validate log level
	if c.Level == "" {
		c.Level = "INFO"
	}

	validLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	c.Level = strings.ToUpper(c.Level)

	valid := false
	for _, level := range validLevels {
		if c.Level == level {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)",
			c.Level, strings.Join(validLevels, ", "))
	}

	// Validate format
	if c.Format == "" {
		c.Format = "console"
	}

	c.Format = strings.ToLower(c.Format)
	if c.Format != "json" && c.Format != "console" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'console')", c.Format)
	}

	return nil
}

// GetLogFileFromEnv returns log file path from environment variable or default
func GetLogFileFromEnv() string {
	if path := os.Getenv("PORTFOLIO_LOG_FILE"); path != "" {
		return path
	}
	return models.GetDefaultLogPath()
}
