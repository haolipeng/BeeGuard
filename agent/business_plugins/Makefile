# Makefile for business plugins

# 变量定义
GO=go
BUILD_DIR=build
PLUGINS_DIR=.
PLUGIN_LIST=baseline

# 编译标志
LDFLAGS=-ldflags "-s -w"
GOFLAGS=-trimpath

# 默认目标
.PHONY: all
all: build

# 编译所有插件
.PHONY: build
build: $(PLUGIN_LIST)
	@echo "All plugins built successfully"

# 编译单个插件
.PHONY: baseline
baseline:
	@echo "Building baseline plugin..."
	@mkdir -p $(BUILD_DIR)
	@cd baseline && $(GO) build $(GOFLAGS) $(LDFLAGS) -o ../$(BUILD_DIR)/baseline main.go
	@echo "Build complete: $(BUILD_DIR)/baseline"

# 清理编译产物
.PHONY: clean
clean:
	@echo "Cleaning plugin build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# 安装依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies for lib..."
	@cd lib && $(GO) mod download && $(GO) mod tidy
	@echo "Downloading dependencies for baseline..."
	@cd baseline && $(GO) mod download && $(GO) mod tidy
	@echo "Dependencies updated"

# 编译并安装所有插件
.PHONY: install
install: build
	@echo "Installing plugins..."
	@mkdir -p /opt/agent/plugins
	@install -m 755 $(BUILD_DIR)/baseline /opt/agent/plugins/baseline
	@echo "Install complete: /opt/agent/plugins/"

# 只编译baseline插件
.PHONY: build-baseline
build-baseline: baseline

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@cd lib && $(GO) fmt ./...
	@cd baseline && $(GO) fmt ./...
	@echo "Format complete"

# 代码检查
.PHONY: vet
vet:
	@echo "Vetting code..."
	@cd lib && $(GO) vet ./...
	@cd baseline && $(GO) vet ./...
	@echo "Vet complete"

# 显示帮助信息
.PHONY: help
help:
	@echo "Business Plugins Makefile Commands:"
	@echo "  make build           - Build all plugins"
	@echo "  make baseline        - Build baseline plugin only"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make deps            - Download and tidy dependencies"
	@echo "  make install         - Install plugins to /opt/agent/plugins"
	@echo "  make fmt             - Format Go code"
	@echo "  make vet             - Run go vet"
	@echo "  make help            - Show this help message"
