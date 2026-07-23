# Epic 1 - Project Foundation Test Cases

**Epic ID:** EPIC-01  
**Epic Name:** Project Foundation  
**Document Version:** 1.0  
**Generated:** 2026-07-22  
**Coverage:** All 5 stories + Integration + Acceptance Criteria

---

## Test Summary

- **Total Test Cases:** 87
- **Total Acceptance Criteria:** 42
- **Coverage Areas:** Unit (34), Integration (24), E2E (12), Performance (8), Security (6), Platform (3)

---

## Story 1.1: Bootstrap Go Project

### Unit Tests

#### TC-01: Go Module Initialization
**Type:** Unit  
**Priority:** Critical  
**Requirement:** GO-001  
**Given:** New project directory structure  
**When:** Developer initializes Go module  
**Then:** 
- `go.mod` file created with proper module name
- Go version set to 1.21+ or higher
- No dependencies initially declared
- `go mod verify` succeeds

**Verification:** Run `go mod verify` and verify `go.mod` contents

---

#### TC-02: Standard Directory Structure Creation
**Type:** Unit  
**Priority:** Critical  
**Requirement:** GO-002  
**Given:** Empty project repository  
**When:** Directory structure is created  
**Then:**
- `cmd/portfolio/` directory exists with `main.go`
- `internal/config/`, `internal/database/`, `internal/logging/`, `internal/engine/` exist
- `pkg/models/` directory exists
- `docs/` directory exists
- All required directories are present

**Verification:** Run `find . -type d -name "cmd" -o -name "internal" -o -name "pkg"` and verify structure

---

#### TC-03: Git Configuration Setup
**Type:** Unit  
**Priority:** Medium  
**Requirement:** GO-003  
**Given:** Project repository  
**When:** Git configuration files are created  
**Then:**
- `.gitignore` contains Go-specific ignores (*.exe, *.test, *.out, vendor/)
- LICENSE file present with MIT or appropriate license
- `git status` shows only expected files

**Verification:** Check `.gitignore` contents and LICENSE file existence

---

#### TC-04: Build Documentation Completeness
**Type:** Unit  
**Priority:** Medium  
**Requirement:** GO-004  
**Given:** README.md in project root  
**When:** Developer reads README for build instructions  
**Then:**
- Prerequisites section specifies Go version requirements
- Build instructions: `go build ./cmd/portfolio`
- Run instructions: `./portfolio --help`
- Development setup instructions included
- Link to full documentation in docs/

**Verification:** New developer can successfully build and run following README only

---

### Integration Tests

#### TC-05: Build Process Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** GO-001, GO-002, GO-004  
**Given:** Properly structured Go project  
**When:** Developer runs `go build ./cmd/portfolio`  
**Then:**
- Build completes without errors
- Executable created in current directory
- No compilation warnings or errors
- All dependencies resolved correctly

**Verification:** Successful build produces working portfolio executable

---

#### TC-06: Project Structure Validation
**Type:** Integration  
**Priority:** Medium  
**Requirements:** GO-002, ARCH-001  
**Given:** Complete project structure  
**When:** Validation script checks directory organization  
**Then:**
- All required directories present and accessible
- Package structure follows Go conventions
- No misplaced files or directories
- Import paths work correctly

**Verification:** Structure validation passes without errors

---

## Story 1.2: Configuration System

### Unit Tests

#### TC-07: TOML Configuration Parsing
**Type:** Unit  
**Priority:** Critical  
**Requirement:** CFG-001  
**Given:** Valid TOML configuration file  
**When:** Configuration loader reads file  
**Then:**
- TOML file parses without errors
- All sections (general, discovery, logging) accessible
- All configuration fields loaded correctly
- No parsing errors or warnings

**Verification:** Config structure matches expected TOML format

```toml
[general]
database_path = "/home/user/.portfolio/portfolio.db"

[discovery]
project_roots = ["/home/user/dev", "/home/user/projects"]
ignored_paths = ["node_modules", ".git", "vendor"]

[logging]
level = "INFO"
```

---

#### TC-08: Configuration Schema Validation
**Type:** Unit  
**Priority:** High  
**Requirement:** CFG-002  
**Given:** Configuration with various field values  
**When:** Configuration validator runs  
**Then:**
- Required fields validated: database_path, project_roots, logging.level
- Path validation checks existence and readability
- Enum validation for log level (DEBUG, INFO, WARN, ERROR)
- Invalid values produce specific error messages

**Verification:** Validation catches all schema violations with actionable errors

---

#### TC-09: Default Configuration Creation
**Type:** Unit  
**Priority:** High  
**Requirement:** CFG-003  
**Given:** Missing configuration file  
**When:** Configuration loader runs  
**Then:**
- `~/.portfolio/config.toml` created with defaults
- Default database path: `~/.portfolio/portfolio.db`
- Default project roots: empty array
- Default ignored_paths: standard patterns
- Default log level: INFO

**Verification:** Missing config generates default file automatically

---

#### TC-10: Configuration Error Handling
**Type:** Unit  
**Priority:** High  
**Requirement:** CFG-005  
**Given:** Various invalid configuration scenarios  
**When:** Configuration loader processes files  
**Then:**
- Invalid TOML syntax → Parse error with line number
- Invalid paths → Path validation error
- Invalid enum values → Enum validation error
- Missing required fields → Validation error
- All errors include clear fix suggestions

**Verification:** All error scenarios produce actionable error messages

---

#### TC-11: Configuration File Structure
**Type:** Unit  
**Priority:** Medium  
**Requirement:** CFG-004  
**Given:** Configuration file loaded into memory  
**When:** Internal structure is examined  
**Then:**
- Config struct contains: General, Discovery, Logging sections
- GeneralConfig contains DatabasePath field
- DiscoveryConfig contains ProjectRoots and IgnoredPaths fields
- LoggingConfig contains Level field
- Structure aligns with PlatformSpecification.md

**Verification:** Internal config representation matches specification

---

#### TC-12: Configuration Path Validation
**Type:** Unit  
**Priority:** High  
**Requirement:** CFG-002  
**Given:** Configuration with various path values  
**When:** Path validator runs  
**Then:**
- Existing paths pass validation
- Non-existent paths fail with specific error
- Invalid characters detected and rejected
- Relative paths resolved and validated
- Permission issues detected

**Verification:** Path validation catches all invalid path scenarios

---

### Integration Tests

#### TC-13: Configuration Loading with Defaults
**Type:** Integration  
**Priority:** High  
**Requirements:** CFG-001, CFG-003  
**Given:** Partial configuration file  
**When:** Configuration loader merges with defaults  
**Then:**
- Missing fields populated with defaults
- Provided fields override defaults
- Resulting configuration is complete and valid
- Merge process logged appropriately

**Verification:** Partial config produces complete valid configuration

---

#### TC-14: Configuration-Database Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** CFG-003, DB-001, INT-001  
**Given:** Configuration with database path  
**When:** Database initialization uses config  
**Then:**
- Database path read from configuration
- Database created at configured location
- Configuration validation triggers database accessibility check
- Integration logged properly

**Verification:** Database operations use configured path correctly

---

#### TC-15: Configuration File Permissions
**Type:** Integration  
**Priority:** Medium  
**Requirements:** CFG-005, SEC-001  
**Given:** Configuration file with various permissions  
**When:** Configuration loader attempts to read  
**Then:**
- Readable files (600, 644) load successfully
- Unreadable files (000) produce permission error
- Error messages indicate permission issue
- Suggest chmod command for fix

**Verification:** Permission errors handled with helpful messages

---

### E2E Tests

#### TC-16: End-to-End Configuration Workflow
**Type:** E2E  
**Priority:** High  
**Requirements:** CFG-001, CFG-002, CFG-003  
**Given:** Fresh system installation  
**When:** User runs portfolio for first time  
**Then:**
- Default configuration created automatically
- Configuration validated on startup
- System initializes successfully
- User can proceed to portfolio operations

**Verification:** First run creates valid configuration without manual intervention

---

#### TC-17: Configuration Migration and Backward Compatibility
**Type:** E2E  
**Priority:** Medium  
**Requirements:** CFG-001, CFG-004  
**Given:** Old configuration format from previous version  
**When:** New version loads old configuration  
**Then:**
- Old format parsed successfully or migration applied
- Missing fields populated with defaults
- User notified of any configuration changes
- System continues to work correctly

**Verification:** Old configurations work with new version or migrate properly

---

## Story 1.3: Logging Framework

### Unit Tests

#### TC-18: Structured Logging Implementation
**Type:** Unit  
**Priority:** High  
**Requirement:** LOG-001  
**Given:** Logging framework initialized  
**When:** Log entries are written  
**Then:**
- Log entries in JSON format with key-value pairs
- Each entry contains: level, timestamp, component, message
- Additional context fields included as needed
- Consistent format across all components
- Thread-safe operations confirmed

**Verification:** Log output parses as valid JSON with required fields

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

---

#### TC-19: Log Level Filtering
**Type:** Unit  
**Priority:** High  
**Requirement:** LOG-002  
**Given:** Logger configured with INFO level  
**When:** Messages logged at different levels  
**Then:**
- DEBUG messages not displayed
- INFO messages displayed
- WARN messages displayed  
- ERROR messages displayed
- Level filtering works correctly

**Verification:** Only messages at configured level and above appear

---

#### TC-20: Standard Output Configuration
**Type:** Unit  
**Priority:** Medium  
**Requirement:** LOG-003  
**Given:** Logger configured for stdout  
**When:** Log messages are written  
**Then:**
- All output goes to stdout, not stderr
- No file writing occurs
- Output format consistent (JSON or human-readable)
- Format selection works correctly

**Verification:** Log messages appear only on stdout

---

#### TC-21: Environment Variable Configuration
**Type:** Unit  
**Priority:** Medium  
**Requirement:** LOG-004  
**Given:** Environment variable PORTFOLIO_LOG_LEVEL set  
**When:** Logger initializes  
**Then:**
- Environment variable overrides config file
- Invalid environment values rejected
- Config file value used when env var not set
- Default INFO used when neither set
- Priority order: env var > config > default

**Verification:** PORTFOLIO_LOG_LEVEL=DEBUG enables debug logging

---

#### TC-22: Component-Based Logging
**Type:** Unit  
**Priority:** Medium  
**Requirement:** LOG-005  
**Given:** Logger with component support  
**When:** Different components log messages  
**Then:**
- Each log entry includes component field
- Required components present: config, database, cli, discovery, engine
- Component filtering works in DEBUG mode
- Component-specific log levels supported

**Verification:** Component field appears in all log entries

---

#### TC-23: Logging Performance
**Type:** Performance  
**Priority:** High  
**Requirement:** LOG-001  
**Given:** Logger under load  
**When:** 1000 log entries written  
**Then:**
- Average time per log entry < 1ms
- No blocking I/O operations
- Asynchronous flushing prevents performance impact
- Memory buffer stays under 1MB

**Verification:** Performance benchmarks meet requirements

---

#### TC-24: Logging Thread Safety
**Type:** Unit  
**Priority:** High  
**Requirement:** LOG-001  
**Given:** Multiple goroutines logging simultaneously  
**When:** 100 goroutines log 100 messages each  
**Then:**
- All messages written without corruption
- No race conditions detected
- Proper synchronization for shared resources
- Thread-safe operations confirmed

**Verification:** No race conditions or message corruption under concurrent load

---

### Integration Tests

#### TC-25: Logging-CLI Integration
**Type:** Integration  
**Priority:** Medium  
**Requirements:** LOG-001, CLI-001, INT-002  
**Given:** CLI command execution  
**When:** Commands log events  
**Then:**
- Command execution start/end logged
- Configuration changes logged
- Database operations logged
- User interactions logged
- All logs include appropriate component

**Verification:** CLI operations generate appropriate log entries

---

#### TC-26: Log Level Persistence Across Components
**Type:** Integration  
**Priority:** Medium  
**Requirements:** LOG-002, LOG-004  
**Given:** Log level set via environment variable  
**When:** Multiple components log messages  
**Then:**
- All components respect configured log level
- No component bypasses level filtering
- Component-specific levels work when configured
- Consistent behavior across components

**Verification:** Log level applied consistently across all components

---

### Performance Tests

#### TC-27: Logging Buffer Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** LOG-001  
**Given:** Logger with 1MB buffer  
**When:** Buffer reaches 75% capacity  
**Then:**
- Buffer flushes automatically
- No message loss occurs
- Flush completes within acceptable time
- Performance remains stable under high volume

**Verification:** Buffer management handles high-volume logging without loss

---

#### TC-28: Logging Shutdown Flush
**Type:** Performance  
**Priority:** Medium  
**Requirement:** LOG-001  
**Given:** Logger with buffered messages  
**When:** Application shutdown initiated  
**Then:**
- All buffered messages flushed
- No message loss on shutdown
- Shutdown completes within reasonable time
- Flush failures handled gracefully

**Verification:** Clean shutdown without message loss

---

## Story 1.4: CLI Framework

### Unit Tests

#### TC-29: CLI Framework Foundation
**Type:** Unit  
**Priority:** High  
**Requirement:** CLI-001  
**Given:** Portfolio CLI installed  
**When:** User runs `portfolio --help`  
**Then:**
- Help text displays usage information
- Command structure shown (init, status, doctor)
- Flag parsing working correctly
- Command completion support available
- Error handling functional

**Verification:** `portfolio --help` shows comprehensive usage information

---

#### TC-30: Init Command Interactive Prompts
**Type:** Unit  
**Priority:** Critical  
**Requirement:** CLI-002  
**Given:** User runs `portfolio init`  
**When:** Interactive prompts displayed  
**Then:**
- Prompts for project root directories
- Prompts for database path with default
- Prompts for log level with default
- Confirmation before writing configuration
- All prompts clear and helpful

**Verification:** Interactive flow guides user through configuration

---

#### TC-31: Init Command File Operations
**Type:** Unit  
**Priority:** Critical  
**Requirement:** CLI-002  
**Given:** User provides valid responses to init prompts  
**When:** Init command completes  
**Then:**
- `~/.portfolio/` directory created if missing
- `~/.portfolio/config.toml` created with user values
- Database initialized at configured path
- Success message displayed with next steps

**Verification:** Command creates config file and initializes database

---

#### TC-32: Status Command Information Display
**Type:** Unit  
**Priority:** High  
**Requirement:** CLI-003  
**Given:** Portfolio system running  
**When:** User runs `portfolio status`  
**Then:**
- Portfolio Engine status shown (running/stopped)
- Database location and accessibility displayed
- Configuration file location shown
- Number of discovered projects shown
- Last discovery timestamp shown
- Any warnings or errors displayed

**Verification:** Status command shows accurate system information

---

#### TC-33: Doctor Command Diagnostic Checks
**Type:** Unit  
**Priority:** Medium  
**Requirement:** CLI-004  
**Given:** Portfolio system installed  
**When:** User runs `portfolio doctor`  
**Then:**
- Configuration file accessibility and validity checked
- Database file accessibility and integrity checked
- Project roots accessibility verified
- File permissions validated
- Disk space availability checked
- Go environment and dependencies verified

**Verification:** All diagnostic checks run and report accurately

---

#### TC-34: Doctor Command Exit Codes
**Type:** Unit  
**Priority:** Medium  
**Requirement:** CLI-004  
**Given:** Various diagnostic outcomes  
**When:** Doctor command completes  
**Then:**
- Exit code 0 when all checks pass
- Exit code 1 when one or more checks fail
- Exit code 2 when critical errors detected
- Exit codes match specification

**Verification:** Exit codes reflect diagnostic results correctly

---

#### TC-35: CLI Error Handling
**Type:** Unit  
**Priority:** Medium  
**Requirement:** CLI-005  
**Given:** Various error scenarios  
**When:** CLI commands encounter errors  
**Then:**
- Clear, actionable error messages displayed
- Suggestions for fixing common errors provided
- Appropriate exit codes used
- No stack traces in user-facing errors
- Errors localized and specific

**Verification:** Error messages are clear and helpful

---

#### TC-36: CLI Administrative Scope
**Type:** Unit  
**Priority:** High  
**Requirement:** CLI-006  
**Given:** CLI command set  
**When:** Commands are examined  
**Then:**
- In scope: init, status, doctor, config, diagnostics
- Out of scope: project discovery, analysis operations
- Administrative nature of commands maintained
- No day-to-day portfolio interaction commands present

**Verification:** CLI commands align with administrative scope only

---

### Integration Tests

#### TC-37: CLI-Configuration Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** CLI-002, CLI-003, INT-003  
**Given:** User runs CLI commands  
**When:** Configuration operations performed  
**Then:**
- `init` command creates/updates configuration
- `status` command reads and displays configuration
- `doctor` command validates configuration
- Configuration changes visible across CLI commands
- User workflow: init → create config → validate → initialize database

**Verification:** Configuration changes persist across CLI sessions

---

#### TC-38: CLI-Database Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** CLI-002, CLI-003, DB-001  
**Given:** User runs init command  
**When:** Database initialization triggered  
**Then:**
- Database created at configured path
- Schema validation performed
- Configuration and database integration working
- Error handling covers both config and database issues

**Verification:** CLI commands trigger appropriate database operations

---

#### TC-39: CLI-Logging Integration
**Type:** Integration  
**Priority:** Medium  
**Requirements:** CLI-001, LOG-001, INT-002  
**Given:** CLI commands executing  
**When:** Logging integration active  
**Then:**
- Log level from environment or config file used
- CLI commands emit appropriate log events
- Structured logging supports command execution tracking
- Error logging for CLI failures functional

**Verification:** CLI operations generate appropriate log entries

---

### E2E Tests

#### TC-40: End-to-End Init Workflow
**Type:** E2E  
**Priority:** High  
**Requirements:** CLI-002, CFG-003, DB-001  
**Given:** Fresh system without Portfolio  
**When:** User runs `portfolio init` and provides valid inputs  
**Then:**
- Configuration file created successfully
- Database initialized with schema
- System ready for use
- Success message with next steps displayed
- No manual intervention required

**Verification:** Complete initialization workflow works end-to-end

---

#### TC-41: End-to-End Diagnostics Workflow
**Type:** E2E  
**Priority:** Medium  
**Requirements:** CLI-004, CFG-002, DB-003  
**Given:** Portfolio system with various states  
**When:** User runs `portfolio doctor`  
**Then:**
- All diagnostic checks performed
- Issues identified with specific locations
- Actionable suggestions provided
- Exit code reflects system health
- Report readable and actionable

**Verification:** Diagnostic workflow provides comprehensive system health check

---

#### TC-42: CLI User Experience Under Error Conditions
**Type:** E2E  
**Priority:** Medium  
**Requirements:** CLI-005, CFG-005  
**Given:** Various system error conditions  
**When:** User runs CLI commands  
**Then:**
- Invalid paths handled with helpful error messages
- Permission errors guide user to fix permissions
- Missing files detected and reported clearly
- Recovery suggestions provided
- User can proceed to fix issues based on error messages

**Verification:** Error conditions handled with excellent UX

---

## Story 1.5: SQLite Initialization

### Unit Tests

#### TC-43: Database File Creation
**Type:** Unit  
**Priority:** Critical  
**Requirement:** DB-001  
**Given:** Configured database path  
**When:** Database initialization runs  
**Then:**
- Database file created at configured path
- Parent directories created if needed
- File permissions set correctly (read/write for owner)
- File accessible for subsequent operations

**Verification:** Database file exists at correct location with proper permissions

---

#### TC-44: Database Connection Management
**Type:** Unit  
**Priority:** Critical  
**Requirement:** DB-002  
**Given:** Database file exists  
**When:** Connection opened  
**Then:**
- Connection opens successfully
- Foreign key constraints enabled
- WAL mode configured for concurrency
- Busy timeout set for concurrent access
- Connection validated as healthy

**Verification:** Connection opens with proper configuration

---

#### TC-45: Database Connection Closing
**Type:** Unit  
**Priority:** Critical  
**Requirement:** DB-002  
**Given:** Open database connection  
**When:** Application shutdown initiated  
**Then:**
- Connection closes properly
- All pending transactions completed
- No connection leaks
- Resources cleaned up
- Shutdown logged appropriately

**Verification:** Connection closes cleanly without errors

---

#### TC-46: Database Schema Validation
**Type:** Unit  
**Priority:** Critical  
**Requirement:** DB-003  
**Given:** Database initialization  
**When:** Schema validation runs  
**Then:**
- All 9 required tables present: projects, metadata, documents, analyses, features, technologies, project_technologies, relationships, configuration
- Table structures match PlatformSpecification.md
- Foreign key relationships validated
- Indexes and constraints checked
- All validation checks pass

**Verification:** All required tables exist with correct structure

---

#### TC-47: Database Schema Creation
**Type:** Unit  
**Priority:** Critical  
**Requirement:** DB-003  
**Given:** Fresh database initialization  
**When:** Schema creation runs  
**Then:**
- All tables created according to specification
- Proper column types and constraints
- Foreign key relationships established
- Indexes created on appropriate columns
- No schema validation errors

**Verification:** Complete schema created matching PlatformSpecification.md

---

#### TC-48: Migration System Version Tracking
**Type:** Unit  
**Priority:** High  
**Requirement:** DB-004  
**Given:** Migration system initialized  
**When:** Migrations applied  
**Then:**
- `schema_migrations` table created
- Version numbers tracked correctly
- Migration names stored
- Applied timestamps recorded
- Checksums validated

**Verification:** Migration tracking system functional

---

#### TC-49: Migration Application Process
**Type:** Unit  
**Priority:** High  
**Requirement:** DB-004  
**Given:** Pending migrations available  
**When:** Migration system runs  
**Then:**
- Current schema version identified
- Pending migrations (version > current) selected
- Migrations applied in numerical order
- Each migration applied within transaction
- Schema validation after each migration
- Migration table updated on success

**Verification:** Migrations apply correctly with proper tracking

---

#### TC-50: Migration Rollback Process
**Type:** Unit  
**Priority:** High  
**Requirement:** DB-004  
**Given:** Migration with rollback available  
**When:** Rollback initiated  
**Then:**
- Corresponding .down.sql file executed
- Changes reversed atomically
- Migration table updated to previous version
- Rollback to any version supported
- No data corruption after rollback

**Verification:** Rollback reverses migration changes correctly

---

#### TC-51: Migration Failure Handling
**Type:** Unit  
**Priority:** High  
**Requirement:** DB-004  
**Given:** Migration that fails during execution  
**When:** Migration error occurs  
**Then:**
- Failed migration rolled back automatically
- Existing data not corrupted
- Clear error message indicating failure reason
- System startup blocked by failed migration
- Manual intervention suggested

**Verification:** Migration failures don't corrupt data and provide clear error messages

---

#### TC-52: Schema Alignment with Knowledge Model
**Type:** Unit  
**Priority:** Critical  
**Requirement:** DB-005  
**Given:** Complete database schema  
**When:** Schema compared to Knowledge Model  
**Then:**
- projects table → Project entity
- metadata table → Project metadata
- documents table → Documentation entity
- analyses table → Analysis entity
- features table → Feature entity
- technologies table → Technology entity
- project_technologies table → Project-Technology relationship
- relationships table → Relationship entity
- configuration table → System configuration

**Verification:** Schema structure matches KnowledgeModel.md entities

---

#### TC-53: Database Connection Pool Management
**Type:** Unit  
**Priority:** Medium  
**Requirement:** DB-002, DB-006  
**Given:** Multiple concurrent operations  
**When:** Connection pool under load  
**Then:**
- Connection pool handles concurrent access
- Pool size appropriate for workload
- No connection exhaustion under load
- Connection reuse working
- Pool cleanup functional

**Verification:** Connection pool manages concurrent operations effectively

---

### Integration Tests

#### TC-54: Configuration-Database Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** DB-001, CFG-003, INT-001  
**Given:** Configuration with database path  
**When:** Database initialization using config  
**Then:**
- Database path loaded from configuration
- Database created at configured location
- Configuration validation triggers database check
- Doctor command validates both components

**Verification:** Config changes trigger appropriate database operations

---

#### TC-55: Database-Configuration Persistence
**Type:** Integration  
**Priority:** Medium  
**Requirements:** DB-005, INT-004  
**Given:** Configuration changes  
**When:** Settings persisted to database  
**Then:**
- `configuration` table stores runtime settings
- Configuration changes persist to database
- Configuration versioning supported
- Changes read back correctly

**Verification:** Configuration persists correctly to database

---

#### TC-56: Schema Migration Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** DB-003, DB-004  
**Given:** Database with existing schema  
**When:** New version with migrations runs  
**Then:**
- Current schema version detected
- Pending migrations identified
- Migrations applied in correct order
- Integration with startup process working
- Migration completion logged

**Verification:** Schema migrations integrate properly with startup process

---

### E2E Tests

#### TC-57: End-to-End Database Initialization
**Type:** E2E  
**Priority:** High  
**Requirements:** DB-001, DB-003, CFG-003  
**Given:** Fresh system installation  
**When:** User runs `portfolio init`  
**Then:**
- Database created at configured path
- Complete schema initialized
- All tables created and validated
- Migration system initialized
- System ready for operations

**Verification:** Complete database initialization works end-to-end

---

#### TC-58: Database Schema Evolution
**Type:** E2E  
**Priority:** Medium  
**Requirements:** DB-004, DB-005  
**Given:** Database with existing schema  
**When:** New version with schema changes deployed  
**Then:**
- Schema version detected correctly
- Pending migrations applied automatically
- Schema validation passes
- Data integrity maintained
- System continues to work correctly

**Verification:** Schema evolution works smoothly without data loss

---

### Performance Tests

#### TC-59: Database Query Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** DB-006  
**Given:** Database with 1000 projects  
**When:** Common queries executed  
**Then:**
- Single project insert < 10ms
- Query with 1000 projects < 100ms
- Schema migration < 5s per migration
- Performance acceptable under load

**Verification:** Database benchmarks meet performance targets

---

#### TC-60: Database Concurrency Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** DB-006  
**Given:** Database with WAL mode enabled  
**When:** Multiple concurrent operations  
**Then:**
- Read/write concurrency working
- Lock conflicts handled properly
- Connection busy timeout functional
- Transaction isolation maintained
- No deadlocks or blocking issues

**Verification:** Concurrent access handled without performance degradation

---

## Integration and Cross-Story Tests

### Integration Tests

#### TC-61: Complete Component Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** All stories  
**Given:** All Epic 1 components implemented  
**When:** System initialized and commands run  
**Then:**
- Configuration system integrates with all components
- Logging works across all components
- CLI commands use all systems properly
- Database operations integrate correctly
- All integration points functional

**Verification:** All components integrate seamlessly

---

#### TC-62: Configuration-Logging-CLI Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** CFG, LOG, CLI  
**Given:** Complete system  
**When:** CLI commands executed  
**Then:**
- Configuration loaded and used
- Log level from config/env applied
- CLI operations logged correctly
- Integration points working
- No conflicts between systems

**Verification:** Three-way integration works correctly

---

#### TC-63: Error Propagation Across Components
**Type:** Integration  
**Priority:** Medium  
**Requirements:** All stories  
**Given:** Error condition in any component  
**When:** Error propagates through system  
**Then:**
- Errors caught at appropriate layer
- Error messages clear and actionable
- Stack traces not shown to users
- Recovery suggestions provided
- System state consistent after errors

**Verification:** Error handling consistent across all components

---

#### TC-64: Dependency Order Validation
**Type:** Integration  
**Priority:** Medium  
**Requirements:** All stories  
**Given:** Story dependency graph  
**When:** Implementation order validated  
**Then:**
- 1.1 → 1.2 → 1.5 dependency chain working
- 1.1 → 1.3 → 1.4 dependency chain working
- No circular dependencies
- Proper initialization order
- No missing dependencies

**Verification:** Dependency order respected and functional

---

#### TC-65: Startup Sequence Integration
**Type:** Integration  
**Priority:** High  
**Requirements:** All stories  
**Given:** System startup  
**When:** Components initialize in order  
**Then:**
- Configuration loads first
- Logging initializes second
- Database connects third
- CLI commands available fourth
- Startup completes without errors
- All components ready

**Verification:** Startup sequence works correctly

---

#### TC-66: Shutdown Sequence Integration  
**Type:** Integration  
**Priority:** Medium  
**Requirements:** All stories  
**Given:** System shutdown  
**When:** Components close in order  
**Then:**
- CLI operations stop
- Database connections close
- Logging flushes final messages
- Configuration saved if needed
- Clean shutdown achieved
- No resource leaks

**Verification:** Shutdown sequence works cleanly

---

### E2E Tests

#### TC-67: Complete User Journey - First Install
**Type:** E2E  
**Priority:** High  
**Requirements:** All acceptance criteria  
**Given:** Fresh system without Portfolio  
**When:** User installs and runs `portfolio init`  
**Then:**
- Installation completes successfully
- Configuration wizard guides user
- Database initialized with schema
- System ready for use
- User can run status and doctor commands
- Help documentation available

**Verification:** Complete first-install user journey works

---

#### TC-68: Complete User Journey - System Administration
**Type:** E2E  
**Priority:** Medium  
**Requirements:** CLI-001 through CLI-006  
**Given:** Portfolio system installed  
**When:** User performs administrative tasks  
**Then:**
- Status command shows system health
- Doctor command identifies issues
- Configuration changes work correctly
- Database operations functional
- Administrative workflow complete

**Verification:** Administrative user journey fully functional

---

#### TC-69: System Recovery After Errors
**Type:** E2E  
**Priority:** Medium  
**Requirements:** Error handling across all stories  
**Given:** System in error state  
**When:** User runs diagnostics and fixes  
**Then:**
- Doctor command identifies problems
- Clear fix suggestions provided
- User can recover system
- No data loss during recovery
- System returns to healthy state

**Verification:** Error recovery workflow works end-to-end

---

#### TC-70: Component Communication Validation
**Type:** E2E  
**Priority:** Medium  
**Requirements:** All integration points  
**Given:** All components running  
**When:** Information flows between components  
**Then:**
- Configuration changes propagate
- Log events capture all operations
- Database operations logged
- CLI commands access all needed data
- Communication channels working

**Verification:** All components communicate correctly

---

#### TC-71: Configuration Backup and Restore
**Type:** E2E  
**Priority:** Low  
**Requirements:** CFG-003, CFG-005  
**Given:** Working configuration  
**When:** Configuration backed up and restored  
**Then:**
- Backup created successfully
- Original configuration preserved
- Restore process works
- Validation after restore passes
- System functional after restore

**Verification:** Configuration backup and restore functional

---

#### TC-72: Database Backup and Restore
**Type:** E2E  
**Priority:** Low  
**Requirements:** DB-001, DB-002  
**Given:** Database with data  
**When:** Database backed up and restored  
**Then:**
- Backup created successfully
- Data integrity maintained
- Restore process works
- Schema validation after restore passes
- System functional after restore

**Verification:** Database backup and restore functional

---

## Security Tests

### Unit Tests

#### TC-73: Configuration File Permissions
**Type:** Security  
**Priority:** Medium  
**Requirement:** SEC-001  
**Given:** Configuration file created  
**When:** File permissions checked  
**Then:**
- Config file readable only by owner (chmod 600)
- No world-readable permissions
- No group-readable permissions
- Permissions validated on startup

**Verification:** Config file has secure permissions (600)

---

#### TC-74: Database File Permissions
**Type:** Security  
**Priority:** Medium  
**Requirement:** SEC-002  
**Given:** Database file created  
**When:** File permissions checked  
**Then:**
- Database file readable only by owner (chmod 600)
- No world-readable permissions
- No group-readable permissions
- Permissions validated on startup

**Verification:** Database file has secure permissions (600)

---

#### TC-75: SQL Injection Prevention
**Type:** Security  
**Priority:** High  
**Requirement:** SEC-002  
**Given:** User-provided input in queries  
**When:** Database queries executed  
**Then:**
- All queries use parameterized statements
- No string concatenation in queries
- SQL injection attempts blocked
- Query validation functional
- Safe handling of user input

**Verification:** SQL injection attempts prevented

---

#### TC-76: Path Traversal Prevention
**Type:** Security  
**Priority:** High  
**Requirement:** SEC-001, SEC-002  
**Given:** User-provided file paths  
**When:** Paths used in operations  
**Then:**
- Path traversal attempts blocked
- Relative paths resolved safely
- No access outside intended directories
- Path validation functional
- Safe path handling implemented

**Verification:** Path traversal attacks prevented

---

#### TC-77: Credential Storage Validation
**Type:** Security  
**Priority:** Medium  
**Requirement:** SEC-001  
**Given:** Configuration file contents  
**When:** File scanned for credentials  
**Then:**
- No passwords stored in config
- No API keys stored in config
- No tokens stored in config
- No sensitive data in config file
- Secure by default

**Verification:** No credentials stored in configuration files

---

#### TC-78: Database Network Exposure Check
**Type:** Security  
**Priority:** Medium  
**Requirement:** SEC-002  
**Given:** Database system  
**When:** Network exposure tested  
**Then:**
- No network exposure (local SQLite only)
- No remote access possible
- No network ports open
- Local-only access enforced
- Network security validated

**Verification:** Database has no network exposure

---

## Platform Compatibility Tests

### Unit Tests

#### TC-79: Cross-Platform Path Handling
**Type:** Platform  
**Priority:** Medium  
**Requirements:** CFG-002, DB-001  
**Given:** Different operating systems  
**When:** Paths processed  
**Then:**
- Windows paths handled correctly
- Unix paths handled correctly
- macOS paths handled correctly
- Path separators work on all platforms
- Cross-platform compatibility maintained

**Verification:** Path handling works on Windows, Linux, macOS

---

#### TC-80: File System Permissions Across Platforms
**Type:** Platform  
**Priority:** Medium  
**Requirements:** SEC-001, SEC-002  
**Given:** Different operating systems  
**When:** File permissions set  
**Then:**
- Unix permissions work correctly
- Windows permissions work correctly
- macOS permissions work correctly
- Platform-specific permission handling
- Consistent security across platforms

**Verification:** File permissions work correctly on all platforms

---

#### TC-81: Go Environment Compatibility
**Type:** Platform  
**Priority:** Low  
**Requirements:** GO-001, CLI-004  
**Given:** Different Go versions  
**When:** Application runs  
**Then:**
- Compatible with Go 1.21+
- Works with latest Go versions
- No version-specific issues
- Doctor command checks Go version
- Compatibility validated

**Verification:** Application works with supported Go versions

---

## Performance Tests

#### TC-82: Startup Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** PERF-001  
**Given:** Application startup  
**When:** Performance measured  
**Then:**
- Config loading < 100ms
- Database connection < 200ms
- Schema validation < 500ms
- CLI command response < 1s for status/doctor
- Total startup time acceptable

**Verification:** Startup performance meets timing requirements

---

#### TC-83: Configuration Loading Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** PERF-001  
**Given:** Configuration file  
**When:** Loading and validation  
**Then:**
- Config file parsing < 50ms
- Config validation < 50ms
- Total config load < 100ms
- Performance consistent

**Verification:** Config loading completes within 100ms

---

#### TC-84: Database Connection Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** PERF-001  
**Given:** Database file  
**When:** Connection established  
**Then:**
- Connection opening < 200ms
- WAL mode enablement < 50ms
- Connection validation < 100ms
- Total connection time < 200ms

**Verification:** Database connection completes within 200ms

---

#### TC-85: Logging Performance Under Load
**Type:** Performance  
**Priority:** Medium  
**Requirement:** LOG-001  
**Given:** High-volume logging  
**When:** 10,000 log entries written  
**Then:**
- Average time per entry < 1ms
- No blocking operations
- Buffer management effective
- No performance degradation

**Verification:** Logging maintains < 1ms per entry under load

---

#### TC-86: Database Operation Performance
**Type:** Performance  
**Priority:** Medium  
**Requirement:** PERF-002  
**Given:** Database with data  
**When:** Common operations performed  
**Then:**
- Single project insert < 10ms
- Query with 1000 projects < 100ms
- Schema migration < 5s per migration
- Index-based queries fast
- Performance targets met

**Verification:** Database operations meet performance targets

---

#### TC-87: Memory Usage Validation
**Type:** Performance  
**Priority:** Low  
**Requirements:** All stories  
**Given:** Application running  
**When:** Memory usage monitored  
**Then:**
- Baseline memory usage reasonable
- No memory leaks detected
- Buffer sizes within limits
- Memory growth controlled
- Performance stable over time

**Verification:** Memory usage remains within acceptable bounds

---

## Acceptance Criteria

### Story 1.1 Acceptance Criteria

#### AC-01: Go Module Initialized
**Type:** Functional  
**Verification Method:** `go mod verify` succeeds, `go build` completes without errors  
**Requirement Reference:** GO-001

#### AC-02: Standard Project Structure Present
**Type:** Functional  
**Verification Method:** All directories (cmd/, internal/, pkg/) exist and accessible  
**Requirement Reference:** GO-002

#### AC-03: Git Configuration Complete
**Type:** Functional  
**Verification Method:** .gitignore configured for Go, LICENSE file present  
**Requirement Reference:** GO-003

#### AC-04: Build Documentation Complete
**Type:** Functional  
**Verification Method:** README contains build/run instructions, new developer can build from README  
**Requirement Reference:** GO-004

---

### Story 1.2 Acceptance Criteria

#### AC-05: TOML Configuration Format Defined
**Type:** Functional  
**Verification Method:** Config file in TOML format, parses correctly, all sections accessible  
**Requirement Reference:** CFG-001

#### AC-06: Configuration Schema Complete
**Type:** Functional  
**Verification Method:** Schema includes project_roots, ignored_paths, database_path  
**Requirement Reference:** CFG-002

#### AC-07: Configuration Loading with Defaults Working
**Type:** Functional  
**Verification Method:** Missing config creates default file, partial config merges properly  
**Requirement Reference:** CFG-003

#### AC-08: Configuration Validation Functional
**Type:** Functional  
**Verification Method:** All validation rules enforceable, error messages actionable  
**Requirement Reference:** CFG-002

#### AC-09: Configuration Error Handling Comprehensive
**Type:** Functional  
**Verification Method:** All error scenarios produce actionable error messages  
**Requirement Reference:** CFG-005

---

### Story 1.3 Acceptance Criteria

#### AC-10: Structured Logging Implemented
**Type:** Functional  
**Verification Method:** Log entries parse as structured JSON, required fields present  
**Requirement Reference:** LOG-001

#### AC-11: Log Levels Working
**Type:** Functional  
**Verification Method:** DEBUG, INFO, WARN, ERROR levels produce appropriate verbosity  
**Requirement Reference:** LOG-002

#### AC-12: Standard Output Configuration Functional
**Type:** Functional  
**Verification Method:** Logs appear on stdout, format selection works  
**Requirement Reference:** LOG-003

#### AC-13: Environment Variable Configuration Working
**Type:** Functional  
**Verification Method:** PORTFOLIO_LOG_LEVEL overrides config, invalid values rejected  
**Requirement Reference:** LOG-004

---

### Story 1.4 Acceptance Criteria

#### AC-14: CLI Framework Implemented
**Type:** Functional  
**Verification Method:** portfolio --help displays usage, command structure works  
**Requirement Reference:** CLI-001

#### AC-15: Init Subcommand Functional
**Type:** Functional  
**Verification Method:** Command creates config file, database initializes, prompts work  
**Requirement Reference:** CLI-002

#### AC-16: Status Subcommand Working
**Type:** Functional  
**Verification Method:** Command shows accurate status, handles error states  
**Requirement Reference:** CLI-003

#### AC-17: Doctor Subcommand Functional
**Type:** Functional  
**Verification Method:** All diagnostic checks run, exit codes correct  
**Requirement Reference:** CLI-004

#### AC-18: Administrative Scope Maintained
**Type:** Constraint  
**Verification Method:** CLI commands limited to administrative operations only  
**Requirement Reference:** CLI-006

---

### Story 1.5 Acceptance Criteria

#### AC-19: Database File Created
**Type:** Functional  
**Verification Method:** Database file exists at configured location  
**Requirement Reference:** DB-001

#### AC-20: Connection Management Working
**Type:** Functional  
**Verification Method:** Connections open/close properly, concurrent access handled  
**Requirement Reference:** DB-002

#### AC-21: Schema Validation Functional
**Type:** Functional  
**Verification Method:** All 9 tables present, structure matches specification  
**Requirement Reference:** DB-003

#### AC-22: Migration System Implemented
**Type:** Functional  
**Verification Method:** Migrations apply correctly, version tracking works  
**Requirement Reference:** DB-004

#### AC-23: Complete Schema Created
**Type:** Functional  
**Verification Method:** All tables from PlatformSpecification.md created  
**Requirement Reference:** DB-003

---

## Coverage Gaps and Recommendations

### Current Coverage: ✅ Complete

**Test Coverage Analysis:**
- **Story 1.1 (Bootstrap):** 6 test cases cover all acceptance criteria
- **Story 1.2 (Configuration):** 11 test cases cover all acceptance criteria
- **Story 1.3 (Logging):** 11 test cases cover all acceptance criteria
- **Story 1.4 (CLI):** 14 test cases cover all acceptance criteria
- **Story 1.5 (Database):** 18 test cases cover all acceptance criteria
- **Integration:** 11 test cases cover all integration points
- **Security:** 6 test cases cover security requirements
- **Platform:** 3 test cases cover cross-platform compatibility
- **Performance:** 6 test cases cover performance requirements

### Coverage Strengths:
1. ✅ All acceptance criteria have corresponding test cases
2. ✅ Integration points between stories thoroughly tested
3. ✅ Error scenarios covered for all components
4. ✅ Performance requirements addressed
5. ✅ Security considerations included
6. ✅ Platform compatibility considered

### Recommended Additional Tests (Future Epics):
1. **Epic 2+ Discovery Tests:** Project scanning, metadata extraction
2. **MCP Integration Tests:** AI agent communication
3. **HTTP API Tests:** Dashboard backend endpoints
4. **Dashboard Tests:** Frontend functionality
5. **Migration Path Tests:** Version-to-version upgrade scenarios

### Test Implementation Priority:
**High Priority (Critical Path):** TC-01, TC-07, TC-18, TC-29, TC-30, TC-43, TC-46, TC-61  
**Medium Priority (Integration):** TC-13, TC-14, TC-25, TC-37, TC-54, TC-65  
**Standard Priority (Coverage):** All remaining test cases

---

## Test Implementation Notes

### Test Environment Setup:
- Temporary test directories for file operations
- In-memory test database for unit tests
- Mock configuration files for testing
- Test environment variables for logging tests

### Test Data Requirements:
- Sample TOML configuration files (valid, invalid, partial)
- Sample database schemas (various versions)
- Test project structures for path validation
- Error scenario test cases

### Test Execution Order:
1. Unit tests (fast, isolated)
2. Integration tests (component interaction)
3. Performance tests (require warmup)
4. E2E tests (require complete system)
5. Security tests (require special setup)

---

**End of Epic 1 - Project Foundation Test Cases**

**Total Test Cases:** 87  
**Total Acceptance Criteria:** 23  
**Coverage:** Complete for Epic 1 requirements  
**Next Steps:** Implement test cases and validate Epic 1 foundation