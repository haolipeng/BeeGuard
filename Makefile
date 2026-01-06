# Makefile for agent main program

# 变量定义
BINARY_NAME=agent
GO=go
BUILD_DIR=build
MAIN_FILE=main.go
MODULE=gitlab.myinterest.top/security/agent

# 编译标志
LDFLAGS=-ldflags "-s -w"
GOFLAGS=-trimpath

# 默认目标
.PHONY: all
all: build

# 编译agent主程序
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 清理编译产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# 运行agent
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# 安装依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Dependencies updated"

# 编译并安装到系统
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	@install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Install complete: /usr/local/bin/$(BINARY_NAME)"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete"

# 代码检查
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GO) vet ./...
	@echo "Vet complete"

# 运行单元测试
.PHONY: test
test:
	@echo "Running unit tests..."
	$(GO) test ./... -v
	@echo "Tests complete"

# 运行 E2E 测试 - Baseline
.PHONY: test-e2e-baseline
test-e2e-baseline:
	@echo "Running Baseline E2E tests..."
	@cd tests/e2e/baseline && ./test.sh

# 运行 E2E 测试 - Collector
.PHONY: test-e2e-collector
test-e2e-collector:
	@echo "Running Collector E2E tests..."
	@cd tests/e2e/collector && ./test.sh

# 运行所有 E2E 测试
.PHONY: test-e2e
test-e2e: test-e2e-baseline test-e2e-collector
	@echo "All E2E tests complete"

# 运行所有测试（单元测试 + E2E 测试）
.PHONY: test-all
test-all: test test-e2e
	@echo "All tests complete"

# 显示帮助信息
.PHONY: help
help:
	@echo "Agent Makefile Commands:"
	@echo "  make build              - Build agent binary"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make run                - Build and run agent"
	@echo "  make deps               - Download and tidy dependencies"
	@echo "  make install            - Install agent to /usr/local/bin"
	@echo "  make fmt                - Format Go code"
	@echo "  make vet                - Run go vet"
	@echo "  make test               - Run unit tests"
	@echo "  make test-e2e-baseline  - Run Baseline E2E tests"
	@echo "  make test-e2e-collector - Run Collector E2E tests"
	@echo "  make test-e2e           - Run all E2E tests"
	@echo "  make test-all           - Run all tests (unit + E2E)"
	@echo "  make help               - Show this help message"
