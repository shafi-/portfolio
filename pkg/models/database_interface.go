package models

import "time"

// DatabaseInterface defines database operations
type DatabaseInterface interface {
	// Connection management
	Connect() error
	Close() error
	IsConnected() bool

	// Schema management
	Initialize() error
	ValidateSchema() error
	GetSchemaVersion() (int, error)

	// Migration management
	Migrate() error
	GetTableCount() (int, error)

	// Project operations
	GetProjectCount() (int, error)
	GetLastDiscoveryTime() (time.Time, error)

	// Health checks
	Ping() error
	ExecuteQuery(query string, args ...interface{}) (*Result, error)
}

// Result represents query results
type Result struct {
	Columns []string
	Rows    [][]interface{}
}
