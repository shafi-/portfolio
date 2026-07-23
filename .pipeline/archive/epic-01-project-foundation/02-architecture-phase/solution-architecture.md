# Epic 1 - Project Foundation Architecture

**Version:** 1.0  
**Epic ID:** EPIC-01  
**Stories Covered:** 1.1, 1.2, 1.3, 1.4, 1.5  
**Generated:** 2026-07-22

---

## Executive Summary

This architecture defines the foundational infrastructure for the Portfolio Engine, establishing 5 core components that enable deterministic project discovery and knowledge storage. The architecture follows the principle: **"Install → Initialize → Forget"** while supporting engineering principles from Guideline.md.

**Component Count:** 5 new components, 2 shared interfaces  
**New Dependencies:** 4 external Go packages  
**Implementation Complexity:** Medium (1M + 2S = ~8 days)

---

## Architecture Overview

### System Philosophy
> **The Engine knows. The AI Agent thinks.**

The foundation established in Epic 1 enables deterministic engine operations while preparing for future AI agent integration. The architecture maintains clear separation between:
- **Deterministic Operations** (Engine): Configuration, logging, database, CLI
- **Semantic Operations** (Future AI Agents): Analysis, relationship discovery, recommendations

### Component Relationship Map
```
┌─────────────────────────────────────────────────────────────┐
│                     Portfolio Engine                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐                    │
│  │    CLI       │◄─────┤  Config      │◄───── User          │
│  │  Framework   │      │   System     │                     │
│  └──────┬───────┘      └──────────────┘                     │
│         │                                                      │
│         ├──► ┌──────────────┐                                │
│         │    │   Logging    │                                │
│         │    │  Framework   │                                │
│         │    └──────────────┘                                │
│         │                                                      │
│         └──► ┌──────────────┐                                │
│              │   Database   │                                │
│              │   System     │                                │
│              └──────────────┘                                │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Architecture

### Component 1: Go Project Bootstrap (Story 1.1)

**Purpose:** Establish standard Go project structure following industry conventions

**Package Structure:**
```
portfolio/
├── cmd/
│   └── portfolio/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── config/                    # Story 1.2
│   │   ├── config.go              # Configuration loading/validation
│   │   ├── config_test.go         # Configuration tests
│   │   └── defaults.go            # Default configuration values
│   ├── database/                  # Story 1.5
│   │   ├── database.go            # SQLite connection management
│   │   ├── migrations.go          # Migration system
│   │   ├── schema.go              # Schema validation
│   │   └── database_test.go       # Database tests
│   ├── logging/                   # Story 1.3
│   │   ├── logger.go              # Structured logging implementation
│   │   ├── logger_test.go         # Logging tests
│   │   └── config.go              # Logging configuration
│   ├── cli/                       # Story 1.4
│   │   ├── root.go                # Root command setup
│   │   ├── init.go                # Init command
│   │   ├── status.go              # Status command
│   │   ├── doctor.go              # Doctor command
│   │   └── cli_test.go            # CLI tests
│   └── engine/                    # Future stories
│       └── (placeholder for future engine logic)
├── pkg/
│   └── models/                    # Shared data structures
│       ├── config.go              # Configuration models
│       ├── database.go            # Database models
│       └── errors.go              # Common error types
├── docs/                          # Existing documentation
├── .gitignore                     # Go-specific ignores
├── LICENSE                        # MIT License
├── README.md                      # Build and run instructions
└── go.mod                         # Go module definition
```

**Go Module Configuration:**
```go
module github.com/nerddevsltd/portfolio

go 1.21

require (
    github.com/spf13/cobra v1.8.0          // CLI framework
    github.com/BurntSushi/toml v1.3.2      // TOML parsing (if needed)
    go.uber.org/zap v1.26.0                // Structured logging
    github.com/mattn/go-sqlite3 v1.14.18   // SQLite driver
)
```

**Key Design Decisions:**
- **`internal/` packages**: Private application code, cannot be imported externally
- **`pkg/models/`**: Shared data structures usable by future extensions
- **Standard layout**: Follows Go project layout conventions
- **Module naming**: Uses GitHub path for maximum compatibility

---

### Component 2: Configuration System (Story 1.2)

**Purpose:** Provide flexible, validated configuration management

**Configuration Schema:**
```toml
# Portfolio Engine Configuration
[general]
database_path = "/home/user/.portfolio/portfolio.db"

[discovery]
project_roots = [
    "/home/user/dev",
    "/home/user/projects"
]
ignored_paths = [
    "node_modules",
    ".git",
    "vendor",
    "build",
    "dist"
]

[logging]
level = "INFO"
```

**Internal Data Structures:**
```go
// pkg/models/config.go
type Config struct {
    General    GeneralConfig    `toml:"general" validate:"required"`
    Discovery  DiscoveryConfig  `toml:"discovery" validate:"required"`
    Logging    LoggingConfig    `toml:"logging" validate:"required"`
}

type GeneralConfig struct {
    DatabasePath string `toml:"database_path" validate:"required,filepath"`
}

type DiscoveryConfig struct {
    ProjectRoots []string `toml:"project_roots" validate:"required,min=1"`
    IgnoredPaths []string `toml:"ignored_paths" validate:"omitempty"`
}

type LoggingConfig struct {
    Level string `toml:"level" validate:"required,oneof=DEBUG INFO WARN ERROR"`
}
```

**Configuration System Architecture:**
```
┌───────────────────────────────────────────────────────────┐
│                   Configuration System                     │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Loader    │───►│ Validator   │───►│   Saver     │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │          │
│         ▼                   ▼                   ▼          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │    TOML     │    │   Schema    │    │   File      │  │
│  │    Parser   │    │   Checks    │    │   Writer    │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

**Key Features:**
- **Smart Defaults**: Create default config if missing
- **Validation Layers**: Path validation, enum validation, required field checks
- **Error Context**: Actionable error messages with specific field locations
- **Merge Strategy**: Partial configs merge with defaults intelligently

**Integration Points:**
- **CLI Integration**: `init` command creates config, `status` reads config
- **Database Integration**: Database path from config, validation triggers DB checks
- **Logging Integration**: Log level from config (with environment override)

---

### Component 3: Logging Framework (Story 1.3)

**Purpose:** Provide high-performance structured logging for operational visibility

**Logging Architecture:**
```
┌───────────────────────────────────────────────────────────┐
│                    Logging Framework                       │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │  Log Level  │───►│   Buffer    │───►│   Output    │  │
│  │  Manager    │    │  Manager    │    │   Manager   │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │          │
│         ▼                   ▼                   ▼          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │ Environment │    │  Asynchronous│    │   stdout    │  │
│  │  Override   │    │   Flushing  │    │   (JSON)    │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

**Log Entry Structure (JSON):**
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

**Logger Interface:**
```go
// pkg/models/logger.go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    With(component string) Logger
    Sync() error
}

type Field struct {
    Key   string
    Value interface{}
}
```

**Performance Requirements:**
- **Latency**: < 1ms per log entry
- **Buffer**: 1MB maximum buffer size
- **Flushing**: Every 1 second or at 75% capacity
- **Thread Safety**: Full concurrent safety

**Component Support:**
- `config` - Configuration loading and validation
- `database` - Database operations and migrations
- `cli` - CLI command execution
- `discovery` - Project discovery (future)
- `engine` - Core engine operations (future)

**Integration Points:**
- **Environment Override**: `PORTFOLIO_LOG_LEVEL` overrides config
- **Component Integration**: All components log through central logger
- **CLI Integration**: Commands emit structured log events

---

### Component 4: CLI Framework (Story 1.4)

**Purpose:** Provide administrative interface for initialization and diagnostics

**CLI Architecture:**
```
┌───────────────────────────────────────────────────────────┐
│                      CLI Framework                          │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Root Cmd  │───►│   Init Cmd  │    │  Status Cmd │  │
│  │   (cobra)   │    │   (setup)   │    │  (health)   │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                                                      │
│         └───► ┌─────────────┐    ┌─────────────┐            │
│              │ Doctor Cmd   │    │  Future...  │            │
│              │  (diagnostics)│   │  Commands   │            │
│              └─────────────┘    └─────────────┘            │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

**Command Specifications:**

**1. `portfolio init` - Initial Setup**
```bash
portfolio init
```
- Interactive prompts for project roots, database path, log level
- Creates `~/.portfolio/` directory structure
- Generates `~/.portfolio/config.toml`
- Initializes database with schema
- Validates configuration

**2. `portfolio status` - Health Check**
```bash
portfolio status
```
- Engine status display
- Configuration location and validity
- Database accessibility and project count
- Last discovery timestamp
- Warnings and errors

**3. `portfolio doctor` - Diagnostics**
```bash
portfolio doctor
```
- Configuration file checks
- Database integrity validation
- Project roots accessibility
- File permissions verification
- Disk space availability
- Go environment validation
- Exit codes: 0 (pass), 1 (warnings), 2 (errors)

**Administrative Scope Constraint:**
- **In Scope**: init, status, doctor, config, diagnostics
- **Out of Scope**: project discovery, analysis operations, day-to-day interaction

**Integration Points:**
- **Configuration**: Uses config system, creates default config
- **Database**: Initializes database, runs migrations
- **Logging**: Logs all operations with appropriate component tags

---

### Component 5: SQLite Initialization (Story 1.5)

**Purpose:** Establish local-first knowledge storage with canonical schema

**Database Architecture:**
```
┌───────────────────────────────────────────────────────────┐
│                    Database System                          │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │ Connection  │───►│   Schema    │───►│ Migration   │  │
│  │   Manager   │    │  Validator  │    │   System    │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │          │
│         ▼                   ▼                   ▼          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │ Connection  │    │   9 Tables   │    │   Version   │  │
│  │   Pool     │    │   Validation │    │  Tracking   │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

**Database Schema (9 Tables):**
1. **projects** - Core project entities
2. **metadata** - Project metadata (git, languages, frameworks)
3. **documents** - Indexed documentation files
4. **analyses** - AI analysis results (versionable)
5. **features** - Extracted features from analyses
6. **technologies** - Technology reference table
7. **project_technologies** - Many-to-many relationship
8. **relationships** - Inter-project relationships
9. **configuration** - System configuration storage

**Schema Validation:**
- Verify all tables exist on startup
- Validate table structures match PlatformSpecification.md
- Check foreign key relationships
- Validate indexes and constraints
- Auto-create missing tables

**Migration System:**
```
internal/database/migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_indexes.up.sql
├── 002_add_indexes.down.sql
└── ...
```

**Migration Features:**
- Version tracking with `schema_migrations` table
- Checksum validation for migration files
- Transactional application with rollback support
- Automatic pending migration detection
- Schema mismatch handling

**Connection Management:**
- **Configuration**: `file:portfolio.db?_foreign_keys=on`
- **Foreign Keys**: Enabled for referential integrity
- **WAL Mode**: Enabled for better concurrency
- **Busy Timeout**: Configured for concurrent access
- **Connection Pool**: Managed for concurrent operations

**Integration Points:**
- **Configuration**: Database path from config, validation checks
- **CLI**: Doctor validates database, init creates database
- **Logging**: All database operations logged

---

## Shared Interfaces and Components

### Interface 1: Configuration Interface
```go
// pkg/models/config.go
type ConfigInterface interface {
    Load() (*Config, error)
    Validate() error
    Save() error
    GetDatabasePath() string
    GetProjectRoots() []string
    GetIgnoredPaths() []string
    GetLogLevel() string
}
```

### Interface 2: Database Interface
```go
// pkg/models/database.go
type DatabaseInterface interface {
    Connect() error
    Close() error
    Migrate() error
    ValidateSchema() error
    GetVersion() int
    ExecuteQuery(query string, args ...interface{}) (*Result, error)
}
```

### Interface 3: Logger Interface
```go
// pkg/models/logger.go
type LoggerInterface interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    With(component string) Logger
    Sync() error
}
```

### Error Handling Standard:
```go
// pkg/models/errors.go
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
    ErrPathNotFound      = "PATH_NOT_FOUND"
    ErrPermission        = "PERMISSION_DENIED"
)
```

---

## Integration Architecture

### Component Lifecycle Management
```
┌───────────────────────────────────────────────────────────┐
│                   Startup Sequence                           │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  1. Load Configuration                                     │
│     ├── Read ~/.portfolio/config.toml                     │
│     ├── Merge with defaults                               │
│     └── Validate configuration                             │
│                                                            │
│  2. Initialize Logging                                     │
│     ├── Check PORTFOLIO_LOG_LEVEL environment              │
│     ├── Apply log level (env > config > default)           │
│     └── Start structured logging                          │
│                                                            │
│  3. Connect Database                                       │
│     ├── Get database path from config                     │
│     ├── Open SQLite connection                            │
│     ├── Run pending migrations                            │
│     └── Validate schema                                   │
│                                                            │
│  4. Initialize CLI                                         │
│     ├── Setup cobra commands                              │
│     ├── Wire up command handlers                          │
│     └── Register command completion                        │
│                                                            │
└───────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────────┐
│                   Shutdown Sequence                        │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  1. Stop CLI Operations                                   │
│     ├── Cancel running commands                           │
│     └── Clean up resources                                │
│                                                            │
│  2. Close Database Connections                            │
│     ├── Complete pending transactions                     │
│     ├── Close connection pool                             │
│     └── Release file locks                                │
│                                                            │
│  3. Flush Logging                                          │
│     ├── Flush buffered messages                          │
│     ├── Close log writers                                │
│     └── Handle flush failures                             │
│                                                            │
│  4. Save Configuration (if needed)                          │
│     ├── Persist any changes                              │
│     └── Release file handles                              │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

### Data Flow Architecture
```
┌───────────────────────────────────────────────────────────┐
│                    Data Flow Diagram                        │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  User Input ──► CLI ──► Config ──► Database ──► Status     │
│        │           │         │           │                  │
│        │           └─────────┴───────────┴──────► Logging   │
│        │                                                      │
│        ▼                                                      │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐    │
│  │  Init Cmd   │───►│  Status Cmd │───►│ Doctor Cmd  │    │
│  │  (setup)    │    │  (health)   │    │ (diagnostics)│    │
│  └─────────────┘    └─────────────┘    └─────────────┘    │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

### Error Handling and Recovery Patterns
```
┌───────────────────────────────────────────────────────────┐
│                 Error Handling Architecture                 │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Config    │───►│   Logger    │───►│    CLI      │  │
│  │   Errors    │    │   Errors    │    │   Errors    │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │          │
│         ▼                   ▼                   ▼          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │ Portfolio   │    │   Structured │    │   User      │  │
│  │ Error Types │    │   Logging   │    │  Feedback   │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │          │
│         └───────────────────┴───────────────────┘          │
│                             ▼                                │
│                   ┌─────────────┐                          │
│                   │  Recovery   │                          │
│                   │  Actions    │                          │
│                   └─────────────┘                          │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

**Error Recovery Patterns:**
1. **Configuration Errors**: Suggest fixes, create defaults, prompt user
2. **Database Errors**: Retry with backoff, suggest migrations, validate paths
3. **Permission Errors**: Suggest chmod, validate ownership, check disk space
4. **Validation Errors**: Specific field feedback, actionable messages

---

## Technical Specifications

### Go Module Dependencies
```go
module github.com/nerddevsltd/portfolio

go 1.21

require (
    github.com/spf13/cobra v1.8.0          // CLI framework
    go.uber.org/zap v1.26.0                // Structured logging
    github.com/mattn/go-sqlite3 v1.14.18   // SQLite driver
)
```

**Dependency Management Strategy:**
- **Minimization**: Only 4 external dependencies
- **Standard Library**: Prefer built-in packages (filepath, io, os, etc.)
- **Version Locking**: Pin versions in go.mod
- **Security**: Regular dependency audits

### Configuration Data Structures
**Validation Rules:**
- **Path Validation**: Must exist, be readable, and be directory
- **Database Validation**: Parent directory writable
- **Enum Validation**: Exact case-sensitive matches
- **Required Fields**: Block startup with actionable errors

**Default Behavior:**
- Create `~/.portfolio/config.toml` if missing
- Apply defaults before validation
- Merge partial configs intelligently
- Suggest fixes for validation failures

### Logging Framework Integration
**Library Selection:** `zap` (uber's high-performance logger)

**Configuration Priority:**
1. Environment variable `PORTFOLIO_LOG_LEVEL`
2. Config file setting
3. Default INFO level

**Performance Specifications:**
- Maximum overhead: < 1ms per entry
- Buffer size: 1MB maximum
- Flush frequency: 1 second or 75% capacity
- Thread safety: Full concurrent safety

### CLI Framework Integration
**Library Selection:** `cobra` (industry standard CLI framework)

**Command Structure:**
```bash
portfolio [command] [flags]
portfolio init [flags]
portfolio status [flags]  
portfolio doctor [flags]
```

**Error Handling:**
- Clear, actionable error messages
- Suggestions for fixing common errors
- Appropriate exit codes (0 for success, non-zero for errors)
- No stack traces in user-facing errors

### Database Connection Management
**Connection Configuration:**
```go
db, err := sql.Open("sqlite3", "file:portfolio.db?_foreign_keys=on")
db.SetMaxOpenConns(25)     // Maximum connections
db.SetMaxIdleConns(25)     // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute)
```

**Migration System:**
- File-based SQL migrations
- Version tracking with checksums
- Transactional application
- Automatic pending detection
- Rollback support for all migrations

---

## Engineering Principles Alignment

### "Engine Knows, Agent Thinks"
**Implementation:**
- Configuration: Deterministic file loading and validation
- Logging: Pure structured output, no AI reasoning
- CLI: Administrative commands only, no semantic operations
- Database: Fact storage with computed indicators

### "Store Facts, Compute Indicators"
**Implementation:**
- **Store**: git_head, timestamps, documentation_hash, file paths
- **Compute**: analysis_outdated, needs_analysis, recently_modified
- **Avoid**: Derived state storage, duplicated metadata

### "Local First"
**Implementation:**
- SQLite database on user's machine
- Configuration in user's home directory
- No network operations
- No cloud dependencies

### "Minimize Dependencies"  
**Implementation:**
- Only 4 external dependencies
- Standard library preference
- Version pinning and auditing
- Security scanning

### "Capabilities over Workflows"
**Implementation:**
- Small composable interfaces
- No high-level workflows in engine
- AI agents compose operations
- CLI provides administrative capabilities only

### "Dashboard is Read-Only"
**Implementation:**
- Database prepared for future read-only access
- No modification commands in CLI
- Logging captures all state changes
- Audit trail through structured logs

### "CLI is Administrative"
**Implementation:**
- Commands limited to init, status, doctor
- Interactive setup only for initialization
- No day-to-day interaction commands
- Clear scope boundaries maintained

### "Agent Agnostic"
**Implementation:**
- No Claude-specific dependencies
- Generic MCP preparation (future stories)
- Extensible configuration system
- Standard interfaces for future integrations

### "Single Knowledge Model"
**Implementation:**
- Database schema matches KnowledgeModel.md
- All interfaces use canonical entities
- Configuration system consistent across interfaces
- No interface-specific representations

---

## Future Readiness

### HTTP API Interface Preparation
**Architecture Support:**
- Database abstraction layer supports multiple access patterns
- Configuration system ready for runtime settings
- Logging framework supports request tracking
- Error handling provides structured responses

### MCP Server Integration Preparation
**Architecture Support:**
- Component-based logging for tool tracking
- Configuration system for agent-specific settings  
- Database abstraction for tool-based access
- Error handling for tool responses

### Extensible Architecture
**Design Patterns:**
- Interface-based design allows alternative implementations
- Component isolation enables independent evolution
- Configuration system supports new sections
- Database migration system supports schema evolution

### Subsequent Epic Support
**Epic 2 (Discovery):** Project scanning, metadata extraction
- Database schema ready for project storage
- Configuration system supports project roots
- Logging framework supports discovery operations

**Epic 3 (Documentation Indexing):** Document discovery and storage
- Database documents table prepared
- Logging supports document processing tracking
- Configuration system supports document sources

**Epic 4 (MCP Server):** AI agent integration
- Database abstraction supports tool-based access
- Component-based logging tracks tool operations
- Configuration system supports agent-specific settings

**Epic 5 (HTTP API):** Dashboard backend
- Database abstraction supports HTTP-based queries
- Configuration system supports runtime settings
- Logging framework supports request tracking

**Epic 6 (Dashboard Frontend):** Web interface
- HTTP API preparation complete
- Database queries optimized for dashboard access
- Logging supports request/response tracking

---

## Security and Privacy Considerations

### File System Security
- Config file permissions: chmod 600 (owner read/write only)
- Database file permissions: chmod 600 (owner read/write only)
- Proper error handling for permission issues
- No credential storage in configuration

### Database Security
- SQLite local-only (no network exposure)
- SQL injection prevention via parameterized queries
- Safe handling of user-provided paths
- Transaction isolation for data consistency

### Error Handling Security
- No stack traces in user-facing errors
- Path traversal prevention
- Safe file path validation
- No sensitive data in log output

---

## Performance Considerations

### Startup Performance
- Config loading: < 100ms
- Database connection: < 200ms
- Schema validation: < 500ms
- CLI command response: < 1s for status/doctor

### Database Performance
- Single project insert: < 10ms
- Query with 1000 projects: < 100ms
- Schema migration: < 5s per migration
- WAL mode for read/write concurrency

### Logging Performance
- Average time per log entry: < 1ms
- Maximum buffer size: 1MB
- Flush frequency: 1 second or 75% capacity
- Asynchronous flushing for performance

---

## Testing and Validation Infrastructure

### Component Testing Strategy
```
┌───────────────────────────────────────────────────────────┐
│                   Testing Architecture                      │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │   Unit      │───►│ Integration │───►│    E2E      │  │
│  │   Tests     │    │    Tests    │    │   Tests     │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│         │                   │                   │          │
│         ▼                   ▼                   ▼          │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐  │
│  │ Individual  │    │  Component  │    │   Complete  │  │
│  │ Functions   │    │ Interaction │    │  Workflows  │  │
│  └─────────────┘    └─────────────┘    └─────────────┘  │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

**Test Coverage Targets:**
- Configuration loading/validation: 90%+
- Database operations: 80%+
- Logging framework: 70%+
- CLI commands: 60%+

**Test Organization:**
- `internal/config/config_test.go`
- `internal/database/database_test.go`
- `internal/logging/logger_test.go`
- `cmd/portfolio/cli_test.go`

---

## Architecture Artifacts

### New Components Created
1. **Go Project Bootstrap** (`cmd/portfolio/`, `internal/`, `pkg/`)
2. **Configuration System** (`internal/config/`)
3. **Logging Framework** (`internal/logging/`)
4. **CLI Framework** (`internal/cli/`)
5. **SQLite Initialization** (`internal/database/`)

### Shared Interfaces Defined
1. **Configuration Interface** (`pkg/models/config.go`)
2. **Database Interface** (`pkg/models/database.go`)
3. **Logger Interface** (`pkg/models/logger.go`)
4. **Error Handling Standard** (`pkg/models/errors.go`)

### New Dependencies Added
1. **cobra** (`github.com/spf13/cobra v1.8.0`)
2. **zap** (`go.uber.org/zap v1.26.0`)
3. **sqlite3** (`github.com/mattn/go-sqlite3 v1.14.18`)

### Integration Points Established
1. **Configuration-Database Integration** (INT-001)
2. **Logging-CLI Integration** (INT-002)
3. **CLI-Configuration Integration** (INT-003)
4. **Database-Configuration Persistence** (INT-004)

---

## Architecture Summary

**Component Count:** 5 new components, 4 shared interfaces  
**New Dependencies:** 3 external Go packages  
**Integration Points:** 4 major integration pathways  
**Implementation Complexity:** Medium (8 days estimated)

**Architecture Strengths:**
1. ✅ **Modular Design**: Clear component boundaries with defined interfaces
2. ✅ **Engineering Principles**: Full alignment with Guideline.md principles
3. ✅ **Future Ready**: Supports HTTP API and MCP server integration
4. ✅ **Local First**: Complete local-only operation
5. ✅ **Deterministic**: No AI reasoning in foundation components
6. ✅ **Extensible**: Migration system and configuration support evolution

**Implementation Readiness:**
- All component dependencies clearly defined
- Integration points specified and testable
- Error handling patterns established
- Performance requirements documented
- Security considerations addressed

**Epic Success Criteria:**
- Developers can initialize Portfolio via CLI
- Configuration system supports user-defined project roots
- Logging supports operational visibility
- SQLite database stores canonical knowledge model
- Foundation supports deterministic operations per Guideline.md

---

**End of Epic 1 - Project Foundation Architecture**

*This architecture document provides the complete technical design for implementing all 5 stories in Epic 1, ensuring cohesive integration while following Portfolio engineering principles and preparing for future epics.*