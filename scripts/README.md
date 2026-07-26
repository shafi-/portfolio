# Portfolio Integration Scripts

This directory contains integration scripts for setting up Portfolio MCP servers with various AI coding agents.

## ⚠️  Important Warning

All scripts in this directory with the `unsafe-` prefix are **UNOFFICIAL** and **UNSAFE**:

- **Not officially supported** by the respective tools
- **Directly manipulate config files** (fragile)
- **Will break** when tools update their config formats
- **Use at your own risk** after understanding the dangers

## Why These Scripts Exist

Some AI coding tools **do not have official CLI commands** for registering local MCP servers. For those tools, we provide:

1. **Transparent scripts** - You can see exactly what they do
2. **Clear warnings** - Multiple safety prompts and risk explanations
3. **Backup protection** - Automatic config backups before changes
4. **Official alternatives** - Documentation of proper methods where available

## Available Scripts

### unsafe-cline-integration.sh
**Tool**: Cline (VS Code extension)
**Official Method**: VS Code Command Palette → "MCP: Add Server"
**Status**: ❌ No official CLI for programmatic setup

```bash
./scripts/unsafe-cline-integration.sh
```

**What it does**:
- Detects portfolio binary location
- Creates/updates `~/.cline/mcp.json`
- Adds portfolio MCP server configuration
- Backs up existing config before changes

**Risks**: Breaks when Cline updates config format

## Safe Integration Methods

For tools with official methods, **DO NOT USE these scripts**. Use the official
commands instead:

### Claude Code (Official Method) ✅
```bash
portfolio install claude     # or: claude mcp add portfolio /path/to/portfolio mcp
```

### OpenCode (Official Method) ✅
```bash
portfolio install opencode   # writes the schema-documented opencode.json (ADR-021)
```

**See**: [`docs/integration-guideline.md`](../docs/integration-guideline.md) for official methods.

## Script Safety Features

Despite being "unsafe", these scripts include safety measures:

1. **User Consent**: Explicit confirmation before proceeding
2. **Backup Creation**: Automatic timestamped backups
3. **Binary Detection**: Verifies portfolio exists
4. **Error Handling**: Clear error messages and exit codes
5. **Transparency**: Open source, reviewable code

## Usage Guidelines

### When to Use These Scripts

✅ **Use when**:
- You understand the risks
- No official method exists for the tool
- You need automated setup
- You accept potential breakage on updates

❌ **Don't use when**:
- Official CLI method exists (use that instead)
- You prefer manual setup (see guidelines doc)
- Production stability is critical
- You're uncomfortable with direct config editing

### Manual Setup Alternative

If you prefer manual setup or scripts don't work, see:
**[Integration Guidelines](../docs/integration-guideline.md)**

### Troubleshooting

If integration doesn't work:

1. **Verify config format**: Check if tool updated its format
2. **Check binary path**: Ensure portfolio is in PATH
3. **Test manually**: Try running `portfolio mcp` directly
4. **Consult official docs**: Tool may have added official methods
5. **Report issues**: Let us know if format changed

## Removing Integrations

To remove portfolio MCP servers:

### OpenCode
```bash
portfolio uninstall opencode
```

### Cline
```bash
# Edit config manually
nano ~/.cline/mcp.json

# Remove the "portfolio" section from "mcpServers" object
```

### Claude Code
```bash
# Use official command
claude mcp remove portfolio
```

## Contributing

If you find that a tool has added official MCP server registration:

1. **Test the official method** thoroughly
2. **Update integration guidelines** in `docs/integration-guideline.md`
3. **Create safe integration** using official methods
4. **Deprecate unsafe script** with clear migration path

## Support

For issues with these scripts:

1. **Check tool updates**: Tool may have added official method
2. **Review script output**: Check error messages
3. **Verify config format**: Compare with official docs
4. **Manual setup**: Fall back to manual configuration

For general Portfolio issues, see main project documentation.

---

**Remember**: These scripts are last-resort solutions. Always prefer official methods when available!
