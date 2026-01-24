#!/bin/bash
#
# Agent 开发环境初始化脚本
# 用于快速搭建 Agent 本地开发和测试环境
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 目录定义
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENT_DIR="$(dirname "$SCRIPT_DIR")"
DEPLOY_DIR="/opt/cloudsec"

# 打印带颜色的信息
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 显示使用帮助
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Agent Development Environment Setup Script"
    echo ""
    echo "Options:"
    echo "  -i, --init       Initialize development environment (create dirs, build all)"
    echo "  -b, --build      Build agent and plugins"
    echo "  -d, --deploy     Deploy to /opt/cloudsec/"
    echo "  -t, --test       Run E2E tests"
    echo "  -c, --clean      Clean all build artifacts"
    echo "  -h, --help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -i            Full initialization (first time setup)"
    echo "  $0 -b -d         Build and deploy"
    echo "  $0 -t            Run E2E tests"
}

# 检查依赖
check_dependencies() {
    info "Checking dependencies..."

    # 检查 Go
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go 1.21 or later."
    fi

    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    info "Go version: $go_version"

    # 检查 make
    if ! command -v make &> /dev/null; then
        error "make is not installed. Please install make."
    fi

    success "All dependencies are installed"
}

# 创建开发目录结构
create_dev_dirs() {
    info "Creating development directory structure..."

    # 创建部署目录
    sudo mkdir -p "$DEPLOY_DIR/bin"
    sudo mkdir -p "$DEPLOY_DIR/plugins"
    sudo mkdir -p "$DEPLOY_DIR/conf"
    sudo mkdir -p "$DEPLOY_DIR/data/agent"
    sudo mkdir -p "$DEPLOY_DIR/data/plugins/collector"
    sudo mkdir -p "$DEPLOY_DIR/logs/agent"
    sudo mkdir -p "$DEPLOY_DIR/logs/plugins/collector"
    sudo mkdir -p "$DEPLOY_DIR/logs/plugins/baseline"

    # 设置目录权限（开发环境使用更宽松的权限）
    sudo chmod -R 777 "$DEPLOY_DIR"

    success "Directory structure created at $DEPLOY_DIR"
}

# 下载依赖
download_deps() {
    info "Downloading Go dependencies..."

    # Agent 依赖
    info "Downloading agent dependencies..."
    cd "$AGENT_DIR"
    go mod download
    go mod tidy

    # Collector 插件依赖
    if [ -d "$AGENT_DIR/business_plugins/collector" ]; then
        info "Downloading collector plugin dependencies..."
        cd "$AGENT_DIR/business_plugins/collector"
        go mod download
        go mod tidy
    fi

    # Baseline 插件依赖
    if [ -d "$AGENT_DIR/business_plugins/baseline" ]; then
        info "Downloading baseline plugin dependencies..."
        cd "$AGENT_DIR/business_plugins/baseline"
        go mod download
        go mod tidy
    fi

    success "All dependencies downloaded"
}

# 编译 Agent 和插件
build_all() {
    info "Building agent and plugins..."

    cd "$AGENT_DIR"
    make build-all

    success "Agent and plugins built successfully"
    echo ""
    echo "Build artifacts:"
    echo "  Agent:     $AGENT_DIR/build/agent"
    echo "  Plugins:   $AGENT_DIR/build/plugins/"
}

# 部署到开发环境
deploy_dev() {
    info "Deploying to development environment..."

    # 部署 Agent
    if [ -f "$AGENT_DIR/build/agent" ]; then
        sudo cp "$AGENT_DIR/build/agent" "$DEPLOY_DIR/bin/"
        sudo chmod 755 "$DEPLOY_DIR/bin/agent"
        success "Agent deployed"
    else
        warn "Agent binary not found, skipping..."
    fi

    # 部署插件
    if [ -d "$AGENT_DIR/build/plugins" ]; then
        sudo cp -r "$AGENT_DIR/build/plugins/"* "$DEPLOY_DIR/plugins/" 2>/dev/null || true
        sudo chmod 755 "$DEPLOY_DIR/plugins/"* 2>/dev/null || true
        success "Plugins deployed"
    else
        warn "Plugins directory not found, skipping..."
    fi

    # 复制配置文件
    if [ ! -f "$DEPLOY_DIR/conf/agent.yaml" ]; then
        # 创建适用于 /opt/cloudsec 的 agent 配置
        cat << 'EOF' | sudo tee "$DEPLOY_DIR/conf/agent.yaml" > /dev/null
server: "127.0.0.1:50051"
connect_timeout: 30
working_directory: "/opt/cloudsec/data/agent"
plugins_directory: "/opt/cloudsec/plugins"
retry_max_count: 10
retry_interval: 5
EOF
        info "Agent config created"
    fi

    success "Deployment complete!"
    echo ""
    echo "Deployed to: $DEPLOY_DIR"
    ls -la "$DEPLOY_DIR/bin/agent" 2>/dev/null || true
    ls -la "$DEPLOY_DIR/plugins/" 2>/dev/null || true
}

# 运行测试
run_tests() {
    info "Running E2E tests..."

    # 确保插件已部署
    if [ ! -f "$DEPLOY_DIR/plugins/collector" ]; then
        warn "Collector plugin not deployed. Building and deploying..."
        build_all
        deploy_dev
    fi

    # 运行 collector E2E 测试
    info "Running collector E2E tests..."
    cd "$AGENT_DIR"
    make test-e2e-collector

    success "E2E tests complete"
}

# 清理所有编译产物
clean_all() {
    info "Cleaning all build artifacts..."

    cd "$AGENT_DIR"
    make clean

    success "All build artifacts cleaned"
}

# 完整初始化
full_init() {
    echo ""
    echo "==========================================="
    echo "  Agent Development Environment Setup"
    echo "==========================================="
    echo ""

    check_dependencies
    create_dev_dirs
    download_deps
    build_all
    deploy_dev

    echo ""
    echo "==========================================="
    echo "  Setup Complete!"
    echo "==========================================="
    echo ""
    echo "Directory structure:"
    echo "  $DEPLOY_DIR/"
    echo "  ├── bin/           # Executables (agent)"
    echo "  ├── plugins/       # Plugin binaries"
    echo "  ├── conf/          # Configuration files"
    echo "  ├── data/          # Runtime data"
    echo "  └── logs/          # Log files"
    echo ""
    echo "Next steps:"
    echo "  Run agent:  sudo $DEPLOY_DIR/bin/agent -config $DEPLOY_DIR/conf/agent.yaml"
    echo "  Run tests:  $0 -t"
    echo ""
}

# 主函数
main() {
    # 如果没有参数，显示帮助
    if [ $# -eq 0 ]; then
        usage
        exit 0
    fi

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -i|--init)
                full_init
                exit 0
                ;;
            -b|--build)
                build_all
                shift
                ;;
            -d|--deploy)
                create_dev_dirs
                deploy_dev
                shift
                ;;
            -t|--test)
                run_tests
                shift
                ;;
            -c|--clean)
                clean_all
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "Unknown option: $1"
                ;;
        esac
    done
}

main "$@"
