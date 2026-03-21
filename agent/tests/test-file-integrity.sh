#!/bin/bash
#
# 文件完整性告警测试脚本
# 用于触发 ebpf_base_detector 插件的文件完整性告警（DataType 6009）
#
# 原理：eBPF 监控敏感文件的创建、修改、删除操作，匹配文件监控规则时产生告警。
#
# 使用前请先在另一个终端启动 Agent：
#   cd /opt/cloudsec
#   sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test
#
# 然后在当前终端执行本脚本：
#   sudo bash scripts/test-file-integrity.sh
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
echo " 文件完整性告警 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo "每个测试用例间隔 ${INTERVAL} 秒"
echo ""

# ------------------------------------------
# FI001: 向 crontab 目录写入文件（create）
# 预期: threat_action=create  file_path=/etc/cron.d/ebpf_test_cron
# ------------------------------------------
echo -e "${YELLOW}[1/4] FI001: crontab 目录创建文件（应触发 create 告警）${NC}"
echo "  执行: echo '# test' > /etc/cron.d/ebpf_test_cron"
echo "# ebpf file integrity test" > /etc/cron.d/ebpf_test_cron
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# FI002: 重命名 crontab 目录文件（rename）
# 预期: threat_action=rename  file_path=/etc/cron.d/ebpf_test_cron_renamed
# ------------------------------------------
echo -e "${YELLOW}[2/4] FI002: crontab 目录重命名文件（应触发 rename 告警）${NC}"
echo "  执行: mv /etc/cron.d/ebpf_test_cron /etc/cron.d/ebpf_test_cron_renamed"
mv /etc/cron.d/ebpf_test_cron /etc/cron.d/ebpf_test_cron_renamed
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# FI003: 删除 crontab 目录文件（delete）
# 预期: threat_action=delete  file_path=/etc/cron.d/ebpf_test_cron_renamed
# ------------------------------------------
echo -e "${YELLOW}[3/4] FI003: crontab 目录删除文件（应触发 delete 告警）${NC}"
echo "  执行: rm /etc/cron.d/ebpf_test_cron_renamed"
rm -f /etc/cron.d/ebpf_test_cron_renamed
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# FI004: 修改 /etc/hosts 文件（modify）
# 预期: threat_action=modify  file_path=/etc/hosts
# ------------------------------------------
echo -e "${YELLOW}[4/4] FI004: 修改 /etc/hosts（应触发 modify 告警）${NC}"
echo "  执行: 追加测试行到 /etc/hosts 后恢复"
# 备份并追加测试行
cp /etc/hosts /etc/hosts.ebpf_test_bak
echo "# ebpf_file_integrity_test" >> /etc/hosts
sleep 1
# 恢复原文件
mv /etc/hosts.ebpf_test_bak /etc/hosts
echo -e "${GREEN}  完成（已恢复原文件）${NC}"

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
echo -e "  ${RED}[1] FI001: threat_action=create  file_path=/etc/cron.d/ebpf_test_cron${NC}"
echo -e "  ${RED}[2] FI002: threat_action=rename  file_path=/etc/cron.d/ebpf_test_cron_renamed${NC}"
echo -e "  ${RED}[3] FI003: threat_action=delete  file_path=/etc/cron.d/ebpf_test_cron_renamed${NC}"
echo -e "  ${RED}[4] FI004: threat_action=rename  file_path=/etc/hosts${NC}"
