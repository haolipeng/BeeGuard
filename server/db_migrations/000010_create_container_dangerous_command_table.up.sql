-- 000012: 容器高危命令告警表
-- 包含: alert_container_dangerous_command

CREATE TABLE IF NOT EXISTS alert_container_dangerous_command (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(256) NOT NULL,
    container_id      VARCHAR(64) NOT NULL,
    container_name    VARCHAR(256),
    image_name        VARCHAR(512),
    command           TEXT NOT NULL,
    command_type      VARCHAR(32) NOT NULL,
    "user"            VARCHAR(64) NOT NULL,
    privilege_level   VARCHAR(32) NOT NULL,
    status            SMALLINT NOT NULL DEFAULT 0,
    alert_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_ccmd_agent_id ON alert_container_dangerous_command(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_ccmd_host_ip ON alert_container_dangerous_command(host_ip);
CREATE INDEX IF NOT EXISTS idx_alert_ccmd_container_id ON alert_container_dangerous_command(container_id);
CREATE INDEX IF NOT EXISTS idx_alert_ccmd_command_type ON alert_container_dangerous_command(command_type);
CREATE INDEX IF NOT EXISTS idx_alert_ccmd_status ON alert_container_dangerous_command(status);
CREATE INDEX IF NOT EXISTS idx_alert_ccmd_alert_time ON alert_container_dangerous_command(alert_time);

COMMENT ON TABLE alert_container_dangerous_command IS '容器安全-容器高危命令告警';
COMMENT ON COLUMN alert_container_dangerous_command.command_type IS '命令类型(存储rule_id)';
COMMENT ON COLUMN alert_container_dangerous_command.status IS '状态: 0-待处理 1-已处理 2-已忽略';
COMMENT ON COLUMN alert_container_dangerous_command.container_id IS '容器ID(64字符hex)';
