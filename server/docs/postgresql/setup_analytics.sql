-- =====================================================
-- 快速初始化概览页统计数据
-- 数据库: PostgreSQL
-- 执行方式: psql -U postgres -d soc -f setup_analytics.sql
-- =====================================================

\echo '1. 创建分析视图...'
\i init_analytics_views.sql

\echo '2. 插入主机资产模拟数据...'
\i mock_data/001_mock_asset_host.sql

\echo '3. 插入告警模拟数据...'
\i mock_data/alert_data/013_mock_alert_brute_force.sql
\i mock_data/alert_data/014_mock_alert_dangerous_command.sql
\i mock_data/alert_data/015_mock_alert_reverse_shell.sql
\i mock_data/alert_data/016_mock_alert_privilege_escalation.sql
\i mock_data/alert_data/017_mock_alert_abnormal_login.sql
\i mock_data/alert_data/018_mock_alert_malicious_request.sql
\i mock_data/alert_data/019_mock_alert_network_attack.sql
\i mock_data/alert_data/020_mock_alert_malware_scan.sql
\i mock_data/alert_data/021_mock_alert_file_integrity.sql

\echo '4. 插入漏洞模拟数据...'
\i mock_data/vuln_data/026_mock_vuln_info.sql
\i mock_data/vuln_data/027_mock_host_vuln_scan.sql
\i mock_data/vuln_data/028_mock_host_vuln_detail.sql
\i mock_data/vuln_data/029_mock_image_vuln_scan.sql
\i mock_data/vuln_data/030_mock_image_vuln_detail.sql

\echo '5. 插入基线检查模拟数据...'
\i mock_data/baseline_data/022_mock_baseline_template.sql
\i mock_data/baseline_data/023_mock_baseline_check_item.sql
\i mock_data/baseline_data/024_mock_baseline_check_result.sql
\i mock_data/baseline_data/025_mock_baseline_check_detail.sql

\echo ''
\echo '初始化完成！验证数据:'
SELECT '主机数量' as metric, COUNT(*)::text as value FROM asset_host
UNION ALL
SELECT '漏洞数量', COUNT(*)::text FROM host_vuln_detail WHERE status = 0
UNION ALL
SELECT '告警数量', COUNT(*)::text FROM v_alert_unified;
