#!/bin/bash

##############################################################################
# ⚠️  UNSAFE - Unofficial Cline Integration Script ⚠️
##############################################################################
#
# This script performs MANUAL integration that is NOT supported by Cline.
# It directly edits configuration files which may break when Cline updates.
#
# WHY THIS IS UNSAFE:
# - Cline has NO official CLI for adding MCP servers
# - Direct config file manipulation is fragile and breaks on updates
# - Cline may change config format without notice
# - This is a VS Code extension, not a standalone tool
#
# OFFICIAL METHOD STATUS:
# ❌ No official CLI or API for MCP server registration
# ✅ Manual UI: Use VS Code Command Palette → "MCP: Add Server"
#
# ALTERNATIVES:
# 1. Use VS Code's built-in MCP support (not Cline-specific)
# 2. Use Claude Code instead (has full official support)
# 3. Request Cline to add official CLI support
#
# This script exists only for users who understand the risks and have
# no other option. Use at your own risk.
#
##############################################################################

set -e

# Colors for warnings
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${RED}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${RED}⚠️  UNSAFE CLINE INTEGRATION - NOT OFFICIALLY SUPPORTED ⚠️${NC}"
echo -e "${RED}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}This script directly edits Cline config files.${NC}"
echo -e "${YELLOW}This will break when Cline updates its config format.${NC}"
echo ""
echo -e "Official method: ${GREEN}VS Code Command Palette → MCP: Add Server${NC}"
echo -e "This method:     ${RED}Direct config file editing (UNSAFE)${NC}"
echo ""
read -p "Do you understand the risks and want to continue? (yes/no): " consent

if [ "$consent" != "yes" ]; then
    echo "Integration cancelled."
    exit 1
fi

# Detect portfolio binary path
PORTFOLIO_BIN=$(which portfolio 2>/dev/null || echo "")
if [ -z "$PORTFOLIO_BIN" ]; then
    # Check current directory
    if [ -f "./portfolio" ]; then
        PORTFOLIO_BIN="$(pwd)/portfolio"
    else
        echo -e "${RED}Error: portfolio binary not found in PATH or current directory${NC}"
        echo "Please install portfolio or provide full path to binary."
        exit 1
    fi
fi

echo -e "${GREEN}✓ Found portfolio binary: $PORTFOLIO_BIN${NC}"

# Detect home directory
HOME_DIR=$(eval echo ~)
CONFIG_DIR="$HOME_DIR/.cline"
CONFIG_FILE="$CONFIG_DIR/mcp.json"

# Create config directory if needed
mkdir -p "$CONFIG_DIR"

# Backup existing config
if [ -f "$CONFIG_FILE" ]; then
    BACKUP_FILE="$CONFIG_FILE.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$CONFIG_FILE" "$BACKUP_FILE"
    echo -e "${GREEN}✓ Backed up existing config to: $BACKUP_FILE${NC}"
fi

# Create or update config
if [ ! -f "$CONFIG_FILE" ]; then
    # Create new config file
    cat > "$CONFIG_FILE" << EOF
{
  "mcpServers": {
    "portfolio": {
      "command": "$PORTFOLIO_BIN",
      "args": ["mcp"],
      "disabled": false,
      "autoApprove": []
    }
  }
}
EOF
    echo -e "${GREEN}✓ Created new Cline config file${NC}"
else
    # Update existing config using Python for JSON manipulation
    if command -v python3 &> /dev/null; then
        python3 << EOF
import json
import os

config_file = "$CONFIG_FILE"
portfolio_bin = "$PORTFOLIO_BIN"

# Read existing config
with open(config_file, 'r') as f:
    config = json.load(f)

# Ensure mcpServers section exists
if 'mcpServers' not in config:
    config['mcpServers'] = {}

# Add portfolio MCP server
config['mcpServers']['portfolio'] = {
    'command': portfolio_bin,
    'args': ['mcp'],
    'disabled': False,
    'autoApprove': []
}

# Write back
with open(config_file, 'w') as f:
    json.dump(config, f, indent=2)

print("✓ Updated existing Cline config file")
EOF
    else
        echo -e "${YELLOW}Warning: python3 not found, creating basic config${NC}"
        # Fallback: create basic config (may overwrite existing)
        cat > "$CONFIG_FILE" << EOF
{
  "mcpServers": {
    "portfolio": {
      "command": "$PORTFOLIO_BIN",
      "args": ["mcp"],
      "disabled": false,
      "autoApprove": []
    }
  }
}
EOF
    fi
fi

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Cline integration completed${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo "Config file: $CONFIG_FILE"
echo ""
echo "Next steps:"
echo "1. Restart VS Code to load new MCP servers"
echo "2. Open Cline panel in VS Code"
echo "3. Click MCP Servers icon (stacked server icon)"
echo "4. Verify 'portfolio' appears in server list"
echo "5. Test with: Call health()"
echo ""
echo -e "${YELLOW}⚠️  Remember: This integration may break on Cline updates!${NC}"
echo -e "${YELLOW}⚠️  This is NOT an official integration method!${NC}"
