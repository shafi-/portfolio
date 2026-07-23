# Epic 9 — Claude Code Integration: Architecture Review

**Status: REJECTED (Corrected Version Approved)**

## Previous Approval Rescinded

The original architecture was **approved** but contained significant over-engineering and duplication:

| Issue | Original | Corrected |
|-------|----------|-----------|
| Upgrade/rollback | Standalone `upgrade.go` with full backup/restore/verify/rollback — duplicating Epic 8 Manager | Integration `Upgrade()` does agent-specific work only (write config + skill). Manager handles snapshot, rollback, verify. |
| Skill version injection | `text/template` injecting version into skill markdown | Static markdown. Version tracked in metadata DB. No template engine. |
| Atomic temp+rename writes | Elevated to design principle with `backupMCPConfig()`/`restoreMCPConfig()` | `os.WriteFile` directly. Temp+rename is internal detail, not exposed in design. |
| In-memory backup snapshots | Full file content in `[]byte` as backup strategy | Manager handles file-level backup on disk. Integration doesn't implement backup at all. |

## Cross-Reference Audit Against Authoritative Docs

| Document | Alignment | Notes |
|----------|-----------|-------|
| **KnowledgeModel.md** | ✅ Strong | No new entities. Skill references canonical MCP tools. |
| **PlatformSpecification.md** | ✅ Strong | Skill lists MCP tools from PlatformSpec §2. Config uses correct file structure. |
| **Guideline.md** | ✅ Strong | "Deterministic by Default" — no upgrade/rollback duplication. "Capabilities over Workflows" — static skill, no template. |
| **ADR-013** | ✅ Strong | Direct implementation — first concrete integration. |
| **Epic 8 (Corrected)** | ✅ Strong | Upgrade lifecycle fully owned by Manager. Integration handles agent-specific mutations only. |

## Cross-Epic Consistency

| Epic | Interface | Status |
|------|-----------|--------|
| Epic 7 (MCP Server) | Consumes Engine via MCP tools | ✅ Consistent |
| Epic 8 (Integration Framework) | Implements Integration interface | ✅ Consistent — no duplicate upgrade/rollback |
| Epic 10 (AI Analysis) | Agent identity "claude-code" | ✅ Consistent |
| Epic 11 (Dashboard) | No integration management in dashboard | ✅ Consistent |

## Findings

### Strengths

1. **No upgrade duplication** — Integration handles agent-specific mutations only. Manager owns snapshot, rollback, verify. Correct separation of concerns.

2. **Static skill file** — No template engine. Version in metadata DB. Simpler and sufficient.

3. **Cleaner package structure** — No `upgrade.go`, no `framework/` sub-package. Integration struct directly implements Epic 8 interface.

4. **Test boundaries clear** — Rollback/corruption recovery tested at Epic 8 level, not duplicated here.

### Issues

None.

---

## Recommendation

**Corrected version approved.** Ready for implementation.

## Review Checklist

| Criterion | Status | Notes |
|-----------|--------|-------|
| Requirements Coverage | ✅ Pass | All FRs and NFRs addressed |
| Architectural Alignment | ✅ Pass | No Epic 8 duplication |
| Design Completeness | ✅ Pass | Agent-specific work only |
| Test Strategy | ✅ Pass | Sandbox pattern, no duplicated tests |
| Upgrade Separation | ✅ Pass | Manager handles lifecycle; integration handles mutations |
