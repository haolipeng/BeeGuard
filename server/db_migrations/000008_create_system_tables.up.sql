-- 000009: 系统管理相关表
-- 包含: systen_user, hids_rules (2 表)

-- 自动更新 updated_at 的触发器函数（如不存在则创建）
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 1. 系统用户表 (systen_user)
CREATE TABLE IF NOT EXISTS systen_user (
    id              SERIAL PRIMARY KEY,
    username        VARCHAR(250),
    name            VARCHAR(250),
    role            VARCHAR(100),
    account_status  VARCHAR(50),
    passwd          VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_app_user_account_status ON systen_user(account_status);
CREATE INDEX IF NOT EXISTS idx_app_user_login_name ON systen_user(name);

COMMENT ON TABLE systen_user IS '系统用户表';

-- 2. HIDS 检测规则表 (hids_rules)
CREATE TABLE IF NOT EXISTS hids_rules (
    id               SERIAL PRIMARY KEY,
    rule_name        VARCHAR(100) NOT NULL UNIQUE,
    rule_feature     TEXT NOT NULL,
    rule_level       VARCHAR(20) NOT NULL,
    trigger_action   TEXT NOT NULL,
    rule_status      VARCHAR(20) NOT NULL DEFAULT '未生效',
    effective_time   TIMESTAMPTZ,
    rule_description TEXT,
    ruler_type       VARCHAR(256),
    threat_type      VARCHAR(255),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT hids_rules_rule_level_check CHECK (rule_level IN ('低', '中', '高', '紧急')),
    CONSTRAINT hids_rules_rule_status_check CHECK (rule_status IN ('未生效', '生效中', '已停用', '已删除'))
);

CREATE INDEX IF NOT EXISTS idx_intrusion_rules_rule_level ON hids_rules(rule_level);
CREATE INDEX IF NOT EXISTS idx_intrusion_rules_rule_status ON hids_rules(rule_status);
CREATE INDEX IF NOT EXISTS idx_intrusion_rules_effective_time ON hids_rules(effective_time);

-- 自动更新 updated_at 触发器
CREATE TRIGGER trigger_intrusion_rules_updated_at
    BEFORE UPDATE ON hids_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE hids_rules IS 'HIDS 入侵检测规则';
COMMENT ON COLUMN hids_rules.rule_level IS '规则级别：低/中/高/紧急';
COMMENT ON COLUMN hids_rules.rule_status IS '规则状态：未生效/生效中/已停用/已删除';
COMMENT ON COLUMN hids_rules.ruler_type IS '规则分类:高危命令、反弹shell、本地提权、异常登录、密码破解、恶意请求、网络攻击、文件查杀、核心文件监控';
