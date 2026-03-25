-- =====================================================
-- SOC 数据库视图初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- 说明: 各模块统计视图定义
-- 依赖: 需在所有基础表创建完成后执行
-- =====================================================


-- =====================================================
-- 1. 主机漏洞统计视图 (v_vuln_count_hosts)
-- 按主机维度统计漏洞数量
-- =====================================================
CREATE OR REPLACE VIEW v_vuln_count_hosts AS
SELECT
    hs.host_ip,
    hs.host_name,
    MAX(hd.scan_time)  AS last_scan_time,
    MIN(hd.scan_time)  AS first_scan_time,
    COUNT(CASE WHEN vi.severity = 'critical' THEN 1 END) AS critical_vulns,
    COUNT(CASE WHEN vi.severity = 'high'     THEN 1 END) AS high_vulns,
    COUNT(CASE WHEN vi.severity = 'medium'   THEN 1 END) AS medium_vulns,
    COUNT(CASE WHEN vi.severity = 'low'      THEN 1 END) AS low_vulns,
    COUNT(*)                                              AS total_vulns
FROM host_vuln_detail hd
JOIN vuln_info vi ON hd.vuln_id = vi.id
JOIN host_vuln_scan_task hs ON hd.scan_id = hs.id
WHERE hd.status = 0
GROUP BY hs.host_ip, hs.host_name;

COMMENT ON VIEW v_vuln_count_hosts IS '漏洞统计-按主机维度';


-- =====================================================
-- 2. 镜像漏洞统计视图 (v_vuln_count_images)
-- 按镜像维度统计漏洞数量
-- =====================================================
CREATE OR REPLACE VIEW v_vuln_count_images AS
SELECT
    ivd.image_id,
    ivs.image_name,
    MAX(ivd.scan_time)  AS last_scan_time,
    MIN(ivd.scan_time)  AS first_scan_time,
    COUNT(CASE WHEN vi.severity = 'critical' THEN 1 END) AS critical_vulns,
    COUNT(CASE WHEN vi.severity = 'high'     THEN 1 END) AS high_vulns,
    COUNT(CASE WHEN vi.severity = 'medium'   THEN 1 END) AS medium_vulns,
    COUNT(CASE WHEN vi.severity = 'low'      THEN 1 END) AS low_vulns,
    COUNT(*)                                              AS total_vulns
FROM image_vuln_detail ivd
JOIN vuln_info vi ON ivd.vuln_id = vi.id
JOIN image_vuln_scan_task ivs ON ivd.scan_id = ivs.id
WHERE ivd.status = 0
GROUP BY ivd.image_id, ivs.image_name;

COMMENT ON VIEW v_vuln_count_images IS '漏洞统计-按镜像维度';


-- =====================================================
-- 3. 漏洞维度主机统计视图 (v_vuln_count_vuls)
-- 按漏洞维度统计影响的主机
-- =====================================================
CREATE OR REPLACE VIEW v_vuln_count_vuls AS
SELECT
    vi.id                AS vuln_id,
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    vi.description,
    vi.fix_suggestion,
    MIN(hd.scan_time)    AS first_scan_time,
    MAX(hd.scan_time)    AS last_scan_time,
    COUNT(DISTINCT hd.agent_id) AS affected_host_count,
    json_agg(json_build_object(
        'host_id',   hd.host_id,
        'host_name', hs.host_name,
        'host_ip',   hs.host_ip,
        'scan_time', hd.scan_time,
        'status',    hd.status
    )) AS affected_hosts
FROM vuln_info vi
JOIN host_vuln_detail hd ON vi.id = hd.vuln_id
JOIN host_vuln_scan_task hs ON hd.scan_id = hs.id
GROUP BY vi.id, vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score, vi.description, vi.fix_suggestion;

COMMENT ON VIEW v_vuln_count_vuls IS '漏洞统计-按漏洞维度(主机)';


-- =====================================================
-- 4. 漏洞维度镜像统计视图 (v_vuln_count_image_vuls)
-- 按漏洞维度统计影响的镜像
-- =====================================================
CREATE OR REPLACE VIEW v_vuln_count_image_vuls AS
SELECT
    vi.id                AS vuln_id,
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    vi.description,
    vi.fix_suggestion,
    MIN(ivd.scan_time)   AS first_scan_time,
    MAX(ivd.scan_time)   AS last_scan_time,
    COUNT(DISTINCT ivd.image_id) AS affected_image_count,
    json_agg(json_build_object(
        'agent_id',   ivd.agent_id,
        'image_id',   ivd.image_id,
        'image_name', ivs.image_name,
        'scan_time',  ivd.scan_time,
        'status',     ivd.status
    )) AS affected_images
FROM vuln_info vi
JOIN image_vuln_detail ivd ON vi.id = ivd.vuln_id
JOIN image_vuln_scan_task ivs ON ivd.scan_id = ivs.id
GROUP BY vi.id, vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score, vi.description, vi.fix_suggestion;

COMMENT ON VIEW v_vuln_count_image_vuls IS '漏洞统计-按漏洞维度(镜像)';


-- =====================================================
-- 5. 基线检查主机统计视图 (baseline_check_host_view)
-- 按主机维度统计基线检查结果
-- 通过 JOIN baseline_check_result 获取 host_name/host_ip（范式化设计）
-- =====================================================
CREATE OR REPLACE VIEW baseline_check_host_view AS
SELECT
    bcr.agent_id,
    bcr.host_name,
    bcr.host_ip,
    COUNT(*)                                         AS total_checks,
    COUNT(CASE WHEN bcd.status = 1 THEN 1 END)      AS passed_checks,
    COUNT(CASE WHEN bcd.status = 0 THEN 1 END)      AS failed_checks,
    COUNT(CASE WHEN bcd.status = 2 THEN 1 END)      AS error_checks,
    MAX(bcd.check_time)                              AS last_check_time
FROM baseline_check_detail bcd
JOIN baseline_check_result bcr ON bcd.result_id = bcr.id
GROUP BY bcr.agent_id, bcr.host_name, bcr.host_ip;

COMMENT ON VIEW baseline_check_host_view IS '基线检查-按主机统计';


-- =====================================================
-- 6. 基线检查项统计视图 (baseline_check_item_view)
-- 按检查项维度统计
-- 通过 JOIN baseline_check_item 获取 item_name（范式化设计）
-- =====================================================
CREATE OR REPLACE VIEW baseline_check_item_view AS
SELECT
    bci.item_name,
    COUNT(DISTINCT bcd.agent_id) AS total_hosts
FROM baseline_check_detail bcd
JOIN baseline_check_item bci ON bcd.item_id = bci.id
GROUP BY bci.item_name;

COMMENT ON VIEW baseline_check_item_view IS '基线检查-按检查项统计';


-- =====================================================
-- 7. 告警联合视图 (v_alert_unified)
-- 整合所有告警表，共同字段单独列出，不同字段转为JSON格式供AI分析
-- =====================================================
CREATE OR REPLACE VIEW v_alert_unified AS
-- 暴力破解告警
SELECT
    'brute_force' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    attack_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'source_ip', source_ip,
        'source_location', source_location,
        'attack_type', attack_type,
        'target_ip', target_ip,
        'target_port', target_port,
        'username', username,
        'attempt_count', attempt_count,
        'first_attack_time', first_attack_time,
        'is_blocked', is_blocked,
        'process_time', process_time,
        'processor', processor,
        'remark', remark
    ) AS details
FROM alert_brute_force

UNION ALL

-- 高危命令告警
SELECT
    'dangerous_command' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    alert_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'command', command,
        'command_type', command_type,
        'user', "user",
        'privilege_level', privilege_level
    ) AS details
FROM alert_dangerous_command

UNION ALL

-- 反弹Shell告警
SELECT
    'reverse_shell' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    victim_ip AS host_ip,
    status,
    event_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'victim_ip', victim_ip,
        'command_line', command_line,
        'shell_type', shell_type,
        'target_host', target_host,
        'target_port', target_port
    ) AS details
FROM alert_reverse_shell

UNION ALL

-- 本地提权告警
SELECT
    'privilege_escalation' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    discover_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'escalated_user', escalated_user,
        'parent_process', parent_process,
        'parent_process_user', parent_process_user,
        'process_id', process_id,
        'process_path', process_path
    ) AS details
FROM alert_privilege_escalation

UNION ALL

-- 异常登录告警
SELECT
    'abnormal_login' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    login_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'source_ip', source_ip,
        'source_location', source_location,
        'source_country', source_country,
        'source_city', source_city,
        'login_user', login_user,
        'risk_level', risk_level,
        'abnormal_type', abnormal_type,
        'is_whitelist', is_whitelist
    ) AS details
FROM alert_abnormal_login

UNION ALL

-- 恶意请求告警
SELECT
    'malicious_request' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    last_request_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'policy_type', policy_type,
        'policy_name', policy_name,
        'malicious_domain', malicious_domain,
        'malicious_ip', malicious_ip,
        'request_count', request_count,
        'first_request_time', first_request_time,
        'risk_description', risk_description
    ) AS details
FROM alert_malicious_request

UNION ALL

-- 网络攻击告警
SELECT
    'network_attack' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    last_attack_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'target_port', target_port,
        'attacker_ip', attacker_ip,
        'attacker_location', attacker_location,
        'attacker_country', attacker_country,
        'vulnerability_name', vulnerability_name,
        'vulnerability_id', vulnerability_id,
        'attack_status', attack_status,
        'attack_count', attack_count,
        'first_attack_time', first_attack_time,
        'attack_payload', attack_payload
    ) AS details
FROM alert_network_attack

UNION ALL

-- 文件查杀告警
SELECT
    'malware_scan' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    scan_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'threat_type', threat_type,
        'file_name', file_name,
        'file_path', file_path,
        'file_size', file_size,
        'file_md5', file_md5,
        'file_sha256', file_sha256,
        'detection_engine', detection_engine,
        'malware_family', malware_family,
        'is_quarantined', is_quarantined,
        'is_deleted', is_deleted
    ) AS details
FROM alert_malware_scan

UNION ALL

-- 核心文件监控告警
SELECT
    'file_integrity' AS alert_type,
    id,
    agent_id,
    host_id,
    host_name,
    host_ip,
    status,
    alert_time AS alert_time,
    created_at,
    updated_at,
    jsonb_build_object(
        'rule_type', rule_type,
        'rule_name', rule_name,
        'rule_id', rule_id,
        'threat_level', threat_level,
        'threat_action', threat_action,
        'file_path', file_path,
        'file_name', file_name,
        'old_content_hash', old_content_hash,
        'new_content_hash', new_content_hash,
        'change_detail', change_detail,
        'operator_user', operator_user,
        'operator_process', operator_process,
        'alert_description', alert_description
    ) AS details
FROM alert_file_integrity;

COMMENT ON VIEW v_alert_unified IS '告警联合视图-整合所有告警类型供AI分析';
COMMENT ON COLUMN v_alert_unified.alert_type IS '告警类型: brute_force/dangerous_command/reverse_shell/privilege_escalation/abnormal_login/malicious_request/network_attack/malware_scan/file_integrity';
COMMENT ON COLUMN v_alert_unified.alert_time IS '告警时间(统一时间字段)';
COMMENT ON COLUMN v_alert_unified.details IS '告警详情(JSON格式,包含各类型特有字段)';


-- =====================================================
-- 初始化完成
-- =====================================================
