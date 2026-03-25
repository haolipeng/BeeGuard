-- =====================================================
-- 模拟数据: alert_malicious_request (恶意请求告警表)
-- 数据量: 35条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- 基于 asset_host 中的主机生成恶意请求告警数据
-- policy_type: mining/c2/phishing/botnet/ransomware
-- =====================================================

INSERT INTO alert_malicious_request (agent_id, host_id, host_name, host_ip, policy_type, policy_name, malicious_domain, malicious_ip, request_count, first_request_time, last_request_time, risk_description, status, created_at, updated_at) VALUES
-- mining: 挖矿
('agent-023-k9l0m1n2', 25, 'aws-eks-master-01', '10.0.4.10', 'mining', '挖矿域名检测规则', 'pool.minexmr.com', '104.18.24.136', 1567, NOW() - INTERVAL '3 days', NOW() - INTERVAL '30 minutes', '检测到与门罗币矿池通信，疑似挖矿程序运行', 0, NOW() - INTERVAL '30 minutes', NOW()),
('agent-024-o3p4q5r6', 26, 'aws-eks-node-01', '10.0.4.11', 'mining', '挖矿域名检测规则', 'xmr.pool.minergate.com', '94.130.12.27', 892, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 hour', '检测到MinerGate矿池连接请求', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-028-e9f0g1h2', 30, 'aws-jenkins-01', '10.0.5.10', 'mining', '挖矿IP检测规则', 'stratum+tcp://pool.hashvault.pro', '51.91.35.76', 2341, NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 hours', '检测到Stratum协议挖矿通信', 1, NOW() - INTERVAL '2 hours', NOW()),
('agent-031-q1r2s3t4', 33, 'aws-nexus-01', '10.0.5.13', 'mining', '挖矿域名检测规则', 'pool.supportxmr.com', '139.99.124.170', 456, NOW() - INTERVAL '1 day', NOW() - INTERVAL '4 hours', '检测到XMR支持矿池连接', 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-005-q7r8s9t0', 5, 'aws-gateway-01', '10.0.1.30', 'mining', '挖矿行为检测规则', 'eth-pool.crypto-pool.fr', '51.255.32.150', 789, NOW() - INTERVAL '4 days', NOW() - INTERVAL '6 hours', '检测到以太坊矿池通信行为', 0, NOW() - INTERVAL '6 hours', NOW()),
('agent-007-y5z6a7b8', 7, 'aws-app-02', '10.0.2.11', 'mining', '挖矿域名检测规则', 'pool.hashrate.to', '185.117.152.92', 1234, NOW() - INTERVAL '2 days', NOW() - INTERVAL '3 hours', '检测到与未知矿池域名通信', 1, NOW() - INTERVAL '3 hours', NOW()),
('agent-050-o7p8q9r0', 50, 'aws-vault-01', '10.0.7.20', 'mining', '挖矿域名检测规则', 'randomx.xmrig.com', '104.24.96.79', 567, NOW() - INTERVAL '1 day', NOW() - INTERVAL '5 hours', '检测到XMRig矿机程序通信', 0, NOW() - INTERVAL '5 hours', NOW()),

-- c2: C2通信
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'c2', 'C2通信检测规则', 'evil-c2.malware.net', '45.33.32.156', 234, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '20 minutes', '检测到与已知C2服务器通信，疑似后门程序', 0, NOW() - INTERVAL '20 minutes', NOW()),
('agent-002-e5f6g7h8', 2, 'aws-web-02', '10.0.1.11', 'c2', 'Cobalt Strike检测规则', 'beacon.c2server.xyz', '185.220.101.35', 456, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 hour', '检测到Cobalt Strike Beacon通信特征', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-017-m5n6o7p8', 17, 'aws-es-02', '10.0.3.31', 'c2', 'C2通信检测规则', 'cmd-control.badactor.com', '91.121.87.18', 123, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '2 hours', '检测到HTTP隧道C2通信', 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', 'c2', 'Linux C2框架检测规则', 'linux-beacon.attacker.io', '103.25.61.114', 345, NOW() - INTERVAL '2 days', NOW() - INTERVAL '3 hours', '检测到Linux C2框架通信', 1, NOW() - INTERVAL '3 hours', NOW()),
('agent-040-a7b8c9d0', 42, 'aws-dns-01', '10.0.7.12', 'c2', 'Metasploit检测规则', 'msf.hacker-domain.net', '45.155.205.233', 567, NOW() - INTERVAL '1 day', NOW() - INTERVAL '45 minutes', '检测到Metasploit Meterpreter通信', 0, NOW() - INTERVAL '45 minutes', NOW()),
('agent-018-q9r0s1t2', 18, 'aws-es-03', '10.0.3.32', 'c2', 'DNS隧道检测规则', 'tunnel.dns-c2.xyz', '185.156.73.54', 890, NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 hours', '检测到DNS隧道C2通信', 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-029-i3j4k5l6', 31, 'aws-gitlab-01', '10.0.5.11', 'c2', 'Sliver C2框架检测规则', 'sliver-implant.c2framework.xyz', '198.51.100.78', 178, NOW() - INTERVAL '1 day', NOW() - INTERVAL '90 minutes', '检测到Sliver C2框架植入体通信', 0, NOW() - INTERVAL '90 minutes', NOW()),

-- phishing: 钓鱼网站
('agent-001-a1b2c3d4', 1, 'aws-web-01', '10.0.1.10', 'phishing', '钓鱼网站检测规则', 'www.paypa1-secure.com', '91.240.118.172', 12, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '30 minutes', '检测到访问仿冒PayPal钓鱼网站', 0, NOW() - INTERVAL '30 minutes', NOW()),
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', 'phishing', '钓鱼网站检测规则', 'aws-console-login.phishing.net', '45.143.220.115', 8, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '1 hour', '检测到访问仿冒AWS控制台钓鱼网站', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-038-s9t0u1v2', 40, 'aws-vpn-01', '10.0.7.10', 'phishing', '钓鱼网站检测规则', 'github-enterprise.phishing.net', '23.129.64.130', 15, NOW() - INTERVAL '1 day', NOW() - INTERVAL '2 hours', '检测到访问仿冒GitHub Enterprise钓鱼网站', 1, NOW() - INTERVAL '2 hours', NOW()),
('agent-045-u7v8w9x0', 47, 'aws-backup-01', '10.0.7.17', 'phishing', '钓鱼邮件检测规则', 'mail-verify.secure-gmail.xyz', '103.74.192.18', 34, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '3 hours', '检测到钓鱼邮件链接访问', 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-041-e1f2g3h4', 43, 'aws-nfs-01', '10.0.7.13', 'phishing', '钓鱼网站检测规则', 'webmail.roundcube-login.xyz', '185.161.248.12', 23, NOW() - INTERVAL '12 hours', NOW() - INTERVAL '5 hours', '检测到仿冒Roundcube邮件系统钓鱼网站访问', 0, NOW() - INTERVAL '5 hours', NOW()),
('agent-047-c5d6e7f8', 23, 'aws-zk-01', '10.0.3.70', 'phishing', '钓鱼网站检测规则', 'sso-login.company-fake.com', '103.153.78.45', 5, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '4 hours', '检测到仿冒企业SSO钓鱼网站', 0, NOW() - INTERVAL '4 hours', NOW()),

-- botnet: 僵尸网络
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'botnet', 'Mirai僵尸网络检测规则', 'mirai-cnc.botnet.io', '58.218.198.160', 1234, NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 hour', '检测到与Mirai僵尸网络C&C通信', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-009-g3h4i5j6', 9, 'aws-worker-01', '10.0.2.20', 'botnet', 'Emotet僵尸网络检测规则', 'emotet.command.net', '61.177.173.25', 567, NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 hours', '检测到Emotet僵尸网络通信特征', 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-025-s7t8u9v0', 27, 'aws-eks-node-02', '10.0.4.12', 'botnet', 'TrickBot检测规则', 'trickbot-c2.malicious.xyz', '119.45.227.38', 890, NOW() - INTERVAL '1 day', NOW() - INTERVAL '4 hours', '检测到TrickBot僵尸网络通信', 1, NOW() - INTERVAL '4 hours', NOW()),
('agent-042-i5j6k7l8', 44, 'aws-mail-01', '10.0.7.14', 'botnet', 'Qakbot检测规则', 'qakbot.botnet-control.com', '222.186.30.112', 456, NOW() - INTERVAL '5 days', NOW() - INTERVAL '6 hours', '检测到Qakbot僵尸网络C2通信', 0, NOW() - INTERVAL '6 hours', NOW()),
('agent-006-u1v2w3x4', 6, 'aws-app-01', '10.0.2.10', 'botnet', 'Dridex检测规则', 'dridex-loader.evil.net', '45.227.255.99', 234, NOW() - INTERVAL '2 days', NOW() - INTERVAL '3 hours', '检测到Dridex银行木马通信', 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-046-y1z2a3b4', 48, 'aws-ftp-01', '10.0.7.18', 'botnet', 'IRC僵尸网络检测规则', 'irc.botnet-army.xyz', '185.220.100.252', 678, NOW() - INTERVAL '4 days', NOW() - INTERVAL '8 hours', '检测到IRC协议僵尸网络通信', 1, NOW() - INTERVAL '8 hours', NOW()),
('agent-010-k7l8m9n0', 10, 'aws-worker-02', '10.0.2.21', 'botnet', 'Mozi僵尸网络检测规则', 'mozi-dht.p2p-botnet.xyz', '176.111.174.26', 345, NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 hours', '检测到Mozi IoT僵尸网络P2P通信', 0, NOW() - INTERVAL '5 hours', NOW()),

-- ransomware: 勒索软件
('agent-003-i9j0k1l2', 3, 'aws-api-01', '10.0.1.20', 'ransomware', 'Lockbit勒索软件检测规则', 'lockbit-payment.onion.ws', '91.121.87.18', 45, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '15 minutes', '检测到与Lockbit勒索软件支付页面通信', 0, NOW() - INTERVAL '15 minutes', NOW()),
('agent-037-o5p6q7r8', 39, 'aws-alertmanager-01', '10.0.6.14', 'ransomware', 'Linux.Encoder勒索软件检测规则', 'linux-encrypt.darkweb.xyz', '195.154.181.128', 23, NOW() - INTERVAL '4 hours', NOW() - INTERVAL '1 hour', '检测到Linux.Encoder勒索软件通信', 0, NOW() - INTERVAL '1 hour', NOW()),
('agent-040-a7b8c9d0', 42, 'aws-dns-01', '10.0.7.12', 'ransomware', 'HelloKitty勒索软件检测规则', 'hellokitty-linux.ransomware.io', '103.75.190.11', 67, NOW() - INTERVAL '6 hours', NOW() - INTERVAL '2 hours', '检测到HelloKitty Linux版勒索软件C2通信', 0, NOW() - INTERVAL '2 hours', NOW()),
('agent-041-e1f2g3h4', 43, 'aws-nfs-01', '10.0.7.13', 'ransomware', 'BlackCat勒索软件检测规则', 'alphv-support.onion.to', '185.156.73.54', 34, NOW() - INTERVAL '2 days', NOW() - INTERVAL '5 hours', '检测到BlackCat(ALPHV)勒索软件通信', 1, NOW() - INTERVAL '5 hours', NOW()),
('agent-013-w9x0y1z2', 13, 'aws-pg-01', '10.0.3.12', 'ransomware', 'Hive勒索软件检测规则', 'hive-ransom.payment.xyz', '45.155.205.233', 12, NOW() - INTERVAL '8 hours', NOW() - INTERVAL '3 hours', '检测到Hive勒索软件支付通信', 0, NOW() - INTERVAL '3 hours', NOW()),
('agent-011-o1p2q3r4', 11, 'aws-mysql-01', '10.0.3.10', 'ransomware', 'Ryuk勒索软件检测规则', 'ryuk-decryptor.darknet.xyz', '91.240.118.172', 56, NOW() - INTERVAL '1 day', NOW() - INTERVAL '4 hours', '检测到Ryuk勒索软件密钥服务器通信', 0, NOW() - INTERVAL '4 hours', NOW()),
('agent-039-w3x4y5z6', 41, 'aws-bastion-01', '10.0.7.11', 'ransomware', 'RansomEXX勒索软件检测规则', 'ransomexx-linux.leak-site.com', '23.129.64.130', 78, NOW() - INTERVAL '3 days', NOW() - INTERVAL '6 hours', '检测到RansomEXX Linux版勒索软件数据泄露站点通信', 2, NOW() - INTERVAL '6 hours', NOW()),
('agent-014-a3b4c5d6', 14, 'aws-redis-01', '10.0.3.20', 'ransomware', 'Royal勒索软件检测规则', 'royal-chat.onion.ws', '193.142.146.35', 28, NOW() - INTERVAL '5 hours', NOW() - INTERVAL '2 hours', '检测到Royal勒索软件加密通信及数据外泄', 0, NOW() - INTERVAL '2 hours', NOW());
