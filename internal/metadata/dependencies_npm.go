package metadata

import (
	"encoding/json"

	"project-dash/pkg/models"
)

// npmParser parses package.json (npm). Production dependencies get scope
// "prod"; devDependencies get scope "dev".
type npmParser struct{}

func (npmParser) Filename() string { return "package.json" }

func (npmParser) Parse(content []byte) ([]models.Dependency, error) {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	// Tolerate malformed manifests: a JSON error yields no deps (the dispatcher
	// skips this file), preserving the pre-refactor "corrupted → nil deps"
	// behavior relied on by TestDetectDependencies_CorruptedManifest.
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, nil
	}

	var deps []models.Dependency
	for name, spec := range pkg.Dependencies {
		ver, kind := parseVersionSpec(spec)
		deps = append(deps, models.Dependency{Name: name, Manager: "npm", Scope: "prod", Version: ver, VersionType: kind})
	}
	for name, spec := range pkg.DevDependencies {
		ver, kind := parseVersionSpec(spec)
		deps = append(deps, models.Dependency{Name: name, Manager: "npm", Scope: "dev", Version: ver, VersionType: kind})
	}
	return deps, nil
}
