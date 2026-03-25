-- =====================================================
-- 建表 + 模拟数据: asset_image_package (镜像软件包表)
-- 数据量: 约120条
-- 说明: 基于 asset_image 中的镜像生成软件包数据
-- =====================================================

-- 模拟数据 (表已在 rebuild_soc_db.sql / init_asset_db.sql 中创建)
INSERT INTO asset_image_package (agent_id, host_name, host_ip, image_id, image_name, package_name, package_version, package_type, os_version, created_at, updated_at) VALUES
-- nginx:1.25.3 (Debian based, web-server-01)
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'nginx', '1.25.3-1~bookworm', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'libc6', '2.36-9+deb12u4', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'libssl3', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'libpcre2-8-0', '10.42-1', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'zlib1g', '1:1.2.13.dfsg-1', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'openssl', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:nginx567jkl890123456789012345678901234nginx', 'nginx', 'libgcc-s1', '12.2.0-14', 'dpkg', 'debian 12', NOW() - INTERVAL '90 days', NOW()),
-- alpine:3.19 (web-server-01)
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:alpine678klm901234567890123456789012alpin', 'alpine', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.19', NOW() - INTERVAL '20 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:alpine678klm901234567890123456789012alpin', 'alpine', 'busybox', '1.36.1-r15', 'apk', 'alpine 3.19', NOW() - INTERVAL '20 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:alpine678klm901234567890123456789012alpin', 'alpine', 'alpine-baselayout', '3.4.3-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '20 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:alpine678klm901234567890123456789012alpin', 'alpine', 'ssl_client', '1.36.1-r15', 'apk', 'alpine 3.19', NOW() - INTERVAL '20 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:alpine678klm901234567890123456789012alpin', 'alpine', 'libcrypto3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '20 days', NOW()),
('agent-001-a1b2c3d4', 'web-server-01', '192.168.1.10', 'sha256:alpine678klm901234567890123456789012alpin', 'alpine', 'libssl3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '20 days', NOW()),
-- redis:7.2.3 (cache-server-01)
('agent-007-y5z6a7b8', 'cache-server-01', '192.168.1.40', 'sha256:redis890uvw123456789012345678901234redis', 'redis', 'redis-server', '7.2.3', 'dpkg', 'debian 12', NOW() - INTERVAL '30 days', NOW()),
('agent-007-y5z6a7b8', 'cache-server-01', '192.168.1.40', 'sha256:redis890uvw123456789012345678901234redis', 'redis', 'libc6', '2.36-9+deb12u4', 'dpkg', 'debian 12', NOW() - INTERVAL '30 days', NOW()),
('agent-007-y5z6a7b8', 'cache-server-01', '192.168.1.40', 'sha256:redis890uvw123456789012345678901234redis', 'redis', 'libssl3', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '30 days', NOW()),
('agent-007-y5z6a7b8', 'cache-server-01', '192.168.1.40', 'sha256:redis890uvw123456789012345678901234redis', 'redis', 'gosu', '1.16', 'dpkg', 'debian 12', NOW() - INTERVAL '30 days', NOW()),
-- mysql:8.0.35 (db-server-01)
('agent-005-q7r8s9t0', 'db-server-01', '192.168.1.30', 'sha256:mysql456qrs789012345678901234567890mysql', 'mysql', 'mysql-community-server-core', '8.0.35-1.el8', 'rpm', 'oracle 8', NOW() - INTERVAL '90 days', NOW()),
('agent-005-q7r8s9t0', 'db-server-01', '192.168.1.30', 'sha256:mysql456qrs789012345678901234567890mysql', 'mysql', 'mysql-community-client', '8.0.35-1.el8', 'rpm', 'oracle 8', NOW() - INTERVAL '90 days', NOW()),
('agent-005-q7r8s9t0', 'db-server-01', '192.168.1.30', 'sha256:mysql456qrs789012345678901234567890mysql', 'mysql', 'mysql-community-common', '8.0.35-1.el8', 'rpm', 'oracle 8', NOW() - INTERVAL '90 days', NOW()),
('agent-005-q7r8s9t0', 'db-server-01', '192.168.1.30', 'sha256:mysql456qrs789012345678901234567890mysql', 'mysql', 'openssl-libs', '1.1.1k-9.el8', 'rpm', 'oracle 8', NOW() - INTERVAL '90 days', NOW()),
('agent-005-q7r8s9t0', 'db-server-01', '192.168.1.30', 'sha256:mysql456qrs789012345678901234567890mysql', 'mysql', 'glibc', '2.28-225.0.4.el8', 'rpm', 'oracle 8', NOW() - INTERVAL '90 days', NOW()),
('agent-005-q7r8s9t0', 'db-server-01', '192.168.1.30', 'sha256:mysql456qrs789012345678901234567890mysql', 'mysql', 'ncurses-libs', '6.1-9.20180224.el8', 'rpm', 'oracle 8', NOW() - INTERVAL '90 days', NOW()),
-- postgres:16.1 (db-server-02)
('agent-006-u1v2w3x4', 'db-server-02', '192.168.1.31', 'sha256:postgres678stu901234567890123456789postg', 'postgres', 'postgresql-16', '16.1-1.pgdg120+1', 'dpkg', 'debian 12', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'db-server-02', '192.168.1.31', 'sha256:postgres678stu901234567890123456789postg', 'postgres', 'libc6', '2.36-9+deb12u4', 'dpkg', 'debian 12', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'db-server-02', '192.168.1.31', 'sha256:postgres678stu901234567890123456789postg', 'postgres', 'libssl3', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'db-server-02', '192.168.1.31', 'sha256:postgres678stu901234567890123456789postg', 'postgres', 'libldap-2.5-0', '2.5.13+dfsg-5', 'dpkg', 'debian 12', NOW() - INTERVAL '80 days', NOW()),
('agent-006-u1v2w3x4', 'db-server-02', '192.168.1.31', 'sha256:postgres678stu901234567890123456789postg', 'postgres', 'gosu', '1.16', 'dpkg', 'debian 12', NOW() - INTERVAL '80 days', NOW()),
-- rabbitmq:3.12-management (mq-server-01)
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:rabbitmq012wxy345678901234567890rabbit', 'rabbitmq', 'rabbitmq-server', '3.12.10-1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:rabbitmq012wxy345678901234567890rabbit', 'rabbitmq', 'erlang-base', '1:25.3.2.8+dfsg-1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:rabbitmq012wxy345678901234567890rabbit', 'rabbitmq', 'libc6', '2.35-0ubuntu3.5', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:rabbitmq012wxy345678901234567890rabbit', 'rabbitmq', 'libssl3', '3.0.2-0ubuntu1.12', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
-- harbor-core (harbor-server-01, Photon OS / rpm)
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:abc123def456789012345678901234567890123456789012345678901234abcd', 'goharbor/harbor-core', 'harbor-core', '2.9.1-1.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:abc123def456789012345678901234567890123456789012345678901234abcd', 'goharbor/harbor-core', 'glibc', '2.36-3.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:abc123def456789012345678901234567890123456789012345678901234abcd', 'goharbor/harbor-core', 'openssl', '3.0.8-5.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:abc123def456789012345678901234567890123456789012345678901234abcd', 'goharbor/harbor-core', 'curl-libs', '8.4.0-1.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
-- harbor-db (harbor-server-01, Photon OS)
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:cde345fgh678901234567890123456789012345678901234567890123456cdef', 'goharbor/harbor-db', 'postgresql13', '13.13-1.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:cde345fgh678901234567890123456789012345678901234567890123456cdef', 'goharbor/harbor-db', 'glibc', '2.36-3.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
('agent-019-u3v4w5x6', 'harbor-server-01', '192.168.2.30', 'sha256:cde345fgh678901234567890123456789012345678901234567890123456cdef', 'goharbor/harbor-db', 'openssl', '3.0.8-5.ph4', 'rpm', 'photon 4.0', NOW() - INTERVAL '80 days', NOW()),
-- redis:7.2-alpine (k8s-node-02)
('agent-024-o3p4q5r6', 'k8s-node-02', '192.168.3.21', 'sha256:redis234lmn567890123456789012345678901234567890123456redis', 'redis', 'redis', '7.2.3-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '55 days', NOW()),
('agent-024-o3p4q5r6', 'k8s-node-02', '192.168.3.21', 'sha256:redis234lmn567890123456789012345678901234567890123456redis', 'redis', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.19', NOW() - INTERVAL '55 days', NOW()),
('agent-024-o3p4q5r6', 'k8s-node-02', '192.168.3.21', 'sha256:redis234lmn567890123456789012345678901234567890123456redis', 'redis', 'libcrypto3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '55 days', NOW()),
('agent-024-o3p4q5r6', 'k8s-node-02', '192.168.3.21', 'sha256:redis234lmn567890123456789012345678901234567890123456redis', 'redis', 'libssl3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '55 days', NOW()),
-- postgres:15-alpine (jenkins-server-01)
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:postgres456def789012345678901234567890123456789012345678901postg', 'postgres', 'postgresql15', '15.5-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:postgres456def789012345678901234567890123456789012345678901postg', 'postgres', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.19', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:postgres456def789012345678901234567890123456789012345678901postg', 'postgres', 'libcrypto3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:postgres456def789012345678901234567890123456789012345678901postg', 'postgres', 'libssl3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:postgres456def789012345678901234567890123456789012345678901postg', 'postgres', 'icu-libs', '74.1-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '100 days', NOW()),
-- sonarqube:10.3-community (jenkins-server-01, Debian/Ubuntu based)
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:sonar345cde678901234567890123456789012345678901234567890123sonar', 'sonarqube', 'openjdk-17-jre-headless', '17.0.9+9-1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:sonar345cde678901234567890123456789012345678901234567890123sonar', 'sonarqube', 'libc6', '2.35-0ubuntu3.5', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:sonar345cde678901234567890123456789012345678901234567890123sonar', 'sonarqube', 'libssl3', '3.0.2-0ubuntu1.12', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
('agent-017-m5n6o7p8', 'jenkins-server-01', '192.168.2.10', 'sha256:sonar345cde678901234567890123456789012345678901234567890123sonar', 'sonarqube', 'libfreetype6', '2.11.1+dfsg-1ubuntu0.2', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '100 days', NOW()),
-- node:20-alpine (dev-server-01)
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:nodejs567bcd890123456789012345678nodej', 'node', 'nodejs', '20.10.0-r1', 'apk', 'alpine 3.19', NOW() - INTERVAL '15 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:nodejs567bcd890123456789012345678nodej', 'node', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.19', NOW() - INTERVAL '15 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:nodejs567bcd890123456789012345678nodej', 'node', 'libcrypto3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '15 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:nodejs567bcd890123456789012345678nodej', 'node', 'libssl3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '15 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:nodejs567bcd890123456789012345678nodej', 'node', 'libstdc++', '13.2.1_git20231014-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '15 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:nodejs567bcd890123456789012345678nodej', 'node', 'nghttp2-libs', '1.58.0-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '15 days', NOW()),
-- python:3.12-slim (dev-server-01, Debian based)
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:python678cde901234567890123456789pytho', 'python', 'python3.12-minimal', '3.12.1-1', 'dpkg', 'debian 12', NOW() - INTERVAL '20 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:python678cde901234567890123456789pytho', 'python', 'libc6', '2.36-9+deb12u4', 'dpkg', 'debian 12', NOW() - INTERVAL '20 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:python678cde901234567890123456789pytho', 'python', 'libssl3', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '20 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:python678cde901234567890123456789pytho', 'python', 'libexpat1', '2.5.0-1', 'dpkg', 'debian 12', NOW() - INTERVAL '20 days', NOW()),
('agent-030-c9d0e1f2', 'dev-server-01', '192.168.4.10', 'sha256:python678cde901234567890123456789pytho', 'python', 'libsqlite3-0', '3.40.1-2', 'dpkg', 'debian 12', NOW() - INTERVAL '20 days', NOW()),
-- elasticsearch:8.11.1 (k8s-node-05, Ubuntu based)
('agent-027-a5b6c7d8', 'k8s-node-05', '192.168.3.24', 'sha256:elastic567pqr890123456789012345678901234567890123456elast', 'docker.elastic.co/elasticsearch/elasticsearch', 'elasticsearch', '8.11.1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '55 days', NOW()),
('agent-027-a5b6c7d8', 'k8s-node-05', '192.168.3.24', 'sha256:elastic567pqr890123456789012345678901234567890123456elast', 'docker.elastic.co/elasticsearch/elasticsearch', 'openjdk-21-jre-headless', '21.0.1+12-1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '55 days', NOW()),
('agent-027-a5b6c7d8', 'k8s-node-05', '192.168.3.24', 'sha256:elastic567pqr890123456789012345678901234567890123456elast', 'docker.elastic.co/elasticsearch/elasticsearch', 'libc6', '2.35-0ubuntu3.5', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '55 days', NOW()),
('agent-027-a5b6c7d8', 'k8s-node-05', '192.168.3.24', 'sha256:elastic567pqr890123456789012345678901234567890123456elast', 'docker.elastic.co/elasticsearch/elasticsearch', 'libssl3', '3.0.2-0ubuntu1.12', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '55 days', NOW()),
('agent-027-a5b6c7d8', 'k8s-node-05', '192.168.3.24', 'sha256:elastic567pqr890123456789012345678901234567890123456elast', 'docker.elastic.co/elasticsearch/elasticsearch', 'zlib1g', '1:1.2.11.dfsg-2ubuntu9.2', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '55 days', NOW()),
-- company/api-service:v2.3.1 (api-server-01, Alpine based)
('agent-003-i9j0k1l2', 'api-server-01', '192.168.1.20', 'sha256:apiapp012opq345678901234567890123456apiap', 'company/api-service', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.18', NOW() - INTERVAL '10 days', NOW()),
('agent-003-i9j0k1l2', 'api-server-01', '192.168.1.20', 'sha256:apiapp012opq345678901234567890123456apiap', 'company/api-service', 'libcrypto3', '3.1.3-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '10 days', NOW()),
('agent-003-i9j0k1l2', 'api-server-01', '192.168.1.20', 'sha256:apiapp012opq345678901234567890123456apiap', 'company/api-service', 'libssl3', '3.1.3-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '10 days', NOW()),
('agent-003-i9j0k1l2', 'api-server-01', '192.168.1.20', 'sha256:apiapp012opq345678901234567890123456apiap', 'company/api-service', 'ca-certificates', '20230506-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '10 days', NOW()),
-- company/frontend:v2.5.0 (k8s-node-01, nginx-based, Debian)
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:frontend901ijk234567890123456789012345678901234567890123front', 'company/frontend', 'nginx', '1.25.3-1~bookworm', 'dpkg', 'debian 12', NOW() - INTERVAL '50 days', NOW()),
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:frontend901ijk234567890123456789012345678901234567890123front', 'company/frontend', 'libc6', '2.36-9+deb12u4', 'dpkg', 'debian 12', NOW() - INTERVAL '50 days', NOW()),
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:frontend901ijk234567890123456789012345678901234567890123front', 'company/frontend', 'libssl3', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '50 days', NOW()),
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:frontend901ijk234567890123456789012345678901234567890123front', 'company/frontend', 'libpcre2-8-0', '10.42-1', 'dpkg', 'debian 12', NOW() - INTERVAL '50 days', NOW()),
-- company/backend:v3.1.2 (k8s-node-01, Alpine)
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:backend012jkl345678901234567890123456789012345678901234backe', 'company/backend', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:backend012jkl345678901234567890123456789012345678901234backe', 'company/backend', 'libcrypto3', '3.1.3-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:backend012jkl345678901234567890123456789012345678901234backe', 'company/backend', 'libssl3', '3.1.3-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
('agent-023-k9l0m1n2', 'k8s-node-01', '192.168.3.20', 'sha256:backend012jkl345678901234567890123456789012345678901234backe', 'company/backend', 'ca-certificates', '20230506-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
-- filebeat:8.11.1 (log-server-01, Ubuntu based)
('agent-011-o1p2q3r4', 'log-server-01', '192.168.1.60', 'sha256:filebeat567zab890123456789012345678901234fileb', 'docker.elastic.co/beats/filebeat', 'filebeat', '8.11.1', 'dpkg', 'ubuntu 20.04', NOW() - INTERVAL '200 days', NOW()),
('agent-011-o1p2q3r4', 'log-server-01', '192.168.1.60', 'sha256:filebeat567zab890123456789012345678901234fileb', 'docker.elastic.co/beats/filebeat', 'libc6', '2.31-0ubuntu9.14', 'dpkg', 'ubuntu 20.04', NOW() - INTERVAL '200 days', NOW()),
('agent-011-o1p2q3r4', 'log-server-01', '192.168.1.60', 'sha256:filebeat567zab890123456789012345678901234fileb', 'docker.elastic.co/beats/filebeat', 'libssl1.1', '1.1.1f-1ubuntu2.20', 'dpkg', 'ubuntu 20.04', NOW() - INTERVAL '200 days', NOW()),
-- minio:RELEASE.2023-12-20 (backup-server-01, Debian based)
('agent-013-w9x0y1z2', 'backup-server-01', '192.168.1.80', 'sha256:minio123fgh456789012345678901234567890minio', 'minio/minio', 'minio', '2023.12.20', 'dpkg', 'debian 12', NOW() - INTERVAL '10 days', NOW()),
('agent-013-w9x0y1z2', 'backup-server-01', '192.168.1.80', 'sha256:minio123fgh456789012345678901234567890minio', 'minio/minio', 'libc6', '2.36-9+deb12u4', 'dpkg', 'debian 12', NOW() - INTERVAL '10 days', NOW()),
('agent-013-w9x0y1z2', 'backup-server-01', '192.168.1.80', 'sha256:minio123fgh456789012345678901234567890minio', 'minio/minio', 'libssl3', '3.0.11-1~deb12u2', 'dpkg', 'debian 12', NOW() - INTERVAL '10 days', NOW()),
-- gitlab-runner:v16.6.1 (gitlab-server-01, Ubuntu based)
('agent-018-q9r0s1t2', 'gitlab-server-01', '192.168.2.20', 'sha256:runner123vwx456789012345678901234567890123456789runne', 'gitlab/gitlab-runner', 'gitlab-runner', '16.6.1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '95 days', NOW()),
('agent-018-q9r0s1t2', 'gitlab-server-01', '192.168.2.20', 'sha256:runner123vwx456789012345678901234567890123456789runne', 'gitlab/gitlab-runner', 'libc6', '2.35-0ubuntu3.5', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '95 days', NOW()),
('agent-018-q9r0s1t2', 'gitlab-server-01', '192.168.2.20', 'sha256:runner123vwx456789012345678901234567890123456789runne', 'gitlab/gitlab-runner', 'git', '1:2.34.1-1ubuntu1.10', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '95 days', NOW()),
('agent-018-q9r0s1t2', 'gitlab-server-01', '192.168.2.20', 'sha256:runner123vwx456789012345678901234567890123456789runne', 'gitlab/gitlab-runner', 'ca-certificates', '20230311ubuntu0.22.04.1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '95 days', NOW()),
('agent-018-q9r0s1t2', 'gitlab-server-01', '192.168.2.20', 'sha256:runner123vwx456789012345678901234567890123456789runne', 'gitlab/gitlab-runner', 'libssl3', '3.0.2-0ubuntu1.12', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '95 days', NOW()),
-- selenium/standalone-chrome:4.16.1 (test-server-01, Debian based)
('agent-031-g3h4i5j6', 'test-server-01', '192.168.4.20', 'sha256:selenium890efg123456789012345678selen', 'selenium/standalone-chrome', 'google-chrome-stable', '120.0.6099.109-1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '10 days', NOW()),
('agent-031-g3h4i5j6', 'test-server-01', '192.168.4.20', 'sha256:selenium890efg123456789012345678selen', 'selenium/standalone-chrome', 'chromium-chromedriver', '120.0.6099.71-0ubuntu0.22.04.1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '10 days', NOW()),
('agent-031-g3h4i5j6', 'test-server-01', '192.168.4.20', 'sha256:selenium890efg123456789012345678selen', 'selenium/standalone-chrome', 'openjdk-11-jre-headless', '11.0.21+9-0ubuntu1~22.04', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '10 days', NOW()),
('agent-031-g3h4i5j6', 'test-server-01', '192.168.4.20', 'sha256:selenium890efg123456789012345678selen', 'selenium/standalone-chrome', 'libc6', '2.35-0ubuntu3.5', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '10 days', NOW()),
('agent-031-g3h4i5j6', 'test-server-01', '192.168.4.20', 'sha256:selenium890efg123456789012345678selen', 'selenium/standalone-chrome', 'libssl3', '3.0.2-0ubuntu1.12', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '10 days', NOW()),
('agent-031-g3h4i5j6', 'test-server-01', '192.168.4.20', 'sha256:selenium890efg123456789012345678selen', 'selenium/standalone-chrome', 'libnss3', '2:3.68.2-0ubuntu1.2', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '10 days', NOW()),
-- php:8.2-fpm-alpine (web-server-02)
('agent-002-e5f6g7h8', 'web-server-02', '192.168.1.11', 'sha256:php890mno123456789012345678901234567phpad', 'php', 'php82', '8.2.13-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '85 days', NOW()),
('agent-002-e5f6g7h8', 'web-server-02', '192.168.1.11', 'sha256:php890mno123456789012345678901234567phpad', 'php', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.19', NOW() - INTERVAL '85 days', NOW()),
('agent-002-e5f6g7h8', 'web-server-02', '192.168.1.11', 'sha256:php890mno123456789012345678901234567phpad', 'php', 'libcrypto3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '85 days', NOW()),
('agent-002-e5f6g7h8', 'web-server-02', '192.168.1.11', 'sha256:php890mno123456789012345678901234567phpad', 'php', 'libssl3', '3.1.4-r2', 'apk', 'alpine 3.19', NOW() - INTERVAL '85 days', NOW()),
('agent-002-e5f6g7h8', 'web-server-02', '192.168.1.11', 'sha256:php890mno123456789012345678901234567phpad', 'php', 'libxml2', '2.11.6-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '85 days', NOW()),
('agent-002-e5f6g7h8', 'web-server-02', '192.168.1.11', 'sha256:php890mno123456789012345678901234567phpad', 'php', 'curl', '8.5.0-r0', 'apk', 'alpine 3.19', NOW() - INTERVAL '85 days', NOW()),
-- awx:21.5.0 (ansible-server-01, CentOS based)
('agent-049-k3l4m5n6', 'ansible-server-01', '192.168.2.40', 'sha256:awxweb789bcd012345678901234567890123456awxwe', 'quay.io/ansible/awx', 'python39', '3.9.18-1.el8', 'rpm', 'centos 8', NOW() - INTERVAL '150 days', NOW()),
('agent-049-k3l4m5n6', 'ansible-server-01', '192.168.2.40', 'sha256:awxweb789bcd012345678901234567890123456awxwe', 'quay.io/ansible/awx', 'glibc', '2.28-225.el8', 'rpm', 'centos 8', NOW() - INTERVAL '150 days', NOW()),
('agent-049-k3l4m5n6', 'ansible-server-01', '192.168.2.40', 'sha256:awxweb789bcd012345678901234567890123456awxwe', 'quay.io/ansible/awx', 'openssl-libs', '1.1.1k-9.el8', 'rpm', 'centos 8', NOW() - INTERVAL '150 days', NOW()),
('agent-049-k3l4m5n6', 'ansible-server-01', '192.168.2.40', 'sha256:awxweb789bcd012345678901234567890123456awxwe', 'quay.io/ansible/awx', 'nginx', '1.22.1-1.el8', 'rpm', 'centos 8', NOW() - INTERVAL '150 days', NOW()),
('agent-049-k3l4m5n6', 'ansible-server-01', '192.168.2.40', 'sha256:awxweb789bcd012345678901234567890123456awxwe', 'quay.io/ansible/awx', 'postgresql-libs', '13.13-1.el8', 'rpm', 'centos 8', NOW() - INTERVAL '150 days', NOW()),
-- kafka:7.5.2 (mq-server-01, Ubuntu based)
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:kafka123xyz456789012345678901234567kafka', 'confluentinc/cp-kafka', 'openjdk-17-jre-headless', '17.0.9+9-1', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '90 days', NOW()),
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:kafka123xyz456789012345678901234567kafka', 'confluentinc/cp-kafka', 'libc6', '2.35-0ubuntu3.5', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '90 days', NOW()),
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:kafka123xyz456789012345678901234567kafka', 'confluentinc/cp-kafka', 'libssl3', '3.0.2-0ubuntu1.12', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '90 days', NOW()),
('agent-008-c9d0e1f2', 'mq-server-01', '192.168.1.50', 'sha256:kafka123xyz456789012345678901234567kafka', 'confluentinc/cp-kafka', 'zlib1g', '1:1.2.11.dfsg-2ubuntu9.2', 'dpkg', 'ubuntu 22.04', NOW() - INTERVAL '90 days', NOW()),
-- company/worker:v1.8.0 (k8s-node-03, Alpine)
('agent-025-s7t8u9v0', 'k8s-node-03', '192.168.3.22', 'sha256:worker456opq789012345678901234567890123456789012345worke', 'company/worker', 'musl', '1.2.4_git20230717-r4', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'k8s-node-03', '192.168.3.22', 'sha256:worker456opq789012345678901234567890123456789012345worke', 'company/worker', 'libcrypto3', '3.1.3-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'k8s-node-03', '192.168.3.22', 'sha256:worker456opq789012345678901234567890123456789012345worke', 'company/worker', 'libssl3', '3.1.3-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW()),
('agent-025-s7t8u9v0', 'k8s-node-03', '192.168.3.22', 'sha256:worker456opq789012345678901234567890123456789012345worke', 'company/worker', 'ca-certificates', '20230506-r0', 'apk', 'alpine 3.18', NOW() - INTERVAL '50 days', NOW());
