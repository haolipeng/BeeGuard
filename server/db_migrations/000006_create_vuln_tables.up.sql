-- 000006: 漏洞发现表
-- 包含: vuln_info, host_vuln_scan_task, host_vuln_detail,
--       image_vuln_scan_task, image_vuln_detail,
--       vulnerability_info, image_vulnerability_info (7 表)

-- 1. 主机漏洞扫描任务表 (host_vuln_scan_task)
CREATE TABLE IF NOT EXISTS host_vuln_scan_task (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_id         BIGINT,
    host_name       VARCHAR(128) NOT NULL,
    host_ip         VARCHAR(256) NOT NULL,
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


-- 2. 漏洞信息表 (vuln_info)
CREATE TABLE IF NOT EXISTS vuln_info (
    id                  BIGSERIAL PRIMARY KEY,
    cve_id              VARCHAR(32) NOT NULL,
    vuln_name           VARCHAR(256) NOT NULL,
    severity            VARCHAR(16) NOT NULL,
    cvss_score          DECIMAL(3,1),
    description         TEXT,
    fix_suggestion      TEXT,
    reference_urls      TEXT,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vi_cve_id ON vuln_info(cve_id);
CREATE INDEX IF NOT EXISTS idx_vi_severity ON vuln_info(severity);
CREATE INDEX IF NOT EXISTS idx_vi_cvss_score ON vuln_info(cvss_score);

COMMENT ON TABLE vuln_info IS '漏洞发现-漏洞信息(主机/容器共用)';
COMMENT ON COLUMN vuln_info.severity IS '漏洞等级: critical/high/medium/low';
COMMENT ON COLUMN vuln_info.cvss_score IS 'CVSS评分(0.0-10.0)';


-- 3. 主机漏洞发现记录表 (host_vuln_detail)
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
    status              SMALLINT NOT NULL DEFAULT 0,
    host_name           VARCHAR(128),
    host_ip             VARCHAR(256),
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


-- 4. 镜像漏洞扫描任务表 (image_vuln_scan_task)
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


-- 5. 镜像漏洞发现记录表 (image_vuln_detail)
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


-- 6. 漏洞基本信息表 (vulnerability_info)
CREATE TABLE IF NOT EXISTS vulnerability_info (
    id              BIGSERIAL PRIMARY KEY,
    cve_id          VARCHAR(32),
    vuln_name       VARCHAR(255) NOT NULL,
    severity        VARCHAR(20) NOT NULL,
    cvss_score      DECIMAL(3,1),
    description     TEXT,
    fix_suggestion  TEXT,
    reference       TEXT,
    publish_date    TIMESTAMP,
    update_time     TIMESTAMP,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vi2_cve_id ON vulnerability_info(cve_id);
CREATE INDEX IF NOT EXISTS idx_vi2_deleted_at ON vulnerability_info(deleted_at);

COMMENT ON TABLE vulnerability_info IS '漏洞基本信息（代码审计/通用漏洞库）';
COMMENT ON COLUMN vulnerability_info.severity IS '严重级别: critical/high/medium/low';
COMMENT ON COLUMN vulnerability_info.status IS '状态: active/inactive';


-- 7. 镜像漏洞基本信息表 (image_vulnerability_info)
CREATE TABLE IF NOT EXISTS image_vulnerability_info (
    id              BIGSERIAL PRIMARY KEY,
    cve_id          VARCHAR(32),
    vuln_name       VARCHAR(255) NOT NULL,
    severity        VARCHAR(20) NOT NULL,
    cvss_score      DECIMAL(3,1),
    description     TEXT,
    fix_suggestion  TEXT,
    reference       TEXT,
    publish_date    TIMESTAMP,
    update_time     TIMESTAMP,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ivi_cve_id ON image_vulnerability_info(cve_id);
CREATE INDEX IF NOT EXISTS idx_ivi_deleted_at ON image_vulnerability_info(deleted_at);

COMMENT ON TABLE image_vulnerability_info IS '镜像漏洞基本信息';
COMMENT ON COLUMN image_vulnerability_info.severity IS '严重级别: critical/high/medium/low';
COMMENT ON COLUMN image_vulnerability_info.status IS '状态: active/inactive';
