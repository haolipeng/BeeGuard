#!/bin/bash
# =====================================================
# 初始化概览页统计数据的脚本
# 数据库: PostgreSQL
# 使用方法: ./setup_analytics.sh
# =====================================================

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="soc"
DB_USER="postgres"

echo "正在初始化概览页统计数据..."

# 1. 创建视图
echo "1. 创建分析视图..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ./init_analytics_views.sql

# 2. 插入主机资产模拟数据 (如果表为空)
echo "2. 检查并插入主机资产数据..."
HOST_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM asset_host;")
if [ "$HOST_COUNT" -eq 0 ]; then
    echo "   插入主机资产模拟数据..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ./mock_data/001_mock_asset_host.sql
else
    echo "   主机资产表已有 $HOST_COUNT 条数据，跳过"
fi

# 3. 插入告警模拟数据 (如果表为空)
echo "3. 检查并插入告警数据..."
for alert_file in ./mock_data/alert_data/*.sql; do
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$alert_file" 2>/dev/null
done

# 4. 插入漏洞模拟数据 (如果表为空)
echo "4. 检查并插入漏洞数据..."
VULN_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM vuln_info;")
if [ "$VULN_COUNT" -eq 0 ]; then
    echo "   插入漏洞信息数据..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ./mock_data/vuln_data/026_mock_vuln_info.sql
fi

SCAN_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM host_vuln_scan_task;")
if [ "$SCAN_COUNT" -eq 0 ]; then
    echo "   插入主机漏洞扫描任务数据..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ./mock_data/vuln_data/027_mock_host_vuln_scan.sql
fi

DETAIL_COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM host_vuln_detail;")
if [ "$DETAIL_COUNT" -eq 0 ]; then
    echo "   插入主机漏洞详情数据..."
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f ./mock_data/vuln_data/028_mock_host_vuln_detail.sql
fi

# 5. 刷新视图
echo "5. 刷新物化视图（如有）..."

echo ""
echo "初始化完成！"
echo ""
echo "验证数据："
echo "主机数量: $(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM asset_host;")"
echo "漏洞数量: $(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM host_vuln_detail WHERE status = 0;")"
echo "告警数量: $(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM v_alert_unified;")"
