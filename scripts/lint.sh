#!/bin/bash
# Run golangci-lint

set -e

if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Install with:"
    echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin"
    exit 1
fi

echo "Running golangci-lint..."
golangci-lint run ./...

echo ""
echo "✓ Linting passed!"
