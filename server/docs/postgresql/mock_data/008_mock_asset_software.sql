-- =====================================================
-- 模拟数据: asset_software (软件资产表)
-- 数据量: 100条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的主机生成软件包数据
-- 包类型: dpkg(Ubuntu), rpm(Amazon Linux), pypi, jar
-- =====================================================

INSERT INTO asset_software (agent_id, host_name, host_ip, os_type, name, version, type, source, status, vendor, path, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================

-- aws-web-01 (Ubuntu 22.04) - Nginx Web 服务器
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'nginx', '1.24.0-1ubuntu1', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'openssh-server', '1:8.9p1-3ubuntu0.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'curl', '7.81.0-1ubuntu1.16', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'openssl', '3.0.2-0ubuntu1.14', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'fail2ban', '0.11.2-6ubuntu2', 'dpkg', 'ubuntu', 'installed', 'Cyril Jaquier', NULL, NOW() - INTERVAL '90 days', NOW()),

-- aws-web-02 (Ubuntu 22.04) - Nginx Web 服务器
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'nginx', '1.24.0-1ubuntu1', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'openssh-server', '1:8.9p1-3ubuntu0.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '88 days', NOW()),

-- aws-api-01 (Ubuntu 22.04) - Spring Boot API 服务器
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'openjdk-17-jdk', '17.0.9+9-1~22.04', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'openssh-server', '1:8.9p1-3ubuntu0.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'vim', '2:8.2.3995-1ubuntu2.15', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'spring-boot', '3.2.1', 'jar', 'maven', 'installed', 'Pivotal Software', '/opt/api/lib/spring-boot-3.2.1.jar', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'spring-web', '6.1.2', 'jar', 'maven', 'installed', 'Pivotal Software', '/opt/api/lib/spring-web-6.1.2.jar', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'jackson-databind', '2.16.1', 'jar', 'maven', 'installed', 'FasterXML', '/opt/api/lib/jackson-databind-2.16.1.jar', NOW() - INTERVAL '85 days', NOW()),

-- aws-api-02 (Ubuntu 22.04) - Spring Boot API 服务器
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'spring-boot', '3.2.1', 'jar', 'maven', 'installed', 'Pivotal Software', '/opt/api/lib/spring-boot-3.2.1.jar', NOW() - INTERVAL '85 days', NOW()),
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'log4j-core', '2.22.1', 'jar', 'maven', 'installed', 'Apache Software Foundation', '/opt/api/lib/log4j-core-2.22.1.jar', NOW() - INTERVAL '85 days', NOW()),
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'slf4j-api', '2.0.11', 'jar', 'maven', 'installed', 'QOS.ch', '/opt/api/lib/slf4j-api-2.0.11.jar', NOW() - INTERVAL '85 days', NOW()),

-- aws-gateway-01 (Amazon Linux 2023) - API 网关
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'nginx', '1.24.0-1.amzn2023.0.2', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'openssh-server', '8.7p1-8.amzn2023.0.6', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'openssl', '3.0.8-1.amzn2023.0.8', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '150 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================

-- aws-app-01 (Ubuntu 22.04) - Java 应用服务器
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'openjdk-17-jdk', '17.0.9+9-1~22.04', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'spring-boot', '3.2.0', 'jar', 'maven', 'installed', 'Pivotal Software', '/opt/tomcat/webapps/app/WEB-INF/lib/spring-boot-3.2.0.jar', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'spring-core', '6.1.1', 'jar', 'maven', 'installed', 'Pivotal Software', '/opt/tomcat/webapps/app/WEB-INF/lib/spring-core-6.1.1.jar', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'hibernate-core', '6.4.0.Final', 'jar', 'maven', 'installed', 'Red Hat', '/opt/tomcat/webapps/app/WEB-INF/lib/hibernate-core-6.4.0.Final.jar', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'mysql-connector-java', '8.2.0', 'jar', 'maven', 'installed', 'Oracle Corporation', '/opt/tomcat/webapps/app/WEB-INF/lib/mysql-connector-java-8.2.0.jar', NOW() - INTERVAL '80 days', NOW()),

-- aws-app-02 (Ubuntu 22.04) - Python/Django 应用服务器
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'python3', '3.10.12-1~22.04.3', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'django', '4.2.9', 'pypi', 'pypi', 'installed', 'Django Software Foundation', NULL, NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'requests', '2.31.0', 'pypi', 'pypi', 'installed', 'Kenneth Reitz', NULL, NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'celery', '5.3.6', 'pypi', 'pypi', 'installed', 'Ask Solem', NULL, NOW() - INTERVAL '78 days', NOW()),

-- aws-app-03 (Amazon Linux 2023) - Go 应用服务器
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'golang', '1.21.5-1.amzn2023.0.1', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '75 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'openssh-server', '8.7p1-8.amzn2023.0.6', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '75 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'curl', '8.5.0-1.amzn2023.0.1', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '75 days', NOW()),

-- aws-worker-01 (Ubuntu 20.04) - Celery Worker
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'python3', '3.8.10-0ubuntu1~20.04.9', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'celery', '5.3.6', 'pypi', 'pypi', 'installed', 'Ask Solem', NULL, NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'redis', '5.0.1', 'pypi', 'pypi', 'installed', 'Redis Inc.', NULL, NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================

-- aws-mysql-01 (Amazon Linux 2) - MySQL 主库
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'mysql-community-server', '8.0.36-1.el7', 'rpm', 'amzn', 'installed', 'Oracle Corporation', NULL, NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'mysql-community-client', '8.0.36-1.el7', 'rpm', 'amzn', 'installed', 'Oracle Corporation', NULL, NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'openssh-server', '7.4p1-22.amzn2.0.2', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'openssl', '1.0.2k-24.amzn2.0.10', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '95 days', NOW()),

-- aws-mysql-02 (Amazon Linux 2) - MySQL 从库
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'mysql-community-server', '8.0.36-1.el7', 'rpm', 'amzn', 'installed', 'Oracle Corporation', NULL, NOW() - INTERVAL '93 days', NOW()),
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'mysql-community-client', '8.0.36-1.el7', 'rpm', 'amzn', 'installed', 'Oracle Corporation', NULL, NOW() - INTERVAL '93 days', NOW()),

-- aws-pg-01 (Amazon Linux 2) - PostgreSQL
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'postgresql15-server', '15.5-1PGDG.amzn2', 'rpm', 'amzn', 'installed', 'PostgreSQL Global Development Group', NULL, NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'postgresql15', '15.5-1PGDG.amzn2', 'rpm', 'amzn', 'installed', 'PostgreSQL Global Development Group', NULL, NOW() - INTERVAL '90 days', NOW()),

-- aws-redis-01 (Amazon Linux 2) - Redis 主
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'redis', '7.2.3-1.amzn2', 'rpm', 'amzn', 'installed', 'Redis Ltd.', NULL, NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'redis-exporter', '1.55.0', 'rpm', 'amzn', 'installed', 'Oliver Letterer', NULL, NOW() - INTERVAL '75 days', NOW()),

-- aws-redis-02 (Amazon Linux 2) - Redis 从
('agent-015-e7f8g9h0', 'aws-redis-02', '10.0.3.21', 'linux', 'redis', '7.2.3-1.amzn2', 'rpm', 'amzn', 'installed', 'Redis Ltd.', NULL, NOW() - INTERVAL '73 days', NOW()),

-- aws-es-01 (Ubuntu 22.04) - Elasticsearch 数据节点
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'elasticsearch', '8.11.3', 'dpkg', 'ubuntu', 'installed', 'Elastic NV', NULL, NOW() - INTERVAL '65 days', NOW()),
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'openjdk-17-jre', '17.0.9+9-1~22.04', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '65 days', NOW()),

-- aws-kafka-01 (Amazon Linux 2) - Kafka Broker
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'java-17-openjdk', '17.0.9.0.9-2.amzn2', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'kafka', '3.6.1', 'jar', 'maven', 'installed', 'Apache Software Foundation', '/opt/kafka/libs/kafka_2.13-3.6.1.jar', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'kafka-clients', '3.6.1', 'jar', 'maven', 'installed', 'Apache Software Foundation', '/opt/kafka/libs/kafka-clients-3.6.1.jar', NOW() - INTERVAL '55 days', NOW()),

-- aws-kafka-02 (Amazon Linux 2) - Kafka Broker
('agent-020-y7z8a9b0', 'aws-kafka-02', '10.0.3.41', 'linux', 'java-17-openjdk', '17.0.9.0.9-2.amzn2', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '53 days', NOW()),
('agent-020-y7z8a9b0', 'aws-kafka-02', '10.0.3.41', 'linux', 'kafka', '3.6.1', 'jar', 'maven', 'installed', 'Apache Software Foundation', '/opt/kafka/libs/kafka_2.13-3.6.1.jar', NOW() - INTERVAL '53 days', NOW()),

-- aws-mq-01 (Ubuntu 20.04) - RabbitMQ
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'rabbitmq-server', '3.12.12-1ubuntu1', 'dpkg', 'ubuntu', 'installed', 'Pivotal Software', NULL, NOW() - INTERVAL '50 days', NOW()),
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'erlang-base', '1:25.3.2.8+dfsg-1', 'dpkg', 'ubuntu', 'installed', 'Erlang/OTP', NULL, NOW() - INTERVAL '50 days', NOW()),

-- aws-mongo-01 (Ubuntu 22.04) - MongoDB
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'mongodb-org-server', '7.0.4', 'dpkg', 'ubuntu', 'installed', 'MongoDB Inc.', NULL, NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'mongodb-org-tools', '7.0.4', 'dpkg', 'ubuntu', 'installed', 'MongoDB Inc.', NULL, NOW() - INTERVAL '48 days', NOW()),

-- aws-zk-01 (Amazon Linux 2) - Zookeeper
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'java-17-openjdk', '17.0.9.0.9-2.amzn2', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'zookeeper', '3.8.3', 'jar', 'maven', 'installed', 'Apache Software Foundation', '/opt/zookeeper/lib/zookeeper-3.8.3.jar', NOW() - INTERVAL '45 days', NOW()),

-- aws-consul-01 (Ubuntu 22.04) - Consul 服务发现
('agent-049-k3l4m5n6', 'aws-consul-01', '10.0.3.72', 'linux', 'consul', '1.17.1-1', 'dpkg', 'ubuntu', 'installed', 'HashiCorp', NULL, NOW() - INTERVAL '40 days', NOW()),
('agent-049-k3l4m5n6', 'aws-consul-01', '10.0.3.72', 'linux', 'openssh-server', '1:8.9p1-3ubuntu0.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '40 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================

-- aws-eks-master-01 (Amazon Linux 2023) - EKS 控制面
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'kubelet', '1.28.4-1.amzn2023', 'rpm', 'amzn', 'installed', 'Kubernetes Authors', NULL, NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'kubeadm', '1.28.4-1.amzn2023', 'rpm', 'amzn', 'installed', 'Kubernetes Authors', NULL, NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'kubectl', '1.28.4-1.amzn2023', 'rpm', 'amzn', 'installed', 'Kubernetes Authors', NULL, NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'containerd.io', '1.7.11-1.amzn2023', 'rpm', 'amzn', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'helm', '3.13.3-1.amzn2023', 'rpm', 'amzn', 'installed', 'Helm Authors', NULL, NOW() - INTERVAL '60 days', NOW()),

-- aws-eks-node-01 (Amazon Linux 2023) - EKS 工作节点
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'kubelet', '1.28.4-1.amzn2023', 'rpm', 'amzn', 'installed', 'Kubernetes Authors', NULL, NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'containerd.io', '1.7.11-1.amzn2023', 'rpm', 'amzn', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '58 days', NOW()),

-- aws-eks-node-02 (Amazon Linux 2023) - EKS 工作节点
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 'kubelet', '1.28.4-1.amzn2023', 'rpm', 'amzn', 'installed', 'Kubernetes Authors', NULL, NOW() - INTERVAL '56 days', NOW()),
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 'containerd.io', '1.7.11-1.amzn2023', 'rpm', 'amzn', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '56 days', NOW()),

-- aws-eks-node-03 (Amazon Linux 2023) - EKS 工作节点
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'linux', 'kubelet', '1.28.4-1.amzn2023', 'rpm', 'amzn', 'installed', 'Kubernetes Authors', NULL, NOW() - INTERVAL '54 days', NOW()),
('agent-026-w1x2y3z4', 'aws-eks-node-03', '10.0.4.13', 'linux', 'containerd.io', '1.7.11-1.amzn2023', 'rpm', 'amzn', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '54 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================

-- aws-jenkins-01 (Ubuntu 22.04) - Jenkins CI/CD
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'jenkins', '2.432.1', 'dpkg', 'ubuntu', 'installed', 'Jenkins Project', NULL, NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'docker-ce', '24.0.7-1~ubuntu.22.04~jammy', 'dpkg', 'ubuntu', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'openjdk-17-jdk', '17.0.9+9-1~22.04', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'ansible', '8.6.1', 'pypi', 'pypi', 'installed', 'Red Hat', NULL, NOW() - INTERVAL '100 days', NOW()),

-- aws-gitlab-01 (Ubuntu 22.04) - GitLab
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'gitlab-ce', '16.7.2-ce.0', 'dpkg', 'ubuntu', 'installed', 'GitLab Inc.', NULL, NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'openssh-server', '1:8.9p1-3ubuntu0.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '95 days', NOW()),

-- aws-harbor-01 (Ubuntu 22.04) - Harbor 镜像仓库
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'docker-ce', '24.0.7-1~ubuntu.22.04~jammy', 'dpkg', 'ubuntu', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'docker-compose-plugin', '2.23.3-1~ubuntu.22.04~jammy', 'dpkg', 'ubuntu', 'installed', 'Docker Inc.', NULL, NOW() - INTERVAL '88 days', NOW()),

-- aws-sonar-01 (Ubuntu 22.04) - SonarQube 代码扫描
('agent-032-u5v6w7x8', 'aws-sonar-01', '10.0.5.14', 'linux', 'openjdk-17-jre', '17.0.9+9-1~22.04', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '76 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================

-- aws-prometheus-01 (Ubuntu 22.04) - Prometheus 监控
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'prometheus', '2.48.1', 'dpkg', 'ubuntu', 'installed', 'Prometheus Authors', NULL, NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'alertmanager', '0.27.0', 'dpkg', 'ubuntu', 'installed', 'Prometheus Authors', NULL, NOW() - INTERVAL '110 days', NOW()),

-- aws-grafana-01 (Ubuntu 22.04) - Grafana 可视化
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'grafana', '10.2.3', 'dpkg', 'ubuntu', 'installed', 'Grafana Labs', NULL, NOW() - INTERVAL '105 days', NOW()),

-- aws-elk-01 (Ubuntu 22.04) - ELK 日志主节点
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'elasticsearch', '8.11.3', 'dpkg', 'ubuntu', 'installed', 'Elastic NV', NULL, NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'kibana', '8.11.3', 'dpkg', 'ubuntu', 'installed', 'Elastic NV', NULL, NOW() - INTERVAL '70 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'logstash', '8.11.3', 'dpkg', 'ubuntu', 'installed', 'Elastic NV', NULL, NOW() - INTERVAL '70 days', NOW()),

-- aws-elk-02 (Ubuntu 22.04) - ELK 日志副节点
('agent-036-k1l2m3n4', 'aws-elk-02', '10.0.6.13', 'linux', 'elasticsearch', '8.11.3', 'dpkg', 'ubuntu', 'installed', 'Elastic NV', NULL, NOW() - INTERVAL '68 days', NOW()),

-- aws-alertmanager-01 (Ubuntu 22.04) - Alertmanager
('agent-037-o5p6q7r8', 'aws-alertmanager-01', '10.0.6.14', 'linux', 'alertmanager', '0.27.0', 'dpkg', 'ubuntu', 'installed', 'Prometheus Authors', NULL, NOW() - INTERVAL '65 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================

-- aws-vpn-01 (Ubuntu 22.04) - VPN 服务器
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'openvpn', '2.5.9-0ubuntu0.22.04.2', 'dpkg', 'ubuntu', 'installed', 'OpenVPN Inc.', NULL, NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'easy-rsa', '3.1.1-1', 'dpkg', 'ubuntu', 'installed', 'OpenVPN Inc.', NULL, NOW() - INTERVAL '120 days', NOW()),

-- aws-bastion-01 (Amazon Linux 2023) - 跳板机
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'openssh-server', '8.7p1-8.amzn2023.0.6', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '45 days', NOW()),
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'vim-enhanced', '9.0.2081-1.amzn2023.0.1', 'rpm', 'amzn', 'installed', 'Amazon', NULL, NOW() - INTERVAL '45 days', NOW()),

-- aws-dns-01 (Amazon Linux 2) - DNS 服务器
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 'bind', '9.11.4-26.P2.amzn2.13', 'rpm', 'amzn', 'installed', 'ISC', NULL, NOW() - INTERVAL '115 days', NOW()),

-- aws-nfs-01 (Ubuntu 20.04) - NFS 文件服务器
('agent-041-e1f2g3h4', 'aws-nfs-01', '10.0.7.13', 'linux', 'nfs-kernel-server', '1:1.3.4-2.5ubuntu3.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '100 days', NOW()),

-- aws-mail-01 (Ubuntu 22.04) - 邮件服务器
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'postfix', '3.6.4-1ubuntu1.3', 'dpkg', 'ubuntu', 'installed', 'Wietse Venema', NULL, NOW() - INTERVAL '90 days', NOW()),

-- aws-ldap-01 (Ubuntu 22.04) - LDAP 目录服务
('agent-043-m9n0o1p2', 'aws-ldap-01', '10.0.7.15', 'linux', 'slapd', '2.5.16+dfsg-0ubuntu0.22.04.2', 'dpkg', 'ubuntu', 'installed', 'OpenLDAP Foundation', NULL, NOW() - INTERVAL '85 days', NOW()),

-- aws-proxy-01 (Amazon Linux 2023) - 反向代理/负载均衡
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'haproxy', '2.8.3-1.amzn2023.0.1', 'rpm', 'amzn', 'installed', 'HAProxy Technologies', NULL, NOW() - INTERVAL '80 days', NOW()),

-- aws-vault-01 (Ubuntu 22.04) - HashiCorp Vault
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'vault', '1.15.4-1', 'dpkg', 'ubuntu', 'installed', 'HashiCorp', NULL, NOW() - INTERVAL '35 days', NOW()),
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'openssh-server', '1:8.9p1-3ubuntu0.6', 'dpkg', 'ubuntu', 'installed', 'Canonical Ltd.', NULL, NOW() - INTERVAL '35 days', NOW());
