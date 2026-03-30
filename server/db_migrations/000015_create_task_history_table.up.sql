-- Agent 任务历史表
CREATE TABLE IF NOT EXISTS agent_task_history (
    id              BIGSERIAL PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL UNIQUE,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(256),
    task_type       INT NOT NULL,
    task_name       VARCHAR(128) NOT NULL,
    parameters      JSONB,
    status          SMALLINT NOT NULL DEFAULT 0,     -- 0=已下发 1=执行中 2=成功 3=失败 4=超时
    result_message  TEXT,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_history_agent_id ON agent_task_history(agent_id);
CREATE INDEX idx_task_history_task_type ON agent_task_history(task_type);
CREATE INDEX idx_task_history_status ON agent_task_history(status);
CREATE INDEX idx_task_history_created ON agent_task_history(created_at);
