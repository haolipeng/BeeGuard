-- =====================================================
-- SOC 数据库重建脚本
-- 数据库: PostgreSQL (soc)
-- 说明: 删除所有视图和表，然后重新创建
-- 警告: 此脚本会清除所有数据！请确认后再执行
-- =====================================================

BEGIN;

-- =====================================================
-- 第一步: 删除所有视图
-- =====================================================
DROP VIEW IF EXISTS v_vuln_count_hosts CASCADE;
DROP VIEW IF EXISTS v_vuln_count_images CASCADE;
DROP VIEW IF EXISTS v_vuln_count_vuls CASCADE;
DROP VIEW IF EXISTS v_vuln_count_image_vuls CASCADE;
DROP VIEW IF EXISTS baseline_check_host_view CASCADE;
DROP VIEW IF EXISTS baseline_check_item_view CASCADE;


-- =====================================================
-- 第二步: 删除所有表 (CASCADE 处理外键依赖)
-- =====================================================

-- 漏洞相关表 (先删子表再删父表)
DROP TABLE IF EXISTS host_vuln_detail CASCADE;
DROP TABLE IF EXISTS image_vuln_detail CASCADE;
DROP TABLE IF EXISTS host_vuln_scan_task CASCADE;
DROP TABLE IF EXISTS image_vuln_scan_task CASCADE;
DROP TABLE IF EXISTS vuln_info CASCADE;
DROP TABLE IF EXISTS vulnerability_info CASCADE;
DROP TABLE IF EXISTS image_vulnerability_info CASCADE;

-- 告警相关表
DROP TABLE IF EXISTS alert_brute_force CASCADE;
DROP TABLE IF EXISTS alert_dangerous_command CASCADE;
DROP TABLE IF EXISTS alert_reverse_shell CASCADE;
DROP TABLE IF EXISTS alert_privilege_escalation CASCADE;
DROP TABLE IF EXISTS alert_abnormal_login CASCADE;
DROP TABLE IF EXISTS alert_malicious_request CASCADE;
DROP TABLE IF EXISTS alert_network_attack CASCADE;
DROP TABLE IF EXISTS alert_malware_scan CASCADE;
DROP TABLE IF EXISTS alert_file_integrity CASCADE;

-- 资产相关表
DROP TABLE IF EXISTS asset_host CASCADE;
DROP TABLE IF EXISTS asset_port CASCADE;
DROP TABLE IF EXISTS asset_account CASCADE;
DROP TABLE IF EXISTS asset_process CASCADE;
DROP TABLE IF EXISTS asset_database CASCADE;
DROP TABLE IF EXISTS asset_web_service CASCADE;
DROP TABLE IF EXISTS asset_system_service CASCADE;
DROP TABLE IF EXISTS asset_software CASCADE;
DROP TABLE IF EXISTS asset_container CASCADE;
DROP TABLE IF EXISTS asset_env_suspicious CASCADE;
DROP TABLE IF EXISTS asset_kmod CASCADE;
DROP TABLE IF EXISTS asset_image CASCADE;
DROP TABLE IF EXISTS asset_image_package CASCADE;

-- 事件相关表
DROP TABLE IF EXISTS event_dns CASCADE;
DROP TABLE IF EXISTS event_execve CASCADE;
DROP TABLE IF EXISTS event_connect CASCADE;
DROP TABLE IF EXISTS event_file CASCADE;

-- 基线相关表
DROP TABLE IF EXISTS baseline_check_detail CASCADE;
DROP TABLE IF EXISTS baseline_check_result CASCADE;
DROP TABLE IF EXISTS baseline_check_item CASCADE;
DROP TABLE IF EXISTS baseline_template_host_link CASCADE;
DROP TABLE IF EXISTS baseline_template CASCADE;

-- Agent相关表
DROP TABLE IF EXISTS agent_info CASCADE;


-- =====================================================
-- 第三步: 重新创建所有表
-- =====================================================


-- =============================================================================
-- 3.1 Agent客户端管理表
-- =============================================================================

CREATE TABLE IF NOT EXISTS agent_info (
    id                  BIGSERIAL PRIMARY KEY,
    agent_id            VARCHAR(64) NOT NULL,
    agent_version       VARCHAR(32),
    connection_status   SMALLINT NOT NULL DEFAULT 0,
    host_name           VARCHAR(128) NOT NULL,
    host_ip             VARCHAR(45) NOT NULL,
    os_type             VARCHAR(16) NOT NULL,
    os_version          VARCHAR(128),
    os_arch             VARCHAR(32),
    cpu_count           INT,
    memory_total        BIGINT,
    disk_total          BIGINT,
    last_connected_at   TIMESTAMP,
    registered_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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


-- =============================================================================
-- 3.2 资产管理表
-- =============================================================================

-- 1. 主机列表表 (asset_host)
CREATE TABLE IF NOT EXISTS asset_host (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    mac_addr        VARCHAR(45),
    os_type         VARCHAR(32),
    os_version      VARCHAR(64),
    agent_status    SMALLINT        DEFAULT 0,
    agent_version   VARCHAR(32),
    last_heartbeat  TIMESTAMP,
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_host_agent_id ON asset_host(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_host_host_ip ON asset_host(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_host_agent_status ON asset_host(agent_status);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_host_agent_id ON asset_host(agent_id);

COMMENT ON TABLE asset_host IS '资产管理-主机列表';
COMMENT ON COLUMN asset_host.agent_id IS 'Agent唯一标识(安装时生成，不随IP/hostname变化)';
COMMENT ON COLUMN asset_host.agent_status IS 'Agent状态: 0=离线 1=在线';


-- 2. 端口列表表 (asset_port)
CREATE TABLE IF NOT EXISTS asset_port (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    port            INTEGER         NOT NULL,
    protocol        SMALLINT        NOT NULL,
    listen_ip       VARCHAR(45)     NOT NULL,
    listen_process  VARCHAR(45)     NOT NULL,
    run_user        VARCHAR(64),
    os_version      VARCHAR(64),
    agent_status    SMALLINT        DEFAULT 0,
    agent_version   VARCHAR(32),
    process_time    TIMESTAMP,
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_port_agent_id ON asset_port(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_port_host_ip ON asset_port(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_port_port ON asset_port(port);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_port_agent_port ON asset_port(agent_id, port, protocol);

COMMENT ON TABLE asset_port IS '资产管理-端口列表';
COMMENT ON COLUMN asset_port.port IS '监听端口';
COMMENT ON COLUMN asset_port.protocol IS '端口协议: 6=TCP, 17=UDP';


-- 3. 账号列表表 (asset_account)
CREATE TABLE IF NOT EXISTS asset_account (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    name            VARCHAR(128)    NOT NULL,
    uid             INTEGER         NOT NULL,
    status          SMALLINT        NOT NULL DEFAULT 0,
    permission      VARCHAR(64)     NOT NULL,
    login_type      VARCHAR(128),
    last_login_time TIMESTAMP,
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_account_agent_id ON asset_account(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_account_host_ip ON asset_account(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_account_name ON asset_account(name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_account_agent_user ON asset_account(agent_id, name);

COMMENT ON TABLE asset_account IS '资产管理-账号列表';
COMMENT ON COLUMN asset_account.name IS '系统账号名称';
COMMENT ON COLUMN asset_account.status IS '账号状态: 0=正常 1=即将过期 2=已过期';
COMMENT ON COLUMN asset_account.permission IS '权限: normal(普通用户)、root、sudo、root,sudo';
COMMENT ON COLUMN asset_account.login_type IS '登录Shell: /bin/bash、/bin/sh、/sbin/nologin等';


-- 4. 进程列表表 (asset_process)
CREATE TABLE IF NOT EXISTS asset_process (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    name            VARCHAR(128)    NOT NULL,
    status          VARCHAR(64),
    version         VARCHAR(64),
    path            VARCHAR(512)    NOT NULL,
    run_name        VARCHAR(128)    NOT NULL,
    start_time      TIMESTAMP,
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_process_agent_id ON asset_process(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_process_host_ip ON asset_process(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_process_name ON asset_process(name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_process_agent_path ON asset_process(agent_id, path);

COMMENT ON TABLE asset_process IS '资产管理-进程列表';
COMMENT ON COLUMN asset_process.name IS '进程名称';
COMMENT ON COLUMN asset_process.run_name IS '运行用户';


-- 5. 数据库列表表 (asset_database)
CREATE TABLE IF NOT EXISTS asset_database (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    db_type         VARCHAR(45)     NOT NULL,
    db_version      VARCHAR(45)     NOT NULL,
    port            INTEGER         NOT NULL,
    run_user        VARCHAR(64),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_database_agent_id ON asset_database(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_database_host_ip ON asset_database(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_database_db_type ON asset_database(db_type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_database_agent_type ON asset_database(agent_id, db_type);

COMMENT ON TABLE asset_database IS '资产管理-数据库列表';
COMMENT ON COLUMN asset_database.db_type IS '数据库类型(MySQL/PostgreSQL/Oracle等)';
COMMENT ON COLUMN asset_database.db_version IS '数据库版本';


-- 6. Web服务表 (asset_web_service)
CREATE TABLE IF NOT EXISTS asset_web_service (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    name            VARCHAR(128)    NOT NULL,
    version         VARCHAR(64)     NOT NULL,
    server_type     VARCHAR(64)     NOT NULL,
    site_domain     VARCHAR(255),
    path            VARCHAR(512),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_web_service_agent_id ON asset_web_service(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_web_service_host_ip ON asset_web_service(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_web_service_server_type ON asset_web_service(server_type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_web_service_agent_type ON asset_web_service(agent_id, server_type);

COMMENT ON TABLE asset_web_service IS '资产管理-Web服务';
COMMENT ON COLUMN asset_web_service.name IS '应用名称';
COMMENT ON COLUMN asset_web_service.server_type IS '服务器类型(Nginx/Apache/Tomcat等)';


-- 7. 系统服务表 (asset_system_service)
CREATE TABLE IF NOT EXISTS asset_system_service (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    name            VARCHAR(255)    NOT NULL,
    version         VARCHAR(64),
    status          VARCHAR(64)     NOT NULL,
    run_user        VARCHAR(255)    NOT NULL,
    path            VARCHAR(512)    NOT NULL,
    describe        VARCHAR(512),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_system_service_agent_id ON asset_system_service(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_system_service_host_ip ON asset_system_service(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_system_service_name ON asset_system_service(name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_system_service_agent_name ON asset_system_service(agent_id, name);

COMMENT ON TABLE asset_system_service IS '资产管理-系统服务';
COMMENT ON COLUMN asset_system_service.name IS '服务名称';
COMMENT ON COLUMN asset_system_service.status IS '服务状态';


-- 8. 软件列表表 (asset_software)
CREATE TABLE IF NOT EXISTS asset_software (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    name            VARCHAR(255)    NOT NULL,
    version         VARCHAR(128),
    type            VARCHAR(32)     NOT NULL,
    source          VARCHAR(255),
    status          VARCHAR(64),
    vendor          VARCHAR(255),
    path            VARCHAR(512),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_software_agent_id ON asset_software(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_software_host_ip ON asset_software(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_software_name ON asset_software(name);
CREATE INDEX IF NOT EXISTS idx_asset_software_type ON asset_software(type);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_software_agent_name_type ON asset_software(agent_id, name, type);

COMMENT ON TABLE asset_software IS '资产管理-软件列表';
COMMENT ON COLUMN asset_software.name IS '软件名称';
COMMENT ON COLUMN asset_software.type IS '软件类型(dpkg/rpm/pypi/jar)';


-- 9. 容器列表表 (asset_container)
CREATE TABLE IF NOT EXISTS asset_container (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    container_id    VARCHAR(128)    NOT NULL,
    name            VARCHAR(255)    NOT NULL,
    state           VARCHAR(32)     NOT NULL,
    image_id        VARCHAR(128),
    image_name      VARCHAR(255),
    runtime         VARCHAR(32),
    pid             VARCHAR(16),
    create_time     VARCHAR(32),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_container_agent_id ON asset_container(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_container_host_ip ON asset_container(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_container_name ON asset_container(name);
CREATE INDEX IF NOT EXISTS idx_asset_container_state ON asset_container(state);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_container_agent_cid ON asset_container(agent_id, container_id);

COMMENT ON TABLE asset_container IS '资产管理-容器列表';
COMMENT ON COLUMN asset_container.container_id IS '容器ID';
COMMENT ON COLUMN asset_container.state IS '容器状态(running/exited/created等)';
COMMENT ON COLUMN asset_container.runtime IS '容器运行时(docker/containerd)';


-- 10. 可疑环境变量表 (asset_env_suspicious)
CREATE TABLE IF NOT EXISTS asset_env_suspicious (
    id                  BIGSERIAL       PRIMARY KEY,
    agent_id            VARCHAR(64)     NOT NULL,
    host_name           VARCHAR(128)    NOT NULL,
    host_ip             VARCHAR(45)     NOT NULL,
    var_name            VARCHAR(255)    NOT NULL,
    var_value           TEXT,
    suspicious_reasons  TEXT,
    source              VARCHAR(128),
    created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_env_suspicious_agent_id ON asset_env_suspicious(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_env_suspicious_host_ip ON asset_env_suspicious(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_env_suspicious_var_name ON asset_env_suspicious(var_name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_env_suspicious_agent_var ON asset_env_suspicious(agent_id, var_name);

COMMENT ON TABLE asset_env_suspicious IS '资产管理-可疑环境变量';
COMMENT ON COLUMN asset_env_suspicious.var_name IS '环境变量名称';
COMMENT ON COLUMN asset_env_suspicious.suspicious_reasons IS '可疑原因';


-- 11. 内核模块表 (asset_kmod)
CREATE TABLE IF NOT EXISTS asset_kmod (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    os_type         VARCHAR(32),
    name            VARCHAR(128)    NOT NULL,
    size            VARCHAR(32),
    refcount        VARCHAR(16),
    used_by         VARCHAR(512),
    state           VARCHAR(32),
    addr            VARCHAR(32),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_kmod_agent_id ON asset_kmod(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_kmod_host_ip ON asset_kmod(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_kmod_name ON asset_kmod(name);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_kmod_agent_name ON asset_kmod(agent_id, name);

COMMENT ON TABLE asset_kmod IS '资产管理-内核模块';
COMMENT ON COLUMN asset_kmod.name IS '内核模块名称';
COMMENT ON COLUMN asset_kmod.state IS '模块状态(Live/Loading/Unloading)';


-- 12. 镜像列表表 (asset_image)
CREATE TABLE IF NOT EXISTS asset_image (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    image_id        VARCHAR(128)    NOT NULL,
    image_name      VARCHAR(255)    NOT NULL,
    image_version   VARCHAR(128),
    image_size      BIGINT,
    container_count INTEGER         DEFAULT 0,
    build_time      TIMESTAMP,
    runtime         VARCHAR(32),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_image_agent_id ON asset_image(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_image_host_ip ON asset_image(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_image_image_name ON asset_image(image_name);
CREATE INDEX IF NOT EXISTS idx_asset_image_image_version ON asset_image(image_version);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_image_agent_imgid ON asset_image(agent_id, image_id);

COMMENT ON TABLE asset_image IS '资产管理-镜像列表';
COMMENT ON COLUMN asset_image.image_id IS '镜像ID(sha256格式)';
COMMENT ON COLUMN asset_image.image_name IS '镜像名称';
COMMENT ON COLUMN asset_image.image_version IS '镜像版本/标签';
COMMENT ON COLUMN asset_image.image_size IS '镜像大小(字节)';
COMMENT ON COLUMN asset_image.container_count IS '关联容器数量';
COMMENT ON COLUMN asset_image.build_time IS '镜像构建时间';
COMMENT ON COLUMN asset_image.runtime IS '容器运行时(docker/containerd)';


-- 13. 镜像软件包表 (asset_image_package)
CREATE TABLE IF NOT EXISTS asset_image_package (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(45)     NOT NULL,
    image_id        VARCHAR(128)    NOT NULL,
    image_name      VARCHAR(255)    NOT NULL,
    package_name    VARCHAR(255)    NOT NULL,
    package_version VARCHAR(128),
    package_type    VARCHAR(32)     NOT NULL,
    os_version      VARCHAR(64),
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_imgpkg_agent_id ON asset_image_package(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_imgpkg_image_id ON asset_image_package(image_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_imgpkg_composite ON asset_image_package(agent_id, image_id, package_name);

COMMENT ON TABLE asset_image_package IS '资产管理-镜像软件包';
COMMENT ON COLUMN asset_image_package.package_type IS '软件包类型(dpkg/rpm/apk)';


-- =============================================================================
-- 3.3 入侵检测告警表
-- =============================================================================

-- 1. 暴力破解告警表 (alert_brute_force)
CREATE TABLE IF NOT EXISTS alert_brute_force (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(45) NOT NULL,
    source_ip         VARCHAR(45) NOT NULL,
    source_location   VARCHAR(128),
    attack_type       VARCHAR(32) NOT NULL,
    target_ip         VARCHAR(45) NOT NULL,
    target_port       INT,
    username          VARCHAR(64) NOT NULL,
    attempt_count     INT NOT NULL,
    attack_time       TIMESTAMP NOT NULL,
    first_attack_time TIMESTAMP,
    status            SMALLINT NOT NULL DEFAULT 0,
    is_blocked        SMALLINT DEFAULT 0,
    process_time      TIMESTAMP,
    processor         VARCHAR(64),
    remark            VARCHAR(512),
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_bf_agent_id ON alert_brute_force(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_bf_source_ip ON alert_brute_force(source_ip);
CREATE INDEX IF NOT EXISTS idx_alert_bf_attack_type ON alert_brute_force(attack_type);
CREATE INDEX IF NOT EXISTS idx_alert_bf_status ON alert_brute_force(status);
CREATE INDEX IF NOT EXISTS idx_alert_bf_attack_time ON alert_brute_force(attack_time);

COMMENT ON TABLE alert_brute_force IS '入侵检测-暴力破解告警';
COMMENT ON COLUMN alert_brute_force.attack_type IS '攻击类型: ssh/ftp/rdp/mysql/redis/web_login';
COMMENT ON COLUMN alert_brute_force.status IS '状态: 0-待处理 1-已处理 2-已忽略';


-- 2. 高危命令告警表 (alert_dangerous_command)
CREATE TABLE IF NOT EXISTS alert_dangerous_command (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    host_ip           VARCHAR(45) NOT NULL,
    command           TEXT NOT NULL,
    command_type      VARCHAR(32) NOT NULL,
    "user"            VARCHAR(64) NOT NULL,
    privilege_level   VARCHAR(32) NOT NULL,
    status            SMALLINT NOT NULL DEFAULT 0,
    alert_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_cmd_agent_id ON alert_dangerous_command(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_command_type ON alert_dangerous_command(command_type);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_status ON alert_dangerous_command(status);
CREATE INDEX IF NOT EXISTS idx_alert_cmd_alert_time ON alert_dangerous_command(alert_time);

COMMENT ON TABLE alert_dangerous_command IS '入侵检测-高危命令告警';
COMMENT ON COLUMN alert_dangerous_command.command_type IS '命令类型: file_delete/privilege_escalation/permission_modify/filesystem_operation/network_scan/data_exfiltration/service_stop/log_tamper';


-- 3. 反弹Shell告警表 (alert_reverse_shell)
CREATE TABLE IF NOT EXISTS alert_reverse_shell (
    id                BIGSERIAL PRIMARY KEY,
    agent_id          VARCHAR(64) NOT NULL,
    host_id           BIGINT,
    host_name         VARCHAR(128) NOT NULL,
    victim_ip         VARCHAR(45) NOT NULL,
    command_line      TEXT NOT NULL,
    shell_type        VARCHAR(32),
    target_host       VARCHAR(45) NOT NULL,
    target_port       INT NOT NULL,
    status            SMALLINT NOT NULL DEFAULT 0,
    event_time        TIMESTAMP NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_rs_agent_id ON alert_reverse_shell(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_rs_shell_type ON alert_reverse_shell(shell_type);
CREATE INDEX IF NOT EXISTS idx_alert_rs_target_host ON alert_reverse_shell(target_host);
CREATE INDEX IF NOT EXISTS idx_alert_rs_status ON alert_reverse_shell(status);
CREATE INDEX IF NOT EXISTS idx_alert_rs_event_time ON alert_reverse_shell(event_time);

COMMENT ON TABLE alert_reverse_shell IS '入侵检测-反弹Shell告警';
COMMENT ON COLUMN alert_reverse_shell.shell_type IS 'Shell类型: bash/python/nc/perl/php/ruby/powershell';


-- 4. 本地提权告警表 (alert_privilege_escalation)
CREATE TABLE IF NOT EXISTS alert_privilege_escalation (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,
    escalated_user        VARCHAR(64) NOT NULL,
    parent_process        VARCHAR(256) NOT NULL,
    parent_process_user   VARCHAR(64) NOT NULL,
    process_id            INT,
    process_path          VARCHAR(512),
    status                SMALLINT NOT NULL DEFAULT 0,
    discover_time         TIMESTAMP NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_pe_agent_id ON alert_privilege_escalation(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_pe_escalated_user ON alert_privilege_escalation(escalated_user);
CREATE INDEX IF NOT EXISTS idx_alert_pe_status ON alert_privilege_escalation(status);
CREATE INDEX IF NOT EXISTS idx_alert_pe_discover_time ON alert_privilege_escalation(discover_time);

COMMENT ON TABLE alert_privilege_escalation IS '入侵检测-本地提权告警';
COMMENT ON COLUMN alert_privilege_escalation.escalated_user IS '提权后的用户(通常为root)';


-- 5. 异常登录告警表 (alert_abnormal_login)
CREATE TABLE IF NOT EXISTS alert_abnormal_login (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,
    source_ip             VARCHAR(45) NOT NULL,
    source_location       VARCHAR(128),
    source_country        VARCHAR(64),
    source_city           VARCHAR(64),
    login_user            VARCHAR(64) NOT NULL,
    login_time            TIMESTAMP NOT NULL,
    risk_level            VARCHAR(16) NOT NULL,
    abnormal_type         VARCHAR(32),
    status                SMALLINT NOT NULL DEFAULT 0,
    is_whitelist          SMALLINT DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_al_agent_id ON alert_abnormal_login(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_al_source_ip ON alert_abnormal_login(source_ip);
CREATE INDEX IF NOT EXISTS idx_alert_al_login_user ON alert_abnormal_login(login_user);
CREATE INDEX IF NOT EXISTS idx_alert_al_abnormal_type ON alert_abnormal_login(abnormal_type);
CREATE INDEX IF NOT EXISTS idx_alert_al_status ON alert_abnormal_login(status);
CREATE INDEX IF NOT EXISTS idx_alert_al_login_time ON alert_abnormal_login(login_time);

COMMENT ON TABLE alert_abnormal_login IS '入侵检测-异常登录告警';
COMMENT ON COLUMN alert_abnormal_login.abnormal_type IS '异常类型: abnormal_location/abnormal_time/abnormal_user';
COMMENT ON COLUMN alert_abnormal_login.risk_level IS '危险等级: low/medium/high';


-- 6. 恶意请求告警表 (alert_malicious_request)
CREATE TABLE IF NOT EXISTS alert_malicious_request (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,
    policy_type           VARCHAR(32) NOT NULL,
    policy_name           VARCHAR(128) NOT NULL,
    malicious_domain      VARCHAR(256) NOT NULL,
    malicious_ip          VARCHAR(45),
    request_count         INT NOT NULL,
    first_request_time    TIMESTAMP,
    last_request_time     TIMESTAMP,
    risk_description      TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_mr_agent_id ON alert_malicious_request(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_mr_policy_type ON alert_malicious_request(policy_type);
CREATE INDEX IF NOT EXISTS idx_alert_mr_malicious_domain ON alert_malicious_request(malicious_domain);
CREATE INDEX IF NOT EXISTS idx_alert_mr_status ON alert_malicious_request(status);
CREATE INDEX IF NOT EXISTS idx_alert_mr_last_request_time ON alert_malicious_request(last_request_time);

COMMENT ON TABLE alert_malicious_request IS '入侵检测-恶意请求告警';
COMMENT ON COLUMN alert_malicious_request.policy_type IS '策略类型: mining/c2/phishing/botnet/ransomware';


-- 7. 网络攻击告警表 (alert_network_attack)
CREATE TABLE IF NOT EXISTS alert_network_attack (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,
    target_port           INT NOT NULL,
    attacker_ip           VARCHAR(45) NOT NULL,
    attacker_location     VARCHAR(128),
    attacker_country      VARCHAR(64),
    vulnerability_name    VARCHAR(256) NOT NULL,
    vulnerability_id      VARCHAR(64),
    attack_status         VARCHAR(32) NOT NULL,
    attack_count          INT NOT NULL,
    first_attack_time     TIMESTAMP,
    last_attack_time      TIMESTAMP NOT NULL,
    attack_payload        TEXT,
    status                SMALLINT NOT NULL DEFAULT 0,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_na_agent_id ON alert_network_attack(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_na_attacker_ip ON alert_network_attack(attacker_ip);
CREATE INDEX IF NOT EXISTS idx_alert_na_vulnerability_id ON alert_network_attack(vulnerability_id);
CREATE INDEX IF NOT EXISTS idx_alert_na_attack_status ON alert_network_attack(attack_status);
CREATE INDEX IF NOT EXISTS idx_alert_na_status ON alert_network_attack(status);
CREATE INDEX IF NOT EXISTS idx_alert_na_last_attack_time ON alert_network_attack(last_attack_time);

COMMENT ON TABLE alert_network_attack IS '入侵检测-网络攻击告警';
COMMENT ON COLUMN alert_network_attack.vulnerability_id IS '漏洞编号(如CVE-2021-44228)';


-- 8. 文件查杀告警表 (alert_malware_scan)
CREATE TABLE IF NOT EXISTS alert_malware_scan (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_ip               VARCHAR(45) NOT NULL,
    host_name             VARCHAR(128) NOT NULL,
    threat_type           VARCHAR(64) NOT NULL,
    file_name             VARCHAR(256) NOT NULL,
    file_path             VARCHAR(512) NOT NULL,
    file_size             BIGINT,
    file_md5              VARCHAR(32),
    file_sha256           VARCHAR(128),
    detection_engine      VARCHAR(64),
    malware_family        VARCHAR(64),
    is_quarantined        SMALLINT DEFAULT 0,
    is_deleted            SMALLINT DEFAULT 0,
    status                SMALLINT NOT NULL DEFAULT 0,
    scan_time             TIMESTAMP NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_ms_agent_id ON alert_malware_scan(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_ms_threat_type ON alert_malware_scan(threat_type);
CREATE INDEX IF NOT EXISTS idx_alert_ms_file_md5 ON alert_malware_scan(file_md5);
CREATE INDEX IF NOT EXISTS idx_alert_ms_malware_family ON alert_malware_scan(malware_family);
CREATE INDEX IF NOT EXISTS idx_alert_ms_status ON alert_malware_scan(status);
CREATE INDEX IF NOT EXISTS idx_alert_ms_scan_time ON alert_malware_scan(scan_time);

COMMENT ON TABLE alert_malware_scan IS '入侵检测-文件查杀告警';
COMMENT ON COLUMN alert_malware_scan.threat_type IS '威胁类型: virus/trojan/webshell/backdoor/ransomware/miner/rootkit';


-- 9. 核心文件监控告警表 (alert_file_integrity)
CREATE TABLE IF NOT EXISTS alert_file_integrity (
    id                    BIGSERIAL PRIMARY KEY,
    agent_id              VARCHAR(64) NOT NULL,
    host_id               BIGINT,
    host_name             VARCHAR(128) NOT NULL,
    host_ip               VARCHAR(45) NOT NULL,

    rule_type             VARCHAR(32) NOT NULL,                       -- 规则类型
    rule_name             VARCHAR(128) NOT NULL,                      -- 命中规则名称
    rule_id               BIGINT,                                     -- 关联规则ID
    threat_level          VARCHAR(16) NOT NULL,                       -- 威胁等级: low/medium/high
    threat_action         VARCHAR(32) NOT NULL,                       -- 威胁行为: add/modify/delete
    file_path             VARCHAR(512) NOT NULL,                      -- 文件路径
    file_name             VARCHAR(256),                               -- 文件名
    old_content_hash      VARCHAR(64),                                -- 原内容哈希
    new_content_hash      VARCHAR(64),                                -- 新内容哈希
    change_detail         TEXT,                                       -- 变更详情
    operator_user         VARCHAR(64),                                -- 操作用户
    operator_process      VARCHAR(256),                               -- 操作进程
    alert_description     TEXT,                                       -- 告警描述

    status                SMALLINT NOT NULL DEFAULT 0,                -- 0-待处理 1-已处理 2-已忽略
    alert_time            TIMESTAMP NOT NULL,                         -- 告警时间
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_fi_agent_id ON alert_file_integrity(agent_id);
CREATE INDEX IF NOT EXISTS idx_alert_fi_rule_type ON alert_file_integrity(rule_type);
CREATE INDEX IF NOT EXISTS idx_alert_fi_threat_level ON alert_file_integrity(threat_level);
CREATE INDEX IF NOT EXISTS idx_alert_fi_file_path ON alert_file_integrity(file_path);
CREATE INDEX IF NOT EXISTS idx_alert_fi_status ON alert_file_integrity(status);
CREATE INDEX IF NOT EXISTS idx_alert_fi_alert_time ON alert_file_integrity(alert_time);

COMMENT ON TABLE alert_file_integrity IS '入侵检测-核心文件监控告警';
COMMENT ON COLUMN alert_file_integrity.threat_level IS '威胁等级: low/medium/high';
COMMENT ON COLUMN alert_file_integrity.threat_action IS '威胁行为: add/modify/delete';


-- =============================================================================
-- 3.4 事件采集表
-- =============================================================================

-- 1. DNS查询事件表 (event_dns)
CREATE TABLE IF NOT EXISTS event_dns (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(45),
    pid             INT NOT NULL,
    tgid            INT,
    ppid            INT,
    uid             INT,
    comm            VARCHAR(16),
    exe_path        VARCHAR(512),
    domain          VARCHAR(255),
    query_type      VARCHAR(16),
    pid_tree        TEXT,
    event_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_dns_agent_id ON event_dns(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_dns_event_time ON event_dns(event_time);
CREATE INDEX IF NOT EXISTS idx_event_dns_domain ON event_dns(domain);
CREATE INDEX IF NOT EXISTS idx_event_dns_comm ON event_dns(comm);
CREATE INDEX IF NOT EXISTS idx_event_dns_exe_path ON event_dns(exe_path);

COMMENT ON TABLE event_dns IS '事件采集-DNS查询事件';
COMMENT ON COLUMN event_dns.query_type IS '查询类型: A/AAAA/CNAME/MX/TXT等';


-- 2. 进程执行事件表 (event_execve)
CREATE TABLE IF NOT EXISTS event_execve (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(45),
    pid             INT NOT NULL,
    tgid            INT,
    ppid            INT,
    pgid            INT,
    uid             INT,
    comm            VARCHAR(16),
    exe_path        VARCHAR(512),
    args            TEXT,
    event_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_execve_agent_id ON event_execve(agent_id);
CREATE INDEX IF NOT EXISTS idx_execve_event_time ON event_execve(event_time);
CREATE INDEX IF NOT EXISTS idx_execve_exe_path ON event_execve(exe_path);
CREATE INDEX IF NOT EXISTS idx_execve_ppid ON event_execve(ppid);
CREATE INDEX IF NOT EXISTS idx_execve_comm ON event_execve(comm);

COMMENT ON TABLE event_execve IS '事件采集-进程执行事件';


-- 3. 出站连接事件表 (event_connect)
CREATE TABLE IF NOT EXISTS event_connect (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        VARCHAR(64) NOT NULL,
    host_name       VARCHAR(128),
    host_ip         VARCHAR(45),
    pid             INT NOT NULL,
    tgid            INT,
    ppid            INT,
    uid             INT,
    comm            VARCHAR(16),
    exe_path        VARCHAR(512),
    protocol        VARCHAR(16),
    remote_ip       VARCHAR(64),
    remote_port     INT,
    pid_tree        TEXT,
    event_time      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_event_connect_agent_id ON event_connect(agent_id);
CREATE INDEX IF NOT EXISTS idx_event_connect_event_time ON event_connect(event_time);
CREATE INDEX IF NOT EXISTS idx_event_connect_remote_ip ON event_connect(remote_ip);
CREATE INDEX IF NOT EXISTS idx_event_connect_comm ON event_connect(comm);
CREATE INDEX IF NOT EXISTS idx_event_connect_exe_path ON event_connect(exe_path);

COMMENT ON TABLE event_connect IS '事件采集-出站连接事件';
COMMENT ON COLUMN event_connect.protocol IS '协议类型: tcp/udp';


-- 4. 文件操作事件表 (event_file)
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


-- =============================================================================
-- 3.5 漏洞发现表
-- =============================================================================

-- 1. 主机漏洞扫描任务表 (host_vuln_scan_task)
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


-- 2. 漏洞信息表 (vuln_info) - 主机/容器共用
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


-- =============================================================================
-- 3.6 合规基线表
-- =============================================================================

-- 1. 基线模板表 (baseline_template)
CREATE TABLE IF NOT EXISTS baseline_template (
    id              BIGSERIAL PRIMARY KEY,
    template_name   VARCHAR(128) NOT NULL,
    template_type   VARCHAR(32) NOT NULL,
    os_type         VARCHAR(32),
    version         VARCHAR(32),
    item_count      INT,
    description     VARCHAR(512),
    is_enabled      SMALLINT NOT NULL DEFAULT 1,
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
    baseline_template_id    BIGINT NOT NULL,
    baseline_template_name  VARCHAR(128) NOT NULL,
    target_range            TEXT NOT NULL,
    scan_frequency          VARCHAR(64) NOT NULL,
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bthl_template_id ON baseline_template_host_link(baseline_template_id);

COMMENT ON TABLE baseline_template_host_link IS '合规基线-基线模板与主机关联';
COMMENT ON COLUMN baseline_template_host_link.baseline_template_id IS '关联基线模板ID';
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
    host_ip         VARCHAR(45) NOT NULL,
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
-- 注意: 复合主键 (id, template_id)
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
    host_ip         VARCHAR(45),                                            -- 主机IP(冗余)
    host_name       VARCHAR(128),                                           -- 主机名称(冗余)
    template_name   VARCHAR(128),                                           -- 基线名称(冗余)
    baseline_id     VARCHAR(255) DEFAULT '',                                  -- 检测批次ID(冗余)
    template_id     INTEGER NOT NULL,                                       -- 模板ID(冗余)
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
COMMENT ON COLUMN baseline_check_detail.result_id IS '关联检查结果ID';
COMMENT ON COLUMN baseline_check_detail.item_id IS '关联检查项ID';
COMMENT ON COLUMN baseline_check_detail.baseline_id IS '检测批次ID(冗余)';
COMMENT ON COLUMN baseline_check_detail.host_ip IS '主机IP(冗余)';
COMMENT ON COLUMN baseline_check_detail.host_name IS '主机名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.template_name IS '基线名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.template_id IS '模板ID(冗余，复合主键组成部分)';
COMMENT ON COLUMN baseline_check_detail.item_name IS '检查项名称(冗余)';
COMMENT ON COLUMN baseline_check_detail.risk_level IS '风险���级(冗余)';
COMMENT ON COLUMN baseline_check_detail.status IS '检查状态: 0-未通过 1-通过 2-检查异常';


-- =============================================================================
-- 第四步: 重新创建所有视图
-- =============================================================================

-- 1. 主机漏洞统计视图
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


-- 2. 镜像漏洞统计视图
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


-- 3. 漏洞维度主机统计视图
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


-- 4. 漏洞维度镜像统计视图
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


-- 5. 基线检查主机统计视图
CREATE OR REPLACE VIEW baseline_check_host_view AS
SELECT
    bcr.agent_id,
    bcr.host_name,
    bcr.host_ip,
    COUNT(*)                                         AS total_checks,
    COUNT(CASE WHEN bcd.status = 1 THEN 1 END)      AS passed_checks,
    COUNT(CASE WHEN bcd.status = 0 THEN 1 END)      AS failed_checks,
    COUNT(CASE WHEN bcd.status = 2 THEN 1 END)      AS error_checks,
    MAX(bcd.check_time)                              AS last_check_time
FROM baseline_check_detail bcd
JOIN baseline_check_result bcr ON bcd.result_id = bcr.id
GROUP BY bcr.agent_id, bcr.host_name, bcr.host_ip;

COMMENT ON VIEW baseline_check_host_view IS '基线检查-按主机统计';


-- 6. 基线检查项统计视图
CREATE OR REPLACE VIEW baseline_check_item_view AS
SELECT
    bci.item_name,
    COUNT(DISTINCT bcd.agent_id) AS total_hosts
FROM baseline_check_detail bcd
JOIN baseline_check_item bci ON bcd.item_id = bci.id
GROUP BY bci.item_name;

COMMENT ON VIEW baseline_check_item_view IS '基线检查-按检查项统计';


COMMIT;

-- =====================================================
-- 重建完成
-- =====================================================
