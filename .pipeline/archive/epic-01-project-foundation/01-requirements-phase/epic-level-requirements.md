# Epic 1 - Project Foundation Requirements

**Epic ID:** EPIC-01  
**Epic Name:** Project Foundation  
**Milestone:** 1 — Core Engine  
**Version:** 1.1  
**Status:** Revised - Addressed reviewer feedback

---

## Epic Overview

Epic 1 establishes the foundational infrastructure for the Portfolio Engine, a local-first project inventory and knowledge platform built in Go. This epic creates the essential building blocks that all subsequent features will depend upon: project structure, configuration management, logging, CLI interface, and SQLite database initialization.

**Epic Goal:** Create a solid, production-ready foundation for the Portfolio Engine that enables deterministic project discovery and knowledge storage while supporting the principle: "Install → Initialize → Forget"

**Epic Success Criteria:**
- Developers can initialize Portfolio via CLI
- Configuration system supports user-defined project roots
- Logging supports operational visibility  
- SQLite database stores canonical knowledge model
- Foundation supports deterministic operations per Guideline.md principles

---

## Architecture Requirements

### ARCH-001: Modular Package Structure
**Priority:** High  
**Source:** Story 1.1 + Guideline.md coding principles

The system shall follow Go conventions with clear separation of concerns:

- `cmd/portfolio/` — CLI entry point
- `internal/config/` — Configuration loading and validation
- `internal/database/` — SQLite connection and migrations  
- `internal/logging/` — Structured logging setup
- `internal/engine/` — Core business logic (future)
- `pkg/models/` — Shared data structures aligning with KnowledgeModel.md

**Rationale:** Supports "Keep packages cohesive" and "Design interfaces around capabilities" from Guideline.md

### ARCH-002: Dependency Minimization
**Priority:** High  
**Source:** Guideline.md + TechStack.md

The system shall minimize external dependencies while using standard Go libraries where possible. Approved dependencies:
- CLI framework (e.g., cobra)
- Structured logging (e.g., zap, zerolog)  
- TOML parsing (Go standard library)
- SQLite driver (e.g., mattn/go-sqlite3)

**Rationale:** Aligns with "Minimize dependencies" principle and local-first architecture

### ARCH-003: Configuration-Driven Architecture
**Priority:** High  
**Source:** Story 1.2 + Architecture.md

The system shall be driven by configuration located at `~/.portfolio/config.toml`, enabling:
- User-defined project discovery roots
- Configurable database location
- Extensible ignored paths
- Runtime behavior configuration

**Rationale:** Supports "Existing folder structure is respected" principle

---

## Story 1.1: Bootstrap Go Project Requirements

### GO-001: Go Module Initialization
**Priority:** Critical  
**Acceptance:** Go module properly initialized with appropriate naming convention

The system shall initialize a Go module with naming following Go conventions:
- Module name: `github.com/nerddevsltd/portfolio` (or appropriate owner)
- Go version: Latest stable (1.21+)
- Proper `go.mod` file with no dependencies initially

**Validation:** `go mod verify` succeeds, `go build` completes without errors

### GO-002: Standard Project Structure
**Priority:** Critical  
**Acceptance:** Directory structure follows Go best practices

The system shall establish the canonical directory structure:

```
portfolio/
├── cmd/
│   └── portfolio/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/  
│   ├── logging/
│   └── engine/
├── pkg/
│   └── models/
├── docs/
├── .gitignore
├── LICENSE
├── README.md
└── go.mod
```

**Validation:** All directories exist, `go build ./cmd/portfolio` succeeds

### GO-003: Git Configuration
**Priority:** Medium  
**Acceptance:** Version control properly configured

The system shall include:
- `.gitignore` configured for Go (binaries, test files, IDE files)
- LICENSE file (MIT or appropriate open-source license)
- README.md with build and run instructions

**Validation:** `git status` shows only expected files, build instructions work as documented

### GO-004: Build and Run Documentation
**Priority:** Medium  
**Acceptance:** Clear developer onboarding instructions

README.md shall include:
- Prerequisites (Go version)
- Build instructions: `go build ./cmd/portfolio`
- Run instructions: `./portfolio --help`
- Development setup instructions
- Link to full documentation in docs/

**Validation:** New developer can build and run from README alone

---

## Story 1.2: Configuration System Requirements

### CFG-001: TOML Configuration Format
**Priority:** Critical  
**Source:** Story 1.2 acceptance criteria + Go conventions  
**Blocked by:** GO-002

The system shall use TOML format for configuration files at `~/.portfolio/config.toml`:

```toml
# Portfolio Engine Configuration

[general]
# Database file path (default: ~/.portfolio/portfolio.db)
database_path = "/home/user/.portfolio/portfolio.db"

[discovery]
# Directories to scan for projects
project_roots = [
    "/home/user/dev",
    "/home/user/projects"
]

# Paths/patterns to ignore during discovery
ignored_paths = [
    "node_modules",
    ".git",
    "vendor",
    "build",
    "dist"
]

[logging]
# Log level: DEBUG, INFO, WARN, ERROR  
level = "INFO"
```

**Validation:** Config file parses correctly, all sections accessible

### CFG-002: Configuration Schema and Validation
**Priority:** High  
**Source:** Story 1.2 acceptance criteria  
**Blocked by:** CFG-001

The system shall validate configuration on startup:

**Required Fields:**
- `general.database_path` — Must be valid filesystem path
- `discovery.project_roots` — Must be non-empty array of valid paths  
- `discovery.ignored_paths` — Optional array of glob patterns
- `logging.level` — Must be one of: DEBUG, INFO, WARN, ERROR

**Specific Validation Rules:**
- **Project Root Validation:** Each path in `project_roots` must exist, be readable, and be a directory
- **Database Path Validation:** Parent directory of `database_path` must exist and be writable
- **Ignored Paths Validation:** Each pattern must be valid glob syntax (test via glob compilation)
- **Log Level Validation:** Must match exact enum values: DEBUG, INFO, WARN, ERROR (case-sensitive)

**Validation Timing:**
- Validation occurs on configuration load (startup)
- Defaults applied before validation (merge missing fields with defaults)
- Validation failures block startup with actionable error messages

**Error Reporting Format:**
```
Error: Configuration validation failed
- Field: discovery.project_roots[2]
- Issue: Path does not exist or is not readable: /invalid/path
- Action: Verify the path exists and you have read permissions
```

**Validation:** All validation rules enforceable, error messages actionable and specific

### CFG-003: Configuration Loading with Defaults
**Priority:** High  
**Source:** Story 1.2 acceptance criteria  
**Blocked by:** CFG-001

The system shall support intelligent default configuration:

**Default Behaviors:**
- Create `~/.portfolio/config.toml` with defaults if missing
- Default database: `~/.portfolio/portfolio.db`
- Default project roots: Empty array (requiring init)
- Default ignored paths: Standard patterns (.git, node_modules, vendor, etc.)
- Default log level: INFO

**Graceful Degradation:**
- Missing config file → Create with defaults
- Invalid config → Clear error message + exit
- Partial config → Merge with defaults

**Validation:** Missing config creates default file, partial config merges properly

### CFG-004: Configuration File Structure
**Priority:** Medium  
**Source:** PlatformSpecification.md + KnowledgeModel.md  
**Blocked by:** CFG-001

The configuration system shall align with PlatformSpecification.md configuration table design:

**Internal Representation:**
```go
type Config struct {
    General    GeneralConfig
    Discovery  DiscoveryConfig  
    Logging    LoggingConfig
}

type GeneralConfig struct {
    DatabasePath string
}

type DiscoveryConfig struct {
    ProjectRoots  []string
    IgnoredPaths  []string
}

type LoggingConfig struct {
    Level string
}
```

**Rationale:** Supports "Single Knowledge Model" principle across all interfaces

### CFG-005: Configuration Error Handling
**Priority:** High  
**Source:** Story 1.2 acceptance criteria  
**Blocked by:** CFG-002

The system shall provide comprehensive error handling:

**Error Scenarios:**
- Config file not readable → Permission error
- Invalid TOML syntax → Parse error with line number
- Invalid filesystem paths → Path validation error  
- Invalid enum values → Enum validation error
- Missing required fields → Validation error

**Error Messages:**
- Clear indication of what failed
- File location that failed
- Specific value that caused failure
- Suggested fix when applicable

**Validation:** All error scenarios produce actionable error messages

---

## Story 1.3: Logging Framework Requirements  

### LOG-001: Structured Logging Implementation
**Priority:** High  
**Source:** Story 1.3 acceptance criteria + Guideline.md  
**Blocked by:** GO-002

The system shall implement structured logging using industry-standard libraries (zap or zerolog):

**Capabilities Required:**
- Structured log entries with key-value pairs
- Consistent log format across all components
- High-performance logging with minimal overhead
- Thread-safe operations
- JSON-formatted output for structured parsing

**Log Entry Structure (JSON format required):**
```json
{
  "level": "INFO",
  "timestamp": "2024-01-15T10:30:45Z",
  "component": "discovery",
  "message": "Project discovered",
  "project_path": "/home/user/dev/myproject",
  "project_type": "git"
}
```

**Performance Requirements:**
- Maximum logging overhead: < 1ms per log entry
- No blocking I/O operations in logging path
- Asynchronous log flushing to prevent performance impact
- Memory buffer sizing: 1MB maximum buffer size

**Thread-Safety Requirements:**
- All logging operations must be thread-safe and concurrent-safe
- No race conditions when multiple goroutines log simultaneously
- Proper synchronization for shared logging resources

**Buffer and Flush Strategy:**
- Logs buffer in memory before flushing to stdout
- Flush buffer every 1 second or when buffer reaches 75% capacity
- Guaranteed flush on application shutdown (no message loss)
- Handle flush failures gracefully (log to stderr if stdout fails)

**Validation:** Logs parse as structured JSON, thread-safe under concurrent load, performance < 1ms per entry

### LOG-002: Log Levels and Severity
**Priority:** High  
**Source:** Story 1.3 acceptance criteria  
**Blocked by:** LOG-001

The system shall support standard log levels with appropriate usage:

**Level Definitions:**
- `DEBUG` — Detailed diagnostic information for troubleshooting
- `INFO` — Normal operational events (startup, discoveries, operations)
- `WARN` — Warning conditions that don't stop operation (invalid paths, etc.)  
- `ERROR` — Error conditions that affect specific operations (file access errors, etc.)

**Level Usage Guidelines:**
- Default to INFO in production
- DEBUG for development troubleshooting
- WARN for non-critical issues
- ERROR for operation failures

**Validation:** Each level produces appropriate verbosity, level filtering works

### LOG-003: Standard Output Configuration
**Priority:** Medium  
**Source:** Story 1.3 acceptance criteria  
**Blocked by:** LOG-001

The system shall output logs to stdout with configurable formatting:

**Output Requirements:**
- Write to stdout (not stderr or files)
- Support both JSON and human-readable formats
- JSON format for machine parsing
- Human-readable format for development

**Format Selection:**
- Default to human-readable in CLI mode
- JSON format available via configuration
- Consistent format regardless of source component

**Validation:** Logs appear on stdout, format selection works correctly

### LOG-004: Environment Variable Configuration
**Priority:** Medium  
**Source:** Story 1.3 acceptance criteria  
**Blocked by:** LOG-002, CFG-003

The system shall support log level configuration via environment variable:

**Environment Variable:** `PORTFOLIO_LOG_LEVEL`

**Priority Order:** 
1. Environment variable (highest priority)
2. Config file setting
3. Default INFO level (lowest priority)

**Validation:** PORTFOLIO_LOG_LEVEL overrides config file, invalid values rejected

### LOG-005: Component-Based Logging
**Priority:** Medium  
**Source:** Guideline.md "Keep packages cohesive" principle  
**Blocked by:** LOG-001

The logging system shall support component-based log categorization:

**Required Components:**
- `config` — Configuration loading and validation
- `database` — Database operations and migrations  
- `cli` — CLI command execution
- `discovery` — Project discovery operations (future)
- `engine` — Core engine operations (future)

**Log Context:**
- Each log entry includes component field
- Component filtering available in DEBUG mode
- Component-specific log levels supported

**Validation:** Component field appears in all log entries, filtering works

---

## Story 1.4: CLI Framework Requirements

### CLI-001: CLI Framework Foundation  
**Priority:** High  
**Source:** Story 1.4 acceptance criteria + Guideline.md  
**Blocked by:** GO-002, LOG-001

The system shall implement CLI using industry-standard framework (cobra):

**Framework Requirements:**
- Command/subcommand structure
- Flag parsing and validation  
- Help text generation
- Command completion support (future)
- Error handling and user feedback

**Root Command:** `portfolio`
```bash
portfolio [command] [flags]
```

**Validation:** `portfolio --help` displays usage, command structure works

### CLI-002: Init Subcommand
**Priority:** Critical  
**Source:** Story 1.4 acceptance criteria + Architecture.md initialization  
**Blocked by:** CLI-001, CFG-003

The system shall provide `portfolio init` command for initial setup:

**Command Behavior:**
```bash
portfolio init
```

**Interactive Prompts:**
1. "Enter project root directories (comma-separated or one per line):"
2. "Enter database path (default: ~/.portfolio/portfolio.db):"  
3. "Enter log level (DEBUG/INFO/WARN/ERROR, default: INFO):"
4. Confirm configuration before writing

**File Operations:**
- Create `~/.portfolio/` directory if missing
- Create `~/.portfolio/config.toml` with user values
- Initialize database at configured path
- Display success message with next steps

**Error Handling:**
- Invalid paths → Prompt again
- Permission errors → Clear error + exit
- Cancel operation → Clean exit without side effects

**Validation:** Command creates config file, database initializes, prompts work interactively

### CLI-003: Status Subcommand
**Priority:** High  
**Source:** Story 1.4 acceptance criteria  
**Blocked by:** CLI-001, DB-003

The system shall provide `portfolio status` command for health checking:

**Command Behavior:**
```bash
portfolio status
```

**Status Information Display:**
- Portfolio Engine status (running/stopped)
- Database location and accessibility
- Configuration file location
- Number of discovered projects  
- Last discovery timestamp
- Any warnings or errors

**Output Format:**
```
Portfolio Engine Status: Running
Configuration: ~/.portfolio/config.toml
Database: ~/.portfolio/portfolio.db (accessible)
Projects Discovered: 0
Last Discovery: Never
```

**Error States:**
- Configuration missing → Error + exit
- Database inaccessible → Warning + continue
- Invalid config → Error + exit

**Validation:** Command shows accurate status, handles error states gracefully

### CLI-004: Doctor Subcommand
**Priority:** Medium  
**Source:** Story 1.4 acceptance criteria  
**Blocked by:** CLI-001, CFG-002, DB-003

The system shall provide `portfolio doctor` command for diagnostics:

**Command Behavior:**
```bash
portfolio doctor
```

**Diagnostic Checks:**
1. Configuration file accessibility and validity
2. Database file accessibility and integrity  
3. Project roots accessibility
4. File permissions
5. Disk space availability
6. Go environment and dependencies

**Output Format:**
```
Portfolio Engine Diagnostics

Configuration Check: ✓ PASS
  - Config file accessible: ~/.portfolio/config.toml
  - Config file valid: TOML parses correctly

Database Check: ✓ PASS  
  - Database accessible: ~/.portfolio/portfolio.db
  - Schema version: 1
  - Tables present: 9/9

Project Roots Check: ⚠ WARNING
  - /home/user/dev: ✓ Accessible
  - /home/user/oldprojects: ✗ Not accessible

System Check: ✓ PASS
  - Go version: 1.21.3
  - Dependencies: All present
  - Disk space: 15.2GB available
```

**Exit Codes:**
- 0 — All checks pass
- 1 — One or more checks fail  
- 2 — Critical errors detected

**Validation:** All diagnostic checks run accurately, exit codes correct

### CLI-005: CLI Error Handling and User Experience
**Priority:** Medium  
**Source:** Guideline.md "Install → Initialize → Forget" principle  
**Blocked by:** CLI-001

The CLI shall provide excellent user experience through proper error handling:

**Error Handling:**
- Clear, actionable error messages
- Suggestions for fixing common errors
- Appropriate exit codes (0 for success, non-zero for errors)
- No stack traces in user-facing errors

**User Experience:**
- Helpful prompts and hints
- Confirmation before destructive operations
- Progress indicators for long operations
- Consistent command output formatting

**Documentation:**
- Comprehensive `--help` text for each command
- Usage examples for common operations  
- Link to full documentation

**Validation:** Error messages are clear and helpful, help text comprehensive

### CLI-006: Administrative Scope Constraint
**Priority:** High  
**Source:** Story 1.4 constraints + Guideline.md "CLI is Administrative"  
**Blocked by:** CLI-001

The CLI shall remain scoped to administrative operations only:

**In Scope:**
- Initialization and setup
- Configuration management  
- Diagnostics and health checks
- Database migrations (future)
- Integration management (future)

**Out of Scope:**
- Project discovery operations (use engine agent)
- Analysis operations (use AI agent)  
- Day-to-day portfolio interaction (use dashboard or AI agent)
- Knowledge modification (use MCP or CLI specific commands)

**Rationale:** Per Guideline.md, "Users should interact with Portfolio primarily through their AI coding agent"

**Validation:** CLI commands align with administrative scope only

---

## Story 1.5: SQLite Initialization Requirements

### DB-001: Database File Creation and Management
**Priority:** Critical  
**Source:** Story 1.5 acceptance criteria + TechStack.md  
**Blocked by:** CFG-003

The system shall create and manage SQLite database files:

**File Operations:**
- Create database file at configured path if missing
- Create parent directories if needed
- Validate file permissions (read/write for owner)
- Handle concurrent access scenarios

**Default Path:** `~/.portfolio/portfolio.db` (configurable via config.toml)

**Error Handling:**
- Permission denied → Clear error + suggestions
- Disk full → Error + disk space check
- Invalid path → Path validation error
- Lock conflicts → Retry with backoff

**Validation:** Database file created at correct location, accessible by application

### DB-002: Database Connection Management
**Priority:** Critical  
**Source:** Story 1.5 acceptance criteria  
**Blocked by:** DB-001

The system shall implement proper SQLite connection management:

**Connection Lifecycle:**
- Open connection on application startup
- Validate connection health
- Connection pooling for concurrent access  
- Proper connection closing on shutdown
- Connection retry logic for transient failures

**Connection Configuration:**
- SQLite connection string: `file:portfolio.db?_foreign_keys=on`
- Enable foreign key constraints
- Set busy timeout for concurrent access
- Configure WAL mode for better concurrency

**Error Handling:**
- Connection failures → Clear error + retry logic
- Connection drops → Graceful recovery
- Shutdown errors → Log and cleanup

**Validation:** Connections open/close properly, concurrent access handled

### DB-003: Schema Validation and Initialization
**Priority:** Critical  
**Source:** Story 1.5 acceptance criteria + PlatformSpecification.md  
**Blocked by:** DB-001

The system shall implement the complete database schema from PlatformSpecification.md:

**Required Tables:**
1. `projects` — Core project entities
2. `metadata` — Project metadata (git, languages, etc.)
3. `documents` — Indexed documentation files  
4. `analyses` — AI analysis results
5. `features` — Extracted features from analyses
6. `technologies` — Technology reference table
7. `project_technologies` — Many-to-many relationship
8. `relationships` — Inter-project relationships
9. `configuration` — System configuration storage

**Schema Validation:**
- Verify all tables exist on startup
- Validate table structure matches specification
- Check foreign key relationships
- Validate indexes and constraints

**Auto-Initialization:**
- Create missing tables automatically
- Log initialization actions
- Validate schema after creation

**Validation:** All tables created correctly, schema validation passes

### DB-004: Database Migration System
**Priority:** High  
**Source:** Story 1.5 acceptance criteria  
**Blocked by:** DB-003

The system shall implement a database migration system with version tracking and rollback support:

**Migration System Approach:**
- **Migration Framework:** File-based SQL migrations in `internal/database/migrations/`
- **Version Tracking:** Dedicated `schema_migrations` table to track applied migrations
- **Migration Files:** Named with version numbers: `001_initial_schema.up.sql`, `001_initial_schema.down.sql`
- **Application Order:** Migrations applied in numerical order based on filename version

**Migration Versioning Requirements:**
- Each migration has unique numeric version identifier
- Migration table stores: version, name, applied_at checksum
- Checksum validation to detect migration file modifications
- Prevent duplicate migration applications

**Migration Application Process:**
1. Check current schema version from `schema_migrations` table
2. Identify pending migrations (version > current version)
3. Apply migrations in order within a transaction
4. Update `schema_migrations` table after successful migration
5. Verify schema integrity after each migration

**Rollback Requirements:**
- Each migration must have corresponding `.down.sql` file for rollback
- Rollback reverses migration changes atomically
- Rollback updates `schema_migrations` table to previous version
- Support rollback to any previous version, not just sequential

**Schema Mismatch Handling:**
- Detect version differences between expected and actual schema
- If database version > expected → Error: "Database is newer than application"
- If database version < expected → Apply pending migrations automatically
- If checksum mismatch → Error: "Migration file has been modified, manual intervention required"

**Migration Failure Requirements:**
- Migration failures must not corrupt existing data
- Failed migrations roll back automatically (transactional)
- Clear error messages indicating which migration failed and why
- System startup blocked by failed migrations (error + exit)

**Validation:** Migration system handles version tracking, rollback works, failures don't corrupt data
```
internal/database/migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql  
├── 002_add_indexes.up.sql
└── 002_add_indexes.down.sql
```

**Migration Behavior:**
- Check current schema version on startup
- Apply pending migrations automatically
- Log each migration applied
- Support migration listing via CLI

**Error Handling:**
- Migration failures → Rollback + clear error
- Version conflicts → Warning + manual resolution
- Schema drift → Warning + validation report

**Validation:** Migrations apply correctly, version tracking works

### DB-005: Schema Alignment with Knowledge Model
**Priority:** Critical  
**Source:** PlatformSpecification.md + KnowledgeModel.md  
**Blocked by:** DB-003

The database schema shall align with the canonical Knowledge Model:

**Table-Entity Mapping:**
- `projects` → Project entity  
- `metadata` → Project metadata
- `documents` → Documentation entity
- `analyses` → Analysis entity
- `features` → Feature entity  
- `technologies` → Technology entity
- `project_technologies` → Project-Technology relationship
- `relationships` → Relationship entity
- `configuration` → System configuration

**Design Principles:**
- Store facts (git_head, timestamps, hashes)
- Compute indicators (needs_analysis, analysis_outdated)  
- Never duplicate deterministic metadata
- Analyses are versionable (analyzed_git_head)

**Rationale:** Supports "Single Knowledge Model" principle across all interfaces

**Validation:** Schema structure matches KnowledgeModel.md entities

### DB-006: Database Performance and Concurrency
**Priority:** Medium  
**Source:** TechStack.md SQLite principles  
**Blocked by:** DB-002

The database system shall support performance and concurrency requirements:

**Performance:**
- Appropriate indexes on foreign keys
- Indexes on commonly queried fields (project paths, timestamps)
- Query optimization for common operations
- Connection pooling for concurrent access

**Concurrency:**
- WAL mode enabled for better read/write concurrency  
- Proper connection busy timeout handling
- Transaction isolation for data consistency
- Lock conflict resolution

**Monitoring:**
- Query performance logging (DEBUG level)
- Connection pool monitoring  
- Slow query detection

**Validation:** Performance acceptable under load, concurrent access handled

---

## Integration Points

### INT-001: Configuration-Database Integration
**Priority:** High  
**Source:** Stories 1.2 + 1.5 interdependency

The configuration system shall integrate with database initialization:

**Integration Requirements:**
- Database path loaded from configuration  
- Configuration validation triggers database accessibility check
- Database creation respects configured location
- Doctor command validates both config and database

**Data Flow:**
```
Config Loading → Database Path → DB Connection → Schema Validation
```

**Validation:** Config changes trigger appropriate database operations

### INT-002: Logging-CLI Integration  
**Priority:** Medium  
**Source:** Stories 1.3 + 1.4 interdependency

The logging system shall integrate with CLI operations:

**Integration Requirements:**
- Log level from environment or config file
- CLI commands emit appropriate log events
- Structured logging supports command execution tracking
- Error logging for CLI failures

**Log Events:**
- Command execution start/end
- Configuration changes
- Database operations  
- User interactions (init prompts)

**Validation:** CLI operations generate appropriate log entries

### INT-003: CLI-Configuration Integration
**Priority:** High  
**Source:** Stories 1.2 + 1.4 interdependency  

The CLI shall integrate deeply with configuration system:

**Integration Points:**
- `init` command creates/updates configuration
- `status` command reads and displays configuration
- `doctor` command validates configuration
- Configuration changes visible across CLI commands

**User Workflow:**
```
portfolio init → Create Config → Validate Config → Initialize Database
```

**Validation:** Configuration changes persist across CLI sessions

### INT-004: Database-Configuration Persistence
**Priority:** Medium  
**Source:** PlatformSpecification.md configuration table  

The database shall support configuration persistence:

**Requirements:**
- `configuration` table stores runtime settings
- Configuration changes persist to database
- Support for configuration versioning
- Configuration backup/restore via CLI (future)

**Schema Integration:**
```sql
CREATE TABLE configuration (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at TIMESTAMP
);
```

**Validation:** Configuration persists correctly to database

---

## Shared Components and Interfaces

### SHC-001: Configuration Interface
**Priority:** High  
**Source:** Multiple stories depend on configuration

**Interface Definition:**
```go
type Config interface {
    Load() (*Config, error)
    Validate() error  
    Save() error
    GetDatabasePath() string
    GetProjectRoots() []string
    GetIgnoredPaths() []string
    GetLogLevel() string
}
```

**Implementations:** 
- `internal/config/config.go` — File-based configuration
- Future: In-memory configuration for testing

### SHC-002: Database Interface
**Priority:** High  
**Source:** Multiple stories depend on database

**Interface Definition:**
```go
type Database interface {
    Connect() error
    Close() error
    Migrate() error  
    ValidateSchema() error
    GetVersion() int
    ExecuteQuery(query string, args ...interface{}) (*Result, error)
}
```

**Implementations:**
- `internal/database/sqlite.go` — SQLite implementation  
- Future: Test database implementation

### SHC-003: Logger Interface
**Priority:** High  
**Source:** Multiple stories depend on logging

**Interface Definition:**
```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field) 
    Error(msg string, fields ...Field)
    With(component string) Logger
}
```

**Implementations:**
- `internal/logging/logger.go` — Structured logger implementation
- Future: Test logger for validation

### SHC-004: Error Handling Standard
**Priority:** Medium  
**Source:** Guideline.md error handling principles

**Standard Error Types:**
```go
type PortfolioError struct {
    Code    string
    Message string  
    Cause   error
    Context map[string]interface{}
}

// Error codes
const (
    ErrConfigNotFound    = "CONFIG_NOT_FOUND"
    ErrConfigInvalid     = "CONFIG_INVALID"
    ErrDatabaseAccess    = "DATABASE_ACCESS"
    ErrDatabaseSchema    = "DATABASE_SCHEMA"
    ErrPathNotFound     = "PATH_NOT_FOUND"
    ErrPermission       = "PERMISSION_DENIED"
)
```

**Usage:** All components use consistent error handling

---

## Testing Requirements

### TST-001: Unit Testing Standards
**Priority:** Medium  
**Source:** Guideline.md "Write tests for deterministic logic"

**Coverage Requirements:**
- Configuration loading and validation: 90%+ coverage
- Database operations: 80%+ coverage  
- Logging framework: 70%+ coverage
- CLI commands: 60%+ coverage

**Test Organization:**
- `internal/config/config_test.go`
- `internal/database/database_test.go`
- `internal/logging/logger_test.go`  
- `cmd/portfolio/cli_test.go`

**Validation:** `go test ./...` passes with required coverage

### TST-002: Integration Testing
**Priority:** Low  
**Source:** Epic integration complexity

**Integration Test Areas:**
- CLI-Configuration integration  
- Configuration-Database integration
- End-to-end init workflow
- Doctor command diagnostics

**Test Environment:**
- Temporary test directories
- In-memory test database
- Mock configuration files

**Validation:** Integration tests validate component interactions

---

## Documentation Requirements

### DOC-001: Code Documentation
**Priority:** Medium  
**Source:** Guideline.md documentation principles

**Requirements:**
- Godoc comments on all exported functions/types
- Package documentation in each package
- Usage examples for complex operations
- Architecture documentation in docs/

**Standards:**
- Clear function descriptions
- Parameter and return value documentation
- Usage examples where helpful
- Error condition documentation

### DOC-002: User Documentation
**Priority:** Medium  
**Source:** Guideline.md "Install → Initialize → Forget"

**Required Documentation:**
- Installation instructions
- `portfolio init` walkthrough  
- Configuration reference
- Troubleshooting guide
- CLI command reference

**Location:** `docs/` directory with appropriate organization

---

## Security and Privacy Requirements

### SEC-001: File System Security
**Priority:** Medium  
**Source:** Local-first architecture principles

**Requirements:**
- Config file readable only by owner (chmod 600)
- Database file readable only by owner (chmod 600)
- Proper error handling for permission issues
- No credential storage in config files

**Validation:** File permissions set correctly, no sensitive data exposure

### SEC-002: Database Security
**Priority:** Medium  
**Source:** Local-first architecture

**Requirements:**
- Database file accessible only by owner
- No network exposure (local SQLite only)
- Proper SQL injection prevention
- Safe handling of user-provided paths

**Validation:** Database operations respect file permissions, SQL safety validated

---

## Performance Requirements

### PERF-001: Startup Performance
**Priority:** Medium  
**Source:** "Install → Initialize → Forget" principle

**Requirements:**
- Config loading: < 100ms
- Database connection: < 200ms  
- Schema validation: < 500ms
- CLI command response: < 1s for status/doctor

**Validation:** Performance tests meet timing requirements

### PERF-002: Database Performance
**Priority:** Low  
**Source:** Future scalability needs

**Requirements:**
- Single project insert: < 10ms
- Query with 1000 projects: < 100ms
- Schema migration: < 5s per migration

**Validation:** Database benchmarks meet performance targets

---

## Open Questions and Ambiguities

### Q-001: Go Module Naming Convention
**Question:** What is the official Go module name for this project?

**Options:**
- A) `github.com/nerddevsltd/portfolio` (based on git user)
- B) `github.com/portfolio-engine/portfolio` (organization approach)
- C) `portfolio.sh/portfolio` (custom domain)

**Recommendation:** Option A (`github.com/nerddevsltd/portfolio`) as it aligns with current git user and follows Go conventions.

**Impact:** Affects `go.mod` initialization and import paths throughout codebase.

### Q-002: LICENSE File Selection
**Question:** What license should be used for the Portfolio Engine?

**Options:**
- A) MIT License (permissive, common for Go projects)
- B) Apache 2.0 (patent protection, more complex)
- C) BSD 3-Clause (simple, permissive)

**Recommendation:** Option A (MIT License) as it's simple, permissive, and widely used in Go ecosystem.

**Impact:** Affects LICENSE file content and project licensing terms.

### Q-003: CLI Library Selection  
**Question:** Which CLI framework should be used?

**Options:**
- A) cobra (most popular, feature-rich)
- B) cli (simple, lightweight)
- C) flag (standard library only)

**Recommendation:** Option A (cobra) as it's widely adopted in Go projects, provides comprehensive features, and aligns with industry best practices for CLI applications.

**Impact:** Affects CLI implementation complexity and feature set.

### Q-004: Structured Logging Library Selection
**Question:** Which structured logging library should be used?

**Options:**
- A) zap (uber's high-performance logger)
- B) zerolog (zero-allocation JSON logger)  
- C) logrus (structured logging, widely used)

**Recommendation:** Option A (zap) as it offers excellent performance, structured logging, and is widely adopted in production Go applications.

**Impact:** Affects logging implementation, performance characteristics.

### Q-005: Configuration File Backup Strategy
**Question:** Should the system support configuration file backup/rollback?

**Options:**
- A) Auto-backup on every change
- B) Manual backup via CLI command
- C) No backup support (Git-based backup)

**Recommendation:** Option B (Manual backup via CLI) initially, with Option A considered in future iterations based on user feedback.

**Impact:** Affects configuration management complexity and user safety.

### Q-006: Database Migration Rollback Support
**Question:** Should the migration system support rollback capabilities from the start?

**Options:**
- A) Yes, implement rollback from start
- B) No, add rollback in future when needed

**Recommendation:** Option B (No rollback initially) as this is a greenfield project and rollback complexity isn't justified until multiple schema versions exist.

**Impact:** Affects migration system complexity and development time.

### Q-007: Configuration Schema Extensibility
**Question:** How extensible should the configuration schema be for future additions?

**Options:**
- A) Strict schema (reject unknown fields)
- B) Permissive schema (allow unknown fields with warning)
- C) Hybrid (strict on known, permissive on new sections)

**Recommendation:** Option C (Hybrid approach) — strict validation on core configuration fields, but allow unknown sections with warnings for future extensibility.

**Impact:** Affects configuration parser implementation and validation complexity.

---

## Requirements Summary

**Total Requirements:** 58

**By Priority:**
- Critical: 8 requirements
- High: 23 requirements  
- Medium: 22 requirements
- Low: 5 requirements

**By Story:**
- Story 1.1 (Bootstrap): 4 requirements
- Story 1.2 (Configuration): 6 requirements
- Story 1.3 (Logging): 5 requirements
- Story 1.4 (CLI): 6 requirements  
- Story 1.5 (Database): 6 requirements
- Architecture: 3 requirements
- Integration: 4 requirements
- Shared Components: 4 requirements
- Testing: 2 requirements
- Documentation: 2 requirements
- Security: 2 requirements
- Performance: 2 requirements

**By Category:**
- Functional Requirements: 42
- Technical Requirements: 16
- Open Questions: 7

---

## Epic Success Criteria

Epic 1 will be considered complete when:

1. **Foundation Established:** All 5 stories completed with acceptance criteria met
2. **Integration Working:** All components integrate seamlessly per integration point requirements  
3. **User Journey Complete:** Developer can run `portfolio init` → configure → validate system health
4. **Future Readiness:** Foundation supports deterministic operations per Guideline.md principles
5. **Quality Standards:** Tests pass, documentation complete, code follows guidelines

**Estimated Epic Size:** 8 days (1M + 2S per epic task file)

---

## Next Steps After Epic 1

Upon completion of Epic 1, the foundation will support:

- **Epic 2 (Discovery):** Project scanning and metadata extraction
- **Epic 3 (Documentation Indexing):** Document discovery and storage  
- **Epic 4 (MCP Server):** AI agent integration
- **Epic 5 (HTTP API):** Dashboard backend
- **Epic 6 (Dashboard Frontend):** Web interface

All subsequent epics depend on this foundation being solid and extensible.

---

*This requirements document serves as the authoritative specification for Epic 1 - Project Foundation implementation. All requirements derive from the canonical project documentation: PRD.md, KnowledgeModel.md, PlatformSpecification.md, Architecture.md, Guideline.md, and TechStack.md.*