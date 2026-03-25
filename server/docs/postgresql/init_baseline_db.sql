-- =====================================================
-- SOC 合规基线数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- 说明: 合并自 migrations/014 的合规基线相关表
-- =====================================================


-- =====================================================
-- 1. 基线模板表 (baseline_template)
-- =====================================================
CREATE TABLE IF NOT EXISTS baseline_template (
    id              BIGSERIAL PRIMARY KEY,
    template_name   VARCHAR(128) NOT NULL,                                  -- 基线名称
    template_type   VARCHAR(32) NOT NULL,                                   -- 基线类型
    os_type         VARCHAR(32),                                            -- 操作系统类型
    version         VARCHAR(32),                                            -- 版本
    item_count      INT,                                                    -- 检查项数量
    description     VARCHAR(512),                                           -- 描述
    is_enabled      SMALLINT NOT NULL DEFAULT 1,                             -- 是否启用: 0-禁用 1-启用
    baseline_ids    TEXT,                                                    -- 基线ID列表
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


-- =====================================================
-- 1.5 基线模板与主机关联表 (baseline_template_host_link)
-- =====================================================
CREATE TABLE IF NOT EXISTS baseline_template_host_link (
    id                      BIGSERIAL PRIMARY KEY,
    baseline_template_id    BIGINT NOT NULL,                                    -- 基线模板ID
    baseline_template_name  VARCHAR(128) NOT NULL,                              -- 基线模板名称
    target_range            TEXT NOT NULL,                                      -- 目标范围（主机ID列表JSON）
    scan_frequency          VARCHAR(64) NOT NULL,                               -- 扫描频率
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bthl_template_id ON baseline_template_host_link(baseline_template_id);

COMMENT ON TABLE baseline_template_host_link IS '合规基线-基线模板与主机关联';
COMMENT ON COLUMN baseline_template_host_link.baseline_template_id IS '关联基线模板ID(业务层关联baseline_template.id)';
COMMENT ON COLUMN baseline_template_host_link.target_range IS '目标范围（存储主机ID列表的JSON格式）';
COMMENT ON COLUMN baseline_template_host_link.scan_frequency IS '扫描频率';


-- =====================================================
-- 2. 基线检查项表 (baseline_check_item)
-- =====================================================
CREATE TABLE IF NOT EXISTS baseline_check_item (
    id              BIGSERIAL PRIMARY KEY,
    template_id     BIGINT NOT NULL,                                        -- 关联基线模板ID
    item_name       VARCHAR(256) NOT NULL,                                  -- 检查项名称
    category        VARCHAR(64) NOT NULL,                                   -- 分类
    risk_level      VARCHAR(16) NOT NULL,                                   -- 风险等级
    check_rules     TEXT,                                                     -- 检查规则
    fix_suggestion  TEXT,                                                     -- 修复建议
    fix_script      TEXT,                                                     -- 修复脚本
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bci_template_id ON baseline_check_item(template_id);
CREATE INDEX IF NOT EXISTS idx_bci_category ON baseline_check_item(category);
CREATE INDEX IF NOT EXISTS idx_bci_risk_level ON baseline_check_item(risk_level);

COMMENT ON TABLE baseline_check_item IS '合规基线-基线检查项';
COMMENT ON COLUMN baseline_check_item.template_id IS '关联基线模板ID(业务层关联baseline_template.id)';
COMMENT ON COLUMN baseline_check_item.risk_level IS '风险等级: high/medium/low';


-- =====================================================
-- 3. 检查结果表 (baseline_check_result)
-- =====================================================
CREATE TABLE IF NOT EXISTS baseline_check_result (
    id              BIGSERIAL PRIMARY KEY,
    baseline_id     VARCHAR(255) NOT NULL DEFAULT '',                        -- 检测批次ID（前端task_id）
    template_id     INTEGER,                                                -- 关联基线模板ID
    agent_id        VARCHAR(64) NOT NULL,                                   -- Agent唯一标识
    host_ip         VARCHAR(45) NOT NULL,                                   -- 主机IP
    host_name       VARCHAR(128) NOT NULL,                                   -- 主机名
    total_items     INT NOT NULL,                                           -- 总检查项数
    passed_items    INT NOT NULL,                                           -- 通过项数
    failed_items    INT NOT NULL,                                           -- 未通过项数
    error_items     INT NOT NULL DEFAULT 0,                                 -- 检查异常项数
    check_time      TIMESTAMP NOT NULL,                                     -- 检查时间
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


-- =====================================================
-- 4. 检查明细表 (baseline_check_detail)
-- 范式化设计：通过 result_id 关联 baseline_check_result，通过 item_id 关联 baseline_check_item
-- 注意: 复合主键 (id, template_id)
-- =====================================================
CREATE TABLE IF NOT EXISTS baseline_check_detail (
    id              BIGSERIAL NOT NULL,
    result_id       BIGINT NOT NULL,                                        -- 关联检查结果ID
    item_id         BIGINT NOT NULL,                                        -- 关联检查项ID
    agent_id        VARCHAR(64) NOT NULL,                                   -- Agent唯一标识
    status          SMALLINT NOT NULL,                                      -- 检查状态
    actual_value    TEXT,                                                    -- 实际值
    expected_value  TEXT,                                                    -- 期望值
    error_message   VARCHAR(512),                                           -- 错误信息
    check_time      TIMESTAMP NOT NULL,                                     -- 检查时间
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    host_ip         VARCHAR(45),                                            -- 主机IP(冗余)
    host_name       VARCHAR(128),                                           -- 主机名称(冗余)
    template_name   VARCHAR(128),                                           -- 基线名称(冗余)
    baseline_id     VARCHAR(255) DEFAULT '',                                  -- 检测批次ID(冗余)
    template_id     INTEGER NOT NULL,                                       -- 模板ID(冗余，复合主键组成部分)
    item_name       VARCHAR(128),                                           -- 检查项名称(冗余)
    risk_level      VARCHAR(255),                                           -- 风险等级(冗余)
    PRIMARY KEY (id, template_id)
);

CREATE INDEX IF NOT EXISTS idx_bcd_result_id ON baseline_check_detail(result_id);
CREATE INDEX IF NOT EXISTS idx_bcd_item_id ON baseline_check_detail(item_id);
CREATE INDEX IF NOT EXISTS idx_bcd_baseline_id ON baseline_check_detail(baseline_id);
CREATE INDEX IF NOT EXISTS idx_bcd_agent_id ON baseline_check_detail(agent_id);
CREATE INDEX IF NOT EXISTS idx_bcd_status ON baseline_check_detail(status);

COMMENT ON TABLE baseline_check_detail IS '合规基线-检查明细';
COMMENT ON COLUMN baseline_check_detail.result_id IS '关联检查结果ID(业务层关联baseline_check_result.id)';
COMMENT ON COLUMN baseline_check_detail.item_id IS '关联检查项ID(业务层关联baseline_check_item.id)';
COMMENT ON COLUMN baseline_check_detail.baseline_id IS '检测批次ID(冗余)';
COMMENT ON COLUMN baseline_check_detail.item_name IS '检查项名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.host_ip IS '主机IP(冗余)';
COMMENT ON COLUMN baseline_check_detail.host_name IS '主机名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.template_name IS '基线名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.template_id IS '模板ID(冗余，复合主键组成部分)';
COMMENT ON COLUMN baseline_check_detail.risk_level IS '风险等级(冗余)';
COMMENT ON COLUMN baseline_check_detail.status IS '检查状态: 0-未通过 1-通过 2-检查异常';


-- =====================================================
-- 初始化完成
-- =====================================================
