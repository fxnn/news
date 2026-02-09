.PHONY: all build story-extractor ui-server test cover fmt vet clean help

VERSION_PKG := github.com/fxnn/news/internal/version
BUILD_TIMESTAMP := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null | grep -v '^HEAD$$' || echo unknown)
LDFLAGS := -X $(VERSION_PKG).BuildTimestamp=$(BUILD_TIMESTAMP) -X $(VERSION_PKG).BuildBranch=$(BUILD_BRANCH)

all: fmt vet test build ## Format, vet, test, and build everything

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

build: story-extractor ui-server ## Build both binaries

story-extractor: ## Build story-extractor
	go build -ldflags "$(LDFLAGS)" -o story-extractor ./cmd/story-extractor

ui-server: ## Build ui-server
	go build -ldflags "$(LDFLAGS)" -o ui-server ./cmd/ui-server

test: ## Run all tests
	go test ./...

cover: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt: ## Format all Go source files
	go fmt ./...

vet: ## Run static analysis
	go vet ./...

clean: ## Remove binaries and coverage output
	rm -f story-extractor ui-server coverage.out
