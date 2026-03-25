-- 000004: 事件采集表
-- 包含: event_execve, event_connect, event_dns, event_file (4 表)

-- 1. 进程执行事件表 (event_execve)
CREATE TABLE IF NOT EXISTS event_execve (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(256),
    pid             INT NOT NULL,
    tgid            INT,
    ppid            INT,
    pgid            INT,
    uid             INT,
    comm            VARCHAR(16),
    exe_path        VARCHAR(512),
    args            TEXT,
    event_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_execve_agent_id ON event_execve(agent_id);
CREATE INDEX IF NOT EXISTS idx_execve_event_time ON event_execve(event_time);
CREATE INDEX IF NOT EXISTS idx_execve_exe_path ON event_execve(exe_path);
CREATE INDEX IF NOT EXISTS idx_execve_ppid ON event_execve(ppid);
CREATE INDEX IF NOT EXISTS idx_execve_comm ON event_execve(comm);

COMMENT ON TABLE event_execve IS '事件采集-进程执行事件';


-- 2. 出站连接事件表 (event_connect)
CREATE TABLE IF NOT EXISTS event_connect (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(256),
    pid             INT NOT NULL,
    tgid            INT,
    ppid            INT,
    uid             INT,
    comm            VARCHAR(16),
    exe_path        VARCHAR(512),
    protocol        VARCHAR(16),
    remote_ip       VARCHAR(64),
    remote_port     INT,
    pid_tree        TEXT,
    event_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_connect_agent_id ON event_connect(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_connect_event_time ON event_connect(event_time);
CREATE INDEX IF NOT EXISTS idx_event_connect_remote_ip ON event_connect(remote_ip);
CREATE INDEX IF NOT EXISTS idx_event_connect_comm ON event_connect(comm);
CREATE INDEX IF NOT EXISTS idx_event_connect_exe_path ON event_connect(exe_path);

COMMENT ON TABLE event_connect IS '事件采集-出站连接事件';
COMMENT ON COLUMN event_connect.protocol IS '协议类型: tcp/udp';


-- 3. DNS查询事件表 (event_dns)
CREATE TABLE IF NOT EXISTS event_dns (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(256),
    pid             INT NOT NULL,
    tgid            INT,
    ppid            INT,
    uid             INT,
    comm            VARCHAR(16),
    exe_path        VARCHAR(512),
    domain          VARCHAR(255),
    query_type      VARCHAR(16),
    pid_tree        TEXT,
    event_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_dns_agent_id ON event_dns(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_dns_event_time ON event_dns(event_time);
CREATE INDEX IF NOT EXISTS idx_event_dns_domain ON event_dns(domain);
CREATE INDEX IF NOT EXISTS idx_event_dns_comm ON event_dns(comm);
CREATE INDEX IF NOT EXISTS idx_event_dns_exe_path ON event_dns(exe_path);

COMMENT ON TABLE event_dns IS '事件采集-DNS查询事件';
COMMENT ON COLUMN event_dns.query_type IS '查询类型: A/AAAA/CNAME/MX/TXT等';


-- 4. 文件操作事件表 (event_file)
CREATE TABLE IF NOT EXISTS event_file (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(64) NOT NULL,
    host_name VARCHAR(128),
    host_ip VARCHAR(256),

    -- 进程信息
    pid INTEGER NOT NULL,
    tgid INTEGER,
    ppid INTEGER,
    uid INTEGER,

    comm VARCHAR(16),
    exe_path VARCHAR(512),

    -- 文件操作信息
    action VARCHAR(16) NOT NULL,
    new_path VARCHAR(512) NOT NULL,
    old_path VARCHAR(512),
    s_id VARCHAR(64),

    -- 进程树
    pid_tree TEXT,

    -- 关联socket信息（可选）
    socket_pid INTEGER,
    remote_ip VARCHAR(64),
    remote_port INTEGER,
    local_ip VARCHAR(64),
    local_port INTEGER,

    event_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_file_agent_id ON event_file(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_file_event_time ON event_file(event_time);
CREATE INDEX IF NOT EXISTS idx_event_file_new_path ON event_file(new_path);
CREATE INDEX IF NOT EXISTS idx_event_file_action ON event_file(action);
CREATE INDEX IF NOT EXISTS idx_event_file_comm ON event_file(comm);

COMMENT ON TABLE event_file IS '文件操作事件记录表，存储eBPF捕获的文件创建/重命名/删除事件';
COMMENT ON COLUMN event_file.action IS '文件操作类型: create/rename/delete';
COMMENT ON COLUMN event_file.new_path IS '目标文件路径';
COMMENT ON COLUMN event_file.old_path IS '原文件路径（仅rename操作）';
