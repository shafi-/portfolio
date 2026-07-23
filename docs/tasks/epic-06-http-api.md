# Epic 6 — HTTP API

**Milestone:** 1 — Core Engine
**Status:** todo

## Overview

Implement RESTful HTTP API endpoints for health, projects, search, configuration, statistics, and relationships to serve the dashboard.

---

## Story 6.1: Health Endpoint

**Status:** todo
**Size:** XS
**Blocked by:** 5.1

**User Story:**
As the Portfolio Engine, I want a health endpoint so that the dashboard can verify engine status.

**Acceptance Criteria:**
- GET /health returns 200 with engine status
- Returns: status (healthy/unhealthy), database_connected, project_count
- Response: JSON

---

## Story 6.2: Projects API

**Status:** todo
**Size:** M
**Blocked by:** 6.1, 5.1

**User Story:**
As the dashboard, I want to list and retrieve projects so that users can browse their portfolio.

**Acceptance Criteria:**
- GET /projects → list of all projects with metadata
- GET /projects/{id} → single project with full details
- Query parameters for filtering, sorting, pagination
- Returns 404 for non-existent projects

**Response Format:**
```json
{
  "id": "uuid",
  "name": "project-name",
  "root_path": "/path/to/project",
  "repository_type": "git",
  "discovered_at": "timestamp",
  "updated_at": "timestamp",
  "metadata": { ... }
}
```

---

## Story 6.3: Search API

**Status:** todo
**Size:** M
**Blocked by:** 6.2

**User Story:**
As the dashboard, I want search endpoints so that users can find projects and documents.

**Acceptance Criteria:**
- GET /search?q=query → search projects and documents
- Query searches: project names, technologies, document contents
- Returns ranked results with type (project/document)
- Fast response (<500ms)

---

## Story 6.4: Configuration API

**Status:** todo
**Size:** S
**Blocked by:** 6.1

**User Story:**
As the dashboard, I want to read configuration so that it can display engine settings.

**Acceptance Criteria:**
- GET /configuration → current config (project roots, ignored paths)
- PATCH /configuration → update config (requires validation)
- Returns 400 for invalid configuration

---

## Story 6.5: Statistics API

**Status:** todo
**Size:** S
**Blocked by:** 6.2

**User Story:**
As the dashboard, I want portfolio statistics so that it can display overview data.

**Acceptance Criteria:**
- GET /statistics → aggregate stats
- Returns: total_projects, projects_with_analysis, technology_counts, recent_activity
- Cached for performance

---

---

## Story 6.6: Relationships API

**Status:** todo
**Size:** S
**Blocked by:** 6.2, Epic 13

**User Story:**
As the dashboard, I want to retrieve project relationships so that users can see connections between projects.

**Acceptance Criteria:**
- GET /relationships/{projectId} → array of relationships for the project
- Each relationship includes: source_project, target_project, type, description, confidence
- Returns 404 if project does not exist
- Returns empty array if no relationships exist

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 6.1 Health Endpoint | todo | XS | 5.1 |
| 6.2 Projects API | todo | M | 6.1, 5.1 |
| 6.3 Search API | todo | M | 6.2 |
| 6.4 Configuration API | todo | S | 6.1 |
| 6.5 Statistics API | todo | S | 6.2 |
| 6.6 Relationships API | todo | S | 6.2, Epic 13 |

**Total Size:** 1XS + 2M + 3S = ~8 days

**Can Start:** Story 6.1 (after 5.1 complete)
