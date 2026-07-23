#!/bin/bash
# Code formatting check for Portfolio Engine

# Include helper functions
source "$(dirname "$0")/../common/helper.sh"

print_info "Checking code formatting with gofmt..."

UNFORMATTED_FILES=$(gofmt -l .)
if [ -n "$UNFORMATTED_FILES" ]; then
    print_error "Code is not properly formatted. Run 'gofmt -w .' to fix:"
    echo "$UNFORMATTED_FILES"
    exit 1
fi

print_info "Code formatting check passed."
exit 0