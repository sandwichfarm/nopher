#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
DATE="${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

echo "==> Building nopher..."
echo "    Version: $VERSION"
echo "    Commit:  $COMMIT"
echo "    Date:    $DATE"

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.version=$VERSION"
LDFLAGS="$LDFLAGS -X main.commit=$COMMIT"
LDFLAGS="$LDFLAGS -X main.date=$DATE"

# Build binary
CGO_ENABLED=0 go build \
    -ldflags "$LDFLAGS" \
    -o dist/nopher \
    ./cmd/nopher

echo "==> Build complete: dist/nopher"

# Show version
./dist/nopher --version
