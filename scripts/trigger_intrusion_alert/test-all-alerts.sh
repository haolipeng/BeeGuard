#!/bin/bash
#
# 告警触发统一入口脚本
# 按顺序调用各告警检测测试脚本，触发入侵检测告警
#
# 用法：
#   sudo bash scripts/test-all-alerts.sh              # 运行所有测试
#   sudo bash scripts/test-all-alerts.sh ebpf          # 仅 ebpf_base_detector 相关
#   sudo bash scripts/test-all-alerts.sh detector      # 仅 detector 相关
#   sudo bash scripts/test-all-alerts.sh nids          # 仅 nids 相关
#   sudo bash scripts/test-all-alerts.sh scanner       # 仅 scanner 相关
#
# 使用前请先在另一个终端启动 Agent：
#   集成测试模式：
#     cd /opt/cloudsec && sudo ./bin/agent -config agent.yaml -test
#   Standalone 模式（按需选择插件）：
#     cd /opt/cloudsec && sudo ./bin/agent -standalone -plugins=ebpf_base_detector,detector,nids -output=stderr -test
#

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

TOTAL=0
PASS=0
SKIP=0
FAIL=0

# ------------------------------------------
# 前置检查
# ------------------------------------------
if [ "$(id -u)" -ne 0 ]; then
    echo -e "${RED}错误：本脚本需要 root 权限运行${NC}"
    echo "  用法: sudo bash $0 [all|ebpf|detector|nids|scanner]"
    exit 1
fi

# ------------------------------------------
# 工具函数
# ------------------------------------------

# 执行单个测试脚本
# 参数: $1=脚本路径  $2=描述  $3...=额外参数
run_test() {
    local script="$1"
    local desc="$2"
    shift 2
    local extra_args=("$@")

    TOTAL=$((TOTAL + 1))

    if [ ! -f "$script" ]; then
        echo -e "  ${RED}[跳过]${NC} ${desc} — 脚本不存在: ${script}"
        SKIP=$((SKIP + 1))
        return
    fi

    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${desc}${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    if bash "$script" "${extra_args[@]}"; then
        echo ""
        echo -e "  ${GREEN}=> ${desc} 执行完成${NC}"
        PASS=$((PASS + 1))
    else
        echo ""
        echo -e "  ${RED}=> ${desc} 执行失败（可能缺少依赖，继续下一项）${NC}"
        FAIL=$((FAIL + 1))
    fi
    echo ""
}

# 分隔符
section_header() {
    echo ""
    echo -e "${BOLD}╔══════════════════════════════════════════╗${NC}"
    echo -e "${BOLD}║  $1${NC}"
    echo -e "${BOLD}╚══════════════════════════════════════════╝${NC}"
    echo ""
}

# ------------------------------------------
# 测试组定义
# ------------------------------------------

run_ebpf_tests() {
    section_header "ebpf_base_detector 插件测试"

    run_test "${SCRIPT_DIR}/test-dangerous-commands.sh" \
        "[eBPF] 高危命令检测 (DataType 6003)"

    run_test "${SCRIPT_DIR}/test-privilege-escalation.sh" \
        "[eBPF] 本地提权检测 (DataType 6006)"

    run_test "${SCRIPT_DIR}/test-reverse-shell.sh" \
        "[eBPF] 反弹Shell检测 (DataType 6004)"

    run_test "${SCRIPT_DIR}/test-malicious-requests.sh" \
        "[eBPF] 恶意请求检测 (DataType 6008)"

    run_test "${SCRIPT_DIR}/test-file-integrity.sh" \
        "[eBPF] 文件完整性告警 (DataType 6009)"
}

run_detector_tests() {
    section_header "Detector 插件测试"

    run_test "${SCRIPT_DIR}/test-ssh-bruteforce.sh" \
        "[Detector] SSH暴力破解检测 (DataType 6001)"

    run_test "${SCRIPT_DIR}/test-ftp-bruteforce.sh" \
        "[Detector] FTP暴力破解检测 (DataType 6002)"

    run_test "${SCRIPT_DIR}/test-ssh-anomaly-login.sh" \
        "[Detector] SSH异常登录检测 (DataType 6005)"
}

run_nids_tests() {
    section_header "NIDS 插件测试"

    run_test "${SCRIPT_DIR}/test-nids.sh" \
        "[NIDS] 网络攻击检测 (DataType 6007)"
}

run_scanner_tests() {
    section_header "Scanner 插件测试"

    echo -e "${YELLOW}注意：scanner 测试仅创建 EICAR 文件，需在启动 Agent 前执行。${NC}"
    echo -e "${YELLOW}      若 Agent 已在运行，需重启 Agent 以触发自动扫描。${NC}"
    echo ""

    run_test "${SCRIPT_DIR}/test-scanner.sh" \
        "[Scanner] 恶意软件扫描 (DataType 6061)" \
        "prepare"
}

# ------------------------------------------
# 主逻辑
# ------------------------------------------

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║     告警触发 — 统一测试入口              ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo "请确认 Agent 已在另一个终端启动"
echo ""

TARGET="${1:-all}"

case "$TARGET" in
    all)
        echo "测试范围: 全部（ebpf + detector + nids + scanner）"
        echo ""
        run_ebpf_tests
        run_detector_tests
        run_nids_tests
        run_scanner_tests
        ;;
    ebpf)
        echo "测试范围: ebpf_base_detector"
        echo ""
        run_ebpf_tests
        ;;
    detector)
        echo "测试范围: detector"
        echo ""
        run_detector_tests
        ;;
    nids)
        echo "测试范围: nids"
        echo ""
        run_nids_tests
        ;;
    scanner)
        echo "测试范围: scanner"
        echo ""
        run_scanner_tests
        ;;
    *)
        echo "用法: sudo bash $0 [all|ebpf|detector|nids|scanner]"
        echo ""
        echo "  all       运行所有测试（默认）"
        echo "  ebpf      高危命令 + 提权 + 反弹Shell + 恶意请求 + 文件完整性"
        echo "  detector  SSH暴力破解 + FTP暴力破解 + SSH异常登录"
        echo "  nids      网络攻击检测（Log4j/SQLi/CMDi 等）"
        echo "  scanner   恶意软件扫描（创建 EICAR 测试文件）"
        exit 1
        ;;
esac

# ------------------------------------------
# 汇总
# ------------------------------------------
echo ""
echo "╔══════════════════════════════════════════╗"
echo "║     全部测试执行完毕                      ║"
echo "╚══════════════════════════════════════════╝"
echo ""
echo -e "  总计: ${TOTAL}  ${GREEN}完成: ${PASS}${NC}  ${YELLOW}跳过: ${SKIP}${NC}  ${RED}失败: ${FAIL}${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
    echo -e "${YELLOW}有 ${FAIL} 个测试组执行失败，通常是缺少依赖（sshpass/vsftpd/nc 等）。${NC}"
    echo "可单独运行失败的脚本查看具体错误。"
    echo ""
fi

echo "后续操作："
echo "  1. 等待 1-2 分钟（detector 类告警有检测周期延迟）"
echo "  2. 查看 Agent 终端输出或日志确认告警"
echo "  3. 集成测试模式下查询远程数据库 alert_* 表验证数据写入"
if [ "$TARGET" = "all" ] || [ "$TARGET" = "scanner" ]; then
    echo "  4. scanner 测试文件清理: sudo bash ${SCRIPT_DIR}/test-scanner.sh cleanup"
fi
echo ""
