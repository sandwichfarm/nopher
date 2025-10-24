.PHONY: help build test lint clean install dev release release-snapshot docker docker-compose-up docker-compose-down ci-setup ci-verify fmt check

# Variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build binary
	@VERSION=$(VERSION) COMMIT=$(COMMIT) DATE=$(DATE) ./scripts/build.sh

test: ## Run tests
	@./scripts/test.sh

lint: ## Run linters
	@./scripts/lint.sh

clean: ## Clean build artifacts
	rm -rf dist/ coverage.txt coverage.html

install: build ## Install to /usr/local/bin
	install -m 755 dist/nopher /usr/local/bin/nopher

dev: ## Run in development mode
	go run ./cmd/nopher --config ./configs/nopher.example.yaml

release: ## Create a release (requires goreleaser)
	goreleaser release --clean

release-snapshot: ## Create a snapshot release
	goreleaser release --snapshot --clean

docker: ## Build Docker image
	docker build -t nopher:$(VERSION) .

docker-compose-up: ## Start with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	docker-compose down

ci-setup: ## Setup CI environment
	@./scripts/ci/setup.sh

ci-verify: ## Verify build artifacts
	@./scripts/ci/verify.sh

fmt: ## Format code
	gofmt -w .
	go mod tidy

check: lint test ## Run all checks
