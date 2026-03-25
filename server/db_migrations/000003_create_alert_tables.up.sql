-- 000003: 入侵检测告警表
-- 包含: alert_brute_force, alert_dangerous_command, alert_reverse_shell,
--       alert_privilege_escalation, alert_abnormal_login, alert_malicious_request,
--       alert_network_attack, alert_malware_scan, alert_file_integrity,
--       alert_process_log (10 表)

-- 1. 暴力破解告警表 (alert_brute_force)
CREATE TABLE IF NOT EXISTS alert_brute_force (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(256) NOT NULL,
    source_ip         VARCHAR(45) NOT NULL,
    source_location   VARCHAR(128),
    attack_type       VARCHAR(32) NOT NULL,
    target_ip         VARCHAR(45) NOT NULL,
    target_port       INT,
    username          VARCHAR(64) NOT NULL,
    attempt_count     INT NOT NULL,
    attack_time       TIMESTAMP NOT NULL,
    first_attack_time TIMESTAMP,
    status            SMALLINT NOT NULL DEFAULT 0,
    is_blocked        SMALLINT DEFAULT 0,
    process_time      TIMESTAMP,
    processor         VARCHAR(64),
    remark            VARCHAR(512),
    result            VARCHAR(16) NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_bf_agent_id ON alert_brute_force(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_bf_source_ip ON alert_brute_force(source_ip);
CREATE INDEX IF NOT EXISTS idx_alert_bf_attack_type ON alert_brute_force(attack_type);
CREATE INDEX IF NOT EXISTS idx_alert_bf_status ON alert_brute_force(status);
CREATE INDEX IF NOT EXISTS idx_alert_bf_attack_time ON alert_brute_force(attack_time);

COMMENT ON TABLE alert_brute_force IS '入侵检测-暴力破解告警';
COMMENT ON COLUMN alert_brute_force.attack_type IS '攻击类型: ssh/ftp/rdp/mysql/redis/web_login';
COMMENT ON COLUMN alert_brute_force.status IS '状态: 0-待处理 1-已处理 2-已忽略';


-- 2. 高危命令告警表 (alert_dangerous_command)
CREATE TABLE IF NOT EXISTS alert_dangerous_command (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(256) NOT NULL,
    command           TEXT NOT NULL,
    command_type      VARCHAR(32) NOT NULL,
    "user"            VARCHAR(64) NOT NULL,
    privilege_level   VARCHAR(32) NOT NULL,
    status            SMALLINT NOT NULL DEFAULT 0,
    alert_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_cmd_agent_id ON alert_dangerous_command(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_command_type ON alert_dangerous_command(command_type);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_status ON alert_dangerous_command(status);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_alert_time ON alert_dangerous_command(alert_time);

COMMENT ON TABLE alert_dangerous_command IS '入侵检测-高危命���告警';
COMMENT ON COLUMN alert_dangerous_command.command_type IS '命令类型: file_delete/privilege_escalation/permission_modify/filesystem_operation/network_scan/data_exfiltration/service_stop/log_tamper';


-- 3. 反弹Shell告警表 (alert_reverse_shell)
CREATE TABLE IF NOT EXISTS alert_reverse_shell (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    victim_ip         VARCHAR(45) NOT NULL,
    command_line      TEXT NOT NULL,
    shell_type        VARCHAR(32),
    target_host       VARCHAR(45) NOT NULL,
    target_port       INT NOT NULL,
    status            SMALLINT NOT NULL DEFAULT 0,
    event_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_rs_agent_id ON alert_reverse_shell(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_rs_shell_type ON alert_reverse_shell(shell_type);
CREATE INDEX IF NOT EXISTS idx_alert_rs_target_host ON alert_reverse_shell(target_host);
CREATE INDEX IF NOT EXISTS idx_alert_rs_status ON alert_reverse_shell(status);
CREATE INDEX IF NOT EXISTS idx_alert_rs_event_time ON alert_reverse_shell(event_time);

COMMENT ON TABLE alert_reverse_shell IS '入侵检测-反弹Shell告警';
COMMENT ON COLUMN alert_reverse_shell.shell_type IS 'Shell类型: bash/python/nc/perl/php/ruby/powershell';


-- 4. 本地提权告警表 (alert_privilege_escalation)
CREATE TABLE IF NOT EXISTS alert_privilege_escalation (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(256) NOT NULL,
    escalated_user        VARCHAR(64) NOT NULL,
    parent_process        VARCHAR(256) NOT NULL,
    parent_process_user   VARCHAR(64) NOT NULL,
    process_id            INT,
    process_path          VARCHAR(512),
    status                SMALLINT NOT NULL DEFAULT 0,
    discover_time         TIMESTAMP NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_pe_agent_id ON alert_privilege_escalation(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_pe_escalated_user ON alert_privilege_escalation(escalated_user);
CREATE INDEX IF NOT EXISTS idx_alert_pe_status ON alert_privilege_escalation(status);
CREATE INDEX IF NOT EXISTS idx_alert_pe_discover_time ON alert_privilege_escalation(discover_time);

COMMENT ON TABLE alert_privilege_escalation IS '入侵检测-本地提权告警';
COMMENT ON COLUMN alert_privilege_escalation.escalated_user IS '提权后的用户(通常为root)';


-- 5. 异常登录告警表 (alert_abnormal_login)
CREATE TABLE IF NOT EXISTS alert_abnormal_login (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(256) NOT NULL,
    source_ip             VARCHAR(45) NOT NULL,
    source_location       VARCHAR(128),
    source_country        VARCHAR(64),
    source_city           VARCHAR(64),
    login_user            VARCHAR(64) NOT NULL,
    login_time            TIMESTAMP NOT NULL,
    risk_level            VARCHAR(16) NOT NULL,
    abnormal_type         VARCHAR(32),
    status                SMALLINT NOT NULL DEFAULT 0,
    is_whitelist          SMALLINT DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_al_agent_id ON alert_abnormal_login(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_al_source_ip ON alert_abnormal_login(source_ip);
CREATE INDEX IF NOT EXISTS idx_alert_al_login_user ON alert_abnormal_login(login_user);
CREATE INDEX IF NOT EXISTS idx_alert_al_abnormal_type ON alert_abnormal_login(abnormal_type);
CREATE INDEX IF NOT EXISTS idx_alert_al_status ON alert_abnormal_login(status);
CREATE INDEX IF NOT EXISTS idx_alert_al_login_time ON alert_abnormal_login(login_time);

COMMENT ON TABLE alert_abnormal_login IS '入侵检测-异常登录告警';
COMMENT ON COLUMN alert_abnormal_login.abnormal_type IS '异常类型: abnormal_location/abnormal_time/abnormal_user';
COMMENT ON COLUMN alert_abnormal_login.risk_level IS '危险等级: low/medium/high';


-- 6. 恶意请求告警表 (alert_malicious_request)
CREATE TABLE IF NOT EXISTS alert_malicious_request (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(256) NOT NULL,
    policy_type           VARCHAR(32) NOT NULL,
    policy_name           VARCHAR(128) NOT NULL,
    malicious_domain      VARCHAR(256) NOT NULL,
    malicious_ip          VARCHAR(45),
    request_count         INT NOT NULL,
    first_request_time    TIMESTAMP,
    last_request_time     TIMESTAMP,
    risk_description      TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_mr_agent_id ON alert_malicious_request(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_mr_policy_type ON alert_malicious_request(policy_type);
CREATE INDEX IF NOT EXISTS idx_alert_mr_malicious_domain ON alert_malicious_request(malicious_domain);
CREATE INDEX IF NOT EXISTS idx_alert_mr_status ON alert_malicious_request(status);
CREATE INDEX IF NOT EXISTS idx_alert_mr_last_request_time ON alert_malicious_request(last_request_time);

COMMENT ON TABLE alert_malicious_request IS '入侵检测-恶意请求告警';
COMMENT ON COLUMN alert_malicious_request.policy_type IS '策略类型: mining/c2/phishing/botnet/ransomware';


-- 7. 网络攻击告警表 (alert_network_attack)
CREATE TABLE IF NOT EXISTS alert_network_attack (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(256) NOT NULL,
    target_port           INT NOT NULL,
    attacker_ip           VARCHAR(45) NOT NULL,
    attacker_location     VARCHAR(128),
    attacker_country      VARCHAR(64),
    vulnerability_name    VARCHAR(256) NOT NULL,
    vulnerability_id      VARCHAR(64),
    attack_status         VARCHAR(32) NOT NULL,
    attack_count          INT NOT NULL,
    first_attack_time     TIMESTAMP,
    last_attack_time      TIMESTAMP NOT NULL,
    attack_payload        TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
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


-- 8. 文件查杀告警表 (alert_malware_scan)
CREATE TABLE IF NOT EXISTS alert_malware_scan (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_ip               VARCHAR(256) NOT NULL,
    host_name             VARCHAR(128) NOT NULL,
    threat_type           VARCHAR(64) NOT NULL,
    file_name             VARCHAR(256) NOT NULL,
    file_path             VARCHAR(512) NOT NULL,
    file_size             BIGINT,
    file_md5              VARCHAR(32),
    file_sha256           VARCHAR(128),
    detection_engine      VARCHAR(64),
    malware_family        VARCHAR(64),
    is_quarantined        SMALLINT DEFAULT 0,
    is_deleted            SMALLINT DEFAULT 0,
    status                SMALLINT NOT NULL DEFAULT 0,
    scan_time             TIMESTAMP NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_ms_agent_id ON alert_malware_scan(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_ms_threat_type ON alert_malware_scan(threat_type);
CREATE INDEX IF NOT EXISTS idx_alert_ms_file_md5 ON alert_malware_scan(file_md5);
CREATE INDEX IF NOT EXISTS idx_alert_ms_malware_family ON alert_malware_scan(malware_family);
CREATE INDEX IF NOT EXISTS idx_alert_ms_status ON alert_malware_scan(status);
CREATE INDEX IF NOT EXISTS idx_alert_ms_scan_time ON alert_malware_scan(scan_time);

COMMENT ON TABLE alert_malware_scan IS '入侵检测-文件查杀告警';
COMMENT ON COLUMN alert_malware_scan.threat_type IS '威胁类型: virus/trojan/webshell/backdoor/ransomware/miner/rootkit';


-- 9. 核心文件监控告警表 (alert_file_integrity)
CREATE TABLE IF NOT EXISTS alert_file_integrity (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(256) NOT NULL,
    rule_type             VARCHAR(32) NOT NULL,
    rule_name             VARCHAR(128) NOT NULL,
    rule_id               BIGINT,
    threat_level          VARCHAR(16) NOT NULL,
    threat_action         VARCHAR(32) NOT NULL,
    file_path             VARCHAR(512) NOT NULL,
    file_name             VARCHAR(256),
    old_content_hash      VARCHAR(64),
    new_content_hash      VARCHAR(64),
    change_detail         TEXT,
    operator_user         VARCHAR(64),
    operator_process      VARCHAR(256),
    alert_description     TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    alert_time            TIMESTAMP NOT NULL,
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


-- 10. 告警处理记录表 (alert_process_log)
CREATE TABLE IF NOT EXISTS alert_process_log (
    id                    BIGSERIAL PRIMARY KEY,
    alert_type            VARCHAR(32) NOT NULL,
    alert_id              BIGINT NOT NULL,
    old_status            SMALLINT,
    new_status            SMALLINT NOT NULL,
    processor             VARCHAR(64) NOT NULL,
    remark                VARCHAR(512),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_pl_alert_type ON alert_process_log(alert_type);
CREATE INDEX IF NOT EXISTS idx_alert_pl_alert_id ON alert_process_log(alert_id);
CREATE INDEX IF NOT EXISTS idx_alert_pl_processor ON alert_process_log(processor);
CREATE INDEX IF NOT EXISTS idx_alert_pl_created_at ON alert_process_log(created_at);

COMMENT ON TABLE alert_process_log IS '入侵检测-告警处理记录';
COMMENT ON COLUMN alert_process_log.alert_type IS '告警类型: dangerous_command/reverse_shell/privilege_escalation/abnormal_login/brute_force/malicious_request/network_attack/malware_scan/file_integrity';
