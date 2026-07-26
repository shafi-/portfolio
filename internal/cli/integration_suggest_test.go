package cli

import (
	"strings"
	"testing"
)

func TestFindUnsupportedAgent(t *testing.T) {
	tests := []struct {
		target   string
		wantOK   bool
		wantName string // display name when found
	}{
		{"cline", true, "Cline"},
		{"Cline", true, "Cline"}, // case-insensitive
		{"CLINE", true, "Cline"}, // case-insensitive
		{"opencode", false, ""},  // supported (official integration), not unsupported
		{"cursor", false, ""},    // unknown
		{"claude", false, ""},    // supported, not unsupported
		{"", false, ""},          // empty
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			a, ok := findUnsupportedAgent(tt.target)
			if ok != tt.wantOK {
				t.Fatalf("findUnsupportedAgent(%q) ok = %v, want %v", tt.target, ok, tt.wantOK)
			}
			if ok && a.display != tt.wantName {
				t.Errorf("display = %q, want %q", a.display, tt.wantName)
			}
		})
	}
}

func TestFormatAgentSuggestion(t *testing.T) {
	agent, ok := findUnsupportedAgent("cline")
	if !ok {
		t.Fatal("expected cline to be a known unsupported agent")
	}

	got := formatAgentSuggestion("install", "cline", agent, "/usr/local/bin/portfolio")

	wantSubstrings := []string{
		"'cline' is not a supported install target",
		"Cline",
		"scripts/unsafe-cline-integration.sh",
		"~/.cline/mcp.json",
		"Unofficial and unsafe",
		"/usr/local/bin/portfolio", // binary path injected into the snippet
		`"args": ["mcp"]`,
		"docs/integration-guideline.md",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("suggestion missing %q;\ngot:\n%s", s, got)
		}
	}
}

func TestFormatUnknownAgentSuggestion(t *testing.T) {
	got := formatUnknownAgentSuggestion("install", "cursor")

	wantSubstrings := []string{
		"'cursor' is not a recognized install target",
		"Supported targets: claude, opencode",
		"portfolio mcp",
		"portfolio manual",
		"docs/agent-integration-manual.md",
	}
	for _, s := range wantSubstrings {
		if !strings.Contains(got, s) {
			t.Errorf("suggestion missing %q;\ngot:\n%s", s, got)
		}
	}
}
