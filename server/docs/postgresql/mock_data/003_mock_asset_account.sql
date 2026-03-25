-- =====================================================
-- 模拟数据: asset_account (账号资产表)
-- 数据量: 80条
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

INSERT INTO asset_account (agent_id, host_name, host_ip, os_type, name, uid, status, permission, login_type, last_login_time, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================

-- aws-web-01 (Ubuntu 22.04)
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'www-data', 33, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'nginx', 990, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '90 days', NOW()),

-- aws-web-02 (Ubuntu 22.04)
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'www-data', 33, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '88 days', NOW()),

-- aws-api-01 (Ubuntu 22.04)
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'www-data', 33, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '85 days', NOW()),

-- aws-gateway-01 (Amazon Linux 2023)
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '3 days', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '3 days', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'nginx', 990, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '150 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================

-- aws-app-01 (Ubuntu 22.04)
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'app', 1001, 0, 'normal', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '75 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'deploy', 1002, 0, 'normal', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '70 days', NOW()),

-- aws-app-02 (Ubuntu 22.04)
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'app', 1001, 0, 'normal', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '73 days', NOW()),

-- aws-worker-01 (Ubuntu 20.04)
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================

-- aws-mysql-01 (Amazon Linux 2)
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'mysql', 27, 0, 'normal', '/bin/false', NULL, NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'dba', 1001, 0, 'normal', '/bin/bash', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '90 days', NOW()),

-- aws-pg-01 (Amazon Linux 2)
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'postgres', 26, 0, 'normal', '/bin/bash', NULL, NOW() - INTERVAL '90 days', NOW()),

-- aws-redis-01 (Amazon Linux 2)
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'redis', 999, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '75 days', NOW()),

-- aws-kafka-01 (Amazon Linux 2)
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '8 hours', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'kafka', 987, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '55 days', NOW()),

-- aws-mq-01 (Ubuntu 20.04)
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '50 days', NOW()),
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '50 days', NOW()),
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'rabbitmq', 998, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '50 days', NOW()),

-- aws-mongo-01 (Ubuntu 22.04)
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'mongod', 985, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '48 days', NOW()),

-- aws-zk-01 (Amazon Linux 2)
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'zookeeper', 986, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '45 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================

-- aws-eks-master-01 (Amazon Linux 2023)
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'etcd', 989, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '60 days', NOW()),

-- aws-eks-node-01 (Amazon Linux 2023)
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '58 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================

-- aws-jenkins-01 (Ubuntu 22.04)
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '8 hours', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'jenkins', 992, 0, 'normal', '/bin/bash', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '100 days', NOW()),

-- aws-gitlab-01 (Ubuntu 22.04)
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'git', 991, 0, 'normal', '/bin/sh', NULL, NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'gitlab-www', 993, 0, 'normal', '/bin/false', NULL, NOW() - INTERVAL '95 days', NOW()),

-- aws-harbor-01 (Ubuntu 22.04)
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'harbor', 10000, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '88 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================

-- aws-prometheus-01 (Ubuntu 22.04)
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'prometheus', 996, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '110 days', NOW()),

-- aws-grafana-01 (Ubuntu 22.04)
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'grafana', 997, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '105 days', NOW()),

-- aws-elk-01 (Ubuntu 22.04)
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'elasticsearch', 988, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'logstash', 983, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'kibana', 982, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================

-- aws-vpn-01 (Ubuntu 22.04)
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'openvpn', 980, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '120 days', NOW()),

-- aws-bastion-01 (Amazon Linux 2023)
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '45 days', NOW()),
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '45 days', NOW()),
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'ops', 1001, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '40 days', NOW()),

-- aws-mail-01 (Ubuntu 22.04)
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '8 hours', NOW() - INTERVAL '90 days', NOW()),
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '90 days', NOW()),
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'postfix', 979, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'dovecot', 978, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '90 days', NOW()),

-- aws-proxy-01 (Amazon Linux 2023)
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '80 days', NOW()),
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'ec2-user', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '80 days', NOW()),
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'squid', 976, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '80 days', NOW()),

-- aws-vault-01 (Ubuntu 22.04)
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'root', 0, 0, 'root', '/bin/bash', NOW() - INTERVAL '6 hours', NOW() - INTERVAL '35 days', NOW()),
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'ubuntu', 1000, 0, 'sudo', '/bin/bash', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '35 days', NOW()),
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'vault', 973, 0, 'normal', '/usr/sbin/nologin', NULL, NOW() - INTERVAL '35 days', NOW());
