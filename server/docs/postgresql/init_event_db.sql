-- =====================================================
-- SOC 事件采集数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- 说明: eBPF事件采集相关表(DNS/Execve/Connect)
-- =====================================================


-- =====================================================
-- 1. DNS查询事件表 (event_dns)
-- =====================================================
CREATE TABLE IF NOT EXISTS event_dns (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(45),

    pid             INT NOT NULL,                                           -- 进程ID（线程ID）
    tgid            INT,                                                    -- 线程组ID（进程ID）
    ppid            INT,                                                    -- 父进程ID
    uid             INT,                                                    -- 用户ID

    comm            VARCHAR(16),                                            -- 进程名（最多16字节）
    exe_path        VARCHAR(512),                                           -- 可执行文件完整路径

    domain          VARCHAR(255),                                           -- 查询域名
    query_type      VARCHAR(16),                                            -- 查询类型（A/AAAA/CNAME/MX/TXT等）

    pid_tree        TEXT,                                                   -- 进程树（预留）

    event_time      TIMESTAMP NOT NULL,                                     -- 事件发生时间
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP            -- 创建时间
);

CREATE INDEX IF NOT EXISTS idx_event_dns_agent_id ON event_dns(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_dns_event_time ON event_dns(event_time);
CREATE INDEX IF NOT EXISTS idx_event_dns_domain ON event_dns(domain);
CREATE INDEX IF NOT EXISTS idx_event_dns_comm ON event_dns(comm);
CREATE INDEX IF NOT EXISTS idx_event_dns_exe_path ON event_dns(exe_path);

COMMENT ON TABLE event_dns IS '事件采集-DNS查询事件';
COMMENT ON COLUMN event_dns.query_type IS '查询类型: A/AAAA/CNAME/MX/TXT等';


-- =====================================================
-- 2. 进程执行事件表 (event_execve)
-- =====================================================
CREATE TABLE IF NOT EXISTS event_execve (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(45),

    pid             INT NOT NULL,                                           -- 进程ID（线程ID）
    tgid            INT,                                                    -- 线程组ID（进程ID）
    ppid            INT,                                                    -- 父进程ID
    pgid            INT,                                                    -- 进程组ID
    uid             INT,                                                    -- 用户ID

    comm            VARCHAR(16),                                            -- 进程名（最多16字节）
    exe_path        VARCHAR(512),                                           -- 可执行文件完整路径
    args            TEXT,                                                   -- 命令行参数

    event_time      TIMESTAMP NOT NULL,                                     -- 事件发生时间
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP            -- 创建时间
);

CREATE INDEX IF NOT EXISTS idx_execve_agent_id ON event_execve(agent_id);
CREATE INDEX IF NOT EXISTS idx_execve_event_time ON event_execve(event_time);
CREATE INDEX IF NOT EXISTS idx_execve_exe_path ON event_execve(exe_path);
CREATE INDEX IF NOT EXISTS idx_execve_ppid ON event_execve(ppid);
CREATE INDEX IF NOT EXISTS idx_execve_comm ON event_execve(comm);

COMMENT ON TABLE event_execve IS '事件采集-进程执行事件';


-- =====================================================
-- 3. 出站连接事件表 (event_connect)
-- =====================================================
CREATE TABLE IF NOT EXISTS event_connect (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(45),

    pid             INT NOT NULL,                                           -- 进程ID（线程ID）
    tgid            INT,                                                    -- 线程组ID（进程ID）
    ppid            INT,                                                    -- 父进程ID
    uid             INT,                                                    -- 用户ID

    comm            VARCHAR(16),                                            -- 进程名（最多16字节）
    exe_path        VARCHAR(512),                                           -- 可执行文件完整路径

    protocol        VARCHAR(16),                                            -- 协议类型（tcp/udp）
    remote_ip       VARCHAR(64),                                            -- 远端IP地址
    remote_port     INT,                                                    -- 远端端口

    pid_tree        TEXT,                                                   -- 进程树（预留）

    event_time      TIMESTAMP NOT NULL,                                     -- 事件发生时间
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP            -- 创建时间
);

CREATE INDEX IF NOT EXISTS idx_event_connect_agent_id ON event_connect(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_connect_event_time ON event_connect(event_time);
CREATE INDEX IF NOT EXISTS idx_event_connect_remote_ip ON event_connect(remote_ip);
CREATE INDEX IF NOT EXISTS idx_event_connect_comm ON event_connect(comm);
CREATE INDEX IF NOT EXISTS idx_event_connect_exe_path ON event_connect(exe_path);

COMMENT ON TABLE event_connect IS '事件采集-出站连接事件';
COMMENT ON COLUMN event_connect.protocol IS '协议类型: tcp/udp';


-- =====================================================
-- 4. 文件操作事件表 (event_file)
-- =====================================================
CREATE TABLE IF NOT EXISTS event_file (
    id BIGSERIAL PRIMARY KEY,
    agent_id VARCHAR(64) NOT NULL,
    host_name VARCHAR(128),
    host_ip VARCHAR(45),

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


-- =====================================================
-- 初始化完成
-- =====================================================
