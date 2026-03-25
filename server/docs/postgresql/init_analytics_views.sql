-- =====================================================
-- Analytics Dashboard Views
-- 概览页所需的数据库视图
-- =====================================================

-- 1. 主机统计视图 (v_asset_host_stats)
CREATE OR REPLACE VIEW v_asset_host_stats AS
SELECT
    COUNT(*) AS today_total,
    COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) AS yesterday_total,
    COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) - COUNT(CASE WHEN created_at >= CURRENT_DATE - 1 AND created_at < CURRENT_DATE THEN 1 END) AS net_increase,
    CASE
        WHEN COUNT(CASE WHEN created_at >= CURRENT_DATE - 1 AND created_at < CURRENT_DATE THEN 1 END) = 0 THEN 0
        ELSE ROUND((COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END)::NUMERIC / NULLIF(COUNT(CASE WHEN created_at >= CURRENT_DATE - 1 AND created_at < CURRENT_DATE THEN 1 END), 0)) * 100, 2)
    END AS growth_percentage
FROM asset_host;

COMMENT ON VIEW v_asset_host_stats IS '主机统计视图-用于概览页';

-- 2. 主机风险资产统计视图 (v_views_host_vuln_stats)
CREATE OR REPLACE VIEW v_views_host_vuln_stats AS
SELECT
    COUNT(DISTINCT hd.agent_id) AS total_count,
    COUNT(DISTINCT CASE WHEN hd.scan_time >= CURRENT_DATE THEN hd.agent_id END) AS today_count,
    CASE
        WHEN COUNT(DISTINCT hd.agent_id) = 0 THEN 0
        ELSE ROUND((COUNT(DISTINCT CASE WHEN hd.scan_time >= CURRENT_DATE THEN hd.agent_id END)::NUMERIC / NULLIF(COUNT(DISTINCT hd.agent_id), 0)) * 100, 2)
    END AS percentage
FROM host_vuln_detail hd
WHERE hd.status = 0;

COMMENT ON VIEW v_views_host_vuln_stats IS '主机风险资产统计';

-- 3. 主机漏洞每日统计视图 (v_views_host_vuln_daily_stats)
CREATE OR REPLACE VIEW v_views_host_vuln_daily_stats AS
SELECT
    COUNT(*) AS total_vuln_count,
    COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) AS today_new_count,
    CASE
        WHEN COUNT(*) = 0 THEN 0
        ELSE ROUND((COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END)::NUMERIC / NULLIF(COUNT(*), 0)) * 100, 2)
    END AS today_new_percentage
FROM host_vuln_detail
WHERE status = 0;

COMMENT ON VIEW v_views_host_vuln_daily_stats IS '主机漏洞每日统计';

-- 4. 安全告警每日统计视图 (v_views_security_alert_daily_stats)
CREATE OR REPLACE VIEW v_views_security_alert_daily_stats AS
SELECT
    COUNT(*) AS total_alert_count,
    COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END) AS today_new_count,
    CASE
        WHEN COUNT(*) = 0 THEN 0
        ELSE ROUND((COUNT(CASE WHEN created_at >= CURRENT_DATE THEN 1 END)::NUMERIC / NULLIF(COUNT(*), 0)) * 100, 2)
    END AS today_new_percentage
FROM v_alert_unified;

COMMENT ON VIEW v_views_security_alert_daily_stats IS '安全告警每日统计';

-- 5. 每小时告警趋势视图 (v_views_total_alert_hourly_stats)
CREATE OR REPLACE VIEW v_views_total_alert_hourly_stats AS
SELECT
    date_trunc('hour', alert_time) AS hour_bucket,
    COUNT(*) AS total_alerts,
    COUNT(CASE WHEN status = 0 THEN 1 END) AS pending_count,
    COUNT(CASE WHEN status = 1 THEN 1 END) AS processed_count,
    COUNT(CASE WHEN status = 2 THEN 1 END) AS ignored_count
FROM v_alert_unified
WHERE alert_time >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY date_trunc('hour', alert_time)
ORDER BY hour_bucket DESC;

COMMENT ON VIEW v_views_total_alert_hourly_stats IS '每小时告警趋势';

-- 6. 每月告警统计视图 (v_views_total_alert_monthly_stats)
CREATE OR REPLACE VIEW v_views_total_alert_monthly_stats AS
SELECT
    date_trunc('month', alert_time) AS month_bucket,
    COUNT(*) AS total_alerts,
    COUNT(CASE WHEN status = 0 THEN 1 END) AS pending_count,
    COUNT(CASE WHEN status = 1 THEN 1 END) AS processed_count,
    COUNT(CASE WHEN status = 2 THEN 1 END) AS ignored_count,
    ROUND(COUNT(*)::NUMERIC / NULLIF(EXTRACT(DAY FROM CURRENT_DATE), 0), 2) AS avg_daily_alerts
FROM v_alert_unified
WHERE alert_time >= date_trunc('year', CURRENT_DATE)
GROUP BY date_trunc('month', alert_time)
ORDER BY month_bucket DESC;

COMMENT ON VIEW v_views_total_alert_monthly_stats IS '每月告警统计';

-- 7. 主机在线状态视图 (v_views_host_status_summary)
CREATE OR REPLACE VIEW v_views_host_status_summary AS
SELECT
    COUNT(*) AS total_hosts,
    COUNT(CASE WHEN agent_status = 1 THEN 1 END) AS online_hosts,
    COUNT(CASE WHEN agent_status = 0 THEN 1 END) AS offline_hosts,
    CASE
        WHEN COUNT(*) = 0 THEN '0%'
        ELSE ROUND((COUNT(CASE WHEN agent_status = 1 THEN 1 END)::NUMERIC / COUNT(*)) * 100) || '%'
    END AS online_rate
FROM asset_host;

COMMENT ON VIEW v_views_host_status_summary IS '主机在线状态统计';

-- 8. 容器镜像漏洞TOP5视图 (v_views_image_vuln_top5_by_cve_all)
CREATE OR REPLACE VIEW v_views_image_vuln_top5_by_cve_all AS
SELECT
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    COUNT(DISTINCT ivd.image_id) AS affected_image_count,
    COUNT(*) AS total_instances,
    COUNT(CASE WHEN ivd.status = 0 THEN 1 END) AS pending_instances,
    COUNT(CASE WHEN ivd.status = 1 THEN 1 END) AS fixed_instances,
    STRING_AGG(DISTINCT ivs.image_name, ', ' LIMIT 3) AS affected_images_sample
FROM vuln_info vi
JOIN image_vuln_detail ivd ON vi.id = ivd.vuln_id
LEFT JOIN image_vuln_scan_task ivs ON ivd.scan_id = ivs.id
GROUP BY vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score
ORDER BY COUNT(DISTINCT ivd.image_id) DESC
LIMIT 5;

COMMENT ON VIEW v_views_image_vuln_top5_by_cve_all IS '容器镜像漏洞TOP5';

-- 9. 风险资产分布TOP5视图 (v_view_host_vuln_package_top5)
CREATE OR REPLACE VIEW v_view_host_vuln_package_top5 AS
SELECT
    vi.vuln_name AS package_name,
    COUNT(*) AS occurrence_count,
    ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC) AS rank
FROM host_vuln_detail hd
JOIN vuln_info vi ON hd.vuln_id = vi.id
WHERE hd.status = 0
GROUP BY vi.vuln_name
ORDER BY COUNT(*) DESC
LIMIT 5;

COMMENT ON VIEW v_view_host_vuln_package_top5 IS '风险资产分布TOP5';

-- 10. 威胁类型统计视图 (v_view_threat_type_total_count)
CREATE OR REPLACE VIEW v_view_threat_type_total_count AS
SELECT
    alert_type AS threat_type,
    COUNT(*) AS count
FROM v_alert_unified
GROUP BY alert_type
ORDER BY COUNT(*) DESC;

COMMENT ON VIEW v_view_threat_type_total_count IS '威胁类型统计';

-- 11. 安全看板漏洞统计视图 (v_view_vuln_chart_data)
CREATE OR REPLACE VIEW v_view_vuln_chart_data AS
SELECT
    vi.cve_id AS id,
    vi.vuln_name AS title,
    vi.severity,
    COUNT(DISTINCT hd.agent_id) AS affected_host_count
FROM vuln_info vi
JOIN host_vuln_detail hd ON vi.id = hd.vuln_id
WHERE hd.status = 0
GROUP BY vi.cve_id, vi.vuln_name, vi.severity
ORDER BY COUNT(DISTINCT hd.agent_id) DESC
LIMIT 10;

COMMENT ON VIEW v_view_vuln_chart_data IS '安全看板漏洞统计';

-- 12. 基线检查不通过主机TOP5 (v_views_host_baseline_fail_top5)
CREATE OR REPLACE VIEW v_views_host_baseline_fail_top5 AS
SELECT
    bcr.agent_id,
    bcr.host_ip,
    bcr.host_name,
    COUNT(CASE WHEN bcd.status = 0 THEN 1 END) AS fail_count
FROM baseline_check_detail bcd
JOIN baseline_check_result bcr ON bcd.result_id = bcr.id
GROUP BY bcr.agent_id, bcr.host_ip, bcr.host_name
HAVING COUNT(CASE WHEN bcd.status = 0 THEN 1 END) > 0
ORDER BY COUNT(CASE WHEN bcd.status = 0 THEN 1 END) DESC
LIMIT 5;

COMMENT ON VIEW v_views_host_baseline_fail_top5 IS '基线检查不通过主机TOP5';

-- 13. 基线检测项TOP5 (v_views_baseline_item_top5_affected)
CREATE OR REPLACE VIEW v_views_baseline_item_top5_affected AS
SELECT
    bci.id AS item_id,
    bci.item_name,
    COUNT(*) AS check_count,
    COUNT(CASE WHEN bcd.status = 0 THEN 1 END) AS failed_count,
    CASE
        WHEN COUNT(*) = 0 THEN '0%'
        ELSE ROUND((COUNT(CASE WHEN bcd.status = 0 THEN 1 END)::NUMERIC / COUNT(*)) * 100) || '%'
    END AS failed_rate,
    STRING_AGG(DISTINCT bcr.host_name, ', ' LIMIT 3) AS affected_hosts
FROM baseline_check_detail bcd
JOIN baseline_check_item bci ON bcd.item_id = bci.id
JOIN baseline_check_result bcr ON bcd.result_id = bcr.id
GROUP BY bci.id, bci.item_name
HAVING COUNT(CASE WHEN bcd.status = 0 THEN 1 END) > 0
ORDER BY COUNT(CASE WHEN bcd.status = 0 THEN 1 END) DESC
LIMIT 5;

COMMENT ON VIEW v_views_baseline_item_top5_affected IS '基线检测项TOP5';

-- 14. 代码仓库漏洞统计视图 (v_views_codeql_vuln_summary)
CREATE OR REPLACE VIEW v_views_codeql_vuln_summary AS
SELECT
    cr.id AS repo_id,
    cr.repo_name,
    cr.project_name,
    COUNT(*) AS total_vulns,
    COUNT(CASE WHEN csr.severity = 'critical' THEN 1 END) AS critical_count,
    COUNT(CASE WHEN csr.severity = 'high' THEN 1 END) AS high_count,
    COUNT(CASE WHEN csr.severity = 'medium' THEN 1 END) AS medium_count,
    COUNT(CASE WHEN csr.severity = 'low' THEN 1 END) AS low_count,
    MAX(csr.scan_time) AS last_scan_time
FROM code_repos cr
LEFT JOIN code_scan_results csr ON cr.id = csr.repo_id
GROUP BY cr.id, cr.repo_name, cr.project_name
HAVING COUNT(*) > 0
ORDER BY COUNT(*) DESC;

COMMENT ON VIEW v_views_codeql_vuln_summary IS '代码仓库漏洞统计';

-- 15. 主机漏洞TOP2 (v_views_vuln_count_vuls)
CREATE OR REPLACE VIEW v_views_vuln_count_vuls AS
SELECT
    vi.id AS vuln_id,
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    vi.description,
    vi.fix_suggestion,
    MIN(hd.scan_time) AS first_scan_time,
    MAX(hd.scan_time) AS last_scan_time,
    COUNT(DISTINCT hd.agent_id) AS affected_host_count,
    STRING_AGG(DISTINCT ah.host_name, ', ' LIMIT 5) AS affected_hosts
FROM vuln_info vi
JOIN host_vuln_detail hd ON vi.id = hd.vuln_id
LEFT JOIN asset_host ah ON hd.agent_id = ah.agent_id
WHERE hd.status = 0
GROUP BY vi.id, vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score, vi.description, vi.fix_suggestion
ORDER BY COUNT(DISTINCT hd.agent_id) DESC
LIMIT 2;

COMMENT ON VIEW v_views_vuln_count_vuls IS '主机漏洞TOP2';

-- 16. 容器漏洞TOP2 (v_views_vuln_count_image_vuls)
CREATE OR REPLACE VIEW v_views_vuln_count_image_vuls AS
SELECT
    vi.id AS vuln_id,
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    vi.description,
    vi.fix_suggestion,
    MIN(ivd.scan_time) AS first_scan_time,
    MAX(ivd.scan_time) AS last_scan_time,
    COUNT(DISTINCT ivd.image_id) AS affected_image_count,
    STRING_AGG(DISTINCT ivs.image_name, ', ' LIMIT 5) AS affected_images
FROM vuln_info vi
JOIN image_vuln_detail ivd ON vi.id = ivd.vuln_id
LEFT JOIN image_vuln_scan_task ivs ON ivd.scan_id = ivs.id
WHERE ivd.status = 0
GROUP BY vi.id, vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score, vi.description, vi.fix_suggestion
ORDER BY COUNT(DISTINCT ivd.image_id) DESC
LIMIT 2;

COMMENT ON VIEW v_views_vuln_count_image_vuls IS '容器漏洞TOP2';
