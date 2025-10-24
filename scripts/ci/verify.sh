#!/usr/bin/env bash
set -euo pipefail

echo "==> Verifying build artifacts..."

# Check that binaries exist
if [ ! -d "dist" ]; then
    echo "ERROR: dist/ directory not found"
    exit 1
fi

# Check that at least one binary exists
BINARY_COUNT=$(find dist -type f -executable 2>/dev/null | wc -l)
if [ "$BINARY_COUNT" -eq 0 ]; then
    echo "ERROR: No binaries found in dist/"
    exit 1
fi

echo "==> Found $BINARY_COUNT binaries"

# Verify each binary
for binary in dist/*; do
    if [ -x "$binary" ] && [ -f "$binary" ]; then
        echo "    Checking: $binary"

        # Check if binary runs
        if ! "$binary" --version &> /dev/null; then
            echo "    WARNING: $binary --version failed"
        else
            echo "    ✓ $binary is valid"
        fi
    fi
done

echo "==> Verification complete!"
