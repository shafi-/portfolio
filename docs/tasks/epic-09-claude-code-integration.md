# Epic 9 — Claude Code Integration

**Milestone:** 2 — Agent Integration
**Status:** todo

## Overview

Implement Claude Code integration including MCP server installation, Portfolio skill installation, verification, upgrade, and uninstall capabilities.

---

## Story 9.1: Install MCP

**Status:** todo
**Size:** M
**Blocked by:** 8.2

**User Story:**
As a Claude Code user, I want to install the Portfolio MCP server so that my AI agent can access my portfolio.

**Acceptance Criteria:**
- `portfolio install claude` command
- Registers Portfolio MCP server with Claude Code
- Configures server connection
- Verification that tools are available

---

## Story 9.2: Install Portfolio Skill

**Status:** todo
**Size:** M
**Blocked by:** 9.1

**User Story:**
As a Claude Code user, I want a Portfolio skill so that my AI agent knows how to interact with Portfolio.

**Acceptance Criteria:**
- Installs Portfolio-specific skill/prompt for Claude Code
- Skill describes available MCP tools
- Provides examples of Portfolio queries
- Skill is upgradable with integration

---

## Story 9.3: Verify Integration

**Status:** todo
**Size:** S
**Blocked by:** 9.2

**User Story:**
As a user, I want to verify the integration so that I know it's working correctly.

**Acceptance Criteria:**
- `portfolio doctor` checks Claude integration
- Verifies MCP server connection
- Tests tool availability
- Reports any issues with remediation steps

---

## Story 9.4: Update Integration

**Status:** todo
**Size:** M
**Blocked by:** 8.4

**User Story:**
As a user, I want to update the integration so that it stays compatible with new features.

**Acceptance Criteria:**
- `portfolio upgrade claude` command
- Updates MCP server and skill
- Preserves configuration
- Reports changes

---

---

## Story 9.5: Uninstall Integration

**Status:** todo
**Size:** S
**Blocked by:** 9.4

**User Story:**
As a user, I want to remove the Claude Code integration so that my Portfolio setup stays clean.

**Acceptance Criteria:**
- `portfolio uninstall claude` command
- Removes MCP server registration from Claude Code config
- Removes Portfolio skill file
- Preserves project data (uninstall is agent-only, not portfolio-wide)
- Idempotent — re-running after removal reports "not installed"

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 9.1 Install MCP | todo | M | 8.2 |
| 9.2 Install Portfolio Skill | todo | M | 9.1 |
| 9.3 Verify Integration | todo | S | 9.2 |
| 9.4 Update Integration | todo | M | 8.4 |
| 9.5 Uninstall Integration | todo | S | 9.4 |

**Total Size:** 3M + 2S = ~12 days

**Can Start:** Story 9.1 (after 8.2 complete)
