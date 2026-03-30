-- 删除告警表上的白名单字段索引
DROP INDEX IF EXISTS idx_alert_dc_wl_hit;
DROP INDEX IF EXISTS idx_alert_rs_wl_hit;
DROP INDEX IF EXISTS idx_alert_pe_wl_hit;
DROP INDEX IF EXISTS idx_alert_al_wl_hit;
DROP INDEX IF EXISTS idx_alert_bf_wl_hit;
DROP INDEX IF EXISTS idx_alert_mr_wl_hit;
DROP INDEX IF EXISTS idx_alert_na_wl_hit;
DROP INDEX IF EXISTS idx_alert_ms_wl_hit;
DROP INDEX IF EXISTS idx_alert_fi_wl_hit;
DROP INDEX IF EXISTS idx_alert_cdc_wl_hit;
DROP INDEX IF EXISTS idx_alert_crs_wl_hit;
DROP INDEX IF EXISTS idx_alert_csf_wl_hit;

-- 删除告警表上的白名单字段
ALTER TABLE alert_dangerous_command DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_dangerous_command DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_reverse_shell DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_reverse_shell DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_privilege_escalation DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_privilege_escalation DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_abnormal_login DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_abnormal_login DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_brute_force DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_brute_force DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_malicious_request DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_malicious_request DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_network_attack DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_network_attack DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_malware_scan DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_malware_scan DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_file_integrity DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_file_integrity DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_container_dangerous_command DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_container_dangerous_command DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_container_reverse_shell DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_container_reverse_shell DROP COLUMN IF EXISTS whitelist_rule_id;
ALTER TABLE alert_container_sensitive_file DROP COLUMN IF EXISTS whitelist_hit;
ALTER TABLE alert_container_sensitive_file DROP COLUMN IF EXISTS whitelist_rule_id;

-- 删除白名单规则表
DROP TABLE IF EXISTS whitelist_container_alert;
DROP TABLE IF EXISTS whitelist_fileguard;
DROP TABLE IF EXISTS whitelist_malware_scan;
DROP TABLE IF EXISTS whitelist_network_attack;
DROP TABLE IF EXISTS whitelist_malicious_request;
DROP TABLE IF EXISTS whitelist_brute_force;
DROP TABLE IF EXISTS whitelist_abnormal_login;
DROP TABLE IF EXISTS whitelist_privilege_escalation;
DROP TABLE IF EXISTS whitelist_reverse_shell;
DROP TABLE IF EXISTS whitelist_dangerous_command;
