package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DetectFrameworks(root string, walker *FileWalker, markers []FrameworkMarker) (*string, error) {
	if walker == nil {
		walker = NewFileWalker(nil)
	}
	if markers == nil {
		markers = DefaultFrameworkMarkers()
	}

	markerIndex := buildMarkerIndex(markers)
	detected := make(map[string]bool)

	err := walker.Walk(root, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		mfMarkers, ok := markerIndex[name]
		if !ok {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		for _, m := range mfMarkers {
			if detected[m.Name] {
				continue
			}
			if frameworkMatch(content, m, name) {
				detected[m.Name] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(detected) == 0 {
		return nil, nil
	}

	var names []string
	for name := range detected {
		names = append(names, name)
	}
	sort.Strings(names)
	result := strings.Join(names, ", ")
	return &result, nil
}

func buildMarkerIndex(markers []FrameworkMarker) map[string][]FrameworkMarker {
	idx := make(map[string][]FrameworkMarker)
	for _, m := range markers {
		idx[m.Manifest] = append(idx[m.Manifest], m)
	}
	return idx
}

func frameworkMatch(content []byte, m FrameworkMarker, manifestName string) bool {
	switch manifestName {
	case "package.json":
		return matchPackageJSON(content, m.Pattern)
	case "go.mod":
		return matchGoMod(content, m.Pattern)
	case "requirements.txt":
		return matchRequirementsTxt(content, m.Pattern)
	case "pyproject.toml":
		return matchPyprojectToml(content, m.Pattern)
	case "Cargo.toml":
		return matchCargoToml(content, m.Pattern)
	case "Gemfile":
		return matchGemfile(content, m.Pattern)
	case "pom.xml":
		return matchPomXML(content, m.Pattern)
	case "build.gradle":
		return matchBuildGradle(content, m.Pattern)
	}
	return false
}

func matchPackageJSON(content []byte, pattern string) bool {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return false
	}
	if _, ok := pkg.Dependencies[pattern]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies[pattern]; ok {
		return true
	}
	return false
}

func matchGoMod(content []byte, pattern string) bool {
	inBlock := false
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
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
			if len(fields) >= 1 && matchGoModDep(fields[0], pattern) {
				return true
			}
			continue
		}
		if strings.HasPrefix(line, "require ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && matchGoModDep(fields[1], pattern) {
				return true
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && matchGoModDep(fields[0], pattern) {
			return true
		}
	}
	return false
}

func matchGoModDep(name string, pattern string) bool {
	return name == pattern || strings.HasPrefix(name, pattern+"/")
}

func matchRequirementsTxt(content []byte, pattern string) bool {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "==")
		name := strings.TrimSpace(parts[0])
		name = strings.Split(name, "[")[0]
		if strings.EqualFold(name, pattern) {
			return true
		}
	}
	return false
}

func matchPyprojectToml(content []byte, pattern string) bool {
	return strings.Contains(strings.ToLower(string(content)), strings.ToLower(pattern))
}

func matchCargoToml(content []byte, pattern string) bool {
	return strings.Contains(string(content), pattern)
}

func matchGemfile(content []byte, pattern string) bool {
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gem '"+pattern+"'") || strings.HasPrefix(line, `gem "`+pattern+`"`) {
			return true
		}
	}
	return false
}

func matchPomXML(content []byte, pattern string) bool {
	return strings.Contains(string(content), pattern)
}

func matchBuildGradle(content []byte, pattern string) bool {
	return strings.Contains(string(content), pattern)
}
