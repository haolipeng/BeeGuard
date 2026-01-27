#!/bin/bash
#
# Agent 部署脚本
# 用于将 Agent 和插件编译产物部署到 /opt/cloudsec/ 目录
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
    echo "Agent Deployment Script"
    echo ""
    echo "Options:"
    echo "  -a, --all        Deploy agent and plugins"
    echo "  -g, --agent      Deploy agent only"
    echo "  -p, --plugins    Deploy plugins only"
    echo "  -b, --build      Build before deploy"
    echo "  -c, --clean      Clean deploy directory before deploy"
    echo "  -h, --help       Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 -a -b         Build and deploy agent with plugins"
    echo "  $0 -g            Deploy agent only (must be built first)"
    echo "  $0 -p -b         Build and deploy plugins only"
}

# 创建部署目录结构
create_deploy_dirs() {
    info "Creating deployment directory structure..."

    sudo mkdir -p "$DEPLOY_DIR/bin"
    sudo mkdir -p "$DEPLOY_DIR/plugins"
    sudo mkdir -p "$DEPLOY_DIR/conf"
    sudo mkdir -p "$DEPLOY_DIR/data/agent"
    sudo mkdir -p "$DEPLOY_DIR/data/plugins/collector"
    sudo mkdir -p "$DEPLOY_DIR/logs/agent"
    sudo mkdir -p "$DEPLOY_DIR/logs/plugins/collector"
    sudo mkdir -p "$DEPLOY_DIR/logs/plugins/baseline"

    # 设置目录权限
    sudo chmod 755 "$DEPLOY_DIR"
    sudo chmod 755 "$DEPLOY_DIR/bin"
    sudo chmod 755 "$DEPLOY_DIR/plugins"
    sudo chmod 755 "$DEPLOY_DIR/conf"
    sudo chmod 755 "$DEPLOY_DIR/data"
    sudo chmod 755 "$DEPLOY_DIR/logs"

    success "Directory structure created"
}

# 清理部署目录
clean_deploy() {
    info "Cleaning agent deployment..."
    sudo rm -f "$DEPLOY_DIR/bin/agent"
    sudo rm -rf "$DEPLOY_DIR/plugins/"*
    success "Agent deployment cleaned"
}

# 编译 Agent 和插件
build_agent() {
    info "Building agent and plugins..."
    cd "$AGENT_DIR"
    make build-all
    success "Agent and plugins built successfully"
}

# 部署 Agent
deploy_agent() {
    info "Deploying agent..."

    if [ ! -f "$AGENT_DIR/build/agent" ]; then
        error "Agent binary not found. Please build first with -b flag"
    fi

    sudo cp "$AGENT_DIR/build/agent" "$DEPLOY_DIR/bin/"
    sudo chmod 755 "$DEPLOY_DIR/bin/agent"

    # 复制配置文件（如果不存在）
    if [ ! -f "$DEPLOY_DIR/conf/agent.yaml" ]; then
        sudo cp "$AGENT_DIR/agent.yaml" "$DEPLOY_DIR/conf/agent.yaml"
        info "Config copied to $DEPLOY_DIR/conf/agent.yaml"
    else
        warn "Config already exists, skipping..."
    fi

    success "Agent deployed to $DEPLOY_DIR/bin/agent"
}

# 部署插件
deploy_plugins() {
    info "Deploying plugins..."

    local plugins_dir="$AGENT_DIR/build/plugins"

    if [ ! -d "$plugins_dir" ]; then
        error "Plugins directory not found. Please build first with -b flag"
    fi

    # 部署 collector 插件
    if [ -f "$plugins_dir/collector" ]; then
        sudo cp "$plugins_dir/collector" "$DEPLOY_DIR/plugins/"
        sudo chmod 755 "$DEPLOY_DIR/plugins/collector"
        success "Collector plugin deployed"
    else
        warn "Collector plugin not found, skipping..."
    fi

    # 部署 baseline 插件
    if [ -f "$plugins_dir/baseline" ]; then
        sudo cp "$plugins_dir/baseline" "$DEPLOY_DIR/plugins/"
        sudo chmod 755 "$DEPLOY_DIR/plugins/baseline"
        success "Baseline plugin deployed"
    else
        warn "Baseline plugin not found, skipping..."
    fi

    success "Plugins deployed to $DEPLOY_DIR/plugins/"
}

# 显示部署结果
show_result() {
    echo ""
    echo "=========================================="
    echo "  Agent Deployment Complete"
    echo "=========================================="
    echo ""
    echo "Deployment directory: $DEPLOY_DIR"
    echo ""

    if [ -f "$DEPLOY_DIR/bin/agent" ]; then
        echo "  Agent:    $DEPLOY_DIR/bin/agent"
    fi

    if [ -d "$DEPLOY_DIR/plugins" ] && [ "$(ls -A $DEPLOY_DIR/plugins 2>/dev/null)" ]; then
        echo "  Plugins:  $DEPLOY_DIR/plugins/"
        ls -1 "$DEPLOY_DIR/plugins/" | while read p; do
            echo "            - $p"
        done
    fi

    echo ""
    echo "Config files: $DEPLOY_DIR/conf/"
    echo "Log files:    $DEPLOY_DIR/logs/"
    echo "Data files:   $DEPLOY_DIR/data/"
    echo ""
}

# 主函数
main() {
    local do_build=false
    local do_clean=false
    local deploy_agent_flag=false
    local deploy_plugins_flag=false

    # 如果没有参数，显示帮助
    if [ $# -eq 0 ]; then
        usage
        exit 0
    fi

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -a|--all)
                deploy_agent_flag=true
                deploy_plugins_flag=true
                shift
                ;;
            -g|--agent)
                deploy_agent_flag=true
                shift
                ;;
            -p|--plugins)
                deploy_plugins_flag=true
                shift
                ;;
            -b|--build)
                do_build=true
                shift
                ;;
            -c|--clean)
                do_clean=true
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

    # 检查是否有部署目标
    if ! $deploy_agent_flag && ! $deploy_plugins_flag; then
        error "No deployment target specified. Use -a, -g, or -p"
    fi

    echo ""
    echo "=========================================="
    echo "  Agent Deployment Script"
    echo "=========================================="
    echo ""

    # 创建目录结构
    create_deploy_dirs

    # 清理（如果需要）
    if $do_clean; then
        clean_deploy
    fi

    # 编译（如果需要）
    if $do_build; then
        build_agent
    fi

    # 部署
    if $deploy_agent_flag; then
        deploy_agent
    fi

    if $deploy_plugins_flag; then
        deploy_plugins
    fi

    # 显示结果
    show_result
}

main "$@"
