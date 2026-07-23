#!/bin/bash
# Dependency check for Portfolio Engine

# Include helper functions
source "$(dirname "$0")/../common/helper.sh"

print_info "Checking dependencies and module health..."

# Run go mod tidy to check for unused/missing dependencies
if ! go mod tidy; then
    print_error "go mod tidy failed - check dependencies"
    exit 1
fi

print_info "Dependency check passed."
cleanup_coverage
exit 0