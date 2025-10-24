# Nopher Makefile

.PHONY: all build test lint clean install dev help

# Default target
all: test build

# Build the binary
build:
	@./scripts/build.sh

# Run tests
test:
	@./scripts/test.sh

# Run linter
lint:
	@./scripts/lint.sh

# Clean build artifacts
clean:
	@./scripts/clean.sh

# Install to GOPATH/bin
install: build
	@echo "Installing nopher..."
	@cp nopher $(shell go env GOPATH)/bin/
	@echo "✓ Installed to $(shell go env GOPATH)/bin/nopher"

# Run in development mode
dev:
	@./scripts/dev.sh

# Run with custom config
run:
	@./scripts/dev.sh $(CONFIG)

# Build Docker image
docker:
	@docker build -t nopher:latest .

# Show help
help:
	@echo "Nopher Makefile targets:"
	@echo ""
	@echo "  make build    - Build the nopher binary"
	@echo "  make test     - Run tests with coverage"
	@echo "  make lint     - Run golangci-lint"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make install  - Install to GOPATH/bin"
	@echo "  make dev      - Run in development mode (uses test-config.yaml)"
	@echo "  make run CONFIG=<path> - Run with custom config"
	@echo "  make docker   - Build Docker image"
	@echo "  make all      - Run tests and build"
	@echo "  make help     - Show this help message"
	@echo ""
