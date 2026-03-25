-- =====================================================
-- 模拟数据: baseline_template (基线模板表)
-- 数据量: 8条
-- 说明: Linux操作系统和数据库安全合规基线模板
-- 环境: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
-- VPC CIDR: 10.0.0.0/16
-- baseline_type: os_security/db_security
-- os_type: linux
-- =====================================================

INSERT INTO baseline_template (id, template_name, template_type, os_type, version, item_count, description, is_enabled, created_at, updated_at) VALUES
-- Linux操作系统安全基线
(1, 'Amazon Linux 2 安全基线', 'os_security', 'linux', 'v1.2', 15, 'Amazon Linux 2 操作系统安全配置基线，涵盖密码策略、SSH安全、防火墙、文件权限、日志审计等检查项', 1, NOW() - INTERVAL '90 days', NOW() - INTERVAL '5 days'),
(2, 'Ubuntu 22.04 安全基线', 'os_security', 'linux', 'v1.1', 14, 'Ubuntu 22.04 LTS 操作系统安全配置基线，涵盖账户安全、SSH加固、内核参数、服务管理等检查项', 1, NOW() - INTERVAL '85 days', NOW() - INTERVAL '3 days'),
(3, 'Linux 通用安全基线', 'os_security', 'linux', 'v1.0', 12, 'Linux 通用操作系统安全配置基线，涵盖基础安全加固检查项，适用于各类Linux发行版', 1, NOW() - INTERVAL '70 days', NOW() - INTERVAL '10 days'),
(4, 'Amazon Linux 2023 安全基线', 'os_security', 'linux', 'v1.0', 13, 'Amazon Linux 2023 操作系统安全配置基线，涵盖密码策略、SSH安全、防火墙、日志审计等检查项', 1, NOW() - INTERVAL '60 days', NOW() - INTERVAL '8 days'),
(5, 'Ubuntu 20.04 安全基线', 'os_security', 'linux', 'v1.1', 12, 'Ubuntu 20.04 LTS 操作系统安全配置基线', 1, NOW() - INTERVAL '80 days', NOW() - INTERVAL '15 days'),

-- 数据库安全基线
(6, 'MySQL 安全基线', 'db_security', 'linux', 'v1.0', 10, 'MySQL 数据库安全配置基线，涵盖账户权限、网络访问、日志配置、数据加密等检查项', 1, NOW() - INTERVAL '45 days', NOW() - INTERVAL '7 days'),
(7, 'Redis 安全基线', 'db_security', 'linux', 'v1.0', 8, 'Redis 数据库安全配置基线，涵盖访问控制、网络绑定、持久化配置、命令禁用等检查项', 1, NOW() - INTERVAL '40 days', NOW() - INTERVAL '6 days'),
(8, 'PostgreSQL 安全基线', 'db_security', 'linux', 'v1.0', 9, 'PostgreSQL 数据库安全配置基线，涵盖认证方式、连接限制、权限管理、日志审计等检查项', 1, NOW() - INTERVAL '35 days', NOW() - INTERVAL '4 days');

-- 重置序列
SELECT setval('baseline_template_id_seq', 8);
