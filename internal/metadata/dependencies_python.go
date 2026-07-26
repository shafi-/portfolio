package metadata

import (
	"bufio"
	"bytes"
	"strings"

	"project-dash/pkg/models"
)

// pipRequirementsParser parses requirements.txt (pip).
type pipRequirementsParser struct{}

func (pipRequirementsParser) Filename() string { return "requirements.txt" }

func (pipRequirementsParser) Parse(content []byte) ([]models.Dependency, error) {
	var deps []models.Dependency
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		name, spec := splitNameSpec(line)
		name = strings.TrimSpace(strings.Split(name, "[")[0])
		if name != "" {
			ver, kind := parseVersionSpec(spec)
			deps = append(deps, models.Dependency{Name: name, Manager: "pip", Version: ver, VersionType: kind})
		}
	}

	return deps, scanner.Err()
}

// pipPyprojectParser parses pyproject.toml (pip: Poetry [tool.poetry.dependencies]
// and PEP 621 [project.dependencies] sections).
type pipPyprojectParser struct{}

func (pipPyprojectParser) Filename() string { return "pyproject.toml" }

func (pipPyprojectParser) Parse(content []byte) ([]models.Dependency, error) {
	text := string(content)
	var deps []models.Dependency

	inDeps := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[tool.poetry.dependencies]") || strings.HasPrefix(line, "[project.dependencies]") {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") && inDeps {
			break
		}
		if inDeps && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			name := strings.TrimSpace(parts[0])
			name = strings.Trim(name, "\"")
			if name != "" && name != "python" {
				ver, kind := parseVersionSpec(extractTableVersion(parts[1]))
				deps = append(deps, models.Dependency{Name: name, Manager: "pip", Version: ver, VersionType: kind})
			}
		}
	}

	return deps, nil
}
