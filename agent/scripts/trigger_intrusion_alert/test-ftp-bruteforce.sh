#!/bin/bash
#
# FTP 暴力破解检测测试脚本
# 用于触发 detector 插件的 FTP 暴力破解告警（DataType 6002）
#
# 原理：detector 插件监控 /var/log/auth.log（或 vsftpd 日志），
#       统计单位时间内来自同一 IP 的 FTP 登录失败次数，超过阈值时产生告警。
#
# 使用前请先在另一个终端启动 Agent：
#   方式一（standalone）：
#     cd /opt/cloudsec
#     sudo ./bin/agent -standalone -plugins=detector -output=stderr -test
#   方式二（集成测试）：
#     cd /opt/cloudsec
#     sudo ./bin/agent -config agent.yaml -test
#
# 然后在当前终端执行本脚本：
#   sudo bash scripts/test-ftp-bruteforce.sh
#
# 依赖：vsftpd
#   sudo apt install vsftpd && sudo systemctl start vsftpd
#

set -e

ATTEMPTS=10      # 登录尝试次数（默认阈值 6 次）
INTERVAL=1       # 每次尝试之间的间隔秒数
TARGET="localhost"

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

# 检查 vsftpd 服务
if ! systemctl is-active --quiet vsftpd 2>/dev/null; then
    echo -e "${RED}错误：vsftpd 服务未运行${NC}"
    echo "  安装并启动: sudo apt install vsftpd && sudo systemctl start vsftpd"
    exit 1
fi

echo "========================================"
echo " FTP 暴力破解检测 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo ""
echo -e "${YELLOW}注意事项：${NC}"
echo "  1. 默认配置中 127.0.0.1 在 FTP 白名单内，本地测试需先移除白名单"
echo "  2. 远程 server 模式：编辑远程 server.yaml 将 ftp task 的"
echo "     whitelist 改为空数组 []，重启 server 和 Agent"
echo "  3. Standalone 模式：编辑 ftp_brute_force.yaml 将 whitelist 改为 []"
echo ""

# ------------------------------------------
# BF002: FTP 密码错误暴力破解
# 预期: attack_type=ftp  attempt_count >= 6
# ------------------------------------------
echo -e "${YELLOW}[1/1] BF002: FTP 暴力破解（${ATTEMPTS} 次错误密码登录）${NC}"
echo "  目标: ${TARGET}"
echo ""

for i in $(seq 1 "$ATTEMPTS"); do
    echo -e "  [${i}/${ATTEMPTS}] curl -u wronguser:wrongpass ftp://${TARGET}/"
    curl -u wronguser:wrongpass ftp://"${TARGET}"/ 2>/dev/null || true
    sleep "$INTERVAL"
done

echo ""
echo -e "${GREEN}  登录尝试完成${NC}"

# ------------------------------------------
# 汇总
# ------------------------------------------
echo ""
echo "========================================"
echo " 测试完成"
echo "========================================"
echo ""
echo "检测触发通常需要 1-2 分钟（检测器按周期扫描日志）"
echo ""
echo "请确认以下告警："
echo ""
echo -e "  ${RED}[1] BF002: attack_type=ftp  attempt_count>=${ATTEMPTS}  source_ip=127.0.0.1${NC}"
echo ""
echo "如果未触发告警，请检查："
echo "  1. 127.0.0.1 是否仍在白名单中（ftp_brute_force.yaml 默认包含 127.0.0.1）"
echo "  2. vsftpd 是否正在运行: systemctl status vsftpd"
echo "  3. /var/log/auth.log 或 /var/log/vsftpd.log 中是否有登录失败记录"
echo "  4. detector 插件是否已加载 ftp 检测器"
