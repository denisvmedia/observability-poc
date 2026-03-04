# Variables
GO_CMD=go
BINARY_NAME=observability
BIN_DIR=bin
FRONTEND_DIR=frontend
BACKEND_DIR=go

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")

LDFLAGS=-X github.com/denisvmedia/observability-poc/internal/version.Version=$(VERSION) \
        -X github.com/denisvmedia/observability-poc/internal/version.Commit=$(COMMIT) \
        -X github.com/denisvmedia/observability-poc/internal/version.Date=$(BUILD_DATE)

MKDIR=mkdir -p $(1)
RM=rm -rf $(1)
CD=cd
BINARY_PATH=$(BIN_DIR)/$(BINARY_NAME)

# Default target
.PHONY: all
all: build

# Build everything (frontend + backend with embed)
.PHONY: build
build: build-frontend build-backend

# Build backend with embedded frontend
.PHONY: build-backend
build-backend:
	$(call MKDIR,$(BIN_DIR))
	$(CD) $(BACKEND_DIR)/cmd/observability && $(GO_CMD) build -tags with_frontend -ldflags "$(LDFLAGS)" -o ../../../$(BINARY_PATH) .

# Build backend without frontend embed
.PHONY: build-nofe
build-nofe:
	$(call MKDIR,$(BIN_DIR))
	$(CD) $(BACKEND_DIR)/cmd/observability && $(GO_CMD) build -ldflags "$(LDFLAGS)" -o ../../../$(BINARY_PATH) .

# Build the frontend
.PHONY: build-frontend
build-frontend:
	$(CD) $(FRONTEND_DIR) && npm install && npm run build

# Run Go tests
.PHONY: test-go
test-go:
	$(CD) $(BACKEND_DIR) && $(GO_CMD) test ./...

# Run frontend tests
.PHONY: test-frontend
test-frontend:
	$(CD) $(FRONTEND_DIR) && npm run test

# Run all tests
.PHONY: test
test: test-go test-frontend

# Lint Go code (three steps in order)
.PHONY: lint-go
lint-go:
	@echo "Running nolintguard..."
	$(CD) $(BACKEND_DIR) && go run github.com/go-extras/nolintguard/cmd/nolintguard@latest ./...
	@echo ""
	@echo "Running qtlint..."
	$(CD) $(BACKEND_DIR) && go run github.com/go-extras/qtlint/cmd/qtlint@latest ./...
	@echo ""
	@echo "Running golangci-lint..."
	$(CD) $(BACKEND_DIR) && golangci-lint run

# Lint Go code with auto-fix
.PHONY: lint-go-fix
lint-go-fix:
	@echo "Running go fix..."
	$(CD) $(BACKEND_DIR) && go fix ./...
	@echo "Running qtlint with auto-fix..."
	$(CD) $(BACKEND_DIR) && go run github.com/go-extras/qtlint/cmd/qtlint@latest -fix ./...
	@echo ""
	@echo "Running golangci-lint with auto-fix..."
	$(CD) $(BACKEND_DIR) && golangci-lint run --fix

# Lint frontend code
.PHONY: lint-frontend
lint-frontend:
	$(CD) $(FRONTEND_DIR) && npm run lint

# Lint everything
.PHONY: lint
lint: lint-go lint-frontend

# Clean build artifacts
.PHONY: clean
clean:
	$(call RM,$(BIN_DIR))
	$(CD) $(FRONTEND_DIR) && npm run clean

# Run targets

# Build and run the binary with the default ClickHouse DSN (requires local ClickHouse).
.PHONY: run
run: build-nofe
	./$(BINARY_PATH) run

# Start ClickHouse + run schema init via Docker, then run the binary pointing at it.
.PHONY: run-clickhouse
run-clickhouse: build-nofe
	docker compose up -d clickhouse init-schema
	OBSERVABILITY_DB_DSN="clickhouse://observability:observability_password@localhost:9000/observability" \
	  ./$(BINARY_PATH) run

# Docker targets
.PHONY: docker-build
docker-build:
	docker build --target production -t observability-poc:latest .

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down

