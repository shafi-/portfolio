# Epic 12 — Dashboard Frontend

**Milestone:** 3 — Dashboard
**Status:** todo

## Overview

Implement read-only dashboard frontend with portfolio overview, project list, project detail, relationship explorer, and statistics pages.

---

## Story 12.1: Portfolio Overview Page

**Status:** todo
**Size:** M
**Blocked by:** 11.2

**User Story:**
As a user, I want a portfolio overview so that I can see my entire software portfolio at a glance.

**Acceptance Criteria:**
- Displays: total projects, active projects, recent discoveries
- Technology summary chart
- Activity timeline (recent modifications)
- Links to project list and detail pages
- Updates automatically (refresh, no live polling)

**Technical Context:**
- Per PlatformSpecification.md and UXGuidelines.md

---

## Story 12.2: Project List Page

**Status:** todo
**Size:** M
**Blocked by:** 12.1

**User Story:**
As a user, I want a project list so that I can browse all my projects.

**Acceptance Criteria:**
- Table or grid view of all projects
- Columns: name, path, languages, last modified, analysis status
- Search bar
- Filters: by technology, by analysis status
- Sorting: by name, by date, by language
- Pagination for large portfolios

---

## Story 12.3: Project Detail Page

**Status:** todo
**Size:** L
**Blocked by:** 12.2

**User Story:**
As a user, I want a project detail page so that I can see everything about a specific project.

**Acceptance Criteria:**
- Sections: Metadata, Documentation, Analysis (if available), Relationships
- Metadata: git info, languages, frameworks, dependencies, statistics
- Documentation: list of indexed docs with content preview
- Analysis: summary, purpose, architecture, features (if analyzed)
- Relationships: links to related projects
- Clear indication when analysis is missing/outdated

---

## Story 12.4: Relationship Explorer

**Status:** todo
**Size:** L
**Blocked by:** 10.4, 12.3

**User Story:**
As a user, I want to explore relationships so that I can understand how projects connect.

**Acceptance Criteria:**
- Visual representation of relationships (graph or list)
- Filter by relationship type
- Navigate to related projects
- Shows relationship description and confidence
- Handles projects with no relationships

---

## Story 12.5: Statistics Page

**Status:** todo
**Size:** M
**Blocked by:** 12.1

**User Story:**
As a user, I want portfolio statistics so that I can understand my development patterns.

**Acceptance Criteria:**
- Technology distribution (chart)
- Project maturity distribution
- Activity over time
- Projects with/without analysis
- Top technologies by project count

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 12.1 Portfolio Overview Page | todo | M | 11.2 |
| 12.2 Project List Page | todo | M | 12.1 |
| 12.3 Project Detail Page | todo | L | 12.2 |
| 12.4 Relationship Explorer | todo | L | 10.4, 12.3 |
| 12.5 Statistics Page | todo | M | 12.1 |

**Total Size:** 2L + 3M = ~21 days

**Can Start:** Story 12.1 (after 11.2 complete)
