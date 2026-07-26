package metadata

import "project-dash/pkg/models"

// ManifestParser extracts dependencies from one manifest file type
// (package.json, go.mod, requirements.txt, ...). Implementations are stateless
// value types, registered once in defaultManifestParsers.
//
// To add an ecosystem: create a new file (e.g. dependencies_php.go) with a
// parser type, then append one entry to defaultManifestParsers. No edits to the
// DetectDependencies dispatcher are needed.
//
// Parse receives the manifest bytes (not the path) so parsers are pure and
// unit-testable without a filesystem — os.ReadFile happens once, in the
// dispatcher. If a future parser needs path context (e.g. npm workspaces
// resolving nested manifests), widen the interface to Parse(path string,
// content []byte).
type ManifestParser interface {
	// Filename is the manifest basename to match on (e.g. "go.mod").
	Filename() string
	// Parse extracts dependencies from the manifest's raw bytes.
	Parse(content []byte) ([]models.Dependency, error)
}

// defaultManifestParsers is the built-in registry, one entry per supported
// ecosystem manifest. Mirrors the catalog Pattern A used by frameworks_data.go,
// capabilities_data.go, and maturity_data.go.
var defaultManifestParsers = []ManifestParser{
	npmParser{},
	goModParser{},
	pipRequirementsParser{},
	pipPyprojectParser{},
	cargoParser{},
	bundlerParser{},
	mavenParser{},
	gradleParser{},
}

// DefaultManifestParsers returns a defensive copy of the default manifest-parser
// registry. Callers may append their own ManifestParser implementations to the
// returned slice and pass it to DetectDependencies without mutating the
// package-level default.
func DefaultManifestParsers() []ManifestParser {
	out := make([]ManifestParser, len(defaultManifestParsers))
	copy(out, defaultManifestParsers)
	return out
}
