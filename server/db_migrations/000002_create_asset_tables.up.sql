-- 000002: 资产管理表
-- 包含: asset_host, asset_port, asset_account, asset_process, asset_database,
--       asset_web_service, asset_system_service, asset_software, asset_container,
--       asset_image, asset_image_package, asset_env_suspicious, asset_kmod (13 表)

-- 1. 主机列表表 (asset_host)
CREATE TABLE IF NOT EXISTS asset_host (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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
    host_ip         VARCHAR(256)     NOT NULL,
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


-- 10. 镜像列表表 (asset_image)
CREATE TABLE IF NOT EXISTS asset_image (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(256)     NOT NULL,
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


-- 11. 镜像软件包表 (asset_image_package)
CREATE TABLE IF NOT EXISTS asset_image_package (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(256)     NOT NULL,
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


-- 12. 可疑环境变量表 (asset_env_suspicious)
CREATE TABLE IF NOT EXISTS asset_env_suspicious (
    id                  BIGSERIAL       PRIMARY KEY,
    agent_id            VARCHAR(64)     NOT NULL,
    host_name           VARCHAR(128)    NOT NULL,
    host_ip             VARCHAR(256)     NOT NULL,
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


-- 13. 内核模块表 (asset_kmod)
CREATE TABLE IF NOT EXISTS asset_kmod (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,
    host_name       VARCHAR(128)    NOT NULL,
    host_ip         VARCHAR(256)     NOT NULL,
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
