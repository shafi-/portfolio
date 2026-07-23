#!/bin/bash
# Testing check for Portfolio Engine

# Include helper functions
source "$(dirname "$0")/../common/helper.sh"

print_info "Running tests..."

# Set environment variables for tests
export GOFLAGS="-mod=mod"

# Run tests for all packages
if ! go test -race -coverprofile=coverage.out ./...; then
    print_error "Tests failed"
    exit 1
fi

print_info "Testing check passed."
cleanup_coverage
exit 0