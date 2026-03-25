-- =====================================================
-- 模拟数据: asset_port (端口资产表)
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
-- protocol: 6=TCP, 17=UDP
-- =====================================================

INSERT INTO asset_port (agent_id, host_name, host_ip, os_type, port, protocol, listen_ip, listen_process, run_user, os_version, agent_status, agent_version, process_time, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================

-- aws-web-01 端口 (Nginx)
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 80, 6, '0.0.0.0', 'nginx', 'www-data', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 443, 6, '0.0.0.0', 'nginx', 'www-data', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
-- aws-web-02 端口 (Nginx)
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 80, 6, '0.0.0.0', 'nginx', 'www-data', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 443, 6, '0.0.0.0', 'nginx', 'www-data', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
-- aws-api-01 端口 (API Gateway / Node.js)
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 8080, 6, '0.0.0.0', 'node', 'app', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
-- aws-api-02 端口 (API Gateway / Node.js)
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 8080, 6, '0.0.0.0', 'node', 'app', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
-- aws-gateway-01 端口 (Kong API Gateway, agent离线)
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2023', 0, '2.1.2', NOW() - INTERVAL '150 days', NOW() - INTERVAL '150 days', NOW() - INTERVAL '2 days'),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 8000, 6, '0.0.0.0', 'kong', 'kong', 'Amazon Linux 2023', 0, '2.1.2', NOW() - INTERVAL '150 days', NOW() - INTERVAL '150 days', NOW() - INTERVAL '2 days'),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================

-- aws-app-01 端口 (Spring Boot)
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '80 days', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 8080, 6, '0.0.0.0', 'java', 'app', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '80 days', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 8443, 6, '0.0.0.0', 'java', 'app', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '80 days', NOW() - INTERVAL '80 days', NOW()),
-- aws-app-02 端口 (Spring Boot)
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '78 days', NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 8080, 6, '0.0.0.0', 'java', 'app', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '78 days', NOW() - INTERVAL '78 days', NOW()),
-- aws-worker-01 端口 (Celery Worker)
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 20.04', 1, '2.1.3', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 5555, 6, '0.0.0.0', 'python3', 'celery', 'Ubuntu 20.04', 1, '2.1.3', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================

-- aws-mysql-01 端口 (MySQL Primary)
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 3306, 6, '0.0.0.0', 'mysqld', 'mysql', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
-- aws-mysql-02 端口 (MySQL Replica)
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '93 days', NOW() - INTERVAL '93 days', NOW()),
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 3306, 6, '0.0.0.0', 'mysqld', 'mysql', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '93 days', NOW() - INTERVAL '93 days', NOW()),
-- aws-pg-01 端口 (PostgreSQL)
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.5', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 5432, 6, '0.0.0.0', 'postgres', 'postgres', 'Amazon Linux 2', 1, '2.1.5', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
-- aws-redis-01 端口 (Redis Primary + Sentinel)
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.5', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 6379, 6, '0.0.0.0', 'redis-server', 'redis', 'Amazon Linux 2', 1, '2.1.5', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 26379, 6, '0.0.0.0', 'redis-sentinel', 'redis', 'Amazon Linux 2', 1, '2.1.5', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
-- aws-redis-02 端口 (Redis Replica)
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '73 days', NOW() - INTERVAL '73 days', NOW()),
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 6379, 6, '0.0.0.0', 'redis-server', 'redis', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '73 days', NOW() - INTERVAL '73 days', NOW()),
-- aws-es-01 端口 (Elasticsearch)
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '65 days', NOW() - INTERVAL '65 days', NOW()),
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 9200, 6, '0.0.0.0', 'java', 'elasticsearch', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '65 days', NOW() - INTERVAL '65 days', NOW()),
-- aws-es-02 端口 (Elasticsearch)
('agent-017-m5n6o7p8', 'aws-es-02', '10.0.3.31', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '63 days', NOW() - INTERVAL '63 days', NOW()),
('agent-017-m5n6o7p8', 'aws-es-02', '10.0.3.31', 'linux', 9200, 6, '0.0.0.0', 'java', 'elasticsearch', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '63 days', NOW() - INTERVAL '63 days', NOW()),
-- aws-kafka-01 端口 (Kafka Broker)
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '55 days', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 9092, 6, '0.0.0.0', 'java', 'kafka', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '55 days', NOW() - INTERVAL '55 days', NOW()),
-- aws-kafka-02 端口 (Kafka Broker)
('agent-020-y7z8a9b0', 'aws-kafka-02', '10.0.3.41', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '53 days', NOW() - INTERVAL '53 days', NOW()),
('agent-020-y7z8a9b0', 'aws-kafka-02', '10.0.3.41', 'linux', 9092, 6, '0.0.0.0', 'java', 'kafka', 'Amazon Linux 2', 1, '2.1.4', NOW() - INTERVAL '53 days', NOW() - INTERVAL '53 days', NOW()),
-- aws-mq-01 端口 (RabbitMQ)
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 20.04', 1, '2.1.3', NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days', NOW()),
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 5672, 6, '0.0.0.0', 'beam.smp', 'rabbitmq', 'Ubuntu 20.04', 1, '2.1.3', NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days', NOW()),
-- aws-mongo-01 端口 (MongoDB)
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '48 days', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 27017, 6, '0.0.0.0', 'mongod', 'mongodb', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '48 days', NOW() - INTERVAL '48 days', NOW()),
-- aws-zk-01 端口 (ZooKeeper)
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.3', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 2181, 6, '0.0.0.0', 'java', 'zookeeper', 'Amazon Linux 2', 1, '2.1.3', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================

-- aws-eks-master-01 端口 (K8s Control Plane)
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 6443, 6, '0.0.0.0', 'kube-apiserver', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 2379, 6, '127.0.0.1', 'etcd', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 10250, 6, '0.0.0.0', 'kubelet', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
-- aws-eks-node-01 端口 (K8s Worker Node)
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '58 days', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 10250, 6, '0.0.0.0', 'kubelet', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '58 days', NOW() - INTERVAL '58 days', NOW()),
-- aws-eks-node-02 端口 (K8s Worker Node)
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '56 days', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 10250, 6, '0.0.0.0', 'kubelet', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '56 days', NOW() - INTERVAL '56 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================

-- aws-jenkins-01 端口 (Jenkins CI)
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.3', NOW() - INTERVAL '100 days', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 8080, 6, '0.0.0.0', 'java', 'jenkins', 'Ubuntu 22.04', 1, '2.1.3', NOW() - INTERVAL '100 days', NOW() - INTERVAL '100 days', NOW()),
-- aws-gitlab-01 端口 (GitLab)
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 80, 6, '0.0.0.0', 'nginx', 'git', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 443, 6, '0.0.0.0', 'nginx', 'git', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
-- aws-harbor-01 端口 (Harbor Registry)
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 80, 6, '0.0.0.0', 'nginx', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 443, 6, '0.0.0.0', 'nginx', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================

-- aws-prometheus-01 端口 (Prometheus)
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '110 days', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 9090, 6, '0.0.0.0', 'prometheus', 'prometheus', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '110 days', NOW() - INTERVAL '110 days', NOW()),
-- aws-grafana-01 端口 (Grafana)
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '105 days', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 3000, 6, '0.0.0.0', 'grafana-server', 'grafana', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '105 days', NOW() - INTERVAL '105 days', NOW()),
-- aws-elk-01 端口 (Logstash + Kibana)
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 5601, 6, '0.0.0.0', 'node', 'kibana', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 5044, 6, '0.0.0.0', 'logstash', 'logstash', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================

-- aws-vpn-01 端口 (OpenVPN)
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '120 days', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 1194, 17, '0.0.0.0', 'openvpn', 'nobody', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '120 days', NOW() - INTERVAL '120 days', NOW()),
-- aws-bastion-01 端口 (Bastion/Jump Host, 仅SSH)
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2023', 1, '2.1.5', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days', NOW()),
-- aws-dns-01 端口 (BIND DNS)
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Amazon Linux 2', 1, '2.1.3', NOW() - INTERVAL '115 days', NOW() - INTERVAL '115 days', NOW()),
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 53, 6, '0.0.0.0', 'named', 'named', 'Amazon Linux 2', 1, '2.1.3', NOW() - INTERVAL '115 days', NOW() - INTERVAL '115 days', NOW()),
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 53, 17, '0.0.0.0', 'named', 'named', 'Amazon Linux 2', 1, '2.1.3', NOW() - INTERVAL '115 days', NOW() - INTERVAL '115 days', NOW()),
-- aws-mail-01 端口 (Postfix + Dovecot)
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 25, 6, '0.0.0.0', 'postfix', 'postfix', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 993, 6, '0.0.0.0', 'dovecot', 'dovecot', 'Ubuntu 22.04', 1, '2.1.4', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
-- aws-ftp-01 端口 (vsftpd, agent离线)
('agent-046-y1z2a3b4', 'aws-ftp-01', '10.0.7.18', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 20.04', 0, '2.1.2', NOW() - INTERVAL '130 days', NOW() - INTERVAL '130 days', NOW() - INTERVAL '3 days'),
('agent-046-y1z2a3b4', 'aws-ftp-01', '10.0.7.18', 'linux', 21, 6, '0.0.0.0', 'vsftpd', 'ftp', 'Ubuntu 20.04', 0, '2.1.2', NOW() - INTERVAL '130 days', NOW() - INTERVAL '130 days', NOW() - INTERVAL '3 days'),
-- aws-vault-01 端口 (HashiCorp Vault)
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 22, 6, '0.0.0.0', 'sshd', 'root', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '35 days', NOW() - INTERVAL '35 days', NOW()),
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 8200, 6, '0.0.0.0', 'vault', 'vault', 'Ubuntu 22.04', 1, '2.1.5', NOW() - INTERVAL '35 days', NOW() - INTERVAL '35 days', NOW());
