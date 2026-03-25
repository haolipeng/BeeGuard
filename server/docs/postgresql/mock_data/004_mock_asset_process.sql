-- =====================================================
-- 模拟数据: asset_process (进程资产表)
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
-- OS: 全部 Linux (无 Windows)
-- =====================================================

INSERT INTO asset_process (agent_id, host_name, host_ip, os_type, name, status, version, path, run_name, start_time, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)  [13 rows]
-- ==========================================

-- aws-web-01: Nginx Web 服务器
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'nginx', 'running', '1.24.0', '/usr/sbin/nginx', 'root', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'systemd', 'running', '249', '/lib/systemd/systemd', 'root', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
-- aws-web-02: Nginx Web 服务器
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'nginx', 'running', '1.24.0', '/usr/sbin/nginx', 'root', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
-- aws-api-01: API 网关 (Node.js)
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'node', 'running', '20.11.0', '/usr/bin/node', 'app', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'systemd', 'running', '249', '/lib/systemd/systemd', 'root', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
-- aws-api-02: API 网关 (Node.js)
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'node', 'running', '20.11.0', '/usr/bin/node', 'app', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days', NOW()),
-- aws-gateway-01: Kong API 网关 (已离线)
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'kong', 'stopped', '3.5.0', '/usr/local/bin/kong', 'kong', NOW() - INTERVAL '150 days', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '150 days', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'systemd', 'running', '253', '/lib/systemd/systemd', 'root', NOW() - INTERVAL '150 days', NOW() - INTERVAL '150 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)  [12 rows]
-- ==========================================

-- aws-app-01: Java Spring Boot 应用
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'app', NOW() - INTERVAL '80 days', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '80 days', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'systemd', 'running', '249', '/lib/systemd/systemd', 'root', NOW() - INTERVAL '80 days', NOW() - INTERVAL '80 days', NOW()),
-- aws-app-02: Java Spring Boot 应用
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'app', NOW() - INTERVAL '78 days', NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '78 days', NOW() - INTERVAL '78 days', NOW()),
-- aws-app-03: Python/Gunicorn 应用
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'python3', 'running', '3.11.6', '/usr/bin/python3', 'app', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'gunicorn', 'running', '21.2.0', '/usr/local/bin/gunicorn', 'app', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
-- aws-worker-01: Celery 异步任务
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'python3', 'running', '3.10.12', '/usr/bin/python3', 'celery', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'sshd', 'running', '8.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),
-- aws-worker-02: Celery 异步任务
('agent-010-k7l8m9n0', 'aws-worker-02', '10.0.2.21', 'linux', 'python3', 'running', '3.10.12', '/usr/bin/python3', 'celery', NOW() - INTERVAL '68 days', NOW() - INTERVAL '68 days', NOW()),
('agent-010-k7l8m9n0', 'aws-worker-02', '10.0.2.21', 'linux', 'sshd', 'running', '8.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '68 days', NOW() - INTERVAL '68 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)  [22 rows]
-- ==========================================

-- aws-mysql-01: MySQL 主节点
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'mysqld', 'running', '8.0.35', '/usr/sbin/mysqld', 'mysql', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'systemd', 'running', '219', '/usr/lib/systemd/systemd', 'root', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
-- aws-mysql-02: MySQL 从节点
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'mysqld', 'running', '8.0.35', '/usr/sbin/mysqld', 'mysql', NOW() - INTERVAL '93 days', NOW() - INTERVAL '93 days', NOW()),
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '93 days', NOW() - INTERVAL '93 days', NOW()),
-- aws-pg-01: PostgreSQL
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'postgres', 'running', '15.4', '/usr/lib/postgresql/15/bin/postgres', 'postgres', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days', NOW()),
-- aws-redis-01: Redis 主节点
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'redis-server', 'running', '7.2.3', '/usr/bin/redis-server', 'redis', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '75 days', NOW() - INTERVAL '75 days', NOW()),
-- aws-redis-02: Redis 从节点/Sentinel
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 'redis-server', 'running', '7.2.3', '/usr/bin/redis-server', 'redis', NOW() - INTERVAL '73 days', NOW() - INTERVAL '73 days', NOW()),
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 'redis-sentinel', 'running', '7.2.3', '/usr/bin/redis-sentinel', 'redis', NOW() - INTERVAL '73 days', NOW() - INTERVAL '73 days', NOW()),
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '73 days', NOW() - INTERVAL '73 days', NOW()),
-- aws-es-01: Elasticsearch 节点
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'elasticsearch', NOW() - INTERVAL '65 days', NOW() - INTERVAL '65 days', NOW()),
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '65 days', NOW() - INTERVAL '65 days', NOW()),
-- aws-kafka-01: Kafka 消息队列
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'kafka', NOW() - INTERVAL '55 days', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '55 days', NOW() - INTERVAL '55 days', NOW()),
-- aws-mq-01: RabbitMQ 消息队列
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'beam.smp', 'running', '25.3', '/usr/lib64/erlang/erts-13.2/bin/beam.smp', 'rabbitmq', NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days', NOW()),
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'sshd', 'running', '8.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days', NOW()),
-- aws-mongo-01: MongoDB
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'mongod', 'running', '7.0.4', '/usr/bin/mongod', 'mongodb', NOW() - INTERVAL '48 days', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '48 days', NOW() - INTERVAL '48 days', NOW()),
-- aws-zk-01: ZooKeeper
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'zookeeper', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'sshd', 'running', '7.4p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)  [13 rows]
-- ==========================================

-- aws-eks-master-01: EKS 控制平面
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'kube-apiserver', 'running', '1.28.4', '/usr/local/bin/kube-apiserver', 'root', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'etcd', 'running', '3.5.10', '/usr/local/bin/etcd', 'root', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'systemd', 'running', '253', '/lib/systemd/systemd', 'root', NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days', NOW()),
-- aws-eks-node-01: EKS 工作节点
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'kubelet', 'running', '1.28.4', '/usr/local/bin/kubelet', 'root', NOW() - INTERVAL '58 days', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'containerd', 'running', '1.7.11', '/usr/bin/containerd', 'root', NOW() - INTERVAL '58 days', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '58 days', NOW() - INTERVAL '58 days', NOW()),
-- aws-eks-node-02: EKS 工作节点
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 'kubelet', 'running', '1.28.4', '/usr/local/bin/kubelet', 'root', NOW() - INTERVAL '56 days', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 'containerd', 'running', '1.7.11', '/usr/bin/containerd', 'root', NOW() - INTERVAL '56 days', NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '56 days', NOW() - INTERVAL '56 days', NOW()),
-- aws-eks-node-03: EKS 工作节点
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'linux', 'kubelet', 'running', '1.28.4', '/usr/local/bin/kubelet', 'root', NOW() - INTERVAL '54 days', NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '54 days', NOW() - INTERVAL '54 days', NOW()),
-- aws-eks-node-04: EKS 工作节点
('agent-027-a5b6c7d8', 'aws-eks-node-04', '10.0.4.14', 'linux', 'kubelet', 'running', '1.28.4', '/usr/local/bin/kubelet', 'root', NOW() - INTERVAL '52 days', NOW() - INTERVAL '52 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)  [10 rows]
-- ==========================================

-- aws-jenkins-01: Jenkins CI/CD
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'jenkins', NOW() - INTERVAL '100 days', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'dockerd', 'running', '24.0.7', '/usr/bin/dockerd', 'root', NOW() - INTERVAL '100 days', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '100 days', NOW() - INTERVAL '100 days', NOW()),
-- aws-gitlab-01: GitLab 代码仓库
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'puma', 'running', '6.4.0', '/opt/gitlab/embedded/bin/puma', 'git', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'sidekiq', 'running', '7.1.6', '/opt/gitlab/embedded/bin/sidekiq', 'git', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '95 days', NOW() - INTERVAL '95 days', NOW()),
-- aws-harbor-01: Harbor 镜像仓库
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'nginx', 'running', '1.22.1', '/usr/sbin/nginx', 'root', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '88 days', NOW() - INTERVAL '88 days', NOW()),
-- aws-sonar-01: SonarQube 代码质量
('agent-032-u5v6w7x8', 'aws-sonar-01', '10.0.5.14', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'sonarqube', NOW() - INTERVAL '76 days', NOW() - INTERVAL '76 days', NOW()),
('agent-032-u5v6w7x8', 'aws-sonar-01', '10.0.5.14', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '76 days', NOW() - INTERVAL '76 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)  [6 rows]
-- ==========================================

-- aws-prometheus-01: Prometheus 监控
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'prometheus', 'running', '2.47.2', '/usr/local/bin/prometheus', 'prometheus', NOW() - INTERVAL '110 days', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '110 days', NOW() - INTERVAL '110 days', NOW()),
-- aws-grafana-01: Grafana 可视化
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'grafana-server', 'running', '10.2.2', '/usr/sbin/grafana-server', 'grafana', NOW() - INTERVAL '105 days', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '105 days', NOW() - INTERVAL '105 days', NOW()),
-- aws-elk-01: ELK 日志分析 (Logstash + Kibana)
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'java', 'running', '17.0.9', '/usr/lib/jvm/java-17-openjdk/bin/java', 'logstash', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)  [4 rows]
-- ==========================================

-- aws-vpn-01: OpenVPN 服务器
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'openvpn', 'running', '2.5.9', '/usr/sbin/openvpn', 'root', NOW() - INTERVAL '120 days', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '120 days', NOW() - INTERVAL '120 days', NOW()),
-- aws-vault-01: HashiCorp Vault 密钥管理
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'vault', 'running', '1.15.4', '/usr/local/bin/vault', 'vault', NOW() - INTERVAL '35 days', NOW() - INTERVAL '35 days', NOW()),
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'sshd', 'running', '8.9p1', '/usr/sbin/sshd', 'root', NOW() - INTERVAL '35 days', NOW() - INTERVAL '35 days', NOW());
