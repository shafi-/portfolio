#!/bin/bash

##############################################################################
# ⚠️  UNSAFE - Unofficial OpenCode Integration Script ⚠️
##############################################################################
#
# This script performs MANUAL integration that is NOT supported by OpenCode.
# It directly edits configuration files which may break when OpenCode updates.
#
# WHY THIS IS UNSAFE:
# - OpenCode has NO official CLI for adding local stdio MCP servers
# - Direct config file manipulation is fragile and breaks on updates
# - OpenCode may change config format without notice
#
# OFFICIAL METHOD STATUS:
# ✅ Remote MCP servers: opencode mcp add --url https://...
# ❌ Local stdio servers: NO OFFICIAL METHOD EXISTS
#
# ALTERNATIVES:
# 1. Use remote MCP server if available
# 2. Request OpenCode to add official local server support
# 3. Use Claude Code instead (has full official support)
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
echo -e "${RED}⚠️  UNSAFE OPENCODE INTEGRATION - NOT OFFICIALLY SUPPORTED ⚠️${NC}"
echo -e "${RED}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}This script directly edits OpenCode config files.${NC}"
echo -e "${YELLOW}This will break when OpenCode updates its config format.${NC}"
echo ""
echo -e "Official method: ${GREEN}opencode mcp add --url https://...${NC} (remote only)"
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
CONFIG_DIR="$HOME_DIR/.config"
CONFIG_FILE="$CONFIG_DIR/opencode/opencode.json"

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
  "\$schema": "https://opencode.ai/config.json",
  "mcp": {
    "portfolio": {
      "type": "local",
      "command": [
        "$PORTFOLIO_BIN",
        "mcp"
      ],
      "enabled": true
    }
  }
}
EOF
    echo -e "${GREEN}✓ Created new OpenCode config file${NC}"
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

# Ensure mcp section exists
if 'mcp' not in config:
    config['mcp'] = {}

# Add portfolio MCP server
config['mcp']['portfolio'] = {
    'type': 'local',
    'command': [portfolio_bin, 'mcp'],
    'enabled': True
}

# Add schema if not present
if '\$schema' not in config:
    config['\$schema'] = 'https://opencode.ai/config.json'

# Write back
with open(config_file, 'w') as f:
    json.dump(config, f, indent=2)

print("✓ Updated existing OpenCode config file")
EOF
    else
        echo -e "${YELLOW}Warning: python3 not found, creating basic config${NC}"
        # Fallback: create basic config (may overwrite existing)
        cat > "$CONFIG_FILE" << EOF
{
  "\$schema": "https://opencode.ai/config.json",
  "mcp": {
    "portfolio": {
      "type": "local",
      "command": [
        "$PORTFOLIO_BIN",
        "mcp"
      ],
      "enabled": true
    }
  }
}
EOF
    fi
fi

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ OpenCode integration completed${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
echo ""
echo "Config file: $CONFIG_FILE"
echo ""
echo "Next steps:"
echo "1. Verify configuration: opencode debug config"
echo "2. Check MCP servers: opencode mcp list"
echo "3. Start OpenCode: opencode"
echo "4. Test portfolio tools: Call health()"
echo ""
echo -e "${YELLOW}⚠️  Remember: This integration may break on OpenCode updates!${NC}"
echo -e "${YELLOW}⚠️  This is NOT an official integration method!${NC}"
