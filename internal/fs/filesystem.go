package fs

import (
	"io/fs"
	"os"
)

// Filesystem interface for filesystem operations
// This interface exists to enable testing with mock filesystems
type Filesystem interface {
	ReadDir(path string) ([]os.DirEntry, error)
	Lstat(path string) (os.FileInfo, error)
	Stat(path string) (os.FileInfo, error)
}

// osFilesystem implements Filesystem using actual os calls
type osFilesystem struct{}

// NewOSFilesystem creates a new Filesystem backed by the operating system
func NewOSFilesystem() Filesystem {
	return &osFilesystem{}
}

// ReadDir returns a list of directory entries
func (o *osFilesystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

// Lstat returns file info without following symlinks
func (o *osFilesystem) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

// Stat returns file info following symlinks
func (o *osFilesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// MapFSAdapter adapts testing/fstest.MapFS to Filesystem interface
// This is useful for testing without a real filesystem
type MapFSAdapter struct {
	MapFS fs.FS
}

// ReadDir implements Filesystem.ReadDir for MapFS
func (a *MapFSAdapter) ReadDir(path string) ([]os.DirEntry, error) {
	entries, err := fs.ReadDir(a.MapFS, path)
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// Lstat implements Filesystem.Lstat for MapFS
// Note: MapFS doesn't support symlinks, so Lstat behaves like Stat
func (a *MapFSAdapter) Lstat(path string) (os.FileInfo, error) {
	return a.Stat(path)
}

// Stat implements Filesystem.Stat for MapFS
func (a *MapFSAdapter) Stat(path string) (os.FileInfo, error) {
	// Handle empty path as root
	if path == "." || path == "" {
		path = "."
	}

	info, err := fs.Stat(a.MapFS, path)
	if err != nil {
		return nil, err
	}
	return info, nil
}
