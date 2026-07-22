package models

// LogLevel represents the logging level
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogComponent represents the logging component
type LogComponent string

const (
	ComponentConfig    LogComponent = "config"
	ComponentDatabase  LogComponent = "database"
	ComponentCLI       LogComponent = "cli"
	ComponentDiscovery LogComponent = "discovery"
	ComponentEngine    LogComponent = "engine"
)

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}
