package opencode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestIntegration(t *testing.T) *OpenCodeIntegration {
	t.Helper()
	dir := t.TempDir()
	return &OpenCodeIntegration{
		config: OpenCodeConfig{
			ConfigPath: filepath.Join(dir, "opencode.json"),
			SkillsDir:  filepath.Join(dir, "skills"),
			BinaryPath: filepath.Join(dir, "portfolio"),
		},
	}
}

func TestSkillOperations(t *testing.T) {
	integration := newTestIntegration(t)

	t.Run("install skill to non-existent directory", func(t *testing.T) {
		if err := integration.installSkill(); err != nil {
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
			t.Fatal("skill file is empty")
		}
		// opencode skills start with YAML frontmatter.
		if !strings.HasPrefix(string(content), "---\n") {
			t.Errorf("skill file does not start with frontmatter; got %q", firstChars(string(content), 20))
		}
	})

	t.Run("overwrite existing skill", func(t *testing.T) {
		skillPath := integration.skillPath()
		oldContent := []byte("old content")
		if err := os.WriteFile(skillPath, oldContent, 0644); err != nil {
			t.Fatalf("write old skill failed: %v", err)
		}

		if err := integration.installSkill(); err != nil {
			t.Fatalf("installSkill failed: %v", err)
		}

		newContent, err := os.ReadFile(skillPath)
		if err != nil {
			t.Fatalf("read skill file failed: %v", err)
		}
		if string(newContent) == string(oldContent) {
			t.Error("skill file was not overwritten")
		}
	})

	t.Run("remove skill directory", func(t *testing.T) {
		if err := integration.removeSkill(); err != nil {
			t.Fatalf("removeSkill failed: %v", err)
		}
		if _, err := os.Stat(integration.skillDir()); err == nil {
			t.Error("skill directory still exists after removal")
		} else if !os.IsNotExist(err) {
			t.Fatalf("unexpected error checking skill dir: %v", err)
		}
	})

	t.Run("remove non-existent skill is idempotent", func(t *testing.T) {
		if err := integration.removeSkill(); err != nil {
			t.Fatalf("removeSkill on non-existent dir failed: %v", err)
		}
	})
}

func TestSkillPath(t *testing.T) {
	dir := t.TempDir()
	o := &OpenCodeIntegration{
		config: OpenCodeConfig{SkillsDir: filepath.Join(dir, "skills")},
	}

	expectedDir := filepath.Join(dir, "skills", "portfolio")
	if got := o.skillDir(); got != expectedDir {
		t.Errorf("skillDir = %s, want %s", got, expectedDir)
	}
	expectedPath := filepath.Join(expectedDir, "SKILL.md")
	if got := o.skillPath(); got != expectedPath {
		t.Errorf("skillPath = %s, want %s", got, expectedPath)
	}
}

// TestSkillContentRendered verifies the installed skill is fully expanded:
// no {{SKILL_COMMON}}/{{ANALYZER}} placeholders survive, the shared body is
// present, and the per-integration analyzer id is filled in.
func TestSkillContentRendered(t *testing.T) {
	o := newTestIntegration(t)

	if err := o.installSkill(); err != nil {
		t.Fatalf("installSkill failed: %v", err)
	}

	content, err := os.ReadFile(o.skillPath())
	if err != nil {
		t.Fatalf("read installed skill: %v", err)
	}
	got := string(content)

	for _, needle := range []string{"{{SKILL_COMMON}}", "{{ANALYZER}}"} {
		if strings.Contains(got, needle) {
			t.Errorf("installed skill contains unresolved placeholder %s", needle)
		}
	}
	if !strings.HasPrefix(got, "---\n") {
		t.Error("installed skill does not start with frontmatter")
	}
	if !strings.Contains(got, "name: portfolio") {
		t.Error("installed skill frontmatter missing name: portfolio")
	}
	if !strings.Contains(got, "## Tools") {
		t.Error("installed skill is missing the shared ## Tools section")
	}
	if !strings.Contains(got, `analyzer: "opencode"`) {
		t.Error(`installed skill missing expanded analyzer: "opencode"`)
	}
}

func firstChars(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
