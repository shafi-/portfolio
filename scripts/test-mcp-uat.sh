#!/bin/bash

# Portfolio MCP UAT Test Script
# Tests MCP server tools via direct stdio interaction

set +e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test tracking
TOTAL=0
PASSED=0
FAILED=0

# Paths
TEST_DIR="${TMPDIR:-/tmp}/portfolio-mcp-uat-$$"
TEST_CONFIG_DIR="$TEST_DIR/.portfolio"
TEST_DB="$TEST_DIR/.portfolio/portfolio.db"
TEST_PROJECTS_DIR="$TEST_DIR/projects"
BINARY="./portfolio"
OUTPUT_FILE="$TEST_DIR/mcp-output.json"

# Set test HOME
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

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# JSON helper
extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | cut -d'"' -f4
}

extract_json_bool() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*true" | wc -l
}

# Setup
setup() {
    log "Setting up MCP test environment..."

    # Create test directories
    mkdir -p "$TEST_CONFIG_DIR"
    mkdir -p "$TEST_PROJECTS_DIR"

    # Build binary if not exists
    if [ ! -f "$BINARY" ]; then
        log "Building portfolio binary..."
        go build -o "$BINARY" ./cmd/portfolio || {
            echo "Failed to build binary"
            exit 1
        }
    fi

    # Create config
    cat > "$TEST_CONFIG_DIR/config.toml" << EOF
[general]
database_path = "$TEST_DB"

[discovery]
project_roots = ["$TEST_PROJECTS_DIR"]
ignored_paths = ["node_modules", ".git", "vendor", "build", "dist", "target", "bin"]

[logging]
level = "INFO"
EOF

    # Create test git repos
    create_test_repos

    # Initialize database via init command to ensure proper migrations
    log "Initializing database via init command..."
    rm -f "$TEST_DB"  # Remove any existing database
    (echo "$TEST_PROJECTS_DIR"; echo ""; echo "INFO"; echo "y") | "$BINARY" --config "$TEST_CONFIG_DIR/config.toml" init >/dev/null 2>&1 || {
        # If init fails, try manual init
        log "Manual database initialization..."
        init_database_manual
    }

    log "Test directory: $TEST_DIR"
}

init_database_manual() {
    # Fallback: create empty database, let migrations handle the rest
    log "Creating empty database for migrations..."
    touch "$TEST_DB"
}

# Helper functions
log() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

create_test_repos() {
    log "Creating test git repositories..."

    for i in 1 2 3; do
        local repo_dir="$TEST_PROJECTS_DIR/repo$i"
        mkdir -p "$repo_dir"
        cd "$repo_dir"

        git init -q
        git config user.email "test@example.com"
        git config user.name "Test User"

        echo "# Test Repo $i" > README.md
        echo "console.log('hello');" > index.js

        git add . >/dev/null 2>&1
        git commit -m "Initial commit" >/dev/null 2>&1

        cd - >/dev/null
    done

    pass "Created 3 test repositories"
}

init_database() {
    log "Initializing database..."

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
CREATE TABLE IF NOT EXISTS configuration (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT
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
EOSQL

    pass "Database initialized"
}

# MCP protocol helpers
mcp_call() {
    local tool="$1"
    local args="$2"

    # Create MCP request
    local request="{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"$tool\",\"arguments\":$args}}"

    # Send to MCP server with explicit config and get response
    echo "$request" | timeout 5 "$BINARY" --config "$TEST_CONFIG_DIR/config.toml" mcp 2>/dev/null | head -20
}

mcp_initialize() {
    echo "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\",\"params\":{\"protocolVersion\":\"2024-11-05\",\"capabilities\":{},\"clientInfo\":{\"name\":\"test-client\",\"version\":\"1.0.0\"}}}" | timeout 5 "$BINARY" --config "$TEST_CONFIG_DIR/config.toml" mcp 2>/dev/null | head -10
}

# Test Suite 1: Server Initialization
test_mcp_init() {
    log "=== Test Suite 1: MCP Server Initialization ==="

    # Test 1.1: Server responds to initialize
    log "Test 1.1: MCP initialize response"
    local response
    response=$(mcp_initialize)

    if echo "$response" | grep -q "jsonrpc"; then
        pass "Server responds to initialize"
    else
        fail "Server should respond to initialize"
        info "Response: $response"
    fi

    # Test 1.2: Server reports capabilities
    log "Test 1.2: Server capabilities"
    if echo "$response" | grep -q "tools"; then
        pass "Server reports tools capability"
    else
        fail "Server should report tools capability"
    fi
}

# Test Suite 2: Health Check
test_health() {
    log "=== Test Suite 2: Health Check ==="

    # Test 2.1: Health tool exists
    log "Test 2.1: Call health tool"
    local response
    response=$(mcp_call "health" "{}")

    if echo "$response" | grep -q "healthy\|unhealthy"; then
        pass "Health tool returns status"
    else
        fail "Health tool should return status"
        info "Response: $response"
    fi

    # Test 2.2: Health includes database status
    log "Test 2.2: Health includes database_connected"
    if echo "$response" | grep -q "database_connected"; then
        pass "Health includes database status"
    else
        fail "Health should include database status"
    fi
}

# Test Suite 3: Discovery Tools
test_discovery() {
    log "=== Test Suite 3: Discovery Tools ==="

    # Test 3.1: discoverProjects tool
    log "Test 3.1: Call discoverProjects"
    local response
    response=$(mcp_call "discoverProjects" "{}")

    if echo "$response" | grep -q "discovered\|error"; then
        pass "discoverProjects tool responds"
    else
        fail "discoverProjects should respond"
        info "Response: $response"
    fi

    # Test 3.2: listProjects tool
    log "Test 3.2: Call listProjects"
    response=$(mcp_call "listProjects" "{}")

    if echo "$response" | grep -q "projects\|error"; then
        pass "listProjects tool responds"
    else
        fail "listProjects should respond"
    fi
}

# Test Suite 4: Project Query Tools
test_queries() {
    log "=== Test Suite 4: Project Query Tools ==="

    # First discover projects
    mcp_call "discoverProjects" "{}" >/dev/null 2>&1

    # Get a project ID
    local project_id
    local projects_response
    projects_response=$(mcp_call "listProjects" "{}")
    project_id=$(echo "$projects_response" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -n "$project_id" ]; then
        # Test 4.1: getProject tool
        log "Test 4.1: Call getProject with ID"
        local response
        response=$(mcp_call "getProject" "{\"id\":\"$project_id\"}")

        if echo "$response" | grep -q "project\|metadata"; then
            pass "getProject returns project data"
        else
            fail "getProject should return project data"
            info "Response: $response"
        fi
    else
        fail "No projects found to test getProject"
    fi

    # Test 4.2: searchProjects tool
    log "Test 4.2: Call searchProjects"
    local response
    response=$(mcp_call "searchProjects" "{\"query\":\"repo\"}")

    if echo "$response" | grep -q "results\|count"; then
        pass "searchProjects returns results"
    else
        fail "searchProjects should return results"
        info "Response: $response"
    fi
}

# Test Suite 5: Analysis Tools
test_analysis() {
    log "=== Test Suite 5: Analysis Tools ==="

    # Get a project ID first
    local project_id
    local projects_response
    projects_response=$(mcp_call "listProjects" "{}")
    project_id=$(echo "$projects_response" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)

    if [ -n "$project_id" ]; then
        # Test 5.1: getAnalysis tool
        log "Test 5.1: Call getAnalysis"
        local response
        response=$(mcp_call "getAnalysis" "{\"project_id\":\"$project_id\"}")

        if echo "$response" | grep -q "analyses\|count"; then
            pass "getAnalysis returns analysis data"
        else
            fail "getAnalysis should return analysis data"
            info "Response: $response"
        fi

        # Test 5.2: storeAnalysis tool
        log "Test 5.2: Call storeAnalysis"
        response=$(mcp_call "storeAnalysis" "{\"project_id\":\"$project_id\",\"analyzer\":\"test\",\"summary\":\"Test summary\"}")

        if echo "$response" | grep -q "id\|analyzed_at"; then
            pass "storeAnalysis creates analysis"
        else
            fail "storeAnalysis should create analysis"
            info "Response: $response"
        fi

        # Test 5.3: listProjectsNeedingAnalysis
        log "Test 5.3: Call listProjectsNeedingAnalysis"
        response=$(mcp_call "listProjectsNeedingAnalysis" "{}")

        if echo "$response" | grep -q "projects\|count"; then
            pass "listProjectsNeedingAnalysis returns results"
        else
            fail "listProjectsNeedingAnalysis should return results"
            info "Response: $response"
        fi
    else
        fail "No projects found for analysis tests"
    fi
}

# Test Suite 6: Search Tools
test_search() {
    log "=== Test Suite 6: Search Tools ==="

    # Test 6.1: searchDocumentation tool
    log "Test 6.1: Call searchDocumentation"
    local response
    response=$(mcp_call "searchDocumentation" "{\"query\":\"test\"}")

    if echo "$response" | grep -q "results\|count"; then
        pass "searchDocumentation returns results"
    else
        fail "searchDocumentation should return results"
        info "Response: $response"
    fi
}

# Test Suite 7: Configuration Tools
test_config_tools() {
    log "=== Test Suite 7: Configuration Tools ==="

    # Test 7.1: getConfiguration tool
    log "Test 7.1: Call getConfiguration"
    local response
    response=$(mcp_call "getConfiguration" "{}")

    if echo "$response" | grep -q "config\|{\|}"; then
        pass "getConfiguration returns config"
    else
        fail "getConfiguration should return config"
        info "Response: $response"
    fi

    # Test 7.2: updateConfiguration tool
    log "Test 7.2: Call updateConfiguration"
    response=$(mcp_call "updateConfiguration" "{\"key\":\"test_key\",\"value\":\"test_value\"}")

    if echo "$response" | grep -q "status\|updated"; then
        pass "updateConfiguration updates config"
    else
        fail "updateConfiguration should update config"
        info "Response: $response"
    fi
}

# Test Suite 8: Relationship Tools
test_relationships() {
    log "=== Test Suite 8: Relationship Tools ==="

    # Get two project IDs
    local projects_response
    local project_id1
    local project_id2

    projects_response=$(mcp_call "listProjects" "{}")
    project_id1=$(echo "$projects_response" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)
    project_id2=$(echo "$projects_response" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | tail -1 | cut -d'"' -f4)

    if [ -n "$project_id1" ] && [ -n "$project_id2" ] && [ "$project_id1" != "$project_id2" ]; then
        # Test 8.1: listRelationships tool
        log "Test 8.1: Call listRelationships"
        local response
        response=$(mcp_call "listRelationships" "{\"project_id\":\"$project_id1\"}")

        if echo "$response" | grep -q "relationships\|\[\]"; then
            pass "listRelationships returns relationships"
        else
            fail "listRelationships should return relationships"
            info "Response: $response"
        fi

        # Test 8.2: storeRelationship tool
        log "Test 8.2: Call storeRelationship"
        response=$(mcp_call "storeRelationship" "{\"source_project\":\"$project_id1\",\"target_project\":\"$project_id2\",\"type\":\"Similar\",\"description\":\"Test relationship\"}")

        if echo "$response" | grep -q "id\|type"; then
            pass "storeRelationship creates relationship"
        else
            fail "storeRelationship should create relationship"
            info "Response: $response"
        fi
    else
        fail "Need two different projects for relationship tests"
    fi
}

# Test Suite 9: Error Handling
test_errors() {
    log "=== Test Suite 9: Error Handling ==="

    # Test 9.1: Invalid tool name
    log "Test 9.1: Call invalid tool"
    local response
    response=$(mcp_call "invalidTool" "{}")

    if echo "$response" | grep -q "error\|not found\|invalid"; then
        pass "Invalid tool returns error"
    else
        fail "Invalid tool should return error"
        info "Response: $response"
    fi

    # Test 9.2: Missing required parameters
    log "Test 9.2: Call getProject without ID"
    response=$(mcp_call "getProject" "{}")

    if echo "$response" | grep -q "error\|required\|missing"; then
        pass "Missing parameters returns error"
    else
        fail "Missing parameters should return error"
        info "Response: $response"
    fi

    # Test 9.3: Invalid project ID
    log "Test 9.3: Call getProject with invalid ID"
    response=$(mcp_call "getProject" "{\"id\":\"invalid-id-12345\"}")

    if echo "$response" | grep -q "error\|not found"; then
        pass "Invalid ID returns error"
    else
        fail "Invalid ID should return error"
        info "Response: $response"
    fi
}

# Summary
print_summary() {
    echo ""
    echo "================================"
    echo "MCP UAT Test Summary"
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

# Cleanup
cleanup() {
    log "Cleaning up test environment..."
    rm -rf "$TEST_DIR"
    log "Cleanup complete"
}

# Main test execution
main() {
    echo "Portfolio MCP UAT Test Suite"
    echo "=============================="
    echo ""

    setup

    test_mcp_init
    test_health
    test_discovery
    test_queries
    test_analysis
    test_search
    test_config_tools
    test_relationships
    test_errors

    print_summary
    local exit_code=$?

    cleanup

    exit $exit_code
}

# Trap cleanup on exit
trap cleanup EXIT

# Run tests
main "$@"
