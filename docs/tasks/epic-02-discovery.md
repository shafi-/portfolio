# Epic 2 — Discovery

**Milestone:** 1 — Core Engine
**Status:** todo

## Overview

Implement project discovery across configured root directories, supporting nested repositories, common project type detection, and ignoring generated directories.

---

## Story 2.1: Configure Project Roots

**Status:** todo
**Size:** S
**Blocked by:** 1.2, 1.4

**User Story:**
As a user, I want to specify which directories contain projects so that Portfolio can discover them.

**Acceptance Criteria:**
- `portfolio init` prompts for project root directories
- Configuration persists to config file
- Support for multiple root directories
- Validation that paths exist and are accessible

---

## Story 2.2: Recursive Project Discovery

**Status:** todo
**Size:** L
**Blocked by:** 2.1, 1.5

**User Story:**
As the Portfolio Engine, I want to recursively scan project roots so that I can discover all repositories.

**Acceptance Criteria:**
- Walks directory trees from configured roots
- Detects Git repositories by `.git` directory presence
- Creates Project records with: id, name, root_path, repository_type, discovered_at
- Handles permission errors gracefully
- Reports discovery count and errors

**Technical Context:**
- Project entity per KnowledgeModel.md
- Store only facts (filesystem paths), not derived state

---

## Story 2.3: Support Nested Folders

**Status:** todo
**Size:** M
**Blocked by:** 2.2

**User Story:**
As the Portfolio Engine, I want to handle nested repositories so that I don't miss projects in subdirectories.

**Acceptance Criteria:**
- Continues recursion when subdirectory contains repository
- Creates separate Project record for each discovered repo
- Handles monorepo structures with nested services
- No depth limit (within filesystem constraints)

---

## Story 2.4: Detect Common Project Types

**Status:** todo
**Size:** M
**Blocked by:** 2.2

**User Story:**
As the Portfolio Engine, I want to identify project types so that I can apply appropriate metadata extraction.

**Acceptance Criteria:**
- Detects presence of: package.json (Node), go.mod (Go), requirements.txt/pyproject.toml (Python), Cargo.toml (Rust), pom.xml (Java)
- Sets repository_type field based on detected markers
- Supports multiple markers per project (polyglot)
- Gracefully handles unknown project types

---

## Story 2.5: Ignore Generated Directories

**Status:** todo
**Size:** S
**Blocked by:** 2.2

**User Story:**
As the Portfolio Engine, I want to skip generated directories so that I don't create noise projects.

**Acceptance Criteria:**
- Skips: node_modules/, vendor/, .venv/, target/, build/, dist/
- Respects .gitignore for additional ignores
- Configurable ignore patterns in config file
- Logs skipped directories at DEBUG level

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 2.1 Configure Project Roots | todo | S | 1.2, 1.4 |
| 2.2 Recursive Project Discovery | todo | L | 2.1, 1.5 |
| 2.3 Support Nested Folders | todo | M | 2.2 |
| 2.4 Detect Common Project Types | todo | M | 2.2 |
| 2.5 Ignore Generated Directories | todo | S | 2.2 |

**Total Size:** 1L + 2M + 2S = ~17 days

**Can Start:** Story 2.1 (after 1.2, 1.4 complete)
