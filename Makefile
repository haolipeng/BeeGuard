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
DETECTOR_SRC=$(PLUGINS_SRC_DIR)/detector
DRIVER_SRC=$(PLUGINS_SRC_DIR)/ebpf_base_detector
NIDS_SRC=$(PLUGINS_SRC_DIR)/nids
SCANNER_SRC=$(PLUGINS_SRC_DIR)/scanner

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
build-plugins: generate-ebpf
	@echo "Building all plugins..."
	@mkdir -p $(PLUGINS_DIR)/collector
	@mkdir -p $(PLUGINS_DIR)/baseline/config
	@mkdir -p $(PLUGINS_DIR)/detector/config/rules
	@mkdir -p $(PLUGINS_DIR)/ebpf_base_detector/config
	@mkdir -p $(PLUGINS_DIR)/nids/config
	@mkdir -p $(PLUGINS_DIR)/scanner/config
	@echo "  Building collector plugin..."
	@cd $(COLLECTOR_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/collector/collector .
	@echo "  Building baseline plugin..."
	@cd $(BASELINE_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/baseline/baseline .
	@cp -r $(BASELINE_SRC)/config/linux $(PLUGINS_DIR)/baseline/config/
	@cp -r $(BASELINE_SRC)/config/container $(PLUGINS_DIR)/baseline/config/ 2>/dev/null || true
	@echo "  Building detector plugin..."
	@cd $(DETECTOR_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/detector/detector .
	@cp $(DETECTOR_SRC)/config/rules/*.yaml $(PLUGINS_DIR)/detector/config/rules/
	@echo "  Building ebpf_base_detector plugin..."
	@cd $(DRIVER_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/ebpf_base_detector/ebpf_base_detector .
	@cp $(DRIVER_SRC)/config/dangerous_commands.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/container_dangerous_commands.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/privilege_escalation_whitelist.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/malicious_request_rules.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/sensitive_file_rules.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/file_monitor_whitelist.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@echo "  Building nids plugin..."
	@cd $(NIDS_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/nids/nids .
	@cp $(NIDS_SRC)/config/nids.yaml $(PLUGINS_DIR)/nids/config/
	@cp $(NIDS_SRC)/config/nids.rules $(PLUGINS_DIR)/nids/config/
	@echo "  Building scanner plugin..."
	@cd $(SCANNER_SRC) && CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/scanner/scanner .
	@cp $(SCANNER_SRC)/config/scanner.yaml $(PLUGINS_DIR)/scanner/config/
	@echo "All plugins built successfully"
	@echo "  $(PLUGINS_DIR)/collector/"
	@echo "  $(PLUGINS_DIR)/baseline/"
	@echo "  $(PLUGINS_DIR)/detector/"
	@echo "  $(PLUGINS_DIR)/ebpf_base_detector/"
	@echo "  $(PLUGINS_DIR)/nids/"
	@echo "  $(PLUGINS_DIR)/scanner/"

# 生成 eBPF 代码 (ebpf_base_detector 插件依赖)
.PHONY: generate-ebpf
generate-ebpf:
	@echo "Generating eBPF code..."
	@cd $(DRIVER_SRC)/ebpf && $(GO) generate ./...
	@echo "eBPF code generation complete"

# 编译 ebpf_base_detector 插件
.PHONY: build-driver
build-driver: generate-ebpf
	@echo "Building ebpf_base_detector plugin..."
	@mkdir -p $(PLUGINS_DIR)/ebpf_base_detector/config
	@cd $(DRIVER_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/ebpf_base_detector/ebpf_base_detector .
	@cp $(DRIVER_SRC)/config/dangerous_commands.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/container_dangerous_commands.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/privilege_escalation_whitelist.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/malicious_request_rules.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/sensitive_file_rules.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@cp $(DRIVER_SRC)/config/file_monitor_whitelist.yaml $(PLUGINS_DIR)/ebpf_base_detector/config/
	@echo "Build complete: $(PLUGINS_DIR)/ebpf_base_detector/"

# 编译 nids 插件
.PHONY: build-nids
build-nids:
	@echo "Building nids plugin..."
	@mkdir -p $(PLUGINS_DIR)/nids/config
	@cd $(NIDS_SRC) && $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/nids/nids .
	@cp $(NIDS_SRC)/config/nids.yaml $(PLUGINS_DIR)/nids/config/
	@cp $(NIDS_SRC)/config/nids.rules $(PLUGINS_DIR)/nids/config/
	@echo "Build complete: $(PLUGINS_DIR)/nids/"

# 编译 scanner 插件
.PHONY: build-scanner
build-scanner:
	@echo "Building scanner plugin..."
	@mkdir -p $(PLUGINS_DIR)/scanner/config
	@cd $(SCANNER_SRC) && CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o ../../$(PLUGINS_DIR)/scanner/scanner .
	@cp $(SCANNER_SRC)/config/scanner.yaml $(PLUGINS_DIR)/scanner/config/
	@echo "Build complete: $(PLUGINS_DIR)/scanner/"

# 编译所有组件 (agent + plugins)
.PHONY: build
build: build-agent build-plugins
	@echo "All components built successfully"
	@echo "  Agent:   $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "  Plugins: $(PLUGINS_DIR)/"

# 编译 cloudsecctl 控制工具
.PHONY: build-ctl
build-ctl:
	@echo "Building cloudsecctl..."
	@mkdir -p $(BUILD_DIR)
	@cd deploy/control && $(GO) build $(GOFLAGS) -o ../../$(BUILD_DIR)/cloudsecctl .
	@echo "Build complete: $(BUILD_DIR)/cloudsecctl"

# 编译所有组件 (agent + plugins + cloudsecctl)
.PHONY: build-all
build-all: build build-ctl
	@echo "All components (including cloudsecctl) built successfully"

# 生成 DEB 安装包
.PHONY: package-deb
package-deb: build-all
	@echo "Generating DEB package..."
	@cd deploy && VERSION=$(VERSION) ARCH=$$(dpkg --print-architecture 2>/dev/null || echo amd64) nfpm package --packager deb --target ../$(BUILD_DIR)/
	@echo "DEB package created in $(BUILD_DIR)/"

# 生成 RPM 安装包
.PHONY: package-rpm
package-rpm: build-all
	@echo "Generating RPM package..."
	@cd deploy && VERSION=$(VERSION) ARCH=$$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/') nfpm package --packager rpm --target ../$(BUILD_DIR)/
	@echo "RPM package created in $(BUILD_DIR)/"

# 生成 DEB + RPM 安装包
.PHONY: package
package: package-deb package-rpm
	@echo "All packages created in $(BUILD_DIR)/"

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
	@sudo mkdir -p $(DEPLOY_DIR)/data/agent
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/collector
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/baseline
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/detector
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/ebpf_base_detector
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/nids
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/scanner
	@sudo mkdir -p $(DEPLOY_DIR)/logs/agent
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/collector
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/baseline
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/detector
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/ebpf_base_detector
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/nids
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/scanner
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(DEPLOY_DIR)/bin/
	@sudo cp -r $(PLUGINS_DIR)/collector/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/baseline/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/detector/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/ebpf_base_detector/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/nids/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/scanner/ $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/bin/$(BINARY_NAME)
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/collector/collector
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/baseline/baseline
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/detector/detector
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/ebpf_base_detector/ebpf_base_detector
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/nids/nids
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/scanner/scanner
	@sudo cp agent.yaml $(DEPLOY_DIR)/
	@sudo cp agent-standalone.yaml $(DEPLOY_DIR)/
	@echo "Deploy complete!"
	@echo "  Agent:   $(DEPLOY_DIR)/bin/$(BINARY_NAME)"
	@echo "  Plugins: $(DEPLOY_DIR)/plugins/"
	@echo "  Config:  $(DEPLOY_DIR)/agent.yaml"
	@echo "  Config:  $(DEPLOY_DIR)/plugins/detector/config/rules/*.yaml"
	@echo "  Config:  $(DEPLOY_DIR)/plugins/ebpf_base_detector/config/dangerous_commands.yaml"
	@echo "  Config:  $(DEPLOY_DIR)/plugins/ebpf_base_detector/config/container_dangerous_commands.yaml"
	@echo "  Config:  $(DEPLOY_DIR)/plugins/ebpf_base_detector/config/privilege_escalation_whitelist.yaml"
	@echo "  Config:  $(DEPLOY_DIR)/plugins/ebpf_base_detector/config/malicious_request_rules.yaml"
	@echo "  Config:  $(DEPLOY_DIR)/plugins/ebpf_base_detector/config/file_monitor_whitelist.yaml"

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
	@sudo cp -r $(PLUGINS_DIR)/collector/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/baseline/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/detector/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/ebpf_base_detector/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/nids/ $(DEPLOY_DIR)/plugins/
	@sudo cp -r $(PLUGINS_DIR)/scanner/ $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/collector/collector
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/baseline/baseline
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/detector/detector
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/ebpf_base_detector/ebpf_base_detector
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/nids/nids
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/scanner/scanner
	@echo "Deploy complete: $(DEPLOY_DIR)/plugins/"

# 仅部署 ebpf_base_detector 插件
.PHONY: deploy-driver
deploy-driver: build-driver
	@echo "Deploying ebpf_base_detector plugin only to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/plugins
	@sudo cp -r $(PLUGINS_DIR)/ebpf_base_detector/ $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/ebpf_base_detector/ebpf_base_detector
	@echo "Deploy complete: $(DEPLOY_DIR)/plugins/ebpf_base_detector/"

# 仅部署 nids 插件
.PHONY: deploy-nids
deploy-nids: build-nids
	@echo "Deploying nids plugin only to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/nids
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/nids
	@sudo mkdir -p $(DEPLOY_DIR)/plugins
	@sudo cp -r $(PLUGINS_DIR)/nids/ $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/nids/nids
	@echo "Deploy complete: $(DEPLOY_DIR)/plugins/nids/"

# 仅部署 scanner 插件
.PHONY: deploy-scanner
deploy-scanner: build-scanner
	@echo "Deploying scanner plugin only to $(DEPLOY_DIR)..."
	@sudo mkdir -p $(DEPLOY_DIR)/data/plugins/scanner
	@sudo mkdir -p $(DEPLOY_DIR)/logs/plugins/scanner
	@sudo mkdir -p $(DEPLOY_DIR)/plugins
	@sudo cp -r $(PLUGINS_DIR)/scanner/ $(DEPLOY_DIR)/plugins/
	@sudo chmod 755 $(DEPLOY_DIR)/plugins/scanner/scanner
	@echo "Deploy complete: $(DEPLOY_DIR)/plugins/scanner/"

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

# 运行 E2E 测试 - Detector
.PHONY: test-e2e-detector
test-e2e-detector:
	@echo "Running Detector E2E tests..."
	@cd tests/e2e/detector && go run main.go

# 运行所有 E2E 测试
.PHONY: test-e2e
test-e2e: test-e2e-baseline test-e2e-collector test-e2e-detector
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
	@echo "  make build-ctl          - Build cloudsecctl control tool"
	@echo "  make build-all          - Build agent + plugins + cloudsecctl"
	@echo "  make build-driver       - Build ebpf_base_detector plugin only"
	@echo "  make build-nids         - Build nids plugin only"
	@echo "  make build-scanner      - Build scanner plugin only (requires libclamav-dev)"
	@echo "  make generate-ebpf      - Generate eBPF code (requires clang, libbpf)"
	@echo "  make clean              - Clean build artifacts"
	@echo ""
	@echo "Package (requires nfpm):"
	@echo "  make package-deb        - Build all + generate DEB package"
	@echo "  make package-rpm        - Build all + generate RPM package"
	@echo "  make package            - Build all + generate DEB + RPM packages"
	@echo ""
	@echo "Deploy (to $(DEPLOY_DIR)):"
	@echo "  make deploy             - Deploy agent + plugins + config"
	@echo "  make deploy-agent       - Deploy agent only"
	@echo "  make deploy-plugins     - Deploy plugins only"
	@echo "  make deploy-driver      - Deploy ebpf_base_detector plugin only"
	@echo "  make deploy-nids        - Deploy nids plugin only"
	@echo "  make deploy-scanner     - Deploy scanner plugin only"
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
