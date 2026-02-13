# 高危命令检测 — 手动测试指南

## 概述

本文档描述如何手动验证 driver 插件的高危命令检测功能（DataType 6003）。

**检测原理**：在 `sched_process_exec` Hook 中捕获所有进程执行事件，将命令行参数与 `dangerous_commands.yaml` 中的规则进行匹配。匹配成功时产生告警。

**匹配方式**：

| 匹配类型 | 说明 | 示例 |
|----------|------|------|
| `regex` | 正则表达式匹配命令行参数 | `rm\s+.*-rf\s+/` |
| `contains` | 命令名包含指定字符串 | `nmap` |
| `prefix` | 命令名以指定字符串开头 | `insmod` |

**关键源文件**：

| 文件 | 说明 |
|------|------|
| `business_plugins/driver/config/dangerous_commands.yaml` | 检测规则配置（12 条规则） |
| `business_plugins/driver/detector/detector.go` | 规则匹配引擎 |
| `business_plugins/driver/main.go` | 事件处理与告警生成 |

---

## 环境要求

| 项目 | 要求 |
|------|------|
| 内核版本 | >= 5.x |
| BTF 支持 | `/sys/kernel/btf/vmlinux` 存在 |
| 编译依赖 | clang、llvm、libbpf-dev、linux-headers |
| 运行权限 | root |

---

## 编译与启动

```bash
# 1. 编译译并部署
cd /home/work/goProject/src/company/agent
make build
make deploy

# 2. 启动 Agent（Terminal A）
# 检测事件输出到 stderr，Agent 运行日志输出到 /opt/cloudsec/logs/agent.log
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=driver -output=stderr -test
```

**可选**：输出到文件以便后续分析：

```bash
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=driver -output=/tmp/detection.json -test
```

---

## 测试用例

在另一个终端（Terminal B）中执行以下命令。每条命令执行后，在 Terminal A 中观察是否输出告警。

> **安全提示**：以下测试命令经过精心设计，使用不存在的路径或 `--help` 等方式避免实际危害。请在隔离的测试环境中执行。

---

### DC001: 危险删除操作（critical）

```bash
# 测试 1：rm -rf /（使用不存在的子目录，安全）
rm -rf /tmp/dc001_nonexistent_test_dir

# 测试 2：rm --no-preserve-root（仅触发规则匹配，实际不会执行危险操作）
echo "rm --no-preserve-root /tmp/test" | cat
```

**预期告警**：

```
rule_id=DC001  rule_name=危险删除操作  severity=critical
```

> **注意**：规则基于命令行参数正则匹配，`rm -rf /tmp/dc001_nonexistent_test_dir` 会匹配 `rm\s+.*-rf\s+/` 模式。

---

### DC002: 敏感文件访问（high）

```bash
# 测试 1：查看 passwd 文件（安全操作）
cat /etc/passwd > /dev/null

# 测试 2：查看 shadow 文件（需要 root）
cat /etc/shadow > /dev/null
```

**预期告警**：

```
rule_id=DC002  rule_name=敏感文件访问  severity=high
```

---

### DC003: 危险权限修改（high）

```bash
# 测试 1：chmod 777（使用临时文件）
touch /tmp/dc003_test && chmod 777 /tmp/dc003_test && rm -f /tmp/dc003_test

# 测试 2：设置 SUID 位（使用临时文件）
touch /tmp/dc003_suid && chmod +s /tmp/dc003_suid && rm -f /tmp/dc003_suid
```

**预期告警**：

```
rule_id=DC003  rule_name=危险权限修改  severity=high
```

---

### DC004: 下载并执行（critical）

```bash
# 测试 1：curl | bash 模式（连接不存在的地址，不会实际下载）
curl http://127.0.0.1:1/test.sh 2>/dev/null | bash 2>/dev/null; true

# 测试 2：wget 管道模式
wget http://127.0.0.1:1/test.sh -O - 2>/dev/null | bash 2>/dev/null; true
```

**预期告警**：

```
rule_id=DC004  rule_name=下载并执行  severity=critical
```

> **说明**：规则匹配的是命令行参数模式，即使实际下载失败也会触发告警。但由于 `curl | bash` 是管道组合，eBPF 捕获的是各个进程的独立 execve 事件，实际告警可能仅匹配到 `curl` 的参数部分。建议同时观察 `curl` 和 `bash` 的事件。

---

### DC005: 计划任务修改（medium）

```bash
# 测试 1：列出当前 crontab（安全操作，但 crontab -e 会触发）
crontab -e <<< ""
# 按 :q! 退出 vim 不保存

# 测试 2：echo 写入 cron 目录（使用不存在的文件）
echo "test" >> /etc/cron.d/dc005_test 2>/dev/null; rm -f /etc/cron.d/dc005_test
```

**预期告警**���

```
rule_id=DC005  rule_name=计划任务修改  severity=medium
```

---

### DC006: 可疑安全工具（medium）

```bash
# 测试 1：nmap（如已安装）
nmap --version 2>/dev/null || echo "nmap not installed, skip"

# 测试 2：masscan（如已安装）
masscan --help 2>/dev/null || echo "masscan not installed, skip"

# 测试 3：使用包含关键字的命令（无需安装工具）
/bin/echo "testing nmap detection"
```

**预期告警**：

```
rule_id=DC006  rule_name=可疑安全工具  severity=medium
```

> **说明**：`contains` 匹配类型会检查命令名是否包含关键字。`nmap --version` 会触发，但 `echo "nmap"` 不会（echo 的命令名不包含 nmap）。

---

### DC007: SSH 密钥操作（high）

```bash
# 测试 1：向 authorized_keys 追加内容（使用临时目录）
mkdir -p /tmp/dc007_ssh && echo "test-key" >> /tmp/dc007_ssh/authorized_keys && rm -rf /tmp/dc007_ssh

# 测试 2：复制私钥
cp /dev/null /tmp/dc007_id_rsa 2>/dev/null; rm -f /tmp/dc007_id_rsa
```

**预期告警**：

```
rule_id=DC007  rule_name=SSH密钥操作  severity=high
```

> **注意**：规则匹配命令行参数中的 `.ssh/authorized_keys` 或 `.ssh/id_rsa` 路径模式。测试用临时路径可能不包含 `.ssh/`，需要根据实际正则调整测试命令。

---

### DC008: 历史记录清除（medium）

```bash
# 测试 1：清除历史记录
history -c

# 测试 2：设置 HISTSIZE
export HISTSIZE=0

# 测试 3：unset HISTFILE
unset HISTFILE
```

**预期告警**：

```
rule_id=DC008  rule_name=历史记录清除  severity=medium
```

> **说明**：`history -c` 和 `export HISTSIZE=0` 是 bash 内建命令，不会触发 execve。只有通过 `bash -c "history -c"` 等方式启动新进程时才能被 eBPF 捕获。可改用以下方式测试：

```bash
bash -c "history -c"
bash -c "unset HISTFILE"
```

---

### DC009: 内核模块操作（high）

```bash
# 测试 1：列出已加载的模块（安全操作，但 modprobe 命令名匹配 prefix 规则）
modprobe --show-depends ext4

# 测试 2：insmod（使用不存在的模块，会报错但触发检测）
insmod /tmp/nonexistent.ko 2>/dev/null; true
```

**预期告警**：

```
rule_id=DC009  rule_name=内核模块操作  severity=high
```

---

### DC010: 防火墙规则修改（medium）

```bash
# 测试 1：列出 iptables 规则（安全操作）
# 注意：iptables -F 会清空规则，测试环境中谨慎执行
iptables -L -n 2>/dev/null; true

# 测试 2：ufw 状态查看（安全操作，但 "ufw disable" 会触发）
ufw status 2>/dev/null; true
```

> **注意**：`iptables -F` 和 `ufw disable` 会实际修改防火墙规则，仅在隔离环境中测试。建议使用以下安全方式验证匹配规则：

```bash
# 仅验证规则匹配，不实际执行
bash -c "echo iptables -F"
```

**预期告警**（仅当执行实际 iptables -F 时）：

```
rule_id=DC010  rule_name=防火墙规则修改  severity=medium
```

---

### DC011: Base64 解码执行（high）

```bash
# 测试 1：base64 解码并执行（解码内容为 "echo hello"）
echo "ZWNobyBoZWxsbw==" | base64 -d | bash

# 测试 2：echo + base64 -d + bash 管道
echo "ZWNobyB0ZXN0" | base64 -d | bash
```

**预期告警**：

```
rule_id=DC011  rule_name=Base64解码执行  severity=high
```

> **说明**：与 DC004 类似，管道命令的每个部分是独立的 execve 事件。告警可能匹配到 `base64 -d` 的参数部分。

---

### DC012: 脚本语言危险执行（high）

```bash
# 测试 1：Python exec（安全内容）
python3 -c "exec('print(1+1)')"

# 测试 2：Python import os（安全内容）
python3 -c "import os; print(os.getpid())"

# 测试 3：Perl system 调用
perl -e 'system("echo perl_test")'
```

**预期告警**：

```
rule_id=DC012  rule_name=脚本语言危险执行  severity=high
```

---

## 验证告警字段

在 Terminal A 的输出中，确认每条告警包含以下关键字段：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `rule_id` | 规则 ID | DC001 |
| `rule_name` | 规则名称 | 危险删除操作 |
| `severity` | 严重程度 | critical / high / medium |
| `matched_pattern` | 匹配的模式 | `rm\s+.*-rf\s+/` |
| `command` | 完整命令行 | `rm -rf /tmp/test` |
| `pid` | 进程 ID | 12345 |
| `uid` | 用户 ID | 0 |
| `exe_path` | 可执行文件路径 | /usr/bin/rm |
| `privilege_level` | 权限级别 | root / normal |
| `detection_type` | 检测类型 | dangerous_command |

---

## 测试结果记录表

| # | 规则 ID | 规则名称 | 严重程度 | 测试命令 | 预期 | 实际 | 备注 |
|---|---------|----------|----------|----------|------|------|------|
| 1 | DC001 | 危险删除操作 | critical | `rm -rf /tmp/dc001_test` | 告警 | | |
| 2 | DC002 | 敏感文件访问 | high | `cat /etc/passwd` | 告警 | | |
| 3 | DC003 | 危险权限修改 | high | `chmod 777 /tmp/test` | 告警 | | |
| 4 | DC004 | 下载并执行 | critical | `curl ... \| bash` | 告警 | | 管道命令注意 |
| 5 | DC005 | 计划任务修改 | medium | `crontab -e` | 告警 | | |
| 6 | DC006 | 可疑安全工具 | medium | `nmap --version` | 告警 | | 需要安装 nmap |
| 7 | DC007 | SSH密钥操作 | high | `echo >> .ssh/authorized_keys` | 告警 | | |
| 8 | DC008 | 历史记录清除 | medium | `bash -c "history -c"` | 告警 | | 内建命令需新进程 |
| 9 | DC009 | 内核模块操作 | high | `insmod /tmp/test.ko` | 告警 | | |
| 10 | DC010 | 防火墙规则修改 | medium | `iptables -F` | 告警 | | 隔离环境 |
| 11 | DC011 | Base64解码执行 | high | `base64 -d \| bash` | 告警 | | 管道命令注意 |
| 12 | DC012 | 脚本语言危险执行 | high | `python3 -c "import os"` | 告警 | | |

---

## 常见问题排查

| 问题 | 排查方法 |
|------|----------|
| Agent 启动失败 `failed to load eBPF` | 确认 root 权限；检查内核版本 >= 5.4；`uname -r` 查看 |
| 所有命令都无告警输出 | 检查 `dangerous_commands.yaml` 是否在正确路径；查看 Agent 日志中 `Detection rules loaded` 行确认规则加载成功 |
| 规则加载 0 条 | 检查 YAML 格式是否正确；确认 `enabled: true` |
| 管道命令未触发告警 | 管道中每个命令是独立 execve 事件；eBPF 捕获的 args 是单个命令的参数，不包含管道符后的部分 |
| bash 内建命令未触发 | `history`、`export`、`unset` 等是 bash 内建命令，不产生 execve 事件；需要 `bash -c "..."` 包装 |
| `contains` 类型误报 | 当进程名包含关键字时会触发（如进程名含 "nmap" 的自定义程序） |

---

## 规则配置说明

规则文件路径：`config/dangerous_commands.yaml`（相对于 driver 二进制所在目录）

### 添加自定义规则

```yaml
  - id: "DC013"
    name: "自定义规则名称"
    description: "规则描述"
    severity: "high"           # critical / high / medium / low
    enabled: true
    match:
      type: "regex"            # regex / contains / prefix
      patterns:
        - "your_pattern_here"
```

修改后重启 Agent 生效。

---

## 测试完成后

1. 在 Terminal A 按 `Ctrl+C` 停止 Agent
2. 清理测试残留文件：`rm -f /tmp/dc00*`
3. 将测试结果填入上方记录表
