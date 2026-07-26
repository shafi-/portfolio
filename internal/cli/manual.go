package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"project-dash/internal/integration"
)

// manualAnalyzer is the per-agent identity shown in the rendered manual. The
// manual targets agents Portfolio does not automate, so it uses a placeholder
// the reader replaces with their own agent's identity.
const manualAnalyzer = "<your-agent>"

const manualDefaultPath = "docs/agent-integration-manual.md"

var (
	manualWrite     bool
	manualWritePath string
)

var manualCmd = &cobra.Command{
	Use:   "manual",
	Short: "Print the agent-integration manual",
	Long: `Render and print the Portfolio agent-integration manual.

The manual documents how to connect any MCP-capable AI coding agent to
Portfolio (the MCP server connection plus the shared skill). It is generated
from the same canonical skill template the supported integrations use, so it
never drifts from the tools Portfolio actually exposes.

Use --write to regenerate the committed copy at
docs/agent-integration-manual.md.`,
	Args: cobra.NoArgs,
	Run:  runManual,
}

func init() {
	rootCmd.AddCommand(manualCmd)
	manualCmd.Flags().BoolVar(&manualWrite, "write", false, "write the manual to disk instead of printing")
	manualCmd.Flags().StringVar(&manualWritePath, "path", manualDefaultPath, "output path when --write is set")
}

func runManual(cmd *cobra.Command, args []string) {
	out := renderManual()

	if !manualWrite {
		fmt.Print(out)
		return
	}

	if err := os.WriteFile(manualWritePath, []byte(out), 0644); err != nil {
		fmt.Printf("Error: failed to write manual: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Agent-integration manual written to %s\n", manualWritePath)
}

// manualTemplate is the manual shell. {{SKILL_COMMON}} injects the shared skill
// body (which itself carries {{ANALYZER}} tokens); RenderSkill expands both.
// It deliberately avoids backticks so it can be a Go raw-string literal.
const manualTemplate = `<!-- DO NOT EDIT — regenerate via: portfolio manual --write -->
# Portfolio — Agent Integration Manual

This is the canonical reference for connecting any MCP-capable AI coding agent
to Portfolio. It is rendered from Portfolio's shared skill template, so it
always matches the tools Portfolio actually exposes — it never drifts.

Supported agents with automated setup:

    portfolio install claude      # Claude Code
    portfolio install opencode    # OpenCode

For any other agent (Cursor, Zed, Continue, Roo, ...), follow the steps below.

This file is generated. Regenerate it with:  portfolio manual --write

## 1. Connect to the Portfolio MCP server

Portfolio runs an MCP server over stdio. Configure your agent to launch it as a
local stdio MCP server using the command:

    portfolio mcp

If your agent config requires an absolute path, use the path to your installed
portfolio binary followed by "mcp". The server needs no other arguments.

## 2. Load the Portfolio skill

Give your agent the skill text below so it knows Portfolio's tools and the
three-tier knowledge protocol. When the skill stores analysis or features it
records an analyzer identity; replace <your-agent> below with a stable name for
your agent (for example cursor or zed) so Portfolio can attribute the work.

---

{{SKILL_COMMON}}
`

func renderManual() string {
	return integration.RenderSkill(manualTemplate, manualAnalyzer)
}
