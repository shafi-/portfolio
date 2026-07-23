package fs

import (
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
