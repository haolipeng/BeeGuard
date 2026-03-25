-- 000005: 合规基线表
-- 包含: baseline_template, baseline_template_host_link, baseline_check_item,
--       baseline_check_result, baseline_check_detail (5 表)

-- 1. 基线模板表 (baseline_template)
CREATE TABLE IF NOT EXISTS baseline_template (
    id              BIGSERIAL PRIMARY KEY,
    template_name   VARCHAR(128) NOT NULL,
    template_type   VARCHAR(32) NOT NULL,
    os_type         VARCHAR(32),
    version         VARCHAR(32),
    item_count      INT,
    description     VARCHAR(512),
    is_enabled      SMALLINT DEFAULT 1,
    baseline_ids    TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bt_template_type ON baseline_template(template_type);
CREATE INDEX IF NOT EXISTS idx_bt_os_type ON baseline_template(os_type);
CREATE INDEX IF NOT EXISTS idx_bt_is_enabled ON baseline_template(is_enabled);

COMMENT ON TABLE baseline_template IS '合规基线-基线模板';
COMMENT ON COLUMN baseline_template.template_type IS '基线类型: os_security/db_security/middleware_security';
COMMENT ON COLUMN baseline_template.os_type IS '操作系统类型: linux/windows';
COMMENT ON COLUMN baseline_template.is_enabled IS '是否启用: 0-禁用 1-启用';


-- 2. 基线模板与主机关联表 (baseline_template_host_link)
CREATE TABLE IF NOT EXISTS baseline_template_host_link (
    id                      BIGSERIAL PRIMARY KEY,
    template_id             BIGINT NOT NULL,
    target_range            TEXT NOT NULL,
    scan_frequency          VARCHAR(64) NOT NULL,
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    template_name           VARCHAR(128) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bthl_template_id ON baseline_template_host_link(template_id);

COMMENT ON TABLE baseline_template_host_link IS '合规基线-基线模板与主机关联';
COMMENT ON COLUMN baseline_template_host_link.template_id IS '关联���线模板ID';
COMMENT ON COLUMN baseline_template_host_link.target_range IS '目标范围（存储主机ID列表的JSON格式）';
COMMENT ON COLUMN baseline_template_host_link.scan_frequency IS '扫描频率';


-- 3. 基线检查项表 (baseline_check_item)
CREATE TABLE IF NOT EXISTS baseline_check_item (
    id              BIGSERIAL PRIMARY KEY,
    template_id     BIGINT NOT NULL,
    item_name       VARCHAR(256) NOT NULL,
    category        VARCHAR(64) NOT NULL,
    risk_level      VARCHAR(16) NOT NULL,
    check_rules     TEXT,
    fix_suggestion  TEXT,
    fix_script      TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bci_template_id ON baseline_check_item(template_id);
CREATE INDEX IF NOT EXISTS idx_bci_category ON baseline_check_item(category);
CREATE INDEX IF NOT EXISTS idx_bci_risk_level ON baseline_check_item(risk_level);

COMMENT ON TABLE baseline_check_item IS '合规基线-基线检查项';
COMMENT ON COLUMN baseline_check_item.template_id IS '关联基线模板ID(业务层关联baseline_template.id)';
COMMENT ON COLUMN baseline_check_item.risk_level IS '风险等级: high/medium/low';


-- 4. 检查结果表 (baseline_check_result)
CREATE TABLE IF NOT EXISTS baseline_check_result (
    id              BIGSERIAL PRIMARY KEY,
    baseline_id     VARCHAR(255) NOT NULL DEFAULT '',
    template_id     INTEGER,
    agent_id        VARCHAR(64) NOT NULL,
    host_ip         VARCHAR(256) NOT NULL,
    host_name       VARCHAR(128) NOT NULL,
    total_items     INT NOT NULL,
    passed_items    INT NOT NULL,
    failed_items    INT NOT NULL,
    error_items     INT NOT NULL DEFAULT 0,
    check_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bcr_baseline_id ON baseline_check_result(baseline_id);
CREATE INDEX IF NOT EXISTS idx_bcr_template_id ON baseline_check_result(template_id);
CREATE INDEX IF NOT EXISTS idx_bcr_agent_id ON baseline_check_result(agent_id);
CREATE INDEX IF NOT EXISTS idx_bcr_check_time ON baseline_check_result(check_time);

COMMENT ON TABLE baseline_check_result IS '合规基线-检查结果';
COMMENT ON COLUMN baseline_check_result.baseline_id IS '检测批次ID（前端task_id）';
COMMENT ON COLUMN baseline_check_result.template_id IS '关联基线模板ID';
COMMENT ON COLUMN baseline_check_result.agent_id IS 'Agent唯一标识';
COMMENT ON COLUMN baseline_check_result.error_items IS '检查异常项数';


-- 5. 检查明细表 (baseline_check_detail)
CREATE TABLE IF NOT EXISTS baseline_check_detail (
    id              BIGSERIAL NOT NULL,
    result_id       BIGINT NOT NULL,
    item_id         BIGINT NOT NULL,
    agent_id        VARCHAR(64) NOT NULL,
    status          SMALLINT NOT NULL,
    actual_value    TEXT,
    expected_value  TEXT,
    error_message   VARCHAR(512),
    check_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    host_ip         VARCHAR(256),
    host_name       VARCHAR(128),
    template_name   VARCHAR(128),
    baseline_id     VARCHAR(255) DEFAULT '',
    template_id     INTEGER NOT NULL,
    item_name       VARCHAR(128),
    risk_level      VARCHAR(255),
    PRIMARY KEY (id, template_id)
);

CREATE INDEX IF NOT EXISTS idx_bcd_result_id ON baseline_check_detail(result_id);
CREATE INDEX IF NOT EXISTS idx_bcd_item_id ON baseline_check_detail(item_id);
CREATE INDEX IF NOT EXISTS idx_bcd_baseline_id ON baseline_check_detail(baseline_id);
CREATE INDEX IF NOT EXISTS idx_bcd_template_id ON baseline_check_detail(template_id);
CREATE INDEX IF NOT EXISTS idx_bcd_agent_id ON baseline_check_detail(agent_id);
CREATE INDEX IF NOT EXISTS idx_bcd_status ON baseline_check_detail(status);

COMMENT ON TABLE baseline_check_detail IS '合规基线-检查明细';
COMMENT ON COLUMN baseline_check_detail.result_id IS '关联检查结果ID';
COMMENT ON COLUMN baseline_check_detail.item_id IS '关联检查项ID';
COMMENT ON COLUMN baseline_check_detail.baseline_id IS '检测批次ID（前端task_id）';
COMMENT ON COLUMN baseline_check_detail.host_ip IS '主机IP(冗余)';
COMMENT ON COLUMN baseline_check_detail.host_name IS '主机名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.template_name IS '模板名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.template_id IS '模板ID(冗余，用于快速查询和过滤)';
COMMENT ON COLUMN baseline_check_detail.item_name IS '检查项名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.risk_level IS '风险等级(冗余)';
COMMENT ON COLUMN baseline_check_detail.status IS '检查状态: 0-未通过 1-通过 2-检查异常';
