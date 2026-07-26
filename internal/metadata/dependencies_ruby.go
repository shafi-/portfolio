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
				deps = append(deps, models.Dependency{Name: name, Manager: "bundler"})
			}
		}
	}
	return deps, nil
}
