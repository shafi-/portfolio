package metadata

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"project-dash/pkg/models"
)

func DetectDependencies(root string, walker *FileWalker) ([]models.Dependency, *string, error) {
	if walker == nil {
		walker = NewFileWalker(nil)
	}

	allDeps := make(map[string]models.Dependency)

	err := walker.Walk(root, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}

		name := filepath.Base(path)
		var deps []models.Dependency
		var err error

		switch name {
		case "package.json":
			deps, err = parsePackageJSON(path)
		case "go.mod":
			deps, err = parseGoMod(path)
		case "requirements.txt":
			deps, err = parseRequirementsTxt(path)
		case "pyproject.toml":
			deps, err = parsePyprojectToml(path)
		case "Cargo.toml":
			deps, err = parseCargoToml(path)
		case "Gemfile":
			deps, err = parseGemfile(path)
		case "pom.xml":
			deps, err = parsePomXML(path)
		case "build.gradle":
			deps, err = parseBuildGradle(path)
		default:
			return nil
		}

		if err != nil {
			return nil
		}

		for _, d := range deps {
			key := d.Name + "|" + d.Manager
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

func parsePackageJSON(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, nil
	}

	var deps []models.Dependency
	for name := range pkg.Dependencies {
		deps = append(deps, models.Dependency{Name: name, Manager: "npm"})
	}
	return deps, nil
}

func parseGoMod(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

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

func parseRequirementsTxt(path string) ([]models.Dependency, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []models.Dependency
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		parts := strings.Split(line, "==")
		name := strings.TrimSpace(parts[0])
		name = strings.Split(name, "[")[0]
		if name != "" {
			deps = append(deps, models.Dependency{Name: name, Manager: "pip"})
		}
	}

	return deps, scanner.Err()
}

func parsePyprojectToml(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

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
				deps = append(deps, models.Dependency{Name: name, Manager: "pip"})
			}
		}
	}

	return deps, nil
}

func parseCargoToml(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

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

func parseGemfile(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

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

func parsePomXML(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(content)
	var deps []models.Dependency

	for {
		start := strings.Index(text, "<groupId>")
		if start == -1 {
			break
		}
		start += len("<groupId>")
		end := strings.Index(text[start:], "</groupId>")
		if end == -1 {
			break
		}
		groupID := text[start : start+end]
		text = text[start+end:]

		artStart := strings.Index(text, "<artifactId>")
		if artStart == -1 {
			break
		}
		artStart += len("<artifactId>")
		artEnd := strings.Index(text[artStart:], "</artifactId>")
		if artEnd == -1 {
			break
		}
		artifactID := text[artStart : artStart+artEnd]
		text = text[artStart+artEnd:]

		if groupID != "" && artifactID != "" {
			deps = append(deps, models.Dependency{Name: groupID + ":" + artifactID, Manager: "maven"})
		}
	}

	return deps, nil
}

func parseBuildGradle(path string) ([]models.Dependency, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var deps []models.Dependency
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "implementation ") || strings.HasPrefix(line, "api ") || strings.HasPrefix(line, "compile ") {
			start := strings.IndexAny(line, "'\"")
			if start == -1 {
				continue
			}
			end := strings.IndexAny(line[start+1:], "'\"")
			if end == -1 {
				continue
			}
			dep := line[start+1 : start+1+end]
			if dep != "" {
				deps = append(deps, models.Dependency{Name: dep, Manager: "gradle"})
			}
		}
	}

	return deps, nil
}
