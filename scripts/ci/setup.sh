#!/usr/bin/env bash
set -euo pipefail

echo "==> Setting up CI environment..."

# Install dependencies
go mod download

# Install tools
echo "==> Installing tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify Go version
echo "==> Go version:"
go version

echo "==> Setup complete!"
