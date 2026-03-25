-- 000010: 代码安全审计相关表
-- 包含: repos, repos_scan_result, codeql_rule, codeql_rules, codeql_scan_results, code_vuldetail (6 表)

-- 1. 代码仓库表 (repos)
CREATE TABLE IF NOT EXISTS repos (
    repo_id               BIGSERIAL PRIMARY KEY,
    repo_name             TEXT NOT NULL,
    repo_url              TEXT NOT NULL,
    language              TEXT,
    scan_frequency        TEXT,
    branch                TEXT,
    total_vulnerabilities BIGINT,
    critical_count        BIGINT,
    high_count            BIGINT,
    medium_count          BIGINT,
    low_count             BIGINT,
    status                TEXT DEFAULT 'PENDING',
    scan_start_time       TIMESTAMP(3),
    scan_end_time         TIMESTAMP(3),
    codeql_rules          TEXT,
    pull_method           TEXT,
    local_path            TEXT,
    is_private            BOOLEAN,
    description           TEXT,
    code_hash             TEXT,
    owner                 TEXT,
    last_scan_time        TIMESTAMP(3),
    deleted               BOOLEAN,
    created_at            TIMESTAMP(3),
    updated_at            TIMESTAMP(3),
    deleted_at            TIMESTAMP(3)
);

CREATE INDEX IF NOT EXISTS idx_repos_deleted_at ON repos(deleted_at);

COMMENT ON TABLE repos IS '代码仓库管理';

-- 2. 仓库扫描结果表 (repos_scan_result)
CREATE TABLE IF NOT EXISTS repos_scan_result (
    result_id             BIGSERIAL PRIMARY KEY,
    repo_name             VARCHAR(100) NOT NULL,
    repo_type             VARCHAR(50),
    rule_set_id           BIGINT,
    total_vulnerabilities BIGINT,
    critical_count        BIGINT,
    high_count            BIGINT,
    medium_count          BIGINT,
    low_count             BIGINT,
    scan_start_time       TIMESTAMP NOT NULL,
    scan_end_time         TIMESTAMP NOT NULL,
    highest_risk_level    VARCHAR(10),
    repo_status           VARCHAR(20),
    scan_report_url       VARCHAR(500),
    deleted               SMALLINT DEFAULT 0,
    update_time           TIMESTAMP DEFAULT NOW(),
    create_time           TIMESTAMP DEFAULT NOW(),
    created_at            TIMESTAMP(3),
    updated_at            TIMESTAMP(3),
    deleted_at            TIMESTAMP(3),
    CONSTRAINT chk_highest_risk_level CHECK (highest_risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    CONSTRAINT chk_repo_status CHECK (repo_status IN ('SAFE', 'WARNING', 'DANGEROUS'))
);

CREATE INDEX IF NOT EXISTS idx_repos_scan_result_deleted_at ON repos_scan_result(deleted_at);

COMMENT ON TABLE repos_scan_result IS '仓库扫描结果';

-- 3. CodeQL 规则表 (codeql_rule)
CREATE TABLE IF NOT EXISTS codeql_rule (
    rule_id                       BIGSERIAL PRIMARY KEY,
    enabled                       BOOLEAN,
    id                            TEXT UNIQUE,
    code                          VARCHAR(255),
    shortdescription_text         VARCHAR(255),
    fulldescription_text          TEXT,
    defaultconfiguration_enabled  VARCHAR(255),
    defaultconfiguration_level    VARCHAR(255),
    properties_tags               VARCHAR(255),
    properties_description        TEXT,
    properties_kind               VARCHAR(255),
    properties_precision          VARCHAR(255),
    properties_problem_severity   VARCHAR(255),
    properties_security_severity  VARCHAR(255),
    short_description_text        VARCHAR(255),
    full_description_text         TEXT,
    default_configuration_enabled VARCHAR(255),
    default_configuration_level   VARCHAR(255),
    deleted                       SMALLINT DEFAULT 0,
    create_time                   TIMESTAMP DEFAULT NOW(),
    update_time                   TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_codeql_rule_enabled ON codeql_rule(enabled);
CREATE INDEX IF NOT EXISTS idx_codeql_rule_deleted ON codeql_rule(deleted);

COMMENT ON TABLE codeql_rule IS 'CodeQL 单条规则';

-- 4. CodeQL 规则集表 (codeql_rules)
CREATE TABLE IF NOT EXISTS codeql_rules (
    rules_id         BIGSERIAL PRIMARY KEY,
    rule_name        TEXT NOT NULL,
    rule_count       BIGINT,
    rule_id          BIGINT,
    applicable_scene TEXT,
    risk_coverage    TEXT,
    total_rules      BIGINT,
    description      TEXT,
    status           TEXT,
    rule_ids         TEXT,
    deleted          SMALLINT,
    create_time      TIMESTAMP,
    update_time      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_codeql_rules_status ON codeql_rules(status);
CREATE INDEX IF NOT EXISTS idx_codeql_rules_deleted ON codeql_rules(deleted);

COMMENT ON TABLE codeql_rules IS 'CodeQL 规则集';

-- 5. CodeQL 扫描结果表 (codeql_scan_results)
CREATE TABLE IF NOT EXISTS codeql_scan_results (
    id              BIGSERIAL PRIMARY KEY,
    repo_id         BIGINT NOT NULL,
    repo_name       TEXT NOT NULL,
    rule_id         BIGINT,
    rule_name       TEXT,
    severity_string VARCHAR(255),
    severity        TEXT,
    file_path       TEXT,
    start_line      BIGINT,
    end_line        BIGINT,
    start_column    BIGINT,
    end_column      BIGINT,
    code_snippet    TEXT,
    message         TEXT,
    language        TEXT,
    code_flows      TEXT,
    related         TEXT,
    scan_time       TIMESTAMP(3),
    remediation     TEXT,
    confidence      TEXT,
    status          TEXT NOT NULL,
    fixed_time      TIMESTAMP,
    project_name    TEXT,
    branch          TEXT,
    hash            VARCHAR(255) UNIQUE,
    commit_id       TEXT,
    scan_type       TEXT NOT NULL,
    started_at      TIMESTAMP,
    finished_at     TIMESTAMP,
    result          TEXT,
    error_msg       TEXT,
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_codeql_scan_results_severity ON codeql_scan_results(severity);
CREATE INDEX IF NOT EXISTS idx_codeql_scan_results_rule_id ON codeql_scan_results(rule_id);
CREATE INDEX IF NOT EXISTS idx_codeql_scan_results_project_status ON codeql_scan_results(project_name, status);
CREATE INDEX IF NOT EXISTS idx_codeql_scan_results_deleted_at ON codeql_scan_results(deleted_at);

COMMENT ON TABLE codeql_scan_results IS 'CodeQL 扫描结果';

-- 6. 代码漏洞详情表 (code_vuldetail)
CREATE TABLE IF NOT EXISTS code_vuldetail (
    id              SERIAL PRIMARY KEY,
    scan_results_id INTEGER,
    path            VARCHAR(225),
    code            TEXT NOT NULL,
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP,
    deleted_at      TIMESTAMP,
    CONSTRAINT code_vuldetail_scan_results_id_path_key UNIQUE (scan_results_id, path)
);

COMMENT ON TABLE code_vuldetail IS '代码漏洞详情';
