-- =====================================================
-- 镜像资产表 (asset_image)
-- 数据库: PostgreSQL
-- 说明: 存储容器镜像资产信息
-- =====================================================

-- 12. 镜像列表表 (asset_image)
-- 存储主机上的容器镜像信息
CREATE TABLE IF NOT EXISTS asset_image (
    id              BIGSERIAL       PRIMARY KEY,
    agent_id        VARCHAR(64)     NOT NULL,                           -- Agent唯一标识
    host_name       VARCHAR(128)    NOT NULL,                           -- 主机名称
    host_ip         VARCHAR(45)     NOT NULL,                           -- 主机IP地址
    image_id        VARCHAR(128)    NOT NULL,                           -- 镜像ID (sha256格式)
    image_name      VARCHAR(255)    NOT NULL,                           -- 镜像名称
    image_version   VARCHAR(128),                                       -- 镜像版本/标签
    image_size      BIGINT,                                             -- 镜像大小(字节)
    container_count INTEGER         DEFAULT 0,                          -- 关联容器数
    build_time      TIMESTAMP,                                          -- 镜像构建时间
    runtime         VARCHAR(32),                                        -- 容器运行时(docker/containerd)
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP  -- 更新时间
);

-- 镜像表索引
CREATE INDEX IF NOT EXISTS idx_asset_image_agent_id ON asset_image(agent_id);
CREATE INDEX IF NOT EXISTS idx_asset_image_host_ip ON asset_image(host_ip);
CREATE INDEX IF NOT EXISTS idx_asset_image_image_name ON asset_image(image_name);
CREATE INDEX IF NOT EXISTS idx_asset_image_image_version ON asset_image(image_version);
-- 唯一约束：同一agent_id+image_id只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS uk_asset_image_agent_imgid ON asset_image(agent_id, image_id);

COMMENT ON TABLE asset_image IS '资产管理-镜像列表';
COMMENT ON COLUMN asset_image.image_id IS '镜像ID(sha256格式，如sha256:a3ed95caeb02...)';
COMMENT ON COLUMN asset_image.image_name IS '镜像名称(如nginx、mysql、redis等)';
COMMENT ON COLUMN asset_image.image_version IS '镜像版本/标签(如1.21.6、8.0.28、latest等)';
COMMENT ON COLUMN asset_image.image_size IS '镜像大小(字节)';
COMMENT ON COLUMN asset_image.container_count IS '关联容器数量';
COMMENT ON COLUMN asset_image.build_time IS '镜像构建时间';
COMMENT ON COLUMN asset_image.runtime IS '容器运行时(docker/containerd)';
