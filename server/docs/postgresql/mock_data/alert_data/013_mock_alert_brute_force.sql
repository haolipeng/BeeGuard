-- =====================================================
-- 模拟数据: alert_brute_force (暴力破解告警表)
-- 数据量: 35条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的主机生成暴力破解告警数据
-- attack_type: ssh/ftp/mysql/redis/web_login
-- OS: 全部为 Linux (Ubuntu 22.04/20.04, Amazon Linux 2/2023)
-- =====================================================

INSERT INTO alert_brute_force (agent_id, host_id, host_name, host_ip, source_ip, source_location, attack_type, target_ip, target_port, username, attempt_count, attack_time, first_attack_time, status, is_blocked, process_time, processor, remark, created_at, updated_at) VALUES

-- ==========================================
-- SSH 暴力破解告警 (13条)
-- ==========================================
-- Web/API 层
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', '45.33.32.156', '美国 加利福尼亚州', 'ssh', '10.0.1.10', 22, 'root', 156, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '3 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '2 hours', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', '185.220.101.35', '德国 柏林', 'ssh', '10.0.1.11', 22, 'admin', 89, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '6 hours', 1, 1, NOW() - INTERVAL '4 hours', 'security_admin', '已封禁攻击IP', NOW() - INTERVAL '5 hours', NOW()),
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', '103.25.61.114', '中国 北京', 'ssh', '10.0.1.20', 22, 'root', 245, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day 2 hours', 1, 1, NOW() - INTERVAL '23 hours', 'ops_admin', '内网测试触发,已处理', NOW() - INTERVAL '1 day', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', '91.121.87.18', '法国 巴黎', 'ssh', '10.0.1.30', 22, 'ubuntu', 67, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '9 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '8 hours', NOW()),
-- EKS/K8s 层
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', '45.155.205.233', '俄罗斯 莫斯科', 'ssh', '10.0.4.10', 22, 'root', 312, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '2 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '30 minutes', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', '195.154.181.128', '法国 巴黎', 'ssh', '10.0.4.11', 22, 'kubernetes', 45, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '13 hours', 2, 0, NOW() - INTERVAL '11 hours', 'security_admin', '误报,扫描工具测试', NOW() - INTERVAL '12 hours', NOW()),
-- DevOps 层
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', '61.177.173.25', '中国 江苏', 'ssh', '10.0.5.10', 22, 'jenkins', 128, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '5 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '4 hours', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', '185.156.73.54', '荷兰 阿姆斯特丹', 'ssh', '10.0.5.11', 22, 'git', 78, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '7 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '6 hours', NOW()),
-- 基础设施/安全层
('agent-039-w3x4y5z6', 41, 'aws-bastion-01', '10.0.7.11', '91.240.118.172', '乌克兰 基辅', 'ssh', '10.0.7.11', 22, 'root', 234, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '3 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '1 hour', NOW()),
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', '45.143.220.115', '美国 纽约', 'ssh', '10.0.7.10', 22, 'ubuntu', 189, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '5 hours', 1, 1, NOW() - INTERVAL '2 hours', 'cloud_admin', '已添加安全组规则', NOW() - INTERVAL '3 hours', NOW()),
-- 应用层
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', '185.220.100.252', '德国 法兰克福', 'ssh', '10.0.2.10', 22, 'ec2-user', 156, NOW() - INTERVAL '10 hours', NOW() - INTERVAL '12 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '10 hours', NOW()),
('agent-007-y5z6a7b8', 7, 'aws-app-02', '10.0.2.11', '23.129.64.130', '美国 西雅图', 'ssh', '10.0.2.11', 22, 'root', 312, NOW() - INTERVAL '45 minutes', NOW() - INTERVAL '2 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '45 minutes', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', '103.75.190.11', '印度 孟买', 'ssh', '10.0.7.14', 22, 'deploy', 98, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days 1 hour', 1, 1, NOW() - INTERVAL '1 day 22 hours', 'security_admin', '已封禁并上报', NOW() - INTERVAL '2 days', NOW()),

-- ==========================================
-- FTP 暴力破解告警 (3条)
-- ==========================================
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', '222.186.30.112', '中国 上海', 'ftp', '10.0.7.18', 21, 'ftpuser', 456, NOW() - INTERVAL '20 minutes', NOW() - INTERVAL '1 hour', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '20 minutes', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', '119.45.227.38', '中国 广东', 'ftp', '10.0.7.18', 21, 'anonymous', 234, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '6 hours', 1, 0, NOW() - INTERVAL '4 hours', 'ops_admin', '匿名登录已禁用', NOW() - INTERVAL '5 hours', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', '45.227.255.99', '巴西 圣保罗', 'ftp', '10.0.7.18', 21, 'root', 178, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day 1 hour', 2, 0, NOW() - INTERVAL '23 hours', 'security_admin', '误报-内部扫描', NOW() - INTERVAL '1 day', NOW()),

-- ==========================================
-- MySQL 暴力破解告警 (4条)
-- ==========================================
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', '58.218.198.160', '中国 江苏', 'mysql', '10.0.3.10', 3306, 'root', 567, NOW() - INTERVAL '15 minutes', NOW() - INTERVAL '45 minutes', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '15 minutes', NOW()),
('agent-012-s5t6u7v8', 12, 'aws-mysql-02', '10.0.3.11', '185.161.248.12', '俄罗斯 莫斯科', 'mysql', '10.0.3.11', 3306, 'mysql', 234, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '5 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '4 hours', NOW()),
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', '103.153.78.45', '越南 河内', 'mysql', '10.0.3.10', 3306, 'admin', 145, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '9 hours', 1, 1, NOW() - INTERVAL '7 hours', 'dba_admin', '已封禁IP段', NOW() - INTERVAL '8 hours', NOW()),
('agent-013-w9x0y1z2', 13, 'aws-pg-01', '10.0.3.12', '45.33.32.156', '美国 加利福尼亚州', 'mysql', '10.0.3.12', 3306, 'root', 89, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day 2 hours', 1, 0, NOW() - INTERVAL '22 hours', 'ops_admin', '服务未开放,误报', NOW() - INTERVAL '1 day', NOW()),

-- ==========================================
-- Redis 未授权访问/暴力破解告警 (3条)
-- ==========================================
('agent-014-a3b4c5d6', 14, 'aws-redis-01', '10.0.3.20', '103.74.192.18', '印度 新德里', 'redis', '10.0.3.20', 6379, 'default', 345, NOW() - INTERVAL '10 minutes', NOW() - INTERVAL '30 minutes', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '10 minutes', NOW()),
('agent-015-e7f8g9h0', 15, 'aws-redis-02', '10.0.3.21', '45.155.205.233', '俄罗斯 莫斯科', 'redis', '10.0.3.21', 6379, 'default', 189, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '7 hours', 1, 1, NOW() - INTERVAL '5 hours', 'security_admin', '已配置认证', NOW() - INTERVAL '6 hours', NOW()),
('agent-014-a3b4c5d6', 14, 'aws-redis-01', '10.0.3.20', '185.220.101.35', '德国 柏林', 'redis', '10.0.3.20', 6379, 'default', 267, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days 1 hour', 1, 1, NOW() - INTERVAL '1 day 20 hours', 'ops_admin', '已修复并封禁', NOW() - INTERVAL '2 days', NOW()),

-- ==========================================
-- Web 登录暴力破解告警 (12条)
-- ==========================================
-- Web/API 层
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', '45.143.220.115', '美国 纽约', 'web_login', '10.0.1.10', 80, 'admin', 234, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '2 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '1 hour', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', '91.240.118.172', '乌克兰 基辅', 'web_login', '10.0.1.11', 443, 'administrator', 156, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '4 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '3 hours', NOW()),
('agent-004-m3n4o5p6', 4, 'aws-api-02', '10.0.1.21', '103.25.61.114', '中国 北京', 'web_login', '10.0.1.21', 443, 'root', 198, NOW() - INTERVAL '7 hours', NOW() - INTERVAL '8 hours', 1, 1, NOW() - INTERVAL '6 hours', 'security_admin', '已启用WAF规则', NOW() - INTERVAL '7 hours', NOW()),
-- DevOps 层
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', '185.156.73.54', '荷兰 阿姆斯特丹', 'web_login', '10.0.5.10', 8080, 'admin', 567, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '1 hour', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '30 minutes', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', '103.25.61.114', '中国 北京', 'web_login', '10.0.5.11', 443, 'root', 345, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '6 hours', 1, 1, NOW() - INTERVAL '4 hours', 'security_admin', '已启用验证码', NOW() - INTERVAL '5 hours', NOW()),
('agent-030-m7n8o9p0', 32, 'aws-harbor-01', '10.0.5.12', '61.177.173.25', '中国 江苏', 'web_login', '10.0.5.12', 443, 'admin', 123, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '9 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '8 hours', NOW()),
-- 监控层
('agent-033-y9z0a1b2', 35, 'aws-prometheus-01', '10.0.6.10', '222.186.30.112', '中国 上海', 'web_login', '10.0.6.10', 9090, 'admin', 89, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days 1 hour', 2, 0, NOW() - INTERVAL '1 day 22 hours', 'ops_admin', '内部测试', NOW() - INTERVAL '2 days', NOW()),
('agent-034-c3d4e5f6', 36, 'aws-grafana-01', '10.0.6.11', '103.153.78.45', '越南 河内', 'web_login', '10.0.6.11', 3000, 'grafana', 167, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '13 hours', 1, 0, NOW() - INTERVAL '11 hours', 'ops_admin', '已加强密码策略', NOW() - INTERVAL '12 hours', NOW()),
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', '45.227.255.99', '巴西 圣保罗', 'web_login', '10.0.6.14', 9093, 'admin', 134, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '5 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '4 hours', NOW()),
-- 基础设施层
('agent-049-k3l4m5n6', 49, 'aws-consul-01', '10.0.3.72', '58.218.198.160', '中国 江苏', 'web_login', '10.0.3.72', 8500, 'consul', 234, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '5 hours', 0, 0, NULL, NULL, NULL, NOW() - INTERVAL '4 hours', NOW()),
('agent-050-o7p8q9r0', 50, 'aws-vault-01', '10.0.7.20', '185.220.100.252', '德国 法兰克福', 'web_login', '10.0.7.20', 8200, 'admin', 312, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '7 hours', 1, 1, NOW() - INTERVAL '5 hours', 'security_admin', '已限制访问IP白名单', NOW() - INTERVAL '6 hours', NOW()),
('agent-043-m9n0o1p2', 45, 'aws-ldap-01', '10.0.7.15', '91.121.87.18', '法国 巴黎', 'web_login', '10.0.7.15', 443, 'admin', 278, NOW() - INTERVAL '9 hours', NOW() - INTERVAL '10 hours', 1, 0, NOW() - INTERVAL '8 hours', 'ops_admin', '已加强LDAP认证策略', NOW() - INTERVAL '9 hours', NOW());
