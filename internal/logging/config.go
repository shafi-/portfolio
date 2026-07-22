package logging

import (
	"fmt"
	"strings"
)

// Config holds logging configuration
type Config struct {
	Level      string
	Format     string // "json" or "console"
	Components map[string]string // Component-specific log levels
}

// LoadConfigFromEnv loads logging configuration from environment
func LoadConfigFromEnv() *Config {
	return &Config{
		Level:  GetLogLevelFromEnv(),
		Format: GetLogFormatFromEnv(),
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

// GetEffectiveLevel returns the effective log level for a component
func (c *Config) GetEffectiveLevel(component string) string {
	if componentLevel, exists := c.Components[component]; exists {
		return componentLevel
	}
	return c.Level
}
