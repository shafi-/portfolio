package cli

import (
	"fmt"
	"os"
	"strings"
)

// unsupportedAgent describes an AI coding agent that has no official,
// automatable Portfolio integration — i.e. no CLI for registering a local
// (stdio) MCP server. For these agents Portfolio falls back to unofficial
// scripts and manual setup, per ADR-016 ("Official Methods Only for Agent
// Integrations") and docs/integration-guideline.md.
type unsupportedAgent struct {
	aliases    []string // lower-case names the user might type
	display    string   // human-readable tool name
	script     string   // repo-relative path to the unsafe script ("" if none)
	configFile string   // agent's MCP config file location
	snippet    string   // manual config snippet (contains one %s for the binary path)
}

// knownUnsupportedAgents mirrors the Integration Decision Matrix in
// docs/integration-guideline.md. Keep this in sync with that document and with
// the scripts/ directory.
var knownUnsupportedAgents = []unsupportedAgent{
	{
		aliases:    []string{"cline"},
		display:    "Cline",
		script:     "scripts/unsafe-cline-integration.sh",
		configFile: "~/.cline/mcp.json",
		snippet: `{
  "mcpServers": {
    "portfolio": {
      "command": "%s",
      "args": ["mcp"]
    }
  }
}`,
	},
	{
		aliases:    []string{"opencode"},
		display:    "OpenCode",
		script:     "scripts/unsafe-opencode-integration.sh",
		configFile: "~/.config/opencode/opencode.json",
		snippet: `{
  "mcp": {
    "portfolio": {
      "type": "local",
      "command": ["%s", "mcp"],
      "enabled": true
    }
  }
}`,
	},
}

// findUnsupportedAgent returns the known unsupported agent matching target
// (case-insensitive, matched against aliases) and true, or false if unknown.
func findUnsupportedAgent(target string) (unsupportedAgent, bool) {
	t := strings.ToLower(target)
	for _, a := range knownUnsupportedAgents {
		for _, alias := range a.aliases {
			if alias == t {
				return a, true
			}
		}
	}
	return unsupportedAgent{}, false
}

// resolveBinaryPath returns the absolute path to the running portfolio binary,
// falling back to os.Args[0] if the executable path cannot be resolved.
func resolveBinaryPath() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return os.Args[0]
}

// formatAgentSuggestion builds guidance for a known unsupported agent.
// verb is "install", "uninstall", or "upgrade".
func formatAgentSuggestion(verb, target string, a unsupportedAgent, binaryPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: '%s' is not a supported %s target.\n\n", target, verb)
	fmt.Fprintf(&b, "Portfolio has no official automated integration for %s\n", a.display)
	b.WriteString("(it has no CLI for registering local MCP servers).\n")
	b.WriteString("Supported target: claude\n\n")
	b.WriteString("You can set it up with the unofficial script approach:\n\n")
	if a.script != "" {
		fmt.Fprintf(&b, "  ⚠️  Unofficial and unsafe — it edits %s directly and may\n", a.configFile)
		fmt.Fprintf(&b, "     break when %s changes its config format. Review it first:\n\n", a.display)
		fmt.Fprintf(&b, "    %s\n\n", a.script)
	}
	fmt.Fprintf(&b, "Or configure %s manually by adding this to %s:\n\n", a.display, a.configFile)
	for _, line := range strings.Split(fmt.Sprintf(a.snippet, binaryPath), "\n") {
		fmt.Fprintf(&b, "    %s\n", line)
	}
	b.WriteString("\nSee docs/integration-guideline.md for full details.\n")
	return b.String()
}

// formatUnknownAgentSuggestion builds guidance for an agent Portfolio doesn't
// recognize at all.
func formatUnknownAgentSuggestion(verb, target string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: '%s' is not a recognized %s target.\n\n", target, verb)
	b.WriteString("Supported target: claude\n\n")
	b.WriteString("Portfolio works with any MCP-capable agent by running its MCP server:\n\n")
	b.WriteString("    portfolio mcp\n\n")
	b.WriteString("Unofficial setup scripts exist for some agents (Cline, OpenCode) in scripts/,\n")
	b.WriteString("and manual setup for others is documented in docs/integration-guideline.md.\n")
	return b.String()
}

// suggestAgentIntegration prints the appropriate guidance for an unsupported
// or unrecognized agent target. Callers should exit non-zero afterwards.
func suggestAgentIntegration(verb, target string) {
	if agent, ok := findUnsupportedAgent(target); ok {
		fmt.Print(formatAgentSuggestion(verb, target, agent, resolveBinaryPath()))
		return
	}
	fmt.Print(formatUnknownAgentSuggestion(verb, target))
}
