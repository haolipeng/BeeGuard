#!/bin/bash
#
# 恶意请求检测测试脚本
# 用于触发 ebpf_base_detector 插件的恶意请求告警（DataType 6008）
#
# 依赖：netcat (nc)、dig 或 nslookup
#   sudo apt install netcat-openbsd dnsutils
#

set -e

INTERVAL=2  # 每个测试用例之间的等待秒数

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

if ! command -v nc &>/dev/null; then
    echo -e "${RED}错误：未找到 nc，请先安装: sudo apt install netcat-openbsd${NC}"
    exit 1
fi

if ! command -v dig &>/dev/null; then
    echo -e "${YELLOW}警告：未找到 dig，DNS 类测试将跳过。安装: sudo apt install dnsutils${NC}"
    HAS_DIG=false
else
    HAS_DIG=true
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
echo " 恶意请求检测 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo "每个测试用例间隔 ${INTERVAL} 秒"
echo ""

# ------------------------------------------
# IOC002: 常见矿池端口（medium） — port 类型
# 预期: rule_id=IOC002  rule_name=常见矿池端口  threat_type=mining  indicator_type=port
# ------------------------------------------
echo -e "${YELLOW}[1/5] IOC002: 常见矿池端口（medium）${NC}"
echo "  启动本地监听: nc -lvp 3333"
nc -lvp 3333 &>/dev/null &
LISTEN_PIDS+=($!)
sleep 1

echo "  触发连接: nc -w 1 127.0.0.1 3333"
echo "test" | nc -w 1 127.0.0.1 3333 2>/dev/null || true
kill "${LISTEN_PIDS[-1]}" 2>/dev/null || true
wait "${LISTEN_PIDS[-1]}" 2>/dev/null || true
echo -e "  ${GREEN}完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# IOC003: 已知矿池域名（high） — domain 类型
# 预期: rule_id=IOC003  rule_name=已知矿池域名  threat_type=mining  indicator_type=domain
# ------------------------------------------
echo -e "${YELLOW}[2/5] IOC003: 已知矿池域名（high）${NC}"
if $HAS_DIG; then
    echo "  触发 DNS 查询: dig minersns.com"
    dig minersns.com +time=2 +tries=1 2>/dev/null || true
    echo -e "  ${GREEN}完成${NC}"
else
    echo -e "  ${RED}跳过${NC} — 未安装 dig"
fi
sleep "$INTERVAL"

# ------------------------------------------
# IOC004: 已知C2域名（critical） — domain 类型（通配符）
# 预期: rule_id=IOC004  rule_name=已知C2域名  threat_type=c2  indicator_type=domain
# ------------------------------------------
echo -e "${YELLOW}[3/5] IOC004: 已知C2域名（critical）${NC}"
if $HAS_DIG; then
    echo "  触发 DNS 查询: dig test.cobalt-strike.example.com"
    dig test.cobalt-strike.example.com +time=2 +tries=1 2>/dev/null || true
    echo -e "  ${GREEN}完成${NC}"
else
    echo -e "  ${RED}跳过${NC} — 未安装 dig"
fi
sleep "$INTERVAL"

# ------------------------------------------
# IOC005: 已知C2端点（critical） — ip_port 类型
# 预期: rule_id=IOC005  rule_name=已知C2端点  threat_type=c2  indicator_type=ip_port
# 注意: 目标 IP 可能不可达，connect 需成功才触发，此测试在多数环境不会生效
# ------------------------------------------
echo -e "${YELLOW}[4/5] IOC005: 已知C2端点（critical）${NC}"
echo "  触发连接: nc -w 2 185.141.27.100 443（目标可能不可达）"
nc -w 2 185.141.27.100 443 2>/dev/null || true
echo -e "  ${GREEN}完成${NC} — 目标不可达时不触发告警（connect 需成功）"
sleep "$INTERVAL"

# ------------------------------------------
# IOC006: 已知钓鱼域名（high） — domain 类型（通配符）
# 预期: rule_id=IOC006  rule_name=已知钓鱼域名  threat_type=phishing  indicator_type=domain
# ------------------------------------------
echo -e "${YELLOW}[5/5] IOC006: 已知钓鱼域名（high）${NC}"
if $HAS_DIG; then
    echo "  触发 DNS 查询: dig login.phishing-example.com"
    dig login.phishing-example.com +time=2 +tries=1 2>/dev/null || true
    echo -e "  ${GREEN}完成${NC}"
else
    echo -e "  ${RED}跳过${NC} — 未安装 dig"
fi

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
echo -e "  ${RED}[1] IOC002: rule_id=IOC002  rule_name=常见矿池端口     threat_type=mining       indicator_type=port${NC}"
if $HAS_DIG; then
echo -e "  ${RED}[2] IOC003: rule_id=IOC003  rule_name=已知矿池域名     threat_type=mining       indicator_type=domain${NC}"
echo -e "  ${RED}[3] IOC004: rule_id=IOC004  rule_name=已知C2域名        threat_type=c2           indicator_type=domain${NC}"
else
echo -e "  [2] IOC003: 跳过（未安装 dig）"
echo -e "  [3] IOC004: 跳过（未安装 dig）"
fi
echo -e "  ${RED}[4] IOC005: rule_id=IOC005  rule_name=已知C2端点        threat_type=c2           indicator_type=ip_port${NC}"
if $HAS_DIG; then
echo -e "  ${RED}[5] IOC006: rule_id=IOC006  rule_name=已知钓鱼域名     threat_type=phishing     indicator_type=domain${NC}"
else
echo -e "  [5] IOC006: 跳过（未安装 dig）"
fi
echo ""
echo "注意事项："
echo "  - connect 类规则仅在连接成功（retval == 0）时触发"
echo "  - IOC002（端口匹配）使用本地监听确保触发"
echo "  - IOC005 目标不可达时不会触发，可将 127.0.0.1 添加到规则 indicators 中测试"
echo "  - DNS 类规则（IOC003/004/006）依赖 eBPF 捕获 recvfrom/recvmsg"
echo ""
