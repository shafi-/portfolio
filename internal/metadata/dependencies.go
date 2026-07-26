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

	// Build a filename → parser index. If two parsers share a Filename() (only
	// possible when a caller appends a custom one alongside the defaults), the
	// later entry wins — an intentional override mechanism.
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
		if depList[i].Name != depList[j].Name {
			return depList[i].Name < depList[j].Name
		}
		// Tiebreak identical (manager, name) by scope so the row that wins the
		// downstream INSERT OR IGNORE (UNIQUE on project_id, name, manager — no
		// scope) is deterministic across scans: prod before dev.
		return scopeRank(depList[i].Scope) < scopeRank(depList[j].Scope)
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

// scopeRank orders dependency scopes so "prod" sorts before "dev" (and any other
// value). This makes the surviving row deterministic when a package is declared
// in both dependencies and devDependencies: the dependencies table's UNIQUE
// constraint is (project_id, name, manager) without scope, so INSERT OR IGNORE
// keeps whichever row is inserted first — the prod row, because it sorts first.
func scopeRank(scope string) int {
	switch scope {
	case "prod", "":
		return 0
	case "dev":
		return 1
	default:
		return 2
	}
}
