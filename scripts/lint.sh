#!/usr/bin/env bash
set -euo pipefail

echo "==> Running linters..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found, installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# Run golangci-lint
golangci-lint run ./...

# Check formatting
echo "==> Checking formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "The following files are not formatted:"
    echo "$UNFORMATTED"
    echo "Run: gofmt -w ."
    exit 1
fi

# Check go mod tidy
echo "==> Checking go.mod..."
go mod tidy
if ! git diff --quiet go.mod go.sum 2>/dev/null; then
    echo "go.mod or go.sum is not tidy. Run: go mod tidy"
    exit 1
fi

echo "==> Linting passed!"
