#!/bin/bash
#
# 恶意软件扫描测试脚本
# 用于触发 scanner 插件的恶意软件检测告警（DataType 6061/6062）
#
# 原理：scanner 插件使用 ClamAV 引擎扫描指定目录，检测木马、Webshell、挖矿程序等。
#       Agent 连接 server 后，服务端自动下发目录扫描任务（默认扫描 /root、/etc、/var/www）。
#       本脚本在扫描目录中创建 EICAR 标准测试文件，等待 scanner 自动检出。
#
# 前置条件：
#   1. ClamAV 已安装: sudo apt install clamav libclamav-dev clamav-freshclam
#   2. 病毒库已更新: sudo freshclam
#   3. 病毒库文件位于: /var/lib/clamav/
#
# 使用方式（集成测试）：
#   1. 先执行本脚本创建测试文件
#   2. 再启动 Agent（Agent 连接后 server 自动下发扫描任务）
#   3. 等待扫描完成后查看告警
#
#   sudo bash scripts/test-scanner.sh prepare   # 创建测试文件
#   sudo bash scripts/test-scanner.sh cleanup   # 清理测试文件
#

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

EICAR_STRING='X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*'
EICAR_MD5="44d88612fea8a8f36de82e1278abb02f"

# 测试文件列表（放在 server 默认扫描目录 /root 下）
TEST_FILES=(
    "/root/eicar_test.com"
    "/root/eicar_1.exe"
    "/root/eicar_2.sh"
)

# ------------------------------------------
# 前置检查
# ------------------------------------------
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}错误：本脚本需要 root 权限运行${NC}"
    echo "  用法: sudo bash $0 [prepare|cleanup]"
    exit 1
fi

# ------------------------------------------
# 子命令：cleanup — 清理测试文件
# ------------------------------------------
do_cleanup() {
    echo -e "${YELLOW}[清理] 删除 EICAR 测试文件${NC}"
    for f in "${TEST_FILES[@]}"; do
        if [ -f "$f" ]; then
            rm -f "$f"
            echo -e "  ${GREEN}已删除${NC}: $f"
        else
            echo -e "  ${YELLOW}不存在${NC}: $f"
        fi
    done
    echo -e "${GREEN}清理完成${NC}"
}

# ------------------------------------------
# 子命令：prepare — 创建测试文件
# ------------------------------------------
do_prepare() {
    echo "========================================"
    echo " 恶意软件扫描 — 自动化测试"
    echo "========================================"
    echo ""

    # 检查 ClamAV
    if [ ! -d "/var/lib/clamav" ]; then
        echo -e "${YELLOW}警告：/var/lib/clamav 不存在，ClamAV 可能未安装${NC}"
        echo "  安装: sudo apt install clamav libclamav-dev clamav-freshclam"
        echo "  更新病毒库: sudo freshclam"
        echo ""
    fi

    echo -e "${YELLOW}[准备] 创建 EICAR 标准测试文件${NC}"
    echo ""

    for f in "${TEST_FILES[@]}"; do
        printf '%s' "$EICAR_STRING" > "$f"
        echo -e "  ${GREEN}已创建${NC}: $f"
    done

    echo ""
    echo -e "${GREEN}测试文件已就绪${NC}"
    echo ""
    echo "后续步骤："
    echo "  1. 启动 Agent 连接 server（scanner 插件会自动接收扫描任务）"
    echo "  2. 等待约 30 秒，scanner 扫描 /root 目录"
    echo "  3. 查询 alert_malware_scan 表验证检测结果"
    echo ""
    echo "预期结果："
    echo ""
    echo -e "  ${RED}[1] file_path=/root/eicar_test.com  file_md5=${EICAR_MD5}  detection_engine=ClamAV${NC}"
    echo -e "  ${RED}[2] file_path=/root/eicar_1.exe     file_md5=${EICAR_MD5}  detection_engine=ClamAV${NC}"
    echo -e "  ${RED}[3] file_path=/root/eicar_2.sh      file_md5=${EICAR_MD5}  detection_engine=ClamAV${NC}"
    echo ""
    echo "测试完成后执行清理："
    echo "  sudo bash $0 cleanup"
}

# ------------------------------------------
# 主逻辑
# ------------------------------------------
case "${1:-prepare}" in
    prepare)
        do_prepare
        ;;
    cleanup)
        do_cleanup
        ;;
    *)
        echo "用法: sudo bash $0 [prepare|cleanup]"
        echo ""
        echo "  prepare  创建 EICAR 测试文件（默认）"
        echo "  cleanup  清理测试文件"
        exit 1
        ;;
esac
