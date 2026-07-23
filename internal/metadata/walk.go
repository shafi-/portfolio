package metadata

import (
	"io/fs"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type WalkConfig struct {
	IgnoredDirs   []string
	MaxFiles      int
	FollowSymlink bool
}

type FileWalker struct {
	logger *zap.Logger
}

func NewFileWalker(logger *zap.Logger) *FileWalker {
	return &FileWalker{logger: logger}
}

func (w *FileWalker) Walk(root string, fn func(path string, info os.FileInfo) error) error {
	cfg := WalkConfig{
		IgnoredDirs:   []string{"vendor", "node_modules", ".git", "build", "dist", ".next", "target"},
		MaxFiles:      0,
		FollowSymlink: false,
	}
	return w.WalkWithConfig(root, cfg, fn)
}

func (w *FileWalker) WalkWithConfig(root string, cfg WalkConfig, fn func(path string, info os.FileInfo) error) error {
	fileCount := 0

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				w.logger.Warn("skipped inaccessible path", zap.String("path", path), zap.Error(err))
				return nil
			}
			return err
		}

		if d.IsDir() {
			for _, ignored := range cfg.IgnoredDirs {
				if d.Name() == ignored {
					return fs.SkipDir
				}
			}
			return nil
		}

		if cfg.MaxFiles > 0 && fileCount >= cfg.MaxFiles {
			return fs.SkipAll
		}

		info, err := d.Info()
		if err != nil {
			if os.IsPermission(err) {
				w.logger.Warn("skipped inaccessible file", zap.String("path", path), zap.Error(err))
				return nil
			}
			return err
		}

		if !cfg.FollowSymlink && info.Mode()&os.ModeSymlink != 0 {
			target, err := filepath.EvalSymlinks(path)
			if err != nil {
				w.logger.Warn("failed to eval symlink", zap.String("path", path), zap.Error(err))
				return nil
			}
			if !filepath.HasPrefix(target, root) {
				w.logger.Warn("skipped symlink outside root", zap.String("path", path), zap.String("target", target))
				return nil
			}
		}

		fileCount++
		return fn(path, info)
	})
}
