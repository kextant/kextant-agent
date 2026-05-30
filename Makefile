.PHONY: help build test test-coverage test-functional test-all clean docker-build docker-run-smoke docker-push validate scan scan-send inventory fmt vet lint tidy

BINARY_NAME := kextant-agent
DOCKER_IMAGE ?= ghcr.io/kextant/kextant-agent
VERSION ?= 0.1.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/agent

test: ## Run unit tests
	go test -race -count=1 ./...

test-coverage: ## Run unit tests with coverage profile and enforce minimum total coverage
	go test -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tee coverage.txt
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/, "", $$3); print $$3}'); \
	awk -v total="$$total" 'BEGIN { if (total < 60.0) { printf "coverage %.1f%% is below required 60.0%%\n", total; exit 1 } }'

test-functional: docker-build docker-run-smoke ## Run containerized functional smoke tests

test-all: fmt vet lint test-coverage test-functional ## Run all local quality gates

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME) coverage.out coverage.html manifest.json

docker-build: ## Build Docker image
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg BUILD_DATE=$(BUILD_DATE) -t $(DOCKER_IMAGE):$(VERSION) .

docker-run-smoke: ## Run containerized smoke tests against the built image
	docker run --rm $(DOCKER_IMAGE):$(VERSION) version --json
	docker run --rm -e LOG_ENABLED=true $(DOCKER_IMAGE):$(VERSION) validate

docker-push: ## Push Docker image
	docker push $(DOCKER_IMAGE):$(VERSION)

validate: ## Validate configuration
	go run ./cmd/agent validate

scan: ## Run a manual scan
	go run ./cmd/agent scan

scan-send: ## Run a manual scan and send report
	go run ./cmd/agent scan --send

inventory: ## Print inventory manifest JSON
	go run ./cmd/agent inventory

fmt: ## Format code
	gofmt -w cmd internal pkg

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	golangci-lint run

tidy: ## Tidy go modules
	go mod tidy

.DEFAULT_GOAL := help
