#!/bin/bash
# Clean build artifacts

set -e

echo "Cleaning build artifacts..."

rm -f nopher
rm -f coverage.out
rm -rf dist/
rm -rf test-data/

echo "✓ Clean complete"
