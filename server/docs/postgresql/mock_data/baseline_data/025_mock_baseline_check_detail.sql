-- =====================================================
-- 模拟数据: baseline_check_detail (检查明细表)
-- 数据量: 96条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 baseline_check_result 生成的逐项检查明细
-- status: 1-通过 0-未通过 2-检查异常
-- =====================================================

INSERT INTO baseline_check_detail (id, result_id, item_id, baseline_id, agent_id, host_ip, host_name, template_name, template_id, status, actual_value, expected_value, error_message, check_time, created_at, updated_at) VALUES

-- ==========================================
-- result_id=1: aws-mysql-01, Amazon Linux 2基线(baseline_id=1), 15项(12通过/3未通过)
-- ==========================================
(1,  1, 1,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '8',       '8',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(2,  1, 2,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '90',      '90',      NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(3,  1, 3,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 0, '0',       '1',       '未配置密码复杂度策略pam_pwquality', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(4,  1, 4,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 0, 'yes',     'no',      'PermitRootLogin当前值为yes', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(5,  1, 5,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '1',       '1',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(6,  1, 6,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '300',     '300',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(7,  1, 7,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '4',       '4',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(8,  1, 8,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 0, 'inactive','active',  'firewalld服务未启动', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(9,  1, 9,  1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '644',     '644',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(10, 1, 10, 1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '000',     '000',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(11, 1, 11, 1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '000',     '000',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(12, 1, 12, 1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, 'active',  'active',  NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(13, 1, 13, 1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, 'active',  'active',  NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(14, 1, 14, 1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '0',       '0',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(15, 1, 15, 1, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'Amazon Linux 2 安全基线', 1, 1, '0',       '0',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),

-- ==========================================
-- result_id=7: aws-web-01, Ubuntu 22.04基线(baseline_id=2), 14项(13通过/1未通过)
-- ==========================================
(16, 7, 16, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '8',       '8',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(17, 7, 17, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '90',      '90',      NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(18, 7, 18, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, 'no',      'no',      NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(19, 7, 19, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '300',     '300',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(20, 7, 20, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(21, 7, 21, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '644',     '644',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(22, 7, 22, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '640',     '640',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(23, 7, 23, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, 'active',  'active',  NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(24, 7, 24, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '0',       '0',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(25, 7, 25, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '0',       '0',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(26, 7, 26, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(27, 7, 27, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 0, '0',       '1',       '未配置core dump限制', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(28, 7, 28, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, '4',       '4',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(29, 7, 29, 2, 'agent-001-a1b2c3d4', '10.0.1.10', 'aws-web-01', 'Ubuntu 22.04 安全基线', 2, 1, 'active',  'active',  NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- ==========================================
-- result_id=10: aws-jenkins-01, Ubuntu 22.04基线(baseline_id=2), 14项(10通过/4未通过)
-- ==========================================
(30, 10, 16, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '8',       '8',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(31, 10, 17, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 0, '99999',   '90',      '密码过期时间为99999天(未限制)', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(32, 10, 18, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 0, 'yes',     'no',      'PermitRootLogin当前值为yes', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(33, 10, 19, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '300',     '300',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(34, 10, 20, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 0, '0',       '1',       'UFW防火墙未启用', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(35, 10, 21, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '644',     '644',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(36, 10, 22, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '640',     '640',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(37, 10, 23, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 0, 'inactive','active',  'auditd服务未安装', NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(38, 10, 24, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '0',       '0',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(39, 10, 25, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '0',       '0',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(40, 10, 26, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(41, 10, 27, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(42, 10, 28, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, '4',       '4',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(43, 10, 29, 2, 'agent-028-e9f0g1h2', '10.0.5.10', 'aws-jenkins-01', 'Ubuntu 22.04 安全基线', 2, 1, 'active',  'active',  NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- ==========================================
-- result_id=6: aws-redis-02, Amazon Linux 2基线(baseline_id=1), 15项(9通过/6未通过)
-- ==========================================
(44, 6, 1,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 0, '5',       '8',       '密码最小长度仅为5位', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(45, 6, 2,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 0, '99999',   '90',      '密码过期时间未限制', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(46, 6, 3,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 0, '0',       '1',       '未配置密码复杂度', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(47, 6, 4,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 0, 'yes',     'no',      'root允许SSH远程登录', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(48, 6, 5,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, '1',       '1',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(49, 6, 6,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 0, '',        '300',     '未配置ClientAliveInterval', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(50, 6, 7,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 0, '6',       '4',       'MaxAuthTries设置过大', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(51, 6, 8,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, 'active',  'active',  NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(52, 6, 9,  1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, '644',     '644',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(53, 6, 10, 1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, '000',     '000',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(54, 6, 11, 1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, '000',     '000',     NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(55, 6, 12, 1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, 'active',  'active',  NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(56, 6, 13, 1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, 'active',  'active',  NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(57, 6, 14, 1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, '0',       '0',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),
(58, 6, 15, 1, 'agent-015-e7f8g9h0', '10.0.3.21', 'aws-redis-02', 'Amazon Linux 2 安全基线', 1, 1, '0',       '0',       NULL, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NOW()),

-- ==========================================
-- result_id=19: aws-mysql-01, MySQL基线(baseline_id=6), 8项展示(5通过/3未通过)
-- ==========================================
(59, 19, 38, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 1, '0',       '0',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(60, 19, 39, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 0, '1',       '0',       '存在1个空密码数据库账户', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(61, 19, 40, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 0, '0',       '1',       'general_log未启用', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(62, 19, 41, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 1, '1',       '1',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(63, 19, 42, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 1, '500',     '500',     NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(64, 19, 43, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 0, '1',       '0',       '存在默认test数据库', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(65, 19, 44, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 1, '1',       '1',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(66, 19, 45, 6, 'agent-011-o1p2q3r4', '10.0.3.10', 'aws-mysql-01', 'MySQL 安全基线', 6, 1, '35',      '1',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- ==========================================
-- result_id=20: aws-mysql-02, MySQL基线(baseline_id=6), 8项展示(4通过/4未通过)
-- ==========================================
(67, 20, 38, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 0, '2',       '0',       '存在2个root远程登录授权', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(68, 20, 39, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 0, '2',       '0',       '存在2个空密码数据库账户', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(69, 20, 40, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 0, '0',       '1',       'general_log未启用', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(70, 20, 41, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 1, '1',       '1',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(71, 20, 42, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 1, '500',     '500',     NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(72, 20, 43, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 1, '0',       '0',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(73, 20, 44, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 0, '0',       '1',       'SSL未启用', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),
(74, 20, 45, 6, 'agent-012-s5t6u7v8', '10.0.3.11', 'aws-mysql-02', 'MySQL 安全基线', 6, 1, '28',      '1',       NULL, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NOW()),

-- ==========================================
-- result_id=9: aws-es-01, Ubuntu 22.04基线(baseline_id=2), 14项(全部通过)
-- ==========================================
(75,  9, 16, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '8',       '8',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(76,  9, 17, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '90',      '90',      NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(77,  9, 18, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, 'no',      'no',      NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(78,  9, 19, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '300',     '300',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(79,  9, 20, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(80,  9, 21, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '644',     '644',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(81,  9, 22, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '640',     '640',     NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(82,  9, 23, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, 'active',  'active',  NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(83,  9, 24, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '0',       '0',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(84,  9, 25, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '0',       '0',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(85,  9, 26, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(86,  9, 27, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '1',       '1',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(87,  9, 28, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, '4',       '4',       NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),
(88,  9, 29, 2, 'agent-016-i1j2k3l4', '10.0.3.30', 'aws-es-01', 'Ubuntu 22.04 安全基线', 2, 1, 'active',  'active',  NULL, NOW() - INTERVAL '3 hours', NOW() - INTERVAL '3 hours', NOW()),

-- ==========================================
-- result_id=13: aws-worker-01, Linux通用基线(baseline_id=3), 5项展示(3通过/2未通过)
-- ==========================================
(89, 13, 30, 3, 'agent-009-g3h4i5j6', '10.0.2.20', 'aws-worker-01', 'Linux 通用安全基线', 3, 1, '8',       '8',       NULL, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
(90, 13, 31, 3, 'agent-009-g3h4i5j6', '10.0.2.20', 'aws-worker-01', 'Linux 通用安全基线', 3, 0, 'yes',     'no',      'PermitRootLogin当前值为yes', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
(91, 13, 32, 3, 'agent-009-g3h4i5j6', '10.0.2.20', 'aws-worker-01', 'Linux 通用安全基线', 3, 1, '644',     '644',     NULL, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
(92, 13, 33, 3, 'agent-009-g3h4i5j6', '10.0.2.20', 'aws-worker-01', 'Linux 通用安全基线', 3, 0, 'inactive','active',  'auditd服务未安装', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),
(93, 13, 34, 3, 'agent-009-g3h4i5j6', '10.0.2.20', 'aws-worker-01', 'Linux 通用安全基线', 3, 1, '0',       '0',       NULL, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours', NOW()),

-- ==========================================
-- result_id=16: aws-eks-node-02, Amazon Linux 2023基线(baseline_id=4), 3项展示(2通过/1未通过)
-- ==========================================
(94, 16, 35, 4, 'agent-025-s7t8u9v0', '10.0.4.12', 'aws-eks-node-02', 'Amazon Linux 2023 安全基线', 4, 1, '8',       '8',       NULL, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
(95, 16, 36, 4, 'agent-025-s7t8u9v0', '10.0.4.12', 'aws-eks-node-02', 'Amazon Linux 2023 安全基线', 4, 0, 'yes',     'no',      'PermitRootLogin当前值为yes', NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW()),
(96, 16, 37, 4, 'agent-025-s7t8u9v0', '10.0.4.12', 'aws-eks-node-02', 'Amazon Linux 2023 安全基线', 4, 1, 'active',  'active',  NULL, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '5 hours', NOW());

-- 重置序列
SELECT setval('baseline_check_detail_id_seq', 96);
