-- =====================================================
-- SOC 漏洞发现数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- 说明: 漏洞发现模块相关表(主机漏洞+容器漏洞)
-- =====================================================


-- =====================================================
-- 1. 主机漏洞扫描任务表 (host_vuln_scan_task)
-- =====================================================
CREATE TABLE IF NOT EXISTS host_vuln_scan_task (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,                                  -- Agent唯一标识
    host_id         BIGINT,                                                -- 关联主机ID
    host_name       VARCHAR(128) NOT NULL,                                 -- 主机名称
    host_ip         VARCHAR(45) NOT NULL,                                  -- 主机IP
    scan_status     SMALLINT NOT NULL DEFAULT 0,                           -- 任务状态: 0-进行中 1-成功 2-失败
    scan_trigger    VARCHAR(16) DEFAULT 'auto',                            -- 触发方式: auto/manual
    total_packages  INT,                                                    -- 扫描软件包总数
    matched_vulns   INT,                                                    -- 匹配到的漏洞总数
    scan_duration   INT,                                                    -- 扫描耗时(ms)
    error_message   TEXT,                                                   -- 失败时的错误信息
    scan_time       TIMESTAMP NOT NULL,                                    -- 扫描时间
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hvst_agent_id ON host_vuln_scan_task(agent_id);
CREATE INDEX IF NOT EXISTS idx_hvst_host_ip ON host_vuln_scan_task(host_ip);
CREATE INDEX IF NOT EXISTS idx_hvst_scan_time ON host_vuln_scan_task(scan_time);
CREATE INDEX IF NOT EXISTS idx_hvst_scan_status ON host_vuln_scan_task(scan_status);

COMMENT ON TABLE host_vuln_scan_task IS '漏洞发现-主机漏洞扫描任务记录';
COMMENT ON COLUMN host_vuln_scan_task.agent_id IS 'Agent唯一标识';
COMMENT ON COLUMN host_vuln_scan_task.host_id IS '关联主机ID(业务层关联asset_host.id)';
COMMENT ON COLUMN host_vuln_scan_task.scan_status IS '任务状态: 0-进行中 1-成功 2-失败';
COMMENT ON COLUMN host_vuln_scan_task.scan_trigger IS '触发方式: auto-定时自动扫描 manual-手动触发';


-- =====================================================
-- 2. 漏洞信息表 (vuln_info) - 主机/容器共用
-- =====================================================
CREATE TABLE IF NOT EXISTS vuln_info (
    id                  BIGSERIAL PRIMARY KEY,
    cve_id              VARCHAR(32) NOT NULL,                              -- CVE编号
    vuln_name           VARCHAR(256) NOT NULL,                             -- 漏洞名称
    severity            VARCHAR(16) NOT NULL,                              -- 漏洞等级
    cvss_score          DECIMAL(3,1),                                      -- CVSS评分
    description         TEXT,                                               -- 漏洞描述
    fix_suggestion      TEXT,                                               -- 修复建议
    reference_urls      TEXT,                                               -- 参考链接
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vi_cve_id ON vuln_info(cve_id);
CREATE INDEX IF NOT EXISTS idx_vi_severity ON vuln_info(severity);
CREATE INDEX IF NOT EXISTS idx_vi_cvss_score ON vuln_info(cvss_score);

COMMENT ON TABLE vuln_info IS '漏洞发现-漏洞信息(主机/容器共用)';
COMMENT ON COLUMN vuln_info.severity IS '漏洞等级: critical/high/medium/low';
COMMENT ON COLUMN vuln_info.cvss_score IS 'CVSS评分(0.0-10.0)';


-- =====================================================
-- 3. 主机漏洞发现记录表 (host_vuln_detail)
-- =====================================================
CREATE TABLE IF NOT EXISTS host_vuln_detail (
    id                  BIGSERIAL PRIMARY KEY,
    scan_id             BIGINT NOT NULL REFERENCES host_vuln_scan_task(id), -- 关联扫描任务
    agent_id            VARCHAR(64) NOT NULL,                              -- Agent唯一标识
    host_id             BIGINT,                                            -- 关联主机ID
    vuln_id             BIGINT NOT NULL,                                   -- 漏洞ID
    cve_id              VARCHAR(32) NOT NULL,                              -- CVE编号(冗余)
    package_name        VARCHAR(128) NOT NULL,                             -- 受影响软件包
    installed_version   VARCHAR(64),                                       -- 当前版本
    fixed_version       VARCHAR(64),                                       -- 修复版本
    status              SMALLINT NOT NULL,                                 -- 状态
    host_name           VARCHAR(128),                                      -- 主机名称(冗余)
    host_ip             VARCHAR(45),                                       -- 主机IP(冗余)
    vuln_name           VARCHAR(256),                                      -- 漏洞名称(冗余)
    severity            VARCHAR(16),                                       -- 漏洞等级(冗余)
    cvss_score          DECIMAL(3,1),                                      -- CVSS评分(冗余)
    description         TEXT,                                               -- 漏洞描述(冗余)
    fix_suggestion      TEXT,                                               -- 修复建议(冗余)
    scan_time           TIMESTAMP NOT NULL,                                -- 扫描时间
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
COMMENT ON COLUMN host_vuln_detail.vuln_id IS '漏洞ID(业务层关联vuln_info.id)';
COMMENT ON COLUMN host_vuln_detail.cve_id IS 'CVE编号(冗余字段，方便查询)';
COMMENT ON COLUMN host_vuln_detail.status IS '状态: 0-未修复 1-已修复 2-已忽略';


-- =====================================================
-- 4. 镜像漏洞扫描任务表 (image_vuln_scan_task)
-- =====================================================
CREATE TABLE IF NOT EXISTS image_vuln_scan_task (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,                                  -- Agent唯一标识(镜像所在主机)
    image_id        VARCHAR(128) NOT NULL,                                 -- 镜像ID
    image_name      VARCHAR(256) NOT NULL,                                 -- 镜像名称(含tag)
    scan_status     SMALLINT NOT NULL DEFAULT 0,                           -- 任务状态: 0-进行中 1-成功 2-失败
    scan_trigger    VARCHAR(16) DEFAULT 'auto',                            -- 触发方式: auto/manual
    total_packages  INT,                                                    -- 扫描软件包总数
    matched_vulns   INT,                                                    -- 匹配到的漏洞总数
    scan_duration   INT,                                                    -- 扫描耗时(ms)
    error_message   TEXT,                                                   -- 失败时的错误信息
    scan_time       TIMESTAMP NOT NULL,                                    -- 扫描时间
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ivst_agent_id ON image_vuln_scan_task(agent_id);
CREATE INDEX IF NOT EXISTS idx_ivst_image_id ON image_vuln_scan_task(image_id);
CREATE INDEX IF NOT EXISTS idx_ivst_scan_time ON image_vuln_scan_task(scan_time);
CREATE INDEX IF NOT EXISTS idx_ivst_scan_status ON image_vuln_scan_task(scan_status);

COMMENT ON TABLE image_vuln_scan_task IS '漏洞发现-镜像漏洞扫描任务记录';
COMMENT ON COLUMN image_vuln_scan_task.agent_id IS 'Agent唯一标识(镜像所在主机)';
COMMENT ON COLUMN image_vuln_scan_task.image_name IS '镜像名称(包含tag标签)';
COMMENT ON COLUMN image_vuln_scan_task.scan_status IS '任务状态: 0-进行中 1-成功 2-失败';
COMMENT ON COLUMN image_vuln_scan_task.scan_trigger IS '触发方式: auto-定时自动扫描 manual-手动触发';


-- =====================================================
-- 5. 镜像漏洞发现记录表 (image_vuln_detail)
-- =====================================================
CREATE TABLE IF NOT EXISTS image_vuln_detail (
    id                  BIGSERIAL PRIMARY KEY,
    scan_id             BIGINT NOT NULL REFERENCES image_vuln_scan_task(id), -- 关联扫描任务
    agent_id            VARCHAR(64) NOT NULL,                              -- Agent唯一标识(镜像所在主机)
    image_id            VARCHAR(128) NOT NULL,                             -- 镜像ID
    vuln_id             BIGINT NOT NULL,                                   -- 漏洞ID
    cve_id              VARCHAR(32) NOT NULL,                              -- CVE编号(冗余)
    package_name        VARCHAR(128) NOT NULL,                             -- 受影响软件包
    installed_version   VARCHAR(64),                                       -- 当前版本
    fixed_version       VARCHAR(64),                                       -- 修复版本
    status              SMALLINT NOT NULL,                                 -- 状态
    image_name          VARCHAR(256),                                      -- 镜像名称(冗余)
    vuln_name           VARCHAR(256),                                      -- 漏洞名称(冗余)
    severity            VARCHAR(16),                                       -- 漏洞等级(冗余)
    cvss_score          DECIMAL(3,1),                                      -- CVSS评分(冗余)
    description         TEXT,                                               -- 漏洞描述(冗余)
    fix_suggestion      TEXT,                                               -- 修复建议(冗余)
    scan_time           TIMESTAMP NOT NULL,                                -- 扫描时间
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
COMMENT ON COLUMN image_vuln_detail.vuln_id IS '漏洞ID(业务层关联vuln_info.id)';
COMMENT ON COLUMN image_vuln_detail.cve_id IS 'CVE编号(冗余字段，方便查询)';
COMMENT ON COLUMN image_vuln_detail.status IS '状态: 0-未修复 1-已修复 2-已忽略';


-- =====================================================
-- 6. 漏洞基本信息表 (vulnerability_info)
-- =====================================================
CREATE TABLE IF NOT EXISTS vulnerability_info (
    id              BIGSERIAL PRIMARY KEY,
    cve_id          VARCHAR(32),                                             -- CVE编号
    vuln_name       VARCHAR(255) NOT NULL,                                   -- 漏洞名称
    severity        VARCHAR(20) NOT NULL,                                    -- 严重级别: critical/high/medium/low
    cvss_score      DECIMAL(3,1),                                            -- CVSS评分
    description     TEXT,                                                    -- 漏洞描述
    fix_suggestion  TEXT,                                                    -- 修复建议
    reference       TEXT,                                                    -- 参考链接
    publish_date    TIMESTAMP,                                               -- 发布日期
    update_time     TIMESTAMP,                                               -- 更新时间
    status          VARCHAR(32) NOT NULL DEFAULT 'active',                   -- 状态: active/inactive
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP                                                -- 软删除时间
);

CREATE INDEX IF NOT EXISTS idx_vi2_cve_id ON vulnerability_info(cve_id);
CREATE INDEX IF NOT EXISTS idx_vi2_deleted_at ON vulnerability_info(deleted_at);

COMMENT ON TABLE vulnerability_info IS '漏洞基本信息（代码审计/通用漏洞库）';
COMMENT ON COLUMN vulnerability_info.severity IS '严重级别: critical/high/medium/low';
COMMENT ON COLUMN vulnerability_info.status IS '状态: active/inactive';


-- =====================================================
-- 7. 镜像漏洞基本信息表 (image_vulnerability_info)
-- =====================================================
CREATE TABLE IF NOT EXISTS image_vulnerability_info (
    id              BIGSERIAL PRIMARY KEY,
    cve_id          VARCHAR(32),                                             -- CVE编号
    vuln_name       VARCHAR(255) NOT NULL,                                   -- 漏洞名称
    severity        VARCHAR(20) NOT NULL,                                    -- 严重级别: critical/high/medium/low
    cvss_score      DECIMAL(3,1),                                            -- CVSS评分
    description     TEXT,                                                    -- 漏洞描述
    fix_suggestion  TEXT,                                                    -- 修复建议
    reference       TEXT,                                                    -- 参考链接
    publish_date    TIMESTAMP,                                               -- 发布日期
    update_time     TIMESTAMP,                                               -- 更新时间
    status          VARCHAR(32) NOT NULL DEFAULT 'active',                   -- 状态: active/inactive
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ivi_cve_id ON image_vulnerability_info(cve_id);
CREATE INDEX IF NOT EXISTS idx_ivi_deleted_at ON image_vulnerability_info(deleted_at);

COMMENT ON TABLE image_vulnerability_info IS '镜像漏洞基本信息';
COMMENT ON COLUMN image_vulnerability_info.severity IS '严重级别: critical/high/medium/low';
COMMENT ON COLUMN image_vulnerability_info.status IS '状态: active/inactive';


-- =====================================================
-- 初始化完成
-- =====================================================
