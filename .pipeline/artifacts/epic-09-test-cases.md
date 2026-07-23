# Epic 9 — Claude Code Integration: Test Cases

**Milestone:** 2 — Agent Integration
**Version:** 1.0

---

## Story 9.1 — Install MCP Server

### TC-9.1.1 — Happy path: fresh MCP server installation

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.1 |
| **Title** | Fresh MCP server installation succeeds |
| **Description** | Verify that `portfolio install claude` registers the Portfolio MCP server in Claude Code's config when no previous registration exists. |
| **Preconditions** | Claude Code CLI is installed; Claude Code config file exists (~/.claude/settings.json or claude_desktop_config.json); Portfolio Engine is installed and functional; No existing Portfolio MCP registration. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output<br>3. Inspect Claude Code config file for Portfolio MCP server entry<br>4. Verify that transport type and server path are correctly configured<br>5. Run `portfolio install claude` again |
| **Expected Result** | First run: command exits 0, outputs success message ("MCP server registered"), config file contains a valid Portfolio MCP server entry with stdio transport and correct binary path. Second run: command exits 0, reports "already registered", no duplicate entry in config. |
| **Story** | 9.1 |

### TC-9.1.2 — Idempotent: re-running install

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.2 |
| **Title** | Re-running install reports already registered |
| **Description** | Verify install is idempotent — re-running does not duplicate entries. |
| **Preconditions** | Portfolio MCP server already registered in Claude Code config. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output<br>3. Inspect config file for Portfolio MCP entries |
| **Expected Result** | Command exits 0, outputs "already registered". Config file contains exactly one Portfolio MCP server entry. |
| **Story** | 9.1 |

### TC-9.1.3 — Claude Code not installed

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.3 |
| **Title** | Error when Claude Code CLI is not detected |
| **Description** | Verify that install fails gracefully when Claude Code is not present. |
| **Preconditions** | Claude Code CLI is not installed on the system; Portfolio Engine is installed. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output and exit code |
| **Expected Result** | Command exits non-zero. Output clearly states that Claude Code CLI was not detected, includes installation instructions (e.g., "Install Claude Code from https://docs.anthropic.com/claude-code"), and does NOT modify any config files. |
| **Story** | 9.1 |

### TC-9.1.4 — Claude Code config file missing

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.4 |
| **Title** | Handles missing Claude Code config file |
| **Description** | Verify that install creates the config file when it doesn't exist. |
| **Preconditions** | Claude Code CLI is installed; ~/.claude/ directory exists but settings.json does not exist; No config file at any known Claude Code config location. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output<br>3. Check that config file was created at the expected location<br>4. Verify the config file contains the Portfolio MCP server entry |
| **Expected Result** | Command exits 0. Config file is created with valid JSON containing the Portfolio MCP server entry. Output reports "created config file" and "MCP server registered". |
| **Story** | 9.1 |

### TC-9.1.5 — Permission denied on config write

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.5 |
| **Title** | Permission denied when writing config |
| **Description** | Verify graceful error when the process lacks write permission to the Claude Code config directory. |
| **Preconditions** | Claude Code CLI installed; ~/.claude/ directory exists but is owned by root or has read-only permissions for the current user. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output and exit code |
| **Expected Result** | Command exits non-zero. Error message includes the config file path and suggests running with elevated permissions or manually editing the file. No partial config is written. |
| **Story** | 9.1 |

### TC-9.1.6 — MCP server binary missing or corrupt

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.6 |
| **Title** | MCP server binary path is invalid |
| **Description** | Verify that install detects if the Portfolio MCP binary is missing or corrupt during the verification step. |
| **Preconditions** | Portfolio Engine install is incomplete — MCP server binary does not exist at the expected path or is corrupt (non-executable). |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output |
| **Expected Result** | Command exits non-zero after attempting verification. Error message includes the expected binary path, suggests re-installing Portfolio Engine, and the config file is NOT updated to avoid registering a dead entry. |
| **Story** | 9.1 |

### TC-9.1.7 — MCP health check fails after registration

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.7 |
| **Title** | MCP server health check fails post-registration |
| **Description** | Verify that install reports failure when the MCP server process starts but health() does not respond. |
| **Preconditions** | Claude Code CLI installed; MCP server binary exists but returns errors on health check (e.g., port conflict, database unavailable). |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output |
| **Expected Result** | Command exits non-zero. Error message includes which tool(s) failed, diagnostic info (stderr if available), and suggests checking for port/socket conflicts or restarting. Config is NOT updated to avoid registering a broken endpoint. |
| **Story** | 9.1 |

### TC-9.1.8 — All MCP tools respond after registration

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.8 |
| **Title** | Verification confirms all MCP tools respond |
| **Description** | Verify that the post-registration verification step checks tool availability and reports success. |
| **Preconditions** | Claude Code CLI installed; Portfolio Engine is running; MCP server is functional. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output<br>3. Check that verification calls health() and listProjects() |
| **Expected Result** | Output includes "Verification: OK" or similar, confirming all MCP tools respond correctly. Command exits 0. |
| **Story** | 9.1 |

### TC-9.1.9 — Transport configuration is correct

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.9 |
| **Title** | MCP server transport is configured as stdio |
| **Description** | Verify that the transport type and path/port in the config match the expected stdio transport configuration. |
| **Preconditions** | Claude Code CLI installed. |
| **Steps** | 1. Run `portfolio install claude`<br>2. After success, read Claude Code config file<br>3. Parse the Portfolio MCP server entry |
| **Expected Result** | The entry uses `"transport": "stdio"` (or the engine's expected transport type). The `command` field points to the correct Portfolio MCP binary path. Args are correctly specified if needed. |
| **Story** | 9.1 |

### TC-9.1.10 — Help flag shows install claude command

| Field | Value |
|-------|-------|
| **ID** | TC-9.1.10 |
| **Title** | `portfolio install claude` is discoverable via help |
| **Description** | Verify the command appears in help output. |
| **Preconditions** | Portfolio Engine is installed. |
| **Steps** | 1. Run `portfolio --help`<br>2. Run `portfolio install --help` |
| **Expected Result** | `--help` output includes `claude` as a subcommand of `install`. `portfolio install --help` describes the Claude Code integration installation including MCP registration, skill installation, and verification. |
| **Story** | 9.1 |

---

## Story 9.2 — Install Portfolio Skill

### TC-9.2.1 — Happy path: skill file installed

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.1 |
| **Title** | Portfolio skill file is installed into Claude Code skills directory |
| **Description** | Verify that `portfolio install claude` copies the skill file to the correct location. |
| **Preconditions** | Claude Code CLI installed; Claude Code skills directory exists at the expected path. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output<br>3. Check Claude Code skills directory for Portfolio skill file |
| **Expected Result** | Command exits 0. Skill file exists at the expected path (e.g., `~/.claude/skills/portfolio.md` or equivalent). Output reports "Skill installed." |
| **Story** | 9.2 |

### TC-9.2.2 — Skill file content: tool descriptions

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.2 |
| **Title** | Skill describes all MCP tools |
| **Description** | Verify the skill file contains descriptions for each MCP tool implemented in Epic 7. |
| **Preconditions** | Skill file has been installed. |
| **Steps** | 1. Read the installed skill file<br>2. Verify each tool name is mentioned with a description |
| **Expected Result** | Skill file covers at minimum: health, discoverProjects, listProjects, getProject, searchProjects, searchDocumentation, getAnalysis, storeAnalysis, listProjectsNeedingAnalysis, getConfiguration, updateConfiguration, listRelationships. Each tool has a brief description of its purpose. |
| **Story** | 9.2 |

### TC-9.2.3 — Skill file content: example queries

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.3 |
| **Title** | Skill includes example queries |
| **Description** | Verify the skill file contains at least 3 example interaction patterns. |
| **Preconditions** | Skill file has been installed. |
| **Steps** | 1. Read the installed skill file<br>2. Count example queries or interaction patterns |
| **Expected Result** | At least 3 distinct example queries are present. Examples should include common patterns such as listing projects, retrieving analysis for a project, and searching by technology. |
| **Story** | 9.2 |

### TC-9.2.4 — Skill format is plain text or markdown

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.4 |
| **Title** | Skill file is plain text or markdown |
| **Description** | Verify the skill file is stored in a plain text format compatible with Claude Code. |
| **Preconditions** | Skill file has been installed. |
| **Steps** | 1. Check the skill file extension and MIME type<br>2. Read the file to confirm it is human-readable plain text |
| **Expected Result** | File extension is `.md` or `.txt` (or no extension if plain text is standard). File content is valid markdown or plain text — NOT binary, PDF, HTML, or any other format. |
| **Story** | 9.2 |

### TC-9.2.5 — Skill file already exists (overwrite)

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.5 |
| **Title** | Existing skill file is overwritten with latest version |
| **Description** | Verify that re-running install overwrites the skill file with the latest version. |
| **Preconditions** | Skill file already exists with an older version's content. |
| **Steps** | 1. Note the current skill file content / version marker<br>2. Run `portfolio install claude`<br>3. Read the skill file<br>4. Observe output |
| **Expected Result** | Skill file is overwritten with the latest version. Output reports "Skill updated." Command exits 0. |
| **Story** | 9.2 |

### TC-9.2.6 — Skills directory does not exist

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.6 |
| **Title** | Skills directory is created if missing |
| **Description** | Verify that install creates the skills directory when it does not exist. |
| **Preconditions** | Claude Code CLI installed; the skills directory (e.g., `~/.claude/skills/`) does not exist; parent config directory exists. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output<br>3. Check that the skills directory was created |
| **Expected Result** | Command exits 0. Skills directory is created with correct permissions. Skill file is placed inside. Output reports "created skills directory" or similar. |
| **Story** | 9.2 |

### TC-9.2.7 — Skill file is upgradable

| Field | Value |
|-------|-------|
| **ID** | TC-9.2.7 |
| **Title** | Upgrade command replaces skill file |
| **Description** | Verify that `portfolio upgrade claude` updates the skill file to a new version. |
| **Preconditions** | Old version of skill file is installed; a newer version is available in the integration package. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Read the skill file<br>3. Observe output |
| **Expected Result** | Skill file content is updated to the new version. Output includes "Skill updated to version X.Y.Z." |
| **Story** | 9.2 |

---

## Story 9.3 — Verify Integration (doctor)

### TC-9.3.1 — Happy path: all checks pass

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.1 |
| **Title** | `portfolio doctor` reports healthy integration |
| **Description** | Verify that when the integration is fully installed and working, doctor reports healthy. |
| **Preconditions** | Claude Code integration is installed and configured; MCP server is running; MCP health() returns OK; skill file exists. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output and exit code |
| **Expected Result** | Command exits 0. Output includes "Claude Code integration: healthy" or equivalent for each check passing. |
| **Story** | 9.3 |

### TC-9.3.2 — Integration not installed

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.2 |
| **Title** | Doctor reports integration not installed |
| **Description** | Verify that doctor correctly identifies when the integration has never been installed. |
| **Preconditions** | Portfolio Engine installed; `portfolio install claude` has never been run; no integration metadata exists in database. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output and exit code |
| **Expected Result** | Command exits non-zero. Output specifically reports "Claude Code integration: not installed" and provides installation instructions (e.g., "Run 'portfolio install claude' to install"). |
| **Story** | 9.3 |

### TC-9.3.3 — Integration partially installed (MCP registered, skill missing)

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.3 |
| **Title** | Doctor detects missing skill file |
| **Description** | Verify that doctor identifies when MCP is registered but the skill file is missing. |
| **Preconditions** | MCP server is registered in Claude Code config; integration metadata exists; skill file has been deleted or is missing. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output |
| **Expected Result** | Command exits non-zero. Output shows "Skill file: MISSING" or similar, with remediation: "Run 'portfolio install claude' to reinstall the skill file." |
| **Story** | 9.3 |

### TC-9.3.4 — MCP server process not running

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.4 |
| **Title** | Doctor detects MCP server is not running |
| **Description** | Verify that doctor detects when the MCP server process is not active. |
| **Preconditions** | Integration installed; MCP server process has been stopped or failed to start. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output |
| **Expected Result** | Command exits non-zero. Output shows "MCP server: NOT RUNNING" with start instructions. |
| **Story** | 9.3 |

### TC-9.3.5 — health() tool returns error

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.5 |
| **Title** | Doctor detects unresponsive MCP tools |
| **Description** | Verify that doctor reports failure when MCP tools do not respond correctly. |
| **Preconditions** | MCP server process is running but health() returns error or times out. |
| **Steps** | 1. Configure MCP server to return errors on health()<br>2. Run `portfolio doctor`<br>3. Observe output |
| **Expected Result** | Command exits non-zero. Output includes "MCP health check: FAILED" with diagnostic details (error message, timeout info) and suggests checking server logs or restarting. |
| **Story** | 9.3 |

### TC-9.3.6 — Doctor reports integration version

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.6 |
| **Title** | Doctor displays installed version |
| **Description** | Verify that doctor reports the current integration version in the output. |
| **Preconditions** | Integration installed at a known version. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output |
| **Expected Result** | Output includes the installed integration version (e.g., "Claude Code integration: healthy (v1.2.3)"). |
| **Story** | 9.3 |

### TC-9.3.7 — All checks individually reported

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.7 |
| **Title** | Each check has a pass/fail status |
| **Description** | Verify that doctor reports individual status for each verification step. |
| **Preconditions** | Integration installed; mixed state (some checks pass, some fail). |
| **Steps** | 1. Create a scenario where MCP config exists but health() fails<br>2. Run `portfolio doctor`<br>3. Observe output |
| **Expected Result** | Each check (integration metadata, MCP server process, health tool, skill file) has its own line with a pass/fail indicator and message. Clear which checks passed and which failed. |
| **Story** | 9.3 |

### TC-9.3.8 — No false warnings on brand new install

| Field | Value |
|-------|-------|
| **ID** | TC-9.3.8 |
| **Title** | Fresh install passes doctor cleanly |
| **Description** | Verify that immediately after a successful install, doctor reports all checks passing. |
| **Preconditions** | Fresh installation just completed successfully via `portfolio install claude`. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Immediately run `portfolio doctor`<br>3. Observe output |
| **Expected Result** | Doctor exits 0. All checks pass. No stale or misleading warnings. |
| **Story** | 9.3 |

---

## Story 9.4 — Update Integration

### TC-9.4.1 — Happy path: upgrade succeeds

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.1 |
| **Title** | Full upgrade succeeds with version change |
| **Description** | Verify that `portfolio upgrade claude` successfully updates MCP server and skill to a newer version. |
| **Preconditions** | Integration is installed at version X.Y.Z; newer version A.B.C is available; engine is compatible. |
| **Steps** | 1. Record current version and skill file content<br>2. Run `portfolio upgrade claude`<br>3. Observe output<br>4. Check new integration version in metadata<br>5. Check skill file content<br>6. Run `portfolio doctor` |
| **Expected Result** | Command exits 0. Output reports "Upgraded from vX.Y.Z to vA.B.C" with a changelog summary. Integration metadata shows new version. Skill file is updated. Doctor passes after upgrade. |
| **Story** | 9.4 |

### TC-9.4.2 — Already up to date

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.2 |
| **Title** | Upgrade reports already up to date |
| **Description** | Verify that re-running upgrade when already at the latest version reports accordingly. |
| **Preconditions** | Integration is installed at the latest available version. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Observe output and exit code<br>3. Run `portfolio upgrade claude` again<br>4. Observe output and exit code |
| **Expected Result** | Both runs exit 0. Output reports "Already up to date (vX.Y.Z)." No files are modified. |
| **Story** | 9.4 |

### TC-9.4.3 — Integration not installed before upgrade

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.3 |
| **Title** | Upgrade fails when integration not installed |
| **Description** | Verify that upgrade errors when no previous installation exists. |
| **Preconditions** | `portfolio install claude` has never been run; no integration metadata exists. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Observe output and exit code |
| **Expected Result** | Command exits non-zero. Output states "Integration not installed" with instructions to run `portfolio install claude` first. |
| **Story** | 9.4 |

### TC-9.4.4 — Engine version incompatible

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.4 |
| **Title** | Upgrade fails on engine version mismatch |
| **Description** | Verify that upgrade checks engine compatibility before proceeding. |
| **Preconditions** | Integration installed; Portfolio Engine version is below the minimum required version for the integration update. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Observe output and exit code |
| **Expected Result** | Command exits non-zero before making any changes. Output states "Incompatible engine version: requires >= X.Y.Z, found A.B.C" with instructions to upgrade the engine first. |
| **Story** | 9.4 |

### TC-9.4.5 — User configuration preserved after upgrade

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.5 |
| **Title** | User MCP config preserved across upgrade |
| **Description** | Verify that any user-specific MCP configuration is not overwritten during upgrade. |
| **Preconditions** | Integration installed; user has custom MCP configuration fields beyond Portfolio's entry. |
| **Steps** | 1. Record current Claude Code config file content<br>2. Run `portfolio upgrade claude`<br>3. Compare config file before and after |
| **Expected Result** | The Portfolio MCP server entry may be updated, but all other entries and user customizations in the config file are preserved unchanged. |
| **Story** | 9.4 |

### TC-9.4.6 — Rollback on upgrade failure (MCP binary update fails)

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.6 |
| **Title** | Upgrade rolls back if MCP binary update fails |
| **Description** | Verify that if the MCP server binary update fails (e.g., disk full, write error), the system rolls back to the previous state. |
| **Preconditions** | Integration installed at version X.Y.Z; upgrade is available; simulate a failure during binary write (e.g., disk quota exceeded). |
| **Steps** | 1. Record current version and config state<br>2. Run `portfolio upgrade claude` with a simulated failure<br>3. Observe output<br>4. Check integration metadata, MCP config, and skill file |
| **Expected Result** | Command exits non-zero. Output reports "Upgrade failed, rolled back to vX.Y.Z" with error details. Integration metadata is unchanged. MCP config is restored to pre-upgrade state. Skill file is restored to pre-upgrade version. |
| **Story** | 9.4 |

### TC-9.4.7 — Rollback on upgrade failure (verification fails)

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.7 |
| **Title** | Upgrade rolls back if post-upgrade verification fails |
| **Description** | Verify that if the new MCP server fails health checks after upgrade, the system rolls back. |
| **Preconditions** | Integration installed at version X.Y.Z; simulate a scenario where the new MCP server binary starts but health() returns errors. |
| **Steps** | 1. Record current version and config state<br>2. Run `portfolio upgrade claude`<br>3. Observe output<br>4. Check integration metadata, MCP config, and skill file |
| **Expected Result** | Command exits non-zero. Output reports "Upgrade failed verification, rolled back to vX.Y.Z" with details. MCP server binary is replaced with the previous version. Skill file is restored. Metadata is unchanged. |
| **Story** | 9.4 |

### TC-9.4.8 — Upgrade reports version diff and changelog

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.8 |
| **Title** | Upgrade output includes version diff and changes |
| **Description** | Verify that upgrade output shows the version transition and a summary of what changed. |
| **Preconditions** | Integration installed at v1.0.0; v1.1.0 is available. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Observe output |
| **Expected Result** | Output includes "Upgraded from v1.0.0 to v1.1.0" followed by a human-readable summary of changes (new features, bug fixes, breaking changes). |
| **Story** | 9.4 |

### TC-9.4.9 — Help flag shows upgrade claude command

| Field | Value |
|-------|-------|
| **ID** | TC-9.4.9 |
| **Title** | `portfolio upgrade claude` is discoverable via help |
| **Description** | Verify the command appears in help output. |
| **Preconditions** | Portfolio Engine is installed. |
| **Steps** | 1. Run `portfolio --help`<br>2. Run `portfolio upgrade --help` |
| **Expected Result** | `--help` output includes `claude` as a subcommand of `upgrade`. `portfolio upgrade --help` describes the upgrade process including version check, backup, and rollback. |
| **Story** | 9.4 |

---

## Integration & Cross-Story Tests

### TC-INT-9.1 — Full install → doctor → upgrade → doctor lifecycle

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9.1 |
| **Title** | Complete lifecycle: install, verify, upgrade, re-verify |
| **Description** | Verify the full lifecycle from fresh state through install, doctor, upgrade, and post-upgrade verification. |
| **Preconditions** | No integration installed; Claude Code CLI present; Portfolio Engine ready. |
| **Steps** | 1. Run `portfolio doctor` — expect "not installed"<br>2. Run `portfolio install claude` — expect success<br>3. Run `portfolio doctor` — expect healthy<br>4. Run `portfolio upgrade claude` — expect success or "already up to date"<br>5. Run `portfolio doctor` — expect healthy with new version |
| **Expected Result** | Step 1 exits non-zero with not-installed message. Steps 2-3 exit 0 with healthy status. Step 4 exits 0. Step 5 exits 0 with the current version showing correct version. |
| **Story** | 9.1, 9.2, 9.3, 9.4 |

### TC-INT-9.2 — Partial rollback recovery

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9.2 |
| **Title** | Recovery from partially installed integration |
| **Description** | Verify that re-running install after a partial/failed install recovers correctly. |
| **Preconditions** | MCP server config is registered but skill file is missing; or vice versa. |
| **Steps** | 1. Run `portfolio doctor` — shows specific failures<br>2. Run `portfolio install claude`<br>3. Run `portfolio doctor` |
| **Expected Result** | Step 1 reports specific partial state. Step 2 completes all pending installation steps without errors (idempotent for already-completed steps). Step 3 reports all checks passing. No duplicate config entries. |
| **Story** | 9.1, 9.2, 9.3 |

### TC-INT-9.3 — MCP tool availability after install

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9.3 |
| **Title** | MCP tools are callable after integration install |
| **Description** | Verify that after installation, Claude Code can successfully call MCP tools via the configured server. |
| **Preconditions** | Integration installed and verified healthy; Claude Code CLI is available. |
| **Steps** | 1. Call `portfolio-mcp-server health` directly via stdio<br>2. Observe response |
| **Expected Result** | The MCP server responds with a valid health response (status OK). This confirms the configured transport path and binary are correct. |
| **Story** | 9.1 |

### TC-INT-9.4 — Upgrade preserves integration metadata

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9.4 |
| **Title** | Integration metadata is consistent after upgrade |
| **Description** | Verify that integration metadata fields (version, installed_at, updated_at) are correctly updated after an upgrade. |
| **Preconditions** | Integration installed at known version with known installed_at timestamp. |
| **Steps** | 1. Record metadata before upgrade<br>2. Run `portfolio upgrade claude`<br>3. Query integration metadata after upgrade |
| **Expected Result** | `version` is updated to new version. `installed_at` remains unchanged from initial install. `updated_at` is set to the upgrade time. |
| **Story** | 9.4 |

### TC-INT-9.5 — Multiple agent integrations coexist

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9.5 |
| **Title** | Claude Code integration does not conflict with other agent integrations |
| **Description** | Verify that installing the Claude Code integration does not break or conflict with other agent integrations (e.g., Codex CLI). |
| **Preconditions** | Another agent integration (e.g., Codex CLI) is already installed and configured separately. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Verify other agent's skill file and config are unaffected<br>3. Run `portfolio doctor` for both integrations |
| **Expected Result** | Both integrations coexist. Claude Code's install does not modify or remove files belonging to other integrations. Both doctors report healthy. |
| **Story** | 9.1, 9.2 |

### TC-INT-9.6 — Skill file is valid for Claude Code consumption

| Field | Value |
|-------|-------|
| **ID** | TC-INT-9.6 |
| **Title** | Skill file follow Claude Code skill format |
| **Description** | Verify the skill file follows the format that Claude Code expects for auto-discovered skills. |
| **Preconditions** | Skill file installed via `portfolio install claude`. |
| **Steps** | 1. Read the skill file<br>2. Verify format conventions (frontmatter, title, sections, etc.) |
| **Expected Result** | The skill file matches Claude Code's expected skill format — it uses the appropriate file naming convention, starts with a clear title/heading, has well-defined sections for tool descriptions and examples, and is placed in the directory where Claude Code auto-discovers skills. |
| **Story** | 9.2 |

---

## Non-Functional / Performance Tests

### TC-NFR-9.1 — Installation completes in under 5 seconds

| Field | Value |
|-------|-------|
| **ID** | TC-NFR-9.1 |
| **Title** | Install completes within time constraint |
| **Description** | Verify that `portfolio install claude` completes in under 5 seconds on modern hardware. |
| **Preconditions** | Standard development machine; cold caches. |
| **Steps** | 1. Time `portfolio install claude` execution<br>2. Record elapsed time |
| **Expected Result** | Total install time (including detection, registration, skill install, and verification) is under 5 seconds. |
| **Story** | 9.1, 9.2 |

### TC-NFR-9.2 — Verification completes in under 3 seconds

| Field | Value |
|-------|-------|
| **ID** | TC-NFR-9.2 |
| **Title** | Doctor completes within time constraint |
| **Description** | Verify that `portfolio doctor` Claude Code checks complete in under 3 seconds. |
| **Preconditions** | Integration installed and healthy. |
| **Steps** | 1. Time `portfolio doctor` execution<br>2. Record elapsed time |
| **Expected Result** | Total doctor time is under 3 seconds. |
| **Story** | 9.3 |

### TC-NFR-9.3 — No network access required

| Field | Value |
|-------|-------|
| **ID** | TC-NFR-9.3 |
| **Title** | All operations work offline |
| **Description** | Verify that install, doctor, and upgrade work without network access. |
| **Preconditions** | Network is disconnected; integration is not installed (or installed for upgrade/doctor). |
| **Steps** | 1. Disconnect network<br>2. Run `portfolio install claude`, `portfolio doctor`, `portfolio upgrade claude` |
| **Expected Result** | All commands complete successfully (or fail gracefully with local-only errors) without attempting any network calls. No timeouts waiting for network. |
| **Story** | 9.1, 9.3, 9.4 |

### TC-NFR-9.4 — All operations are idempotent

| Field | Value |
|-------|-------|
| **ID** | TC-NFR-9.4 |
| **Title** | All integration operations are idempotent |
| **Description** | Verify that install, doctor, and upgrade can be run multiple times with consistent results. |
| **Preconditions** | Integration installed; any state. |
| **Steps** | 1. Run `portfolio install claude` three times in a row<br>2. Run `portfolio doctor` three times in a row<br>3. Run `portfolio upgrade claude` three times in a row |
| **Expected Result** | All commands produce identical exit codes and similar output on each run. No duplicate config entries, no skill file duplicates, no side effects from re-running. |
| **Story** | 9.1, 9.3, 9.4 |

### TC-NFR-9.5 — Engine code is not modified

| Field | Value |
|-------|-------|
| **ID** | TC-NFR-9.5 |
| **Title** | Integration does not modify engine code |
| **Description** | Verify that installing the Claude Code integration does not modify any Portfolio Engine files. |
| **Preconditions** | Portfolio Engine installed at known path; integration not yet installed. |
| **Steps** | 1. Record checksums of all engine binary files<br>2. Run `portfolio install claude`<br>3. Recompute checksums and compare |
| **Expected Result** | All engine file checksums are identical before and after install. The integration only creates/writes files in Claude Code's config and skill directories and in the integration metadata store. |
| **Story** | 9.1 |

---

## Edge Cases

### TC-EDGE-9.1 — Multiple Claude Code config locations

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.1 |
| **Title** | Handles multiple possible Claude Code config locations |
| **Description** | Verify that install detects and uses the correct config location when multiple possible locations exist. |
| **Preconditions** | Both `claude_desktop_config.json` and `~/.claude/settings.json` exist; one contains a Portfolio entry, the other does not. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Check which config file was modified<br>3. Run `portfolio doctor` |
| **Expected Result** | The command selects the standard/preferred location based on priority. Only one config file is modified. Doctor reports consistent state. |
| **Story** | 9.1 |

### TC-EDGE-9.2 — Config file contains invalid JSON

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.2 |
| **Title** | Handles corrupt Claude Code config file |
| **Description** | Verify that install handles a malformed config file gracefully. |
| **Preconditions** | Claude Code config file exists but contains invalid JSON (e.g., trailing comma, unclosed brace). |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output and exit code |
| **Expected Result** | Command exits non-zero. Error message includes the config file path and the JSON parse error. Suggests manual repair or backing up and regenerating the file. Does NOT overwrite the corrupt file to avoid data loss. |
| **Story** | 9.1 |

### TC-EDGE-9.3 — Skill file deleted after install

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.3 |
| **Title** | Doctor detects deleted skill file |
| **Description** | Verify that doctor correctly reports when the skill file is deleted post-install. |
| **Preconditions** | Integration was installed and healthy; skill file manually deleted. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output |
| **Expected Result** | Doctor reports "Skill file: MISSING" with remediation. Does NOT report the integration as entirely missing (metadata still exists, MCP may still be configured). |
| **Story** | 9.3 |

### TC-EDGE-9.4 — MCP config entry manually removed

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.4 |
| **Title** | Doctor detects removed MCP config entry |
| **Description** | Verify that doctor correctly reports when the MCP config entry is removed post-install. |
| **Preconditions** | Integration was installed; user manually removed the Portfolio MCP entry from Claude Code config. |
| **Steps** | 1. Run `portfolio doctor`<br>2. Observe output |
| **Expected Result** | Doctor reports "MCP server: NOT CONFIGURED" with remediation to run `portfolio install claude`. |
| **Story** | 9.3 |

### TC-EDGE-9.5 — Upgrade with no newer version available (fresh install)

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.5 |
| **Title** | Upgrade on latest integration matches current version |
| **Description** | Verify that running upgrade immediately after a fresh install reports "already up to date." |
| **Preconditions** | Integration was just installed to latest version. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Observe output |
| **Expected Result** | Command exits 0. Output reports "Already up to date (vX.Y.Z)." No backup files remain after command completion. |
| **Story** | 9.4 |

### TC-EDGE-9.6 — Install interrupted mid-way

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.6 |
| **Title** | Recovery from interrupted install |
| **Description** | Verify that if `portfolio install claude` is interrupted (SIGINT, power loss), the system can recover. |
| **Preconditions** | Integrations state is unknown — MCP may be partially configured. |
| **Steps** | 1. Simulate an interruption during install (e.g., kill the process after MCP config is written but before skill is installed)<br>2. Run `portfolio install claude` again |
| **Expected Result** | Re-running install completes successfully. Idempotency ensures MCP config is not duplicated. Missing skill file is installed. Doctor passes. |
| **Story** | 9.1, 9.2 |

### TC-EDGE-9.7 — Symbolic link in config or skill path

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.7 |
| **Title** | Handles symlinked config/skills directories |
| **Description** | Verify that install works correctly when the Claude Code config or skills directory is a symbolic link. |
| **Preconditions** | `~/.claude/` is a symlink to another directory; or config/skills directories are symlinked. |
| **Steps** | 1. Run `portfolio install claude`<br>2. Verify files are written to the symlink target<br>3. Run `portfolio doctor` |
| **Expected Result** | Installation succeeds. Files are written to the target of the symlink (not overwriting the symlink itself). Doctor passes. |
| **Story** | 9.1, 9.2 |

### TC-EDGE-9.8 — Non-standard Claude Code install path

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.8 |
| **Title** | Handles non-standard Claude Code location |
| **Description** | Verify that install works when Claude Code is installed in a non-default location. |
| **Preconditions** | Claude Code CLI is installed in a custom location (not in PATH or not in the default expected path). |
| **Steps** | 1. Run `portfolio install claude`<br>2. Observe output |
| **Expected Result** | The integration uses configurable discovery mechanisms to find Claude Code. If found, install proceeds normally. If not found, it reports a clear error with the paths checked. |
| **Story** | 9.1 |

### TC-EDGE-9.9 — Collision: skill file name conflict

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.9 |
| **Title** | Skill file name does not collide with other skills |
| **Description** | Verify that the Portfolio skill uses a unique name that won't collide with other installed skills. |
| **Preconditions** | Other skill files already exist in the Claude Code skills directory. |
| **Steps** | 1. List existing skills before install<br>2. Run `portfolio install claude`<br>3. List skills after install |
| **Expected Result** | The Portfolio skill file is named uniquely (e.g., `portfolio.md` or `portfolio-*.md`). Existing skill files are not modified or deleted. |
| **Story** | 9.2 |

### TC-EDGE-9.10 — Version comparison with pre-release tags

| Field | Value |
|-------|-------|
| **ID** | TC-EDGE-9.10 |
| **Title** | Version comparison handles pre-release and build metadata |
| **Description** | Verify that the upgrade version comparison correctly handles semver pre-release tags. |
| **Preconditions** | Integration installed at v1.0.0-alpha.1; v1.0.0 is available. Or installed at v1.0.0; v1.0.0-beta.1 is available. |
| **Steps** | 1. Run `portfolio upgrade claude`<br>2. Observe output |
| **Expected Result** | Semver precedence is respected. v1.0.0 > v1.0.0-alpha.1 (pre-release treated as lower). Upgrade from v1.0.0-alpha.1 to v1.0.0 is offered. Upgrade from v1.0.0 to v1.0.0-beta.1 is NOT offered (stable > pre-release). |
| **Story** | 9.4 |

---

## Error Message Quality Tests

### TC-ERR-9.1 — Error messages include actionable remediation

| Field | Value |
|-------|-------|
| **ID** | TC-ERR-9.1 |
| **Title** | All error messages include actionable remediation |
| **Description** | Verify that every error condition produces a message telling the user what to do next. |
| **Preconditions** | Any failure scenario. |
| **Steps** | 1. Induce each failure scenario<br>2. Observe error output |
| **Expected Result** | Every error message includes not just what went wrong, but what the user should do to fix it (e.g., "Run 'portfolio install claude' to install", "Upgrade Portfolio Engine to >= X.Y.Z", "Check the binary at /path/to/mcp-server"). |
| **Story** | 9.1, 9.3, 9.4 |

### TC-ERR-9.2 — Exit codes are correct per scenario

| Field | Value |
|-------|-------|
| **ID** | TC-ERR-9.2 |
| **Title** | Correct exit codes for success and failure |
| **Description** | Verify that all commands use exit code 0 for success and non-zero for failure. |
| **Preconditions** | Various states of integration. |
| **Steps** | 1. Run each command in success and failure scenarios<br>2. Check exit codes |
| **Expected Result** | Exit 0 for: fresh install success, re-install (already installed), doctor healthy, upgrade success, already up to date. Exit non-zero for: Claude Code not found, permission denied, MCP health fails, skill missing, engine incompatible, upgrade failure with rollback. |
| **Story** | 9.1, 9.3, 9.4 |

### TC-ERR-9.3 — Error output goes to stderr

| Field | Value |
|-------|-------|
| **ID** | TC-ERR-9.3 |
| **Title** | Error output is on stderr |
| **Description** | Verify that error messages are written to stderr, not stdout. |
| **Preconditions** | Any failure scenario. |
| **Steps** | 1. Run a failing command<br>2. Capture stdout and stderr separately |
| **Expected Result** | Normal/progress output goes to stdout. Error messages and diagnostics go to stderr. |
| **Story** | 9.1, 9.3, 9.4 |
