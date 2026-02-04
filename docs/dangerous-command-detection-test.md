# 高危命令检测功能测试文档

> 本文档记录使用 Standalone 模式测试高危命令检测功能的完整流程

---

## 测试环境

| 项目 | 说明 |
|------|------|
| 操作系统 | Linux 6.5.0-18-generic |
| Agent 路径 | `/home/work/goProject/src/company/agent` |
| 插件目录 | `/opt/cloudsec/plugins` |
| 工作目录 | `/tmp/cloudsec-agent` |

---

## 测试流程

### 步骤 1: 编译 Agent

```bash
cd /home/work/goProject/src/company/agent
go build -o agent-bin main.go
```

### 步骤 2: 编译 Driver 插件

```bash
cd /home/work/goProject/src/company/agent/business_plugins/driver
go build -o driver .
```

### 步骤 3: 准备插件目录

```bash
# 创建插件目录
sudo mkdir -p /opt/cloudsec/plugins/driver/config

# 复制 driver 可执行文件
sudo cp business_plugins/driver/driver /opt/cloudsec/plugins/driver/

# 复制规则配置文件
sudo cp business_plugins/driver/config/dangerous_commands.yaml /opt/cloudsec/plugins/driver/config/

# 创建工作目录
sudo mkdir -p /tmp/cloudsec-agent
```

### 步骤 4: 启动 Standalone 模式

```bash
# 使用配置文件启动
sudo ./agent-bin -config=agent-standalone.yaml -test

# 或使用命令行参数启动
sudo ./agent-bin -standalone -plugins=driver -output=log -test
```

### 步骤 5: 执行测试命令

在另一个终端执行以下命令触发检测规则：

```bash
# DC001 (regex) - 危险删除操作
rm -rf /tmp/test_nonexistent_dir

# DC002 (regex) - 敏感文件访问
cat /etc/passwd

# DC006 (contains) - 可疑安全工具
which nmap

# DC009 (prefix) - 内核模块操作
modprobe --version
```

### 步骤 6: 查看检测结果

检测结果将输出到日志（默认）或 JSON 文件（配置 output=file 时）。

---

## 测试结果示例

### 日志输出格式

```
2026-02-03T18:45:30.381+0800  INFO  standalone/output.go:151  dangerous command detected
    {"rule_id": "DC001", "rule_name": "危险删除操作", "severity": "critical",
     "command": "rm -rf /tmp/test_nonexistent_dir_12345",
     "matched_pattern": "rm\\s+.*-rf\\s+/", "pid": "727422", "uid": "0"}

2026-02-03T18:45:30.381+0800  INFO  standalone/output.go:151  dangerous command detected
    {"rule_id": "DC006", "rule_name": "可疑安全工具", "severity": "medium",
     "command": "/bin/sh /usr/bin/which nmap",
     "matched_pattern": "nmap", "pid": "727423", "uid": "0"}

2026-02-03T18:45:30.381+0800  INFO  standalone/output.go:151  dangerous command detected
    {"rule_id": "DC009", "rule_name": "内核模块操作", "severity": "high",
     "command": "modprobe --version",
     "matched_pattern": "modprobe", "pid": "727424", "uid": "0"}

2026-02-03T18:45:30.381+0800  INFO  standalone/output.go:151  dangerous command detected
    {"rule_id": "DC002", "rule_name": "敏感文件访问", "severity": "high",
     "command": "cat /etc/passwd",
     "matched_pattern": "cat\\s+.*/etc/(passwd|shadow|sudoers)", "pid": "727426", "uid": "0"}
```

---

## 规则覆盖测试

已测试的规则及对应匹配类型：

| 规则 ID | 规则名称 | 匹配类型 | 测试命令 | 结果 |
|---------|----------|----------|----------|------|
| DC001 | 危险删除操作 | regex | `rm -rf /tmp/xxx` | PASS |
| DC002 | 敏感文件访问 | regex | `cat /etc/passwd` | PASS |
| DC006 | 可疑安全工具 | contains | `which nmap` | PASS |
| DC009 | 内核模块操作 | prefix | `modprobe --version` | PASS |

---

## 配置文件说明

### agent-standalone.yaml

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

### 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-config` | 配置文件路径 | `-config=agent-standalone.yaml` |
| `-standalone` | 启用 standalone 模式 | `-standalone` |
| `-output` | 输出方式 (log/file) | `-output=log` |
| `-output-path` | 输出文件路径 | `-output-path=/tmp/results.json` |
| `-plugins` | 加载的插件列表 | `-plugins=driver` |
| `-test` | 测试模式（固定 agent ID） | `-test` |

---

## 检测规则清单

当前配置了 12 条高危命令检测规则：

| ID | 名称 | 严重级别 | 匹配类型 |
|----|------|----------|----------|
| DC001 | 危险删除操作 | critical | regex |
| DC002 | 敏感文件访问 | high | regex |
| DC003 | 危险权限修改 | high | regex |
| DC004 | 下载并执行 | critical | regex |
| DC005 | 计划任务修改 | medium | regex |
| DC006 | 可疑安全工具 | medium | contains |
| DC007 | SSH密钥操作 | high | regex |
| DC008 | 历史记录清除 | medium | regex |
| DC009 | 内核模块操作 | high | prefix |
| DC010 | 防火墙规则修改 | medium | regex |
| DC011 | Base64解码执行 | high | regex |
| DC012 | 脚本语言危险执行 | high | regex |

---

## 注意事项

1. **需要 root 权限**: eBPF 程序加载需要 root 权限
2. **内核版本要求**: 需要 Linux 内核 5.4+ 且支持 BTF
3. **规则文件位置**: 规则文件需放在 `插件目录/driver/config/dangerous_commands.yaml`
4. **日志查看**: standalone 模式仅输出触发规则的命令，未触发规则的命令不会显示

---

## 常见问题

### Q: 为什么没有检测结果输出？

A: 检查以下几点：
- driver 插件是否成功加载（查看日志中的 `plugin loaded in standalone mode`）
- 规则配置文件是否正确放置
- 执行的命令是否匹配规则模式

### Q: 如何添加自定义规则？

A: 编辑 `config/dangerous_commands.yaml` 文件，按照现有规则格式添加新规则，然后重启 agent。

### Q: standalone 模式和正常模式有什么区别？

A: standalone 模式不连接 gRPC server，检测结果输出到本地日志或文件，适合本地开发测试。

---

## 测试时间

- **测试日期**: 2026-02-03
- **测试执行人**: Claude Code
