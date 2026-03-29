#!/bin/bash
#
# Agent 服务管理脚本
# 用于启动/停止/管理 cloudsec-agent
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 目录定义
DEPLOY_DIR="/opt/cloudsec/agent"

# 打印函数
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

step() {
    echo -e "\n${CYAN}========== $1 ==========${NC}\n"
}

# 显示帮助
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Agent 服务管理脚本"
    echo ""
    echo "Options:"
    echo "  -r, --run       启动 Agent"
    echo "  -s, --status    查看 Agent 状态"
    echo "  -k, --kill      停止 Agent"
    echo "  -c, --clean     清理日志和运行时数据"
    echo "  -h, --help      显示帮助"
    echo ""
    echo "Examples:"
    echo "  $0 -r           # 启动 Agent"
    echo "  $0 -s           # 查看 Agent 状态"
    echo "  $0 -k           # 停止 Agent"
    echo "  $0 -c           # 清理日志和数据"
    echo ""
    echo "注意: 请先手动编译和部署 Agent"
    echo "  cd /home/work/goProject/src/BeeGuard/agent && make deploy"
}

# 启动 Agent
run_agent() {
    step "启动 Agent"

    # 检查是否已部署
    if [ ! -f "${DEPLOY_DIR}/bin/agent" ]; then
        error "Agent 未部署，请先手动部署: cd agent && make deploy"
    fi

    # 停止已有进程
    info "停止已有 Agent 进程..."
    sudo pkill -f "${DEPLOY_DIR}/bin/agent" 2>/dev/null || true
    sleep 1

    # 启动 Agent (需要 root 权限)
    info "启动 Agent..."
    sudo bash -c "nohup ${DEPLOY_DIR}/bin/agent -config ${DEPLOY_DIR}/agent.yaml >> ${DEPLOY_DIR}/logs/agent/agent.log 2>&1 &"
    sleep 2

    AGENT_PID=$(pgrep -f "${DEPLOY_DIR}/bin/agent" | head -1)
    if [ -n "$AGENT_PID" ]; then
        echo "$AGENT_PID" | sudo tee /tmp/cloudsec-agent.pid > /dev/null
        success "Agent 已启动 (PID: $AGENT_PID)"
    else
        error "Agent 启动失败，查看日志: ${DEPLOY_DIR}/logs/agent/agent.log"
    fi

    echo ""
    info "查看日志:"
    echo "  Agent: tail -f ${DEPLOY_DIR}/logs/agent/agent.log"
}

# 查看 Agent 状态
show_status() {
    step "Agent 状态"

    echo "Agent:"
    if pgrep -f "${DEPLOY_DIR}/bin/agent" > /dev/null 2>&1; then
        PID=$(pgrep -f "${DEPLOY_DIR}/bin/agent")
        echo -e "  ${GREEN}● 运行中${NC} (PID: $PID)"
    else
        echo -e "  ${RED}○ 未运行${NC}"
    fi
}

# 停止 Agent
kill_agent() {
    step "停止 Agent"

    info "停止 Agent..."
    sudo pkill -f "${DEPLOY_DIR}/bin/agent" 2>/dev/null && success "Agent 已停止" || warn "Agent 未运行"
}

# 清理环境
clean_all() {
    step "清理 Agent 环境"

    # 停止服务
    kill_agent 2>/dev/null || true

    # 清理日志和运行时数据
    info "清理日志..."
    sudo rm -rf "${DEPLOY_DIR}/logs/"* 2>/dev/null || true

    info "清理运行时数据..."
    sudo rm -rf "${DEPLOY_DIR}/data/"* 2>/dev/null || true

    success "清理完成"
}

# 主函数
main() {
    local do_run=false
    local do_clean=false
    local do_status=false
    local do_kill=false

    # 如果没有参数，显示帮助
    if [ $# -eq 0 ]; then
        usage
        exit 0
    fi

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -r|--run)
                do_run=true
                shift
                ;;
            -c|--clean)
                do_clean=true
                shift
                ;;
            -s|--status)
                do_status=true
                shift
                ;;
            -k|--kill)
                do_kill=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                error "未知选项: $1"
                ;;
        esac
    done

    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║   Agent 服务管理脚本                           ║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════╝${NC}"
    echo ""

    # 执行操作
    if $do_status; then
        show_status
        exit 0
    fi

    if $do_kill; then
        kill_agent
        exit 0
    fi

    if $do_clean; then
        clean_all
        exit 0
    fi

    if $do_run; then
        run_agent
    fi

    echo ""
    success "操作完成!"
}

main "$@"
