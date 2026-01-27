# Makefile for agent main program

# 变量定义
BINARY_NAME=agent
GO=go
BUILD_DIR=build
PLUGINS_DIR=$(BUILD_DIR)/plugins
MAIN_FILE=main.go
MODULE=gitlab.myinterest.top/security/agent

# 插件源码目录
PLUGINS_SRC_DIR=business_plugins
COLLECTOR_SRC=$(PLUGINS_SRC_DIR)/collector
BASELINE_SRC=$(PLUGINS_SRC_DIR)/baseline

# 部署目录
DEPLOY_DIR=/opt/cloudsec

# 版本信息（可通过命令行覆盖）
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 编译标志
LDFLAGS=-ldflags "-s -w \
	-X $(MODULE)/agent.Version=$(VERSION) \
	-X $(MODULE)/agent.BuildTime=$(BUILD_TIME) \
	-X $(MODULE)/agent.GitCommit=$(GIT_COMMIT)"
GOFLAGS=-trimpath

# 默认目标
.PHONY: all
all: build

# 编译 agent 主程序
.PHONY: build-agent
build-agent:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# 编译所有插件
.PHONY: build-plugins
build-plugins:
	@echo "Building all plugins..."
	@mkdir -p $(PLUGINS_DIR)
	@echo "  Building collector plugin..."
	@cd $(COLLECTOR_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/collector .
	@echo "  Building baseline plugin..."
	@cd $(BASELINE_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/baseline .
	@echo "All plugins built successfully"
	@echo "  $(PLUGINS_DIR)/collector"
	@echo "  $(PLUGINS_DIR)/baseline"

# 编译所有组件 (agent + plugins)
.PHONY: build
build: build-agent build-plugins
	@echo "All components built successfully"
	@echo "  Agent:   $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "  Plugins: $(PLUGINS_DIR)/"

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

# 部署到 /opt/cloudsec/
.PHONY: deploy
deploy: build
	@echo "Deploying to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/bin
	@sudo mkdir -p $(DEPLOY_DIR)/plugins
	@sudo mkdir -p $(DEPLOY_DIR)/conf
	@sudo mkdir -p $(DEPLOY_DIR)/data/agent
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/collector
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/baseline
	@sudo mkdir -p $(DEPLOY_DIR)/logs/agent
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/collector
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/baseline
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(DEPLOY_DIR)/bin/
	@sudo cp $(PLUGINS_DIR)/collector $(DEPLOY_DIR)/plugins/
	@sudo cp $(PLUGINS_DIR)/baseline $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/bin/$(BINARY_NAME)
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/*
	@if [ ! -f $(DEPLOY_DIR)/conf/agent.yaml ]; then \
		sudo cp agent.yaml $(DEPLOY_DIR)/conf/agent.yaml; \
		echo "Config copied to $(DEPLOY_DIR)/conf/agent.yaml"; \
	else \
		echo "Config already exists, skipping..."; \
	fi
	@echo "Deploy complete!"
	@echo "  Agent:   $(DEPLOY_DIR)/bin/$(BINARY_NAME)"
	@echo "  Plugins: $(DEPLOY_DIR)/plugins/"
	@echo "  Config:  $(DEPLOY_DIR)/conf/agent.yaml"

# 仅部署 agent（不含插件）
.PHONY: deploy-agent
deploy-agent: build-agent
	@echo "Deploying agent only to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/bin
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(DEPLOY_DIR)/bin/
	@sudo chmod 755 $(DEPLOY_DIR)/bin/$(BINARY_NAME)
	@echo "Deploy complete: $(DEPLOY_DIR)/bin/$(BINARY_NAME)"

# 仅部署插件
.PHONY: deploy-plugins
deploy-plugins: build-plugins
	@echo "Deploying plugins only to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/plugins
	@sudo cp $(PLUGINS_DIR)/* $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/*
	@echo "Deploy complete: $(DEPLOY_DIR)/plugins/"

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
	@echo ""
	@echo "Build:"
	@echo "  make build              - Build agent + all plugins"
	@echo "  make build-agent        - Build agent only"
	@echo "  make build-plugins      - Build all plugins"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Deploy (to $(DEPLOY_DIR)):"
	@echo "  make deploy             - Deploy agent + plugins + config"
	@echo "  make deploy-agent       - Deploy agent only"
	@echo "  make deploy-plugins     - Deploy plugins only"
	@echo ""
	@echo "Run & Test:"
	@echo "  make run                - Build and run agent"
	@echo "  make test               - Run unit tests"
	@echo "  make test-e2e-baseline  - Run Baseline E2E tests"
	@echo "  make test-e2e-collector - Run Collector E2E tests"
	@echo "  make test-e2e           - Run all E2E tests"
	@echo "  make test-all           - Run all tests (unit + E2E)"
	@echo ""
	@echo "Other:"
	@echo "  make deps               - Download and tidy dependencies"
	@echo "  make install            - Install agent to /usr/local/bin"
	@echo "  make fmt                - Format Go code"
	@echo "  make vet                - Run go vet"
	@echo "  make help               - Show this help message"
