# 恶意请求检测 — 手动测试指南

## 概述

本文档描述如何手动验证 ebpf_base_detector 插件的恶意请求检测功能（DataType 6008）。

**检测原理**：通过 eBPF Hook `raw_tracepoint/sys_exit` 捕获 `connect`（出站连接）和 `recvfrom/recvmsg`（DNS 响应）事件，将目标 IP、端口、域名与 `malicious_request_rules.yaml` 中的威胁情报指标进行匹配。匹配成功时产生告警。

**指标类型**：

| 匹配类型 | 匹配对象 | 触发事件 |
|----------|----------|----------|
| `ip` | 目标 IP 地址 | connect 出站连接 |
| `port` | 目标端口 | connect 出站连接 |
| `ip_port` | IP:端口 复合 | connect 出站连接 |
| `domain` | DNS 查询域名 | DNS 响应（recvfrom/recvmsg） |

**关键源文���**：

| 文件 | 说明 |
|------|------|
| `business_plugins/ebpf_base_detector/config/malicious_request_rules.yaml` | 恶意请求规则配置（7 条规则） |
| `business_plugins/ebpf_base_detector/malicious_request_detector.go` | 匹配引擎 |
| `business_plugins/ebpf_base_detector/malicious_request_loader.go` | 规则加载 |
| `business_plugins/ebpf_base_detector/main.go` | 事件处理与告警生成 |

## 编译部署与启动

```bash
# 1. 编译并部署
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

> **注意**：当前仅捕获 `connect` 返回值为 0（成功）的事件。目标 IP 不可达时不触发。推荐使用本地监听方式测试，或临时将 `127.0.0.1` 添加到规则的 indicators 列表中。

---

### IOC001: 常见矿池端口（medium）

```bash
# 在本地启动矿池端口监听
nc -lvp 3333 &>/dev/null &
nc -w 1 127.0.0.1 3333 <<< "test" 2>/dev/null; true
kill %1 2>/dev/null
```

**预期告警**：

```
rule_id=IOC002  rule_name=常见矿池端口  threat_type=mining  indicator_type=port
```

> **注意**：`port` 类型匹配所有目标端口为指定值的连接，不区分目标 IP。如需精确匹配使用 `ip_port` 类型。

---

### IOC002: 已知矿池域名（high）

```bash
# 使用 dig 查询已知矿池域名
dig pool.minexmr.com 2>/dev/null; true

# 使用 nslookup
nslookup xmr.pool.minergate.com 2>/dev/null; true
```

**预期告警**：

```
rule_id=IOC003  rule_name=已知矿池域名  threat_type=mining  indicator_type=domain
```

> **注意**：DNS 检测依赖 eBPF 捕获 `recvfrom`/`recvmsg` 中的 DNS 响应包（UDP 端口 53/5353）。部分系统使用 `systemd-resolved` 代理 DNS 查询，此时捕获的进程可能是 `systemd-resolve` 而非 `dig`。

---

### IOC003: 已知C2域名（critical）

```bash
# 查询 C2 域名（示例域名，通常不可解析）
dig test.cobalt-strike.example.com 2>/dev/null; true
dig beacon.darkside.example.com 2>/dev/null; true
```

**预期告警**：

```
rule_id=IOC004  rule_name=已知C2域名  threat_type=c2  indicator_type=domain
```

> **注意**：域名规则支持通配符匹配，`*.cobalt-strike.example.com` 会匹配所有子域名。

---

### IOC004: 已知C2端点（critical）

```bash
# 连接到已知 C2 端点（IP:端口复合匹配）
nc -w 2 185.141.27.100 443 2>/dev/null; true
nc -w 2 45.33.32.156 8443 2>/dev/null; true
```

**预期告警**：

```
rule_id=IOC005  rule_name=已知C2端点  threat_type=c2  indicator_type=ip_port
```

> **注意**：`ip_port` 类型要求 IP 和端口同时匹配。仅连接到匹配的 IP 但不同端口，不会触发此规则。

---

### IOC005: 已知钓鱼域名（high）

```bash
# 查询钓鱼域名
dig login.phishing-example.com 2>/dev/null; true
dig login-secure.example.net 2>/dev/null; true
```

**预期告警**：

```
rule_id=IOC006  rule_name=已知钓鱼域名  threat_type=phishing  indicator_type=domain
```

---

## 验证告警字段

在 Terminal A 的输出中，确认每条告警包含以下关键字段：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `detection_type` | 检测类型 | malicious_request |
| `event_type` | 事件来源 | connect / dns |
| `rule_id` | 规则 ID | IOC001 |
| `rule_name` | 规则名称 | 已知矿池IP |
| `threat_type` | 威胁类型 | mining / c2 / phishing / data_leakage |
| `severity` | 严重程度 | critical / high / medium |
| `indicator_type` | 指标类型 | ip / port / domain / ip_port |
| `matched_value` | 匹配到的指标值 | 94.23.23.52 |
| `pid` | 进程 ID | 12345 |
| `comm` | 进程名 | nc |
| `exe_path` | 可执行文件路径 | /usr/bin/nc |
| `remote_ip` | 目标 IP（connect 事件） | 94.23.23.52 |
| `remote_port` | 目标端口（connect 事件） | 80 |
| `domain` | 查询域名（DNS 事件） | pool.minexmr.com |

---

## 测试结果记录表

| # | 规则 ID | 规则名称 | 指标类型 | 测试方法 | 预期 | 实际 | 备注 |
|---|---------|----------|----------|----------|------|------|------|
| 1 | IOC001 | 已知矿池IP | ip | `nc 94.23.23.52 80` | 告警 | | connect 需成功 |
| 2 | IOC002 | 常见矿池端口 | port | 本地监听 3333 + nc 连接 | 告警 | | 推荐本地监听 |
| 3 | IOC003 | 已知矿池域名 | domain | `dig pool.minexmr.com` | 告警 | | 依赖 DNS 响应 |
| 4 | IOC004 | 已知C2域名 | domain | `dig test.cobalt-strike.example.com` | 告警 | | 通配符匹配 |
| 5 | IOC005 | 已知C2端点 | ip_port | `nc 185.141.27.100 443` | 告警 | | IP+端口同时匹配 |
| 6 | IOC006 | 已知钓鱼域名 | domain | `dig login.phishing-example.com` | 告警 | | 通配符匹配 |
| 7 | IOC007 | 已知数据泄露IP | ip | `nc 198.51.100.1 443` | 告警 | | connect 需成功 |
| 8 | - | 正常流量 | - | `curl www.baidu.com` | 不告警 | | 反面用例 |

---

## 常见问题排查

| 问题 | 排查方法 |
|------|----------|
| Agent 启动后规则加载 0 条 | 检查 `malicious_request_rules.yaml` 文件路径和 YAML 格式；确认规则 `enabled: true` |
| IP 类型规则不触发告警 | 当前仅捕获 `connect` 成功（retval == 0）的事件；目标 IP 不可达时不会触发；使用本地监听验证 |
| 域名规则不触发告警 | DNS 检测依赖 UDP 端口 53/5353 的 recvfrom/recvmsg；检查系统是否使用 `systemd-resolved` 代理 |
| 通配符域名不匹配 | 确认规则格式为 `*.example.com`；精确域名不需要通配符 |
| 端口规则误报 | `port` 类型匹配所有目标端口为指定值的连接，不区分目标 IP；如需精确匹配使用 `ip_port` 类型 |
| DNS 事件中域名显示乱码 | DNS 包解析问题；检查 eBPF 代码中 `query_dns_record` 函数 |
| 同一请求触发多条规则 | 正常现象；例如连接到矿池 IP 的矿池端口可能同时触发 IOC001 和 IOC002 |

---

## 规则配置说明

规则文件路径：`config/malicious_request_rules.yaml`（相对于 ebpf_base_detector 二进制所在目录）

### 添加自定义规则

```yaml
  - id: "IOC_CUSTOM"
    name: "自定义威胁情报"
    description: "规则描述"
    threat_type: "c2"              # mining / c2 / phishing / data_leakage
    indicator_type: "ip"           # ip / port / domain / ip_port
    severity: "high"               # critical / high / medium / low
    enabled: true
    indicators:
      - "1.2.3.4"
```

修改后重启 Agent 生效。

---

## 测试完成后

1. 在 Terminal A 按 `Ctrl+C` 停止 Agent
2. 清理本地监听进程：`killall nc 2>/dev/null`
3. 将测试结果填入上方记录表
