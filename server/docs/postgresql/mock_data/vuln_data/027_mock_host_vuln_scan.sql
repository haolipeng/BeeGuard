-- =====================================================
-- 模拟数据: host_vuln_scan_task (主机漏洞扫描任务表)
-- 数据量: 30条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
--       VPC CIDR: 10.0.0.0/16
--       引用 001_mock_asset_host.sql 中的主机数据
--       scan_status: 1-成功
--       scan_trigger: auto-自动扫描
--       matched_vulns 与 host_vuln_detail 数据匹配
-- =====================================================

INSERT INTO host_vuln_scan_task (id, agent_id, host_id, host_name, host_ip, scan_status, scan_trigger, total_packages, matched_vulns, scan_duration, error_message, scan_time, created_at, updated_at) VALUES
-- Web/API 层 (10.0.1.x)
(1,  'agent-001-a1b2c3d4', 1,  'aws-web-01',        '10.0.1.10',  1, 'auto', 187, 7, 3520, NULL, NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(2,  'agent-002-e5f6g7h8', 2,  'aws-web-02',        '10.0.1.11',  1, 'auto', 195, 6, 2850, NULL, NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),
(3,  'agent-003-i9j0k1l2', 3,  'aws-api-01',        '10.0.1.20',  1, 'auto', 156, 6, 4210, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(4,  'agent-004-m3n4o5p6', 4,  'aws-api-02',        '10.0.1.21',  1, 'auto', 162, 5, 3780, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),
(5,  'agent-005-q7r8s9t0', 5,  'aws-gateway-01',    '10.0.1.30',  1, 'auto', 210, 4, 5120, NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '365 days', NOW()),
-- 应用层 (10.0.2.x)
(6,  'agent-006-u1v2w3x4', 6,  'aws-app-01',        '10.0.2.10',  1, 'auto', 178, 5, 2960, NULL, NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(7,  'agent-007-y5z6a7b8', 7,  'aws-app-02',        '10.0.2.11',  1, 'auto', 183, 5, 3150, NULL, NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(8,  'agent-009-g3h4i5j6', 9,  'aws-worker-01',     '10.0.2.20',  1, 'auto', 145, 4, 2380, NULL, NOW() - INTERVAL '9 hours',  NOW() - INTERVAL '120 days', NOW()),
-- 数据层 (10.0.3.x)
(9,  'agent-011-o1p2q3r4', 11, 'aws-mysql-01',      '10.0.3.10',  1, 'auto', 134, 6, 4530, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(10, 'agent-012-s5t6u7v8', 12, 'aws-mysql-02',      '10.0.3.11',  1, 'auto', 138, 5, 4150, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),
(11, 'agent-014-a3b4c5d6', 14, 'aws-redis-01',      '10.0.3.20',  1, 'auto', 121, 5, 1980, NULL, NOW() - INTERVAL '10 hours', NOW() - INTERVAL '90 days', NOW()),
(12, 'agent-015-e7f8g9h0', 15, 'aws-redis-02',      '10.0.3.21',  1, 'auto', 126, 5, 2140, NULL, NOW() - INTERVAL '2 days',   NOW() - INTERVAL '88 days', NOW()),
(13, 'agent-016-i1j2k3l4', 16, 'aws-es-01',         '10.0.3.30',  1, 'auto', 245, 4, 5680, NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '200 days', NOW()),
(14, 'agent-017-m5n6o7p8', 17, 'aws-es-02',         '10.0.3.31',  1, 'auto', 238, 4, 5430, NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '200 days', NOW()),
(15, 'agent-019-u3v4w5x6', 19, 'aws-kafka-01',      '10.0.3.40',  1, 'auto', 198, 3, 3860, NULL, NOW() - INTERVAL '11 hours', NOW() - INTERVAL '250 days', NOW()),
(16, 'agent-021-c1d2e3f4', 21, 'aws-mq-01',         '10.0.3.50',  1, 'auto', 175, 4, 3240, NULL, NOW() - INTERVAL '9 hours',  NOW() - INTERVAL '120 days', NOW()),
(17, 'agent-022-g5h6i7j8', 22, 'aws-mongo-01',      '10.0.3.60',  1, 'auto', 168, 4, 3070, NULL, NOW() - INTERVAL '9 hours',  NOW() - INTERVAL '118 days', NOW()),
-- EKS/K8s 层 (10.0.4.x)
(18, 'agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10',  1, 'auto', 267, 6, 5950, NULL, NOW() - INTERVAL '2 hours',  NOW() - INTERVAL '60 days', NOW()),
(19, 'agent-024-o3p4q5r6', 26, 'aws-eks-node-01',   '10.0.4.11',  1, 'auto', 258, 5, 5620, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(20, 'agent-025-s7t8u9v0', 27, 'aws-eks-node-02',   '10.0.4.12',  1, 'auto', 253, 5, 5340, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(21, 'agent-026-w1x2y3z4', 28, 'aws-eks-node-03',   '10.0.4.13',  1, 'auto', 246, 5, 5180, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
-- DevOps 层 (10.0.5.x)
(22, 'agent-028-e9f0g1h2', 30, 'aws-jenkins-01',    '10.0.5.10',  1, 'auto', 219, 6, 4780, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '100 days', NOW()),
(23, 'agent-029-i3j4k5l6', 31, 'aws-gitlab-01',     '10.0.5.11',  1, 'auto', 225, 4, 4920, NULL, NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '95 days', NOW()),
-- 监控层 (10.0.6.x)
(24, 'agent-033-y9z0a1b2', 35, 'aws-prometheus-01', '10.0.6.10',  1, 'auto', 142, 4, 2560, NULL, NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '180 days', NOW()),
(25, 'agent-035-g7h8i9j0', 37, 'aws-elk-01',        '10.0.6.12',  1, 'auto', 276, 4, 5870, NULL, NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '200 days', NOW()),
-- 基础设施/安全层 (10.0.7.x)
(26, 'agent-038-s9t0u1v2', 40, 'aws-vpn-01',        '10.0.7.10',  1, 'auto', 131, 5, 1890, NULL, NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '365 days', NOW()),
(27, 'agent-040-a7b8c9d0', 42, 'aws-dns-01',        '10.0.7.12',  1, 'auto', 128, 6, 1750, NULL, NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '300 days', NOW()),
(28, 'agent-041-e1f2g3h4', 43, 'aws-nfs-01',        '10.0.7.13',  1, 'auto', 152, 3, 2670, NULL, NOW() - INTERVAL '11 hours', NOW() - INTERVAL '250 days', NOW()),
(29, 'agent-042-i5j6k7l8', 44, 'aws-mail-01',       '10.0.7.14',  1, 'auto', 189, 5, 3410, NULL, NOW() - INTERVAL '9 hours',  NOW() - INTERVAL '250 days', NOW()),
(30, 'agent-043-m9n0o1p2', 45, 'aws-ldap-01',       '10.0.7.15',  1, 'auto', 147, 3, 2480, NULL, NOW() - INTERVAL '11 hours', NOW() - INTERVAL '180 days', NOW());

-- 重置序列
SELECT setval('host_vuln_scan_task_id_seq', 30);
