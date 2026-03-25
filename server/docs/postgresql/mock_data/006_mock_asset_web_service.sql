-- =====================================================
-- 模拟数据: asset_web_service (Web服务资产表)
-- 数据量: 50条
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
-- 注意: (agent_id, server_type) 必须唯一
-- =====================================================

INSERT INTO asset_web_service (agent_id, host_name, host_ip, os_type, name, version, server_type, site_domain, path, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================
-- aws-web-01: Nginx 反向代理 + 前端静态站点
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'company-website', '1.24.0', 'Nginx', 'www.company.com', '/var/www/html', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'frontend-ssr', '18.19.0', 'Node.js', 'app.company.com', '/opt/apps/frontend', NOW() - INTERVAL '88 days', NOW()),
-- aws-web-02: 备份 Web 节点
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'cdn-proxy', '1.24.0', 'Nginx', 'cdn.company.com', '/var/www/cdn', NOW() - INTERVAL '88 days', NOW()),
-- aws-api-01: Spring Boot API 服务
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'user-service', '10.1.16', 'Tomcat', 'api.company.com', '/opt/tomcat/webapps', NOW() - INTERVAL '85 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'linux', 'api-docs', '4.18.2', 'Swagger-UI', 'api-docs.company.com', '/opt/apps/swagger-ui', NOW() - INTERVAL '83 days', NOW()),
-- aws-api-02: Spring Boot API 服务 (备)
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'linux', 'order-service', '10.1.16', 'Tomcat', 'api.company.com', '/opt/tomcat/webapps', NOW() - INTERVAL '85 days', NOW()),
-- aws-gateway-01: API 网关
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'api-gateway', '2.10.7', 'Traefik', 'gateway.company.com', '/etc/traefik', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'haproxy-stats', '2.8.4', 'HAProxy', 'lb-stats.company.com', '/etc/haproxy', NOW() - INTERVAL '148 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================
-- aws-app-01: Java 微服务
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'payment-service', '10.1.16', 'Tomcat', 'pay.company.com', '/opt/tomcat/webapps', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'admin-portal', '1.24.0', 'Nginx', 'admin.company.com', '/var/www/admin', NOW() - INTERVAL '78 days', NOW()),
-- aws-app-02: Java 微服务
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'inventory-service', '10.1.16', 'Tomcat', 'inventory.company.com', '/opt/tomcat/webapps', NOW() - INTERVAL '78 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'linux', 'notification-service', '4.0.0', 'Uvicorn', 'notify.company.com', '/opt/apps/notification', NOW() - INTERVAL '76 days', NOW()),
-- aws-app-03: Python 微服务
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'ml-service', '4.0.0', 'Gunicorn', 'ml.company.com', '/opt/apps/ml-service', NOW() - INTERVAL '75 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'linux', 'data-api', '4.0.0', 'Uvicorn', 'data-api.company.com', '/opt/apps/data-api', NOW() - INTERVAL '73 days', NOW()),
-- aws-worker-01: 异步任务节点
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'celery-flower', '2.0.1', 'Flower', 'flower.company.com', '/opt/apps/flower', NOW() - INTERVAL '70 days', NOW()),
-- aws-worker-02: 异步任务节点
('agent-010-k7l8m9n0', 'aws-worker-02', '10.0.2.21', 'linux', 'airflow-web', '2.8.1', 'Gunicorn', 'airflow.company.com', '/opt/airflow', NOW() - INTERVAL '68 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================
-- aws-mysql-01: MySQL 主库管理
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'phpmyadmin', '5.2.1', 'Apache', 'pma.company.com', '/usr/share/phpmyadmin', NOW() - INTERVAL '95 days', NOW()),
-- aws-mysql-02: MySQL 从库管理
('agent-012-s5t6u7v8', 'aws-mysql-02', '10.0.3.11', 'linux', 'mysql-exporter-web', '0.15.1', 'Prometheus-Exporter', 'mysql-metrics.company.com', '/opt/mysqld_exporter', NOW() - INTERVAL '93 days', NOW()),
-- aws-pg-01: PostgreSQL 管理
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'pgadmin', '8.1', 'Gunicorn', 'pgadmin.company.com', '/usr/pgadmin4', NOW() - INTERVAL '90 days', NOW()),
-- aws-redis-01: Redis 可视化管理
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'redis-insight', '2.38.0', 'Redis-Insight', 'redis.company.com', '/opt/redis-insight', NOW() - INTERVAL '75 days', NOW()),
-- aws-es-01: Elasticsearch HQ
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'linux', 'elasticsearch-hq', '11.0.18', 'Jetty', 'es.company.com', '/opt/elasticsearch-hq', NOW() - INTERVAL '65 days', NOW()),
-- aws-kafka-01: Kafka 管理 UI
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'kafka-ui', '0.7.1', 'Kafka-UI', 'kafka.company.com', '/opt/kafka-ui', NOW() - INTERVAL '55 days', NOW()),
-- aws-mq-01: RabbitMQ 管理
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'linux', 'rabbitmq-mgmt', '3.12.10', 'RabbitMQ', 'mq.company.com', '/usr/lib/rabbitmq', NOW() - INTERVAL '50 days', NOW()),
-- aws-mongo-01: Mongo Express
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'mongo-express', '1.0.2', 'Node.js', 'mongo.company.com', '/opt/mongo-express', NOW() - INTERVAL '48 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================
-- aws-eks-master-01: K8s Dashboard + Rancher
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'kubernetes-dashboard', '2.7.0', 'Go', 'k8s.company.com', '/opt/kubernetes-dashboard', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'rancher', '2.8.0', 'Rancher', 'rancher.company.com', '/opt/rancher', NOW() - INTERVAL '58 days', NOW()),
-- aws-eks-node-01: Ingress Controller
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'ingress-nginx', '1.9.5', 'Nginx', 'ingress.company.com', '/etc/nginx', NOW() - INTERVAL '58 days', NOW()),
-- aws-eks-node-02: 服务网格
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'linux', 'istio-pilot', '1.20.2', 'Go', 'istio.company.com', '/opt/istio', NOW() - INTERVAL '56 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================
-- aws-jenkins-01: CI/CD 平台
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'jenkins', '2.426.2', 'Jenkins', 'ci.company.com', '/var/lib/jenkins', NOW() - INTERVAL '100 days', NOW()),
-- aws-gitlab-01: 代码仓库
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'gitlab-web', '1.24.0', 'Nginx', 'git.company.com', '/opt/gitlab/embedded/sbin', NOW() - INTERVAL '95 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'linux', 'gitlab', '16.6.1', 'Puma', 'git.company.com', '/opt/gitlab/embedded/service/gitlab-rails', NOW() - INTERVAL '93 days', NOW()),
-- aws-harbor-01: 容器镜像仓库
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'harbor-portal', '1.22.1', 'Nginx', 'harbor.company.com', '/etc/nginx', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'harbor-core', '2.9.1', 'Go', 'harbor.company.com', '/harbor', NOW() - INTERVAL '86 days', NOW()),
-- aws-nexus-01: 制品仓库
('agent-031-q1r2s3t4', 'aws-nexus-01', '10.0.5.13', 'linux', 'nexus', '3.63.0', 'Nexus', 'nexus.company.com', '/opt/nexus', NOW() - INTERVAL '82 days', NOW()),
-- aws-sonar-01: 代码质量
('agent-032-u5v6w7x8', 'aws-sonar-01', '10.0.5.14', 'linux', 'sonarqube', '10.3.0', 'SonarQube', 'sonar.company.com', '/opt/sonarqube', NOW() - INTERVAL '76 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================
-- aws-prometheus-01: Prometheus 监控
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'prometheus', '2.47.2', 'Prometheus', 'prometheus.company.com', '/usr/local/prometheus', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'thanos-query', '0.34.0', 'Go', 'thanos.company.com', '/opt/thanos', NOW() - INTERVAL '108 days', NOW()),
-- aws-grafana-01: Grafana 可视化
('agent-034-c3d4e5f6', 'aws-grafana-01', '10.0.6.11', 'linux', 'grafana', '10.2.2', 'Grafana', 'grafana.company.com', '/usr/share/grafana', NOW() - INTERVAL '105 days', NOW()),
-- aws-elk-01: 日志分析 (Kibana)
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'linux', 'kibana', '8.11.3', 'Kibana', 'kibana.company.com', '/usr/share/kibana', NOW() - INTERVAL '70 days', NOW()),
-- aws-elk-02: 日志采集 (Logstash Web)
('agent-036-k1l2m3n4', 'aws-elk-02', '10.0.6.13', 'linux', 'logstash-web', '8.11.3', 'Logstash', 'logstash.company.com', '/usr/share/logstash', NOW() - INTERVAL '68 days', NOW()),
-- aws-alertmanager-01: 告警管理
('agent-037-o5p6q7r8', 'aws-alertmanager-01', '10.0.6.14', 'linux', 'alertmanager', '0.26.0', 'AlertManager', 'alerts.company.com', '/usr/local/alertmanager', NOW() - INTERVAL '65 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================
-- aws-vpn-01: VPN 管理界面
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'openvpn-admin', '2.6.8', 'Nginx', 'vpn.company.com', '/var/www/openvpn-admin', NOW() - INTERVAL '120 days', NOW()),
-- aws-bastion-01: 堡垒机 Web 终端
('agent-039-w3x4y5z6', 'aws-bastion-01', '10.0.7.11', 'linux', 'jumpserver', '3.8.0', 'Gunicorn', 'jump.company.com', '/opt/jumpserver', NOW() - INTERVAL '45 days', NOW()),
-- aws-dns-01: DNS 管理
('agent-040-a7b8c9d0', 'aws-dns-01', '10.0.7.12', 'linux', 'powerdns-admin', '0.4.1', 'Gunicorn', 'dns.company.com', '/opt/powerdns-admin', NOW() - INTERVAL '115 days', NOW()),
-- aws-mail-01: 邮件 Web 界面
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'linux', 'webmail', '1.6.5', 'Apache', 'mail.company.com', '/var/www/roundcube', NOW() - INTERVAL '90 days', NOW()),
-- aws-ldap-01: LDAP 管理
('agent-043-m9n0o1p2', 'aws-ldap-01', '10.0.7.15', 'linux', 'ldap-admin', '2.4.0', 'uWSGI', 'ldap.company.com', '/opt/ldap-admin', NOW() - INTERVAL '85 days', NOW()),
-- aws-proxy-01: 正向代理管理
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'linux', 'squid-analyzer', '6.6', 'Apache', 'proxy-report.company.com', '/var/www/squid-analyzer', NOW() - INTERVAL '80 days', NOW()),
-- aws-backup-01: MinIO 对象存储
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'minio-console', 'RELEASE.2023-12-20', 'MinIO', 's3.company.com', '/opt/minio', NOW() - INTERVAL '100 days', NOW()),
-- aws-consul-01: Consul 服务发现
('agent-049-k3l4m5n6', 'aws-consul-01', '10.0.3.72', 'linux', 'consul-ui', '1.17.1', 'Consul', 'consul.company.com', '/opt/consul', NOW() - INTERVAL '40 days', NOW()),
-- aws-vault-01: HashiCorp Vault
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'linux', 'vault-ui', '1.15.4', 'Vault', 'vault.company.com', '/opt/vault', NOW() - INTERVAL '35 days', NOW());
