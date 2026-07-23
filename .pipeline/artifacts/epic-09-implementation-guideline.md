# Epic 9 — Claude Code Integration: Implementation Guideline

**Reference:** `.architecture/epic-09-architecture.md`, `.requirements/epic-09-requirements.md`, `docs/tasks/epic-09-claude-code-integration.md`

---

## 1. Technical Standards

### Language & Runtime
- Go 1.22.6, module `github.com/nerddevsltd/portfolio`
- Claude Code CLI configuration via filesystem (Claude Code config file location: platform-dependent)
- MCP server from Epic 7 registered in Claude Code's MCP config

### Package Organization
```
internal/integration/claude/
├── claude.go         — Claude Code integration (implements integration.Integration)
├── install.go        — MCP server + skill installation
├── skill.go          — Portfolio skill prompt generation
├── verify.go         — Integration verification
└── uninstall.go      — Removal logic
```

### Key Design Decisions
- **MCP server registration:** Write to Claude Code's `claude.json` MCP servers config
- **Skill installation:** Write a CLAUDE.md or skill file to the user's project
- **Analyzer identity:** Claude Code uses `analyzer: "claude-code"` in storeAnalysis calls
- **CLI commands:** `portfolio install claude`, `portfolio uninstall claude`, `portfolio upgrade claude`

## 2. Implementation Details

### MCP Registration
Claude Code's MCP config is at platform-specific path. Register as:
```json
{
  "mcpServers": {
    "portfolio": {
      "command": "portfolio",
      "args": ["mcp"],
      "transport": "stdio"
    }
  }
}
```

### Portfolio Skill
Install a skill file or CLAUDE.md addition that tells Claude Code:
- Available MCP tools (listProjects, searchProjects, storeAnalysis, etc.)
- When to use each tool
- Analyzer identity: always set `analyzer` to `"claude-code"`
- Workflow: health → discover → search → analyze → store

### Verification
`portfolio doctor` checks:
- Claude Code CLI is installed (`which claude`)
- MCP server is registered in Claude Code config
- MCP tools are available (calls health())
- Skill file exists

## 3. Implementation Order

### Story 9.1 — Install MCP
- Detect Claude Code config path
- Write MCP server entry
- Verify tools available

### Story 9.2 — Install Portfolio Skill
- Generate skill prompt file
- Write to Claude Code's skills directory or CLAUDE.md

### Story 9.3 — Verify Integration
- `portfolio doctor` checks Claude integration
- Reports status with remediation steps

### Story 9.4 — Update Integration
- `portfolio upgrade claude` command
- Updates MCP config and skill

### Story 9.5 — Uninstall Integration
- `portfolio uninstall claude` command
- Removes MCP config entry, skill file
- Preserves portfolio data

## 4. Testing Strategy

### Unit Tests
- MCP config file read/write
- Skill prompt generation
- Path detection for Claude Code config

### Integration Tests
- Full install/uninstall lifecycle in temp dir
- Verify MCP config file format matches Claude Code expectations
- `portfolio doctor` correctly reports installed/missing states

## 5. Build & Verification

```bash
go build ./cmd/portfolio
go test ./internal/integration/claude/... -v -cover
```

## 6. Quality Gates

- [ ] Claude Code integration installs without errors
- [ ] `portfolio doctor` reports correct integration status
- [ ] Uninstall is clean (no leftover config)
- [ ] Upgrade preserves user configuration
- [ ] All acceptance criteria from `.requirements/epic-09-requirements.md` pass
