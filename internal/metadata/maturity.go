package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// DetectMaturity computes a maturity score from file-presence signals at the
// project root. It returns the weighted score and a JSON object of detected
// artifact kinds (e.g. {"ci":true,"readme":true}). JSON map keys are sorted
// alphabetically by encoding/json, so output is deterministic.
func DetectMaturity(root string, artifacts []MaturityArtifact) (int, string, error) {
	if artifacts == nil {
		artifacts = DefaultMaturityArtifacts()
	}

	indicators := make(map[string]bool)
	score := 0
	for _, a := range artifacts {
		if pathExists(root, a.Paths) {
			indicators[a.Kind] = true
			score += a.Weight
		}
	}

	if len(indicators) == 0 {
		return 0, "", nil
	}

	b, err := json.Marshal(indicators)
	if err != nil {
		return score, "", err
	}
	return score, string(b), nil
}

// pathExists reports whether any of the relative paths exists at root. Paths
// containing glob metacharacters (*?[]) are matched with filepath.Glob.
func pathExists(root string, rels []string) bool {
	for _, rel := range rels {
		full := filepath.Join(root, rel)
		if strings.ContainsAny(rel, "*?[") {
			if m, _ := filepath.Glob(full); len(m) > 0 {
				return true
			}
			continue
		}
		if _, err := os.Stat(full); err == nil {
			return true
		}
	}
	return false
}
