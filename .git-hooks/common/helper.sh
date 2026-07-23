#!/bin/bash
# Common helper functions for Portfolio Engine git hooks

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to run check and handle result
run_check() {
    local check_name="$1"
    local check_command="$2"
    local quiet="$3"  # "quiet" or "verbose"

    print_info "Running $check_name..."
    if [ "$quiet" = "quiet" ]; then
        if ! eval "$check_command"; then
            print_error "$check_name failed"
            return 1
        fi
    else
        if ! eval "$check_command"; then
            return 1
        fi
    fi
    return 0
}

# Function to cleanup coverage file
cleanup_coverage() {
    if [ -f "coverage.out" ]; then
        print_info "Code coverage summary:"
        go tool cover -func=coverage.out | tail -n 1
        rm -f coverage.out
    fi
}

# Function to check Go version
check_go_version() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        return 1
    fi

    local GO_VERSION=$(go version | sed 's/go version go\//')
    print_info "Go version: $GO_VERSION"
    return 0
}

# Function to check for vendor directory
check_vendor() {
    if [ -d "vendor/" ] && ! grep -q "vendor/" .gitignore 2>/dev/null; then
        print_warning "vendor/ directory exists but not in .gitignore"
        print_warning "Consider adding 'vendor/' to .gitignore to avoid committing dependencies"
    fi
}