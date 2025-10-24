#!/usr/bin/env bash
set -euo pipefail

# Run all tests with coverage
echo "==> Running tests..."

# Unit tests
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Display coverage summary
go tool cover -func=coverage.txt | tail -n 1

# Optional: generate HTML coverage report
if [ "${HTML_COVERAGE:-false}" = "true" ]; then
    go tool cover -html=coverage.txt -o coverage.html
    echo "Coverage report: coverage.html"
fi

echo "==> Tests passed!"
