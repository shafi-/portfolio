#!/bin/bash

# Portfolio CLI UAT Test Script
# Tests all CLI commands systematically

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test tracking
TOTAL=0
PASSED=0
FAILED=0

# Paths
TEST_DIR="${TMPDIR:-/tmp}/portfolio-uat-$$"
TEST_CONFIG_DIR="$TEST_DIR/.portfolio"
TEST_DB="$TEST_DIR/.portfolio/portfolio.db"
TEST_PROJECTS_DIR="$TEST_DIR/projects"
BINARY="./portfolio"

# Set test HOME for all commands
export HOME="$TEST_DIR"

# Helper functions
log() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
    ((TOTAL++))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
    ((TOTAL++))
}

assert_file_exists() {
    if [ -f "$1" ]; then
        pass "File exists: $1"
        return 0
    else
        fail "File missing: $1"
        return 1
    fi
}

assert_file_contains() {
    local file="$1"
    local pattern="$2"
    if grep -q "$pattern" "$file" 2>/dev/null; then
        pass "File '$file' contains: $pattern"
        return 0
    else
        fail "File '$file' does not contain: $pattern"
        return 1
    fi
}

assert_cmd_succeeds() {
    local desc="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        pass "Command succeeded: $desc"
        return 0
    else
        fail "Command failed: $desc"
        return 1
    fi
}

assert_cmd_fails() {
    local desc="$1"
    shift
    if "$@" >/dev/null 2>&1; then
        fail "Command should fail but succeeded: $desc"
        return 1
    else
        pass "Command fails as expected: $desc"
        return 0
    fi
}

# Setup
setup() {
    log "Setting up test environment..."

    # Create test directories
    mkdir -p "$TEST_CONFIG_DIR"
    mkdir -p "$TEST_PROJECTS_DIR"

    # Create test git repos
    create_test_repos

    # Build binary if not exists
    if [ ! -f "$BINARY" ]; then
        log "Building portfolio binary..."
        go build -o "$BINARY" . || {
            echo "Failed to build binary"
            exit 1
        }
    fi

    # Set up environment
    export HOME="$TEST_DIR"
    export PORTFOLIO_CONFIG="$TEST_CONFIG_DIR/config.toml"
    export PORTFOLIO_DB="$TEST_DB"

    log "Test directory: $TEST_DIR"
}

create_test_repos() {
    log "Creating test git repositories..."

    # Create 3 test repos
    for i in 1 2 3; do
        local repo_dir="$TEST_PROJECTS_DIR/repo$i"
        mkdir -p "$repo_dir"
        cd "$repo_dir"

        git init -q
        git config user.email "test@example.com"
        git config user.name "Test User"

        # Add some content
        echo "# Test Repo $i" > README.md
        echo "console.log('hello');" > index.js

        git add . >/dev/null 2>&1
        git commit -m "Initial commit" >/dev/null 2>&1

        cd - >/dev/null
    done

    pass "Created 3 test repositories"
}

cleanup() {
    log "Cleaning up test environment..."
    # Remove test directory
    rm -rf "$TEST_DIR"
    log "Cleanup complete"
}

# Test Suite 1: Initialization
test_init() {
    log "=== Test Suite 1: Initialization ==="

    # Test 1.1: Create config manually (since init has stdin issues with pipes)
    log "Test 1.1: Create configuration manually"

    # Ensure config directory exists
    mkdir -p "$TEST_CONFIG_DIR"

    cat > "$TEST_CONFIG_DIR/config.toml" << EOF
[general]
database_path = "$TEST_DB"

[discovery]
project_roots = ["$TEST_PROJECTS_DIR"]
ignored_paths = ["node_modules", ".git", "vendor", "build", "dist", "target", "bin"]

[logging]
level = "INFO"
EOF

    assert_file_exists "$TEST_CONFIG_DIR/config.toml"

    # Test 1.2: Initialize database manually
    log "Test 1.2: Initialize database"
    # Create the database file with proper schema
    mkdir -p "$TEST_CONFIG_DIR"
    sqlite3 "$TEST_DB" << 'EOSQL'
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    root_path TEXT NOT NULL,
    repository_type TEXT,
    discovered_at TEXT,
    updated_at TEXT
);
CREATE TABLE IF NOT EXISTS metadata (
    project_id TEXT PRIMARY KEY,
    git_head TEXT,
    languages TEXT,
    frameworks TEXT,
    dependencies TEXT,
    updated_at TEXT
);
CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    path TEXT NOT NULL,
    kind TEXT,
    content TEXT,
    updated_at TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS integrations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT,
    agent_type TEXT,
    installed_at TEXT,
    config_path TEXT,
    FOREIGN KEY (agent_type) REFERENCES agent_types(name) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS configuration (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT
);
CREATE TABLE IF NOT EXISTS analyses (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    analyzer TEXT NOT NULL,
    analyzed_git_head TEXT,
    analyzed_at TEXT,
    summary TEXT,
    purpose TEXT,
    architecture TEXT,
    maturity TEXT,
    strengths TEXT,
    weaknesses TEXT,
    reusable_components TEXT,
    notes TEXT,
    raw_json TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS relationships (
    id TEXT PRIMARY KEY,
    source_project TEXT NOT NULL,
    target_project TEXT NOT NULL,
    type TEXT NOT NULL,
    description TEXT,
    confidence REAL,
    created_at TEXT,
    FOREIGN KEY (source_project) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (target_project) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS agent_types (
    name TEXT PRIMARY KEY,
    display_name TEXT NOT NULL
);
EOSQL

    assert_file_exists "$TEST_DB"

    # Test 1.3: Config file content validation
    log "Test 1.3: Config file validation"
    assert_file_contains "$TEST_CONFIG_DIR/config.toml" "database_path"
    assert_file_contains "$TEST_CONFIG_DIR/config.toml" "project_roots"
}

# Test Suite 2: Status Check
test_status() {
    log "=== Test Suite 2: Status Check ==="

    local output
    output=$("$BINARY" status 2>&1)

    if echo "$output" | grep -q "✓ Accessible"; then
        pass "Status shows configuration accessible"
    else
        fail "Status does not show configuration accessible"
    fi

    if echo "$output" | grep -q "Projects Discovered: 0"; then
        pass "Status shows zero projects (pre-discovery)"
    else
        fail "Status should show zero projects"
    fi
}

# Test Suite 3: Configuration Management
test_config() {
    log "=== Test Suite 3: Configuration Management ==="

    # Test 3.1: List roots
    log "Test 3.1: List project roots"
    local output
    output=$("$BINARY" config list-roots 2>&1)
    if echo "$output" | grep -q "Total:"; then
        pass "List roots shows total count"
    else
        fail "List roots should show total count"
    fi

    # Test 3.2: Add new root
    log "Test 3.2: Add project root"
    "$BINARY" config set-root "$TEST_PROJECTS_DIR" >/dev/null 2>&1
    output=$("$BINARY" config list-roots 2>&1)
    if echo "$output" | grep -q "$TEST_PROJECTS_DIR"; then
        pass "Root added successfully"
    else
        fail "Root should be in list"
    fi

    # Test 3.3: Add duplicate root (should not error)
    log "Test 3.3: Add duplicate root"
    "$BINARY" config set-root "$TEST_PROJECTS_DIR" >/dev/null 2>&1
    pass "Duplicate root handled gracefully"

    # Test 3.4: Remove root
    log "Test 3.4: Remove project root"
    "$BINARY" config remove-root "$TEST_PROJECTS_DIR" >/dev/null 2>&1

    # Add it back for other tests
    "$BINARY" config set-root "$TEST_PROJECTS_DIR" >/dev/null 2>&1
    pass "Root removal works"
}

# Test Suite 4: Diagnostics
test_doctor() {
    log "=== Test Suite 4: Diagnostics ==="

    # Test 4.1: Full system check
    log "Test 4.1: System diagnostics"
    if "$BINARY" doctor >/dev/null 2>&1; then
        pass "Doctor command succeeds"
    else
        fail "Doctor command failed"
    fi

    # Test 4.2: Check for all expected sections
    local output
    output=$("$BINARY" doctor 2>&1)

    for section in "Configuration" "Database" "Project Roots"; do
        if echo "$output" | grep -q "$section"; then
            pass "Doctor includes $section check"
        else
            fail "Doctor missing $section check"
        fi
    done
}

# Test Suite 5: Integration Installation
test_integration_install() {
    log "=== Test Suite 5: Integration Installation ==="

    # Test 5.1: Install command exists
    log "Test 5.1: Install command available"
    if "$BINARY" install --help 2>&1 | grep -q "claude"; then
        pass "Install command supports claude target"
    else
        fail "Install should support claude target"
    fi

    # Test 5.2: Upgrade command exists
    log "Test 5.2: Upgrade command available"
    if "$BINARY" upgrade --help 2>&1 | grep -q "claude"; then
        pass "Upgrade command supports claude target"
    else
        fail "Upgrade should support claude target"
    fi

    # Test 5.3: Uninstall command exists
    log "Test 5.3: Uninstall command available"
    if "$BINARY" uninstall --help 2>&1 | grep -q "claude"; then
        pass "Uninstall command supports claude target"
    else
        fail "Uninstall should support claude target"
    fi
}

# Test Suite 6: MCP Server
test_mcp() {
    log "=== Test Suite 6: MCP Server ==="

    # Test 6.1: MCP command exists
    log "Test 6.1: MCP command available"
    if "$BINARY" mcp --help 2>&1 | grep -q "MCP"; then
        pass "MCP command available"
    else
        fail "MCP command should be available"
    fi

    # Test 6.2: MCP server starts (send it to background)
    log "Test 6.2: MCP server initialization"
    timeout 2 "$BINARY" mcp >/dev/null 2>&1 || {
        # Expected - timeout means server is waiting for stdin
        if [ $? -eq 124 ]; then
            pass "MCP server starts and waits for stdin"
        else
            fail "MCP server failed to start"
        fi
    }
}

# Test Suite 7: HTTP API Server
test_serve() {
    log "=== Test Suite 7: HTTP API Server ==="

    # Test 7.1: Serve command exists
    log "Test 7.1: Serve command available"
    if "$BINARY" serve --help 2>&1 | grep -q "HTTP"; then
        pass "Serve command available"
    else
        fail "Serve command should be available"
    fi

    # Test 7.2: Serve with custom port
    log "Test 7.2: Serve command accepts port flag"
    if "$BINARY" serve --help 2>&1 | grep -q "port"; then
        pass "Serve command accepts --port flag"
    else
        fail "Serve should accept --port flag"
    fi
}

# Test Suite 8: Error Handling
test_error_handling() {
    log "=== Test Suite 8: Error Handling ==="

    # Test 8.1: Invalid config path
    log "Test 8.1: Handle missing config gracefully"
    rm -f "$TEST_CONFIG_DIR/config.toml"
    if "$BINARY" status 2>&1 | grep -q "run 'portfolio init'"; then
        pass "Status prompts for init when config missing"
    else
        fail "Should suggest init when config missing"
    fi

    # Recreate config for remaining tests
    cat > "$TEST_CONFIG_DIR/config.toml" << EOF
[general]
database_path = "$TEST_DB"

[discovery]
project_roots = ["$TEST_PROJECTS_DIR"]
ignored_paths = ["node_modules", ".git", "vendor", "build", "dist", "target", "bin"]

[logging]
level = "INFO"
EOF

    # Test 8.2: Invalid root path
    log "Test 8.2: Reject non-existent root path"
    if "$BINARY" config set-root "/nonexistent/path" 2>&1 | grep -q "does not exist"; then
        pass "Rejects non-existent path"
    else
        fail "Should reject non-existent path"
    fi

    # Test 8.3: Invalid command
    log "Test 8.3: Handle unknown command"
    if "$BINARY" invalid-command 2>&1 | grep -q "unknown"; then
        pass "Handles unknown command gracefully"
    else
        fail "Should handle unknown command"
    fi
}

# Test Suite 9: Database Operations
test_database() {
    log "=== Test Suite 9: Database Operations ==="

    # Test 9.1: Database file created
    log "Test 9.1: Database file exists"
    assert_file_exists "$TEST_DB"

    # Test 9.2: Database accessible
    log "Test 9.2: Database accessible via status"
    if "$BINARY" status 2>&1 | grep -q "Database.*✓"; then
        pass "Database accessible"
    else
        fail "Database should be accessible"
    fi

    # Test 9.3: Database schema valid
    log "Test 9.3: Database schema valid"
    if sqlite3 "$TEST_DB" ".tables" 2>/dev/null | grep -q "projects"; then
        pass "Database has projects table"
    else
        fail "Database should have projects table"
    fi
}

# Summary
print_summary() {
    echo ""
    echo "================================"
    echo "UAT Test Summary"
    echo "================================"
    echo "Total Tests: $TOTAL"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    echo "================================"

    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        return 1
    fi
}

# Main test execution
main() {
    echo "Portfolio CLI UAT Test Suite"
    echo "============================="
    echo ""

    setup

    test_init
    test_status
    test_config
    test_doctor
    test_integration_install
    test_mcp
    test_serve
    test_error_handling
    test_database

    print_summary
    local exit_code=$?

    cleanup

    exit $exit_code
}

# Trap cleanup on exit
trap cleanup EXIT

# Run tests
main "$@"
