package metadata

import (
	"sort"
	"strings"

	"project-dash/pkg/models"
)

// DetectCapabilities derives capability categories from a list of parsed
// dependencies. It returns a sorted, comma-joined string of unique categories
// (mirrors framework_summary), or nil if none are detected.
func DetectCapabilities(deps []models.Dependency, markers []CapabilityMarker) (*string, error) {
	if markers == nil {
		markers = DefaultCapabilityMarkers()
	}

	detected := make(map[string]bool)
	for _, d := range deps {
		name := strings.ToLower(d.Name)
		if name == "" {
			continue
		}
		for _, m := range markers {
			if strings.Contains(name, strings.ToLower(m.Pattern)) {
				detected[m.Category] = true
			}
		}
	}

	if len(detected) == 0 {
		return nil, nil
	}

	var cats []string
	for c := range detected {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	result := strings.Join(cats, ", ")
	return &result, nil
}
