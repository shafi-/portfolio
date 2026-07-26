package integration

import (
	_ "embed"
	"strings"
)

// SkillCommonContent is the shared skill body — the tool catalog, the
// three-tier knowledge protocol, notes, and example workflows. It is injected
// into each integration's skill template in place of its {{SKILL_COMMON}}
// placeholder, so the tool documentation lives in exactly one place.
//
//go:embed skill_common.md
var SkillCommonContent string

// SkillCommonPlaceholder marks where the shared skill body is injected into an
// integration's skill template.
const SkillCommonPlaceholder = "{{SKILL_COMMON}}"

// AnalyzerPlaceholder marks where an integration's analyzer identity is injected
// into the shared skill body (e.g. "claude-code", "opencode").
const AnalyzerPlaceholder = "{{ANALYZER}}"

// RenderSkill expands an integration skill template: it substitutes the shared
// skill body for {{SKILL_COMMON}}, then the analyzer identity for {{ANALYZER}}.
// The shared body is injected first so {{ANALYZER}} tokens carried inside it are
// resolved by the second substitution.
func RenderSkill(template, analyzer string) string {
	rendered := strings.ReplaceAll(template, SkillCommonPlaceholder, SkillCommonContent)
	rendered = strings.ReplaceAll(rendered, AnalyzerPlaceholder, analyzer)
	return rendered
}
