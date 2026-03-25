-- =====================================================
-- 模拟数据: agent_info (Agent客户端信息表)
-- 数据量: 10条
-- 说明: AWS ap-southeast-1 区域 EC2 实例上的 HCIDS Agent
-- VPC CIDR: 10.0.0.0/16
-- =====================================================

INSERT INTO agent_info (agent_id, agent_version, connection_status, host_name, host_ip, os_type, os_version, os_arch, cpu_count, memory_total, disk_total, last_connected_at, registered_at, created_at, updated_at) VALUES
('agent-001-a1b2c3d4', '2.1.5', 1, 'aws-web-01',        '10.0.1.10', 'linux', 'Ubuntu 22.04.4 LTS',    'x86_64',  4,  8589934592,   107374182400, NOW() - INTERVAL '5 minutes',  NOW() - INTERVAL '90 days',  NOW() - INTERVAL '90 days',  NOW()),
('agent-006-u1v2w3x4', '2.1.5', 1, 'aws-app-01',        '10.0.2.10', 'linux', 'Ubuntu 22.04.4 LTS',    'x86_64',  8,  34359738368,  214748364800, NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '85 days',  NOW() - INTERVAL '85 days',  NOW()),
('agent-011-o1p2q3r4', '2.1.4', 1, 'aws-mysql-01',      '10.0.3.10', 'linux', 'Amazon Linux 2',        'x86_64',  8,  34359738368,  536870912000, NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '80 days',  NOW() - INTERVAL '80 days',  NOW()),
('agent-014-a3b4c5d6', '2.1.5', 1, 'aws-redis-01',      '10.0.3.20', 'linux', 'Amazon Linux 2',        'x86_64',  4,  17179869184,  107374182400, NOW() - INTERVAL '1 minute',   NOW() - INTERVAL '75 days',  NOW() - INTERVAL '75 days',  NOW()),
('agent-024-o3p4q5r6', '2.1.5', 1, 'aws-eks-node-01',   '10.0.4.11', 'linux', 'Amazon Linux 2023',     'x86_64',  8,  34359738368,  214748364800, NOW() - INTERVAL '4 minutes',  NOW() - INTERVAL '60 days',  NOW() - INTERVAL '60 days',  NOW()),
('agent-028-e9f0g1h2', '2.1.3', 1, 'aws-jenkins-01',    '10.0.5.10', 'linux', 'Ubuntu 22.04.4 LTS',    'x86_64',  4,  17179869184,  214748364800, NOW() - INTERVAL '6 minutes',  NOW() - INTERVAL '100 days', NOW() - INTERVAL '100 days', NOW()),
('agent-035-g7h8i9j0', '2.1.5', 1, 'aws-elk-01',        '10.0.6.12', 'linux', 'Ubuntu 22.04.4 LTS',    'x86_64',  8,  34359738368,  536870912000, NOW() - INTERVAL '2 minutes',  NOW() - INTERVAL '70 days',  NOW() - INTERVAL '70 days',  NOW()),
('agent-033-y9z0a1b2', '2.1.4', 1, 'aws-prometheus-01',  '10.0.6.10', 'linux', 'Ubuntu 22.04.4 LTS',    'x86_64',  4,  8589934592,   214748364800, NOW() - INTERVAL '3 minutes',  NOW() - INTERVAL '110 days', NOW() - INTERVAL '110 days', NOW()),
('agent-039-w3x4y5z6', '2.1.5', 1, 'aws-bastion-01',    '10.0.7.11', 'linux', 'Amazon Linux 2023',     'x86_64',  2,  4294967296,   53687091200,  NOW() - INTERVAL '1 minute',   NOW() - INTERVAL '45 days',  NOW() - INTERVAL '45 days',  NOW()),
('agent-005-q7r8s9t0', '2.1.2', 0, 'aws-gateway-01',    '10.0.1.30', 'linux', 'Amazon Linux 2023',     'x86_64',  2,  4294967296,   107374182400, NOW() - INTERVAL '2 days',     NOW() - INTERVAL '150 days', NOW() - INTERVAL '150 days', NOW());
