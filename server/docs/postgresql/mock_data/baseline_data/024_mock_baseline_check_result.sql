-- =====================================================
-- 模拟数据: baseline_check_result (检查结果表)
-- 数据量: 20条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的Linux主机生成基线检查结果
-- 每条记录表示一次针对某主机执行某基线模板的检查汇总
-- =====================================================

INSERT INTO baseline_check_result (id, baseline_id, agent_id, host_ip, host_name, total_items, passed_items, failed_items, check_time, created_at, updated_at) VALUES

-- CentOS 7 安全基线 (baseline_id=1, 15项) -> Amazon Linux 2 主机
(1,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01',      15, 12, 3,  NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '2 hours',  NOW()),
(2,  1, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02',      15, 10, 5,  NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '2 hours',  NOW()),
(3,  1, 'agent-013-w9x0y1z2', '10.0.3.12', 'aws-pg-01',         15, 14, 1,  NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '2 hours',  NOW()),
(4,  1, 'agent-019-u3v4w5x6', '10.0.3.40', 'aws-kafka-01',      15, 11, 4,  NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '2 hours',  NOW()),
(5,  1, 'agent-014-a3b4c5d6', '10.0.3.20', 'aws-redis-01',      15, 13, 2,  NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '2 hours',  NOW()),
(6,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02',      15,  9, 6,  NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '2 hours',  NOW()),

-- Ubuntu 22.04 安全基线 (baseline_id=2, 14项) -> Ubuntu 22.04 主机
(7,  2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01',        14, 13, 1,  NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '3 hours',  NOW()),
(8,  2, 'agent-002-e5f6g7h8', '10.0.1.11', 'aws-web-02',        14, 11, 3,  NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '3 hours',  NOW()),
(9,  2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01',         14, 14, 0,  NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '3 hours',  NOW()),
(10, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01',    14, 10, 4,  NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '3 hours',  NOW()),
(11, 2, 'agent-029-i3j4k5l6', '10.0.5.11', 'aws-gitlab-01',     14, 12, 2,  NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '3 hours',  NOW()),
(12, 2, 'agent-006-u1v2w3x4', '10.0.2.10', 'aws-app-01',        14, 11, 3,  NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '3 hours',  NOW()),

-- Debian 11 安全基线 (baseline_id=3, 12项) -> Ubuntu 20.04 / Amazon Linux 主机
(13, 3, 'agent-009-g3h4i5j6', '10.0.2.20', 'aws-worker-01',     12, 10, 2,  NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '4 hours',  NOW()),
(14, 3, 'agent-010-k7l8m9n0', '10.0.2.21', 'aws-worker-02',     12, 11, 1,  NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '4 hours',  NOW()),
(15, 3, 'agent-021-c1d2e3f4', '10.0.3.50', 'aws-mq-01',         12,  8, 4,  NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '4 hours',  NOW()),

-- CentOS 8 安全基线 (baseline_id=4, 13项) -> Amazon Linux 2023 主机
(16, 4, 'agent-025-s7t8u9v0', '10.0.4.12', 'aws-eks-node-02',   13, 11, 2,  NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '5 hours',  NOW()),
(17, 4, 'agent-044-q3r4s5t6', '10.0.7.16', 'aws-proxy-01',      13, 12, 1,  NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '5 hours',  NOW()),

-- Ubuntu 20.04 安全基线 (baseline_id=5, 12项) -> Ubuntu 20.04 主机
(18, 5, 'agent-041-e1f2g3h4', '10.0.7.13', 'aws-nfs-01',        12, 10, 2,  NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '6 hours',  NOW()),

-- MySQL 安全基线 (baseline_id=6, 10项) -> Amazon Linux 2 数据库主机
(19, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01',      10,  7, 3,  NOW() - INTERVAL '1 hour',   NOW() - INTERVAL '1 hour',   NOW()),
(20, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02',      10,  6, 4,  NOW() - INTERVAL '1 hour',   NOW() - INTERVAL '1 hour',   NOW());

-- 重置序列
SELECT setval('baseline_check_result_id_seq', 20);
