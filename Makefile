.PHONY: help build run clean test lint fmt deps setup interactive daemon peer-network stress-test check docs

# Variables
BINARY_NAME=node
BUILD_DIR=build
BIN_PATH=$(BUILD_DIR)/$(BINARY_NAME)
GO_VERSION=$(shell go version | awk '{print $$3}')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Default target
help:
	@echo "P2P-DPoS v2 - Delegated Proof of Stake Blockchain"
	@echo ""
	@echo "Setup & Dependencies:"
	@echo "  make setup          - Install dependencies and setup"
	@echo "  make deps           - Download and verify go.mod dependencies"
	@echo "  make check          - Verify environment and dependencies"
	@echo ""
	@echo "Build:"
	@echo "  make build          - Build the binary"
	@echo "  make build-debug    - Build with debug symbols"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Run & Test:"
	@echo "  make interactive    - Run single interactive node (dev)"
	@echo "  make daemon         - Run node as daemon service"
	@echo "  make test-single    - Test single node setup"
	@echo "  make test-multi     - Test multi-node network (2 nodes)"
	@echo "  make test-stress    - Stress test with 5 nodes"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make test           - Run unit tests"
	@echo "  make coverage       - Test with coverage report"
	@echo ""
	@echo "Documentation:"
	@echo "  make docs           - Generate documentation"
	@echo "  make info           - Show environment info"
	@echo ""

# Setup & Dependencies
setup: check deps
	@echo "✓ Setup complete!"

check:
	@echo "=== Environment Check ==="
	@echo "Go Version: $(GO_VERSION)"
	@if ! command -v go &> /dev/null; then \
		echo "✗ Go is not installed"; exit 1; \
	fi
	@if [ "$$(go version | grep -oP 'go1\.\d+' | cut -d. -f2)" -lt 22 ]; then \
		echo "✗ Go 1.22+ required"; exit 1; \
	fi
	@echo "✓ Go version OK"
	@echo "✓ Environment check passed"

deps:
	@echo "=== Downloading Dependencies ==="
	go mod download
	go mod verify
	go mod tidy
	@echo "✓ Dependencies downloaded"

# Build targets
build: check
	@echo "=== Building $(BINARY_NAME) ==="
	mkdir -p $(BUILD_DIR)
	go build \
		-ldflags="-X main.Version=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)" \
		-o $(BIN_PATH) \
		./cmd/node
	@echo "✓ Build complete: $(BIN_PATH)"
	@ls -lh $(BIN_PATH)

build-debug: check
	@echo "=== Building $(BINARY_NAME) (debug mode) ==="
	mkdir -p $(BUILD_DIR)
	go build -gcflags="all=-N -l" -o $(BIN_PATH).debug ./cmd/node
	@echo "✓ Debug build complete: $(BIN_PATH).debug"

build-all: check
	@echo "=== Cross-compiling for multiple platforms ==="
	mkdir -p $(BUILD_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build -o $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch ./cmd/node; \
		done; \
	done
	@echo "✓ Cross-compilation complete"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)-*

clean:
	@echo "=== Cleaning build artifacts ==="
	rm -rf $(BUILD_DIR)
	rm -rf blockchain/
	rm -rf *.db
	@echo "✓ Clean complete"

# Run targets
interactive: build
	@echo "=== Starting Interactive Node ==="
	@echo "Available commands at prompt:"
	@echo "  status              - Show node status"
	@echo "  balance             - Check your balance"
	@echo "  register-validator  - Become a validator"
	@echo "  validators          - List all validators"
	@echo "  transfer            - Send tokens"
	@echo "  delegate            - Delegate tokens"
	@echo "  exit                - Stop node"
	@echo ""
	$(BIN_PATH) interactive --initial-balance 5000

daemon: build
	@echo "=== Starting Node in Daemon Mode ==="
	$(BIN_PATH) start --datadir ./blockchain --loglevel info

# Test targets
test-single: build
	@echo "=== Test 1: Single Node Setup ==="
	@echo "Starting node with 1000 initial tokens..."
	@echo ""
	@echo "Instructions:"
	@echo "1. Wait 5 seconds"
	@echo "2. Type: status"
	@echo "3. Type: register-validator --stake 500"
	@echo "4. Type: status (should show 1 validator)"
	@echo "5. Wait 10 seconds for block production"
	@echo "6. Type: exit"
	@echo ""
	$(BIN_PATH) interactive --initial-balance 1000

test-multi:
	@echo "=== Test 2: Multi-Node Network (2 Nodes) ==="
	@echo ""
	@echo "Run in two separate terminals:"
	@echo ""
	@echo "Terminal 1:"
	@echo "  make run-node1"
	@echo ""
	@echo "Terminal 2 (after node1 starts):"
	@echo "  make run-node2"
	@echo ""
	@echo "Then in each terminal:"
	@echo "  > register-validator --stake 2000"
	@echo "  > status  (verify 2 validators)"
	@echo "  > validators  (see ranking)"
	@echo ""

run-node1: build
	@echo "=== Node 1 ==="
	@echo "Your peer ID will appear below. Copy it for Node 2."
	$(BIN_PATH) interactive --initial-balance 5000 --port 30333

run-node2: build
	@echo "=== Node 2 ==="
	@echo "Connect to Node 1:"
	@read -p "Enter Node 1 peer ID: " peer_id; \
	$(BIN_PATH) interactive --initial-balance 5000 --port 30334 \
		--bootstrap "/ip4/127.0.0.1/tcp/30333/p2p/$$peer_id"

test-stress:
	@echo "=== Test 3: Stress Test (5-Node Network) ==="
	@echo ""
	@echo "This test requires 5 terminal windows."
	@echo ""
	@echo "Run in 5 separate terminals:"
	@for i in 1 2 3 4 5; do \
		port=$$(( 30333 + $$i )); \
		echo "Terminal $$i: make run-node-$$i"; \
	done
	@echo ""
	@echo "Then in each terminal:"
	@echo "  > register-validator --stake 2000"
	@echo "  > wait 30 seconds"
	@echo "  > status  (check block production)"
	@echo ""

run-node-1: build
	@$(BIN_PATH) interactive --initial-balance 5000 --port 30334

run-node-2: build
	@echo "Enter Node 1 peer ID:" && read peer_id && \
	$(BIN_PATH) interactive --initial-balance 5000 --port 30335 --bootstrap "/ip4/127.0.0.1/tcp/30334/p2p/$$peer_id"

run-node-3: build
	@echo "Enter Node 1 peer ID:" && read peer_id && \
	$(BIN_PATH) interactive --initial-balance 5000 --port 30336 --bootstrap "/ip4/127.0.0.1/tcp/30334/p2p/$$peer_id"

# Code quality targets
fmt:
	@echo "=== Formatting code ==="
	go fmt ./...
	@echo "✓ Code formatted"

lint:
	@echo "=== Running linter ==="
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...
	@echo "✓ Linting complete"

test:
	@echo "=== Running unit tests ==="
	go test -v -race ./...
	@echo "✓ Tests passed"

coverage:
	@echo "=== Running tests with coverage ==="
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Documentation targets
docs:
	@echo "=== Documentation ==="
	@echo ""
	@echo "Available docs:"
	@echo "  README.md              - Architecture and overview"
	@echo "  QUICKSTART.md          - Getting started guide"
	@echo "  VERSION_COMPARISON.md  - V1 vs V2 comparison"
	@echo ""

info:
	@echo "=== Environment Information ==="
	@echo "Go Version: $(GO_VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Binary Path: $(BIN_PATH)"
	@echo ""
	@echo "Project Structure:"
	@ls -d */ | sed 's/^/  /'
	@echo ""

# Test all functionality
test-all: clean build test lint
	@echo "✓ All tests passed!"

.DEFAULT_GOAL := help
