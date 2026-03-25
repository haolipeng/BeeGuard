-- =====================================================
-- SOC 入侵检测告警数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- 说明: 合并自 migrations/004-011 的入侵检测相关表
-- =====================================================

-- =====================================================
-- 1. 暴力破解告警表 (alert_brute_force)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_brute_force (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(45) NOT NULL,

    source_ip         VARCHAR(45) NOT NULL,                           -- 攻击来源IP
    source_location   VARCHAR(128),                                   -- 来源地理位置(国家)
    attack_type       VARCHAR(32) NOT NULL,                           -- 攻击类型: ssh/ftp/rdp/mysql/redis/web_login
    target_ip         VARCHAR(45) NOT NULL,                           -- 目标IP
    target_port       INT,                                            -- 目标端口
    username          VARCHAR(64) NOT NULL,                           -- 被尝试的用户名
    attempt_count     INT NOT NULL,                                   -- 尝试次数
    attack_time       TIMESTAMP NOT NULL,                             -- 攻击时间(最近一次)
    first_attack_time TIMESTAMP,                                      -- 首次攻击时间

    status            SMALLINT NOT NULL DEFAULT 0,                    -- 0-待处理 1-已处理 2-已忽略
    is_blocked        SMALLINT DEFAULT 0,                             -- 是否已封禁: 0-否 1-是
    process_time      TIMESTAMP,                                      -- 处理时间
    processor         VARCHAR(64),                                    -- 处理人
    remark            VARCHAR(512),                                   -- 备注
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- attack_type 枚举值:
-- ssh: SSH密码暴力破解
-- ftp: FTP暴力破解
-- mysql: MySQL暴力破解
-- redis: Redis未授权访问
-- web_login: Web登录暴力破解

CREATE INDEX IF NOT EXISTS idx_alert_bf_agent_id ON alert_brute_force(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_bf_source_ip ON alert_brute_force(source_ip);
CREATE INDEX IF NOT EXISTS idx_alert_bf_attack_type ON alert_brute_force(attack_type);
CREATE INDEX IF NOT EXISTS idx_alert_bf_status ON alert_brute_force(status);
CREATE INDEX IF NOT EXISTS idx_alert_bf_attack_time ON alert_brute_force(attack_time);

COMMENT ON TABLE alert_brute_force IS '入侵检测-暴力破解告警';
COMMENT ON COLUMN alert_brute_force.attack_type IS '攻击类型: ssh/ftp/rdp/mysql/redis/web_login';
COMMENT ON COLUMN alert_brute_force.status IS '状态: 0-待处理 1-已处理 2-已忽略';


-- =====================================================
-- 2. 高危命令告警表 (alert_dangerous_command)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_dangerous_command (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(45) NOT NULL,

    command           TEXT NOT NULL,                                  -- 执行的命令内容
    command_type      VARCHAR(32) NOT NULL,                           -- 命令类型(见枚举)
    "user"            VARCHAR(64) NOT NULL,                           -- 执行用户
    privilege_level   VARCHAR(32) NOT NULL,                           -- 权限级别

    status            SMALLINT NOT NULL DEFAULT 0,                    -- 0-待处理 1-已处理 2-已忽略
    alert_time        TIMESTAMP NOT NULL,                             -- 告警时间
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- command_type 枚举值:
-- file_delete: 文件删除
-- privilege_escalation: 权限提升
-- permission_modify: 权限修改
-- filesystem_operation: 文件系统操作
-- network_scan: 网络扫描
-- data_exfiltration: 数据外传
-- service_stop: 服务停止
-- log_tamper: 日志篡改

CREATE INDEX IF NOT EXISTS idx_alert_cmd_agent_id ON alert_dangerous_command(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_command_type ON alert_dangerous_command(command_type);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_status ON alert_dangerous_command(status);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_alert_time ON alert_dangerous_command(alert_time);

COMMENT ON TABLE alert_dangerous_command IS '入侵检测-高危命令告警';
COMMENT ON COLUMN alert_dangerous_command.command_type IS '命令类型: file_delete/privilege_escalation/permission_modify/filesystem_operation/network_scan/data_exfiltration/service_stop/log_tamper';


-- =====================================================
-- 3. 反弹Shell告警表 (alert_reverse_shell)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_reverse_shell (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    victim_ip         VARCHAR(45) NOT NULL,                           -- 受害主机IP

    command_line      TEXT NOT NULL,                                  -- 反弹Shell命令行
    shell_type        VARCHAR(32),                                    -- Shell类型
    target_host       VARCHAR(45) NOT NULL,                           -- 目标主机(攻击者IP)
    target_port       INT NOT NULL,                                   -- 目标端口

    status            SMALLINT NOT NULL DEFAULT 0,                    -- 0-待处理 1-已处理 2-已忽略
    event_time        TIMESTAMP NOT NULL,                             -- 事件时间
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- shell_type 枚举值:
-- bash: Bash Shell
-- python: Python
-- nc: Netcat
-- perl: Perl
-- php: PHP
-- ruby: Ruby
-- powershell: PowerShell

CREATE INDEX IF NOT EXISTS idx_alert_rs_agent_id ON alert_reverse_shell(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_rs_shell_type ON alert_reverse_shell(shell_type);
CREATE INDEX IF NOT EXISTS idx_alert_rs_target_host ON alert_reverse_shell(target_host);
CREATE INDEX IF NOT EXISTS idx_alert_rs_status ON alert_reverse_shell(status);
CREATE INDEX IF NOT EXISTS idx_alert_rs_event_time ON alert_reverse_shell(event_time);

COMMENT ON TABLE alert_reverse_shell IS '入侵检测-反弹Shell告警';
COMMENT ON COLUMN alert_reverse_shell.shell_type IS 'Shell类型: bash/python/nc/perl/php/ruby/powershell';


-- =====================================================
-- 4. 本地提权告警表 (alert_privilege_escalation)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_privilege_escalation (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,

    escalated_user        VARCHAR(64) NOT NULL,                       -- 提权后用户
    parent_process        VARCHAR(256) NOT NULL,                      -- 父进程名称
    parent_process_user   VARCHAR(64) NOT NULL,                       -- 父进程所属用户
    process_id            INT,                                        -- 进程ID
    process_path          VARCHAR(512),                               -- 进程路径

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    discover_time         TIMESTAMP NOT NULL,                         -- 发现时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_pe_agent_id ON alert_privilege_escalation(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_pe_escalated_user ON alert_privilege_escalation(escalated_user);
CREATE INDEX IF NOT EXISTS idx_alert_pe_status ON alert_privilege_escalation(status);
CREATE INDEX IF NOT EXISTS idx_alert_pe_discover_time ON alert_privilege_escalation(discover_time);

COMMENT ON TABLE alert_privilege_escalation IS '入侵检测-本地提权告警';
COMMENT ON COLUMN alert_privilege_escalation.escalated_user IS '提权后的用户(通常为root)';


-- =====================================================
-- 5. 异常登录告警表 (alert_abnormal_login)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_abnormal_login (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,

    source_ip             VARCHAR(45) NOT NULL,                       -- 来源IP
    source_location       VARCHAR(128),                               -- 来源地理位置
    source_country        VARCHAR(64),                                -- 来源国家
    source_city           VARCHAR(64),                                -- 来源城市
    login_user            VARCHAR(64) NOT NULL,                       -- 登录用户名
    login_time            TIMESTAMP NOT NULL,                         -- 登录时间
    risk_level            VARCHAR(16) NOT NULL,                       -- 危险等级: low/medium/high
    abnormal_type         VARCHAR(32),                                -- 异常类型

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    is_whitelist          SMALLINT DEFAULT 0,                         -- 是否白名单: 0-否 1-是
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- abnormal_type 枚举值:
-- abnormal_location: 异常地域
-- abnormal_time: 异常时间
-- abnormal_user: 异常用户

CREATE INDEX IF NOT EXISTS idx_alert_al_agent_id ON alert_abnormal_login(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_al_source_ip ON alert_abnormal_login(source_ip);
CREATE INDEX IF NOT EXISTS idx_alert_al_login_user ON alert_abnormal_login(login_user);
CREATE INDEX IF NOT EXISTS idx_alert_al_abnormal_type ON alert_abnormal_login(abnormal_type);
CREATE INDEX IF NOT EXISTS idx_alert_al_status ON alert_abnormal_login(status);
CREATE INDEX IF NOT EXISTS idx_alert_al_login_time ON alert_abnormal_login(login_time);

COMMENT ON TABLE alert_abnormal_login IS '入侵检测-异常登录告警';
COMMENT ON COLUMN alert_abnormal_login.abnormal_type IS '异常类型: abnormal_location/abnormal_time/abnormal_user';
COMMENT ON COLUMN alert_abnormal_login.risk_level IS '危险等级: low/medium/high';


-- =====================================================
-- 6. 恶意请求告警表 (alert_malicious_request)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_malicious_request (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,

    policy_type           VARCHAR(32) NOT NULL,                       -- 命中策略类型
    policy_name           VARCHAR(128) NOT NULL,                      -- 命中策略名称
    malicious_domain      VARCHAR(256) NOT NULL,                      -- 恶意请求域名
    malicious_ip          VARCHAR(45),                                -- 恶意请求IP
    request_count         INT NOT NULL,                               -- 请求次数
    first_request_time    TIMESTAMP,                                  -- 首次请求时间
    last_request_time     TIMESTAMP,                                  -- 最近请求时间
    risk_description      TEXT,                                       -- 危害描述

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- policy_type 枚举值:
-- mining: 挖矿
-- c2: C2通信
-- phishing: 钓鱼网站
-- botnet: 僵尸网络
-- ransomware: 勒索软件

CREATE INDEX IF NOT EXISTS idx_alert_mr_agent_id ON alert_malicious_request(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_mr_policy_type ON alert_malicious_request(policy_type);
CREATE INDEX IF NOT EXISTS idx_alert_mr_malicious_domain ON alert_malicious_request(malicious_domain);
CREATE INDEX IF NOT EXISTS idx_alert_mr_status ON alert_malicious_request(status);
CREATE INDEX IF NOT EXISTS idx_alert_mr_last_request_time ON alert_malicious_request(last_request_time);

COMMENT ON TABLE alert_malicious_request IS '入侵检测-恶意请求告警';
COMMENT ON COLUMN alert_malicious_request.policy_type IS '策略类型: mining/c2/phishing/botnet/ransomware';


-- =====================================================
-- 7. 网络攻击告警表 (alert_network_attack)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_network_attack (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,                       -- 被攻击主机IP

    target_port           INT NOT NULL,                               -- 目标端口
    attacker_ip           VARCHAR(45) NOT NULL,                       -- 攻击来源IP
    attacker_location     VARCHAR(128),                               -- 攻击来源地理位置
    attacker_country      VARCHAR(64),                                -- 攻击来源国家
    vulnerability_name    VARCHAR(256) NOT NULL,                      -- 漏洞名称
    vulnerability_id      VARCHAR(64),                                -- 漏洞编号(CVE等)
    attack_status         VARCHAR(32) NOT NULL,                       -- 攻击状态
    attack_count          INT NOT NULL,                               -- 攻击次数
    first_attack_time     TIMESTAMP,                                  -- 首次攻击时间
    last_attack_time      TIMESTAMP NOT NULL,                         -- 最近攻击时间
    attack_payload        TEXT,                                       -- 攻击载荷

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_na_agent_id ON alert_network_attack(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_na_attacker_ip ON alert_network_attack(attacker_ip);
CREATE INDEX IF NOT EXISTS idx_alert_na_vulnerability_id ON alert_network_attack(vulnerability_id);
CREATE INDEX IF NOT EXISTS idx_alert_na_attack_status ON alert_network_attack(attack_status);
CREATE INDEX IF NOT EXISTS idx_alert_na_status ON alert_network_attack(status);
CREATE INDEX IF NOT EXISTS idx_alert_na_last_attack_time ON alert_network_attack(last_attack_time);

COMMENT ON TABLE alert_network_attack IS '入侵检测-网络攻击告警';
COMMENT ON COLUMN alert_network_attack.vulnerability_id IS '漏洞编号(如CVE-2021-44228)';


-- =====================================================
-- 8. 文件查杀告警表 (alert_malware_scan)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_malware_scan (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_ip               VARCHAR(45) NOT NULL,
    host_name             VARCHAR(128) NOT NULL,

    threat_type           VARCHAR(64) NOT NULL,                       -- 威胁类型
    file_name             VARCHAR(256) NOT NULL,                      -- 文件名
    file_path             VARCHAR(512) NOT NULL,                      -- 文件路径
    file_size             BIGINT,                                     -- 文件大小(字节)
    file_md5              VARCHAR(32),                                -- 文件MD5哈希
    file_sha256           VARCHAR(128),                               -- 文件SHA256哈希
    detection_engine      VARCHAR(64),                                -- 检测引擎
    malware_family        VARCHAR(64),                                -- 恶意软件家族
    is_quarantined        SMALLINT DEFAULT 0,                         -- 是否已隔离: 0-否 1-是
    is_deleted            SMALLINT DEFAULT 0,                         -- 是否已删除: 0-否 1-是

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    scan_time             TIMESTAMP NOT NULL,                         -- 扫描时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- threat_type 枚举值:
-- virus: 病毒程序
-- trojan: 木马程序
-- webshell: Webshell
-- backdoor: 后门程序
-- ransomware: 勒索软件
-- miner: 挖矿程序
-- rootkit: Rootkit

CREATE INDEX IF NOT EXISTS idx_alert_ms_agent_id ON alert_malware_scan(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_ms_threat_type ON alert_malware_scan(threat_type);
CREATE INDEX IF NOT EXISTS idx_alert_ms_file_md5 ON alert_malware_scan(file_md5);
CREATE INDEX IF NOT EXISTS idx_alert_ms_malware_family ON alert_malware_scan(malware_family);
CREATE INDEX IF NOT EXISTS idx_alert_ms_status ON alert_malware_scan(status);
CREATE INDEX IF NOT EXISTS idx_alert_ms_scan_time ON alert_malware_scan(scan_time);

COMMENT ON TABLE alert_malware_scan IS '入侵检测-文件查杀告警';
COMMENT ON COLUMN alert_malware_scan.threat_type IS '威胁类型: virus/trojan/webshell/backdoor/ransomware/miner/rootkit';


-- =====================================================
-- 9. 核心文件监控告警表 (alert_file_integrity)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_file_integrity (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,

    rule_type             VARCHAR(32) NOT NULL,                       -- 规则类型
    rule_name             VARCHAR(128) NOT NULL,                      -- 命中规则名称
    rule_id               BIGINT,                                     -- 关联规则ID
    threat_level          VARCHAR(16) NOT NULL,                       -- 威胁等级: low/medium/high
    threat_action         VARCHAR(32) NOT NULL,                       -- 威胁行为: add/modify/delete
    file_path             VARCHAR(512) NOT NULL,                      -- 文件路径
    file_name             VARCHAR(256),                               -- 文件名
    old_content_hash      VARCHAR(64),                                -- 原内容哈希
    new_content_hash      VARCHAR(64),                                -- 新内容哈希
    change_detail         TEXT,                                       -- 变更详情
    operator_user         VARCHAR(64),                                -- 操作用户
    operator_process      VARCHAR(256),                               -- 操作进程
    alert_description     TEXT,                                       -- 告警描述

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    alert_time            TIMESTAMP NOT NULL,                         -- 告警时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_fi_agent_id ON alert_file_integrity(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_fi_rule_type ON alert_file_integrity(rule_type);
CREATE INDEX IF NOT EXISTS idx_alert_fi_threat_level ON alert_file_integrity(threat_level);
CREATE INDEX IF NOT EXISTS idx_alert_fi_file_path ON alert_file_integrity(file_path);
CREATE INDEX IF NOT EXISTS idx_alert_fi_status ON alert_file_integrity(status);
CREATE INDEX IF NOT EXISTS idx_alert_fi_alert_time ON alert_file_integrity(alert_time);

COMMENT ON TABLE alert_file_integrity IS '入侵检测-核心文件监控告警';
COMMENT ON COLUMN alert_file_integrity.threat_level IS '威胁等级: low/medium/high';
COMMENT ON COLUMN alert_file_integrity.threat_action IS '威胁行为: add/modify/delete';


-- =====================================================
-- 10. 告警处理记录表 (alert_process_log)
-- =====================================================
CREATE TABLE IF NOT EXISTS alert_process_log (
    id                    BIGSERIAL PRIMARY KEY,
    alert_type            VARCHAR(32) NOT NULL,                       -- 告警类型
    alert_id              BIGINT NOT NULL,                            -- 关联的告警ID
    old_status            SMALLINT,                                   -- 变更前状态
    new_status            SMALLINT NOT NULL,                          -- 变更后状态
    processor             VARCHAR(64) NOT NULL,                       -- 处理人
    remark                VARCHAR(512),                               -- 处理备注
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- alert_type 枚举值:
-- dangerous_command: 高危命令
-- reverse_shell: 反弹Shell
-- privilege_escalation: 本地提权
-- abnormal_login: 异常登录
-- brute_force: 密码破解
-- malicious_request: 恶意请求
-- network_attack: 网络攻击
-- malware_scan: 文件查杀
-- file_integrity: 核心文件监控

-- status 枚举值:
-- 0: 待处理
-- 1: 已处理
-- 2: 已忽略

CREATE INDEX IF NOT EXISTS idx_alert_pl_alert_type ON alert_process_log(alert_type);
CREATE INDEX IF NOT EXISTS idx_alert_pl_alert_id ON alert_process_log(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_pl_processor ON alert_process_log(processor);
CREATE INDEX IF NOT EXISTS idx_alert_pl_created_at ON alert_process_log(created_at);

COMMENT ON TABLE alert_process_log IS '入侵检测-告警处理记录';
COMMENT ON COLUMN alert_process_log.alert_type IS '告警类型: dangerous_command/reverse_shell/privilege_escalation/abnormal_login/brute_force/malicious_request/network_attack/malware_scan/file_integrity';


-- =====================================================
-- 初始化完成
-- =====================================================
