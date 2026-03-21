# Agent 配置文件详解

本文档描述 Agent 及各插件的配置文件格式和配置项说明。

---

## 一、主配置文件

### 1.1 配置文件位置

按优先级查找：
1. 命令行 `-config` 参数指定
2. `/etc/cloudsec-agent/agent.yaml`
3. 当前目录 `agent.yaml`

### 1.2 配置项说明

**文件：** `agent.yaml`

```yaml
# gRPC Server 地址
server: "127.0.0.1:50051"

# 连接超时时间（秒）
connect_timeout: 30

# Agent 工作目录（存放运行时数据）
working_directory: "/var/run/cloudsec-agent"

# 插件目录
plugins_directory: "/opt/cloudsec/agent/plugins"

# 日志目录
log_directory: "/opt/cloudsec/agent/logs"

# 连接失败最大重试次数
retry_max_count: 10

# 重试间隔（秒）
retry_interval: 5

# 日志配置
log:
  level: "info"           # 日志级别: debug/info/warn/error
  file: ""                # 日志文件路径，空或 "stderr" 输出到 stderr
  max_size: 10            # 单文件最大 MB
  max_backups: 5          # 保留旧文件数
  compress: false         # 是否压缩旧文件
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| server | string | 必填 | gRPC Server 地址（standalone 模式可省略） |
| connect_timeout | int | 30 | 连接超时（秒） |
| working_directory | string | /var/run/cloudsec-agent | 工作目录 |
| plugins_directory | string | /opt/cloudsec/agent/plugins | 插件目录 |
| log_directory | string | /opt/cloudsec/agent/logs | 日志目录 |
| retry_max_count | int | 10 | 最大重试次数 |
| retry_interval | int | 5 | 重试间隔（秒） |
| log.level | string | info | 日志级别 |
| log.file | string | "" | 日志文件路径，空或 "stderr" 输出到 stderr |
| log.max_size | int | 10 | 单日志文件最大 MB |
| log.max_backups | int | 5 | 保留旧文件数 |
| log.compress | bool | false | 是否压缩旧文件 |

### 1.3 Standalone 模式配置

Standalone 模式在同一个 `agent.yaml` 中通过 `standalone:` 块配置，也可通过命令行参数 `-standalone` 启用。

```yaml
# agent.yaml 中的 standalone 配置块
working_directory: "/tmp/cloudsec-agent"
plugins_directory: "/opt/cloudsec/agent/plugins"

standalone:
  enabled: true
  output: "stderr"                   # "stderr" 或文件路径（如 "/tmp/results.json"）
  flush_interval: 1                  # 刷新间隔（秒）
  plugins:
    - ebpf_base_detector             # 要加载的插件列表
```

| 配置项 | 类型 | 说明 |
|--------|------|------|
| standalone.enabled | bool | 是否启用独立模式 |
| standalone.output | string | 输出方式：stderr 或文件路径 |
| standalone.flush_interval | int | 输出刷新间隔（秒） |
| standalone.plugins | []string | 要加载的插件列表 |

---

## 二、Detector 插件配置

配置目录：`/opt/cloudsec/agent/plugins/detector/config/rules/`

### 2.1 SSH 暴力破解检测

**文件：** `ssh_brute_force.yaml`

```yaml
ssh:
  enabled: true
  log_paths:
    - /var/log/auth.log
    - /var/log/secure
  rules:
    - name: auth_failure_brute_force
      description: "SSH认证失败暴力破解检测"
      pattern: 'Failed (password|publickey) for .* from (\S+)'
      action: failed
      frequency: 6          # 触发阈值
      timeframe: 120        # 时间窗口（秒）
      level: 10             # 告警级别
      ignore: 60            # 告警抑制时间（秒）
      group_by: source_ip
    - name: invalid_user_brute_force
      description: "SSH非法用户暴力破解���测"
      pattern: '(Invalid|Illegal) user .* from (\S+)'
      action: invalid_user
      frequency: 6
      timeframe: 120
      level: 10
      ignore: 60
      group_by: source_ip
  whitelist:
    - 127.0.0.1
    - "::1"
```

| 配置项 | 说明 |
|--------|------|
| enabled | 是否启用 |
| log_paths | 监控的日志文件路径 |
| whitelist | IP 白名单（不触发告警） |
| rules[].frequency | 触发告警的失败次数阈值 |
| rules[].timeframe | 统计时间窗口（秒） |
| rules[].ignore | 告警抑制时间（秒） |

### 2.2 FTP 暴力破解检测

**文件：** `ftp_brute_force.yaml`

```yaml
ftp:
  enabled: true
  log_paths:
    - /var/log/vsftpd.log
    - /var/log/xferlog
  rules:
    - name: auth_failure_brute_force
      description: "FTP认证失败暴力破解检测"
      action: failed
      frequency: 6
      timeframe: 120
      level: 10
      ignore: 60
      group_by: source_ip
    - name: multiple_connection_attempt
      description: "FTP多次连接尝试检测"
      action: connect
      frequency: 10
      timeframe: 60
      level: 10
      ignore: 60
      group_by: source_ip
  whitelist:
    - 127.0.0.1
    - "::1"
```

### 2.3 SSH 异常登录检测

**文件：** `ssh_anomaly_login.yaml`

```yaml
ssh_anomaly_login:
  enabled: false
  log_paths:
    - /var/log/auth.log
    - /var/log/secure
  alert_level: 8
  ignore_time: 300          # 告警抑制时间（秒）
  anomaly_rules: []
  # 示例规则配置:
  # anomaly_rules:
  #   - name: office_ips
  #     description: "办公网段IP"
  #     enabled: true
  #     ips:
  #       - 192.168.1.100
  #       - 192.168.1.101
  #       - 10.0.0.1
  #     users:           # 允许的用户列表，为空或不配置表示不限制用户
  #       - root
  #       - admin
  #       - deploy
  #
  #   - name: ops_ips
  #     description: "运维IP"
  #     enabled: true
  #     ips:
  #       - 10.10.10.50
  #       - 10.10.10.51
  #     users:
  #       - ops
  #       - root
```

**说明：** 从不在 `ips` 列表中的 IP 成功登录时触发告警。`users` 字段可限定仅当指定用户登录时才检查。

---

## 三、ebpf_base_detector 插件配置

配置目录：`/opt/cloudsec/agent/plugins/ebpf_base_detector/config/`

### 3.1 高危命令检测规则

**文件：** `dangerous_commands.yaml`

```yaml
version: "1.0"
description: "高危命令检测规则"

rules:
  - id: 2001
    name: "危险删除操作"
    description: "检测可能导致系统损坏的删除命令，如 rm -rf /"
    severity: critical
    enabled: true
    match:
      type: "regex"
      patterns:
        - 'rm\s+.*-rf\s+/'
        - 'rm\s+.*--no-preserve-root'
        - 'rm\s+-rf\s+/\*'
        - 'rm\s+-rf\s+~'

  - id: 2002
    name: "敏感文件访问"
    description: "检测对敏感系统文件的读取或修改操作"
    severity: high
    enabled: true
    category: "filesystem_operation"
    match:
      type: "regex"
      patterns:
        - 'cat\s+.*/etc/(passwd|shadow|sudoers)'
        - 'vi(m)?\s+.*/etc/(passwd|shadow|sudoers)'
        - 'nano\s+.*/etc/(passwd|shadow|sudoers)'
        - 'less\s+.*/etc/shadow'
        - 'more\s+.*/etc/shadow'
        - 'head\s+.*/etc/shadow'
        - 'tail\s+.*/etc/shadow'

  - id: 2003
    name: "危险权限修改"
    description: "检测危险的文件权限修改操作"
    severity: high
    enabled: true
    category: "permission_modify"
    match:
      type: "regex"
      patterns:
        - 'chmod\s+.*777\s+/'
        - 'chmod\s+.*\+s\s+'
        - 'chmod\s+4[0-7]{3}\s+'
        - 'chown\s+root:\s+'
        - 'chown\s+root:root\s+/'

  - id: 2009
    name: "内核模块操作"
    description: "检测内核模块的加载或卸载操作"
    severity: high
    enabled: true
    match:
      type: "prefix"
      patterns:
        - "insmod"
        - "rmmod"
        - "modprobe"
```

**规则结构：**

每条规则包含嵌套的 `match` 对象：

```go
// Rule 检测规则（rule_types.go）
type Rule struct {
    ID          int64  `yaml:"id"`
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
    Severity    string `yaml:"severity"`    // critical/high/medium/low
    Enabled     bool   `yaml:"enabled"`
    Category    string `yaml:"category"`    // 规则分类（如 file_delete/filesystem_operation/permission_modify）
    Match       Match  `yaml:"match"`       // 嵌套匹配配置
}

type Match struct {
    Type     string   `yaml:"type"`     // regex/contains/prefix/exact
    Patterns []string `yaml:"patterns"` // 匹配模式列表
}
```

**匹配类型：**

| match.type | 说明 | 示例 |
|------------|------|------|
| regex | 正则表达式匹配 | `rm\s+.*-rf\s+/` |
| prefix | 命令前缀匹配 | `insmod` |
| contains | 参数包含匹配 | `/etc/shadow` |
| exact | 精确匹配 | `shutdown` |

**严重程度：**

| severity | 说明 |
|----------|------|
| critical | 严重 |
| high | 高危 |
| medium | 中危 |
| low | 低危 |

### 3.2 完整规则列表

| 规则 ID | 名称 | 严重程度 | 匹配类型 | 分类 |
|--------|------|---------|----------|------|
| 2001 | 危险删除操作 | critical | regex | file_delete |
| 2002 | 敏感文件访问 | high | regex | filesystem_operation |
| 2003 | 危险权限修改 | high | regex | permission_modify |
| 2009 | 内核模块操作 | high | prefix | filesystem_operation |

### 3.3 敏感文件检测规则

**文件：** `sensitive_file_rules.yaml`

检测对系统敏感文件的创建或重命名操作，规则格式与高危命令相同。

| 规则 ID | 名称 | 严重程度 | 匹配类型 |
|--------|------|---------|----------|
| 1001 | Crontab 文件修改 | high | regex |
| 1002 | SSH 配置文件修改 | high | regex |
| 1003 | 动态链接器劫持 | critical | regex |
| 1004 | 系统服务文件创建 | high | regex |
| 1005 | 内核模块文件创建 | critical | regex |
| 1006 | Web Shell 文件创建 | critical | regex |
| 1007 | PAM 模块文件创建 | critical | regex |
| 1008 | passwd/shadow 文件修改 | critical | regex |

### 3.4 恶意请求检测规则

**文件：** `malicious_request_rules.yaml`

基于 IOC（威胁指标）检测恶意网络请求，用于 Connect 和 DNS 事件匹配。

```yaml
rules:
  - id: "IOC001"
    name: "已知矿池IP"
    threat_type: "mining"        # 威胁类型
    indicator_type: "ip"         # 指标类型
    severity: "high"
    enabled: true
    indicators:                  # IOC 指标列表
      - "94.23.23.52"
```

**支持的指标类型：**

| indicator_type | 说明 | 匹配的事件 |
|----------------|------|-----------|
| ip | IP 地址匹配 | Connect 事件的 remote_ip |
| domain | 域名匹配（支持通配符 `*`） | DNS 查询的 domain |
| port | 端口匹配 | Connect 事件的 remote_port |
| ip_port | IP:端口复合匹配 | Connect 事件的 remote_ip:remote_port |

**支持的威胁类型：** `mining`（挖矿）、`c2`（C2通信）、`phishing`（钓鱼）、`data_leakage`（数据泄露）

### 3.5 提权检测白名单

**文件：** `privilege_escalation_whitelist.yaml`

定义可信可执行文件列表，这些进程的提权行为不会触发告警。

```yaml
version: "1.0"
enabled: true
log_filtered_events: false  # 是否记录被过滤的事件
trusted_executables:
  - "/usr/bin/sudo"
  - "/usr/bin/su"
  - "/usr/bin/pkexec"
  # ...
```

### 3.6 文件监控白名单

**文件：** `file_monitor_whitelist.yaml`

定义不触发敏感文件告警的可信进程列表，格式与提权白名单相同。

```yaml
version: "1.0"
enabled: true
log_filtered_events: false
trusted_executables:
  - "/usr/bin/apt"
  - "/usr/bin/dpkg"
  - "/usr/bin/yum"
  # ...
```

### 3.7 容器高危命令检测规则

**文件：** `container_dangerous_commands.yaml`

容器内专用的高危命令检测规则，规则格式与宿主机高危命令相同。

| 规则 ID | 名称 | 严重程度 | 匹配类型 |
|--------|------|---------|----------|
| 3001 | 容器内包管理器安装 | medium | regex |
| 3002 | 容器内网络下载工具 | high | prefix |
| 3003 | 容器内系统配置修改 | high | regex |

### 3.8 容器敏感文件检测规则

**文件：** `container_sensitive_file_rules.yaml`

容器内专用的敏感文件检测规则，规则格式与宿主机敏感文件规则相同。

| 规则 ID | 名称 | 严重程度 | 匹配类型 |
|--------|------|---------|----------|
| 2001 | 容器内 passwd/shadow 文件修改 | critical | regex |
| 2005 | 容器内 DNS/hosts 篡改 | high | exact |

---

## 四、NIDS 插件配置

配置目录：`/opt/cloudsec/agent/plugins/nids/config/`

### 4.1 NIDS 配置文件

**文件：** `nids.yaml`

```yaml
# 抓包网卡（生产环境改为实际网卡如 eth0/ens33）
interface: "lo"
# BPF 过滤器（限定监控端口）
bpf_filter: "tcp port 80 or tcp port 8080"
# 抓包每帧最大字节数
snaplen: 65535
# TCP 流重组参数
tcp_reassembly:
  max_buffer_size: 262144   # 每个流最大缓冲区（字节）
  max_streams: 10000        # 最大并发流数
  stream_timeout: 120s      # 流超时时间
# Suricata 规则文件路径（相对于可执行文件目录）
rules_file: "config/nids.rules"
```

| 配置项 | 类型 | 说明 |
|--------|------|------|
| interface | string | 抓包网卡名称 |
| bpf_filter | string | BPF 过滤表达式 |
| snaplen | int | 每帧最大捕获字节数 |
| tcp_reassembly.max_buffer_size | int | TCP 流重组单流最大缓冲区（字节） |
| tcp_reassembly.max_streams | int | 最大并发流数 |
| tcp_reassembly.stream_timeout | duration | 流超时时间 |
| rules_file | string | Suricata 规则文件路径（相对路径） |

### 4.2 NIDS 规则文件

**文件：** `nids.rules`

使用 Suricata 兼容的规则语法，详见 Suricata 文档。

---

## 五、Scanner 插件配置

配置目录：`/opt/cloudsec/agent/plugins/scanner/config/`

**文件：** `scanner.yaml`

```yaml
scanner:
  engine:
    db_path: "/var/lib/clamav"      # ClamAV 病毒库路径
    max_file_size: 18874368         # 单文件最大扫描大小（字节，18MB）
    max_scan_time: 5                # 单文件最大扫描时间（秒）
  cronjob:
    dir_scan_interval: "24h"        # 目录扫描间隔
    proc_scan_interval: "1h"        # 进程扫描间隔
    throttle: "1s"                  # 扫描节流间隔
  scan_dirs:                        # 定期扫描目录列表
    - path: "/root"
      max_depth: 3
    - path: "/bin"
      max_depth: 2
    # ...
  filter:
    path_whitelist:                 # 路径白名单（跳过扫描）
      - "/dev"
      - "/proc"
      - "/sys"
    skip_file_types:                # 跳过的文件类型
      - "video"
      - "audio"
      - "image"
    min_file_size: 4                # 最小文件大小（字节）
    max_file_size: 18874368         # 最大文件大小（字节）
  fullscan:
    max_workers: 6                  # 全盘扫描最大并发数
    max_memory_mb: 512              # 全盘扫描最大内存（MB）
    max_cpu_percent: 600            # 最大 CPU 百分比（600 = 6 核）
    quick_timeout: "1h"             # 快速扫描超时
    full_timeout: "48h"             # 全盘扫描超时
  cgroup:
    enabled: true                   # 是否启用 cgroup 资源限制
    normal_memory_mb: 180           # 常规扫描内存限制（MB）
    normal_cpu_quota: 10000         # 常规扫描 CPU 配额
    fullscan_memory_mb: 512         # 全盘扫描内存限制（MB）
    fullscan_cpu_quota: 600000      # 全盘扫描 CPU 配额
```

---

## 六、命令行参数

### 6.1 Agent 命令行参数

```bash
./agent [options]
```

| 参数 | 说明 | 示例 |
|------|------|------|
| -config | 配置文件路径 | `-config=/etc/cloudsec-agent/agent.yaml` |
| -test | 测试模式（固定 Agent ID） | `-test` |
| -standalone | 启用独立模式 | `-standalone` |
| -output | 输出方式 (stderr/文件路径) | `-output=/opt/cloudsec/agent/logs/agent.log` |
| -plugins | 加载的插件列表 | `-plugins=ebpf_base_detector,collector` |

### 6.2 常用启动命令

```bash
# 正常模式
sudo ./agent -config=agent.yaml

# Standalone 模式（输出到 stderr）
sudo ./agent -standalone -plugins=ebpf_base_detector -output=stderr -test

# Standalone 模式（输出到文件）
sudo ./agent -standalone -plugins=ebpf_base_detector -output=/tmp/results.json -test
```

---

## 相关文档

- [架构设计文档](architecture.md) — 系统架构和模块职责
- [DataType 详细说明](data-types.md) — 各 DataType 的字段定义
- [功能测试文档](testing.md) — 各插件的测试流程
- [编译部署文档](build-deploy.md) — 配置文件的部署路径
