#!/bin/bash

set -e

echo "Building dashboard frontend..."
cd dashboard
npm run build

echo "Copying dashboard assets to Go package..."
mkdir -p ../internal/dashboard/dist
cp -r dist/* ../internal/dashboard/dist/

echo "Building Go binary with embedded dashboard..."
cd ..
go build -o portfolio ./cmd/portfolio

echo "Build complete!"
echo "Binary: ./portfolio"
echo "Start with: ./portfolio dashboard"
