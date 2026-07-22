BINARY_NAME := dizzie
BUILD_DIR   := bin
GO          := go
GOFLAGS     := -trimpath
VERSION     := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
LDFLAGS     := -X main.version=$(VERSION)

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) .

.PHONY: run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" .

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: clean
clean:
	@rm -rf $(BUILD_DIR)

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: fmt
fmt:
	$(GO) fmt ./...
