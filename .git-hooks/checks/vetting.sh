#!/bin/bash
# Static analysis check for Portfolio Engine

# Include helper functions
source "$(dirname "$0")/../common/helper.sh"

print_info "Running go vet for static analysis..."

if ! go vet ./...; then
    print_error "go vet failed"
    exit 1
fi

print_info "Static analysis check passed."
exit 0