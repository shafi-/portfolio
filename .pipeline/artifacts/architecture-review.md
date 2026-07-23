# Epic 3 — Metadata Extraction: Architecture Review

**Reviewer:** Architect Advisor
**Date:** 2026-07-22
**Status:** Approved with Minor Comments

---

## Review Checklist

### 1. Requirements Coverage

| Req | Description | Covered? | Notes |
|-----|------------|----------|-------|
| F-3.1 | Git Metadata | ✅ | Section 1.7 — command-per-field table with edge cases |
| F-3.2 | Language Detection | ✅ | Section 1.8 — Walk → extension → sort by prevalence |
| F-3.3 | Framework Detection | ✅ | Section 1.9 — manifest scanning with marker rules |
| F-3.4 | Dependency Detection | ✅ | Section 1.10 — 8 manifest formats, top-10 + full table |
| F-3.5 | Project Statistics | ✅ | Section 1.11 — binary skip, sampling for >100k |
| F-3.6 | Documentation Hashes | ✅ | Section 1.12 — SHA-256 hash-of-hashes |
| NFR-1 | Determinism | ✅ | All capabilities are purely deterministic |
| NFR-2 | Performance | ✅ | <5s/30s targets, sampling strategy, early-pruning walk |
| NFR-3 | Isolation | ✅ | Each capability is independent, no shared state |
| NFR-4 | Idempotency | ✅ | UpsertMetadata + deterministic computation |
| NFR-5 | Disk Respect | ✅ | `fs.SkipDir` at walk level, not post-filter |
| NFR-6 | Graceful Degradation | ✅ | Section 1.13 — accumulate partial results |

### 2. Edge Case Coverage

| EC | Description | Covered? | Notes |
|----|------------|----------|-------|
| EC-1 | Empty repo | ✅ | NULL git_head, commit_count=0 |
| EC-2 | Bare repo | ✅ | .git missing → skip lang/framework |
| EC-3 | Detached HEAD | ✅ | SHA still valid, branch from remote HEAD |
| EC-4 | No recognized language | ✅ | language_summary = NULL |
| EC-5 | Unknown extensions | ✅ | Not included in language_summary |
| EC-6 | Polyglot 20+ | ✅ | No truncation, all sorted |
| EC-7 | Multi-framework | ✅ | Union of all matches |
| EC-8 | Monorepo | ✅ | Each manifest parsed independently |
| EC-9 | Corrupted manifest | ✅ | Skip, log WARN, continue |
| EC-10 | >100k files | ✅ | Sampling for LOC |
| EC-11 | No doc files | ✅ | Empty string → NULL |
| EC-12 | Deleted/moved repo | ✅ | os.Stat check → skip |
| EC-13 | Symlink escape | ✅ | filepath.EvalSymlinks |
| EC-14 | Non-UTF-8 docs | ✅ | Hash raw bytes |
| EC-15 | No remote | ✅ | Fallback to local HEAD ref |
| EC-16 | Permission error | ✅ | Log WARN, skip file |

### 3. Architectural Soundness

| Criterion | Verdict | Notes |
|-----------|---------|-------|
| Package structure | ✅ | Clean separation: internal/metadata/, internal/store/, pkg/models/ |
| Interface design | ✅ | Store + FileWalker interfaces allow testability |
| Capability separation | ✅ | Each story is a standalone exported function |
| Error handling | ✅ | 4-category classification + partial-failure orchestrator |
| Migration strategy | ✅ | ALTER TABLE IF NOT EXISTS with column-existence check |
| MCP tool mapping | ✅ | 6 individual + 1 aggregate tool, capabilities over workflows |
| Test strategy | ✅ | Real git repos as fixtures, integration tests |
| Implementation order | ✅ | Well-sequenced, walk.go as shared dependency |

### 4. Consistency with Project Principles

| Principle | Alignment | Notes |
|-----------|-----------|-------|
| Engine Knows, Agent Thinks | ✅ | All 6 stories are purely deterministic |
| Deterministic by Default | ✅ | Same HEAD → same metadata |
| Store Facts, Compute Indicators | ✅ | documentation_hash stored; computed fields deferred |
| Capabilities over Workflows | ✅ | Independent capability functions |
| Local First | ✅ | All extraction from local filesystem |
| AI is Optional | ✅ | Metadata extraction provides value without AI |
| Single Knowledge Model | ✅ | All outputs map to metadata table |

---

## Issues Found

### 🔴 Critical Issues

None.

### 🟡 Minor Issues

1. **Inconsistent `ComputeDocHash` signatures** (§1.4 vs §1.12)
   - §1.4 shows `func(docPaths []string)` but §1.12 shows `ComputeDocHash(docContents [][]byte)`.
   - **Fix:** Align to a single signature. The `[][]byte` variant is more correct since it separates I/O from hashing.

2. **Binary file detection is underspecified** (§1.11)
   - "check first 512 bytes, or extension-based" — two approaches, no commitment.
   - **Fix:** Pick one. Extension-based is faster (NFR-2) but misses edge cases. Recommend extension blacklist + first-512-bytes for ambiguous cases.

3. **Story 3.6 ↔ Epic 4 interface not defined** (§1.13)
   - Orchestrator doesn't show how doc paths are obtained. Currently says "Epic 4 provides sorted paths" in a comment.
   - **Fix:** Define the interface contract now (e.g., `DocPathProvider` interface with `func GetDocPaths(projectID string) ([]string, error)`).

4. **No thread-safety discussion** (§5.2 — TestConcurrentExtraction)
   - Integration test plans concurrent extraction but nothing addresses SQLite concurrent access or walker goroutine safety.
   - **Fix:** Add note about SQLite WAL mode and whether Store methods are safe for concurrent callers.

5. **Story 3.6 estimate likely too low** (§6 — 1 day)
   - Context pack rates it "M" (medium), same as 3.3 and 3.4 (both estimated at 3 days).
   - **Suggestion:** Bump to 2 days to account for Epic 4 integration wiring.

---

## Open Questions from Requirements (Closed)

| Q# | Question | Architecture Decision | Verdict |
|----|----------|----------------------|---------|
| Q1 | `dependency_summary` as JSON array vs comma-separated string | Chose comma-separated string | ✅ Acceptable for top-10 display |
| Q2 | Sampling approach for >100k files | Every Nth file, N=ceil(totalFiles/10000) | ✅ Clear and implementable |
| Q3 | Hash-of-hashes vs rolling hash | Hash-of-hashes (sort → sha256 each → combine) | ✅ More debuggable |
| Q4 | Built-in mapping vs config file | Embedded defaults + TOML overrides | ✅ Follows convention |

---

## Cross-Reference Audit Against Authoritative Docs

| Document | Alignment | Notes |
|----------|-----------|-------|
| **KnowledgeModel.md** | ✅ Strong | Metadata entity (git info, languages, frameworks, dependencies, hashes) maps directly to KnowledgeModel's Engine-owned fields. |
| **PlatformSpecification.md** | ✅ Strong | Schema additions in §2.1 match PlatformSpec specification. `commit_count` is valid extension per design principles. |
| **Guideline.md** | ✅ Strong | "Engine Knows, Agent Thinks" — all 6 stories are purely deterministic. "Deterministic by Default" — same HEAD → same metadata. "Store Facts, Compute Indicators" — documentation_hash stored; computed fields deferred. |
| **ADR-015** (KnowledgeModel canonical) | ✅ Strong | Metadata fields derive from KnowledgeModel.md. No redefinition. |
| **Architecture.md** (High-level) | ✅ Strong | Metadata extraction is an Engine responsibility per Architecture.md. |

## Cross-Epic Consistency

| Epic | Interface | Status |
|------|-----------|--------|
| Epic 2 (Discovery) | Projects must exist before metadata extraction | ✅ Consistent — MetadataService.ExtractAll loads project from store |
| Epic 5 (Knowledge Store) | Store interface: UpsertMetadata, GetMetadata, ReplaceDependencies | ✅ Consistent — signatures match expected store pattern |
| Epic 7 (MCP Server) | Individual metadata extraction tools | ✅ Consistent — 6 individual + 1 aggregate tool |
| Epic 14 (Insights) | Activity timeline depends on metadata.last_commit_at | ✅ Consistent — field is present and documented |
| Epic 10 (AI Analysis) | Staleness detection uses metadata.git_head | ✅ Consistent — field is present |

**Note:** The `source_project != target_project` CHECK constraint from Epic 13's relationships schema should not affect metadata — no overlap.

## Overall Approval Status

**✅ APPROVED with Minor Comments**

The architecture is thorough, well-structured, and strongly aligned with project principles. All 16 functional requirements are fully covered. All 16 edge cases are addressed. The capability-separated design and partial-failure orchestration are clean. Fix the two signature inconsistencies and document the Epic 4 interface before starting Story 3.6.
