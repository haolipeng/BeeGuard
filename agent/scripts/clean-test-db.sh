#!/bin/bash
#
# 集成测试前数据库清理脚本
# ���空本地 PostgreSQL soc 数据库中所有表的记录
#
# 用法：
#   bash scripts/clean-test-db.sh
#
# 默认连接本地 PostgreSQL（postgres/root），可通过环境变量覆盖：
#   DB_HOST=192.168.1.100 DB_USER=myuser DB_PASS=mypass bash scripts/clean-test-db.sh
#

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-root}"
DB_NAME="${DB_NAME:-soc}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================"
echo " 集成测试 — 数据库清理"
echo "========================================"
echo ""
echo "  数据库: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "  用户:   ${DB_USER}"
echo ""

# 检查 psql 是否可用
if ! command -v psql &>/dev/null; then
    echo -e "${RED}错误: psql 命令未找到，请先安装 postgresql-client${NC}"
    exit 1
fi

# 测试数据库连接
if ! PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" &>/dev/null; then
    echo -e "${RED}错误: 无法连接到数据库 ${DB_HOST}:${DB_PORT}/${DB_NAME}${NC}"
    echo "请检查："
    echo "  1. PostgreSQL 服务是否已启动"
    echo "  2. 数据库 ${DB_NAME} 是否已创建"
    echo "  3. 用户名和密码是否正确"
    exit 1
fi

echo -e "${GREEN}数据库连接成功${NC}"
echo ""

# 定义所有���要清理的表（按依赖顺序排列）
TABLES=(
    # Collector 资产表
    "asset_process"
    "asset_port"
    "asset_account"
    "asset_system_service"
    "asset_software"
    "asset_kmod"
    "asset_container"
    "asset_image"
    "asset_image_package"
    "asset_web_service"
    "asset_database"
    "asset_env_suspicious"
    # eBPF 事件表
    "event_execve"
    "event_connect"
    "event_dns"
    "event_file"
    # 告警表
    "alert_brute_force"
    "alert_dangerous_command"
    "alert_privilege_escalation"
    "alert_reverse_shell"
    "alert_abnormal_login"
    "alert_malicious_request"
    "alert_malware_scan"
    "alert_network_attack"
    "alert_file_integrity"
    # Baseline 表
    "baseline_check_detail"
    "baseline_check_result"
    # Agent 信息表
    "agent_info"
)

success=0
skipped=0
failed=0

for table in "${TABLES[@]}"; do
    # 检查表是否存在
    exists=$(PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc \
        "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '${table}');")

    if [ "$exists" = "t" ]; then
        if PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
            "TRUNCATE TABLE ${table} CASCADE;" &>/dev/null; then
            echo -e "  ${GREEN}✓${NC} ${table}"
            ((success++))
        else
            echo -e "  ${RED}✗${NC} ${table} (清空失败)"
            ((failed++))
        fi
    else
        echo -e "  ${YELLOW}-${NC} ${table} (表不存在，跳过)"
        ((skipped++))
    fi
done

echo ""
echo "========================================"
echo " 清理完成"
echo "========================================"
echo ""
echo -e "  ${GREEN}成功: ${success}${NC}  ${YELLOW}跳过: ${skipped}${NC}  ${RED}失败: ${failed}${NC}"
echo ""

if [ "$failed" -gt 0 ]; then
    echo -e "${RED}存在清理失败的表，请手动检查${NC}"
    exit 1
fi

echo -e "${GREEN}数据库已清空，可以开始集成测试${NC}"
