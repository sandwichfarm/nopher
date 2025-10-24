#!/bin/bash
# Build nophr binary

set -e

VERSION=${VERSION:-dev}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILT_BY=${BUILT_BY:-$(whoami)}

echo "Building nophr..."
echo "  Version: $VERSION"
echo "  Commit:  $COMMIT"
echo "  Date:    $DATE"
echo "  Built by: $BUILT_BY"
echo ""

go build \
    -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE -X main.builtBy=$BUILT_BY" \
    -o nophr \
    ./cmd/nophr

echo ""
echo "✓ Build complete: ./nophr"
./nophr --version
