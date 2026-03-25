-- =====================================================
-- 模拟数据: asset_database (数据库资产表)
-- 数据量: 50条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的 AWS 主机生成数据库数据
-- OS: 全部为 Linux (无 Windows / 无 SQL Server)
-- =====================================================

INSERT INTO asset_database (agent_id, host_name, host_ip, os_type, db_type, db_version, port, run_user, created_at, updated_at) VALUES
-- ==========================================
-- MySQL 数据库
-- ==========================================
('agent-011-o1p2q3r4', 'aws-mysql-01',      '10.0.3.10', 'linux', 'MySQL', '8.0.35', 3306, 'mysql', NOW() - INTERVAL '95 days', NOW()),
('agent-012-s5t6u7v8', 'aws-mysql-02',      '10.0.3.11', 'linux', 'MySQL', '8.0.35', 3306, 'mysql', NOW() - INTERVAL '93 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01',        '10.0.2.10', 'linux', 'MySQL', '8.0.35', 3306, 'mysql', NOW() - INTERVAL '80 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02',        '10.0.2.11', 'linux', 'MySQL', '8.0.35', 3306, 'mysql', NOW() - INTERVAL '78 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01',    '10.0.5.10', 'linux', 'MySQL', '8.0.32', 3306, 'mysql', NOW() - INTERVAL '100 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01',     '10.0.5.11', 'linux', 'MySQL', '8.0.35', 3306, 'mysql', NOW() - INTERVAL '95 days', NOW()),
-- ==========================================
-- PostgreSQL 数据库
-- ==========================================
('agent-013-w9x0y1z2', 'aws-pg-01',         '10.0.3.12', 'linux', 'PostgreSQL', '15.4', 5432, 'postgres', NOW() - INTERVAL '90 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01',     '10.0.5.12', 'linux', 'PostgreSQL', '13.13', 5432, 'postgres', NOW() - INTERVAL '88 days', NOW()),
('agent-032-u5v6w7x8', 'aws-sonar-01',      '10.0.5.14', 'linux', 'PostgreSQL', '15.4', 5432, 'postgres', NOW() - INTERVAL '76 days', NOW()),
('agent-034-c3d4e5f6', 'aws-grafana-01',    '10.0.6.11', 'linux', 'PostgreSQL', '14.10', 5432, 'postgres', NOW() - INTERVAL '105 days', NOW()),
-- ==========================================
-- Redis 数据库
-- ==========================================
('agent-014-a3b4c5d6', 'aws-redis-01',      '10.0.3.20', 'linux', 'Redis', '7.2.3', 6379, 'redis', NOW() - INTERVAL '75 days', NOW()),
('agent-015-e7f8g9h0', 'aws-redis-02',      '10.0.3.21', 'linux', 'Redis', '7.2.3', 6379, 'redis', NOW() - INTERVAL '73 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01',    '10.0.5.10', 'linux', 'Redis', '7.0.15', 6379, 'redis', NOW() - INTERVAL '100 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01',     '10.0.5.11', 'linux', 'Redis', '7.0.15', 6379, 'gitlab-redis', NOW() - INTERVAL '95 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'Redis', '7.2.3', 6379, 'redis', NOW() - INTERVAL '60 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01',        '10.0.2.10', 'linux', 'Redis', '7.2.3', 6379, 'redis', NOW() - INTERVAL '80 days', NOW()),
('agent-031-q1r2s3t4', 'aws-nexus-01',      '10.0.5.13', 'linux', 'Redis', '7.0.15', 6379, 'redis', NOW() - INTERVAL '82 days', NOW()),
-- ==========================================
-- Elasticsearch
-- ==========================================
('agent-016-i1j2k3l4', 'aws-es-01',         '10.0.3.30', 'linux', 'Elasticsearch', '8.11.1', 9200, 'elasticsearch', NOW() - INTERVAL '65 days', NOW()),
('agent-017-m5n6o7p8', 'aws-es-02',         '10.0.3.31', 'linux', 'Elasticsearch', '8.11.1', 9200, 'elasticsearch', NOW() - INTERVAL '63 days', NOW()),
('agent-018-q9r0s1t2', 'aws-es-03',         '10.0.3.32', 'linux', 'Elasticsearch', '8.11.1', 9200, 'elasticsearch', NOW() - INTERVAL '60 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01',        '10.0.6.12', 'linux', 'Elasticsearch', '8.11.1', 9200, 'elasticsearch', NOW() - INTERVAL '70 days', NOW()),
('agent-036-k1l2m3n4', 'aws-elk-02',        '10.0.6.13', 'linux', 'Elasticsearch', '8.11.1', 9200, 'elasticsearch', NOW() - INTERVAL '68 days', NOW()),
-- ==========================================
-- MongoDB
-- ==========================================
('agent-022-g5h6i7j8', 'aws-mongo-01',      '10.0.3.60', 'linux', 'MongoDB', '7.0.4', 27017, 'mongodb', NOW() - INTERVAL '48 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01',        '10.0.2.10', 'linux', 'MongoDB', '7.0.4', 27017, 'mongodb', NOW() - INTERVAL '80 days', NOW()),
-- ==========================================
-- etcd
-- ==========================================
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'etcd', '3.5.10', 2379, 'root', NOW() - INTERVAL '60 days', NOW()),
-- ==========================================
-- ZooKeeper
-- ==========================================
('agent-047-c5d6e7f8', 'aws-zk-01',         '10.0.3.70', 'linux', 'ZooKeeper', '3.8.3', 2181, 'zookeeper', NOW() - INTERVAL '45 days', NOW()),
('agent-048-g9h0i1j2', 'aws-zk-02',         '10.0.3.71', 'linux', 'ZooKeeper', '3.8.3', 2181, 'zookeeper', NOW() - INTERVAL '43 days', NOW()),
-- ==========================================
-- Kafka (内置数据存储)
-- ==========================================
('agent-019-u3v4w5x6', 'aws-kafka-01',      '10.0.3.40', 'linux', 'Kafka', '3.6.1', 9092, 'kafka', NOW() - INTERVAL '55 days', NOW()),
('agent-020-y7z8a9b0', 'aws-kafka-02',      '10.0.3.41', 'linux', 'Kafka', '3.6.1', 9092, 'kafka', NOW() - INTERVAL '53 days', NOW()),
-- ==========================================
-- InfluxDB (时序数据库)
-- ==========================================
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'InfluxDB', '2.7.3', 8086, 'influxdb', NOW() - INTERVAL '110 days', NOW()),
('agent-035-g7h8i9j0', 'aws-elk-01',        '10.0.6.12', 'linux', 'InfluxDB', '2.7.3', 8086, 'influxdb', NOW() - INTERVAL '70 days', NOW()),
-- ==========================================
-- ClickHouse
-- ==========================================
('agent-036-k1l2m3n4', 'aws-elk-02',        '10.0.6.13', 'linux', 'ClickHouse', '23.11.2', 8123, 'clickhouse', NOW() - INTERVAL '68 days', NOW()),
-- ==========================================
-- SQLite (嵌入式)
-- ==========================================
('agent-028-e9f0g1h2', 'aws-jenkins-01',    '10.0.5.10', 'linux', 'SQLite', '3.40.1', 0, 'jenkins', NOW() - INTERVAL '100 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01',     '10.0.5.11', 'linux', 'SQLite', '3.40.1', 0, 'gitlab-rails', NOW() - INTERVAL '95 days', NOW()),
('agent-049-k3l4m5n6', 'aws-consul-01',     '10.0.3.72', 'linux', 'SQLite', '3.40.1', 0, 'consul', NOW() - INTERVAL '40 days', NOW()),
-- ==========================================
-- Cassandra
-- ==========================================
('agent-019-u3v4w5x6', 'aws-kafka-01',      '10.0.3.40', 'linux', 'Cassandra', '4.1.3', 9042, 'cassandra', NOW() - INTERVAL '55 days', NOW()),
('agent-020-y7z8a9b0', 'aws-kafka-02',      '10.0.3.41', 'linux', 'Cassandra', '4.1.3', 9042, 'cassandra', NOW() - INTERVAL '53 days', NOW()),
-- ==========================================
-- Neo4j (图数据库)
-- ==========================================
('agent-008-c9d0e1f2', 'aws-app-03',        '10.0.2.12', 'linux', 'Neo4j', '5.14.0', 7474, 'neo4j', NOW() - INTERVAL '75 days', NOW()),
-- ==========================================
-- CouchDB
-- ==========================================
('agent-007-y5z6a7b8', 'aws-app-02',        '10.0.2.11', 'linux', 'CouchDB', '3.3.3', 5984, 'couchdb', NOW() - INTERVAL '78 days', NOW()),
-- ==========================================
-- RabbitMQ (消息队列内嵌数据库)
-- ==========================================
('agent-021-c1d2e3f4', 'aws-mq-01',         '10.0.3.50', 'linux', 'RabbitMQ', '3.12.10', 5672, 'rabbitmq', NOW() - INTERVAL '50 days', NOW()),
-- ==========================================
-- TiDB
-- ==========================================
('agent-013-w9x0y1z2', 'aws-pg-01',         '10.0.3.12', 'linux', 'TiDB', '7.5.0', 4000, 'tidb', NOW() - INTERVAL '90 days', NOW()),
-- ==========================================
-- CockroachDB
-- ==========================================
('agent-045-u7v8w9x0', 'aws-backup-01',     '10.0.7.17', 'linux', 'CockroachDB', '23.1.13', 26257, 'cockroach', NOW() - INTERVAL '100 days', NOW()),
-- ==========================================
-- Memcached (缓存)
-- ==========================================
('agent-014-a3b4c5d6', 'aws-redis-01',      '10.0.3.20', 'linux', 'Memcached', '1.6.22', 11211, 'memcached', NOW() - INTERVAL '75 days', NOW()),
('agent-015-e7f8g9h0', 'aws-redis-02',      '10.0.3.21', 'linux', 'Memcached', '1.6.22', 11211, 'memcached', NOW() - INTERVAL '73 days', NOW()),
-- ==========================================
-- Consul (键值存储)
-- ==========================================
('agent-049-k3l4m5n6', 'aws-consul-01',     '10.0.3.72', 'linux', 'Consul', '1.17.1', 8500, 'consul', NOW() - INTERVAL '40 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'Consul', '1.17.1', 8500, 'consul', NOW() - INTERVAL '60 days', NOW()),
-- ==========================================
-- MariaDB
-- ==========================================
('agent-042-i5j6k7l8', 'aws-mail-01',       '10.0.7.14', 'linux', 'MariaDB', '10.11.6', 3306, 'mysql', NOW() - INTERVAL '90 days', NOW()),
('agent-043-m9n0o1p2', 'aws-ldap-01',       '10.0.7.15', 'linux', 'MariaDB', '10.6.16', 3306, 'mysql', NOW() - INTERVAL '85 days', NOW()),
-- ==========================================
-- RethinkDB
-- ==========================================
('agent-009-g3h4i5j6', 'aws-worker-01',     '10.0.2.20', 'linux', 'RethinkDB', '2.4.4', 28015, 'rethinkdb', NOW() - INTERVAL '70 days', NOW()),
-- ==========================================
-- Oracle
-- ==========================================
('agent-011-o1p2q3r4', 'aws-mysql-01',      '10.0.3.10', 'linux', 'Oracle', '19c', 1521, 'oracle', NOW() - INTERVAL '95 days', NOW());
