#!/bin/bash
# Run tests with coverage

set -e

echo "Running tests..."
go test ./... -v -cover -coverprofile=coverage.out

echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "Coverage report saved to coverage.out"
echo "View HTML report with: go tool cover -html=coverage.out"
