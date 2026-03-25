-- =====================================================
-- 模拟数据: host_vuln_detail (主机漏洞关联表)
-- 数据量: 50条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
--       VPC CIDR: 10.0.0.0/16
--       关联 host_vuln_scan_task + vuln_info
--       引用真实软件包名，包含 0/1/2 三种状态
--       status: 0-未修复 1-已修复 2-已忽略
--       scan_id 引用 host_vuln_scan_task.id
-- =====================================================

INSERT INTO host_vuln_detail (scan_id, agent_id, host_id, vuln_id, cve_id, package_name, installed_version, fixed_version, status, host_name, host_ip, vuln_name, severity, cvss_score, description, fix_suggestion, scan_time, created_at, updated_at) VALUES
-- aws-web-01 (critical:1, high:3, medium:2, low:1)
(1, 'agent-001-a1b2c3d4', 1,  7,  'CVE-2023-38545', 'curl',           '7.88.1',    '8.4.0',    0, 'aws-web-01', '10.0.1.10', 'curl SOCKS5 堆缓冲区溢出漏洞', 'critical', 9.8, 'curl SOCKS5 堆缓冲区溢出漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(1, 'agent-001-a1b2c3d4', 1,  13, 'CVE-2023-3446',  'openssl',        '3.0.9',     '3.0.10',   0, 'aws-web-01', '10.0.1.10', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(1, 'agent-001-a1b2c3d4', 1,  14, 'CVE-2023-5363',  'openssl',        '3.0.9',     '3.0.12',   0, 'aws-web-01', '10.0.1.10', 'OpenSSL密钥和IV长度处理漏洞', 'high', 7.5, 'OpenSSL密钥和IV长度处理漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(1, 'agent-001-a1b2c3d4', 1,  21, 'CVE-2023-48795', 'openssh-client', '9.2p1',     '9.6p1',    0, 'aws-web-01', '10.0.1.10', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9, 'OpenSSH Terrapin 前缀截断攻击漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(1, 'agent-001-a1b2c3d4', 1,  23, 'CVE-2023-5678',  'openssl',        '3.0.9',     '3.0.13',   0, 'aws-web-01', '10.0.1.10', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(1, 'agent-001-a1b2c3d4', 1,  29, 'CVE-2024-0727',  'openssl',        '3.0.9',     '3.0.13',   0, 'aws-web-01', '10.0.1.10', 'OpenSSL PKCS12解析空指针解引用', 'medium', 5.5, 'OpenSSL PKCS12解析空指针解引用', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),
(1, 'agent-001-a1b2c3d4', 1,  33, 'CVE-2023-5156',  'glibc',          '2.35',      '2.38-4',   1, 'aws-web-01', '10.0.1.10', 'glibc getaddrinfo() 内存泄漏', 'low', 3.7, 'glibc getaddrinfo() 内存泄漏', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '30 days', NOW()),

-- aws-web-02 (critical:1, high:2, medium:2, low:1)
(2, 'agent-002-e5f6g7h8', 2,  2,  'CVE-2024-6387',  'openssh-server', '8.9p1',     '9.8p1',    0, 'aws-web-02', '10.0.1.11', 'OpenSSH regreSSHion 远程代码执行漏洞', 'critical', 8.1, 'OpenSSH regreSSHion 远程代码执行漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),
(2, 'agent-002-e5f6g7h8', 2,  13, 'CVE-2023-3446',  'openssl',        '3.0.9',     '3.0.10',   0, 'aws-web-02', '10.0.1.11', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),
(2, 'agent-002-e5f6g7h8', 2,  21, 'CVE-2023-48795', 'openssh-client', '9.2p1',     '9.6p1',    0, 'aws-web-02', '10.0.1.11', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9, 'OpenSSH Terrapin 前缀截断攻击漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),
(2, 'agent-002-e5f6g7h8', 2,  23, 'CVE-2023-5678',  'openssl',        '3.0.9',     '3.0.13',   0, 'aws-web-02', '10.0.1.11', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),
(2, 'agent-002-e5f6g7h8', 2,  29, 'CVE-2024-0727',  'openssl',        '3.0.9',     '3.0.13',   2, 'aws-web-02', '10.0.1.11', 'OpenSSL PKCS12解析空指针解引用', 'medium', 5.5, 'OpenSSL PKCS12解析空指针解引用', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),
(2, 'agent-002-e5f6g7h8', 2,  34, 'CVE-2023-6237',  'openssl',        '3.0.9',     '3.0.13',   1, 'aws-web-02', '10.0.1.11', 'OpenSSL RSA解密性能漏洞', 'low', 3.7, 'OpenSSL RSA解密性能漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '28 days', NOW()),

-- aws-api-01 (critical:2, high:2, medium:1, low:1)
(3, 'agent-003-i9j0k1l2', 3,  2,  'CVE-2024-6387',  'openssh-server', '7.4p1',     '9.8p1',    0, 'aws-api-01', '10.0.1.20', 'OpenSSH regreSSHion 远程代码执行漏洞', 'critical', 8.1, 'OpenSSH regreSSHion 远程代码执行漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(3, 'agent-003-i9j0k1l2', 3,  4,  'CVE-2021-3156',  'sudo',           '1.8.23',    '1.9.5p2',  0, 'aws-api-01', '10.0.1.20', 'Sudo Buffer Overflow 提权漏洞 (Baron Samedit)', 'critical', 7.8, 'Sudo Buffer Overflow 提权漏洞 (Baron Samedit)', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(3, 'agent-003-i9j0k1l2', 3,  18, 'CVE-2023-2650',  'openssl',        '1.0.2k',    '1.0.2zj',  0, 'aws-api-01', '10.0.1.20', 'OpenSSL ASN.1对象标识符处理DoS', 'high', 6.5, 'OpenSSL ASN.1对象标识符处理DoS', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(3, 'agent-003-i9j0k1l2', 3,  13, 'CVE-2023-3446',  'openssl',        '1.0.2k',    '1.0.2zj',  0, 'aws-api-01', '10.0.1.20', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(3, 'agent-003-i9j0k1l2', 3,  27, 'CVE-2023-4527',  'glibc',          '2.17',      '2.17-326', 0, 'aws-api-01', '10.0.1.20', 'glibc getaddrinfo() 栈缓冲区溢出', 'medium', 6.5, 'glibc getaddrinfo() 栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),
(3, 'agent-003-i9j0k1l2', 3,  35, 'CVE-2023-4016',  'procps-ng',      '3.3.10',    '3.3.17',   2, 'aws-api-01', '10.0.1.20', 'procps-ng ps命令栈缓冲区溢出', 'low', 3.3, 'procps-ng ps命令栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '60 days', NOW()),

-- aws-api-02 (critical:1, high:2, medium:2, low:0)
(4, 'agent-004-m3n4o5p6', 4,  4,  'CVE-2021-3156',  'sudo',           '1.8.23',    '1.9.5p2',  0, 'aws-api-02', '10.0.1.21', 'Sudo Buffer Overflow 提权漏洞 (Baron Samedit)', 'critical', 7.8, 'Sudo Buffer Overflow 提权漏洞 (Baron Samedit)', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),
(4, 'agent-004-m3n4o5p6', 4,  18, 'CVE-2023-2650',  'openssl',        '1.0.2k',    '1.0.2zj',  0, 'aws-api-02', '10.0.1.21', 'OpenSSL ASN.1对象标识符处理DoS', 'high', 6.5, 'OpenSSL ASN.1对象标识符处理DoS', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),
(4, 'agent-004-m3n4o5p6', 4,  13, 'CVE-2023-3446',  'openssl',        '1.0.2k',    '1.0.2zj',  0, 'aws-api-02', '10.0.1.21', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),
(4, 'agent-004-m3n4o5p6', 4,  24, 'CVE-2023-4806',  'glibc',          '2.17',      '2.17-326', 0, 'aws-api-02', '10.0.1.21', 'glibc getaddrinfo() UAF漏洞', 'medium', 5.9, 'glibc getaddrinfo() UAF漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),
(4, 'agent-004-m3n4o5p6', 4,  27, 'CVE-2023-4527',  'glibc',          '2.17',      '2.17-326', 1, 'aws-api-02', '10.0.1.21', 'glibc getaddrinfo() 栈缓冲区溢出', 'medium', 6.5, 'glibc getaddrinfo() 栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '55 days', NOW()),

-- aws-gateway-01 (critical:1, high:2, medium:1, low:1)
(5, 'agent-005-q7r8s9t0', 5,  8,  'CVE-2023-4911',  'glibc',          '2.31',      '2.31-18',  0, 'aws-gateway-01', '10.0.1.30', 'glibc ld.so 本地提权漏洞 (Looney Tunables)', 'critical', 7.8, 'glibc ld.so 本地提权漏洞 (Looney Tunables)', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(5, 'agent-005-q7r8s9t0', 5,  17, 'CVE-2023-45853', 'zlib',           '1.2.11',    '1.3',      0, 'aws-gateway-01', '10.0.1.30', 'zlib MiniZip 整数溢出漏洞', 'high', 9.8, 'zlib MiniZip 整数溢出漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(5, 'agent-005-q7r8s9t0', 5,  21, 'CVE-2023-48795', 'openssh-client', '8.4p1',     '9.6p1',    0, 'aws-gateway-01', '10.0.1.30', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9, 'OpenSSH Terrapin 前缀截断攻击漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(5, 'agent-005-q7r8s9t0', 5,  28, 'CVE-2023-52425', 'libexpat',       '2.4.1',     '2.6.0',    0, 'aws-gateway-01', '10.0.1.30', 'libexpat XML解析DoS漏洞', 'medium', 5.5, 'libexpat XML解析DoS漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(5, 'agent-005-q7r8s9t0', 5,  39, 'CVE-2023-2975',  'openssl',        '3.0.7',     '3.0.10',   1, 'aws-gateway-01', '10.0.1.30', 'OpenSSL AES-SIV空关联数据处理漏洞', 'low', 3.7, 'OpenSSL AES-SIV空关联数据处理漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),

-- aws-app-01 (critical:0, high:2, medium:2, low:1)
(6, 'agent-006-u1v2w3x4', 6,  17, 'CVE-2023-45853', 'zlib',           '1.2.11',    '1.3',      0, 'aws-app-01', '10.0.2.10', 'zlib MiniZip 整数溢出漏洞', 'high', 9.8, 'zlib MiniZip 整数溢出漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(6, 'agent-006-u1v2w3x4', 6,  21, 'CVE-2023-48795', 'openssh-client', '8.4p1',     '9.6p1',    0, 'aws-app-01', '10.0.2.10', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9, 'OpenSSH Terrapin 前缀截断攻击漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(6, 'agent-006-u1v2w3x4', 6,  28, 'CVE-2023-52425', 'libexpat',       '2.4.1',     '2.6.0',    0, 'aws-app-01', '10.0.2.10', 'libexpat XML解析DoS漏洞', 'medium', 5.5, 'libexpat XML解析DoS漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(6, 'agent-006-u1v2w3x4', 6,  24, 'CVE-2023-4806',  'glibc',          '2.31',      '2.31-18',  0, 'aws-app-01', '10.0.2.10', 'glibc getaddrinfo() UAF漏洞', 'medium', 5.9, 'glibc getaddrinfo() UAF漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),
(6, 'agent-006-u1v2w3x4', 6,  33, 'CVE-2023-5156',  'glibc',          '2.31',      '2.31-18',  0, 'aws-app-01', '10.0.2.10', 'glibc getaddrinfo() 内存泄漏', 'low', 3.7, 'glibc getaddrinfo() 内存泄漏', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '45 days', NOW()),

-- aws-app-02 (critical:1, high:2, medium:1, low:1)
(7, 'agent-007-y5z6a7b8', 7,  2,  'CVE-2024-6387',  'openssh-server', '8.2p1',     '9.8p1',    0, 'aws-app-02', '10.0.2.11', 'OpenSSH regreSSHion 远程代码执行漏洞', 'critical', 8.1, 'OpenSSH regreSSHion 远程代码执行漏洞', '升级到修复版本', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '90 days', NOW()),
(7, 'agent-007-y5z6a7b8', 7,  13, 'CVE-2023-3446',  'openssl',        '1.1.1f',    '1.1.1v',   0, 'aws-app-02', '10.0.2.11', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '90 days', NOW()),
(7, 'agent-007-y5z6a7b8', 7,  20, 'CVE-2023-6246',  'glibc',          '2.31',      '2.39',     0, 'aws-app-02', '10.0.2.11', 'glibc __fortify_fail 本地提权漏洞', 'high', 7.8, 'glibc __fortify_fail 本地提权漏洞', '升级到修复版本', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '90 days', NOW()),
(7, 'agent-007-y5z6a7b8', 7,  23, 'CVE-2023-5678',  'openssl',        '1.1.1f',    '1.1.1x',   0, 'aws-app-02', '10.0.2.11', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '90 days', NOW()),
(7, 'agent-007-y5z6a7b8', 7,  35, 'CVE-2023-4016',  'procps-ng',      '3.3.16',    '3.3.17',   2, 'aws-app-02', '10.0.2.11', 'procps-ng ps命令栈缓冲区溢出', 'low', 3.3, 'procps-ng ps命令栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '10 hours', NOW() - INTERVAL '90 days', NOW()),

-- aws-eks-master-01 (critical:1, high:2, medium:1, low:1)
(18, 'agent-023-k9l0m1n2', 25, 9,  'CVE-2024-1086',  'linux-kernel',   '5.15.0',    '5.15.149', 0, 'aws-eks-master-01', '10.0.4.10', 'Linux内核netfilter nf_tables本地提权漏洞', 'critical', 7.8, 'Linux内核netfilter nf_tables本地提权漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(18, 'agent-023-k9l0m1n2', 25, 21, 'CVE-2023-48795', 'openssh-client', '9.0p1',     '9.6p1',    0, 'aws-eks-master-01', '10.0.4.10', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9, 'OpenSSH Terrapin 前缀截断攻击漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(18, 'agent-023-k9l0m1n2', 25, 13, 'CVE-2023-3446',  'openssl',        '3.0.9',     '3.0.10',   1, 'aws-eks-master-01', '10.0.4.10', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(18, 'agent-023-k9l0m1n2', 25, 23, 'CVE-2023-5678',  'openssl',        '3.0.9',     '3.0.13',   0, 'aws-eks-master-01', '10.0.4.10', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(18, 'agent-023-k9l0m1n2', 25, 40, 'CVE-2023-4641',  'shadow-utils',   '4.8.1',     '4.14.0',   0, 'aws-eks-master-01', '10.0.4.10', 'shadow-utils useradd密码信息泄露', 'low', 3.3, 'shadow-utils useradd密码信息泄露', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),

-- aws-pg-01 (critical:0, high:1, medium:1, low:1)
(9, 'agent-013-w9x0y1z2', 13, 18, 'CVE-2023-2650',  'openssl',        '1.0.2k',    '1.0.2zj',  0, 'aws-pg-01', '10.0.3.12', 'OpenSSL ASN.1对象标识符处理DoS', 'high', 6.5, 'OpenSSL ASN.1对象标识符处理DoS', '升级到修复版本', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '150 days', NOW()),
(9, 'agent-013-w9x0y1z2', 13, 27, 'CVE-2023-4527',  'glibc',          '2.17',      '2.17-326', 0, 'aws-pg-01', '10.0.3.12', 'glibc getaddrinfo() 栈缓冲区溢出', 'medium', 6.5, 'glibc getaddrinfo() 栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '150 days', NOW()),
(9, 'agent-013-w9x0y1z2', 13, 35, 'CVE-2023-4016',  'procps-ng',      '3.3.10',    '3.3.17',   1, 'aws-pg-01', '10.0.3.12', 'procps-ng ps命令栈缓冲区溢出', 'low', 3.3, 'procps-ng ps命令栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '12 hours', NOW() - INTERVAL '150 days', NOW()),

-- aws-zk-01 (critical:0, high:1, medium:1, low:1)
(15, 'agent-047-c5d6e7f8', 23, 13, 'CVE-2023-3446',  'openssl',        '1.0.2k',    '1.0.2zj',  0, 'aws-zk-01', '10.0.3.72', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '11 hours', NOW() - INTERVAL '180 days', NOW()),
(15, 'agent-047-c5d6e7f8', 23, 24, 'CVE-2023-4806',  'glibc',          '2.17',      '2.17-326', 0, 'aws-zk-01', '10.0.3.72', 'glibc getaddrinfo() UAF漏洞', 'medium', 5.9, 'glibc getaddrinfo() UAF漏洞', '升级到修复版本', NOW() - INTERVAL '11 hours', NOW() - INTERVAL '180 days', NOW()),
(15, 'agent-047-c5d6e7f8', 23, 35, 'CVE-2023-4016',  'procps-ng',      '3.3.10',    '3.3.17',   2, 'aws-zk-01', '10.0.3.72', 'procps-ng ps命令栈缓冲区溢出', 'low', 3.3, 'procps-ng ps命令栈缓冲区溢出', '升级到修复版本', NOW() - INTERVAL '11 hours', NOW() - INTERVAL '180 days', NOW());
