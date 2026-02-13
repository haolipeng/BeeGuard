# 恶意请求检测 — 手动测试指南

## 概述

本文档描述如何手动验证 ebpf_base_detector 插件的恶意请求检测功能（DataType 6008）。

**检测原理**：通过 eBPF Hook `raw_tracepoint/sys_exit` 捕获 `connect`（出站连接）和 `recvfrom/recvmsg`（DNS 响应）事件，将目标 IP、端口、域名与 `malicious_request_rules.yaml` 中的威胁情报指标进行匹配。匹配成功时产生告警。

**支持的指标类型**：

| 指标类型 | 匹配对象 | 触发事件 |
|----------|----------|----------|
| `ip` | 目标 IP 地址 | connect 出站连接 |
| `port` | 目标端口 | connect 出站连接 |
| `ip_port` | IP:端口 复合 | connect 出站连接 |
| `domain` | DNS 查询域名 | DNS 响应（recvfrom/recvmsg） |

**支持的威胁类型**：

| 威胁类型 | 说明 |
|----------|------|
| `mining` | 加密货币挖矿 |
| `c2` | C2（Command & Control）通信 |
| `phishing` | 钓鱼网站 |
| `data_leakage` | 数据泄露 |

**关键源文件**：

| 文件 | 说明 |
|------|------|
| `business_plugins/ebpf_base_detector/config/malicious_request_rules.yaml` | 恶意请求规则配置（7 条规则） |
| `business_plugins/ebpf_base_detector/malicious_request_loader.go` | 规则加载 |
| `business_plugins/ebpf_base_detector/malicious_request_detector.go` | 匹配引擎 |
| `business_plugins/ebpf_base_detector/malicious_request_types.go` | 类型定义 |
| `business_plugins/ebpf_base_detector/main.go` | 事件处理与告警生成 |

---

## 环境要求

| 项目 | 要求 |
|------|------|
| 内核版本 | >= 5.x |
| BTF 支持 | `/sys/kernel/btf/vmlinux` 存在 |
| 编译依赖 | clang、llvm、libbpf-dev、linux-headers |
| 运行权限 | root |
| 测试工具 | `curl`、`nc`（netcat）、`dig` 或 `nslookup` |

---

## 编译与启动

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

**可选**：输出到文件以便后续分析：

```bash
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test

# 另一终端实时查看
tail -f /tmp/malicious_request_detection.json
```

---

## 测试用例

在另一个终端（Terminal B）中执行以下命令。

> **安全提示**：以下测试使用恶意请求规则中配置的示例地址和域名。这些地址可能不可达或已被接管，测试目的是验证检测机制是否正确触发告警，不依赖连接是否成功。

---

### 测试 1: IP 地址匹配 — 已知矿池 IP（IOC001）

**规则**：检测连接到已知加密货币矿池 IP 地址。

```bash
# 连接到已知矿池 IP（连接会失败，但 eBPF 仍会捕获 connect 事件）
# 注意：仅捕获 connect 成功（retval == 0）的事件
# 使用 nc 尝试连接（超时 2 秒）
nc -w 2 94.23.23.52 80 2>/dev/null; true

# 或使用 curl（超时 2 秒）
curl --connect-timeout 2 http://94.23.23.52/ 2>/dev/null; true

# 其他矿池 IP
nc -w 2 104.140.201.18 80 2>/dev/null; true
nc -w 2 5.196.23.240 80 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on connect
    rule_id=IOC001  rule_name=已知矿池IP  threat_type=mining
    matched_value=94.23.23.52  pid=...  comm=nc
```

> **说明**：当前实现仅捕获 `connect` 返回值为 0（成功）的事件。如果目标 IP 不可达，connect 返回错误码，则不会触发恶意请求匹配。可通过临时添加本机 IP 到规则来验证功能。

---

### 测试 2: 端口匹配 — 常见矿池端口（IOC002）

**规则**：检测连接到常见加密货币挖矿端口（3333、4444、5555、7777 等）。

**方法 A — 使用本地监听模拟**：

```bash
# Terminal C：在本地启动监听（模拟矿池端口）
nc -lvp 3333 &

# Terminal B：连接到本地矿池端口（connect 会成功）
nc -w 1 127.0.0.1 3333 <<< "test"; true

# 清理
kill %1 2>/dev/null
```

**方法 B — 直接连接**：

```bash
# 连接到矿池端口（可能不可达）
nc -w 2 127.0.0.1 4444 2>/dev/null; true
nc -w 2 127.0.0.1 5555 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on connect
    rule_id=IOC002  rule_name=常见矿池端口  threat_type=mining
    matched_value=3333  pid=...  comm=nc
```

> **推荐方法 A**：本地监听确保 connect 成功，能可靠触发检测。

---

### 测试 3: 域名匹配 — 已知矿池域名（IOC003）

**规则**：检测 DNS 解析已知矿池域名。

```bash
# 使用 dig 查询已知矿池域名
dig pool.minexmr.com 2>/dev/null; true

# 使用 nslookup
nslookup xmr.pool.minergate.com 2>/dev/null; true

# 使用 curl 触发 DNS 解析
curl --connect-timeout 2 http://pool.minexmr.com/ 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on DNS
    rule_id=IOC003  rule_name=已知矿池域名  threat_type=mining
    matched_value=pool.minexmr.com  pid=...  comm=dig
```

> **说明**：DNS 检测依赖 eBPF 捕获 `recvfrom`/`recvmsg` 系统调用中的 DNS 响应包。检测条件：(1) UDP 协议，(2) 对端端口为 53 或 5353，(3) DNS 响应包（QR=1）。部分系统使用 `systemd-resolved` 代理 DNS 查询，此时 eBPF 捕获的进程可能是 `systemd-resolve` 而非 `dig`。

---

### 测试 4: C2 域名匹配（IOC004）

**规则**：检测 DNS 解析已知 C2 服务器域名。

```bash
# 查询 C2 域名（这些是示例域名，通常不可解析）
dig test.cobalt-strike.example.com 2>/dev/null; true
dig beacon.darkside.example.com 2>/dev/null; true

# 使用 nslookup
nslookup cmd.cobalt-strike.example.com 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on DNS
    rule_id=IOC004  rule_name=已知C2域名  threat_type=c2  severity=critical
    matched_value=test.cobalt-strike.example.com  pid=...  comm=dig
```

> **说明**：域名规则支持通配符匹配，`*.cobalt-strike.example.com` 会匹配所有子域名。

---

### 测试 5: C2 端点匹配 — IP:Port 复合（IOC005）

**规则**：检测连接到已知 C2 服务器的特定 IP:端口组合。

```bash
# 连接到已知 C2 端点
nc -w 2 185.141.27.100 443 2>/dev/null; true
nc -w 2 45.33.32.156 8443 2>/dev/null; true

# 使用 curl
curl --connect-timeout 2 https://185.141.27.100:443/ 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on connect
    rule_id=IOC005  rule_name=已知C2端点  threat_type=c2  severity=critical
    matched_value=185.141.27.100:443  pid=...  comm=nc
```

> **说明**：`ip_port` 类型要求 IP 和端口同时匹配。仅连接到匹配的 IP 但不同端口，不会触发此规则（但可能触发 IP 类型规则，如果存在的话）。

---

### 测试 6: 钓鱼域名匹配（IOC006）

**规则**：检测 DNS 解析已知钓鱼网站域名。

```bash
# 查询钓鱼域名
dig login.phishing-example.com 2>/dev/null; true
dig login-secure.example.net 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on DNS
    rule_id=IOC006  rule_name=已知钓鱼域名  threat_type=phishing
    matched_value=login.phishing-example.com  pid=...  comm=dig
```

---

### 测试 7: 数据泄露目标 IP（IOC007）

**规则**：检测连接到已知数据泄露目标服务器。

```bash
# 连接到已知数据泄露目标 IP
nc -w 2 198.51.100.1 443 2>/dev/null; true
nc -w 2 203.0.113.50 80 2>/dev/null; true
```

**预期告警**：

```
WARN  Malicious request detected on connect
    rule_id=IOC007  rule_name=已知数据泄露目标IP  threat_type=data_leakage  severity=critical
    matched_value=198.51.100.1  pid=...  comm=nc
```

---

### 测试 8: 反面用例 — 正常流量不应告警

```bash
# 正常 DNS 查询
dig www.baidu.com

# 正常 HTTP 请求
curl -s https://www.baidu.com > /dev/null

# 连接到正常端口
nc -w 1 127.0.0.1 22 2>/dev/null; true
```

**预期**：Terminal A **不应** 输出恶意请求相关告警。

---

## 本地模拟测试（推荐）

由于恶意请求规则中的 IP 和域名可能不可达（connect 不成功则不触发），推荐通过修改规则文件使用本地地址进行可靠测试。

### 步骤

1. 备份原规则文件：

```bash
cp build/config/malicious_request_rules.yaml build/config/malicious_request_rules.yaml.bak
```

2. 添加测试规则（使用本机地址）：

```yaml
  - id: "IOC_TEST"
    name: "本地测试规则"
    description: "用于测试的本地恶意请求规则"
    threat_type: "c2"
    indicator_type: "ip"
    severity: "high"
    enabled: true
    indicators:
      - "127.0.0.1"
```

3. 重启 Agent 后，任何连接到 `127.0.0.1` 的操作都会触发告警。

4. 测试完成后恢复原规则文件：

```bash
mv build/config/malicious_request_rules.yaml.bak build/config/malicious_request_rules.yaml
```

---

## 验证告警字段

### Connect 事件恶意请求告警字段

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `detection_type` | 检测类型 | malicious_request |
| `event_type` | 事件类型 | connect |
| `rule_id` | 规则 ID | IOC001 |
| `rule_name` | 规则名称 | 已知矿池IP |
| `threat_type` | 威胁类型 | mining |
| `severity` | 严重程度 | high |
| `indicator_type` | 指标类型 | ip |
| `matched_value` | 匹配到的指标值 | 94.23.23.52 |
| `pid` | 进程 ID | 12345 |
| `comm` | 进程名 | nc |
| `exe_path` | 可执行文件路径 | /usr/bin/nc |
| `remote_ip` | 目标 IP | 94.23.23.52 |
| `remote_port` | 目标端口 | 80 |
| `protocol` | 协议 | tcp |

### DNS 事件恶意请求告警字段

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `detection_type` | 检测类型 | malicious_request |
| `event_type` | 事件类型 | dns |
| `rule_id` | 规则 ID | IOC003 |
| `rule_name` | 规则名称 | 已知矿池域名 |
| `threat_type` | 威胁类型 | mining |
| `indicator_type` | 指标类型 | domain |
| `matched_value` | 匹配到的域名 | pool.minexmr.com |
| `domain` | 查询域名 | pool.minexmr.com |
| `query_type` | 查询类型 | A |
| `dns_server_ip` | DNS 服务器 IP | 127.0.0.53 |
| `pid` | 进程 ID | 12345 |
| `comm` | 进程名 | dig |

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
| Agent 启动后 `Malicious request rules loaded` 显示 0 条规则 | 检查 `malicious_request_rules.yaml` 文件路径和 YAML 格式；确认规则 `enabled: true` |
| IP 类型规则不触发告警 | 当前仅捕获 `connect` 成功（retval == 0）的事件；目标 IP 不可达时不会触发；使用本地模拟测试验证 |
| 域名规则不触发告警 | DNS 检测依赖 UDP 端口 53/5353 的 recvfrom/recvmsg；检查系统是否使用 `systemd-resolved` 代理（此时 DNS 进程可能不同）|
| 通配符域名不匹配 | 确认规则格式为 `*.example.com`，匹配所有子域名；精确域名不需要通配符 |
| 端口规则误报 | `port` 类型匹配所有目标端口为指定值的连接，不区分目标 IP；如需精确匹配使用 `ip_port` 类型 |
| DNS 事件中域名显示乱码 | DNS 包解析状态机问题；检查 eBPF 代码中 `query_dns_record` 函数；确认响应包格式正确 |
| 同一请求触发多条规则 | 正常现象；例如连接到矿池 IP 的矿池端口可能同时触发 IOC001（IP）和 IOC002（端口） |
| `systemd-resolve` 代替实际进程 | 系统使用本地 DNS 代理时，eBPF 捕获的是代理进程而非发起查询的进程；这是已知局限 |

---

## 规则配置说明

规则文件路径：`config/malicious_request_rules.yaml`（相对于 ebpf_base_detector 二进制所在目录）

### 添加自定义恶意请求规则

```yaml
  - id: "IOC_CUSTOM"
    name: "自定义威胁情报"
    description: "规则描述"
    threat_type: "c2"              # mining / c2 / phishing / data_leakage
    indicator_type: "ip"           # ip / port / domain / ip_port
    severity: "high"               # critical / high / medium / low
    enabled: true
    indicators:
      - "1.2.3.4"                  # IP 地址
      - "5.6.7.8"
```

### 域名通配符

```yaml
    indicators:
      - "evil.com"                 # 精确匹配
      - "*.evil.com"               # 匹配所有子域名（如 a.evil.com、b.c.evil.com）
```

### IP:端口复合

```yaml
    indicator_type: "ip_port"
    indicators:
      - "1.2.3.4:443"             # 仅匹配该 IP 的 443 端口
      - "5.6.7.8:8080"
```

修改后重启 Agent 生效。

---

## 已知局限

| 局限 | 原因 | 后续方案 |
|------|------|----------|
| 仅检测 connect 成功的事件 | eBPF 过滤 retval != 0 的 connect | 可选支持失败的连接尝试 |
| 仅支持 IPv4 | eBPF 代码仅处理 AF_INET (family=2) | 增加 AF_INET6 支持 |
| DNS 进程归属不准确 | systemd-resolved 代理 DNS 查询 | 增加 DNS 请求与进程的关联追踪 |
| 不支持 DNS over HTTPS/TLS | 仅监控 UDP 53/5353 端口 | 需要 TLS 拦截或代理方案 |
| 域名解析仅支持前 64 字节 | BPF 状态机循环上限 | 扩大解析范围 |

---

## 测试完成后

1. 在 Terminal A 按 `Ctrl+C` 停止 Agent
2. 如使用本地模拟测试，恢复原规则文件
3. 清理本地监听进程：`killall nc 2>/dev/null`
4. 将测试结果填入上方记录表
