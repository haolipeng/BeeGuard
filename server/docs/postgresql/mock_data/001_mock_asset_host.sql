-- =====================================================
-- 模拟数据: asset_host (主机资产表)
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
-- OS: Ubuntu 22.04/20.04, Amazon Linux 2/2023 (无 Windows)
-- =====================================================

INSERT INTO asset_host (id, agent_id, host_name, host_ip, mac_addr, os_type, os_version, agent_status, agent_version, last_heartbeat, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================
(1,  'agent-001-a1b2c3d4', 'aws-web-01',        '10.0.1.10', '02:a1:b2:c3:d4:01', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '90 days',  NOW()),
(2,  'agent-002-e5f6g7h8', 'aws-web-02',        '10.0.1.11', '02:e5:f6:g7:h8:02', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '88 days',  NOW()),
(3,  'agent-003-i9j0k1l2', 'aws-api-01',        '10.0.1.20', '02:i9:j0:k1:l2:03', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '85 days',  NOW()),
(4,  'agent-004-m3n4o5p6', 'aws-api-02',        '10.0.1.21', '02:m3:n4:o5:p6:04', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '85 days',  NOW()),
(5,  'agent-005-q7r8s9t0', 'aws-gateway-01',    '10.0.1.30', '02:q7:r8:s9:t0:05', 'linux', 'Amazon Linux 2023',  0, '2.1.2', NOW() - INTERVAL '2 days',     NOW() - INTERVAL '150 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================
(6,  'agent-006-u1v2w3x4', 'aws-app-01',        '10.0.2.10', '02:u1:v2:w3:x4:06', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '80 days',  NOW()),
(7,  'agent-007-y5z6a7b8', 'aws-app-02',        '10.0.2.11', '02:y5:z6:a7:b8:07', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '78 days',  NOW()),
(8,  'agent-008-c9d0e1f2', 'aws-app-03',        '10.0.2.12', '02:c9:d0:e1:f2:08', 'linux', 'Amazon Linux 2023',  1, '2.1.4', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '75 days',  NOW()),
(9,  'agent-009-g3h4i5j6', 'aws-worker-01',     '10.0.2.20', '02:g3:h4:i5:j6:09', 'linux', 'Ubuntu 20.04',       1, '2.1.3', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '70 days',  NOW()),
(10, 'agent-010-k7l8m9n0', 'aws-worker-02',     '10.0.2.21', '02:k7:l8:m9:n0:10', 'linux', 'Ubuntu 20.04',       1, '2.1.3', NOW() - INTERVAL '6 minutes',  NOW() - INTERVAL '68 days',  NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================
(11, 'agent-011-o1p2q3r4', 'aws-mysql-01',      '10.0.3.10', '02:o1:p2:q3:r4:11', 'linux', 'Amazon Linux 2',     1, '2.1.4', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '95 days',  NOW()),
(12, 'agent-012-s5t6u7v8', 'aws-mysql-02',      '10.0.3.11', '02:s5:t6:u7:v8:12', 'linux', 'Amazon Linux 2',     1, '2.1.4', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '93 days',  NOW()),
(13, 'agent-013-w9x0y1z2', 'aws-pg-01',         '10.0.3.12', '02:w9:x0:y1:z2:13', 'linux', 'Amazon Linux 2',     1, '2.1.5', NOW() - INTERVAL '1 minute',   NOW() - INTERVAL '90 days',  NOW()),
(14, 'agent-014-a3b4c5d6', 'aws-redis-01',      '10.0.3.20', '02:a3:b4:c5:d6:14', 'linux', 'Amazon Linux 2',     1, '2.1.5', NOW() - INTERVAL '1 minute',   NOW() - INTERVAL '75 days',  NOW()),
(15, 'agent-015-e7f8g9h0', 'aws-redis-02',      '10.0.3.21', '02:e7:f8:g9:h0:15', 'linux', 'Amazon Linux 2',     1, '2.1.4', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '73 days',  NOW()),
(16, 'agent-016-i1j2k3l4', 'aws-es-01',         '10.0.3.30', '02:i1:j2:k3:l4:16', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '65 days',  NOW()),
(17, 'agent-017-m5n6o7p8', 'aws-es-02',         '10.0.3.31', '02:m5:n6:o7:p8:17', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '63 days',  NOW()),
(18, 'agent-018-q9r0s1t2', 'aws-es-03',         '10.0.3.32', '02:q9:r0:s1:t2:18', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '60 days',  NOW()),
(19, 'agent-019-u3v4w5x6', 'aws-kafka-01',      '10.0.3.40', '02:u3:v4:w5:x6:19', 'linux', 'Amazon Linux 2',     1, '2.1.4', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '55 days',  NOW()),
(20, 'agent-020-y7z8a9b0', 'aws-kafka-02',      '10.0.3.41', '02:y7:z8:a9:b0:20', 'linux', 'Amazon Linux 2',     1, '2.1.4', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '53 days',  NOW()),
(21, 'agent-021-c1d2e3f4', 'aws-mq-01',         '10.0.3.50', '02:c1:d2:e3:f4:21', 'linux', 'Ubuntu 20.04',       1, '2.1.3', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '50 days',  NOW()),
(22, 'agent-022-g5h6i7j8', 'aws-mongo-01',      '10.0.3.60', '02:g5:h6:i7:j8:22', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '48 days',  NOW()),
(23, 'agent-047-c5d6e7f8', 'aws-zk-01',         '10.0.3.70', '02:c5:d6:e7:f8:23', 'linux', 'Amazon Linux 2',     1, '2.1.3', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '45 days',  NOW()),
(24, 'agent-048-g9h0i1j2', 'aws-zk-02',         '10.0.3.71', '02:g9:h0:i1:j2:24', 'linux', 'Amazon Linux 2',     1, '2.1.3', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '43 days',  NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================
(25, 'agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', '02:k9:l0:m1:n2:25', 'linux', 'Amazon Linux 2023',  1, '2.1.5', NOW() - INTERVAL '1 minute',   NOW() - INTERVAL '60 days',  NOW()),
(26, 'agent-024-o3p4q5r6', 'aws-eks-node-01',   '10.0.4.11', '02:o3:p4:q5:r6:26', 'linux', 'Amazon Linux 2023',  1, '2.1.5', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '58 days',  NOW()),
(27, 'agent-025-s7t8u9v0', 'aws-eks-node-02',   '10.0.4.12', '02:s7:t8:u9:v0:27', 'linux', 'Amazon Linux 2023',  1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '56 days',  NOW()),
(28, 'agent-026-w1x2y3z4', 'aws-eks-node-03',   '10.0.4.13', '02:w1:x2:y3:z4:28', 'linux', 'Amazon Linux 2023',  1, '2.1.4', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '54 days',  NOW()),
(29, 'agent-027-a5b6c7d8', 'aws-eks-node-04',   '10.0.4.14', '02:a5:b6:c7:d8:29', 'linux', 'Amazon Linux 2023',  1, '2.1.4', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '52 days',  NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================
(30, 'agent-028-e9f0g1h2', 'aws-jenkins-01',    '10.0.5.10', '02:e9:f0:g1:h2:30', 'linux', 'Ubuntu 22.04',       1, '2.1.3', NOW() - INTERVAL '6 minutes',  NOW() - INTERVAL '100 days', NOW()),
(31, 'agent-029-i3j4k5l6', 'aws-gitlab-01',     '10.0.5.11', '02:i3:j4:k5:l6:31', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '95 days',  NOW()),
(32, 'agent-030-m7n8o9p0', 'aws-harbor-01',     '10.0.5.12', '02:m7:n8:o9:p0:32', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '88 days',  NOW()),
(33, 'agent-031-q1r2s3t4', 'aws-nexus-01',      '10.0.5.13', '02:q1:r2:s3:t4:33', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '82 days',  NOW()),
(34, 'agent-032-u5v6w7x8', 'aws-sonar-01',      '10.0.5.14', '02:u5:v6:w7:x8:34', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '76 days',  NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================
(35, 'agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', '02:y9:z0:a1:b2:35', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '110 days', NOW()),
(36, 'agent-034-c3d4e5f6', 'aws-grafana-01',    '10.0.6.11', '02:c3:d4:e5:f6:36', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '105 days', NOW()),
(37, 'agent-035-g7h8i9j0', 'aws-elk-01',        '10.0.6.12', '02:g7:h8:i9:j0:37', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '70 days',  NOW()),
(38, 'agent-036-k1l2m3n4', 'aws-elk-02',        '10.0.6.13', '02:k1:l2:m3:n4:38', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '68 days',  NOW()),
(39, 'agent-037-o5p6q7r8', 'aws-alertmanager-01','10.0.6.14','02:o5:p6:q7:r8:39', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '65 days',  NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================
(40, 'agent-038-s9t0u1v2', 'aws-vpn-01',        '10.0.7.10', '02:s9:t0:u1:v2:40', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '120 days', NOW()),
(41, 'agent-039-w3x4y5z6', 'aws-bastion-01',    '10.0.7.11', '02:w3:x4:y5:z6:41', 'linux', 'Amazon Linux 2023',  1, '2.1.5', NOW() - INTERVAL '1 minute',   NOW() - INTERVAL '45 days',  NOW()),
(42, 'agent-040-a7b8c9d0', 'aws-dns-01',        '10.0.7.12', '02:a7:b8:c9:d0:42', 'linux', 'Amazon Linux 2',     1, '2.1.3', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '115 days', NOW()),
(43, 'agent-041-e1f2g3h4', 'aws-nfs-01',        '10.0.7.13', '02:e1:f2:g3:h4:43', 'linux', 'Ubuntu 20.04',       1, '2.1.3', NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '100 days', NOW()),
(44, 'agent-042-i5j6k7l8', 'aws-mail-01',       '10.0.7.14', '02:i5:j6:k7:l8:44', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '6 minutes',  NOW() - INTERVAL '90 days',  NOW()),
(45, 'agent-043-m9n0o1p2', 'aws-ldap-01',       '10.0.7.15', '02:m9:n0:o1:p2:45', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '85 days',  NOW()),
(46, 'agent-044-q3r4s5t6', 'aws-proxy-01',      '10.0.7.16', '02:q3:r4:s5:t6:46', 'linux', 'Amazon Linux 2023',  1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '80 days',  NOW()),
(47, 'agent-045-u7v8w9x0', 'aws-backup-01',     '10.0.7.17', '02:u7:v8:w9:x0:47', 'linux', 'Ubuntu 20.04',       1, '2.1.3', NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '100 days', NOW()),
(48, 'agent-046-y1z2a3b4', 'aws-ftp-01',        '10.0.7.18', '02:y1:z2:a3:b4:48', 'linux', 'Ubuntu 20.04',       0, '2.1.2', NOW() - INTERVAL '3 days',     NOW() - INTERVAL '130 days', NOW()),
(49, 'agent-049-k3l4m5n6', 'aws-consul-01',     '10.0.3.72', '02:k3:l4:m5:n6:49', 'linux', 'Ubuntu 22.04',       1, '2.1.4', NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '40 days',  NOW()),
(50, 'agent-050-o7p8q9r0', 'aws-vault-01',      '10.0.7.20', '02:o7:p8:q9:r0:50', 'linux', 'Ubuntu 22.04',       1, '2.1.5', NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '35 days',  NOW());

-- 重置序列
SELECT setval('asset_host_id_seq', 50);
