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
			if capabilityMatch(name, strings.ToLower(m.Pattern)) {
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

// capabilityMatch reports whether the (lower-cased) dependency name matches the
// (lower-cased) marker pattern. Short patterns (3 chars or fewer — e.g. "pg",
// "ent", "jwt", "nsq") are matched on a word boundary so they don't fire inside
// unrelated names like "client", "moment", or "jpg". Longer patterns are
// distinctive enough to match as a plain substring, which preserves legitimate
// prefixed variants such as "ioredis" -> redis.
func capabilityMatch(name, pattern string) bool {
	if len(pattern) <= 3 {
		return boundaryContains(name, pattern)
	}
	return strings.Contains(name, pattern)
}

// boundaryContains reports whether sub occurs in s at a position bounded by a
// non-alphanumeric character (or the start/end of s), so e.g. "ent" matches the
// standalone token "ent" but not the interior of "client". s and sub are
// expected to be already lower-cased.
func boundaryContains(s, sub string) bool {
	for off := 0; off < len(s); {
		idx := strings.Index(s[off:], sub)
		if idx < 0 {
			return false
		}
		pos := off + idx
		end := pos + len(sub)
		beforeOK := pos == 0 || !isAlphaNum(s[pos-1])
		afterOK := end == len(s) || !isAlphaNum(s[end])
		if beforeOK && afterOK {
			return true
		}
		off = pos + 1
	}
	return false
}

func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}
