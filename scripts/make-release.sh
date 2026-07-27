#!/bin/bash
# Quick release shortcut - just provide version and go!
# This is a convenience wrapper around scripts/release.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Usage: ./make-release.sh v0.2.0

if [ $# -eq 0 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.2.0"
    exit 1
fi

# Run the full release script
exec "$SCRIPT_DIR/release.sh" "$@"