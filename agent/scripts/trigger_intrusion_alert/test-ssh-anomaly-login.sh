#!/bin/bash
#
# SSH 异常登录检测测试脚本
# 用于触发 detector 插件的 SSH 异常登录告警（DataType 6005）
#
# 原理：detector 插件的 ssh_anomaly_login 检测器采用白名单机制，
#       从不在白名单中的 IP 成功登录 SSH 时触发告警。
#
# 前置条件（三项全部满足才能触发告警）：
#   1. Agent 端：/opt/cloudsec/plugins/detector/config/rules/ssh_anomaly_login.yaml
#      中 enabled=true 且 anomaly_rules 至少有一条含 IP 的规则
#   2. 远程 server 模式：server.yaml 中 ssh_anomaly_login 的 enabled=true 且有规则
#      （否则服务端配置会覆盖本地配置）
#   3. 确认 detector 日志出现: "compiled N IPs from M rules"（N > 0, M > 0）
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
#   sudo bash scripts/test-ssh-anomaly-login.sh
#

set -e

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

# 获取本机非回环 IP
LOCAL_IP=$(hostname -I | awk '{print $1}')
if [ -z "$LOCAL_IP" ]; then
    echo -e "${RED}错误：无法获取本机 IP${NC}"
    exit 1
fi

echo "========================================"
echo " SSH 异常登录检测 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo ""
echo -e "${YELLOW}前置条件检查清单：${NC}"
echo "  1. ssh_anomaly_login.yaml 中 enabled=true"
echo "  2. anomaly_rules 中配置了可信 IP 白名单（不包含 ${LOCAL_IP}）"
echo "  3. detector 日志出现: compiled N IPs from M rules（N > 0）"
echo ""
echo "本机 IP: ${LOCAL_IP}"
echo ""

# 检查 detector 日志（可选）
DETECTOR_LOG="/opt/cloudsec/agent/logs/plugins/detector/detector.log"
if [ -f "$DETECTOR_LOG" ]; then
    COMPILED_LINE=$(grep "compiled.*IPs from.*rules" "$DETECTOR_LOG" | tail -1)
    if [ -n "$COMPILED_LINE" ]; then
        echo -e "  ${GREEN}检测器状态: ${COMPILED_LINE}${NC}"
    else
        echo -e "  ${RED}警告: 未在日志中找到规则加载记录，检测器可能未生效${NC}"
    fi
else
    echo -e "  ${YELLOW}跳过日志检查: ${DETECTOR_LOG} 不存在${NC}"
fi
echo ""

# ------------------------------------------
# AL001: 从非白名单 IP 成功 SSH 登录
# 预期: source_ip 不在白名单  risk_level=critical
# ------------------------------------------
echo -e "${YELLOW}[1/1] AL001: 从非白名单 IP 成功 SSH 登录${NC}"

# 确保 SSH 密钥认证可用
if [ ! -f ~/.ssh/id_rsa ]; then
    echo "  生成 SSH 密钥对..."
    ssh-keygen -t rsa -f ~/.ssh/id_rsa -N "" -q
fi

# 添加公钥到 authorized_keys
if ! grep -qF "$(cat ~/.ssh/id_rsa.pub)" ~/.ssh/authorized_keys 2>/dev/null; then
    cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
    chmod 600 ~/.ssh/authorized_keys
fi

echo "  执行: ssh -i ~/.ssh/id_rsa root@${LOCAL_IP} 'echo login success'"
ssh -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa root@"${LOCAL_IP}" "echo 'login success'" 2>/dev/null || true
echo -e "${GREEN}  完成${NC}"

# ------------------------------------------
# 汇总
# ------------------------------------------
echo ""
echo "========================================"
echo " 测试完成"
echo "========================================"
echo ""
echo "请确认以下告警："
echo ""
echo -e "  ${RED}[1] AL001: source_ip=${LOCAL_IP}  risk_level=critical  login_user=root${NC}"
echo ""
echo "如果未触发告警，请检查："
echo "  1. ssh_anomaly_login 检测器是否已启用（日志中搜索 ssh_anomaly）"
echo "  2. ${LOCAL_IP} 是否不在白名单 anomaly_rules 的 ips 列表中"
echo "  3. 远程 server 模式下服务端配置是否覆盖了本地 enabled=true"
