# Agent 功能测试文档

本文档描述 Agent 及其插件的完整测试流程，包括自动化测试和手动验证。

---

## 一、概述

### 测试类型

| 类型 | 说明 | 执行方式 |
|------|------|---------|
| 单元测试 | 模块级功能验证 | `make test` |
| E2E 测试 | 端到端插件流程测试 | `make test-e2e` |
| 手动验证 | 功能完整性验证 | 手动执行 |

### 测试覆盖范围

| 插件 | 单元测试 | E2E 测试 | 手动验证 |
|------|---------|---------|---------|
| collector | port_test.go | 自动化 | - |
| baseline | - | 自动化 | - |
| detector | - | - | 需手动验证 |
| driver | - | - | 需手动验证 |

---

## 二、Detector 插件手动验证

Detector 插件目前没有自动化测试，需要手动验证。

### 2.1 准备工作

```bash
# 1. 编译并部署
make build-plugins
make deploy-plugins

# 2. 确认插件已部署
ls -la /opt/cloudsec/plugins/detector/
```

### 2.2 SSH 暴力破解检测

**注意:** 默认配置中 `127.0.0.1` 和 `::1` 在白名单内，本地测试需先移除白名单。

**验证步骤:**

```bash
# 1. 修改配置，移除白名单 (本地测试必须)
cat > /tmp/ssh_test.yaml << 'EOF'
ssh:
  enabled: true
  log_paths:
    - /var/log/auth.log
    - /var/log/secure
  whitelist: []  # 清空白名单
  rules:
    - name: auth_failure_brute_force
      description: "SSH认证失败暴力破解检测"
      action: failed
      frequency: 6
      timeframe: 120
      level: 10
      ignore: 60
      group_by: source_ip
EOF
# 将此配置复制到 /opt/cloudsec/plugins/detector/config/rules/ssh_brute_force.yaml

# 2. 终端 1: 查看日志
tail -f /opt/cloudsec/logs/plugins/detector/detector.log

# 3. 终端 2: 模拟 SSH 密码错误 (6次以上触发告警)
for i in {1..10}; do
  ssh -o BatchMode=yes -o ConnectTimeout=1 root@localhost 2>/dev/null
  sleep 1
done
```

**预期结果:**

日志中出现暴力破解告警：
```
INFO    brute force detected    {"source_ip": "127.0.0.1", "count": 6, "rule": "auth_failure_brute_force"}
```

**告警数据类型:** DataType 6001

**配置说明:**

默认规则 (config/rules/ssh_brute_force.yaml):
- 检测阈值: 6 次失败
- 时间窗口: 120 秒
- 告警抑制: 60 秒
- 默认白名单: 127.0.0.1, ::1 (本地测试需移除)

### 2.3 FTP 暴力破解检测

**前提条件:** 需要安装并启动 vsftpd

```bash
# 安装 vsftpd
apt install vsftpd

# 启动服务
systemctl start vsftpd
```

**验证步骤:**

```bash
# 模拟 FTP 登录失败
for i in {1..10}; do
  curl -u wronguser:wrongpass ftp://localhost/ 2>/dev/null
  sleep 1
done

# 检查日志
tail -f /opt/cloudsec/logs/plugins/detector/detector.log
```

**预期结果:** 日志中出现 FTP 暴力破解告警

**告警数据类型:** DataType 6002

### 2.4 SSH 异常登录检测

**前提条件:** 配置异常登录规则

```bash
# 查看/编辑规则配置
cat /opt/cloudsec/plugins/detector/config/rules/ssh_anomaly_login.yaml
```

**配置示例:**

```yaml
ssh_anomaly_login:
  enabled: true
  log_paths:
    - /var/log/auth.log
    - /var/log/secure
  alert_level: 8
  ignore_time: 300
  anomaly_rules:
    - name: office_ips
      description: "办公网段 IP"
      enabled: true
      ips:
        - 192.168.1.100
        - 10.0.0.1
```

**验证步骤:**

```bash
# 从非白名单 IP 成功登录 SSH
# (需要从另一台机器，IP 不在白名单中)
ssh user@target_host

# 检查日志
tail -f /opt/cloudsec/logs/plugins/detector/detector.log
```

**预期结果:**

```
INFO    anomaly login detected  {"user": "root", "source_ip": "45.33.32.156", "service": "ssh"}
```

**告警数据类型:** DataType 6005

**注意:** 如果未配置任何规则，不会产生告警。

---

## 三、Driver 插件手动验证

Driver 插件基于 eBPF，**必须使用 root 权限**运行。

### 3.1 准备工作

```bash
# 1. 检查 eBPF 环境
ls /sys/kernel/btf/vmlinux  # BTF 支持
cat /proc/version           # 内核版本 >= 5.x

# 2. 编译并部署
make build-driver
make deploy-driver

# 3. 确认部署
ls -la /opt/cloudsec/plugins/driver/
ls -la /opt/cloudsec/plugins/driver/config/
```

### 3.2 Standalone 模式测试

Standalone 模式允许不连接 gRPC Server 进行本地测试，检测结果输出到日志或文件。**在测试 driver 插件时，推荐优先使用此模式。**

**配置文件 (agent-standalone.yaml):**

```yaml
# Agent Standalone 模式配置
working_directory: "/tmp/cloudsec-agent"
plugins_directory: "/opt/cloudsec/plugins"

standalone:
  enabled: true
  output: "log"                    # "log" 或 "file"
  output_path: "/tmp/cloudsec-detection-results.json"
  flush_interval: 1                # 刷新间隔（秒）
  plugins:
    - driver                       # 仅加载 driver 插件
```

**启动方式:**

```bash
# 方式一：使用配置文件
sudo ./build/agent -config=agent-standalone.yaml -test

# 方式二：使用命令行参数
sudo ./build/agent -standalone -plugins=driver -output=log -test
```

**命令行参数:**

| 参数 | 说明 | 示例 |
|------|------|------|
| `-config` | 配置文件路径 | `-config=agent-standalone.yaml` |
| `-standalone` | 启用 standalone 模式 | `-standalone` |
| `-output` | 输出方式 (log/file) | `-output=log` |
| `-output-path` | 输出文件路径 | `-output-path=/tmp/results.json` |
| `-plugins` | 加载的插件列表 | `-plugins=driver` |
| `-test` | 测试模式（固定 agent ID） | `-test` |

**日志输出示例:**

```
INFO  standalone/output.go:151  dangerous command detected
    {"rule_id": "DC001", "rule_name": "危险删除操作", "severity": "critical",
     "command": "rm -rf /tmp/test_nonexistent_dir",
     "matched_pattern": "rm\\s+.*-rf\\s+/", "pid": "727422", "uid": "0"}
```

### 3.3 eBPF 进程监控验证

**验证步骤:**

```bash
# 终端 1: 以 standalone 模式启动
sudo ./build/agent -standalone -plugins=driver -output=log -test

# 终端 2: 执行一些命令
ls /tmp
whoami
cat /etc/hostname
```

**预期结果:**

终端 1 输出进程执行事件：
```
[EXEC] pid=12345 ppid=1000 uid=0 comm=ls exe=/usr/bin/ls args=ls /tmp
[EXEC] pid=12346 ppid=1000 uid=0 comm=whoami exe=/usr/bin/whoami args=whoami
[EXEC] pid=12347 ppid=1000 uid=0 comm=cat exe=/usr/bin/cat args=cat /etc/hostname
```

**数据字段说明:**
- pid: 进程 ID
- ppid: 父进程 ID
- uid: 用户 ID
- comm: 命令名 (最多 16 字符)
- exe: 可执行文件路径
- args: 命令行参数

### 3.4 高危命令检测验证

**验证步骤:**

```bash
# 终端 1: 以 standalone 模式启动
sudo ./build/agent -standalone -plugins=driver -output=log -test

# 终端 2: 执行高危命令 (测试环境，注意安全)
# 示例 1: 敏感文件访问
cat /etc/shadow

# 示例 2: 历史记录清除
history -c

# 示例 3: 可疑工具
nmap --version
```

**预期结果:**

触发高危命令告警：
```
[ALERT] Dangerous command detected!
  Rule: DC002 - 敏感文件访问
  Severity: high
  Command: cat /etc/shadow
```

**规则配置文件:**

`/opt/cloudsec/plugins/driver/config/dangerous_commands.yaml`

**主要规则类型:**

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

### 3.5 Standalone 模式测试命令汇总

以下命令可用于快速验证各类检测规则：

```bash
# 终端 1: 启动 standalone 模式
sudo ./build/agent -standalone -plugins=driver -output=log -test

# 终端 2: 执行测试命令
rm -rf /tmp/test_nonexistent_dir   # DC001 - 危险删除
cat /etc/passwd                     # DC002 - 敏感文件访问
which nmap                          # DC006 - 可疑安全工具
modprobe --version                  # DC009 - 内核模块操作
```

---

## 四、单元测试

### 执行方式

```bash
cd /home/work/goProject/src/company/agent

# 运行所有单元测试
make test

# 运行指定包的测试
go test -v ./host/...
go test -v ./agent/...
go test -v ./transport/...

# 运行指定测试函数
go test -v ./host/... -run TestHostname
```

### 测试模块说明

| 模块 | 测试文件 | 测试内容 |
|------|---------|---------|
| host | host_test.go, platform_test.go | 主机信息采集 |
| agent | id_test.go, state_test.go | Agent ID 生成、状态管理 |
| transport | connection_test.go, stats_handler_test.go | gRPC 连接、统计 |
| collector/port | port_test.go | 端口采集、IP 解析 |

### 开发者调试技巧

```bash
# 显示详细输出
go test -v ./... 2>&1 | tee test.log

# 只编译不运行
go test -c ./host/...

# 覆盖率报告
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 五、E2E 测试

### 5.1 Baseline 插件测试

**执行方式:**

```bash
# 方式一：使用 Makefile (推荐)
make test-e2e-baseline

# 方式二：直接执行脚本
cd tests/e2e/baseline
chmod +x test.sh
./test.sh
```

**测试流程:**
1. 编译 baseline 插件
2. 复制到 `/tmp/plugin/baseline/`
3. 启动测试程序，发送测试任务
4. 接收并验证结果

**预期结果:**
```
========== Baseline Check Result ==========
Baseline ID: 1200
Status: success
Token: test-token-123
Check Items Count: 3
  [1] CheckID: 1001, Result: PASS
  [2] CheckID: 1002, Result: PASS
  [3] CheckID: 1003, Result: FAIL
==========================================

========== Task Status ==========
Status: succeed
Token: test-token-123
================================
```

**关键数据类型:**
- DataType 8000: 基线检查结果
- DataType 8010: 任务状态

### 5.2 Collector 插件测试

**执行方式:**

```bash
make test-e2e-collector
```

**测试内容:**
- 5050: 进程采集
- 5051: 端口采集
- 5052: 用户采集
- 5054: 系统服务采集
- 5055: 软件采集
- 5056: 容器采集
- 5060: Web 服务采集
- 5061: 数据库服务采集
- 5062: 内核模块采集

**预期结果:**

收到各类采集数据，格式化输出到控制台。

