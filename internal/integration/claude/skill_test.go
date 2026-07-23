package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillOperations(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")

	integration := &ClaudeCodeIntegration{
		config: ClaudeConfig{
			SkillsDir: skillsDir,
		},
		logger: nil,
		store:  nil,
		mcp:    nil,
	}

	t.Run("install skill to non-existent directory", func(t *testing.T) {
		err := integration.installSkill()
		if err != nil {
			t.Fatalf("installSkill failed: %v", err)
		}

		skillPath := integration.skillPath()
		if _, err := os.Stat(skillPath); err != nil {
			t.Fatalf("skill file not created: %v", err)
		}

		content, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("read skill file failed: %v", err)
		}

		if len(content) == 0 {
			t.Error("skill file is empty")
		}

		expectedContent := "# Portfolio — Claude Code Skill"
		if string(content)[:len(expectedContent)] != expectedContent {
			t.Error("skill file content doesn't start with expected header")
		}
	})

	t.Run("overwrite existing skill", func(t *testing.T) {
		skillPath := integration.skillPath()
		oldContent := []byte("old content")
		if err := os.WriteFile(skillPath, oldContent, 0644); err != nil {
			t.Fatalf("write old skill failed: %v", err)
		}

		err := integration.installSkill()
		if err != nil {
			t.Fatalf("installSkill failed: %v", err)
		}

		newContent, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("read skill file failed: %v", err)
		}

		if string(newContent) == string(oldContent) {
			t.Error("skill file was not overwritten")
		}

		expectedContent := "# Portfolio — Claude Code Skill"
		if string(newContent)[:len(expectedContent)] != expectedContent {
			t.Error("skill file content doesn't start with expected header after overwrite")
		}
	})

	t.Run("remove skill file", func(t *testing.T) {
		err := integration.removeSkill()
		if err != nil {
			t.Fatalf("removeSkill failed: %v", err)
		}

		skillPath := integration.skillPath()
		if _, err := os.Stat(skillPath); err == nil {
			t.Error("skill file still exists after removal")
		} else if !os.IsNotExist(err) {
			t.Fatalf("unexpected error checking skill file: %v", err)
		}
	})

	t.Run("remove non-existent skill is idempotent", func(t *testing.T) {
		err := integration.removeSkill()
		if err != nil {
			t.Fatalf("removeSkill on non-existent file failed: %v", err)
		}
	})
}

func TestSkillPath(t *testing.T) {
	tempDir := t.TempDir()
	skillsDir := filepath.Join(tempDir, "skills")

	integration := &ClaudeCodeIntegration{
		config: ClaudeConfig{
			SkillsDir: skillsDir,
		},
	}

	expectedPath := filepath.Join(skillsDir, "portfolio.md")
	actualPath := integration.skillPath()

	if actualPath != expectedPath {
		t.Errorf("expected skill path %s, got %s", expectedPath, actualPath)
	}
}
