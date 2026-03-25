-- =====================================================
-- HCIDS 网络攻击告警测试数据
-- 目标服务器: 192.168.215.165
-- Agent ID: 77533a6b-9edc-4198-a834-a52bbc42b340
-- 生成时间: 2026-03-19
-- =====================================================

-- 1. SQL 注入攻击 (Web应用攻击)
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    80,
    '192.168.1.100',
    NULL,
    NULL,
    'sql_injection',
    'CVE-2024-1234',
    'detected',
    15,
    NOW() - INTERVAL '1 hour',
    NOW(),
    ''' OR ''1''=''1 --',
    0,
    NOW(),
    NOW()
);

-- 2. SSH 暴力破解攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    22,
    '10.0.0.50',
    NULL,
    NULL,
    'brute_force',
    NULL,
    'detected',
    56,
    NOW() - INTERVAL '30 minutes',
    NOW(),
    'SSH authentication failure for user root from 10.0.0.50 port 52341',
    0,
    NOW(),
    NOW()
);

-- 3. XSS 跨站脚本攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    8080,
    '172.16.0.25',
    NULL,
    NULL,
    'xss',
    'CVE-2024-5678',
    'detected',
    8,
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '10 minutes',
    '<script>alert("XSS")</script>',
    0,
    NOW(),
    NOW()
);

-- 4. 远程代码执行 (RCE) 攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    443,
    '192.168.2.200',
    NULL,
    NULL,
    'rce',
    'CVE-2023-9999',
    'detected',
    3,
    NOW() - INTERVAL '15 minutes',
    NOW(),
    'GET /api/exec?cmd=whoami HTTP/1.1',
    0,
    NOW(),
    NOW()
);

-- 5. 端口扫描检测
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    3306,
    '192.168.3.50',
    NULL,
    NULL,
    'port_scan',
    NULL,
    'detected',
    22,
    NOW() - INTERVAL '45 minutes',
    NOW(),
    'TCP SYN scan detected on multiple ports',
    0,
    NOW(),
    NOW()
);

-- 6. 缓冲区溢出攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    21,
    '192.168.5.100',
    NULL,
    NULL,
    'buffer_overflow',
    'CVE-2023-1111',
    'detected',
    2,
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '2 hours',
    'FTP USER command buffer overflow attempt',
    0,
    NOW(),
    NOW()
);

-- 7. DoS 拒绝服务攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    80,
    '192.168.10.0',
    NULL,
    NULL,
    'dos',
    NULL,
    'mitigated',
    1500,
    NOW() - INTERVAL '4 hours',
    NOW() - INTERVAL '3 hours',
    'SYN flood attack detected, connection rate: 500/s',
    1,
    NOW(),
    NOW()
);

-- 8. DDoS 分布式拒绝服务攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    443,
    '10.10.10.10',
    NULL,
    NULL,
    'ddos',
    NULL,
    'detected',
    8000,
    NOW() - INTERVAL '6 hours',
    NOW() - INTERVAL '5 hours',
    'Distributed amplification attack from multiple sources',
    0,
    NOW(),
    NOW()
);

-- 9. Webshell 上传攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    80,
    '192.168.20.50',
    NULL,
    NULL,
    'webshell',
    NULL,
    'detected',
    1,
    NOW() - INTERVAL '20 minutes',
    NOW() - INTERVAL '20 minutes',
    'POST /upload.php: suspicious file shell.php detected',
    0,
    NOW(),
    NOW()
);

-- 10. LDAP 注入攻击
INSERT INTO alert_network_attack (
    agent_id, host_name, host_ip, target_port, attacker_ip,
    attacker_location, attacker_country, vulnerability_name, vulnerability_id,
    attack_status, attack_count, first_attack_time, last_attack_time,
    attack_payload, status, created_at, updated_at
) VALUES (
    '77533a6b-9edc-4198-a834-a52bbc42b340',
    'server-165',
    '192.168.215.165',
    389,
    '192.168.15.100',
    NULL,
    NULL,
    'ldap_injection',
    'CVE-2024-2222',
    'detected',
    5,
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '30 minutes',
    '(*)(uid=*))(|(uid=*',
    0,
    NOW(),
    NOW()
);

-- =====================================================
-- 查询验证
-- =====================================================
-- SELECT * FROM alert_network_attack
-- WHERE agent_id = '77533a6b-9edc-4198-a834-a52bbc42b340'
-- ORDER BY last_attack_time DESC;

-- =====================================================
-- 清理测试数据 (如需要)
-- =====================================================
-- DELETE FROM alert_network_attack
-- WHERE agent_id = '77533a6b-9edc-4198-a834-a52bbc42b340';
