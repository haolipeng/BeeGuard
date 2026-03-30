-- 白名单规则表（10 张表，每类告警一张）
-- 统一结构：规则名、描述、作用范围、匹配条件(JSONB)、启用状态、命中统计

-- 1. 高危命令白名单
CREATE TABLE IF NOT EXISTS whitelist_dangerous_command (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,       -- 0=全局, 1=指定Agent
    agent_ids   TEXT,                               -- scope=1 时的 agent_id JSON 数组
    conditions  JSONB NOT NULL,                     -- 匹配条件
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_dangerous_command_enabled ON whitelist_dangerous_command(enabled);

-- 2. 反弹Shell白名单
CREATE TABLE IF NOT EXISTS whitelist_reverse_shell (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_reverse_shell_enabled ON whitelist_reverse_shell(enabled);

-- 3. 本地提权白名单
CREATE TABLE IF NOT EXISTS whitelist_privilege_escalation (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_privilege_escalation_enabled ON whitelist_privilege_escalation(enabled);

-- 4. 异常登录白名单
CREATE TABLE IF NOT EXISTS whitelist_abnormal_login (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_abnormal_login_enabled ON whitelist_abnormal_login(enabled);

-- 5. 暴力破解白名单
CREATE TABLE IF NOT EXISTS whitelist_brute_force (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_brute_force_enabled ON whitelist_brute_force(enabled);

-- 6. 恶意请求白名单
CREATE TABLE IF NOT EXISTS whitelist_malicious_request (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_malicious_request_enabled ON whitelist_malicious_request(enabled);

-- 7. 网络攻击白名单
CREATE TABLE IF NOT EXISTS whitelist_network_attack (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_network_attack_enabled ON whitelist_network_attack(enabled);

-- 8. 恶意文件白名单
CREATE TABLE IF NOT EXISTS whitelist_malware_scan (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_malware_scan_enabled ON whitelist_malware_scan(enabled);

-- 9. 文件完整性白名单
CREATE TABLE IF NOT EXISTS whitelist_fileguard (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_fileguard_enabled ON whitelist_fileguard(enabled);

-- 10. 容器告警白名单（合并容器高危命令/反弹Shell/敏感文件）
CREATE TABLE IF NOT EXISTS whitelist_container_alert (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(512),
    scope       SMALLINT NOT NULL DEFAULT 0,
    agent_ids   TEXT,
    conditions  JSONB NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT true,
    hit_count   BIGINT NOT NULL DEFAULT 0,
    created_by  VARCHAR(64),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_wl_container_alert_enabled ON whitelist_container_alert(enabled);

-- 为所有告警表添加白名单命中标记字段
ALTER TABLE alert_dangerous_command ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_dangerous_command ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_reverse_shell ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_reverse_shell ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_privilege_escalation ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_privilege_escalation ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_abnormal_login ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_abnormal_login ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_brute_force ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_brute_force ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_malicious_request ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_malicious_request ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_network_attack ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_network_attack ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_malware_scan ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_malware_scan ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_file_integrity ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_file_integrity ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_container_dangerous_command ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_container_dangerous_command ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_container_reverse_shell ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_container_reverse_shell ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

ALTER TABLE alert_container_sensitive_file ADD COLUMN IF NOT EXISTS whitelist_hit BOOLEAN DEFAULT false;
ALTER TABLE alert_container_sensitive_file ADD COLUMN IF NOT EXISTS whitelist_rule_id BIGINT DEFAULT NULL;

-- 为白名单命中字段创建索引（用于告警列表筛选）
CREATE INDEX IF NOT EXISTS idx_alert_dc_wl_hit ON alert_dangerous_command(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_rs_wl_hit ON alert_reverse_shell(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_pe_wl_hit ON alert_privilege_escalation(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_al_wl_hit ON alert_abnormal_login(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_bf_wl_hit ON alert_brute_force(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_mr_wl_hit ON alert_malicious_request(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_na_wl_hit ON alert_network_attack(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_ms_wl_hit ON alert_malware_scan(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_fi_wl_hit ON alert_file_integrity(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_cdc_wl_hit ON alert_container_dangerous_command(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_crs_wl_hit ON alert_container_reverse_shell(whitelist_hit);
CREATE INDEX IF NOT EXISTS idx_alert_csf_wl_hit ON alert_container_sensitive_file(whitelist_hit);
