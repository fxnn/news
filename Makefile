.PHONY: all build story-extractor ui-server test cover fmt vet clean help

all: fmt vet test build ## Format, vet, test, and build everything

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-20s %s\n", $$1, $$2}'

build: story-extractor ui-server ## Build both binaries

story-extractor: ## Build story-extractor
	go build -o story-extractor ./cmd/story-extractor

ui-server: ## Build ui-server
	go build -o ui-server ./cmd/ui-server

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
