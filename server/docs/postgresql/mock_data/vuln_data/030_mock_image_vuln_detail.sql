-- =====================================================
-- 模拟数据: image_vuln_detail (镜像漏洞关联表)
-- 数据量: 35条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
--       VPC CIDR: 10.0.0.0/16
--       关联 image_vuln_scan_task + vuln_info
--       引用真实软件包名，包含 0/1/2 三种状态
--       status: 0-未修复 1-已修复 2-已忽略
--       scan_id 引用 image_vuln_scan_task.id
-- =====================================================

INSERT INTO image_vuln_detail (scan_id, agent_id, image_id, vuln_id, cve_id, package_name, installed_version, fixed_version, status, image_name, vuln_name, severity, cvss_score, description, fix_suggestion, scan_time, created_at, updated_at) VALUES
-- registry.k8s.io/ingress-nginx/controller:v1.10.0 on aws-eks-node-01 (critical:1, high:2, medium:1, low:1)
(8, 'agent-024-o3p4q5r6', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 5,  'CVE-2023-44487', 'ngx_http_v2_module', '1.25.3',   '1.25.4',   0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'HTTP/2 Rapid Reset 拒绝服务漏洞', 'critical', 7.5, 'HTTP/2 Rapid Reset 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '90 days', NOW()),
(8, 'agent-024-o3p4q5r6', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 7,  'CVE-2023-38545', 'curl',              '7.88.1',   '8.4.0',    0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'curl SOCKS5 堆缓冲区溢出漏洞', 'critical', 9.8, 'curl SOCKS5 堆缓冲区溢出漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '90 days', NOW()),
(8, 'agent-024-o3p4q5r6', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 13, 'CVE-2023-3446',  'openssl',           '3.0.9',    '3.0.10',   0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '90 days', NOW()),
(8, 'agent-024-o3p4q5r6', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 23, 'CVE-2023-5678',  'openssl',           '3.0.9',    '3.0.13',   0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '90 days', NOW()),
(8, 'agent-024-o3p4q5r6', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 37, 'CVE-2023-45803', 'python3-urllib3',   '1.26.12',  '1.26.18',  2, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'Python urllib3 请求体泄露漏洞', 'low', 4.2, 'Python urllib3 请求体泄露漏洞', '升级到修复版本', NOW() - INTERVAL '6 hours',  NOW() - INTERVAL '90 days', NOW()),

-- redis:7.2-alpine on aws-eks-node-02 (critical:0, high:1, medium:1, low:1)
(2, 'agent-025-s7t8u9v0', 'sha256:01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2',
 13, 'CVE-2023-3446',  'openssl',           '3.1.2',    '3.1.4',    0, 'redis:7.2-alpine', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '55 days', NOW()),
(2, 'agent-025-s7t8u9v0', 'sha256:01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2',
 23, 'CVE-2023-5678',  'openssl',           '3.1.2',    '3.1.5',    0, 'redis:7.2-alpine', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '55 days', NOW()),
(2, 'agent-025-s7t8u9v0', 'sha256:01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2',
 34, 'CVE-2023-6237',  'openssl',           '3.1.2',    '3.1.5',    0, 'redis:7.2-alpine', 'OpenSSL RSA解密性能漏洞', 'low', 3.7, 'OpenSSL RSA解密性能漏洞', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '55 days', NOW()),

-- mysql:8.0 on aws-eks-node-03 (critical:1, high:2, medium:1, low:0)
(3, 'agent-026-w1x2y3z4', 'sha256:23c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4',
 8,  'CVE-2023-4911',  'glibc',             '2.31',     '2.31-18',  0, 'mysql:8.0', 'glibc ld.so 本地提权漏洞 (Looney Tunables)', 'critical', 7.8, 'glibc ld.so 本地提权漏洞 (Looney Tunables)', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '57 days', NOW()),
(3, 'agent-026-w1x2y3z4', 'sha256:23c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4',
 7,  'CVE-2023-38545', 'curl',              '7.74.0',   '8.4.0',    0, 'mysql:8.0', 'curl SOCKS5 堆缓冲区溢出漏洞', 'critical', 9.8, 'curl SOCKS5 堆缓冲区溢出漏洞', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '57 days', NOW()),
(3, 'agent-026-w1x2y3z4', 'sha256:23c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4',
 18, 'CVE-2023-2650',  'openssl',           '1.1.1n',   '1.1.1u',   0, 'mysql:8.0', 'OpenSSL ASN.1对象标识符处理DoS', 'high', 6.5, 'OpenSSL ASN.1对象标识符处理DoS', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '57 days', NOW()),
(3, 'agent-026-w1x2y3z4', 'sha256:23c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4',
 24, 'CVE-2023-4806',  'glibc',             '2.31',     '2.31-18',  1, 'mysql:8.0', 'glibc getaddrinfo() UAF漏洞', 'medium', 5.9, 'glibc getaddrinfo() UAF漏洞', '升级到修复版本', NOW() - INTERVAL '4 hours',  NOW() - INTERVAL '57 days', NOW()),

-- postgres:16-alpine on aws-jenkins-01 (critical:0, high:1, medium:2, low:1)
(4, 'agent-028-e9f0g1h2', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4',
 13, 'CVE-2023-3446',  'openssl',           '3.1.1',    '3.1.2',    1, 'postgres:16-alpine', 'OpenSSL DH密钥参数检查DoS漏洞', 'high', 5.3, 'OpenSSL DH密钥参数检查DoS漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '100 days', NOW()),
(4, 'agent-028-e9f0g1h2', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4',
 23, 'CVE-2023-5678',  'openssl',           '3.1.1',    '3.1.5',    0, 'postgres:16-alpine', 'OpenSSL DH密钥生成性能问题', 'medium', 5.3, 'OpenSSL DH密钥生成性能问题', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '100 days', NOW()),
(4, 'agent-028-e9f0g1h2', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4',
 29, 'CVE-2024-0727',  'openssl',           '3.1.1',    '3.1.5',    0, 'postgres:16-alpine', 'OpenSSL PKCS12解析空指针解引用', 'medium', 5.5, 'OpenSSL PKCS12解析空指针解引用', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '100 days', NOW()),
(4, 'agent-028-e9f0g1h2', 'sha256:a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4',
 39, 'CVE-2023-2975',  'openssl',           '3.1.1',    '3.1.2',    1, 'postgres:16-alpine', 'OpenSSL AES-SIV空关联数据处理漏洞', 'low', 3.7, 'OpenSSL AES-SIV空关联数据处理漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '100 days', NOW()),

-- docker.elastic.co/elasticsearch/elasticsearch:8.12.2 on aws-eks-node-04 (critical:0, high:2, medium:1, low:1)
(5, 'agent-027-a5b6c7d8', 'sha256:56f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
 15, 'CVE-2023-39325', 'golang.org/x/net',  '0.15.0',   '0.17.0',   0, 'docker.elastic.co/elasticsearch/elasticsearch:8.12.2', 'Go net/http HTTP/2 拒绝服务漏洞', 'high', 7.5, 'Go net/http HTTP/2 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '55 days', NOW()),
(5, 'agent-027-a5b6c7d8', 'sha256:56f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
 17, 'CVE-2023-45853', 'zlib',              '1.2.13',   '1.3',      0, 'docker.elastic.co/elasticsearch/elasticsearch:8.12.2', 'zlib MiniZip 整数溢出漏洞', 'high', 9.8, 'zlib MiniZip 整数溢出漏洞', '升级到修复版本', NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '55 days', NOW()),
(5, 'agent-027-a5b6c7d8', 'sha256:56f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
 25, 'CVE-2023-39326', 'golang.org/x/net',  '0.15.0',   '0.19.0',   0, 'docker.elastic.co/elasticsearch/elasticsearch:8.12.2', 'Go net/http 请求体读取漏洞', 'medium', 5.3, 'Go net/http 请求体读取漏洞', '升级到修复版本', NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '55 days', NOW()),
(5, 'agent-027-a5b6c7d8', 'sha256:56f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7',
 34, 'CVE-2023-6237',  'openssl',           '3.1.3',    '3.1.5',    0, 'docker.elastic.co/elasticsearch/elasticsearch:8.12.2', 'OpenSSL RSA解密性能漏洞', 'low', 3.7, 'OpenSSL RSA解密性能漏洞', '升级到修复版本', NOW() - INTERVAL '5 hours',  NOW() - INTERVAL '55 days', NOW()),

-- company/frontend:v3.2.0 on aws-eks-node-01 (critical:0, high:1, medium:2, low:1)
(6, 'agent-024-o3p4q5r6', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8',
 15, 'CVE-2023-39325', 'golang.org/x/net',  '0.14.0',   '0.17.0',   0, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend:v3.2.0', 'Go net/http HTTP/2 拒绝服务漏洞', 'high', 7.5, 'Go net/http HTTP/2 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
(6, 'agent-024-o3p4q5r6', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8',
 16, 'CVE-2023-44270', 'postcss',           '8.4.28',   '8.4.31',   0, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend:v3.2.0', 'PostCSS 换行符解析漏洞', 'high', 5.3, 'PostCSS 换行符解析漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
(6, 'agent-024-o3p4q5r6', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8',
 26, 'CVE-2023-45287', 'golang.org/x/crypto','0.12.0',  '0.17.0',   0, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend:v3.2.0', 'Go crypto/tls RSA密钥交换时序泄露漏洞', 'medium', 5.3, 'Go crypto/tls RSA密钥交换时序泄露漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
(6, 'agent-024-o3p4q5r6', 'sha256:c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8',
 37, 'CVE-2023-45803', 'python3-urllib3',   '1.26.15',  '1.26.18',  2, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/frontend:v3.2.0', 'Python urllib3 请求体泄露漏洞', 'low', 4.2, 'Python urllib3 请求体泄露漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),

-- company/backend:v4.1.0 on aws-eks-node-01 (critical:1, high:2, medium:1, low:0)
(7, 'agent-024-o3p4q5r6', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9',
 6,  'CVE-2024-21626', 'runc',              '1.1.10',   '1.1.12',   0, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend:v4.1.0', 'runc 容器逃逸漏洞 (Leaky Vessels)', 'critical', 8.6, 'runc 容器逃逸漏洞 (Leaky Vessels)', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
(7, 'agent-024-o3p4q5r6', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9',
 15, 'CVE-2023-39325', 'golang.org/x/net',  '0.13.0',   '0.17.0',   0, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend:v4.1.0', 'Go net/http HTTP/2 拒绝服务漏洞', 'high', 7.5, 'Go net/http HTTP/2 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
(7, 'agent-024-o3p4q5r6', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9',
 21, 'CVE-2023-48795', 'golang.org/x/crypto','0.14.0',  '0.17.0',   0, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend:v4.1.0', 'OpenSSH Terrapin 前缀截断攻击漏洞', 'high', 5.9, 'OpenSSH Terrapin 前缀截断攻击漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),
(7, 'agent-024-o3p4q5r6', 'sha256:d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9',
 32, 'CVE-2023-29406', 'golang.org/x/net',  '0.13.0',   '0.14.0',   1, '123456789012.dkr.ecr.ap-southeast-1.amazonaws.com/company/backend:v4.1.0', 'Go net/http Host头注入漏洞', 'medium', 6.5, 'Go net/http Host头注入漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '50 days', NOW()),

-- registry.k8s.io/ingress-nginx/controller:v1.10.0 on aws-eks-master-01 (critical:1, high:1, medium:1, low:0)
(8, 'agent-023-k9l0m1n2', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 5,  'CVE-2023-44487', 'ngx_http_v2_module', '1.25.3',  '1.25.4',   0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'HTTP/2 Rapid Reset 拒绝服务漏洞', 'critical', 7.5, 'HTTP/2 Rapid Reset 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(8, 'agent-023-k9l0m1n2', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 15, 'CVE-2023-39325', 'golang.org/x/net',  '0.14.0',   '0.17.0',   0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'Go net/http HTTP/2 拒绝服务漏洞', 'high', 7.5, 'Go net/http HTTP/2 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),
(8, 'agent-023-k9l0m1n2', 'sha256:b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7',
 25, 'CVE-2023-39326', 'golang.org/x/net',  '0.14.0',   '0.19.0',   0, 'registry.k8s.io/ingress-nginx/controller:v1.10.0', 'Go net/http 请求体读取漏洞', 'medium', 5.3, 'Go net/http 请求体读取漏洞', '升级到修复版本', NOW() - INTERVAL '3 hours',  NOW() - INTERVAL '58 days', NOW()),

-- goharbor/harbor-core:v2.10.0 on aws-harbor-01 (critical:0, high:1, medium:1, low:1)
(9, 'agent-030-m7n8o9p0', 'sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2',
 15, 'CVE-2023-39325', 'golang.org/x/net',  '0.12.0',   '0.17.0',   0, 'goharbor/harbor-core:v2.10.0', 'Go net/http HTTP/2 拒绝服务漏洞', 'high', 7.5, 'Go net/http HTTP/2 拒绝服务漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '80 days', NOW()),
(9, 'agent-030-m7n8o9p0', 'sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2',
 26, 'CVE-2023-45287', 'golang.org/x/crypto','0.11.0',  '0.17.0',   0, 'goharbor/harbor-core:v2.10.0', 'Go crypto/tls RSA密钥交换时序泄露漏洞', 'medium', 5.3, 'Go crypto/tls RSA密钥交换时序泄露漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '80 days', NOW()),
(9, 'agent-030-m7n8o9p0', 'sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2',
 37, 'CVE-2023-45803', 'python3-urllib3',   '1.26.14',  '1.26.18',  0, 'goharbor/harbor-core:v2.10.0', 'Python urllib3 请求体泄露漏洞', 'low', 4.2, 'Python urllib3 请求体泄露漏洞', '升级到修复版本', NOW() - INTERVAL '7 hours',  NOW() - INTERVAL '80 days', NOW()),

-- python:3.12-slim on aws-gitlab-01 (critical:0, high:1, medium:1, low:1)
(14, 'agent-029-i3j4k5l6', 'sha256:45c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6',
 20, 'CVE-2023-6246',  'glibc',             '2.36',     '2.39',     0, 'python:3.12-slim', 'glibc __fortify_fail 本地提权漏洞', 'high', 7.8, 'glibc __fortify_fail 本地提权漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '20 days', NOW()),
(14, 'agent-029-i3j4k5l6', 'sha256:45c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6',
 28, 'CVE-2023-52425', 'libexpat',          '2.5.0',    '2.6.0',    0, 'python:3.12-slim', 'libexpat XML解析DoS漏洞', 'medium', 5.5, 'libexpat XML解析DoS漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '20 days', NOW()),
(14, 'agent-029-i3j4k5l6', 'sha256:45c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6',
 38, 'CVE-2023-44271', 'python3-pil',       '9.5.0',    '10.0.1',   0, 'python:3.12-slim', 'Pillow 图像解析DoS漏洞', 'low', 3.3, 'Pillow 图像解析DoS漏洞', '升级到修复版本', NOW() - INTERVAL '8 hours',  NOW() - INTERVAL '20 days', NOW());
