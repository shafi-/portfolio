package metadata

import (
	"strings"

	"project-dash/pkg/models"
)

// cargoParser parses Cargo.toml (Cargo / Rust) [dependencies] section.
type cargoParser struct{}

func (cargoParser) Filename() string { return "Cargo.toml" }

func (cargoParser) Parse(content []byte) ([]models.Dependency, error) {
	text := string(content)
	var deps []models.Dependency

	inDeps := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[dependencies]") {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") && inDeps {
			break
		}
		if inDeps && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			name := strings.TrimSpace(parts[0])
			if name != "" {
				deps = append(deps, models.Dependency{Name: name, Manager: "cargo"})
			}
		}
	}

	return deps, nil
}
