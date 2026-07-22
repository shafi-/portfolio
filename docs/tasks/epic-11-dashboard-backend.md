# Epic 11 — Dashboard Backend

**Milestone:** 3 — Dashboard
**Status:** todo

## Overview

Implement dashboard backend including asset serving, API integration, and search endpoints to support the read-only dashboard frontend.

---

## Story 11.1: Asset Serving

**Status:** todo
**Size:** S
**Blocked by:** Epic 6

**User Story:**
As the Portfolio Engine, I want to serve static assets so that the dashboard can load its frontend.

**Acceptance Criteria:**
- Serves dashboard HTML/CSS/JS from configured path
- Proper MIME types
- Cache headers for static assets
- Supports embedded or external dashboard files

---

## Story 11.2: Dashboard API Integration

**Status:** todo
**Size:** M
**Blocked by:** Epic 6

**User Story:**
As the dashboard, I want API endpoints so that I can fetch portfolio data.

**Acceptance Criteria:**
- All HTTP API endpoints available for dashboard
- CORS configuration for local development
- Authentication not required (local only)
- Error responses with proper status codes

---

## Story 11.3: Search Endpoints

**Status:** todo
**Size:** M
**Blocked by:** 6.3

**User Story:**
As the dashboard, I want rich search so that users can find anything in their portfolio.

**Acceptance Criteria:**
- Unified search endpoint
- Search filters: by technology, by framework, by date range
- Pagination support
- Highlighted snippets in results

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 11.1 Asset Serving | todo | S | Epic 6 |
| 11.2 Dashboard API Integration | todo | M | Epic 6 |
| 11.3 Search Endpoints | todo | M | 6.3 |

**Total Size:** 1S + 2M = ~7 days

**Can Start:** Story 11.1 (after Epic 6 complete)
