# 高危命令检测 — 手动测试指南

## 概述

本文档描述如何手动验证 ebpf_base_detector 插件的高危命令检测功能（DataType 6003）。

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
| `business_plugins/ebpf_base_detector/config/dangerous_commands.yaml` | 检测规则配置（12 条规则） |
| `business_plugins/ebpf_base_detector/detector/detector.go` | 规则匹配引擎 |
| `business_plugins/ebpf_base_detector/main.go` | 事件处理与告警生成 |

## 编译部署与启动

```bash
# 1. 编译译并部署
cd /home/work/goProject/src/company/agent
make build
make deploy

# 2. 启动 Agent（Terminal A）
# 检测事件输出到 stderr，Agent 运行日志输出到 /opt/cloudsec/logs/agent.log
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test
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

### DC004: 内核模块操作（high）

```bash
# 测试 1：insmod（使用不存在的模块，会报错但触发检测）
insmod /tmp/nonexistent.ko 2>/dev/null; true
```

**预期告警**：

```
rule_id=DC009  rule_name=内核模块操作  severity=high
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
| 4 | DC009 | 内核模块操作 | high | `insmod /tmp/test.ko` | 告警 | | |

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

规则文件路径：`config/dangerous_commands.yaml`（相对于 ebpf_base_detector 二进制所在目录）

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
