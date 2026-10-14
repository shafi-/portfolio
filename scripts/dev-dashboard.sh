#!/bin/bash

set -e

echo "Starting dashboard development environment..."

# Check if Vite dev server is already running
if lsof -ti:5173 > /dev/null 2>&1; then
    echo "Vite dev server already running on port 5173"
else
    # Start Vite dev server in background
    cd dashboard
    npm run dev &
    VITE_PID=$!
    cd ..
    echo "Vite dev server started (PID: $VITE_PID)"
fi

# Start Portfolio dashboard server in background
# This serves both API and dashboard assets in dev mode
go run ./cmd/portfolio dashboard --dev --dist ./dashboard/dist &
DASHBOARD_PID=$!

echo "Development environment started!"
echo "Dashboard: http://localhost:3000"
echo "Press Ctrl+C to stop"

# Handle cleanup
trap "kill $VITE_PID $DASHBOARD_PID 2>/dev/null || true" EXIT INT TERM

wait
