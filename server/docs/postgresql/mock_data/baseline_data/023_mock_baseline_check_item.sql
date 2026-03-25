-- =====================================================
-- 模拟数据: baseline_check_item (基线检查项表)
-- 数据量: 45条
-- 说明: AWS ap-southeast-1 (Singapore) 区域 EC2 实例
--       各基线模板对应的安全检查项
-- check_rules: JSON格式的检查规则
--   type: file_line_check / file_permission_check / command_check / service_status_check / sysctl_check
--   param: 参数列表（如文件路径、命令、服务名等）
--   filter: 正则表达式，提取检查值
--   result: 期望结果，格式 $(op)value，op支持 ==, !=, <=, >=, <, >
-- =====================================================

INSERT INTO baseline_check_item (id, template_id, item_name, category, risk_level, check_rules, fix_suggestion, fix_script, created_at, updated_at) VALUES

-- ==========================================
-- Amazon Linux 2 安全基线 (baseline_id=1) 15项
-- ==========================================

-- 密码策略
(1,  1, '密码最小长度检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MIN_LEN\\s*\\t*(\\d+)","result":"$(>=)8"}]}',
 '设置密码最小长度不低于8位', 'sed -i "s/^PASS_MIN_LEN.*/PASS_MIN_LEN    8/" /etc/login.defs',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(2,  1, '密码过期时间检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MAX_DAYS\\s*\\t*(\\d+)","result":"$(<=)90"}]}',
 '设置密码最长使用天数为90天', 'sed -i "s/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/" /etc/login.defs',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(3,  1, '密码复杂度检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/pam.d/system-auth"],"filter":"pam_pwquality.*minclass=(\\d+)","result":"$(>=)3"}]}',
 '配置密码复杂度至少包含3类字符', 'authconfig --passminclass=3 --update',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- SSH安全
(4,  1, 'SSH禁止root远程登录', 'SSH安全', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*PermitRootLogin\\s+(\\S+)","result":"$(==)no"}]}',
 '禁止root用户直接通过SSH远程登录', 'sed -i "s/^#*PermitRootLogin.*/PermitRootLogin no/" /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(5,  1, 'SSH协议版本检查', 'SSH安全', 'high',
 '{"rules":[{"type":"command_check","param":["ssh -V 2>&1"],"filter":"OpenSSH_(\\d+)","result":"$(>=)7"}]}',
 '确保使用SSH Protocol 2（OpenSSH 7+默认仅支持v2）', NULL,
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(6,  1, 'SSH空闲超时设置', 'SSH安全', 'medium',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*ClientAliveInterval\\s+(\\d+)","result":"$(<=)300"}]}',
 '设置SSH空闲超时时间不超过300秒', 'echo "ClientAliveInterval 300" >> /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(7,  1, 'SSH最大认证尝试次数', 'SSH安全', 'medium',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*MaxAuthTries\\s+(\\d+)","result":"$(<=)4"}]}',
 '设置SSH最大认证尝试次数不超过4次', 'sed -i "s/^#*MaxAuthTries.*/MaxAuthTries 4/" /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- 防火墙
(8,  1, '防火墙服务状态检查', '防火墙', 'high',
 '{"rules":[{"type":"service_status_check","param":["iptables"],"filter":"","result":"$(==)active"}]}',
 '确保iptables服务已启动', 'systemctl enable --now iptables',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- 文件权限
(9,  1, '/etc/passwd文件权限检查', '文件权限', 'high',
 '{"rules":[{"type":"file_permission_check","param":["/etc/passwd"],"filter":"","result":"$(==)644"}]}',
 '设置/etc/passwd文件权限为644', 'chmod 644 /etc/passwd',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(10, 1, '/etc/shadow文件权限检查', '文件权限', 'high',
 '{"rules":[{"type":"file_permission_check","param":["/etc/shadow"],"filter":"","result":"$(==)000"}]}',
 '设置/etc/shadow文件权限为000', 'chmod 000 /etc/shadow',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(11, 1, '/etc/gshadow文件权限检查', '文件权限', 'medium',
 '{"rules":[{"type":"file_permission_check","param":["/etc/gshadow"],"filter":"","result":"$(==)000"}]}',
 '设置/etc/gshadow文件权限为000', 'chmod 000 /etc/gshadow',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- 日志审计
(12, 1, '审计服务状态检查', '日志审计', 'high',
 '{"rules":[{"type":"service_status_check","param":["auditd"],"filter":"","result":"$(==)active"}]}',
 '确保auditd服务已启动', 'systemctl enable --now auditd',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

(13, 1, 'rsyslog服务状态检查', '日志审计', 'medium',
 '{"rules":[{"type":"service_status_check","param":["rsyslog"],"filter":"","result":"$(==)active"}]}',
 '确保rsyslog服务已启动', 'systemctl enable --now rsyslog',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- 账户安全
(14, 1, '空密码账户检查', '账户安全', 'high',
 '{"rules":[{"type":"command_check","param":["awk -F: ''($2 == \"\") {print $1}'' /etc/shadow | wc -l"],"filter":"(\\d+)","result":"$(==)0"}]}',
 '确保不存在空密码账户', NULL,
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- 内核参数
(15, 1, 'IP转发禁用检查', '内核参数', 'medium',
 '{"rules":[{"type":"sysctl_check","param":["net.ipv4.ip_forward"],"filter":"","result":"$(==)0"}]}',
 '禁用IP转发功能', 'sysctl -w net.ipv4.ip_forward=0 && echo "net.ipv4.ip_forward = 0" >> /etc/sysctl.conf',
 NOW() - INTERVAL '90 days', NOW() - INTERVAL '90 days'),

-- ==========================================
-- Ubuntu 22.04 安全基线 (baseline_id=2) 14项
-- ==========================================
(16, 2, '密码最小长度检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MIN_LEN\\s*\\t*(\\d+)","result":"$(>=)8"}]}',
 '设置密码最小长度不低于8位', 'sed -i "s/^PASS_MIN_LEN.*/PASS_MIN_LEN    8/" /etc/login.defs',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(17, 2, '密码过期时间检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MAX_DAYS\\s*\\t*(\\d+)","result":"$(<=)90"}]}',
 '设置密码最长使用天数为90天', 'sed -i "s/^PASS_MAX_DAYS.*/PASS_MAX_DAYS   90/" /etc/login.defs',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(18, 2, 'SSH禁止root远程登录', 'SSH安全', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*PermitRootLogin\\s+(\\S+)","result":"$(==)no"}]}',
 '禁止root用户直接通过SSH远程登录', 'sed -i "s/^#*PermitRootLogin.*/PermitRootLogin no/" /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(19, 2, 'SSH空闲超时设置', 'SSH安全', 'medium',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*ClientAliveInterval\\s+(\\d+)","result":"$(<=)300"}]}',
 '设置SSH空闲超时时间不超过300秒', 'echo "ClientAliveInterval 300" >> /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(20, 2, 'UFW防火墙状态检查', '防火墙', 'high',
 '{"rules":[{"type":"command_check","param":["ufw status"],"filter":"Status:\\s+(\\S+)","result":"$(==)active"}]}',
 '确保UFW防火墙已启用', 'ufw enable',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(21, 2, '/etc/passwd文件权限检查', '文件权限', 'high',
 '{"rules":[{"type":"file_permission_check","param":["/etc/passwd"],"filter":"","result":"$(==)644"}]}',
 '设置/etc/passwd文件权限为644', 'chmod 644 /etc/passwd',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(22, 2, '/etc/shadow文件权限检查', '文件权限', 'high',
 '{"rules":[{"type":"file_permission_check","param":["/etc/shadow"],"filter":"","result":"$(==)640"}]}',
 '设置/etc/shadow文件权限为640', 'chmod 640 /etc/shadow',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(23, 2, '审计服务状态检查', '日志审计', 'high',
 '{"rules":[{"type":"service_status_check","param":["auditd"],"filter":"","result":"$(==)active"}]}',
 '确保auditd服务已启动', 'apt install auditd -y && systemctl enable --now auditd',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(24, 2, '空密码账户检查', '账户安全', 'high',
 '{"rules":[{"type":"command_check","param":["awk -F: ''($2 == \"\") {print $1}'' /etc/shadow | wc -l"],"filter":"(\\d+)","result":"$(==)0"}]}',
 '确保不存在空密码账户', NULL,
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(25, 2, 'IP转发禁用检查', '内核参数', 'medium',
 '{"rules":[{"type":"sysctl_check","param":["net.ipv4.ip_forward"],"filter":"","result":"$(==)0"}]}',
 '禁用IP转发功能', 'sysctl -w net.ipv4.ip_forward=0',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(26, 2, 'SYN Cookie启用检查', '内核参数', 'medium',
 '{"rules":[{"type":"sysctl_check","param":["net.ipv4.tcp_syncookies"],"filter":"","result":"$(==)1"}]}',
 '启用SYN Cookie防护', 'sysctl -w net.ipv4.tcp_syncookies=1',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(27, 2, 'core dump限制检查', '账户安全', 'medium',
 '{"rules":[{"type":"file_line_check","param":["/etc/security/limits.conf"],"filter":"\\*\\s+hard\\s+core\\s+(\\d+)","result":"$(==)0"}]}',
 '禁止生成core dump文件', 'echo "* hard core 0" >> /etc/security/limits.conf',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(28, 2, 'SSH最大认证尝试次数', 'SSH安全', 'medium',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*MaxAuthTries\\s+(\\d+)","result":"$(<=)4"}]}',
 '设置SSH最大认证尝试次数不超过4次', 'sed -i "s/^#*MaxAuthTries.*/MaxAuthTries 4/" /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

(29, 2, 'Cron守护进程启用检查', '服务安全', 'low',
 '{"rules":[{"type":"service_status_check","param":["cron"],"filter":"","result":"$(==)active"}]}',
 '确保Cron守护进程已启用', 'systemctl enable --now cron',
 NOW() - INTERVAL '85 days', NOW() - INTERVAL '85 days'),

-- ==========================================
-- Ubuntu 20.04 安全基线 (baseline_id=3) 5项
-- ==========================================
(30, 3, '密码最小长度检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MIN_LEN\\s*\\t*(\\d+)","result":"$(>=)8"}]}',
 '设置密码最小长度不低于8位', 'sed -i "s/^PASS_MIN_LEN.*/PASS_MIN_LEN    8/" /etc/login.defs',
 NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days'),

(31, 3, 'SSH禁止root远程登录', 'SSH安全', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*PermitRootLogin\\s+(\\S+)","result":"$(==)no"}]}',
 '禁止root用户直接通过SSH远程登录', 'sed -i "s/^#*PermitRootLogin.*/PermitRootLogin no/" /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days'),

(32, 3, '/etc/passwd文件权限检查', '文件权限', 'high',
 '{"rules":[{"type":"file_permission_check","param":["/etc/passwd"],"filter":"","result":"$(==)644"}]}',
 '设置/etc/passwd文件权限为644', 'chmod 644 /etc/passwd',
 NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days'),

(33, 3, '审计服务状态检查', '日志审计', 'high',
 '{"rules":[{"type":"service_status_check","param":["auditd"],"filter":"","result":"$(==)active"}]}',
 '确保auditd服务已启动', 'apt install auditd -y && systemctl enable --now auditd',
 NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days'),

(34, 3, '空密码账户检查', '账户安全', 'high',
 '{"rules":[{"type":"command_check","param":["awk -F: ''($2 == \"\") {print $1}'' /etc/shadow | wc -l"],"filter":"(\\d+)","result":"$(==)0"}]}',
 '确保不存在空密码账户', NULL,
 NOW() - INTERVAL '70 days', NOW() - INTERVAL '70 days'),

-- ==========================================
-- Amazon Linux 2023 安全基线 (baseline_id=4) 3项
-- ==========================================
(35, 4, '密码最小长度检查', '密码策略', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/login.defs"],"filter":"\\s*\\t*PASS_MIN_LEN\\s*\\t*(\\d+)","result":"$(>=)8"}]}',
 '设置密码最小长度不低于8位', 'sed -i "s/^PASS_MIN_LEN.*/PASS_MIN_LEN    8/" /etc/login.defs',
 NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days'),

(36, 4, 'SSH禁止root远程登录', 'SSH安全', 'high',
 '{"rules":[{"type":"file_line_check","param":["/etc/ssh/sshd_config"],"filter":"^\\s*PermitRootLogin\\s+(\\S+)","result":"$(==)no"}]}',
 '禁止root用户直接通过SSH远程登录', 'sed -i "s/^#*PermitRootLogin.*/PermitRootLogin no/" /etc/ssh/sshd_config && systemctl restart sshd',
 NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days'),

(37, 4, '防火墙服务状态检查', '防火墙', 'high',
 '{"rules":[{"type":"service_status_check","param":["iptables"],"filter":"","result":"$(==)active"}]}',
 '确保iptables服务已启动', 'systemctl enable --now iptables',
 NOW() - INTERVAL '60 days', NOW() - INTERVAL '60 days'),

-- ==========================================
-- MySQL/RDS 安全基线 (baseline_id=6) 8项
-- ==========================================
(38, 6, 'MySQL远程root登录检查', '数据库安全', 'high',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SELECT COUNT(*) FROM mysql.user WHERE user=''root'' AND host NOT IN (''localhost'',''127.0.0.1'',''::1'')\""],"filter":"(\\d+)","result":"$(==)0"}]}',
 '禁止root用户远程登录MySQL', 'mysql -e "DELETE FROM mysql.user WHERE user=''root'' AND host NOT IN (''localhost'',''127.0.0.1'',''::1''); FLUSH PRIVILEGES;"',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(39, 6, 'MySQL空密码账户检查', '数据库安全', 'high',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SELECT COUNT(*) FROM mysql.user WHERE authentication_string='''' OR authentication_string IS NULL\""],"filter":"(\\d+)","result":"$(==)0"}]}',
 '确保不存在空密码的数据库账户', NULL,
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(40, 6, 'MySQL审计日志启用检查', '数据库安全', 'medium',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SHOW VARIABLES LIKE ''general_log''\""],"filter":"general_log\\s+(\\S+)","result":"$(==)ON"}]}',
 '启用MySQL general_log审计日志', 'mysql -e "SET GLOBAL general_log = ON;"',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(41, 6, 'MySQL binlog启用检查', '数据库安全', 'medium',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SHOW VARIABLES LIKE ''log_bin''\""],"filter":"log_bin\\s+(\\S+)","result":"$(==)ON"}]}',
 '启用MySQL binlog日志', '在my.cnf中配置 log-bin=mysql-bin 并重启MySQL',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(42, 6, 'MySQL最大连接数检查', '数据库安全', 'low',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SHOW VARIABLES LIKE ''max_connections''\""],"filter":"max_connections\\s+(\\d+)","result":"$(>=)500"}]}',
 '设置合理的最大连接数', 'mysql -e "SET GLOBAL max_connections = 500;"',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(43, 6, 'MySQL test数据库检查', '数据库安全', 'medium',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SHOW DATABASES LIKE ''test''\" | wc -l"],"filter":"(\\d+)","result":"$(==)0"}]}',
 '删除默认test数据库', 'mysql -e "DROP DATABASE IF EXISTS test;"',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(44, 6, 'MySQL SSL连接检查', '数据库安全', 'medium',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SHOW VARIABLES LIKE ''have_ssl''\""],"filter":"have_ssl\\s+(\\S+)","result":"$(==)YES"}]}',
 '启用MySQL SSL加密连接', '配置SSL证书并在my.cnf中启用SSL',
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),

(45, 6, 'MySQL错误日志启用检查', '数据库安全', 'low',
 '{"rules":[{"type":"command_check","param":["mysql -e \"SHOW VARIABLES LIKE ''log_error''\""],"filter":"log_error\\s+(\\S+)","result":"$(!=)"}]}',
 '确保错误日志路径已配置', NULL,
 NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days');

-- 重置序列
SELECT setval('baseline_check_item_id_seq', 45);
