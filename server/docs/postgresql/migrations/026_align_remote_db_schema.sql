-- =====================================================
-- 迁移脚本: 026_align_remote_db_schema.sql
-- 目的: 将远程数据库 schema 与本地 SQL 定义 / Go 模型对齐
-- 执行顺序: 必须在维护窗口期执行，涉及表重建
-- =====================================================

BEGIN;

-- =====================================================
-- 1. 漏洞扫描表重建（最大变更）
-- =====================================================

-- 1a. 删除旧视图（依赖这些表）
DROP VIEW IF EXISTS v_vuln_count_hosts CASCADE;
DROP VIEW IF EXISTS v_vuln_count_images CASCADE;
DROP VIEW IF EXISTS v_vuln_count_vuls CASCADE;
DROP VIEW IF EXISTS v_vuln_count_image_vuls CASCADE;

-- 1b. 删除旧的 detail 表（有外键依赖或需要添加 scan_id）
DROP TABLE IF EXISTS host_vuln_detail CASCADE;
DROP TABLE IF EXISTS image_vuln_detail CASCADE;

-- 1c. 删除旧的 scan 表
DROP TABLE IF EXISTS host_vuln_scan CASCADE;
DROP TABLE IF EXISTS image_vuln_scan CASCADE;

-- 1d. 创建新的 host_vuln_scan_task 表
CREATE TABLE IF NOT EXISTS host_vuln_scan_task (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_id         BIGINT,
    host_name       VARCHAR(128) NOT NULL,
    host_ip         VARCHAR(45) NOT NULL,
    scan_status     SMALLINT NOT NULL DEFAULT 0,
    scan_trigger    VARCHAR(16) DEFAULT 'auto',
    total_packages  INT,
    matched_vulns   INT,
    scan_duration   INT,
    error_message   TEXT,
    scan_time       TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hvst_agent_id ON host_vuln_scan_task(agent_id);
CREATE INDEX IF NOT EXISTS idx_hvst_host_ip ON host_vuln_scan_task(host_ip);
CREATE INDEX IF NOT EXISTS idx_hvst_scan_time ON host_vuln_scan_task(scan_time);
CREATE INDEX IF NOT EXISTS idx_hvst_scan_status ON host_vuln_scan_task(scan_status);

COMMENT ON TABLE host_vuln_scan_task IS '漏洞发现-主机漏洞扫描任务记录';
COMMENT ON COLUMN host_vuln_scan_task.scan_status IS '任务状态: 0-进行中 1-成功 2-失败';
COMMENT ON COLUMN host_vuln_scan_task.scan_trigger IS '触发方式: auto-定时自动扫描 manual-手动触发';

-- 1e. 创建新的 image_vuln_scan_task 表
CREATE TABLE IF NOT EXISTS image_vuln_scan_task (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    image_id        VARCHAR(128) NOT NULL,
    image_name      VARCHAR(256) NOT NULL,
    scan_status     SMALLINT NOT NULL DEFAULT 0,
    scan_trigger    VARCHAR(16) DEFAULT 'auto',
    total_packages  INT,
    matched_vulns   INT,
    scan_duration   INT,
    error_message   TEXT,
    scan_time       TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ivst_agent_id ON image_vuln_scan_task(agent_id);
CREATE INDEX IF NOT EXISTS idx_ivst_image_id ON image_vuln_scan_task(image_id);
CREATE INDEX IF NOT EXISTS idx_ivst_scan_time ON image_vuln_scan_task(scan_time);
CREATE INDEX IF NOT EXISTS idx_ivst_scan_status ON image_vuln_scan_task(scan_status);

COMMENT ON TABLE image_vuln_scan_task IS '漏洞发现-镜像漏洞扫描任务记录';
COMMENT ON COLUMN image_vuln_scan_task.scan_status IS '任务状态: 0-进行中 1-成功 2-失败';
COMMENT ON COLUMN image_vuln_scan_task.scan_trigger IS '触发方式: auto-定时自动扫描 manual-手动触发';

-- 1f. 重建 host_vuln_detail 表（含 scan_id FK, vuln_name VARCHAR(256)）
CREATE TABLE IF NOT EXISTS host_vuln_detail (
    id                  BIGSERIAL PRIMARY KEY,
    scan_id             BIGINT NOT NULL REFERENCES host_vuln_scan_task(id),
    agent_id            VARCHAR(64) NOT NULL,
    host_id             BIGINT,
    vuln_id             BIGINT NOT NULL,
    cve_id              VARCHAR(32) NOT NULL,
    package_name        VARCHAR(128) NOT NULL,
    installed_version   VARCHAR(64),
    fixed_version       VARCHAR(64),
    status              SMALLINT NOT NULL,
    host_name           VARCHAR(128),
    host_ip             VARCHAR(45),
    vuln_name           VARCHAR(256),
    severity            VARCHAR(16),
    cvss_score          DECIMAL(3,1),
    description         TEXT,
    fix_suggestion      TEXT,
    scan_time           TIMESTAMP NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hvd_scan_id ON host_vuln_detail(scan_id);
CREATE INDEX IF NOT EXISTS idx_hvd_agent_id ON host_vuln_detail(agent_id);
CREATE INDEX IF NOT EXISTS idx_hvd_vuln_id ON host_vuln_detail(vuln_id);
CREATE INDEX IF NOT EXISTS idx_hvd_cve_id ON host_vuln_detail(cve_id);
CREATE INDEX IF NOT EXISTS idx_hvd_status ON host_vuln_detail(status);
CREATE INDEX IF NOT EXISTS idx_hvd_scan_time ON host_vuln_detail(scan_time);

COMMENT ON TABLE host_vuln_detail IS '漏洞发现-主机漏洞发现记录';
COMMENT ON COLUMN host_vuln_detail.scan_id IS '关联扫描任务ID(host_vuln_scan_task.id)';
COMMENT ON COLUMN host_vuln_detail.status IS '状态: 0-未修复 1-已修复 2-已忽略';

-- 1g. 重建 image_vuln_detail 表（含 scan_id FK, vuln_name VARCHAR(256)）
CREATE TABLE IF NOT EXISTS image_vuln_detail (
    id                  BIGSERIAL PRIMARY KEY,
    scan_id             BIGINT NOT NULL REFERENCES image_vuln_scan_task(id),
    agent_id            VARCHAR(64) NOT NULL,
    image_id            VARCHAR(128) NOT NULL,
    vuln_id             BIGINT NOT NULL,
    cve_id              VARCHAR(32) NOT NULL,
    package_name        VARCHAR(128) NOT NULL,
    installed_version   VARCHAR(64),
    fixed_version       VARCHAR(64),
    status              SMALLINT NOT NULL,
    image_name          VARCHAR(256),
    vuln_name           VARCHAR(256),
    severity            VARCHAR(16),
    cvss_score          DECIMAL(3,1),
    description         TEXT,
    fix_suggestion      TEXT,
    scan_time           TIMESTAMP NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ivd_scan_id ON image_vuln_detail(scan_id);
CREATE INDEX IF NOT EXISTS idx_ivd_agent_id ON image_vuln_detail(agent_id);
CREATE INDEX IF NOT EXISTS idx_ivd_image_id ON image_vuln_detail(image_id);
CREATE INDEX IF NOT EXISTS idx_ivd_vuln_id ON image_vuln_detail(vuln_id);
CREATE INDEX IF NOT EXISTS idx_ivd_cve_id ON image_vuln_detail(cve_id);
CREATE INDEX IF NOT EXISTS idx_ivd_status ON image_vuln_detail(status);

COMMENT ON TABLE image_vuln_detail IS '漏洞发现-镜像漏洞发现记录';
COMMENT ON COLUMN image_vuln_detail.scan_id IS '关联扫描任务ID(image_vuln_scan_task.id)';
COMMENT ON COLUMN image_vuln_detail.status IS '状态: 0-未修复 1-已修复 2-已忽略';

-- 1h. 重建所有 4 个漏洞统计视图
CREATE OR REPLACE VIEW v_vuln_count_hosts AS
SELECT
    hs.host_ip,
    hs.host_name,
    MAX(hd.scan_time)  AS last_scan_time,
    MIN(hd.scan_time)  AS first_scan_time,
    COUNT(CASE WHEN vi.severity = 'critical' THEN 1 END) AS critical_vulns,
    COUNT(CASE WHEN vi.severity = 'high'     THEN 1 END) AS high_vulns,
    COUNT(CASE WHEN vi.severity = 'medium'   THEN 1 END) AS medium_vulns,
    COUNT(CASE WHEN vi.severity = 'low'      THEN 1 END) AS low_vulns,
    COUNT(*)                                              AS total_vulns
FROM host_vuln_detail hd
JOIN vuln_info vi ON hd.vuln_id = vi.id
JOIN host_vuln_scan_task hs ON hd.scan_id = hs.id
WHERE hd.status = 0
GROUP BY hs.host_ip, hs.host_name;

COMMENT ON VIEW v_vuln_count_hosts IS '漏洞统计-按主机维度';

CREATE OR REPLACE VIEW v_vuln_count_images AS
SELECT
    ivd.image_id,
    ivs.image_name,
    MAX(ivd.scan_time)  AS last_scan_time,
    MIN(ivd.scan_time)  AS first_scan_time,
    COUNT(CASE WHEN vi.severity = 'critical' THEN 1 END) AS critical_vulns,
    COUNT(CASE WHEN vi.severity = 'high'     THEN 1 END) AS high_vulns,
    COUNT(CASE WHEN vi.severity = 'medium'   THEN 1 END) AS medium_vulns,
    COUNT(CASE WHEN vi.severity = 'low'      THEN 1 END) AS low_vulns,
    COUNT(*)                                              AS total_vulns
FROM image_vuln_detail ivd
JOIN vuln_info vi ON ivd.vuln_id = vi.id
JOIN image_vuln_scan_task ivs ON ivd.scan_id = ivs.id
WHERE ivd.status = 0
GROUP BY ivd.image_id, ivs.image_name;

COMMENT ON VIEW v_vuln_count_images IS '漏洞统计-按镜像维度';

CREATE OR REPLACE VIEW v_vuln_count_vuls AS
SELECT
    vi.id                AS vuln_id,
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    vi.description,
    vi.fix_suggestion,
    MIN(hd.scan_time)    AS first_scan_time,
    MAX(hd.scan_time)    AS last_scan_time,
    COUNT(DISTINCT hd.agent_id) AS affected_host_count,
    json_agg(json_build_object(
        'host_id',   hd.host_id,
        'host_name', hs.host_name,
        'host_ip',   hs.host_ip,
        'scan_time', hd.scan_time,
        'status',    hd.status
    )) AS affected_hosts
FROM vuln_info vi
JOIN host_vuln_detail hd ON vi.id = hd.vuln_id
JOIN host_vuln_scan_task hs ON hd.scan_id = hs.id
GROUP BY vi.id, vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score, vi.description, vi.fix_suggestion;

COMMENT ON VIEW v_vuln_count_vuls IS '漏洞统计-按漏洞维度(主机)';

CREATE OR REPLACE VIEW v_vuln_count_image_vuls AS
SELECT
    vi.id                AS vuln_id,
    vi.cve_id,
    vi.vuln_name,
    vi.severity,
    vi.cvss_score,
    vi.description,
    vi.fix_suggestion,
    MIN(ivd.scan_time)   AS first_scan_time,
    MAX(ivd.scan_time)   AS last_scan_time,
    COUNT(DISTINCT ivd.image_id) AS affected_image_count,
    json_agg(json_build_object(
        'agent_id',   ivd.agent_id,
        'image_id',   ivd.image_id,
        'image_name', ivs.image_name,
        'scan_time',  ivd.scan_time,
        'status',     ivd.status
    )) AS affected_images
FROM vuln_info vi
JOIN image_vuln_detail ivd ON vi.id = ivd.vuln_id
JOIN image_vuln_scan_task ivs ON ivd.scan_id = ivs.id
GROUP BY vi.id, vi.cve_id, vi.vuln_name, vi.severity, vi.cvss_score, vi.description, vi.fix_suggestion;

COMMENT ON VIEW v_vuln_count_image_vuls IS '漏洞统计-按漏洞维度(镜像)';


-- =====================================================
-- 2. asset_image 添加 runtime 列
-- =====================================================
ALTER TABLE asset_image ADD COLUMN IF NOT EXISTS runtime VARCHAR(32);


-- =====================================================
-- 3. agent_info 修复 NOT NULL 约束
-- =====================================================

-- agent_version: Go 模型中为 *string（可空），远程可能是 NOT NULL，需要 DROP
ALTER TABLE agent_info ALTER COLUMN agent_version DROP NOT NULL;

-- os_type: Go 模型中为 string + gorm:"not null"，远程可能缺少 NOT NULL
-- 先填充可能存在的 NULL 值
UPDATE agent_info SET os_type = 'linux' WHERE os_type IS NULL;
ALTER TABLE agent_info ALTER COLUMN os_type SET NOT NULL;


-- =====================================================
-- 4. 补齐缺失的唯一索引（10 个）
-- =====================================================

-- 创建唯一索引前清理可能的重复数据
-- asset_host: 按 agent_id 去重，保留最新记录
DELETE FROM asset_host a USING asset_host b
WHERE a.agent_id = b.agent_id AND a.id < b.id;

-- asset_port: 按 (agent_id, port, protocol) 去重
DELETE FROM asset_port a USING asset_port b
WHERE a.agent_id = b.agent_id AND a.port = b.port AND a.protocol = b.protocol AND a.id < b.id;

-- asset_process: 按 (agent_id, path) 去重
DELETE FROM asset_process a USING asset_process b
WHERE a.agent_id = b.agent_id AND a.path = b.path AND a.id < b.id;

-- asset_database: 按 (agent_id, db_type) 去重
DELETE FROM asset_database a USING asset_database b
WHERE a.agent_id = b.agent_id AND a.db_type = b.db_type AND a.id < b.id;

-- asset_web_service: 按 (agent_id, server_type) 去重
DELETE FROM asset_web_service a USING asset_web_service b
WHERE a.agent_id = b.agent_id AND a.server_type = b.server_type AND a.id < b.id;

-- asset_system_service: 按 (agent_id, name) 去重
DELETE FROM asset_system_service a USING asset_system_service b
WHERE a.agent_id = b.agent_id AND a.name = b.name AND a.id < b.id;

-- asset_software: 按 (agent_id, name, type) 去重
DELETE FROM asset_software a USING asset_software b
WHERE a.agent_id = b.agent_id AND a.name = b.name AND a.type = b.type AND a.id < b.id;

-- asset_container: 按 (agent_id, container_id) 去重
DELETE FROM asset_container a USING asset_container b
WHERE a.agent_id = b.agent_id AND a.container_id = b.container_id AND a.id < b.id;

-- asset_env_suspicious: 按 (agent_id, var_name) 去重
DELETE FROM asset_env_suspicious a USING asset_env_suspicious b
WHERE a.agent_id = b.agent_id AND a.var_name = b.var_name AND a.id < b.id;

-- asset_kmod: 按 (agent_id, name) 去重
DELETE FROM asset_kmod a USING asset_kmod b
WHERE a.agent_id = b.agent_id AND a.name = b.name AND a.id < b.id;

-- 创建唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_host_agent_id ON asset_host(agent_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_port_agent_port ON asset_port(agent_id, port, protocol);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_process_agent_path ON asset_process(agent_id, path);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_database_agent_type ON asset_database(agent_id, db_type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_web_service_agent_type ON asset_web_service(agent_id, server_type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_system_service_agent_name ON asset_system_service(agent_id, name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_software_agent_name_type ON asset_software(agent_id, name, type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_container_agent_cid ON asset_container(agent_id, container_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_env_suspicious_agent_var ON asset_env_suspicious(agent_id, var_name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_kmod_agent_name ON asset_kmod(agent_id, name);


COMMIT;

-- =====================================================
-- 迁移完成
-- =====================================================
