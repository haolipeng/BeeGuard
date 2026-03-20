#!/bin/bash
#
# 高危命令检测测试脚本
# 用于触发 ebpf_base_detector 插件的高危命令告警
#
# 使用前请先在另一个终端启动 Agent：
#   cd /opt/cloudsec
#   sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test
#
# 然后在当前终端执行本脚本：
#   sudo bash scripts/test-dangerous-commands.sh
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

echo "========================================"
echo " 高危命令检测 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo "每个测试用例间隔 ${INTERVAL} 秒"
echo ""

# ------------------------------------------
# 2001: 危险删除操作（critical）
# 预期: rule_id=2001  rule_name=危险删除操作  severity=critical
# ------------------------------------------
echo -e "${YELLOW}[1/4] 2001: 危险删除操作（critical）${NC}"
echo "  执行: rm -rf /tmp/dc001_nonexistent_test_dir"
rm -rf /tmp/dc001_nonexistent_test_dir
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# 2002: 敏感文件访问（high）
# 预期: rule_id=2002  rule_name=敏感文件访问  severity=high
# ------------------------------------------
echo -e "${YELLOW}[2/4] 2002: 敏感文件访问（high）${NC}"
echo "  执行: cat /etc/passwd > /dev/null"
cat /etc/passwd > /dev/null
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# 2003: 危险权限修改（high）
# 预期: rule_id=2003  rule_name=危险权限修改  severity=high
# ------------------------------------------
echo -e "${YELLOW}[3/4] 2003: 危险权限修改（high）${NC}"
echo "  执行: chmod 777 /tmp/dc003_test"
touch /tmp/dc003_test && chmod 777 /tmp/dc003_test && rm -f /tmp/dc003_test
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# 2009: 内核模块操作（high）
# 预期: rule_id=2009  rule_name=内核模块操作  severity=high
# ------------------------------------------
echo -e "${YELLOW}[4/4] 2009: 内核模块操作（high）${NC}"
echo "  执行: insmod /tmp/nonexistent.ko"
insmod /tmp/nonexistent.ko 2>/dev/null; true
echo -e "${GREEN}  完成${NC}"

# ------------------------------------------
# 清理与汇总
# ------------------------------------------
echo ""
echo "========================================"
echo " 测试完成"
echo "========================================"
echo ""
echo "请在 Agent 终端确认以下告警："
echo ""
echo -e "  ${RED}[1] rule_id=2001  rule_name=危险删除操作     severity=critical${NC}"
echo -e "  ${RED}[2] rule_id=2002  rule_name=敏感文件访问     severity=high${NC}"
echo -e "  ${RED}[3] rule_id=2003  rule_name=危险权限修改     severity=high${NC}"
echo -e "  ${RED}[4] rule_id=2009  rule_name=内核模块操作     severity=high${NC}"
