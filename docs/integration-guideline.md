# Portfolio Integration Guidelines

This document provides official integration methods for AI coding agents and manual setup guidelines where official methods are insufficient.

## Philosophy

**Official Methods Only**: Portfolio integrations must use the tool's official CLI commands or APIs for MCP server registration. Direct config file manipulation is **not permitted** in production code as it's fragile and breaks when tools update.

**Unsafe Scripts Available**: For tools without official methods, we provide transparent `unsafe-*.sh` scripts in `scripts/` directory. These scripts:
- Clearly state they are unsafe and unofficial
- Require explicit user consent
- Provide automatic backups
- Are fully visible and reviewable
- Should be used only when no official method exists

## Integration Decision Matrix

| Tool | Official Method | Status | Approach |
|------|---------------|--------|----------|
| Claude Code | `claude mcp add` | ✅ Full Support | Use official CLI |
| OpenCode | `opencode mcp add` | ⚠️ Partial | Remote only. Local: use `scripts/unsafe-opencode-integration.sh` |
| Cline | None found | ❌ No Support | Use `scripts/unsafe-cline-integration.sh` or manual setup |

---

## Claude Code Integration

### Official Method ✅

Claude Code has **full official support** for adding local MCP servers via CLI:

```bash
claude mcp add portfolio /path/to/portfolio mcp
```

#### Examples

```bash
# Add with full path
claude mcp add portfolio /usr/local/bin/portfolio mcp

# Add with environment variables
claude mcp add portfolio -e API_KEY=xxx -- /usr/local/bin/portfolio mcp

# Add with subprocess flags
claude mcp add portfolio -- /usr/local/bin/portfolio --flag arg
```

#### Verification

```bash
# List all MCP servers
claude mcp list

# Get portfolio server details
claude mcp get portfolio
```

#### Removal

```bash
claude mcp remove portfolio
```

### Implementation Status

- **Automated Integration**: ✅ Available - Use `claude mcp add` in integration code
- **Manual Setup**: ✅ Documented - See commands above
- **Official Documentation**: [Configuring MCP Tools in Claude Code](https://scottspence.com/posts/configuring-mcp-tools-in-claude-code)

### Config Location (Reference Only - Do Not Edit Directly)

- **File**: `~/.config/claude-code/settings.json`
- **Warning**: Never edit this file directly. Always use `claude mcp` commands.

---

## OpenCode Integration

### Official Method ⚠️

OpenCode has **limited official support** for MCP servers:

```bash
# For remote MCP servers only
opencode mcp add portfolio --url https://your-mcp-server.com/mcp

# For local servers with environment variables (no direct command support)
opencode mcp add portfolio --env PORTFOLIO_PATH=/path/to/portfolio
```

**Limitation**: `opencode mcp add` does NOT support local stdio-based MCP servers with direct command execution. It only supports:
- Remote HTTP servers via `--url`
- Environment variable configuration via `--env`

### Manual Setup Guidelines

Since OpenCode lacks official CLI for local stdio MCP servers, users must manually configure the config file:

#### Step 1: Create/Edit Config File

```bash
# Edit OpenCode config
nano ~/.config/opencode/opencode.json
```

#### Step 2: Add Portfolio MCP Server

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "portfolio": {
      "type": "local",
      "command": [
        "/usr/local/bin/portfolio",
        "mcp"
      ],
      "enabled": true
    }
  }
}
```

#### Step 3: Verify Configuration

```bash
# Check if OpenCode recognizes the server
opencode mcp list

# Debug MCP configuration
opencode debug config
```

#### Step 4: Test MCP Tools

```bash
# Start OpenCode and test portfolio tools
opencode

# Within OpenCode, test:
# Call health()
# Call listProjects()
```

### Config Format Details

**File**: `~/.config/opencode/opencode.json` or `~/.config/opencode/opencode.jsonc`

**Structure for Local Servers**:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "server-name": {
      "type": "local",
      "command": ["executable", "arg1", "arg2"],
      "enabled": true
    }
  }
}
```

**Structure for Remote Servers**:
```json
{
  "mcp": {
    "server-name": {
      "type": "remote",
      "url": "https://server.com/mcp",
      "enabled": true
    }
  }
}
```

### Implementation Status

- **Automated Integration**: ❌ Not Available - No official CLI for local stdio servers
- **Manual Setup**: ✅ Required - Follow steps above
- **Official Documentation**: [OpenCode MCP Configuration Guide](https://blog.wenhaofree.com/en/posts/articles/opencode-mcp-configuration-guide/)

### Important Notes

- **Do Not Automate**: Do not create integration code that directly edits this file
- **Tool Changes**: OpenCode may change config format without notice
- **Alternative**: Consider using remote MCP server if available

---

## Cline Integration

### Official Method ❌

Cline (VS Code extension) **does not have a CLI command** for adding MCP servers. The official methods are:

1. **VS Code Command Palette**: Use UI to add servers
2. **Manual Config Editing**: Edit `~/.cline/mcp.json` directly
3. **VS Code CLI**: `code --add-mcp` (but this is VS Code, not Cline-specific)

### Manual Setup Guidelines

#### Step 1: Locate Config File

```bash
# Cline config location
~/.cline/mcp.json
```

#### Step 2: Create/Edit Config File

```bash
# Create directory if needed
mkdir -p ~/.cline

# Edit config
nano ~/.cline/mcp.json
```

#### Step 3: Add Portfolio MCP Server

```json
{
  "mcpServers": {
    "portfolio": {
      "command": "/usr/local/bin/portfolio",
      "args": ["mcp"],
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

#### Step 4: Restart VS Code

```bash
# Restart VS Code to load new MCP servers
code --reload
```

#### Step 5: Verify in Cline

1. Open Cline panel in VS Code
2. Click MCP Servers icon
3. Verify "portfolio" appears in the server list
4. Test with: `Call health()`

### Config Format Details

**File**: `~/.cline/mcp.json`

**Structure**:
```json
{
  "mcpServers": {
    "server-name": {
      "command": "executable",
      "args": ["arg1", "arg2"],
      "env": {
        "KEY": "value"
      },
      "disabled": false,
      "autoApprove": ["tool1", "tool2"]
    }
  }
}
```

**Field Descriptions**:
- `command`: Executable to run (string)
- `args`: Array of command arguments
- `env`: Optional environment variables
- `disabled`: Set to `true` to disable
- `autoApprove`: Array of tool names to auto-approve

### Implementation Status

- **Automated Integration**: ❌ Not Available - No official CLI method
- **Manual Setup**: ✅ Required - Follow steps above
- **Official Documentation**: [Cline MCP Overview](https://docs.cline.bot/mcp/mcp-overview)

### VS Code Alternative (Not Cline-Specific)

VS Code has MCP support but it's not Cline-specific:

```bash
# Add MCP server to VS Code (not Cline extension)
code --add-mcp '{"name":"portfolio","command":"/usr/local/bin/portfolio","args":["mcp"]}'
```

**Note**: This adds to VS Code's MCP system, not Cline's configuration.

---

## Unsafe Integration Scripts

For tools that lack official MCP server registration methods, we provide transparent "unsafe" scripts:

### Available Scripts

**unsafe-opencode-integration.sh**
```bash
./scripts/unsafe-opencode-integration.sh
```
- **Purpose**: Automate OpenCode MCP server setup (no official method exists)
- **Config**: `~/.config/opencode/opencode.json`
- **Safety**: Automatic backups, user consent required
- **See**: [scripts/README.md](../scripts/README.md)

**unsafe-cline-integration.sh**
```bash
./scripts/unsafe-cline-integration.sh
```
- **Purpose**: Automate Cline MCP server setup (no official method exists)
- **Config**: `~/.cline/mcp.json`
- **Safety**: Automatic backups, user consent required
- **See**: [scripts/README.md](../scripts/README.md)

### Why These Are "Unsafe"

- **Not Official**: Tools don't support these methods officially
- **Fragile**: Break when tools update config formats
- **No Guarantees**: May stop working without notice
- **Direct File Manipulation**: Edits config files directly

### Safety Features

Despite being "unsafe", these scripts include:

✅ **Transparent**: Open source, fully reviewable code
✅ **Warnings**: Multiple safety prompts and risk explanations
✅ **Backups**: Automatic timestamped backups before changes
✅ **Error Handling**: Clear error messages and safe failures
✅ **Documentation**: Explicit risk disclosures

### When to Use

✅ **Use scripts when**:
- You understand and accept the risks
- No official method exists for the tool
- You need automated setup
- You can handle potential breakage

❌ **Don't use scripts when**:
- Official CLI method exists (use that instead)
- Production stability is critical
- You prefer manual setup
- You're uncomfortable with risks

### Manual Setup Alternative

If you prefer manual setup or scripts don't work, follow the manual guidelines in each tool's section above.

---

## General MCP Integration Pattern

When evaluating new tools for integration, follow this decision tree:

```
Does tool have official CLI for adding local MCP servers?
│
├── YES → Create automated integration using official CLI
│         ✅ Example: claude mcp add
│
└── NO  → Do NOT create automated integration
          │
          ├── Document manual setup in this file
          └── Consider alternatives (remote servers, different tools)
```

## Adding New Tools

When adding a new tool to this document:

1. **Research Official Methods**: Search for `[tool] mcp add` or `[tool] MCP CLI`
2. **Test Official Methods**: Verify the CLI works for local stdio servers
3. **Document**: Add section with official commands and examples
4. **Manual Setup**: If no official method exists, provide clear manual steps
5. **Status Indicators**: Use ✅ ⚠️ ❌ for clear communication

## Integration Maintenance

### When Tools Update

1. **Monitor Release Notes**: Check for MCP configuration changes
2. **Test Official Methods**: Verify commands still work after updates
3. **Update Documentation**: Keep this file synchronized with tool changes
4. **Deprecate Warnings**: Mark methods that become obsolete

### Breaking Changes

If a tool changes its MCP configuration format:

1. **Update This Document**: Document the new format immediately
2. **Add Migration Guide**: Help users transition from old to new format
3. **Clear Warnings**: Mark deprecated approaches clearly

## Sources and References

- [Claude Code MCP Configuration](https://scottspence.com/posts/configuring-mcp-tools-in-claude-code)
- [OpenCode MCP Configuration Guide](https://blog.wenhaofree.com/en/posts/articles/opencode-mcp-configuration-guide/)
- [Cline MCP Overview](https://docs.cline.bot/mcp/mcp-overview)
- [VS Code MCP Server Documentation](https://code.visualstudio.com/docs/agent-customization/mcp-servers)
- [MCP Official Specification](https://modelcontextprotocol.io/introduction)

## Contributing

When adding new tools or updating existing ones:

1. **Verify Official Methods**: Test commands personally
2. **Provide Examples**: Include copy-pasteable examples
3. **Document Limitations**: Be clear about what doesn't work
4. **Update Status**: Keep ✅ ⚠️ ❌ indicators accurate
5. **Add Sources**: Reference official documentation

---

**Last Updated**: 2025-07-25
**Maintained By**: Portfolio Team
**Version**: 1.0.0
