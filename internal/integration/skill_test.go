package integration

import (
	_ "embed"
	"strings"
	"testing"
)

func TestRenderSkill(t *testing.T) {
	t.Run("expands common body and analyzer identity", func(t *testing.T) {
		tmpl := "# Portfolio — Test\n\nIntro.\n\n" + SkillCommonPlaceholder + "\n"
		got := RenderSkill(tmpl, "test-analyzer")

		if strings.Contains(got, SkillCommonPlaceholder) {
			t.Errorf("output still contains %q placeholder", SkillCommonPlaceholder)
		}
		if strings.Contains(got, AnalyzerPlaceholder) {
			t.Errorf("output still contains %q placeholder", AnalyzerPlaceholder)
		}
		if !strings.HasPrefix(got, "# Portfolio — Test") {
			t.Errorf("wrapper header lost; got prefix %q", firstLine(got))
		}
		if !strings.Contains(got, "## Tools") {
			t.Error("shared body (## Tools) was not injected")
		}
		if !strings.Contains(got, `analyzer: "test-analyzer"`) {
			t.Error("analyzer identity was not substituted into shared body")
		}
	})

	t.Run("per-integration analyzer value", func(t *testing.T) {
		tmpl := "# X\n\n" + SkillCommonPlaceholder + "\n"
		if got := RenderSkill(tmpl, "opencode"); !strings.Contains(got, `analyzer: "opencode"`) {
			t.Error("expected opencode analyzer in rendered output")
		}
		if got := RenderSkill(tmpl, "claude-code"); !strings.Contains(got, `analyzer: "claude-code"`) {
			t.Error("expected claude-code analyzer in rendered output")
		}
	})

	t.Run("shared body is embedded and non-empty", func(t *testing.T) {
		if SkillCommonContent == "" {
			t.Fatal("SkillCommonContent is empty; skill_common.md not embedded")
		}
		if !strings.Contains(SkillCommonContent, "## Three-Tier Knowledge Protocol") {
			t.Error("skill_common.md missing expected section")
		}
		// The shared body must reference the analyzer placeholder, not a hard-coded id.
		if !strings.Contains(SkillCommonContent, AnalyzerPlaceholder) {
			t.Errorf("skill_common.md must contain %q so each integration gets its analyzer id", AnalyzerPlaceholder)
		}
		if strings.Contains(SkillCommonContent, "claude-code") || strings.Contains(SkillCommonContent, "opencode") {
			t.Error("skill_common.md must not hard-code an analyzer id; use {{ANALYZER}}")
		}
	})
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// Real, shipped skill templates for each integration. Embedding them here lets us
// assert that the actual files (not just a synthetic template) render cleanly.

//go:embed claude/skill.md
var claudeSkillTemplate string

//go:embed opencode/skill.md
var opencodeSkillTemplate string

func TestRenderSkillRealTemplates(t *testing.T) {
	cases := []struct {
		name     string
		template string
		analyzer string
		header   string
	}{
		{"claude", claudeSkillTemplate, "claude-code", "# Portfolio — Claude Code Skill"},
		{"opencode", opencodeSkillTemplate, "opencode", "# Portfolio — OpenCode Skill"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RenderSkill(tc.template, tc.analyzer)

			for _, needle := range []string{SkillCommonPlaceholder, AnalyzerPlaceholder} {
				if strings.Contains(got, needle) {
					t.Errorf("rendered %s skill still contains placeholder %s", tc.name, needle)
				}
			}
			if !strings.Contains(got, tc.header) {
				t.Errorf("rendered %s skill missing header %q", tc.name, tc.header)
			}
			if !strings.Contains(got, "## Tools") || !strings.Contains(got, "## Three-Tier Knowledge Protocol") {
				t.Errorf("rendered %s skill missing shared sections", tc.name)
			}
			if !strings.Contains(got, `analyzer: "`+tc.analyzer+`"`) {
				t.Errorf("rendered %s skill missing analyzer %q", tc.name, tc.analyzer)
			}
		})
	}
}
