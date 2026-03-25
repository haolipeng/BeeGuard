#!/bin/bash
# 基线数据验证脚本
# 用法: bash check_data.sh

HOST="54.179.163.116"
PORT="5432"
USER="user_daEJ8N"
PASSWORD="password_72kmbz"
DB="soc"

export PGPASSWORD="$PASSWORD"
export PAGER=cat

echo "========== 1. baseline_template (id=1400) =========="
psql -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c \
  "SELECT id, baseline_name, baseline_type, os_type, version, item_count, is_enabled FROM baseline_template WHERE id = 1400;"

echo ""
echo "========== 2. baseline_check_item 列表 =========="
psql -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c \
  "SELECT id, item_name, category, risk_level FROM baseline_check_item WHERE baseline_id = 1400 ORDER BY id;"

echo ""
echo "========== 3. 第一条 check_item 完整内容 =========="
psql -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c \
  "SELECT id, item_name, check_rules, fix_suggestion FROM baseline_check_item WHERE baseline_id = 1400 ORDER BY id LIMIT 1;"

echo ""
echo "========== 4. check_item 总数统计 =========="
psql -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c \
  "SELECT COUNT(*) AS total FROM baseline_check_item WHERE baseline_id = 1400;"

echo ""
echo "========== 5. baseline_template_host_link (template_id=1400) =========="
psql -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -c \
  "SELECT id, baseline_template_id, baseline_template_name, target_range, scan_frequency, created_at, updated_at FROM baseline_template_host_link WHERE baseline_template_id = 1400;"

unset PGPASSWORD PAGER
