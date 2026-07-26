package opencode

import (
	_ "embed"
	"os"
	"path/filepath"

	"project-dash/internal/integration"
)

//go:embed skill.md
var skillContent string

func (o *OpenCodeIntegration) installSkill() error {
	if err := os.MkdirAll(o.skillDir(), 0755); err != nil {
		return err
	}

	// skillContent is a template carrying the {{SKILL_COMMON}} placeholder (and
	// {{ANALYZER}} inside the shared body); expand both before writing so the
	// installed file is the fully resolved skill.
	rendered := integration.RenderSkill(skillContent, AnalyzerID)
	return os.WriteFile(o.skillPath(), []byte(rendered), 0644)
}

func (o *OpenCodeIntegration) removeSkill() error {
	// Remove the whole skill directory (name dir + SKILL.md).
	if err := os.RemoveAll(o.skillDir()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// skillDir is the opencode skill directory; its name must match the
// frontmatter "name" field: <skills>/portfolio
func (o *OpenCodeIntegration) skillDir() string {
	return filepath.Join(o.config.SkillsDir, "portfolio")
}

// skillPath is the file opencode loads: <skills>/portfolio/SKILL.md
func (o *OpenCodeIntegration) skillPath() string {
	return filepath.Join(o.skillDir(), "SKILL.md")
}
