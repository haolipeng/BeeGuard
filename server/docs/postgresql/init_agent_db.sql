-- =====================================================
-- SOC 管理中心-客户端管理数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- 说明: 管理中心模块-Agent客户端管理相关表
-- =====================================================


-- =====================================================
-- 1. Agent客户端信息表 (agent_info)
-- =====================================================
CREATE TABLE IF NOT EXISTS agent_info (
    id                  BIGSERIAL PRIMARY KEY,
    agent_id            VARCHAR(64) NOT NULL,                                  -- Agent唯一标识(如AGT-20251225-001)
    agent_version       VARCHAR(32),                                           -- 安装版本
    connection_status   SMALLINT NOT NULL DEFAULT 0,                           -- 连接状态
    host_name           VARCHAR(128) NOT NULL,                                 -- 主机名
    host_ip             VARCHAR(45) NOT NULL,                                  -- IP地址
    os_type             VARCHAR(16) NOT NULL,                                  -- 操作系统类型
    os_version          VARCHAR(128),                                          -- 操作系统版本
    os_arch             VARCHAR(32),                                           -- CPU架构
    cpu_count           INT,                                                   -- CPU核数
    memory_total        BIGINT,                                                -- 内存总量(字节)
    disk_total          BIGINT,                                                -- 磁盘总量(字节)
    last_connected_at   TIMESTAMP,                                             -- 最后连接时间
    registered_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,          -- 注册时间
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_agent_id ON agent_info(agent_id);
CREATE INDEX IF NOT EXISTS idx_ai_connection_status ON agent_info(connection_status);
CREATE INDEX IF NOT EXISTS idx_ai_host_name ON agent_info(host_name);
CREATE INDEX IF NOT EXISTS idx_ai_host_ip ON agent_info(host_ip);
CREATE INDEX IF NOT EXISTS idx_ai_last_connected_at ON agent_info(last_connected_at);

COMMENT ON TABLE agent_info IS '管理中心-Agent客户端信息';
COMMENT ON COLUMN agent_info.agent_id IS 'Agent唯一标识(如AGT-20251225-001)';
COMMENT ON COLUMN agent_info.agent_version IS '安装版本(如2.1.5)';
COMMENT ON COLUMN agent_info.connection_status IS '连接状态: 0-已断开 1-已连接';
COMMENT ON COLUMN agent_info.host_name IS '主机名';
COMMENT ON COLUMN agent_info.host_ip IS 'IP地址(支持IPv4/IPv6)';
COMMENT ON COLUMN agent_info.os_type IS '操作系统类型: linux/windows';
COMMENT ON COLUMN agent_info.os_version IS '操作系统版本(如Ubuntu 20.04.3 LTS)';
COMMENT ON COLUMN agent_info.os_arch IS 'CPU架构(如x86_64, aarch64)';
COMMENT ON COLUMN agent_info.cpu_count IS 'CPU核数';
COMMENT ON COLUMN agent_info.memory_total IS '内存总量(字节)';
COMMENT ON COLUMN agent_info.disk_total IS '磁盘总量(字节)';
COMMENT ON COLUMN agent_info.last_connected_at IS '最后连接时间';
COMMENT ON COLUMN agent_info.registered_at IS 'Agent首次注册时间';


-- =====================================================
-- 初始化完成
-- =====================================================
