package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderManual verifies the manual is fully expanded (no placeholders), is
// headed as the agent-integration manual, carries the connection section, and
// injects the shared skill body with the per-agent placeholder expanded.
func TestRenderManual(t *testing.T) {
	got := renderManual()

	for _, needle := range []string{"{{SKILL_COMMON}}", "{{ANALYZER}}"} {
		if strings.Contains(got, needle) {
			t.Errorf("rendered manual contains unresolved placeholder %s", needle)
		}
	}

	for _, want := range []string{
		"<!-- DO NOT EDIT — regenerate via: portfolio manual --write -->",
		"# Portfolio — Agent Integration Manual",
		"portfolio mcp",
		"portfolio install opencode",
		"## 1. Connect to the Portfolio MCP server",
		"## Tools", // shared skill body injected
		`analyzer: "<your-agent>"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered manual missing %q", want)
		}
	}
}

// TestRenderManualDeterministic asserts the manual is a pure function of the
// embedded templates — two renders produce byte-identical output — so the
// committed docs/agent-integration-manual.md can be regenerated without churn.
func TestRenderManualDeterministic(t *testing.T) {
	a := renderManual()
	b := renderManual()
	if a != b {
		t.Error("renderManual() is not deterministic; two calls produced different output")
	}
}

// TestRunManualWriteRegeneratesFile exercises the --write path: the file is
// created and its contents match an in-process render.
func TestRunManualWriteRegeneratesFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "agent-integration-manual.md")

	manualWrite = true
	manualWritePath = outPath
	t.Cleanup(func() {
		manualWrite = false
		manualWritePath = manualDefaultPath
	})

	runManual(manualCmd, nil)

	written, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("manual file not written: %v", err)
	}
	if got, want := string(written), renderManual(); got != want {
		t.Error("--write output differs from renderManual()")
	}
}
