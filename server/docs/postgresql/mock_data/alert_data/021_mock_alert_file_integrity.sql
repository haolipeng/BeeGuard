-- =====================================================
-- 模拟数据: alert_file_integrity (核心文件监控告警表)
-- 数据量: 35条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的主机生成核心文件监控告警数据
-- change_type: created/modified/deleted/permission_changed/owner_changed
-- =====================================================

INSERT INTO alert_file_integrity (agent_id, host_id, host_name, host_ip, rule_type, rule_name, file_path, change_type, before_hash, after_hash, process_name, process_user, status, event_time, created_at, updated_at) VALUES
-- 系统关键文件修改
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'system', '系统认证文件监控', '/etc/passwd', 'modified', 'abc123def456789012345678901234ab', 'def456abc789012345678901234567cd', 'useradd', 'root', 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 'system', '系统认证文件监控', '/etc/shadow', 'modified', '123456abc789def012345678901234ef', '789012def345abc678901234567890gh', 'passwd', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'system', '系统认证文件监控', '/etc/sudoers', 'modified', '456789def012abc345678901234567ij', '012345abc678def901234567890123kl', 'visudo', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'ssh_key', 'SSH配置文件监控', '/etc/ssh/sshd_config', 'modified', '789012abc345def678901234567890mn', '345678def901abc234567890123456op', 'vim', 'root', 0, NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '45 minutes', NOW()),
('agent-044-q3r4s5t6', 46, 'aws-proxy-01', '10.0.7.16', 'config_file', 'PAM配置监控', '/etc/pam.d/sshd', 'modified', '012345def678abc901234567890123qr', '678901abc234def567890123456789st', 'echo', 'root', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- 系统关键文件添加
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'ssh_key', 'SSH密钥监控', '/root/.ssh/authorized_keys', 'created', NULL, 'abc123def456789012345678901234uv', 'bash', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'cron_file', 'Cron任务监控', '/etc/cron.d/backdoor', 'created', NULL, 'def456abc789012345678901234567wx', 'crontab', 'root', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'system_binary', '系统启动项监控', '/etc/rc.local', 'created', NULL, '789012def345abc678901234567890yz', 'vim', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),

-- 系统关键文件删除
('agent-035-g7h8i9j0', 37, 'aws-elk-01', '10.0.6.12', 'log_file', '系统日志监控', '/var/log/auth.log', 'deleted', '123456abc789def012345678901234ab', NULL, 'rm', 'root', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-036-k1l2m3n4', 38, 'aws-elk-02', '10.0.6.13', 'log_file', '审计日志监控', '/var/log/audit/audit.log', 'deleted', '456789def012abc345678901234567cd', NULL, 'shred', 'root', 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),

-- 应用配置文件修改
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', 'config_file', 'MySQL配置监控', '/etc/mysql/my.cnf', 'modified', '789012abc345def678901234567890ef', '012345def678abc901234567890123gh', 'vim', 'root', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-014-a3b4c5d6', 14, 'aws-redis-01', '10.0.3.20', 'config_file', 'Redis配置监控', '/etc/redis/redis.conf', 'modified', '345678abc901def234567890123456ij', '901234def567abc890123456789012kl', 'sed', 'root', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'config_file', 'Nginx配置监控', '/etc/nginx/nginx.conf', 'modified', '678901def234abc567890123456789mn', '234567abc890def123456789012345op', 'nginx', 'www-data', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', 'config_file', 'GitLab配置监控', '/etc/gitlab/gitlab.rb', 'modified', '901234abc567def890123456789012qr', '567890def123abc456789012345678st', 'vim', 'root', 1, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'config_file', 'Jenkins配置监控', '/var/lib/jenkins/config.xml', 'modified', '234567def890abc123456789012345uv', '890123abc456def789012345678901wx', 'java', 'jenkins', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- Web目录文件监控
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'web_root', 'Web目录监控', '/var/www/html/uploads/shell.php', 'created', NULL, '567890abc123def456789012345678yz', 'php-fpm', 'www-data', 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 'web_root', 'Web目录监控', '/var/www/html/.htaccess', 'created', NULL, '890123def456abc789012345678901ab', 'apache2', 'www-data', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'web_root', 'Web目录监控', '/opt/tomcat/webapps/ROOT/index.jsp', 'modified', '123456abc789def012345678901234cd', '789012def345abc678901234567890ef', 'java', 'tomcat', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- 二进制文件监控
('agent-020-y7z8a9b0', 20, 'aws-kafka-02', '10.0.3.41', 'system_binary', '系统命令监控', '/usr/bin/ps', 'modified', '456789def012abc345678901234567gh', '012345abc678def901234567890123ij', 'cp', 'root', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', 'system_binary', '系统命令监控', '/usr/bin/netstat', 'modified', '789012abc345def678901234567890kl', '345678def901abc234567890123456mn', 'mv', 'root', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 'system_binary', '系统命令监控', '/usr/bin/ls', 'modified', '012345def678abc901234567890123op', '678901abc234def567890123456789qr', 'bash', 'root', 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),

-- 内核模块监控
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'system_binary', '内核模块监控', '/lib/modules/5.15.0/kernel/drivers/misc/rootkit.ko', 'created', NULL, '345678abc901def234567890123456st', 'insmod', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'system_binary', '内核模块监控', '/lib/modules/5.15.0/kernel/net/netfilter/hidden.ko', 'created', NULL, '678901def234abc567890123456789uv', 'modprobe', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),

-- 低级别配置变更
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 'config_file', '通用配置监控', '/etc/hosts', 'modified', '901234abc567def890123456789012wx', '567890def123abc456789012345678yz', 'vim', 'root', 2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),
('agent-009-g3h4i5j6', 9, 'aws-worker-01', '10.0.2.20', 'config_file', '通用配置监控', '/etc/resolv.conf', 'modified', '234567def890abc123456789012345ab', '890123abc456def789012345678901cd', 'dhclient', 'root', 1, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NOW()),
('agent-045-u7v8w9x0', 47, 'aws-backup-01', '10.0.7.17', 'config_file', '通用配置监控', '/etc/fstab', 'modified', '567890abc123def456789012345678ef', '123456def789abc012345678901234gh', 'mount', 'root', 1, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW()),
('agent-040-a7b8c9d0', 42, 'aws-dns-01', '10.0.7.12', 'config_file', 'DNS配置监控', '/etc/named.conf', 'modified', '890123def456abc789012345678901ij', '456789abc012def345678901234567kl', 'rndc', 'named', 2, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days', NOW()),

-- AWS EC2 系统监控
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', 'system_binary', 'Systemd服务监控', '/etc/systemd/system/backdoor.service', 'created', NULL, 'def456abc789012345678901234567op', 'systemctl', 'root', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-039-w3x4y5z6', 41, 'aws-bastion-01', '10.0.7.11', 'cron_file', 'Crontab监控', '/etc/cron.d/malicious_job', 'created', NULL, '789012abc345def678901234567890qr', 'crontab', 'root', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-041-e1f2g3h4', 43, 'aws-nfs-01', '10.0.7.13', 'web_root', 'Web目录监控', '/var/www/html/roundcube/shell.php', 'created', NULL, '012345def678abc901234567890123st', 'php-fpm', 'www-data', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- 权限变更监控
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', 'config_file', 'Linux安全策略监控', '/etc/security/limits.conf', 'modified', '345678abc901def234567890123456uv', '901234def567abc890123456789012wx', 'vim', 'root', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),

-- 权限和属主变更
('agent-004-m3n4o5p6', 4, 'aws-api-02', '10.0.1.21', 'system_binary', '系统二进制监控', '/usr/bin/find', 'permission_changed', NULL, NULL, 'chmod', 'root', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-007-y5z6a7b8', 7, 'aws-app-02', '10.0.2.11', 'system_binary', '系统二进制监控', '/usr/bin/nmap', 'permission_changed', NULL, NULL, 'chmod', 'root', 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),
('agent-010-k7l8m9n0', 10, 'aws-worker-02', '10.0.2.21', 'config_file', '配置文件监控', '/etc/ld.so.preload', 'created', NULL, '123456abc789def012345678901234mn', 'echo', 'root', 0, NOW() - INTERVAL '7 hours', NOW() - INTERVAL '7 hours', NOW()),
('agent-008-c9d0e1f2', 8, 'aws-batch-01', '10.0.2.30', 'log_file', '日志文件监控', '/var/log/wtmp', 'modified', '456789def012abc345678901234567op', '012345abc678def901234567890123qr', 'utmpdump', 'root', 0, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW());
