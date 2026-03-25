-- 000003 down: 删除告警表
DROP TABLE IF EXISTS alert_process_log CASCADE;
DROP TABLE IF EXISTS alert_file_integrity CASCADE;
DROP TABLE IF EXISTS alert_malware_scan CASCADE;
DROP TABLE IF EXISTS alert_network_attack CASCADE;
DROP TABLE IF EXISTS alert_malicious_request CASCADE;
DROP TABLE IF EXISTS alert_abnormal_login CASCADE;
DROP TABLE IF EXISTS alert_privilege_escalation CASCADE;
DROP TABLE IF EXISTS alert_reverse_shell CASCADE;
DROP TABLE IF EXISTS alert_dangerous_command CASCADE;
DROP TABLE IF EXISTS alert_brute_force CASCADE;
