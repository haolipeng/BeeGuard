-- 000014: 容器核心文件监控告警表
-- 包含: alert_container_sensitive_file

CREATE TABLE IF NOT EXISTS alert_container_sensitive_file (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(256) NOT NULL,
    container_id      VARCHAR(64) NOT NULL,
    container_name    VARCHAR(256),
    image_name        VARCHAR(512),
    rule_id           VARCHAR(32) NOT NULL,
    rule_name         VARCHAR(128) NOT NULL,
    severity          VARCHAR(16) NOT NULL,
    rule_description  TEXT,
    matched_pattern   VARCHAR(512),
    action            VARCHAR(16) NOT NULL,
    file_path         VARCHAR(1024) NOT NULL,
    old_path          VARCHAR(1024),
    operator_user     VARCHAR(64),
    operator_process  VARCHAR(256),
    status            SMALLINT NOT NULL DEFAULT 0,
    alert_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_csf_agent_id ON alert_container_sensitive_file(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_csf_host_ip ON alert_container_sensitive_file(host_ip);
CREATE INDEX IF NOT EXISTS idx_alert_csf_container_id ON alert_container_sensitive_file(container_id);
CREATE INDEX IF NOT EXISTS idx_alert_csf_severity ON alert_container_sensitive_file(severity);
CREATE INDEX IF NOT EXISTS idx_alert_csf_status ON alert_container_sensitive_file(status);
CREATE INDEX IF NOT EXISTS idx_alert_csf_alert_time ON alert_container_sensitive_file(alert_time);

COMMENT ON TABLE alert_container_sensitive_file IS '容器安全-容器核心文件监控告警';
COMMENT ON COLUMN alert_container_sensitive_file.rule_id IS '命中规则ID';
COMMENT ON COLUMN alert_container_sensitive_file.severity IS '严重等级: low/medium/high';
COMMENT ON COLUMN alert_container_sensitive_file.action IS '文件操作类型: create/rename/delete';
COMMENT ON COLUMN alert_container_sensitive_file.status IS '状态: 0-待处理 1-已处理 2-已忽略';
COMMENT ON COLUMN alert_container_sensitive_file.container_id IS '容器ID(64字符hex)';
