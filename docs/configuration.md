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
working_directory: "/opt/cloudsec/data/agent"

# 插件目录
plugins_directory: "/opt/cloudsec/plugins"

# 连接失败最大重试次数
retry_max_count: 10

# 重试间隔（秒）
retry_interval: 5
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| server | string | 必填 | gRPC Server 地址 |
| connect_timeout | int | 30 | 连接超时（秒） |
| working_directory | string | /opt/cloudsec/data/agent | 工作目录 |
| plugins_directory | string | /opt/cloudsec/plugins | 插件目录 |
| retry_max_count | int | 10 | 最大重试次数 |
| retry_interval | int | 5 | 重试间隔（秒） |

### 1.3 Standalone 模式配置

**文件：** `agent-standalone.yaml`

```yaml
working_directory: "/tmp/cloudsec-agent"
plugins_directory: "/opt/cloudsec/plugins"

standalone:
  enabled: true
  output: "log"                    # "log" 或 "file"
  output_path: "/tmp/cloudsec-detection-results.json"
  flush_interval: 1                # 刷新间隔（秒）
  plugins:
    - driver                       # 要加载的插件列表
```

| 配置项 | 类型 | 说明 |
|--------|------|------|
| standalone.enabled | bool | 是否启用独立模式 |
| standalone.output | string | 输出方式：log 或 file |
| standalone.output_path | string | 输出文件路径（output=file 时有效） |
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
  whitelist:
    - "127.0.0.1"
    - "::1"
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
    - /var/log/auth.log
  whitelist:
    - "127.0.0.1"
  rules:
    - name: ftp_auth_failure_brute_force
      description: "FTP认证失败暴力破解检测"
      action: failed
      frequency: 6
      timeframe: 120
      level: 10
      ignore: 60
      group_by: source_ip
```

### 2.3 SSH 异常登录检测

**文件：** `ssh_anomaly_login.yaml`

```yaml
ssh_anomaly_login:
  enabled: true
  log_paths:
    - /var/log/auth.log
    - /var/log/secure
  alert_level: 8
  ignore_time: 300          # 告警抑制时间（秒）
  anomaly_rules:
    - name: office_ips
      description: "办公网段 IP"
      enabled: true
      ips:                   # IP 白名单
        - 192.168.1.100
        - 10.0.0.1
```

**说明：** 从不在 `ips` 列表中的 IP 成功登录时触发告警。

---

## 三、Driver 插件配置

配置目录：`/opt/cloudsec/plugins/driver/config/`

### 3.1 高危命令检测规则

**文件：** `dangerous_commands.yaml`

```yaml
rules:
  - id: DC001
    name: "危险删除操作"
    description: "检测 rm -rf 等危险删除命令"
    severity: critical
    match_type: regex
    patterns:
      - 'rm\s+.*-rf\s+/'
      - 'rm\s+-rf\s+/'

  - id: DC002
    name: "敏感文件访问"
    description: "检测访问敏感文件的命令"
    severity: high
    match_type: contains
    patterns:
      - /etc/shadow
      - /etc/passwd

  - id: DC006
    name: "可疑安全工具"
    description: "检测安全扫描工具"
    severity: medium
    match_type: prefix
    commands:
      - nmap
      - masscan
      - hydra
```

**匹配类型：**

| match_type | 说明 | 示例 |
|------------|------|------|
| regex | 正则表达式匹配 | `rm\s+.*-rf\s+/` |
| prefix | 命令前缀匹配 | `nmap` |
| contains | 参数包含匹配 | `/etc/shadow` |

**严重程度：**

| severity | 说明 |
|----------|------|
| critical | 严重 |
| high | 高危 |
| medium | 中危 |
| low | 低危 |

### 3.2 完整规则列表

| 规则 ID | 名称 | 严重程度 |
|--------|------|---------|
| DC001 | 危险删除操作 | critical |
| DC002 | 敏感文件访问 | high |
| DC003 | 危险权限修改 | high |
| DC004 | 下载并执行 | critical |
| DC005 | 计划任务修改 | medium |
| DC006 | 可疑安全工具 | medium |
| DC007 | SSH 密钥操作 | high |
| DC008 | 历史记录清除 | medium |
| DC009 | 内核模块操作 | high |
| DC010 | 防火墙规则修改 | medium |
| DC011 | Base64 解码执行 | high |
| DC012 | 脚本语言危险执行 | high |

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
| -output | 输出方式 (log/file) | `-output=log` |
| -output-path | 输出文件路径 | `-output-path=/tmp/results.json` |
| -plugins | 加载的插件列表 | `-plugins=driver,collector` |

### 4.2 常用启动命令

```bash
# 正常模式
sudo ./agent -config=agent.yaml

# Standalone 模式（日志输出）
sudo ./agent -standalone -plugins=driver -output=log -test

# Standalone 模式（文件输出）
sudo ./agent -standalone -plugins=driver -output=file -output-path=/tmp/results.json -test

# 使用配置文件启动 Standalone
sudo ./agent -config=agent-standalone.yaml -test
```
