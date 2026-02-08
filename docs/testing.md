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

**配置文件说明：**

| 配置文件 | 路径 | 用途 |
|---------|------|------|
| 高危命令规则 | `/opt/cloudsec/plugins/driver/config/dangerous_commands.yaml` | 定义 12 条高危命令检测规则 |
| 可信任程序白名单 | `/opt/cloudsec/plugins/driver/config/trusted_executables.yaml` | 本地提权检测白名单（sudo、su 等） |

**数据类型（DataType）：**

| DataType | 用途 | 生成场景 | 日志级别 |
|----------|------|---------|---------|
| 59 | 基础进程执行事件 | 所有进程执行 | DEBUG/INFO |
| 6003 | 高危命令告警 | 匹配危险命令规则 | INFO |
| 60 | 本地提权告警 | 检测到非法提权 | WARN |

### 3.2 Standalone 模式说明

Standalone 模式允许不连接 gRPC Server 进行本地测试，检测结果输出到日志或文件。**在测试 driver 插件时，推荐优先使用此模式。**

**配置文件 (agent-standalone.yaml):**

```yaml
# Agent Standalone 模式配置
working_directory: "/opt/cloudsec/data/agent"
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

# 方式二：使用命令行参数 (推荐)
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

**日志输出说明:**

- **输出位置**: 标准错误（stderr）
- **日志格式**: 结构化日志（zap），控制台格式
- **日志级别**: INFO（高危命令）、WARN（本地提权）、ERROR（错误）
- **保存日志**: 可通过重定向保存到文件
  ```bash
  sudo ./build/agent -standalone -plugins=driver -output=log -test 2>&1 | tee driver.log
  ```

**日志输出示例:**

```
INFO  Dangerous command detected
    rule_id=DC001  rule_name=危险删除操作  severity=critical
    command=rm -rf /tmp/test_dir  matched_pattern=rm\s+.*-rf\s+/
    pid=727422  uid=0  comm=rm
```

### 3.3 高危命令检测

Driver 插件通过 eBPF 监控进程执行事件，在用户态使用规则引擎检测高危命令。

**检测原理:**
- eBPF 层: Hook `sched_process_exec` 捕获进程执行事件
- 用户态: 对命令行参数进行规则匹配（支持正则、包含、前缀、精确 4 种模式）
- 告警类型: DataType 6003

**Agent 启动命令:**

```bash
# 终端 1: 启动 agent（standalone 模式）
sudo ./build/agent -standalone -plugins=driver -output=log -test
```

**日志输出:**
- **位置**: 标准错误（stderr）
- **级别**: INFO
- **格式示例**:
  ```
  INFO  Dangerous command detected
      rule_id=DC001  rule_name=危险删除操作  severity=critical
      uid=0  comm=rm  args=rm -rf /tmp/test_dir
      matched_pattern=rm\s+.*-rf\s+/
  ```

**成功判断标准:**
1. 日志中出现 `INFO  Dangerous command detected`
2. 包含 `rule_id`、`rule_name`、`severity` 字段
3. `matched_pattern` 显示匹配的规则模式

---

#### 测试场景：12 条检测规则

以下每条规则提供 1-2 个测试命令。在启动 agent 后，在另一个终端执行测试命令。

**DC001 - 危险删除操作 (critical)**

```bash
# 测试命令 1: 删除临时目录（推荐）
mkdir -p /tmp/test_dir && rm -rf /tmp/test_dir
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC001  rule_name=危险删除操作  severity=critical
```

---

**DC002 - 敏感文件访问 (high)**

```bash
# 测试命令 1: 读取 shadow 文件
cat /etc/shadow

# 测试命令 2: 读取 passwd 文件
cat /etc/passwd
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC002  rule_name=敏感文件访问  severity=high
```

---

**DC003 - 危险权限修改 (high)**

```bash
# 测试命令 1: 设置 777 权限
touch /tmp/test_file && chmod 777 /tmp/test_file

# 测试命令 2: 设置 SUID 位
chmod 4755 /tmp/test_file
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC003  rule_name=危险权限修改  severity=high
```

---

**DC005 - 计划任务修改 (medium)**

```bash
# 测试命令 1: crontab 编辑
crontab -e

# 测试命令 2: 追加 crontab
echo "* * * * * /tmp/script" >> /tmp/test_crontab
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC005  rule_name=计划任务修改  severity=medium
```

---

**DC006 - 可疑安全工具 (medium)**

```bash
# 测试命令 1: nmap 扫描
nmap localhost

# 测试命令 2: 检查 nmap 版本
nmap --version
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC006  rule_name=可疑安全工具  severity=medium
```

---

**DC010 - 防火墙规则修改 (medium)**

```bash
# 测试命令 1: 清空 iptables 规则（需 root）
iptables -F

# 测试命令 2: 禁用 ufw
ufw disable
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC010  rule_name=防火墙规则修改  severity=medium
```

---

**DC011 - Base64解码执行 (high)**

```bash
# 测试命令: Base64 解码并执行
echo "ZWNobyB0ZXN0" | base64 -d | bash
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC011  rule_name=Base64解码执行  severity=high
```

---

**DC012 - 脚本语言危险执行 (high)**

```bash
# 测试命令 1: Python 执行系统命令
python -c 'import os; os.system("id")'

# 测试命令 2: Bash eval 执行
eval $(echo "id")
```

**预期日志:**
```
INFO  Dangerous command detected  rule_id=DC012  rule_name=脚本语言危险执行  severity=high
```

---

#### 配置文件

**规则配置文件:**
`/opt/cloudsec/plugins/driver/config/dangerous_commands.yaml`

**规则结构示例:**
```yaml
rules:
  - id: "DC001"
    name: "危险删除操作"
    description: "检测可能导致系统损坏的删除命令"
    severity: "critical"
    enabled: true
    match:
      type: "regex"                  # 匹配类型: regex/contains/prefix/exact
      patterns:
        - "rm\\s+.*-rf\\s+/"
        - "rm\\s+.*--no-preserve-root"
```

**修改规则:** 编辑配置文件后重启 agent 生效

---

### 3.4 本地提权检测

Driver 插件通过 eBPF Hook `kprobe/commit_creds` 检测进程的 uid/euid 提权行为，并在内核层进行白名单过滤。

**检测原理:**
- Hook 点: `kprobe/commit_creds` (内核凭证变更函数)
- 检测条件: 原 uid 和 euid 都非 0，新 uid 或 euid 为 0（提权到 root）
- 白名单过滤: 内核层过滤 sudo、su、pkexec 等合法提权程序
- 告警类型: DataType 60

**Agent 启动命令:**

```bash
# 终端 1: 启动 agent（standalone 模式）
sudo ./build/agent -standalone -plugins=driver -output=log -test
```

**日志输出:**
- **位置**: 标准错误（stderr）
- **级别**: WARN
- **格式示例**:
  ```
  WARN  Privilege escalation detected
      pid=12345  tgid=12345  ppid=12344  comm=privesc_test
      exe_path=/tmp/privesc_test  uid=1000
      old_uid=1000  old_euid=1000  new_uid=0  new_euid=0
  ```

**成功判断标准:**
1. 日志中出现 `WARN  Privilege escalation detected`
2. `old_uid` 和 `old_euid` 非 0
3. `new_uid` 或 `new_euid` 为 0
4. `exe_path` 不在白名单中

---

#### 模拟测试

由于本地提权涉及到特权操作，需要在授权的测试环境中进行。以下提供多种测试方法。

**方法一：使用 SUID 程序测试（推荐，无需编写代码）**

创建一个带 SUID 位的测试脚本包装器：

```bash
# 终端 2: 创建测试环境（需要 root 权限）

# 1. 创建一个简单的 C 包装器（会调用 setuid）
cat > /tmp/suid_wrapper.c << 'EOF'
#include <unistd.h>
#include <stdio.h>
int main() {
    printf("Before: uid=%d euid=%d\n", getuid(), geteuid());
    setuid(0);
    setgid(0);
    printf("After: uid=%d euid=%d\n", getuid(), geteuid());
    execl("/bin/bash", "bash", "-c", "id", NULL);
    return 0;
}
EOF

# 2. 编译并设置 SUID 位（需要 root）
gcc -o /tmp/suid_wrapper /tmp/suid_wrapper.c
sudo chown root:root /tmp/suid_wrapper
sudo chmod 4755 /tmp/suid_wrapper

# 3. 以普通用户身份运行（切换到普通用户）
su - haolipeng -c "/tmp/suid_wrapper"
```

**预期结果:**
- Agent 日志中出现 `WARN Privilege escalation detected`
- `exe_path=/tmp/suid_wrapper`
- `old_uid=1000` (你的用户 ID), `new_uid=0`

**清理测试环境:**
```bash
sudo rm -f /tmp/suid_wrapper /tmp/suid_wrapper.c
```

---

**方法二：使用已有的 SUID 程序（最简单）**

直接使用系统中已有的 SUID 程序进行测试：

```bash
# 查找系统中的 SUID 程序
find /usr/bin /bin /usr/sbin -perm -4000 -type f 2>/dev/null

# 示例：使用 passwd 命令（会改变凭证但在白名单中）
passwd  # 输入密码后取消

# 创建一个不在白名单中的 SUID 副本进行测试
sudo cp /usr/bin/passwd /tmp/test_passwd
sudo chmod 4755 /tmp/test_passwd
/tmp/test_passwd  # 这应该触发检测（如果不在白名单）
```

**预期结果:**
- 如果 `/tmp/test_passwd` 不在白名单，应该触发 `WARN Privilege escalation detected`

---

**方法三：使用 Python/Perl 一行命令（需要有漏洞的环境）**

在某些特殊配置的测试环境中，可以尝试：

```bash
# Python 调用 setuid（通常会失败，除非在特殊环境）
python3 -c "import os; os.setuid(0); os.system('id')"

# Perl 调用 setuid（通常会失败）
perl -e 'use POSIX; POSIX::setuid(0); system("id")'
```

**注意:** 这些命令在正常环境下会因权限不足而失败，只能在特殊测试环境中触发检测。

---

**方法四：白名单验证（验证不触发告警）**

测试合法提权程序不应触发告警：

```bash
# 终端 2: 使用 sudo（应该不触发告警）
sudo id

# 使用 su（应该不触发告警）
su - root -c "id"
```

**预期结果:** 日志中**不应该**出现 `WARN Privilege escalation detected`，因为 sudo 和 su 在白名单中。

---

**方法五：使用完整测试程序（最可靠）**

如果以上方法都无法触发，使用以下测试程序：

```c
/* privilege_escalation_test.c
 * 用途：测试本地提权检测功能
 * 编译：gcc -o privesc_test privilege_escalation_test.c
 * 运行：需要在具有 SUID 漏洞的环境中测试
 *
 * 注意：此程序仅用于测试，在正常环境中会因权限不足而失败
 */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>

int main() {
    printf("Current UID: %d, EUID: %d\n", getuid(), geteuid());
    printf("Attempting privilege escalation...\n");

    // 方法 1: 直接调用 setuid (仅在特定环境有效)
    if (setuid(0) == 0) {
        printf("✓ Successfully escalated to root via setuid\n");
        printf("  New UID: %d, EUID: %d\n", getuid(), geteuid());
        system("id");
        return 0;
    } else {
        perror("✗ setuid failed (expected in normal environment)");
    }

    // 方法 2: 测试 setreuid
    if (setreuid(0, 0) == 0) {
        printf("✓ Successfully escalated to root via setreuid\n");
        return 0;
    } else {
        perror("✗ setreuid failed (expected in normal environment)");
    }

    printf("\n⚠ Test requires a vulnerable environment (e.g., SUID exploit)\n");
    printf("   In normal conditions, all privilege escalation attempts will fail.\n");

    return 1;
}
```

**编译和运行:**

```bash
# 终端 2: 编译测试程序
gcc -o /tmp/privesc_test privilege_escalation_test.c

# 运行测试（正常环境下会失败，这是预期的）
/tmp/privesc_test
```

**注意事项:**
- 在正常环境中，此程序会因权限不足而失败，这是预期行为
- 要触发提权检测，需要在具有 SUID 漏洞的测试环境中运行
- 或者使用已知的 CVE 漏洞利用工具（仅限授权的安全测试环境）

---

#### 白名单配置

**配置文件:**
`/opt/cloudsec/plugins/driver/config/trusted_executables.yaml`

**默认白名单内容:**
```yaml
version: "1.0"
description: "Trusted executables whitelist for privilege escalation filtering"

# 可信任的可执行文件（绝对路径）
trusted_executables:
  - "/usr/bin/sudo"
  - "/usr/bin/su"
  - "/usr/bin/pkexec"
  - "/usr/lib/polkit-1/polkit-agent-helper-1"
  - "/usr/bin/doas"
  - "/usr/lib/systemd/systemd"
  - "/usr/lib/systemd/systemd-logind"
  - "/usr/sbin/unix_chkpwd"

enabled: true
log_filtered_events: false       # 是否记录被过滤的事件
```

**白名单匹配机制:**
- 基于可执行文件的**绝对路径**（通过 eBPF dentry 链遍历获取）
- 使用 Murmur OAAT64 哈希算法进行快速匹配
- 内核层直接过滤，不产生用户态事件
- 路径最深支持 16 层目录，最长 255 字节

**修改白名单:** 编辑配置文件后重启 agent 生效

---

#### 调试技巧

如果需要查看内核层的调试信息：

```bash
# 查看 eBPF 内核日志（需要 root 权限）
sudo cat /sys/kernel/debug/tracing/trace_pipe | grep hids
```

**内��日志示例:**
```
hids: commit_creds pid=12345 tgid=12345 ppid=12344
hids: commit_creds uid=1000 old_uid=1000 old_euid=1000
hids: commit_creds new_uid=0 new_euid=0
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

