package indexer

import (
	"os"
	"path/filepath"
	"strings"

	"project-dash/pkg/models"
)

type DocDiscoverer struct {
	maxDepth int
}

func NewDocDiscoverer() *DocDiscoverer {
	return &DocDiscoverer{maxDepth: 50}
}

var readmeVariants = []string{
	"README.md", "README.rst", "README.txt", "README",
}

func (d *DocDiscoverer) FindREADME(rootPath string) []DocFile {
	for _, variant := range readmeVariants {
		absPath := filepath.Join(rootPath, variant)
		info, err := os.Stat(absPath)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		if info.IsDir() {
			continue
		}
		return []DocFile{{
			RelPath: variant,
			AbsPath: absPath,
			Kind:    models.DocKindREADME,
		}}
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)

		var stem, ext string
		if dot := strings.LastIndex(lower, "."); dot >= 0 {
			stem = lower[:dot]
			ext = lower[dot+1:]
		} else {
			stem = lower
		}

		if stem != "readme" {
			continue
		}

		if ext != "" && ext != "md" && ext != "rst" && ext != "txt" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}

		return []DocFile{{
			RelPath: name,
			AbsPath: filepath.Join(rootPath, name),
			Kind:    models.DocKindREADME,
		}}
	}

	return nil
}
