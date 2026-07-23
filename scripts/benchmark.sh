#!/bin/bash
# Local benchmark runner - use this instead of CI workflow
# Usage: ./scripts/benchmark.sh [quick|full]
# Default: quick (1s per benchmark)

set -e

BENCHTIME="${1:-quick}"

case "$BENCHTIME" in
  quick)
    TIME="1s"
    echo "Running quick benchmarks (1s each)..."
    ;;
  full)
    TIME="10s"
    echo "Running full benchmarks (10s each)..."
    ;;
  *)
    echo "Usage: $0 [quick|full]"
    exit 1
    ;;
esac

echo "======================================="
echo "Benchmark Mode: $BENCHTIME"
echo "======================================="

go test -bench=. -benchmem -benchtime="$TIME" ./...

echo ""
echo "✓ Benchmarks complete"
