package metadata

import (
	"strings"

	"project-dash/pkg/models"
)

// mavenParser parses pom.xml (Maven). Each <dependency> becomes a single
// dependency named "groupId:artifactId".
type mavenParser struct{}

func (mavenParser) Filename() string { return "pom.xml" }

func (mavenParser) Parse(content []byte) ([]models.Dependency, error) {
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

// gradleParser parses build.gradle (Gradle) implementation/api/compile
// declarations.
type gradleParser struct{}

func (gradleParser) Filename() string { return "build.gradle" }

func (gradleParser) Parse(content []byte) ([]models.Dependency, error) {
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
