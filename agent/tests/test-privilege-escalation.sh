#!/bin/bash
#
# 本地提权检测测试脚本
# 用于触发 ebpf_base_detector 插件的本地提权告警（DataType 6006）
#
# 使用前请先在另一个终端启动 Agent：
#   cd /opt/cloudsec
#   sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test
#
# 然后在当前终端执行本脚本：
#   sudo bash scripts/test-privilege-escalation.sh
#

set -e

INTERVAL=3  # 每个测试用例之间的等待秒数
SUID_WRAPPER=/tmp/suid_wrapper
SUID_WRAPPER_SRC=/tmp/suid_wrapper.c

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

if ! command -v gcc &>/dev/null; then
    echo -e "${RED}错误：未找到 gcc，请先安装: apt install gcc${NC}"
    exit 1
fi

# 自动查找非 root 用户（UID >= 1000）
TEST_USER=$(awk -F: '$3 >= 1000 && $3 < 65534 {print $1; exit}' /etc/passwd)
if [ -z "$TEST_USER" ]; then
    echo -e "${RED}错误：未找到可用的非 root 用户${NC}"
    exit 1
fi
TEST_UID=$(id -u "$TEST_USER")

echo "========================================"
echo " 本地提权检测 — 自动化测试"
echo "========================================"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo "测试用户: ${TEST_USER} (uid=${TEST_UID})"
echo ""

# ------------------------------------------
# 准备：编译 SUID 测试程序
# ------------------------------------------
echo -e "${YELLOW}[准备] 编译 SUID 测试程序${NC}"
echo -e "  ${RED}注意: chown/chmod 会触发 DC003 告警（data_type=6003），属于预期行为，请忽略${NC}"

cat > "$SUID_WRAPPER_SRC" << 'EOF'
#include <unistd.h>
#include <stdio.h>
int main() {
    printf("uid=%d euid=%d\n", getuid(), geteuid());
    return 0;
}
EOF

gcc -o "$SUID_WRAPPER" "$SUID_WRAPPER_SRC"
chown root:root "$SUID_WRAPPER"
chmod 4755 "$SUID_WRAPPER"
echo -e "${GREEN}  编译完成: ${SUID_WRAPPER}（SUID 已设置）${NC}"
echo ""

# ------------------------------------------
# 用例一：SUID 程序提权（应触发告警）
# 预期: Privilege escalation detected, exe_path=/tmp/suid_wrapper
# ------------------------------------------
echo -e "${YELLOW}[1/3] SUID 程序提权（应触发告警）${NC}"
echo "  执行: su - ${TEST_USER} -c '${SUID_WRAPPER}'"
su - "$TEST_USER" -c "$SUID_WRAPPER" 2>/dev/null || true
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# 用例二：白名单验证 — sudo（不应触发告警）
# 预期: sudo 在白名单中，不应出现 Privilege escalation detected
# ------------------------------------------
echo -e "${YELLOW}[2/3] 白名单验证 — sudo（不应触发告警）${NC}"
echo "  执行: sudo id"
sudo id > /dev/null 2>&1
echo -e "${GREEN}  完成${NC}"
sleep "$INTERVAL"

# ------------------------------------------
# 用例三：白名单验证 — su（不应触发告警）
# 预期: su 在白名单中，不应出现 Privilege escalation detected
# ------------------------------------------
echo -e "${YELLOW}[3/3] 白名单验证 — su（不应触发告警）${NC}"
echo "  执行: su - root -c 'id'"
su - root -c "id" > /dev/null 2>&1
echo -e "${GREEN}  完成${NC}"

# ------------------------------------------
# 清理与汇总
# ------------------------------------------
echo ""
echo -e "${YELLOW}[清理] 删除测试文件${NC}"
rm -f "$SUID_WRAPPER" "$SUID_WRAPPER_SRC"
echo -e "${GREEN}  已清理${NC}"

echo ""
echo "========================================"
echo " 测试完成"
echo "========================================"
echo ""
echo "请在 Agent 终端确认以下结果："
echo ""
echo -e "  ${RED}[1] 应触发:   Privilege escalation detected  exe_path=${SUID_WRAPPER}  old_uid=${TEST_UID}  new_uid=0${NC}"
echo -e "  ${GREEN}[2] 不应触发: sudo id（在白名单中）${NC}"
echo -e "  ${GREEN}[3] 不应触发: su（在白名单中）${NC}"
