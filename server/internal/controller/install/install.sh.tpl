#!/bin/bash
set -e

BASE_URL="{{.BaseURL}}"
GRPC_ADDR="{{.GRPCAddr}}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 检查 root 权限
if [ "$(id -u)" -ne 0 ]; then
    error "请使用 root 权限运行此脚本"
fi

# 检查 systemctl
if ! command -v systemctl &>/dev/null; then
    error "系统不支持 systemctl，无法安装"
fi

# 检测包管理器
PKG_TYPE=""
if command -v dpkg &>/dev/null; then
    PKG_TYPE="deb"
elif command -v rpm &>/dev/null; then
    PKG_TYPE="rpm"
else
    error "未检测到 dpkg 或 rpm 包管理器"
fi
info "检测到包管理器类型: ${PKG_TYPE}"

# 检测架构
ARCH=""
MACHINE=$(uname -m)
case "${MACHINE}" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       error "不支持的架构: ${MACHINE}" ;;
esac
info "检测到系统架构: ${ARCH}"

# 创建临时目录
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

# 下载安装包
PKG_FILE="${TMP_DIR}/cloudsec-agent.${PKG_TYPE}"
DOWNLOAD_URL="${BASE_URL}/api1/agent/download?type=${PKG_TYPE}&arch=${ARCH}"
info "正在下载安装包: ${DOWNLOAD_URL}"

if command -v curl &>/dev/null; then
    curl -fSL -o "${PKG_FILE}" "${DOWNLOAD_URL}"
elif command -v wget &>/dev/null; then
    wget -q -O "${PKG_FILE}" "${DOWNLOAD_URL}"
else
    error "未找到 curl 或 wget，无法下载安装包"
fi
info "安装包下载完成"

# 安装
info "正在安装 Agent..."
export SPECIFIED_SERVER="${GRPC_ADDR}"
if [ "${PKG_TYPE}" = "deb" ]; then
    dpkg -i "${PKG_FILE}"
elif [ "${PKG_TYPE}" = "rpm" ]; then
    rpm -i "${PKG_FILE}" || rpm -U "${PKG_FILE}"
fi

# 验证服务状态
sleep 2
if systemctl is-active --quiet cloudsec-agent; then
    info "Agent 安装成功，服务已启动"
    info "gRPC 服务器地址: ${GRPC_ADDR}"
    systemctl status cloudsec-agent --no-pager
else
    warn "Agent 已安装，但服务可能尚未启动，请检查日志:"
    warn "  journalctl -u cloudsec-agent -f"
fi
