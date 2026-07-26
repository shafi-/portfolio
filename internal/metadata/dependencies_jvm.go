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
			var ver, kind string
			// The version sits in the same <dependency> block, after
			// </artifactId> and before </dependency>.
			search := text
			if blockEnd := strings.Index(text, "</dependency>"); blockEnd != -1 {
				search = text[:blockEnd]
			}
			if vStart := strings.Index(search, "<version>"); vStart != -1 {
				vStart += len("<version>")
				if vEnd := strings.Index(search[vStart:], "</version>"); vEnd != -1 {
					ver, kind = parseVersionSpec(search[vStart : vStart+vEnd])
				}
			}
			deps = append(deps, models.Dependency{
				Name: groupID + ":" + artifactID, Manager: "maven",
				Version: ver, VersionType: kind,
			})
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
				var ver, kind string
				// Gradle coordinate form group:name:version — the final ':'
				// segment is the version (also covers name:version).
				if segs := strings.Split(dep, ":"); len(segs) > 1 {
					ver, kind = parseVersionSpec(segs[len(segs)-1])
				}
				deps = append(deps, models.Dependency{Name: dep, Manager: "gradle", Version: ver, VersionType: kind})
			}
		}
	}

	return deps, nil
}
