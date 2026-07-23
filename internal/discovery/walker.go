package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"project-dash/internal/fs"
)

// Walker performs breadth-first search directory walking
type Walker struct {
	fs          WalkerFS
	detector    *Detector
	ignorePaths map[string]bool
	seenInodes  map[uint64]string
	inodeMutex  sync.Mutex
	maxDepth    int
}

// WalkerFS interface for walker operations
type WalkerFS interface {
	ReadDir(path string) ([]os.DirEntry, error)
	Lstat(path string) (os.FileInfo, error)
}

// NewWalker creates a new BFS walker
func NewWalker(filesystem fs.Filesystem, detector *Detector, ignorePaths []string, maxDepth int) *Walker {
	ignoreMap := make(map[string]bool)
	for _, path := range ignorePaths {
		ignoreMap[path] = true
	}

	// Create adapter for the filesystem interface
	walkerFS := &walkerFSAdapter{fs: filesystem}

	return &Walker{
		fs:          walkerFS,
		detector:    detector,
		ignorePaths: ignoreMap,
		seenInodes:  make(map[uint64]string),
		maxDepth:    maxDepth,
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

// Walk performs a breadth-first search of the directory tree
func (w *Walker) Walk(ctx context.Context, root string, callback WalkCallback) error {
	// Normalize root path
	root = filepath.Clean(root)

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
			if w.ignorePaths[entry.Name()] {
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
				// Don't recurse into repository directories
				continue
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
