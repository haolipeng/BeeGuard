-- =====================================================
-- 模拟数据: alert_privilege_escalation (本地提权告警表)
-- 数据量: 30条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 子网划分:
--   10.0.1.x  Web/API 层 (公有子网)
--   10.0.2.x  应用层 (私有子网)
--   10.0.3.x  数据层 (私有子网)
--   10.0.4.x  EKS/K8s 层 (私有子网)
--   10.0.5.x  DevOps 层 (私有子网)
--   10.0.6.x  监控层 (私有子网)
--   10.0.7.x  基础设施/安全层 (私有子网)
-- OS: Ubuntu 22.04/20.04, Amazon Linux 2/2023 (无 Windows)
-- =====================================================

INSERT INTO alert_privilege_escalation (agent_id, host_id, host_name, host_ip, escalated_user, parent_process, parent_process_user, process_id, process_path, status, discover_time, created_at, updated_at) VALUES

-- sudo提权
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'root', 'bash', 'www-data', 12345, '/bin/bash', 0, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 'root', 'nginx', 'nginx', 23456, '/usr/sbin/nginx', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'root', 'java', 'appuser', 34567, '/usr/bin/java', 1, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 'root', 'node', 'nodejs', 45678, '/usr/bin/node', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- CVE漏洞提权
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'root', 'python3', 'apiuser', 56789, '/usr/bin/python3', 0, NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '45 minutes', NOW()),
('agent-017-m5n6o7p8', 17, 'aws-es-02', '10.0.3.31', 'root', 'java', 'jenkins', 67890, '/opt/java/bin/java', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-018-q9r0s1t2', 18, 'aws-es-03', '10.0.3.32', 'root', 'gitlab-rails', 'git', 78901, '/opt/gitlab/embedded/bin/ruby', 1, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),
('agent-020-y7z8a9b0', 20, 'aws-kafka-02', '10.0.3.41', 'root', 'kubelet', 'kubelet', 89012, '/usr/bin/kubelet', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- SUID提权
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'root', 'find', 'kubernetes', 90123, '/usr/bin/find', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', 'root', 'vim.basic', 'deploy', 10234, '/usr/bin/vim.basic', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-025-s7t8u9v0', 27, 'aws-eks-node-02', '10.0.4.12', 'root', 'nmap', 'devops', 11345, '/usr/bin/nmap', 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),
('agent-044-q3r4s5t6', 46, 'aws-proxy-01', '10.0.7.16', 'root', 'python3', 'proxy', 12456, '/usr/bin/python3', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),

-- Capability提权
('agent-007-y5z6a7b8', 7, 'aws-app-02', '10.0.2.11', 'root', 'redis-server', 'redis', 13567, '/usr/bin/redis-server', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-009-g3h4i5j6', 9, 'aws-worker-01', '10.0.2.20', 'root', 'rabbitmq-server', 'rabbitmq', 14678, '/usr/lib/rabbitmq/bin/rabbitmq-server', 1, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '12 hours', NOW()),
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', 'root', 'python3', 'logstash', 15789, '/usr/bin/python3', 0, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '6 hours', NOW()),

-- 内核漏洞提权
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'root', 'exploit', 'elasticsearch', 16890, '/tmp/.hidden/exploit', 0, NOW() - INTERVAL '20 minutes', NOW() - INTERVAL '20 minutes', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', 'root', 'dirty_cow', 'elastic', 17901, '/tmp/dirty_cow', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-019-u3v4w5x6', 19, 'aws-kafka-01', '10.0.3.40', 'root', 'overlayfs_exp', 'kafka', 18012, '/var/tmp/overlayfs_exp', 0, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
('agent-034-c3d4e5f6', 36, 'aws-grafana-01', '10.0.6.11', 'root', 'pkexec', 'zookeeper', 19123, '/usr/bin/pkexec', 1, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW()),

-- Docker/容器逃逸
('agent-030-m7n8o9p0', 32, 'aws-harbor-01', '10.0.5.12', 'root', 'docker', 'harbor', 20234, '/usr/bin/docker', 0, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'root', 'containerd-shim', 'containerd', 21345, '/usr/bin/containerd-shim', 0, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', 'root', 'runc', 'containerd', 22456, '/usr/bin/runc', 1, NOW() - INTERVAL '10 hours', NOW() - INTERVAL '10 hours', NOW()),

-- cron/at提权
('agent-012-s5t6u7v8', 12, 'aws-mysql-02', '10.0.3.11', 'root', 'cron', 'prometheus', 23567, '/usr/sbin/cron', 0, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
('agent-013-w9x0y1z2', 13, 'aws-pg-01', '10.0.3.12', 'root', 'atd', 'backup', 24678, '/usr/sbin/atd', 2, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW()),
('agent-049-k3l4m5n6', 49, 'aws-consul-01', '10.0.3.72', 'root', 'anacron', 'ansible', 25789, '/usr/sbin/anacron', 0, NOW() - INTERVAL '7 hours', NOW() - INTERVAL '7 hours', NOW()),

-- 其他提权方式
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', 'root', 'sudo', 'nginx', 26890, '/usr/bin/sudo', 0, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 'root', 'postfix', 'postfix', 32456, '/usr/lib/postfix/sbin/master', 0, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '8 hours', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', 'root', 'vsftpd', 'ftp', 33567, '/usr/sbin/vsftpd', 2, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW()),
('agent-043-m9n0o1p2', 45, 'aws-ldap-01', '10.0.7.15', 'root', 'slapd', 'openldap', 34678, '/usr/sbin/slapd', 0, NOW() - INTERVAL '9 hours', NOW() - INTERVAL '9 hours', NOW()),
('agent-050-o7p8q9r0', 50, 'aws-vault-01', '10.0.7.20', 'root', 'prometheus', 'prometheus', 35789, '/usr/bin/prometheus', 1, NOW() - INTERVAL '15 hours', NOW() - INTERVAL '15 hours', NOW());
