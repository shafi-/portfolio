package metadata

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"project-dash/pkg/models"
)

// DetectDependencies walks root and dispatches each manifest file to the
// ManifestParser registered for its filename, then returns the merged,
// deduplicated dependency list plus a short comma-joined summary of the first
// 10 names.
//
// walker defaults to a nil-logger FileWalker; parsers defaults to
// DefaultManifestParsers(). To add an ecosystem, append a ManifestParser to the
// passed slice (or to defaultManifestParsers) — no edits here are required.
// Read or parse errors for an individual manifest are tolerated (that file is
// skipped), matching the prior behavior.
func DetectDependencies(root string, walker *FileWalker, parsers []ManifestParser) ([]models.Dependency, *string, error) {
	if walker == nil {
		walker = NewFileWalker(nil)
	}
	if parsers == nil {
		parsers = DefaultManifestParsers()
	}

	byFilename := make(map[string]ManifestParser, len(parsers))
	for _, p := range parsers {
		byFilename[p.Filename()] = p
	}

	allDeps := make(map[string]models.Dependency)

	err := walker.Walk(root, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}

		parser, ok := byFilename[filepath.Base(path)]
		if !ok {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		deps, err := parser.Parse(content)
		if err != nil {
			return nil
		}

		for _, d := range deps {
			key := d.Name + "|" + d.Manager + "|" + d.Scope
			allDeps[key] = d
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if len(allDeps) == 0 {
		return nil, nil, nil
	}

	var depList []models.Dependency
	for _, d := range allDeps {
		depList = append(depList, d)
	}

	sort.Slice(depList, func(i, j int) bool {
		if depList[i].Manager != depList[j].Manager {
			return depList[i].Manager < depList[j].Manager
		}
		return depList[i].Name < depList[j].Name
	})

	var names []string
	for i, d := range depList {
		if i >= 10 {
			break
		}
		names = append(names, d.Name)
	}
	summary := strings.Join(names, ", ")

	return depList, &summary, nil
}
