-- 创建分析报告表
CREATE TABLE IF NOT EXISTS analysis_report (
    id              SERIAL PRIMARY KEY,
    analysis_type   VARCHAR(50) NOT NULL,      -- host/source_ip/single
    scope_key       VARCHAR(255) NOT NULL,     -- 主机IP/攻击源IP/告警ID
    alert_count     INTEGER NOT NULL DEFAULT 0,
    alert_snapshot  JSONB,                      -- 告警快照
    risk_level      VARCHAR(20),                -- low/medium/high/critical
    attack_pattern  TEXT,
    attack_stage    VARCHAR(100),
    summary         TEXT,
    recommendations JSONB,                      -- 建议列表
    ioc_indicators  JSONB,                      -- IOC指标
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_analysis_report_analysis_type ON analysis_report(analysis_type);
CREATE INDEX idx_analysis_report_scope_key ON analysis_report(scope_key);
CREATE INDEX idx_analysis_report_risk_level ON analysis_report(risk_level);
CREATE INDEX idx_analysis_report_created_at ON analysis_report(created_at DESC);

-- 添加注释
COMMENT ON TABLE analysis_report IS 'AI分析报告表';
COMMENT ON COLUMN analysis_report.analysis_type IS '分析类型: host(按主机)/source_ip(按攻击源)/single(单条告警)';
COMMENT ON COLUMN analysis_report.scope_key IS '分析范围标识: 主机IP/攻击源IP/告警ID';
COMMENT ON COLUMN analysis_report.alert_count IS '分析的告警数量';
COMMENT ON COLUMN analysis_report.alert_snapshot IS '告警快照(JSONB数组)';
COMMENT ON COLUMN analysis_report.risk_level IS '风险等级: low/medium/high/critical';
COMMENT ON COLUMN analysis_report.attack_pattern IS '攻击模式描述';
COMMENT ON COLUMN analysis_report.attack_stage IS '攻击阶段';
COMMENT ON COLUMN analysis_report.summary IS '分析摘要';
COMMENT ON COLUMN analysis_report.recommendations IS '处置建议(JSONB数组)';
COMMENT ON COLUMN analysis_report.ioc_indicators IS 'IOC指标(JSONB对象)';
