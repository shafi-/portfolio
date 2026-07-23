package discovery

import (
	"fmt"
)

// WalkEvent represents the type of event during directory walking
type WalkEvent int

const (
	EventEnterDir  WalkEvent = iota // entering a directory
	EventFoundRepo                  // directory contains .git
	EventError                      // error during walk
	EventSkipped                    // directory skipped
)

// String returns the string representation of WalkEvent
func (e WalkEvent) String() string {
	switch e {
	case EventEnterDir:
		return "EnterDir"
	case EventFoundRepo:
		return "FoundRepo"
	case EventError:
		return "Error"
	case EventSkipped:
		return "Skipped"
	default:
		return "Unknown"
	}
}

// DiscoveryResult represents the result of a discovery operation
type DiscoveryResult struct {
	Discovered int                 // number of projects discovered
	Errors     []DiscoveryError    // errors encountered during discovery
	RootStats  map[string]RootStat // per-root statistics
}

// RootStat represents statistics for a single root directory
type RootStat struct {
	Discovered int   // number of projects discovered in this root
	Errors     int   // number of errors in this root
	DurationMs int64 // time taken to process this root (milliseconds)
}

// DiscoveryError represents an error encountered during discovery
type DiscoveryError struct {
	RootPath string // root directory being processed
	DirPath  string // specific directory that caused the error
	Err      error  // underlying error
}

// Error returns the error message
func (e *DiscoveryError) Error() string {
	if e.DirPath != "" {
		return fmt.Sprintf("discovery error in root %s at %s: %v", e.RootPath, e.DirPath, e.Err)
	}
	return fmt.Sprintf("discovery error in root %s: %v", e.RootPath, e.Err)
}

// Unwrap returns the underlying error
func (e *DiscoveryError) Unwrap() error {
	return e.Err
}

// WalkCallback is the callback function invoked during directory walking
type WalkCallback func(path string, event WalkEvent, err error) error

// ConcurrentDiscoveryError is returned when DiscoverProjects is called concurrently
type ConcurrentDiscoveryError struct {
	CurrentOperation string // description of current operation
}

// Error returns the error message
func (e *ConcurrentDiscoveryError) Error() string {
	return fmt.Sprintf("concurrent discovery not allowed: %s already in progress", e.CurrentOperation)
}

// Unwrap returns the underlying error (nil for this error type)
func (e *ConcurrentDiscoveryError) Unwrap() error {
	return nil
}
