-- =====================================================
-- SOC 资产管理数据库初始化脚本
-- 数据库: PostgreSQL
-- 版本: 1.0
-- =====================================================

-- 1. 主机列表表 (asset_host)
-- 存储主机基础信息
CREATE TABLE IF NOT EXISTS asset_host (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识(安装时生成，不随IP/hostname变化)
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称(可变，Agent上报更新)
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址(可变，Agent上报更新，支持IPv6)
    mac_addr        VARCHAR(45),                                        -- MAC地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    os_version      VARCHAR(64),                                        -- 操作系统版本
    agent_status    SMALLINT        DEFAULT 0,                          -- Agent状态: 0=离线 1=在线
    agent_version   VARCHAR(32),                                        -- Agent版本号
    last_heartbeat  TIMESTAMP,                                          -- 最后心跳时间
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 主机表索引
CREATE INDEX IF NOT EXISTS idx_asset_host_agent_id ON asset_host(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_host_host_ip ON asset_host(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_host_agent_status ON asset_host(agent_status);
-- 唯一约束：同一agent_id只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_host_agent_id ON asset_host(agent_id);

COMMENT ON TABLE asset_host IS '资产管理-主机列表';
COMMENT ON COLUMN asset_host.agent_id IS 'Agent唯一标识(安装时生成，不随IP/hostname变化)';
COMMENT ON COLUMN asset_host.agent_status IS 'Agent状态: 0=离线 1=在线 (Agent上报数据时设置为1，心跳超时或断连时更新为0)';


-- 2. 端口列表表 (asset_port)
-- 存储主机端口监听信息
CREATE TABLE IF NOT EXISTS asset_port (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    port            INTEGER         NOT NULL,                           -- 端口
    protocol        SMALLINT        NOT NULL,                           -- 端口协议: 6=TCP, 17=UDP
    listen_ip       VARCHAR(45)     NOT NULL,                           -- 监听IP
    listen_process  VARCHAR(45)     NOT NULL,                           -- 监听进程
    run_user        VARCHAR(64),                                        -- 运行用户
    os_version      VARCHAR(64),                                        -- 操作系统版本
    agent_status    SMALLINT        DEFAULT 0,                          -- Agent状态: 0=离线 1=在线
    agent_version   VARCHAR(32),                                        -- Agent版本号
    process_time    TIMESTAMP,                                          -- 进程启动时间
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 端口表索引
CREATE INDEX IF NOT EXISTS idx_asset_port_agent_id ON asset_port(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_port_host_ip ON asset_port(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_port_port ON asset_port(port);
-- 唯一约束：同一agent_id+port+protocol只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_port_agent_port ON asset_port(agent_id, port, protocol);

COMMENT ON TABLE asset_port IS '资产管理-端口列表';
COMMENT ON COLUMN asset_port.port IS '监听端口';
COMMENT ON COLUMN asset_port.protocol IS '端口协议: 6=TCP, 17=UDP';


-- 3. 账号列表表 (asset_account)
-- 存储主机账号信息
CREATE TABLE IF NOT EXISTS asset_account (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    name            VARCHAR(128)    NOT NULL,                           -- 账号名称
    uid             INTEGER         NOT NULL,                           -- UID
    status          SMALLINT        NOT NULL DEFAULT 0,                 -- 账号状态: 0=正常 1=即将过期 2=已过期
    permission      VARCHAR(64)     NOT NULL,                           -- 权限: normal、root、sudo、root,sudo
    login_type      VARCHAR(128),                                       -- 登录Shell: /bin/bash、/bin/sh、/sbin/nologin等
    last_login_time TIMESTAMP,                                          -- 最后登录时间
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 账号表索引
CREATE INDEX IF NOT EXISTS idx_asset_account_agent_id ON asset_account(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_account_host_ip ON asset_account(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_account_name ON asset_account(name);
-- 唯一约束：同一agent_id+name只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_account_agent_user ON asset_account(agent_id, name);

COMMENT ON TABLE asset_account IS '资产管理-账号列表';
COMMENT ON COLUMN asset_account.name IS '系统账号名称';
COMMENT ON COLUMN asset_account.status IS '账号状态: 0=正常 1=即将过期 2=已过期';
COMMENT ON COLUMN asset_account.permission IS '权限: normal(普通用户)、root、sudo、root,sudo';
COMMENT ON COLUMN asset_account.login_type IS '登录Shell: /bin/bash、/bin/sh、/sbin/nologin等';


-- 4. 进程列表表 (asset_process)
-- 存储主机进程信息
CREATE TABLE IF NOT EXISTS asset_process (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    name            VARCHAR(128)    NOT NULL,                           -- 进程名称
    status          VARCHAR(64),                                        -- 进程状态
    version         VARCHAR(64),                                        -- 进程版本
    path            VARCHAR(512)    NOT NULL,                           -- 进程路径
    run_name        VARCHAR(128)    NOT NULL,                           -- 运行用户
    start_time      TIMESTAMP,                                          -- 进程启动时间
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 进程表索引
CREATE INDEX IF NOT EXISTS idx_asset_process_agent_id ON asset_process(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_process_host_ip ON asset_process(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_process_name ON asset_process(name);
-- 唯一约束：同一agent_id+path只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_process_agent_path ON asset_process(agent_id, path);

COMMENT ON TABLE asset_process IS '资产管理-进程列表';
COMMENT ON COLUMN asset_process.name IS '进程名称';
COMMENT ON COLUMN asset_process.run_name IS '运行用户';


-- 5. 数据库列表表 (asset_database)
-- 存储主机上运行的数据库信息
CREATE TABLE IF NOT EXISTS asset_database (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    db_type         VARCHAR(45)     NOT NULL,                           -- 数据库类型
    db_version      VARCHAR(45)     NOT NULL,                           -- 数据库版本
    port            INTEGER         NOT NULL,                           -- 监听端口
    run_user        VARCHAR(64),                                        -- 运行用户
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 数据库表索引
CREATE INDEX IF NOT EXISTS idx_asset_database_agent_id ON asset_database(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_database_host_ip ON asset_database(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_database_db_type ON asset_database(db_type);
-- 唯一约束：同一agent_id+db_type只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_database_agent_type ON asset_database(agent_id, db_type);

COMMENT ON TABLE asset_database IS '资产管理-数据库列表';
COMMENT ON COLUMN asset_database.db_type IS '数据库类型(MySQL/PostgreSQL/Oracle等)';
COMMENT ON COLUMN asset_database.db_version IS '数据库版本';


-- 6. Web服务表 (asset_web_service)
-- 存储Web服务信息
CREATE TABLE IF NOT EXISTS asset_web_service (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    name            VARCHAR(128)    NOT NULL,                           -- 应用名
    version         VARCHAR(64)     NOT NULL,                           -- 版本
    server_type     VARCHAR(64)     NOT NULL,                           -- 服务器类型
    site_domain     VARCHAR(255),                                       -- 站点域名
    path            VARCHAR(512),                                       -- 根路径
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- Web服务表索引
CREATE INDEX IF NOT EXISTS idx_asset_web_service_agent_id ON asset_web_service(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_web_service_host_ip ON asset_web_service(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_web_service_server_type ON asset_web_service(server_type);
-- 唯一约束：同一agent_id+server_type只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_web_service_agent_type ON asset_web_service(agent_id, server_type);

COMMENT ON TABLE asset_web_service IS '资产管理-Web服务';
COMMENT ON COLUMN asset_web_service.name IS '应用名称';
COMMENT ON COLUMN asset_web_service.server_type IS '服务器类型(Nginx/Apache/Tomcat等)';


-- 7. 系统服务表 (asset_system_service)
-- 存储系统服务信息
CREATE TABLE IF NOT EXISTS asset_system_service (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    name            VARCHAR(255)    NOT NULL,                           -- 服务名称
    version         VARCHAR(64),                                        -- 版本
    status          VARCHAR(64)     NOT NULL,                           -- 状态
    run_user        VARCHAR(255)    NOT NULL,                           -- 运行用户
    path            VARCHAR(512)    NOT NULL,                           -- 根路径
    describe        VARCHAR(512),                                       -- 描述
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 系统服务表索引
CREATE INDEX IF NOT EXISTS idx_asset_system_service_agent_id ON asset_system_service(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_system_service_host_ip ON asset_system_service(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_system_service_name ON asset_system_service(name);
-- 唯一约束：同一agent_id+name只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_system_service_agent_name ON asset_system_service(agent_id, name);

COMMENT ON TABLE asset_system_service IS '资产管理-系统服务';
COMMENT ON COLUMN asset_system_service.name IS '服务名称';
COMMENT ON COLUMN asset_system_service.status IS '服务状态';


-- =====================================================
-- 初始化完成
-- =====================================================


-- =====================================================
-- 新增资产表（软件、容器、可疑环境变量、内核模块）
-- =====================================================

-- 8. 软件列表表 (asset_software)
-- 存储主机安装的软件包信息
CREATE TABLE IF NOT EXISTS asset_software (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    name            VARCHAR(255)    NOT NULL,                           -- 软件名称
    version         VARCHAR(128),                                       -- 软件版本
    type            VARCHAR(32)     NOT NULL,                           -- 软件类型: dpkg, rpm, pypi, jar
    source          VARCHAR(255),                                       -- 来源
    status          VARCHAR(64),                                        -- 状态
    vendor          VARCHAR(255),                                       -- 厂商
    path            VARCHAR(512),                                       -- 路径(jar类型)
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 软件表索引
CREATE INDEX IF NOT EXISTS idx_asset_software_agent_id ON asset_software(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_software_host_ip ON asset_software(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_software_name ON asset_software(name);
CREATE INDEX IF NOT EXISTS idx_asset_software_type ON asset_software(type);
-- 唯一约束：同一agent_id+name+type只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_software_agent_name_type ON asset_software(agent_id, name, type);

COMMENT ON TABLE asset_software IS '资产管理-软件列表';
COMMENT ON COLUMN asset_software.name IS '软件名称';
COMMENT ON COLUMN asset_software.type IS '软件类型(dpkg/rpm/pypi/jar)';


-- 9. 容器列表表 (asset_container)
-- 存储主机上运行的容器信息
CREATE TABLE IF NOT EXISTS asset_container (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    container_id    VARCHAR(128)    NOT NULL,                           -- 容器ID
    name            VARCHAR(255)    NOT NULL,                           -- 容器名称
    state           VARCHAR(32)     NOT NULL,                           -- 容器状态
    image_id        VARCHAR(128),                                       -- 镜像ID
    image_name      VARCHAR(255),                                       -- 镜像名称
    runtime         VARCHAR(32),                                        -- 运行时(docker/containerd)
    pid             VARCHAR(16),                                        -- 容器主进程PID
    create_time     VARCHAR(32),                                        -- 容器创建时间
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 容器表索引
CREATE INDEX IF NOT EXISTS idx_asset_container_agent_id ON asset_container(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_container_host_ip ON asset_container(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_container_name ON asset_container(name);
CREATE INDEX IF NOT EXISTS idx_asset_container_state ON asset_container(state);
-- 唯一约束：同一agent_id+container_id只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_container_agent_cid ON asset_container(agent_id, container_id);

COMMENT ON TABLE asset_container IS '资产管理-容器列表';
COMMENT ON COLUMN asset_container.container_id IS '容器ID';
COMMENT ON COLUMN asset_container.state IS '容器状态(running/exited/created等)';
COMMENT ON COLUMN asset_container.runtime IS '容器运行时(docker/containerd)';


-- 10. 可疑环境变量表 (asset_env_suspicious)
-- 存储检测到的可疑环境变量
CREATE TABLE IF NOT EXISTS asset_env_suspicious (
    id                  BIGSERIAL       PRIMARY KEY,
    agent_id            VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name           VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip             VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    var_name            VARCHAR(255)    NOT NULL,                           -- 环境变量名
    var_value           TEXT,                                               -- 环境变量值
    suspicious_reasons  TEXT,                                               -- 可疑原因
    source              VARCHAR(128),                                       -- 来源
    created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 可疑环境变量表索引
CREATE INDEX IF NOT EXISTS idx_asset_env_suspicious_agent_id ON asset_env_suspicious(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_env_suspicious_host_ip ON asset_env_suspicious(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_env_suspicious_var_name ON asset_env_suspicious(var_name);
-- 唯一约束：同一agent_id+var_name只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_env_suspicious_agent_var ON asset_env_suspicious(agent_id, var_name);

COMMENT ON TABLE asset_env_suspicious IS '资产管理-可疑环境变量';
COMMENT ON COLUMN asset_env_suspicious.var_name IS '环境变量名称';
COMMENT ON COLUMN asset_env_suspicious.suspicious_reasons IS '可疑原因';


-- 11. 内核模块表 (asset_kmod)
-- 存储主机加载的内核模块信息
CREATE TABLE IF NOT EXISTS asset_kmod (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    os_type         VARCHAR(32),                                        -- 操作系统类型: linux/windows
    name            VARCHAR(128)    NOT NULL,                           -- 模块名称
    size            VARCHAR(32),                                        -- 模块大小
    refcount        VARCHAR(16),                                        -- 引用计数
    used_by         VARCHAR(512),                                       -- 使用该模块的模块列表
    state           VARCHAR(32),                                        -- 模块状态
    addr            VARCHAR(32),                                        -- 内存地址
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 内核模块表索引
CREATE INDEX IF NOT EXISTS idx_asset_kmod_agent_id ON asset_kmod(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_kmod_host_ip ON asset_kmod(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_kmod_name ON asset_kmod(name);
-- 唯一约束：同一agent_id+name只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_kmod_agent_name ON asset_kmod(agent_id, name);

COMMENT ON TABLE asset_kmod IS '资产管理-内核模块';
COMMENT ON COLUMN asset_kmod.name IS '内核模块名称';
COMMENT ON COLUMN asset_kmod.state IS '模块状态(Live/Loading/Unloading)';


-- 12. 镜像软件包表 (asset_image_package)
-- 存储容器镜像中的软件包信息
CREATE TABLE IF NOT EXISTS asset_image_package (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    image_id        VARCHAR(128)    NOT NULL,                           -- 镜像ID
    image_name      VARCHAR(255)    NOT NULL,                           -- 镜像名称
    package_name    VARCHAR(255)    NOT NULL,                           -- 软件包名称
    package_version VARCHAR(128),                                       -- 软件包版本
    package_type    VARCHAR(32)     NOT NULL,                           -- 软件包类型: dpkg/rpm/apk
    os_version      VARCHAR(64),                                        -- 操作系统版本
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 镜像软件包表索引
CREATE INDEX IF NOT EXISTS idx_asset_imgpkg_agent_id ON asset_image_package(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_imgpkg_image_id ON asset_image_package(image_id);
-- 唯一约束：同一agent_id+image_id+package_name只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_imgpkg_composite ON asset_image_package(agent_id, image_id, package_name);

COMMENT ON TABLE asset_image_package IS '资产管理-镜像软件包';
COMMENT ON COLUMN asset_image_package.package_type IS '软件包类型(dpkg/rpm/apk)';
