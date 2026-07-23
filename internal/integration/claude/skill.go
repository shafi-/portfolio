package claude

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed skill.md
var skillContent string

func (c *ClaudeCodeIntegration) installSkill() error {
	if err := os.MkdirAll(c.config.SkillsDir, 0755); err != nil {
		return err
	}

	skillPath := c.skillPath()
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
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
