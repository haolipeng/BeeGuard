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

# 显示帮助信息
.PHONY: help
help:
	@echo "Agent Makefile Commands:"
	@echo "  make build    - Build agent binary"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make run      - Build and run agent"
	@echo "  make deps     - Download and tidy dependencies"
	@echo "  make install  - Install agent to /usr/local/bin"
	@echo "  make fmt      - Format Go code"
	@echo "  make vet      - Run go vet"
	@echo "  make help     - Show this help message"
