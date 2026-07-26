package claude

import (
	_ "embed"
	"os"
	"path/filepath"

	"project-dash/internal/integration"
)

//go:embed skill.md
var skillContent string

func (c *ClaudeCodeIntegration) installSkill() error {
	if err := os.MkdirAll(c.config.SkillsDir, 0755); err != nil {
		return err
	}

	// skillContent is a template carrying the {{SKILL_COMMON}} placeholder (and
	// {{ANALYZER}} inside the shared body); expand both before writing so the
	// installed file is the fully resolved skill.
	skillPath := c.skillPath()
	if err := os.WriteFile(skillPath, []byte(integration.RenderSkill(skillContent, "claude-code")), 0644); err != nil {
		return err
	}

	return nil
}

func (c *ClaudeCodeIntegration) removeSkill() error {
	skillPath := c.skillPath()
	if err := os.Remove(skillPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c *ClaudeCodeIntegration) skillPath() string {
	return filepath.Join(c.config.SkillsDir, "portfolio.md")
}
