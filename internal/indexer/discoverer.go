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

var docsExtensions = map[string]bool{
	".md":   true,
	".rst":  true,
	".txt":  true,
	".adoc": true,
}

var changelogNames = map[string]bool{
	"changelog.md": true,
	"changes.md":   true,
	"history.md":   true,
}

var adrDirs = []string{
	"docs/adr",
	".adr",
	"adr",
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

func (d *DocDiscoverer) FindDocs(rootPath string) []DocFile {
	docsDir := filepath.Join(rootPath, "docs")
	info, err := os.Stat(docsDir)
	if err != nil || !info.IsDir() {
		return nil
	}

	var files []DocFile
	d.walkDocs(docsDir, rootPath, 0, &files)
	return files
}

func (d *DocDiscoverer) walkDocs(dir, rootPath string, depth int, files *[]DocFile) {
	if depth > d.maxDepth {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dir, entry.Name())

		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		if entry.IsDir() {
			d.walkDocs(entryPath, rootPath, depth+1, files)
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !docsExtensions[ext] {
			continue
		}

		if isBinaryFile(entryPath) {
			continue
		}

		relPath, _ := filepath.Rel(rootPath, entryPath)
		*files = append(*files, DocFile{
			RelPath: relPath,
			AbsPath: entryPath,
			Kind:    models.DocKindDOC,
		})
	}
}

func (d *DocDiscoverer) FindADRs(rootPath string) []DocFile {
	var files []DocFile

	for _, adrDir := range adrDirs {
		dirPath := filepath.Join(rootPath, adrDir)
		info, err := os.Stat(dirPath)
		if err != nil || !info.IsDir() {
			continue
		}

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}

			if !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
				continue
			}

			relPath := filepath.Join(adrDir, entry.Name())
			files = append(files, DocFile{
				RelPath: relPath,
				AbsPath: filepath.Join(dirPath, entry.Name()),
				Kind:    models.DocKindADR,
			})
		}
	}

	return files
}

func (d *DocDiscoverer) FindCHANGELOG(rootPath string) []DocFile {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil
	}

	var files []DocFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		lower := strings.ToLower(entry.Name())
		if !changelogNames[lower] {
			continue
		}

		files = append(files, DocFile{
			RelPath: entry.Name(),
			AbsPath: filepath.Join(rootPath, entry.Name()),
			Kind:    models.DocKindCHANGELOG,
		})
	}

	return files
}

func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return false
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}
