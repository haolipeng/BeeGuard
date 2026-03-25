-- =====================================================
-- 模拟数据: asset_kmod (内核模块资产表)
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
-- AWS 通用模块: ena (Elastic Network Adapter), nvme (NVMe 存储)
-- =====================================================

INSERT INTO asset_kmod (agent_id, host_name, host_ip, os_type, name, size, refcount, used_by, state, addr, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================
-- aws-web-01 内核模块 (Ubuntu 22.04)
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0a00000', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'nvme', '45056', '2', '', 'Live', '0xffffffffc0a10000', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,nf_conntrack_netlink,xt_conntrack,nft_ct', 'Live', '0xffffffffc0a20000', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'linux', 'ext4', '1015808', '1', '', 'Live', '0xffffffffc0a40000', NOW() - INTERVAL '90 days', NOW()),
-- aws-web-02 内核模块 (Ubuntu 22.04)
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0a50000', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'nvme', '45056', '2', '', 'Live', '0xffffffffc0a60000', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack,nft_ct', 'Live', '0xffffffffc0a70000', NOW() - INTERVAL '88 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'linux', 'ext4', '1015808', '1', '', 'Live', '0xffffffffc0a80000', NOW() - INTERVAL '88 days', NOW()),
-- aws-gateway-01 内核模块 (Amazon Linux 2023)
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0b30000', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'nf_conntrack', '172032', '8', 'nf_nat,ip_vs,xt_conntrack,nft_ct', 'Live', '0xffffffffc0b50000', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'ip_vs', '180224', '2', '', 'Live', '0xffffffffc0b60000', NOW() - INTERVAL '150 days', NOW()),
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'linux', 'ip_vs_rr', '16384', '1', '', 'Live', '0xffffffffc0b70000', NOW() - INTERVAL '150 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================
-- aws-app-01 内核模块 (Ubuntu 22.04)
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0c00000', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'nvme', '45056', '2', '', 'Live', '0xffffffffc0c10000', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack,nft_ct', 'Live', '0xffffffffc0c20000', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'linux', 'ext4', '1015808', '1', '', 'Live', '0xffffffffc0c30000', NOW() - INTERVAL '80 days', NOW()),
-- aws-worker-01 内核模块 (Ubuntu 20.04)
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0c60000', NOW() - INTERVAL '70 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0c80000', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================
-- aws-mysql-01 内核模块 (Amazon Linux 2)
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0d00000', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'nvme', '45056', '4', '', 'Live', '0xffffffffc0d10000', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'xfs', '1511424', '3', '', 'Live', '0xffffffffc0d20000', NOW() - INTERVAL '95 days', NOW()),
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'linux', 'nf_conntrack', '139264', '4', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0d30000', NOW() - INTERVAL '95 days', NOW()),
-- aws-pg-01 内核模块 (Amazon Linux 2)
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0d50000', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'xfs', '1511424', '2', '', 'Live', '0xffffffffc0d70000', NOW() - INTERVAL '90 days', NOW()),
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'linux', 'nf_conntrack', '139264', '3', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0d80000', NOW() - INTERVAL '90 days', NOW()),
-- aws-redis-01 内核模块 (Amazon Linux 2)
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0e00000', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0e10000', NOW() - INTERVAL '75 days', NOW()),
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'linux', 'tcp_bbr', '20480', '1', '', 'Live', '0xffffffffc0e20000', NOW() - INTERVAL '75 days', NOW()),
-- aws-kafka-01 内核模块 (Amazon Linux 2)
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0e70000', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'xfs', '1511424', '2', '', 'Live', '0xffffffffc0e90000', NOW() - INTERVAL '55 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0ea0000', NOW() - INTERVAL '55 days', NOW()),
-- aws-mongo-01 内核模块 (Ubuntu 22.04)
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0f30000', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'ext4', '1015808', '2', '', 'Live', '0xffffffffc0f50000', NOW() - INTERVAL '48 days', NOW()),
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'linux', 'nf_conntrack', '139264', '3', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0f60000', NOW() - INTERVAL '48 days', NOW()),
-- aws-zk-01 内核模块 (Amazon Linux 2)
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc0f70000', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'xfs', '1511424', '2', '', 'Live', '0xffffffffc0f80000', NOW() - INTERVAL '45 days', NOW()),
('agent-047-c5d6e7f8', 'aws-zk-01', '10.0.3.70', 'linux', 'nf_conntrack', '139264', '3', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc0f90000', NOW() - INTERVAL '45 days', NOW()),

-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================
-- aws-eks-master-01 内核模块 (Amazon Linux 2023)
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1000000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'nvme', '45056', '2', '', 'Live', '0xffffffffc1010000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'overlay', '155648', '52', '', 'Live', '0xffffffffc1020000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'br_netfilter', '28672', '0', '', 'Live', '0xffffffffc1030000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'ip_vs', '180224', '0', '', 'Live', '0xffffffffc1040000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'ip_vs_rr', '16384', '0', '', 'Live', '0xffffffffc1050000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'ip_vs_wrr', '16384', '0', '', 'Live', '0xffffffffc1060000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'ip_vs_sh', '16384', '0', '', 'Live', '0xffffffffc1070000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'vxlan', '90112', '2', '', 'Live', '0xffffffffc1080000', NOW() - INTERVAL '60 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'linux', 'nf_conntrack', '172032', '6', 'nf_nat,ip_vs,xt_conntrack', 'Live', '0xffffffffc1090000', NOW() - INTERVAL '60 days', NOW()),
-- aws-eks-node-01 内核模块 (Amazon Linux 2023)
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1100000', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'overlay', '155648', '128', '', 'Live', '0xffffffffc1110000', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'br_netfilter', '28672', '0', '', 'Live', '0xffffffffc1120000', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'ip_vs', '180224', '4', '', 'Live', '0xffffffffc1130000', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'vxlan', '90112', '2', '', 'Live', '0xffffffffc1140000', NOW() - INTERVAL '58 days', NOW()),
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'linux', 'nf_conntrack', '172032', '8', 'nf_nat,ip_vs,xt_conntrack,nft_ct', 'Live', '0xffffffffc1150000', NOW() - INTERVAL '58 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x) - Docker 主机
-- ==========================================
-- aws-jenkins-01 内核模块 (Ubuntu 22.04)
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1200000', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'overlay', '155648', '20', '', 'Live', '0xffffffffc1220000', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'br_netfilter', '28672', '0', '', 'Live', '0xffffffffc1230000', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'bridge', '311296', '1', 'br_netfilter', 'Live', '0xffffffffc1240000', NOW() - INTERVAL '100 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'linux', 'veth', '32768', '0', '', 'Live', '0xffffffffc1250000', NOW() - INTERVAL '100 days', NOW()),
-- aws-harbor-01 内核模块 (Ubuntu 22.04)
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1300000', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'overlay', '155648', '35', '', 'Live', '0xffffffffc1320000', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'br_netfilter', '28672', '0', '', 'Live', '0xffffffffc1330000', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'bridge', '311296', '1', 'br_netfilter', 'Live', '0xffffffffc1340000', NOW() - INTERVAL '88 days', NOW()),
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'linux', 'veth', '32768', '0', '', 'Live', '0xffffffffc1350000', NOW() - INTERVAL '88 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================
-- aws-prometheus-01 内核模块 (Ubuntu 22.04)
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1400000', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'nvme', '45056', '2', '', 'Live', '0xffffffffc1410000', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc1420000', NOW() - INTERVAL '110 days', NOW()),
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'linux', 'ext4', '1015808', '2', '', 'Live', '0xffffffffc1430000', NOW() - INTERVAL '110 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================
-- aws-vpn-01 内核模块 (Ubuntu 22.04)
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1500000', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'tun', '57344', '2', '', 'Live', '0xffffffffc1510000', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'nf_conntrack', '172032', '6', 'nf_nat,xt_conntrack,xt_state', 'Live', '0xffffffffc1520000', NOW() - INTERVAL '120 days', NOW()),
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'linux', 'iptable_nat', '16384', '1', '', 'Live', '0xffffffffc1530000', NOW() - INTERVAL '120 days', NOW()),
-- aws-nfs-01 内核模块 (Ubuntu 20.04)
('agent-041-e1f2g3h4', 'aws-nfs-01', '10.0.7.13', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1600000', NOW() - INTERVAL '100 days', NOW()),
('agent-041-e1f2g3h4', 'aws-nfs-01', '10.0.7.13', 'linux', 'nfsd', '577536', '13', '', 'Live', '0xffffffffc1620000', NOW() - INTERVAL '100 days', NOW()),
('agent-041-e1f2g3h4', 'aws-nfs-01', '10.0.7.13', 'linux', 'nfs', '393216', '0', '', 'Live', '0xffffffffc1630000', NOW() - INTERVAL '100 days', NOW()),
('agent-041-e1f2g3h4', 'aws-nfs-01', '10.0.7.13', 'linux', 'sunrpc', '577536', '26', 'nfsd,nfs,nfs_acl,lockd', 'Live', '0xffffffffc1650000', NOW() - INTERVAL '100 days', NOW()),
-- aws-backup-01 内核模块 (Ubuntu 20.04)
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'ena', '114688', '0', '', 'Live', '0xffffffffc1800000', NOW() - INTERVAL '100 days', NOW()),
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'nvme', '45056', '4', '', 'Live', '0xffffffffc1810000', NOW() - INTERVAL '100 days', NOW()),
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'ext4', '1015808', '4', '', 'Live', '0xffffffffc1820000', NOW() - INTERVAL '100 days', NOW()),
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'fuse', '163840', '5', '', 'Live', '0xffffffffc1830000', NOW() - INTERVAL '100 days', NOW()),
('agent-045-u7v8w9x0', 'aws-backup-01', '10.0.7.17', 'linux', 'nf_conntrack', '172032', '4', 'nf_nat,xt_conntrack', 'Live', '0xffffffffc1840000', NOW() - INTERVAL '100 days', NOW());
