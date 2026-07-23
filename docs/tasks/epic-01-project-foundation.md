# Epic 1 — Project Foundation

**Milestone:** 1 — Core Engine
**Status:** completed

## Overview

Bootstrap the Go project, establish configuration, logging, CLI framework, and SQLite database initialization.

---

## Story 1.1: Bootstrap Go Project

**Status:** completed
**Size:** S
**Blocked by:** None

**User Story:**
As a developer, I want a properly structured Go project so that I can begin implementing the Portfolio Engine.

**Acceptance Criteria:**
- Go module initialized with appropriate name
- Standard project structure: `cmd/`, `internal/`, `pkg/`
- `.gitignore` configured for Go
- LICENSE file present
- README with build and run instructions

---

## Story 1.2: Configuration System

**Status:** completed
**Size:** M
**Blocked by:** 1.1

**User Story:**
As the Portfolio Engine, I want to load and store configuration so that I can discover projects in user-defined directories.

**Acceptance Criteria:**
- Configuration file format defined (TOML per Go conventions)
- Configuration schema includes: project roots, ignored paths, database path
- Configuration loading with defaults
- Configuration validation on startup
- Error handling for missing or invalid config

**Technical Context:**
- Config location: `~/.portfolio/config.toml`
- Stores list of directories to scan for projects

---

## Story 1.3: Logging Framework

**Status:** completed
**Size:** S
**Blocked by:** 1.1

**User Story:**
As a developer, I want structured logging so that I can diagnose engine behavior.

**Acceptance Criteria:**
- Structured logging implementation (e.g., zap, zerolog)
- Log levels: DEBUG, INFO, WARN, ERROR
- Log output to stdout
- Configurable log level via environment variable

**Implementation Notes:**
- Fixed critical concurrency bug in `With` method (2026-07-23)
- Bug: Copying struct containing `sync.Once` violated Go concurrency safety
- Fix: Properly create new Logger instances instead of struct copying
- Location: `internal/logging/logger.go:145`

---

## Story 1.4: CLI Framework

**Status:** completed
**Size:** M
**Blocked by:** 1.1, 1.3

**User Story:**
As a user, I want a command-line interface so that I can initialize and administer Portfolio.

**Acceptance Criteria:**
- CLI framework (e.g., cobra)
- Subcommands: `init`, `status`, `doctor`
- `init` prompts for project roots and creates config
- `status` shows engine health and project count
- `doctor` runs diagnostics (config check, database access)

**Constraints:**
- CLI is for administration only, not primary interaction (per Guideline.md)

---

## Story 1.5: SQLite Initialization

**Status:** completed
**Size:** M
**Blocked by:** 1.2

**User Story:**
As the Portfolio Engine, I want a SQLite database so that I can store project knowledge locally.

**Acceptance Criteria:**
- Database file created at configured path
- Connection management with proper closing
- Basic schema validation on startup
- Migration system in place
- Database creation with all tables from PlatformSpecification.md

**Technical Context:**
- Tables: projects, metadata, documents, analyses, features, technologies, project_technologies, relationships, configuration

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 1.1 Bootstrap Go Project | completed | S | — |
| 1.2 Configuration System | completed | M | 1.1 |
| 1.3 Logging Framework | completed | S | 1.1 |
| 1.4 CLI Framework | completed | M | 1.1, 1.3 |
| 1.5 SQLite Initialization | completed | M | 1.2 |

**Total Size:** 1M + 2S = ~8 days

**Can Start:** Story 1.1 (no dependencies)
