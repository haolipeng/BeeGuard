-- =====================================================
-- 模拟数据: alert_abnormal_login (异常登录告警表)
-- 数据量: 30条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- abnormal_type: abnormal_location/abnormal_time/abnormal_user
-- risk_level: low/medium/high
-- =====================================================

INSERT INTO alert_abnormal_login (agent_id, host_id, host_name, host_ip, source_ip, source_location, source_country, source_city, login_user, login_time, risk_level, abnormal_type, status, is_whitelist, created_at, updated_at) VALUES
-- 异常地域登录 (abnormal_location) - 11条
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', '45.33.32.156', '美国 加利福尼亚州 洛杉矶', '美国', '洛杉矶', 'root', NOW() - INTERVAL '30 minutes', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '30 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', '185.220.101.35', '德国 柏林', '德国', '柏林', 'admin', NOW() - INTERVAL '2 hours', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', '91.121.87.18', '法国 巴黎', '法国', '巴黎', 'ubuntu', NOW() - INTERVAL '5 hours', 'high', 'abnormal_location', 1, 0, NOW() - INTERVAL '5 hours', NOW()),
('agent-017-m5n6o7p8', 17, 'aws-es-02', '10.0.3.31', '103.25.61.114', '中国 北京', '中国', '北京', 'elastic', NOW() - INTERVAL '1 day', 'medium', 'abnormal_location', 1, 1, NOW() - INTERVAL '1 day', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', '45.155.205.233', '俄罗斯 莫斯科', '俄罗斯', '莫斯科', 'git', NOW() - INTERVAL '3 hours', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', '195.154.181.128', '法国 巴黎', '法国', '巴黎', 'ec2-user', NOW() - INTERVAL '6 hours', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '6 hours', NOW()),
('agent-035-g7h8i9j0', 37, 'aws-elk-01', '10.0.6.12', '91.240.118.172', '乌克兰 基辅', '乌克兰', '基辅', 'root', NOW() - INTERVAL '1 hour', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', '23.129.64.130', '美国 西雅图', '美国', '西雅图', 'ubuntu', NOW() - INTERVAL '4 hours', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-044-q3r4s5t6', 46, 'aws-proxy-01', '10.0.7.16', '103.74.192.18', '印度 新德里', '印度', '新德里', 'ec2-user', NOW() - INTERVAL '8 hours', 'medium', 'abnormal_location', 1, 0, NOW() - INTERVAL '8 hours', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', '185.156.73.54', '荷兰 阿姆斯特丹', '荷兰', '阿姆斯特丹', 'ftpuser', NOW() - INTERVAL '12 hours', 'high', 'abnormal_location', 0, 0, NOW() - INTERVAL '12 hours', NOW()),
('agent-049-k3l4m5n6', 49, 'aws-consul-01', '10.0.3.72', '45.227.255.99', '巴西 圣保罗', '巴西', '圣保罗', 'consul', NOW() - INTERVAL '2 days', 'high', 'abnormal_location', 2, 0, NOW() - INTERVAL '2 days', NOW()),

-- 异常时间登录 (abnormal_time) - 9条
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', '10.0.1.200', '内网', '新加坡', '本地', 'root', NOW() - INTERVAL '6 hours' + INTERVAL '3 hours', 'medium', 'abnormal_time', 0, 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', '10.0.3.200', '内网', '新加坡', '本地', 'mysql', NOW() - INTERVAL '1 day' + INTERVAL '2 hours', 'medium', 'abnormal_time', 0, 0, NOW() - INTERVAL '1 day', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', '10.0.1.201', '内网', '新加坡', '本地', 'deploy', NOW() - INTERVAL '2 days' + INTERVAL '4 hours', 'low', 'abnormal_time', 1, 1, NOW() - INTERVAL '2 days', NOW()),
('agent-033-y9z0a1b2', 35, 'aws-prometheus-01', '10.0.6.10', '10.0.6.200', '内网', '新加坡', '本地', 'prometheus', NOW() - INTERVAL '5 hours' + INTERVAL '1 hour', 'medium', 'abnormal_time', 0, 0, NOW() - INTERVAL '5 hours', NOW()),
('agent-020-y7z8a9b0', 20, 'aws-kafka-02', '10.0.3.41', '10.0.3.201', '内网', '新加坡', '本地', 'admin', NOW() - INTERVAL '8 hours' + INTERVAL '2 hours', 'medium', 'abnormal_time', 0, 0, NOW() - INTERVAL '8 hours', NOW()),
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', '10.0.6.201', '内网', '新加坡', '本地', 'ubuntu', NOW() - INTERVAL '10 hours' + INTERVAL '3 hours', 'low', 'abnormal_time', 1, 0, NOW() - INTERVAL '10 hours', NOW()),
('agent-041-e1f2g3h4', 43, 'aws-nfs-01', '10.0.7.13', '10.0.7.200', '内网', '新加坡', '本地', 'deploy', NOW() - INTERVAL '1 day 2 hours', 'medium', 'abnormal_time', 0, 0, NOW() - INTERVAL '1 day', NOW()),
('agent-043-m9n0o1p2', 45, 'aws-ldap-01', '10.0.7.15', '10.0.7.201', '内网', '新加坡', '本地', 'ldapadmin', NOW() - INTERVAL '3 days' + INTERVAL '5 hours', 'low', 'abnormal_time', 2, 1, NOW() - INTERVAL '3 days', NOW()),
('agent-050-o7p8q9r0', 50, 'aws-vault-01', '10.0.7.20', '10.0.7.202', '内网', '新加坡', '本地', 'ops', NOW() - INTERVAL '12 hours' + INTERVAL '4 hours', 'medium', 'abnormal_time', 0, 0, NOW() - INTERVAL '12 hours', NOW()),

-- 异常用户登录 (abnormal_user) - 10条
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', '10.0.1.150', '内网', '新加坡', '本地', 'test', NOW() - INTERVAL '45 minutes', 'high', 'abnormal_user', 0, 0, NOW() - INTERVAL '45 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', '10.0.1.151', '内网', '新加坡', '本地', 'nobody', NOW() - INTERVAL '4 hours', 'high', 'abnormal_user', 0, 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', '10.0.1.152', '内网', '新加坡', '本地', 'guest', NOW() - INTERVAL '7 hours', 'high', 'abnormal_user', 1, 0, NOW() - INTERVAL '7 hours', NOW()),
('agent-007-y5z6a7b8', 7, 'aws-app-02', '10.0.2.11', '10.0.2.150', '内网', '新加坡', '本地', 'default', NOW() - INTERVAL '2 hours', 'medium', 'abnormal_user', 0, 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', '10.0.5.150', '内网', '新加坡', '本地', 'backup', NOW() - INTERVAL '1 day', 'medium', 'abnormal_user', 1, 0, NOW() - INTERVAL '1 day', NOW()),
('agent-030-m7n8o9p0', 32, 'aws-harbor-01', '10.0.5.12', '10.0.5.151', '内网', '新加坡', '本地', 'sync', NOW() - INTERVAL '6 hours', 'low', 'abnormal_user', 2, 1, NOW() - INTERVAL '6 hours', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', '10.0.4.150', '内网', '新加坡', '本地', 'kubelet', NOW() - INTERVAL '3 hours', 'medium', 'abnormal_user', 0, 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-035-g7h8i9j0', 37, 'aws-elk-01', '10.0.6.12', '10.0.6.150', '内网', '新加坡', '本地', 'nobody', NOW() - INTERVAL '2 hours', 'high', 'abnormal_user', 0, 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', '10.0.7.150', '内网', '新加坡', '本地', 'mail', NOW() - INTERVAL '11 hours', 'medium', 'abnormal_user', 1, 0, NOW() - INTERVAL '11 hours', NOW()),
('agent-045-u7v8w9x0', 47, 'aws-backup-01', '10.0.7.17', '10.0.7.151', '内网', '新加坡', '本地', 'daemon', NOW() - INTERVAL '4 hours', 'high', 'abnormal_user', 0, 0, NOW() - INTERVAL '4 hours', NOW());
