-- 000013: 容器反弹Shell告警表
-- 包含: alert_container_reverse_shell

CREATE TABLE IF NOT EXISTS alert_container_reverse_shell (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(256) NOT NULL,
    container_id      VARCHAR(64) NOT NULL,
    container_name    VARCHAR(256),
    image_name        VARCHAR(512),
    pid               INT NOT NULL,
    ppid              INT,
    uid               VARCHAR(16) NOT NULL,
    comm              VARCHAR(256) NOT NULL,
    exe_path          VARCHAR(512),
    args              TEXT,
    shell_type        VARCHAR(32),
    remote_ip         VARCHAR(45) NOT NULL,
    remote_port       INT NOT NULL,
    status            SMALLINT NOT NULL DEFAULT 0,
    event_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_crs_agent_id ON alert_container_reverse_shell(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_crs_host_ip ON alert_container_reverse_shell(host_ip);
CREATE INDEX IF NOT EXISTS idx_alert_crs_container_id ON alert_container_reverse_shell(container_id);
CREATE INDEX IF NOT EXISTS idx_alert_crs_remote_ip ON alert_container_reverse_shell(remote_ip);
CREATE INDEX IF NOT EXISTS idx_alert_crs_shell_type ON alert_container_reverse_shell(shell_type);
CREATE INDEX IF NOT EXISTS idx_alert_crs_status ON alert_container_reverse_shell(status);
CREATE INDEX IF NOT EXISTS idx_alert_crs_event_time ON alert_container_reverse_shell(event_time);

COMMENT ON TABLE alert_container_reverse_shell IS '容器安全-容器反弹Shell告警';
COMMENT ON COLUMN alert_container_reverse_shell.uid IS '用户UID(字符串)';
COMMENT ON COLUMN alert_container_reverse_shell.comm IS '进程名';
COMMENT ON COLUMN alert_container_reverse_shell.shell_type IS 'Shell类型(bash/python/nc等)';
COMMENT ON COLUMN alert_container_reverse_shell.status IS '状态: 0-待处理 1-已处理 2-已忽略';
COMMENT ON COLUMN alert_container_reverse_shell.container_id IS '容器ID(64字符hex)';
