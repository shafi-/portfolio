package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"project-dash/internal/fs"
	"project-dash/internal/logging"
	"project-dash/pkg/models"
)

// IgnoreMatcher defines the interface for pattern-based ignore matching
type IgnoreMatcher interface {
	// ShouldIgnore returns true if the given path should be ignored
	ShouldIgnore(path string, isDir bool) bool
	// Match returns true if the given name matches any ignore pattern
	Match(name string) bool
}

// DefaultIgnoreMatcher implements pattern-based ignore matching
type DefaultIgnoreMatcher struct {
	patterns []string
	logger   *logging.Logger
}

// NewDefaultIgnoreMatcher creates a new ignore matcher with default patterns
func NewDefaultIgnoreMatcher(customPatterns []string, logger *logging.Logger) *DefaultIgnoreMatcher {
	// Start with default patterns
	patterns := []string{
		"node_modules",
		"vendor",
		".venv",
		"target",
		"build",
		"dist",
	}

	// Add custom patterns from config
	patterns = append(patterns, customPatterns...)

	return &DefaultIgnoreMatcher{
		patterns: patterns,
		logger:   logger,
	}
}

// ShouldIgnore returns true if the given path should be ignored
func (m *DefaultIgnoreMatcher) ShouldIgnore(path string, isDir bool) bool {
	// Extract the directory/file name from the path
	name := filepath.Base(path)

	// Check each pattern
	for _, pattern := range m.patterns {
		if m.matchPattern(name, pattern) {
			m.logger.Debug("Ignoring directory",
				models.Field{Key: "path", Value: path},
				models.Field{Key: "pattern", Value: pattern},
			)
			return true
		}
	}

	return false
}

// Match returns true if the given name matches any ignore pattern
func (m *DefaultIgnoreMatcher) Match(name string) bool {
	for _, pattern := range m.patterns {
		if m.matchPattern(name, pattern) {
			return true
		}
	}
	return false
}

// matchPattern performs pattern matching with support for wildcards
func (m *DefaultIgnoreMatcher) matchPattern(name string, pattern string) bool {
	// Direct match
	if name == pattern {
		return true
	}

	// Simple wildcard support: pattern*
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	// Simple wildcard support: *pattern
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	return false
}

// Walker performs breadth-first search directory walking
type Walker struct {
	fs            WalkerFS
	detector      *Detector
	ignoreMatcher IgnoreMatcher
	seenInodes    map[uint64]string
	inodeMutex    sync.Mutex
	maxDepth      int
	logger        *logging.Logger
}

// WalkerFS interface for walker operations
type WalkerFS interface {
	ReadDir(path string) ([]os.DirEntry, error)
	Lstat(path string) (os.FileInfo, error)
}

// NewWalker creates a new BFS walker
func NewWalker(filesystem fs.Filesystem, detector *Detector, ignorePaths []string, maxDepth int, logger *logging.Logger) *Walker {
	// Create ignore matcher with default patterns and custom patterns from config
	ignoreMatcher := NewDefaultIgnoreMatcher(ignorePaths, logger)

	// Create adapter for the filesystem interface
	walkerFS := &walkerFSAdapter{fs: filesystem}

	return &Walker{
		fs:            walkerFS,
		detector:      detector,
		ignoreMatcher: ignoreMatcher,
		seenInodes:    make(map[uint64]string),
		maxDepth:      maxDepth,
		logger:        logger,
	}
}

// walkerFSAdapter adapts fs.Filesystem to WalkerFS
type walkerFSAdapter struct {
	fs fs.Filesystem
}

func (a *walkerFSAdapter) ReadDir(path string) ([]os.DirEntry, error) {
	return a.fs.ReadDir(path)
}

func (a *walkerFSAdapter) Lstat(path string) (os.FileInfo, error) {
	return a.fs.Lstat(path)
}

// normalizePath performs comprehensive path normalization
func normalizePath(path string) (string, error) {
	// Check for empty path
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Expand home directory if present
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot expand home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Clean the path (removes redundant separators, . , .. etc)
	// This handles trailing slashes and other path issues
	path = filepath.Clean(path)

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("cannot convert to absolute path: %w", err)
		}
		path = abs
	}

	return path, nil
}

// Walk performs a breadth-first search of the directory tree
func (w *Walker) Walk(ctx context.Context, root string, callback WalkCallback) error {
	// Normalize root path with comprehensive handling
	normalizedRoot, err := normalizePath(root)
	if err != nil {
		return fmt.Errorf("path normalization failed for %s: %w", root, err)
	}
	root = normalizedRoot

	// Check if root exists
	info, err := w.fs.Lstat(root)
	if err != nil {
		return fmt.Errorf("cannot access root %s: %w", root, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("root is not a directory: %s", root)
	}

	// Track inode to prevent symlink loops
	if err := w.trackInode(root, info); err != nil {
		callback(root, EventSkipped, err)
		return nil
	}

	// Check if root itself is a repository
	if w.detector.IsGitRepository(root) {
		if err := callback(root, EventFoundRepo, nil); err != nil {
			return err
		}
		return nil
	}

	// BFS queue
	queue := []walkEntry{{path: root, depth: 0}}

	for len(queue) > 0 {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Dequeue
		current := queue[0]
		queue = queue[1:]

		// Check max depth
		if w.maxDepth > 0 && current.depth >= w.maxDepth {
			continue
		}

		// Emit enter directory event
		if err := callback(current.path, EventEnterDir, nil); err != nil {
			return err
		}

		// Read directory entries
		entries, err := w.fs.ReadDir(current.path)
		if err != nil {
			// Handle permission errors gracefully
			if err := callback(current.path, EventError, err); err != nil {
				return err
			}
			continue
		}

		// Process each entry
		for _, entry := range entries {
			entryPath := filepath.Join(current.path, entry.Name())

			// Skip if ignored
			if w.ignoreMatcher.Match(entry.Name()) {
				callback(entryPath, EventSkipped, nil)
				continue
			}

			// Get file info
			info, err := w.fs.Lstat(entryPath)
			if err != nil {
				callback(entryPath, EventError, err)
				continue
			}

			// Skip regular files and special files
			if !entry.IsDir() {
				continue
			}

			// Track inode to prevent symlink loops
			if err := w.trackInode(entryPath, info); err != nil {
				callback(entryPath, EventSkipped, err)
				continue
			}

			// Check if this is a Git repository
			if w.detector.IsGitRepository(entryPath) {
				if err := callback(entryPath, EventFoundRepo, nil); err != nil {
					return err
				}
				// Continue recursion to support nested repositories (monorepos)
				// This allows discovery of repos within repos
			}

			// Add to queue for BFS processing
			queue = append(queue, walkEntry{
				path:  entryPath,
				depth: current.depth + 1,
			})
		}
	}

	return nil
}

// trackInode tracks visited directories to prevent symlink loops
func (w *Walker) trackInode(path string, info os.FileInfo) error {
	// Get inode number (works on Unix systems)
	sys := info.Sys()
	if sys == nil {
		// Cannot get inode info, proceed without tracking
		return nil
	}

	// Try to get inode from different OS types
	var inode uint64
	switch stat := sys.(type) {
	case interface{ Ino() uint64 }:
		inode = stat.Ino()
	default:
		// Cannot get inode, proceed without tracking
		return nil
	}

	w.inodeMutex.Lock()
	defer w.inodeMutex.Unlock()

	// Check if we've seen this inode before
	if prevPath, seen := w.seenInodes[inode]; seen {
		return fmt.Errorf("already visited %s (inode %d)", prevPath, inode)
	}

	// Mark this inode as seen
	w.seenInodes[inode] = path
	return nil
}

// walkEntry represents an entry in the BFS queue
type walkEntry struct {
	path  string
	depth int
}

// ResetInodeTracking clears the inode tracking map
// This should be called before each new walk operation
func (w *Walker) ResetInodeTracking() {
	w.inodeMutex.Lock()
	defer w.inodeMutex.Unlock()
	w.seenInodes = make(map[uint64]string)
}
