package metadata

import (
	"strings"

	"project-dash/pkg/models"
)

// goModParser parses go.mod (Go modules).
type goModParser struct{}

func (goModParser) Filename() string { return "go.mod" }

func (goModParser) Parse(content []byte) ([]models.Dependency, error) {
	var deps []models.Dependency
	inBlock := false

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "module ") || strings.HasPrefix(line, "go ") {
			continue
		}
		if strings.HasPrefix(line, "require (") {
			inBlock = true
			continue
		}
		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 1 && fields[0] != "" && !strings.HasPrefix(fields[0], "//") {
				deps = append(deps, models.Dependency{Name: fields[0], Manager: "go_mod"})
			}
			continue
		}
		if strings.HasPrefix(line, "require ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				deps = append(deps, models.Dependency{Name: fields[1], Manager: "go_mod"})
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && !strings.HasPrefix(fields[0], "//") {
			deps = append(deps, models.Dependency{Name: fields[0], Manager: "go_mod"})
		}
	}

	return deps, nil
}
