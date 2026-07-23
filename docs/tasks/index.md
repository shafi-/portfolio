# Tasks Index

Portfolio Implementation Roadmap — Epic Index

---

## Overview

This directory contains the complete implementation roadmap for Portfolio, organized by epic. Each epic file contains detailed user stories with acceptance criteria, dependencies, and status tracking.

**Status Values:**
- `todo` — Not started
- `in_progress` — Currently being worked on
- `completed` — Done
- `blocked` — Waiting on dependencies or external factors

---

## Legend

**Dependencies:**
- `Blocked by: [story]` — This story cannot start until the blocking story is complete
- `Unblocks: [story]` — This story must be complete before the dependent story can start

**Sizing:**
- `XS` — < 1 day
- `S` — 1-2 days
- `M` — 3-5 days
- `L` — 1-2 weeks
- `XL` — > 2 weeks

---

## Milestone 1 — Core Engine

**Total Estimated Size:** ~74 days

| Epic | Status | Stories | Size | File |
|------|--------|---------|------|------|
| [Epic 1 — Project Foundation](epic-01-project-foundation.md) | todo | 5 | ~8 days | [epic-01-project-foundation.md](epic-01-project-foundation.md) |
| [Epic 2 — Discovery](epic-02-discovery.md) | todo | 5 | ~17 days | [epic-02-discovery.md](epic-02-discovery.md) |
| [Epic 3 — Metadata Extraction](epic-03-metadata-extraction.md) | todo | 6 | ~18 days | [epic-03-metadata-extraction.md](epic-03-metadata-extraction.md) |
| [Epic 4 — Documentation Indexing](epic-04-documentation-indexing.md) | todo | 5 | ~15 days | [epic-04-documentation-indexing.md](epic-04-documentation-indexing.md) |
| [Epic 5 — Knowledge Store](epic-05-knowledge-store.md) | todo | 3 | ~12 days | [epic-05-knowledge-store.md](epic-05-knowledge-store.md) |
| [Epic 6 — HTTP API](epic-06-http-api.md) | todo | 5 | ~7 days | [epic-06-http-api.md](epic-06-http-api.md) |
| [Epic 7 — MCP Server](epic-07-mcp-server.md) | todo | 5 | ~14 days | [epic-07-mcp-server.md](epic-07-mcp-server.md) |

**Milestone 1 Progress:** 0/7 epics completed (0%)

---

## Milestone 2 — Agent Integration

**Total Estimated Size:** ~34 days

| Epic | Status | Stories | Size | File |
|------|--------|---------|------|------|
| [Epic 8 — Integration Framework](epic-08-integration-framework.md) | todo | 4 | ~11 days | [epic-08-integration-framework.md](epic-08-integration-framework.md) |
| [Epic 9 — Claude Code Integration](epic-09-claude-code-integration.md) | todo | 4 | ~11 days | [epic-09-claude-code-integration.md](epic-09-claude-code-integration.md) |
| [Epic 10 — AI Analysis](epic-10-ai-analysis.md) | todo | 4 | ~16 days | [epic-10-ai-analysis.md](epic-10-ai-analysis.md) |

**Milestone 2 Progress:** 0/3 epics completed (0%)

---

## Milestone 3 — Dashboard

**Total Estimated Size:** ~28 days

| Epic | Status | Stories | Size | File |
|------|--------|---------|------|------|
| [Epic 11 — Dashboard Backend](epic-11-dashboard-backend.md) | todo | 2 | ~4 days | [epic-11-dashboard-backend.md](epic-11-dashboard-backend.md) |
| [Epic 12 — Dashboard Frontend](epic-12-dashboard-frontend.md) | todo | 5 | ~21 days | [epic-12-dashboard-frontend.md](epic-12-dashboard-frontend.md) |

**Milestone 3 Progress:** 0/2 epics completed (0%)

---

## Milestone 4 — Portfolio Intelligence

**Total Estimated Size:** ~27 days

| Epic | Status | Stories | Size | File |
|------|--------|---------|------|------|
| [Epic 13 — Relationships](epic-13-relationships.md) | todo | 1 | ~4 days | [epic-13-relationships.md](epic-13-relationships.md) |
| [Epic 14 — Insights](epic-14-insights.md) | todo | 4 | ~18 days | [epic-14-insights.md](epic-14-insights.md) |

**Milestone 4 Progress:** 0/2 epics completed (0%)

---

## Overall Progress

**Total Epics:** 14
**Completed:** 0
**In Progress:** 0
**Todo:** 14

**Total Estimated Size:** ~163 days (~6-7 months for one developer)

---

## Dependency Graph

```
Epic 1 (Project Foundation)
    ↓
Epic 2 (Discovery)
    ↓
Epic 3 (Metadata Extraction)
    ↓
Epic 4 (Documentation Indexing) → Epic 5 (Knowledge Store)
    ↓                              ↓
Epic 6 (HTTP API) ←───────────────┘
    ↓
Epic 7 (MCP Server)
    ↓
─────────────────────────────────────────────────
Epic 8 (Integration Framework)
    ↓
Epic 9 (Claude Code Integration)
    ↓
Epic 10 (AI Analysis)
    ↓
─────────────────────────────────────────────────
Epic 11 (Dashboard Backend) ← Epic 6
    ↓
Epic 12 (Dashboard Frontend)
    ↓
─────────────────────────────────────────────────
Epic 13 (Relationships) ← Epic 10
    ↓
Epic 14 (Insights)
```

---

## Can Start Now

These stories have no dependencies and can be started immediately:

- **Story 1.1** — Bootstrap Go Project (Epic 1)

---

## Definition of Done

A story is complete when:

1. **Acceptance criteria are satisfied** — All criteria in the story pass
2. **Tests are added for deterministic logic** — Engine functionality has automated tests
3. **Documentation is updated** — If the story changes the domain model or APIs:
   - Update KnowledgeModel.md
   - Update PlatformSpecification.md
   - Update CLAUDE.md
4. **Architectural guidelines are respected** — Implementation follows Guideline.md principles
5. **No regressions** — Existing functionality remains working

---

## Updating Status

When working on stories:

1. Update the story's `Status:` field in its epic file
2. Update epic status summaries as stories complete
3. Update this index's progress tables
4. When an epic is complete, update its `Status:` at the top of the file

---

## Related Documentation

- [PRD.md](../PRD.md) — Product Requirements Document
- [Guideline.md](../Guideline.md) — Engineering Principles
- [KnowledgeModel.md](../KnowledgeModel.md) — Canonical Domain Model
- [PlatformSpecification.md](../PlatformSpecification.md) — Implementation Contracts
- [TechStack.md](../TechStack.md) — Technology Stack
- [UXGuidelines.md](../UXGuidelines.md) — Dashboard UX Guidelines
- [ADR.md](../ADR.md) — Architecture Decision Records

---

*This roadmap derives from the authoritative specifications above. All implementation should follow the documented principles and contracts.*
