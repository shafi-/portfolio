# Epic 8 — Integration Framework

**Milestone:** 2 — Agent Integration
**Status:** Completed

## Overview

Implement integration abstraction, installation framework, validation, upgrade mechanism, and removal/uninstall to support multiple AI agent integrations.

---

## Story 8.1: Integration Abstraction

**Status:** completed
**Size:** M
**Blocked by:** Epic 7

**User Story:**
As the Portfolio Engine, I want an integration abstraction so that different AI agents can be supported.

**Acceptance Criteria:**
- Integration interface defines: install, validate, upgrade, remove
- Integration stores its own metadata
- Engine remains agent-agnostic
- Integrations are discoverable

**Technical Context:**
- Per ADR-013, integrations are first-class components

---

## Story 8.2: Installation Framework

**Status:** completed
**Size:** M
**Blocked by:** 8.1

**User Story:**
As an integration, I want to register with Portfolio so that agents can discover it.

**Acceptance Criteria:**
- Integration registration command
- Stores integration metadata in database
- Validates integration requirements
- Lists installed integrations

---

## Story 8.3: Validation

**Status:** completed
**Size:** S
**Blocked by:** 8.2

**User Story:**
As an integration, I want to validate its installation so that errors are detected early.

**Acceptance Criteria:**
- Integration provides validation command
- Checks: MCP server reachable, tools available
- Returns validation result with diagnostics and remediation steps
- `portfolio doctor` reports issues; `--fix` flag applies self-healable fixes
- Self-healable: restart MCP, recreate config, recreate dir

---

## Story 8.4: Upgrade Mechanism

**Status:** completed
**Size:** M
**Blocked by:** 8.3

**User Story:**
As an integration, I want to upgrade itself so that it stays compatible with engine versions.

**Acceptance Criteria:**
- Integration version tracking
- Upgrade command per integration
- Compatibility check with engine version
- Snapshot files before upgrade, restore on failure
- Manager handles backup/rollback; integration handles agent-specific mutations

---

---

## Story 8.5: Removal/Uninstall

**Status:** completed
**Size:** S
**Blocked by:** 8.4

**User Story:**
As a user, I want to remove an integration so that unused agent integrations don't clutter my setup.

**Acceptance Criteria:**
- CLI command: `portfolio integration remove <name>`
- Removes integration metadata and artifacts
- Unregisters MCP server configuration for the integration
- Warns if integration is in active use
- Idempotent — removing an already-removed integration is a no-op

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 8.1 Integration Abstraction | completed | M | Epic 7 |
| 8.2 Installation Framework | completed | M | 8.1 |
| 8.3 Validation | completed | S | 8.2 |
| 8.4 Upgrade Mechanism | completed | M | 8.3 |
| 8.5 Removal/Uninstall | completed | S | 8.4 |

**Total Size:** 3M + 2S = ~12 days

**Completed:** All stories completed
