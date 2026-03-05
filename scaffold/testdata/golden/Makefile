.DEFAULT_GOAL := help

SERVICE   := example
MODULE    := github.com/example/example
IMAGE_TAG ?= $(SERVICE):dev

# ── local development ─────────────────────────────────────────────────────────
.PHONY: run
run: ## Run the service locally (no TLS, port 8080)
	EXAMPLE_NO_TLS=true \
	EXAMPLE_PORT=8080 \
	EXAMPLE_SERVICE_NAME=$(SERVICE) \
	EXAMPLE_LOG_LEVEL=debug \
	EXAMPLE_EXPORTER_TYPE=stdout \
	EXAMPLE_HEALTHCHECK_CHECK_INTERVAL=10s \
	go run ./cmd/$(SERVICE)

# ── code quality ──────────────────────────────────────────────────────────────
.PHONY: lint
lint: ## Run golangci-lint
	@bash scripts/lint.sh

.PHONY: test
test: ## Run unit tests
	@bash scripts/test.sh

.PHONY: test-cover
test-cover: ## Run tests with HTML coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ── build ─────────────────────────────────────────────────────────────────────
.PHONY: build
build: ## Compile binary
	go build -o bin/$(SERVICE) ./cmd/$(SERVICE)

.PHONY: docker
docker: ## Build Docker image
	docker build -t $(IMAGE_TAG) .

# ── module hygiene ────────────────────────────────────────────────────────────
.PHONY: tidy
tidy: ## Tidy and vendor Go modules
	go mod tidy

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
	  /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
