# Architecture Review: Epic 8 — Integration Framework

## Status: REJECTED (Corrected Version Approved)

## Previous Approval Rescinded

The original architecture was **approved** but contained significant over-engineering. The following issues were identified during cross-document audit against Guideline.md engineering principles:

| Issue | Original | Corrected |
|-------|----------|-----------|
| State machine | 6 states (not-installed, installing, installed, validated, upgrading, removing) with 19 transitions + `state.go` | 3 states (not_installed, installed). 7 transitions. `Status` is a field on `IntegrationMeta`, not a separate file. |
| Self-healing | Doctor automatically attempted fixes without opt-in | Doctor reports only. `--fix` flag applies self-healable fixes. |
| File lock | `~/.portfolio/.integration.lock` with `flock`-style advisory locking | Removed. Single-user CLI. SQLite serializes metadata writes. |
| Factory registration | Config-driven factory with `init()` global registry, blank imports, config.yaml entries | Direct import in Manager constructor. 3 agent types don't justify the indirection. |
| Backup as Store methods | `SaveBackup`, `GetBackup`, `DeleteBackup` on `Store` interface | File-level backup on disk under `integration/<name>/backup/`. No schema coupling. |

## Cross-Reference Audit Against Authoritative Docs

| Document | Alignment | Notes |
|----------|-----------|-------|
| **KnowledgeModel.md** | ✅ Strong | No new entities. Agent-agnostic per principle 7. |
| **PlatformSpecification.md** | ✅ Strong | Reuses `configuration` table. No schema changes. |
| **Guideline.md** | ✅ Strong | "Deterministic by Default" — 3-state machine, report-only doctor, no auto-healing. "CLI is Administrative" — CLI-only. "Agent Agnostic" — Integration interface. |
| **ADR-013** | ✅ Strong | Direct implementation of integration architecture. |

## Cross-Epic Consistency

| Epic | Interface | Status |
|------|-----------|--------|
| Epic 7 (MCP Server) | MCPClient interface | ✅ Consistent |
| Epic 9 (Claude Code) | Implements Integration interface | ✅ Consistent — Claude upgrade reuses Manager upgrade flow (no standalone upgrade.go) |
| Epic 10 (AI Analysis) | Analyzer identity string | ✅ Consistent |
| Epic 11 (Dashboard) | No integration management in dashboard | ✅ Consistent |

## Decision Summary

Corrected design: Integration interface, Manager, Store (4 methods, no backup), MCPClient, CLI, 3-state machine (no state.go), doctor with --fix (no auto-heal), upgrade with file-level rollback (Manager handles snapshot/restore), direct imports in Manager (no init() registry).

---

## Findings

### Strengths

1. **Pragmatic state machine** — 3 states, 7 transitions. No persistence of intermediate states. No separate `state.go`.

2. **Doctor/Fix separation** — Report by default, fix by opt-in (`--fix`). Correctly separates diagnosis from remediation.

3. **Direct integration construction** — Manager imports agent packages directly. Config-driven factory with `init()` registry adds indirection for no benefit.

4. **File-level backup** — Upgrade backup stored on disk, not in configuration table. No schema coupling.

5. **Absence of file lock** — Correct for a single-user CLI tool. SQLite serialization is sufficient.

### Issues

None.

### Considerations

1. **Agent packages inside `internal/` means new agents require engine rebuilds** — Acceptable for M2. Future externalization should be tracked post-M2.

2. **`--fix` governance** — The list of self-healable fixes will grow. Recommend documented boundaries with explicit review gate for additions.

3. **Process isolation is at goroutine level** — `recover()` guards. Process-level isolation deferred to post-M2.

---

## Recommendation

**Corrected version approved.** Ready for implementation.

## Review Checklist

| Criterion | Status | Notes |
|-----------|--------|-------|
| Requirement Alignment | ✅ Pass | All FRs and NFRs addressed |
| Agent Agnosticism | ✅ Pass | Integration interface |
| Design Simplicity | ✅ Pass | 3 states, report-only doctor, direct imports |
| Edge Case Coverage | ✅ Pass | 11 ECs defined with specific handling |
| Error Handling | ✅ Pass | Typed errors with codes and remediation |
| Test Coverage | ✅ Pass | Fakes, unit, integration tests defined |
| Rollback Strategy | ✅ Pass | File-level backup with atomic restore |
| Store Design | ✅ Pass | 4 methods, no backup coupling |
| MCP Integration | ✅ Pass | RegisterTools on MCPClient interface |
| CLI Design | ✅ Pass | All subcommands mapped to Manager |
| Self-Heal Scope | ✅ Pass | Report only; --fix flag for opt-in remediation |
