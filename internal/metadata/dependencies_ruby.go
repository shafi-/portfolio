package metadata

import (
	"strings"

	"project-dash/pkg/models"
)

// bundlerParser parses Gemfile (Bundler / Ruby) gem declarations.
type bundlerParser struct{}

func (bundlerParser) Filename() string { return "Gemfile" }

func (bundlerParser) Parse(content []byte) ([]models.Dependency, error) {
	var deps []models.Dependency
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gem ") {
			start := strings.IndexAny(line, "'\"")
			if start == -1 {
				continue
			}
			end := strings.IndexAny(line[start+1:], "'\"")
			if end == -1 {
				continue
			}
			name := line[start+1 : start+1+end]
			if name != "" {
				var ver, kind string
				// A second quoted token after the name is the version spec:
				// gem "rails", "7.0.0" or gem "rails", "~> 7.0".
				rest := line[start+1+end+1:]
				if vStart := strings.IndexAny(rest, "'\""); vStart != -1 {
					if vEnd := strings.IndexAny(rest[vStart+1:], "'\""); vEnd != -1 {
						ver, kind = parseVersionSpec(rest[vStart+1 : vStart+1+vEnd])
					}
				}
				deps = append(deps, models.Dependency{Name: name, Manager: "bundler", Version: ver, VersionType: kind})
			}
		}
	}
	return deps, nil
}
