package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// simpleTOMLParser provides basic TOML parsing functionality
// This is a simplified implementation for standard library compliance
type simpleTOMLParser struct {
	data string
}

// parseTOMLFile parses a TOML file and returns a Config structure
func parseTOMLFile(configPath string) (*ConfigData, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	parser := &simpleTOMLParser{data: string(data)}
	return parser.parse()
}

// ConfigData represents the parsed configuration data
type ConfigData struct {
	GeneralConfig    GeneralConfigData
	DiscoveryConfig  DiscoveryConfigData
	LoggingConfig    LoggingConfigData
}

// GeneralConfigData represents general configuration data
type GeneralConfigData struct {
	DatabasePath string
}

// DiscoveryConfigData represents discovery configuration data
type DiscoveryConfigData struct {
	ProjectRoots []string
	IgnoredPaths []string
}

// LoggingConfigData represents logging configuration data
type LoggingConfigData struct {
	Level string
}

// parse parses the TOML data
func (p *simpleTOMLParser) parse() (*ConfigData, error) {
	config := &ConfigData{
		GeneralConfig: GeneralConfigData{
			DatabasePath: getDefaultDatabasePath(),
		},
		DiscoveryConfig: DiscoveryConfigData{
			ProjectRoots: []string{},
			IgnoredPaths: getDefaultIgnoredPaths(),
		},
		LoggingConfig: LoggingConfigData{
			Level: "INFO",
		},
	}

	lines := strings.Split(p.data, "\n")
	currentSection := ""
	var currentArrayBuilder *arrayBuilder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			currentArrayBuilder = nil // Reset array builder on section change
			continue
		}

		// Handle array continuation
		if currentArrayBuilder != nil {
			if strings.HasPrefix(line, "]") {
				// End of array
				currentArrayBuilder.finish(config)
				currentArrayBuilder = nil
				continue
			}
			// Add to array
			currentArrayBuilder.addValue(line)
			continue
		}

		// Handle array start
		if strings.Contains(line, "[") {
			if idx := strings.Index(line, "="); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])

				// Check if this is an array assignment
				if strings.Contains(value, "[") {
					currentArrayBuilder = &arrayBuilder{
						section: currentSection,
						key:     key,
						values:  []string{},
					}

					// Check if array starts and ends on same line
					if strings.Contains(value, "]") {
						// Single line array
						arrayContent := strings.TrimSpace(strings.TrimPrefix(value, "["))
						arrayContent = strings.TrimSuffix(arrayContent, "]")
						if arrayContent != "" {
							currentArrayBuilder.values = parseInlineArray(arrayContent)
						}
						currentArrayBuilder.finish(config)
						currentArrayBuilder = nil
					} else {
						// Multi-line array - values will be collected in subsequent iterations
						if remaining := strings.TrimSpace(strings.TrimPrefix(value, "[")); remaining != "" && !strings.Contains(remaining, "]") {
							currentArrayBuilder.addValue(remaining)
						}
					}
					continue
				}
			}
		}

		// Handle key-value pairs
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Remove quotes from value
			value = strings.Trim(value, "\"")

			p.processKeyValue(currentSection, key, value, config)
		}
	}

	return config, nil
}

// processKeyValue processes a key-value pair
func (p *simpleTOMLParser) processKeyValue(section, key, value string, config *ConfigData) {
	switch section {
	case "general":
		if key == "database_path" {
			config.GeneralConfig.DatabasePath = value
		}
	case "discovery":
		if key == "project_roots" {
			config.DiscoveryConfig.ProjectRoots = parseArray(value)
		} else if key == "ignored_paths" {
			config.DiscoveryConfig.IgnoredPaths = parseArray(value)
		}
	case "logging":
		if key == "level" {
			config.LoggingConfig.Level = value
		}
	}
}

// parseArray parses a simple array string
func parseArray(value string) []string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		// Remove brackets and split by comma
		inner := strings.TrimSpace(value[1 : len(value)-1])
		if inner == "" {
			return []string{}
		}

		items := strings.Split(inner, ",")
		result := []string{}
		for _, item := range items {
			item = strings.TrimSpace(item)
			item = strings.Trim(item, "\"")
			if item != "" {
				result = append(result, item)
			}
		}
		return result
	}
	return []string{}
}

// arrayBuilder handles multi-line array parsing
type arrayBuilder struct {
	section string
	key     string
	values  []string
}

// addValue adds a value to the array being built
func (ab *arrayBuilder) addValue(line string) {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, ",")
	line = strings.Trim(line, "\"")
	if line != "" {
		ab.values = append(ab.values, line)
	}
}

// finish completes the array and assigns it to the config
func (ab *arrayBuilder) finish(config *ConfigData) {
	switch ab.section {
	case "discovery":
		if ab.key == "project_roots" {
			config.DiscoveryConfig.ProjectRoots = ab.values
		} else if ab.key == "ignored_paths" {
			config.DiscoveryConfig.IgnoredPaths = ab.values
		}
	}
}

// parseInlineArray parses an inline array (comma-separated values)
func parseInlineArray(value string) []string {
	if value == "" {
		return []string{}
	}

	items := strings.Split(value, ",")
	result := []string{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, "\"")
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

// getDefaultDatabasePath returns the default database path
func getDefaultDatabasePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".portfolio", "portfolio.db")
}

// getDefaultIgnoredPaths returns default ignored paths
func getDefaultIgnoredPaths() []string {
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

// writeTOMLFile writes configuration data to a TOML file
func writeTOMLFile(configPath string, config *ConfigData) error {
	var builder strings.Builder

	builder.WriteString("# Portfolio Engine Configuration\n")
	builder.WriteString("\n")
	builder.WriteString("[general]\n")
	builder.WriteString(fmt.Sprintf("database_path = \"%s\"\n", config.GeneralConfig.DatabasePath))
	builder.WriteString("\n")
	builder.WriteString("[discovery]\n")
	builder.WriteString("# Directories to scan for projects\n")
	builder.WriteString("project_roots = [\n")
	for i, root := range config.DiscoveryConfig.ProjectRoots {
		if i < len(config.DiscoveryConfig.ProjectRoots)-1 {
			builder.WriteString(fmt.Sprintf("    \"%s\",\n", root))
		} else {
			builder.WriteString(fmt.Sprintf("    \"%s\"\n", root))
		}
	}
	builder.WriteString("]\n\n")
	builder.WriteString("# Paths/patterns to ignore during discovery\n")
	builder.WriteString("ignored_paths = [\n")
	for i, path := range config.DiscoveryConfig.IgnoredPaths {
		if i < len(config.DiscoveryConfig.IgnoredPaths)-1 {
			builder.WriteString(fmt.Sprintf("    \"%s\",\n", path))
		} else {
			builder.WriteString(fmt.Sprintf("    \"%s\"\n", path))
		}
	}
	builder.WriteString("]\n\n")
	builder.WriteString("[logging]\n")
	builder.WriteString("# Log level: DEBUG, INFO, WARN, ERROR\n")
	builder.WriteString(fmt.Sprintf("level = \"%s\"\n", config.LoggingConfig.Level))

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write with secure permissions
	return os.WriteFile(configPath, []byte(builder.String()), 0600)
}