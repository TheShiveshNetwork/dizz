# Project
BINARY_NAME := dizz
BUILD_DIR   := bin
MAIN_PKG    := .

# Local development install
INSTALL_DIR := $(HOME)/.local/bin
TARGET_NAME := dizz-dev

# Go env
GO          := go
GOFLAGS     := -trimpath

# Default target
.PHONY: all
all: build

## Build the binary
.PHONY: build
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

## Run the CLI (after build)
.PHONY: run
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## Run with args: make run ARGS="status"
.PHONY: run-args
run-args: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## Dev shortcut (fast rebuild + run)
.PHONY: dev
dev:
	@$(GO) build -o /tmp/$(BINARY_NAME)
	@/tmp/$(BINARY_NAME)

## Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f /tmp/$(BINARY_NAME)

## Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@$(GO) test -v ./tests/...

## Run benchmarks
.PHONY: bench
bench:
	@echo "📊 Running benchmarks..."
	@$(GO) test -bench=. -benchmem ./tests/...

## Format code
.PHONY: fmt
fmt:
	@$(GO) fmt ./...

## Lint (requires golangci-lint)
.PHONY: lint
lint:
	@golangci-lint run

## Install locally (go install)
.PHONY: install
install: build
	@echo "📦 Installing $(TARGET_NAME) to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(TARGET_NAME)
	@chmod +x $(INSTALL_DIR)/$(TARGET_NAME)
	@echo "✅ Installed as $(TARGET_NAME)"

.PHONY: uninstall
uninstall:
	@echo "🗑 Removing $(TARGET_NAME)..."
	@rm -f $(INSTALL_DIR)/$(TARGET_NAME)

build-site:
	./scripts/build-site.sh

wasm:
	cd site/wasm && GOOS=js GOARCH=wasm go build -o ../public/dizz.wasm
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" site/public/

serve-site:
	cd site/server && go run server.go

