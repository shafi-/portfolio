#!/bin/bash

# Epic 9 — Claude Code Integration: Quality Gate Verification Script
# Simplified version

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counter for pass/fail
PASSED=0
FAILED=0

# Helper functions
print_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAILED++))
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Run check with output
run_check() {
    local name=$1
    local command=$2
    
    print_info "Checking: $name"
    
    if eval "$command" >/dev/null 2>&1; then
        print_pass "$name"
        return 0
    else
        print_fail "$name"
        return 1
    fi
}

# Main verification
echo "======================================="
echo "Epic 9 Quality Gate Verification"
echo "======================================="
echo ""

# Check prerequisites
print_info "Checking prerequisites..."
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}✗${NC} Go is not installed"
    exit 1
fi
print_pass "Go is installed"

if ! command -v bc >/dev/null 2>&1; then
    echo -e "${RED}✗${NC} bc is not installed"
    exit 1
fi
print_pass "bc is installed"

echo ""

# Priority 1: CLI Integration - Wire commands
print_info "Priority 1: CLI Integration - Wire commands"
echo ""

# Check if all commands are wired
run_check "CLI: install command exists" "./portfolio --help | grep -q 'install'"
run_check "CLI: upgrade command exists" "./portfolio --help | grep -q 'upgrade'"
run_check "CLI: uninstall command exists" "./portfolio --help | grep -q 'uninstall'"
run_check "CLI: doctor command exists" "./portfolio --help | grep -q 'doctor'"

# Check if claude subcommands are wired
run_check "CLI: install claude subcommand exists" "./portfolio install --help | grep -q 'claude'"
run_check "CLI: upgrade claude subcommand exists" "./portfolio upgrade --help | grep -q 'claude'"
run_check "CLI: uninstall claude subcommand exists" "./portfolio uninstall --help | grep -q 'claude'"
run_check "CLI: doctor claude subcommand exists" "./portfolio doctor --help | grep -q 'claude'"

# Check help text completeness
run_check "CLI: install claude has help text" "./portfolio install claude --help | grep -q 'Install Claude Code integration'"
run_check "CLI: upgrade claude has help text" "./portfolio upgrade claude --help | grep -q 'Upgrade Claude Code integration'"
run_check "CLI: uninstall claude has help text" "./portfolio uninstall claude --help | grep -q 'Uninstall Claude Code integration'"
run_check "CLI: doctor claude has help text" "./portfolio doctor claude --help | grep -q 'Check Claude Code integration health'"

echo ""

# Priority 2: Integration Tests - Full lifecycle
print_info "Priority 2: Integration Tests - Full lifecycle"
echo ""

# Check if test files exist
run_check "Integration: lifecycle test exists" "test -f ./internal/integration/claude/lifecycle_test.go"
run_check "Integration: mcp_integration test exists" "test -f ./internal/integration/claude/mcp_integration_test.go"
run_check "Integration: test_sandbox exists" "test -f ./internal/integration/claude/test_sandbox.go"
run_check "Integration: error_scenarios test exists" "test -f ./internal/integration/claude/error_scenarios_test.go"

# Run integration tests
print_info "Running integration tests..."
if go test -v -run TestLifecycle ./internal/integration/claude/... >/tmp/lifecycle_tests.log 2>&1; then
    print_pass "Lifecycle integration tests pass"
else
    print_fail "Lifecycle integration tests failed"
    echo "  See /tmp/lifecycle_tests.log for details"
fi

if go test -v -run TestMCPServer ./internal/integration/claude/... >/tmp/mcp_tests.log 2>&1; then
    print_pass "MCP server integration tests pass"
else
    print_fail "MCP server integration tests failed"
    echo "  See /tmp/mcp_tests.log for details"
fi

if go test -v -run TestError ./internal/integration/claude/... >/tmp/error_tests.log 2>&1; then
    print_pass "Error scenario tests pass"
else
    print_fail "Error scenario tests failed"
    echo "  See /tmp/error_tests.log for details"
fi

echo ""

# Priority 3: Test Coverage improvements
print_info "Priority 3: Test Coverage improvements"
echo ""

# Generate coverage report
print_info "Generating coverage report..."
go test -coverprofile=/tmp/coverage.out ./... >/dev/null 2>&1

# Check overall coverage
overall_coverage=$(go tool cover -func=/tmp/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
print_pass "Overall coverage: ${overall_coverage}%"

# Check integration package coverage
integration_coverage=$(go tool cover -func=/tmp/coverage.out | grep "internal/integration/claude/" | tail -1 | awk '{print $3}' | sed 's/%//')
print_pass "Integration package coverage: ${integration_coverage}%"

# Check CLI package coverage
cli_coverage=$(go tool cover -func=/tmp/coverage.out | grep "internal/cli/" | tail -1 | awk '{print $3}' | sed 's/%//')
print_pass "CLI package coverage: ${cli_coverage}%"

# Verify coverage targets
if (( $(echo "$overall_coverage >= 80" | bc -l) )); then
    print_pass "Overall coverage meets target (≥80%)"
else
    print_fail "Overall coverage below target (≥80%)"
fi

if (( $(echo "$integration_coverage >= 85" | bc -l) )); then
    print_pass "Integration package coverage meets target (≥85%)"
else
    print_fail "Integration package coverage below target (≥85%)"
fi

if (( $(echo "$cli_coverage >= 75" | bc -l) )); then
    print_pass "CLI package coverage meets target (≥75%)"
else
    print_fail "CLI package coverage below target (≥75%)"
fi

# Generate HTML coverage report
go tool cover -html=/tmp/coverage.out -o /tmp/coverage.html
print_pass "HTML coverage report generated: /tmp/coverage.html"

echo ""

# Priority 4: Quality gate verification
print_info "Priority 4: Quality gate verification"
echo ""

# Check MCP tools availability
run_check "Quality: Database store exists" "test -f ./internal/integration/database_store.go"
run_check "Quality: Integration manager exists" "test -f ./internal/integration/manager.go"

# Check error handling patterns
run_check "Quality: Error types defined" "test -f ./internal/integration/errors.go"
run_check "Quality: Integration interface defined" "test -f ./internal/integration/interface.go"

# Final summary
echo ""
echo "======================================="
echo "Verification Summary"
echo "======================================="
echo ""
echo -e "Passed: ${GREEN}${PASSED}${NC}"
echo -e "Failed: ${RED}${FAILED}${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All quality gates passed!"
    exit 0
else
    echo -e "${RED}✗${NC} Some quality gates failed. Please review the failures above."
    exit 1
fi