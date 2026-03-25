-- =====================================================
-- 模拟数据: asset_system_service (系统服务资产表)
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

INSERT INTO asset_system_service (agent_id, host_name, host_ip, os_type, name, version, status, run_user, path, describe, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================

-- aws-web-01 (Ubuntu 22.04) Nginx Web 服务器
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'nginx.service', '1.24.0', 'active', 'root', '/lib/systemd/system/nginx.service', 'A high performance web server and reverse proxy', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '90 days', NOW()),

-- aws-web-02 (Ubuntu 22.04) Nginx Web 服务器
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'nginx.service', '1.24.0', 'active', 'root', '/lib/systemd/system/nginx.service', 'A high performance web server and reverse proxy', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '88 days', NOW()),

-- aws-api-01 (Ubuntu 22.04) API 网关
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'nginx.service', '1.24.0', 'active', 'root', '/lib/systemd/system/nginx.service', 'A high performance web server and reverse proxy', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '85 days', NOW()),

-- aws-api-02 (Ubuntu 22.04) API 网关
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'nginx.service', '1.24.0', 'active', 'root', '/lib/systemd/system/nginx.service', 'A high performance web server and reverse proxy', NOW() - INTERVAL '85 days', NOW()),
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '85 days', NOW()),

-- aws-gateway-01 (Amazon Linux 2023, agent_status=0 离线)
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'nginx.service', '1.24.0', 'inactive', 'root', '/usr/lib/systemd/system/nginx.service', 'A high performance web server and reverse proxy', NOW() - INTERVAL '150 days', NOW() - INTERVAL '2 days'),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'sshd.service', '8.2p1', 'inactive', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '150 days', NOW() - INTERVAL '2 days'),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================

-- aws-app-01 (Ubuntu 22.04) Java 应用服务器
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'tomcat.service', '10.1.18', 'active', 'app', '/lib/systemd/system/tomcat.service', 'Apache Tomcat Web Application Container', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '80 days', NOW()),

-- aws-app-02 (Ubuntu 22.04) Java 应用服务器
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'tomcat.service', '10.1.18', 'active', 'app', '/lib/systemd/system/tomcat.service', 'Apache Tomcat Web Application Container', NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '78 days', NOW()),

-- aws-app-03 (Amazon Linux 2023) Node.js 应用服务器
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'node-app.service', '20.11.0', 'active', 'app', '/usr/lib/systemd/system/node-app.service', 'Node.js Application Service', NOW() - INTERVAL '75 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'sshd.service', '8.2p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '75 days', NOW()),

-- aws-worker-01 (Ubuntu 20.04) Celery 后台任务
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'celery.service', '5.3.6', 'active', 'app', '/lib/systemd/system/celery.service', 'Celery distributed task queue', NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'sshd.service', '8.4p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '70 days', NOW()),

-- aws-worker-02 (Ubuntu 20.04) Celery 后台任务
('agent-010-k7l8m9n0', 'aws-worker-02', '10.0.2.21', 'linux', 'celery.service', '5.3.6', 'active', 'app', '/lib/systemd/system/celery.service', 'Celery distributed task queue', NOW() - INTERVAL '68 days', NOW()),
('agent-010-k7l8m9n0', 'aws-worker-02', '10.0.2.21', 'linux', 'sshd.service', '8.4p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '68 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================

-- aws-mysql-01 (Amazon Linux 2) MySQL 主库
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'mysqld.service', '8.0.36', 'active', 'mysql', '/usr/lib/systemd/system/mysqld.service', 'MySQL Server', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '95 days', NOW()),

-- aws-mysql-02 (Amazon Linux 2) MySQL 从库
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'mysqld.service', '8.0.36', 'active', 'mysql', '/usr/lib/systemd/system/mysqld.service', 'MySQL Server', NOW() - INTERVAL '93 days', NOW()),
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '93 days', NOW()),

-- aws-pg-01 (Amazon Linux 2) PostgreSQL
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'postgresql-15.service', '15.5', 'active', 'postgres', '/usr/lib/systemd/system/postgresql-15.service', 'PostgreSQL 15 database server', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '90 days', NOW()),

-- aws-redis-01 (Amazon Linux 2) Redis 主节点
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'redis.service', '7.2.4', 'active', 'redis', '/usr/lib/systemd/system/redis.service', 'Advanced key-value store', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '75 days', NOW()),

-- aws-redis-02 (Amazon Linux 2) Redis 从节点
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 'redis.service', '7.2.4', 'active', 'redis', '/usr/lib/systemd/system/redis.service', 'Advanced key-value store', NOW() - INTERVAL '73 days', NOW()),
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '73 days', NOW()),

-- aws-es-01 (Ubuntu 22.04) Elasticsearch 数据节点
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'elasticsearch.service', '8.12.0', 'active', 'elasticsearch', '/lib/systemd/system/elasticsearch.service', 'Elasticsearch', NOW() - INTERVAL '65 days', NOW()),
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '65 days', NOW()),

-- aws-kafka-01 (Amazon Linux 2) Kafka broker
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'kafka.service', '3.7.0', 'active', 'kafka', '/usr/lib/systemd/system/kafka.service', 'Apache Kafka', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '55 days', NOW()),

-- aws-mq-01 (Ubuntu 20.04) RabbitMQ
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'rabbitmq-server.service', '3.13.0', 'active', 'rabbitmq', '/lib/systemd/system/rabbitmq-server.service', 'RabbitMQ broker', NOW() - INTERVAL '50 days', NOW()),
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'sshd.service', '8.4p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '50 days', NOW()),

-- aws-mongo-01 (Ubuntu 22.04) MongoDB
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'mongod.service', '7.0.5', 'active', 'mongod', '/lib/systemd/system/mongod.service', 'MongoDB Database Server', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '48 days', NOW()),

-- aws-zk-01 (Amazon Linux 2) ZooKeeper
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'zookeeper.service', '3.9.1', 'active', 'zookeeper', '/usr/lib/systemd/system/zookeeper.service', 'Apache ZooKeeper', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '45 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================

-- aws-eks-master-01 (Amazon Linux 2023) EKS 控制面
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'kubelet.service', '1.29.1', 'active', 'root', '/usr/lib/systemd/system/kubelet.service', 'kubelet: The Kubernetes Node Agent', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'containerd.service', '1.7.11', 'active', 'root', '/usr/lib/systemd/system/containerd.service', 'containerd container runtime', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'sshd.service', '8.2p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '60 days', NOW()),

-- aws-eks-node-01 (Amazon Linux 2023) EKS 工作节点
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'kubelet.service', '1.29.1', 'active', 'root', '/usr/lib/systemd/system/kubelet.service', 'kubelet: The Kubernetes Node Agent', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'containerd.service', '1.7.11', 'active', 'root', '/usr/lib/systemd/system/containerd.service', 'containerd container runtime', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'sshd.service', '8.2p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '58 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================

-- aws-jenkins-01 (Ubuntu 22.04) Jenkins CI
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'jenkins.service', '2.440.1', 'active', 'jenkins', '/lib/systemd/system/jenkins.service', 'Jenkins Continuous Integration Server', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '100 days', NOW()),

-- aws-gitlab-01 (Ubuntu 22.04) GitLab
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'gitlab-runsvdir.service', '16.8.1', 'active', 'root', '/lib/systemd/system/gitlab-runsvdir.service', 'GitLab Runit supervisor', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '95 days', NOW()),

-- aws-harbor-01 (Ubuntu 22.04) Harbor 镜像仓库
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'docker.service', '25.0.3', 'active', 'root', '/lib/systemd/system/docker.service', 'Docker Application Container Engine', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '88 days', NOW()),

-- aws-nexus-01 (Ubuntu 22.04) Nexus 制品库
('agent-031-q1r2s3t4', 'aws-nexus-01', '10.0.5.13', 'linux', 'nexus.service', '3.64.0', 'active', 'nexus', '/lib/systemd/system/nexus.service', 'Sonatype Nexus Repository Manager', NOW() - INTERVAL '82 days', NOW()),
('agent-031-q1r2s3t4', 'aws-nexus-01', '10.0.5.13', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '82 days', NOW()),

-- aws-sonar-01 (Ubuntu 22.04) SonarQube 代码质量
('agent-032-u5v6w7x8', 'aws-sonar-01', '10.0.5.14', 'linux', 'sonarqube.service', '10.3.0', 'active', 'sonarqube', '/lib/systemd/system/sonarqube.service', 'SonarQube code quality platform', NOW() - INTERVAL '76 days', NOW()),
('agent-032-u5v6w7x8', 'aws-sonar-01', '10.0.5.14', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '76 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================

-- aws-prometheus-01 (Ubuntu 22.04) Prometheus
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'prometheus.service', '2.49.1', 'active', 'prometheus', '/lib/systemd/system/prometheus.service', 'Prometheus monitoring system', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '110 days', NOW()),

-- aws-grafana-01 (Ubuntu 22.04) Grafana
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'grafana-server.service', '10.3.1', 'active', 'grafana', '/lib/systemd/system/grafana-server.service', 'Grafana instance', NOW() - INTERVAL '105 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '105 days', NOW()),

-- aws-elk-01 (Ubuntu 22.04) ELK 日志集群
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'elasticsearch.service', '8.12.0', 'active', 'elasticsearch', '/lib/systemd/system/elasticsearch.service', 'Elasticsearch', NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '70 days', NOW()),

-- aws-elk-02 (Ubuntu 22.04) ELK 日志集群
('agent-036-k1l2m3n4', 'aws-elk-02', '10.0.6.13', 'linux', 'logstash.service', '8.12.0', 'active', 'logstash', '/lib/systemd/system/logstash.service', 'Logstash', NOW() - INTERVAL '68 days', NOW()),
('agent-036-k1l2m3n4', 'aws-elk-02', '10.0.6.13', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '68 days', NOW()),

-- aws-alertmanager-01 (Ubuntu 22.04) Alertmanager
('agent-037-o5p6q7r8', 'aws-alertmanager-01', '10.0.6.14', 'linux', 'alertmanager.service', '0.27.0', 'active', 'prometheus', '/lib/systemd/system/alertmanager.service', 'Prometheus Alertmanager', NOW() - INTERVAL '65 days', NOW()),
('agent-037-o5p6q7r8', 'aws-alertmanager-01', '10.0.6.14', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '65 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================

-- aws-vpn-01 (Ubuntu 22.04) VPN 网关
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'openvpn@server.service', '2.6.8', 'active', 'root', '/lib/systemd/system/openvpn@.service', 'OpenVPN service for server', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '120 days', NOW()),

-- aws-bastion-01 (Amazon Linux 2023) 堡垒机
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'amazon-ssm-agent.service', '3.2.1630', 'active', 'root', '/usr/lib/systemd/system/amazon-ssm-agent.service', 'Amazon SSM Agent', NOW() - INTERVAL '45 days', NOW()),
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'sshd.service', '8.2p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '45 days', NOW()),

-- aws-dns-01 (Amazon Linux 2) DNS 服务器
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 'named.service', '9.11.36', 'active', 'named', '/usr/lib/systemd/system/named.service', 'Berkeley Internet Name Domain (DNS)', NOW() - INTERVAL '115 days', NOW()),
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 'sshd.service', '7.4p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '115 days', NOW()),

-- aws-mail-01 (Ubuntu 22.04) 邮件服务器
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'postfix.service', '3.6.4', 'active', 'root', '/lib/systemd/system/postfix.service', 'Postfix Mail Transport Agent', NOW() - INTERVAL '90 days', NOW()),
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '90 days', NOW()),

-- aws-proxy-01 (Amazon Linux 2023) 反向代理
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'nginx.service', '1.24.0', 'active', 'root', '/usr/lib/systemd/system/nginx.service', 'A high performance web server and reverse proxy', NOW() - INTERVAL '80 days', NOW()),
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'sshd.service', '8.2p1', 'active', 'root', '/usr/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '80 days', NOW()),

-- aws-backup-01 (Ubuntu 20.04) 备份服务器
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'minio.service', 'RELEASE.2024-01-16', 'active', 'minio', '/lib/systemd/system/minio.service', 'MinIO Object Storage', NOW() - INTERVAL '100 days', NOW()),
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'sshd.service', '8.4p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '100 days', NOW()),

-- aws-vault-01 (Ubuntu 22.04) HashiCorp Vault
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'vault.service', '1.15.4', 'active', 'vault', '/lib/systemd/system/vault.service', 'HashiCorp Vault secret management', NOW() - INTERVAL '35 days', NOW()),
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'sshd.service', '8.9p1', 'active', 'root', '/lib/systemd/system/sshd.service', 'OpenSSH server daemon', NOW() - INTERVAL '35 days', NOW());
