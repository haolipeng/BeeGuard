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
plugins_directory: "/opt/cloudsec/plugins"

# 日志目录
log_directory: "/opt/cloudsec/logs"

# 连接失败最大重试次数
retry_max_count: 10

# 重试间隔（秒）
retry_interval: 5

# 日志配置
log:
  level: "info"           # 日志级别: debug/info/warn/error
  max_size: 10            # 单文件最大 MB
  max_backups: 5          # 保留旧文件数
  compress: false         # 是否压缩旧文件
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| server | string | 必填 | gRPC Server 地址（standalone 模式可省略） |
| connect_timeout | int | 30 | 连接超时（秒） |
| working_directory | string | /var/run/cloudsec-agent | 工作目录 |
| plugins_directory | string | /opt/cloudsec/plugins | 插件目录 |
| log_directory | string | /opt/cloudsec/logs | 日志目录 |
| retry_max_count | int | 10 | 最大重试次数 |
| retry_interval | int | 5 | 重试间隔（秒） |
| log.level | string | info | 日志级别 |
| log.max_size | int | 10 | 单日志文件最大 MB |
| log.max_backups | int | 5 | 保留旧文件数 |
| log.compress | bool | false | 是否压缩旧文件 |

### 1.3 Standalone 模式配置

Standalone 模式在同一个 `agent.yaml` 中通过 `standalone:` 块配置，也可通过命令行参数 `-standalone` 启用。

```yaml
# agent.yaml 中的 standalone 配置块
working_directory: "/tmp/cloudsec-agent"
plugins_directory: "/opt/cloudsec/plugins"

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

配置目录：`/opt/cloudsec/plugins/detector/config/rules/`

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
  #     description: "办公网段 IP"
  #     enabled: true
  #     ips:
  #       - 192.168.1.100
  #       - 10.0.0.1
```

**说明：** 从不在 `ips` 列表中的 IP 成功登录时触发告警。

---

## 三、ebpf_base_detector 插件配置

配置目录：`/opt/cloudsec/plugins/ebpf_base_detector/config/`

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
    match:
      type: "regex"
      patterns:
        - 'cat\s+.*/etc/(passwd|shadow|sudoers)'
        - 'vi(m)?\s+.*/etc/(passwd|shadow|sudoers)'

  - id: 2003
    name: "危险权限修改"
    description: "检测危险的文件权限修改操作"
    severity: high
    enabled: true
    match:
      type: "regex"
      patterns:
        - 'chmod\s+.*777\s+/'
        - 'chmod\s+.*\+s\s+'

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

| 规则 ID | 名称 | 严重程度 | 匹配类型 |
|--------|------|---------|----------|
| 2001 | 危险删除操作 | critical | regex |
| 2002 | 敏感文件访问 | high | regex |
| 2003 | 危险权限修改 | high | regex |
| 2009 | 内核模块操作 | high | prefix |

---

## 四、命令行参数

### 4.1 Agent 命令行参数

```bash
./agent [options]
```

| 参数 | 说明 | 示例 |
|------|------|------|
| -config | 配置文件路径 | `-config=/etc/cloudsec-agent/agent.yaml` |
| -test | 测试模式（固定 Agent ID） | `-test` |
| -standalone | 启用独立模式 | `-standalone` |
| -output | 输出方式 (stderr/文件路径) | `-output=/opt/cloudsec/logs/agent.log` |
| -plugins | 加载的插件列表 | `-plugins=ebpf_base_detector,collector` |

### 4.2 常用启动命令

```bash
# 正常模式
sudo ./agent -config=agent.yaml

# Standalone 模式（输出到 stderr）
sudo ./agent -standalone -plugins=ebpf_base_detector -output=stderr -test

# Standalone 模式（输出到文件）
sudo ./agent -standalone -plugins=ebpf_base_detector -output=/tmp/results.json -test
```
