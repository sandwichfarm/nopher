#!/bin/bash
# Run nopher in development mode

set -e

CONFIG=${1:-test-config.yaml}

if [ ! -f "$CONFIG" ]; then
    echo "Error: Configuration file not found: $CONFIG"
    echo "Usage: $0 [config-file]"
    exit 1
fi

echo "Starting nopher in development mode..."
echo "Config: $CONFIG"
echo ""

go run ./cmd/nopher --config "$CONFIG"
