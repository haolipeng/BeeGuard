# 恶意请求检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的恶意请求检测功能（DataType 6008）：eBPF 通过 `raw_tracepoint/sys_exit` Hook 捕获 `connect`（出站连接）和 `recvfrom/recvmsg`（DNS 响应）事件，用户态将目标 IP、端口、域名与 `malicious_request_rules.yaml` 中的威胁情报指标进行匹配，匹配成功时产生告警。本文档选取 5 条代表性规则进行验证，覆盖 port、domain、ip_port 三种指标类型和 critical、high、medium 三种严重程度。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | 编译环境 | `go version` | Go 已安装 |
| 6 | nc 工具 | `which nc` | 路径存在 |
| 7 | dig 工具 | `which dig` | 路径存在 |
| 8 | DNS 解析可用 | `dig +short example.com` | 返回 IP 地址（非空） |

- 条件 1-7 不满足时测试无法进行
- 条件 8 不满足时，**用例 2/3/5（domain 类型）无法通过外网方式测试**，需改用"无外网环境替代方案"（见附录 A）

---

## Step 1：编译部署

```bash
cd /home/work/goProject/src/company/agent
make build
make deploy
```

**验证**：执行 `ls -la /opt/cloudsec/bin/agent /opt/cloudsec/plugins/ebpf_base_detector/ebpf_base_detector`，两个文件都存在即成功。

---

## Step 2：启动 Agent

清空之前的测试输出并启动 Agent：

```bash
cd /opt/cloudsec
rm -f /tmp/ebpf_test.log
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/tmp/ebpf_test.log -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**看到以下日志行：

```
INFO  Malicious request rules loaded  count=7  source=config/malicious_request_rules.yaml
```

**判定规则**：
- `count=7` → 启动成功，7 条规则全部加载，进入 Step 3
- `count=0` 或该行未出现 → 启动失败，检查 `malicious_request_rules.yaml` 是否在 `/opt/cloudsec/plugins/ebpf_base_detector/config/` 目录下
- `failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3

### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stderr) | 操作日志（启动、错误等），用于确认启动状态和规则加载 |
| `/opt/cloudsec/logs/ebpf_base_detector.log` | 操作日志持久化文件 |
| `/tmp/ebpf_test.log` | **检测结果输出文件**，JSON 格式，每行一条记录，**主要验证位置** |

### 搜索技巧

检测结果以 JSON 格式写入 `/tmp/ebpf_test.log`，可使用以下方式查询：

```bash
# 搜索恶意请求告警（按 data_type 过滤）
grep '"data_type":6008' /tmp/ebpf_test.log

# 按规则 ID 精确搜索
grep '"rule_id":"IOC001"' /tmp/ebpf_test.log

# 按事件类型搜索
grep '"event_type":"dns"' /tmp/ebpf_test.log

# 使用 jq 格式化输出（需安装 jq）
cat /tmp/ebpf_test.log | jq 'select(.data_type==6008)'

# 实时监控新告警
tail -f /tmp/ebpf_test.log
```

---

## Step 3：执行测试用例

打开 **Terminal B**，逐条执行以下测试命令。每执行一条后，检查 `/tmp/ebpf_test.log` 是否出现对应告警。

### 告警日志格式

每条告警在 `/tmp/ebpf_test.log` 中以一行 JSON 输出：

```json
{"timestamp":1234567890,"data_type":6008,"event_type":"connect|dns","rule_id":"ID","rule_name":"名称","severity":"级别","threat_type":"威胁类型","indicator_type":"指标类型","matched_value":"匹配值","pid":"PID","comm":"进程名","exe_path":"路径","remote_ip":"IP","remote_port":"端口","domain":"域名"}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. `/tmp/ebpf_test.log` 中出现 `"data_type":6008` 的 JSON 行
2. `"rule_id"` 与测试用例的规则 ID 一致
3. `"indicator_type"` 与预期一致
4. `"matched_value"` 包含预期的匹配值

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 `/tmp/ebpf_test.log` 无任何 `"data_type":6008` 的记录
- `"rule_id"` 与预期不一致

---

### 用例 1：规则 IOC001 — 常见矿池端口（medium）

**指标类型**：port，匹配目标端口 3333

**测试命令**（Terminal B）：

```bash
# 启动本地监听（端口类型需要 connect 成功，retval==0）
nc -lvp 3333 &>/dev/null &
nc -w 1 127.0.0.1 3333 <<< "test" 2>/dev/null; true
kill %1 2>/dev/null
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6008,"event_type":"connect","rule_id":"IOC001","rule_name":"常见矿池端口","severity":"medium","threat_type":"mining","indicator_type":"port","matched_value":"3333","comm":"nc",...}
```

**验证命令**：

```bash
grep '"rule_id":"IOC001"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"rule_id":"IOC001"`，`"indicator_type":"port"`，`"matched_value":"3333"`。

> 说明：`port` 类型匹配所有目标端口为指定值的连接，不区分目标 IP。connect 必须成功（retval==0）才会触发，因此需要先启动本地监听。

---

### 用例 2：规则 IOC002 — 已知矿池域名（high）

**指标类型**：domain，匹配 `minersns.com` 等矿池域名

**测试命令**（Terminal B）：

```bash
dig minersns.com 2>/dev/null; true
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6008,"event_type":"dns","rule_id":"IOC002","rule_name":"已知矿池域名","severity":"high","threat_type":"mining","indicator_type":"domain","matched_value":"minersns.com","comm":"dig",...}
```

**验证命令**：

```bash
grep '"rule_id":"IOC002"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"rule_id":"IOC002"`，`"indicator_type":"domain"`，`"matched_value"` 包含 `minersns.com`。

> 说明：DNS 检测依赖 eBPF 捕获 `recvfrom`/`recvmsg` 中的 DNS 响应包（UDP 端口 53/5353）。部分系统使用 `systemd-resolved` 代理 DNS 查询，此时捕获的进程可能是 `systemd-resolve` 而非 `dig`。

---

### 用例 3：规则 IOC003 — 已知C2域名（critical）

**指标类型**：domain，匹配 `*.cobalt-strike.example.com` 等 C2 域名

**测试命令**（Terminal B）：

```bash
dig test.cobalt-strike.example.com 2>/dev/null; true
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6008,"event_type":"dns","rule_id":"IOC003","rule_name":"已知C2域名","severity":"critical","threat_type":"c2","indicator_type":"domain","matched_value":"test.cobalt-strike.example.com","comm":"dig",...}
```

**验证命令**：

```bash
grep '"rule_id":"IOC003"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"rule_id":"IOC003"`，`"indicator_type":"domain"`，`"matched_value"` 包含 `cobalt-strike.example.com`。

> 说明：域名规则支持通配符匹配，`*.cobalt-strike.example.com` 会匹配所有子域名。示例域名通常不可解析，但 DNS 查询请求本身会被捕获。

---

### 用例 4：规则 IOC004 — 已知C2端点（critical）

**指标类型**：ip_port，匹配特定 IP:端口组合

**测试命令**（Terminal B）：

```bash
# ip_port 类型需要 connect 成功，建议使用本地监听模拟
# 如果目标不可达，改用本地监听方式：
# nc -lvp 443 &>/dev/null &
# nc -w 2 127.0.0.1 443 2>/dev/null; true
# kill %1 2>/dev/null

nc -w 2 185.141.27.100 443 2>/dev/null; true
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6008,"event_type":"connect","rule_id":"IOC004","rule_name":"已知C2端点","severity":"critical","threat_type":"c2","indicator_type":"ip_port","matched_value":"185.141.27.100:443","comm":"nc",...}
```

**验证命令**：

```bash
grep '"rule_id":"IOC004"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"rule_id":"IOC004"`，`"indicator_type":"ip_port"`，`"matched_value"` 包含 `185.141.27.100:443`。

> 说明：`ip_port` 类型要求 IP 和端口同时匹配。仅连接到匹配的 IP 但不同端口，不会触发此规则。connect 必须成功（retval==0），如果目标不可达，需使用本地监听方式测试。

---

### 用例 5：规则 IOC005 — 已知钓鱼域名（high）

**指标类型**：domain，匹配 `login.phishing-example.com` 等钓鱼域名

**测试命令**（Terminal B）：

```bash
dig login.phishing-example.com 2>/dev/null; true
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6008,"event_type":"dns","rule_id":"IOC005","rule_name":"已知钓鱼域名","severity":"high","threat_type":"phishing","indicator_type":"domain","matched_value":"login.phishing-example.com","comm":"dig",...}
```

**验证命令**：

```bash
grep '"rule_id":"IOC005"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"rule_id":"IOC005"`，`"indicator_type":"domain"`，`"matched_value"` 包含 `login.phishing-example.com`。

> 说明：DNS 检测依赖 `recvfrom`/`recvmsg` 捕获 DNS 响应，与用例 2、3 的检测机制相同。

---

## Step 4：记录测试结果

| # | 规则 ID | 规则名称 | 严重程度 | 测试命令 | 预期 | 实际 | PASS/FAIL |
|---|---------|----------|----------|----------|------|------|-----------|
| 1 | IOC001 | 常见矿池端口 | medium | `nc -w 1 127.0.0.1 3333` | 告警 | | |
| 2 | IOC002 | 已知矿池域名 | high | `dig minersns.com` | 告警 | | |
| 3 | IOC003 | 已知C2域名 | critical | `dig test.cobalt-strike.example.com` | 告警 | | |
| 4 | IOC004 | 已知C2端点 | critical | `nc -w 2 185.141.27.100 443` | 告警 | | |
| 5 | IOC005 | 已知钓鱼域名 | high | `dig login.phishing-example.com` | 告警 | | |

---

## Step 5：清理与停止

```bash
# 1. Terminal A：按 Ctrl+C 停止 Agent

# 2. Terminal B：清理本地监听进程
killall nc 2>/dev/null

# 3. 清理输出文件（可选）
rm -f /tmp/ebpf_test.log
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动报 `failed to load eBPF` | 内核不支持或无 root 权限 | 1) `whoami` 确认 root；2) `uname -r` 确认 >= 5.4；3) `ls /sys/kernel/btf/vmlinux` 确认 BTF |
| `/tmp/ebpf_test.log` 未生成 | Agent 未成功启动或路径无写权限 | 1) 确认 Terminal A 中出现 `detection results will be written to: /tmp/ebpf_test.log`；2) `ls -la /tmp/ebpf_test.log` 确认文件存在 |
| 规则加载 count=0 | 配置文件缺失或格式错误 | 1) `ls /opt/cloudsec/plugins/ebpf_base_detector/config/malicious_request_rules.yaml` 确认文件存在；2) 用 `python3 -c "import yaml; yaml.safe_load(open('...'))"` 检查 YAML 语法；3) 确认规则 `enabled: true` |
| 端口规则（IOC001）不触发 | connect 未成功（retval != 0） | 1) 确认本地监听已启动：`ss -tlnp \| grep 3333`；2) connect 必须返回 0 才触发，目标不可达时不会告警 |
| 域名规则（IOC002/003/005）不触发 | DNS 响应未被捕获 | 1) 先确认 DNS 可用：`dig +short example.com`，无输出则 DNS 不可用，改用附录 A 方案；2) DNS 检测依赖 UDP 端口 53/5353 的 recvfrom/recvmsg；3) 检查系统是否使用 `systemd-resolved` 代理：`systemctl status systemd-resolved`；4) 尝试 `nslookup` 替代 `dig` |
| ip_port 规则（IOC004）不触发 | 目标不可达或端口不匹配 | 1) `ip_port` 要求 IP 和端口同时匹配；2) connect 必须成功；3) 如目标不可达，使用本地监听方式测试 |
| 同一请求触发多条规则 | 命令同时匹配多条规则 | 正常现象，例如连接到矿池 IP 的矿池端口可能同时触发端口和 IP 规则 |
| 告警延迟超过 5 秒 | standalone 刷新间隔较长 | 检查配置中 `flush_interval`（默认 1 秒）；eBPF 事件本身无延迟，延迟来自用户态轮询 |
| 所有域名规则均不触发 | DNS 不可用（无外网） | `dig +short example.com` 无输出，确认 DNS 不可用后改用附录 A 的本地 DNS 方案 |

---

## 附录 A：无外网环境替代方案

当测试机无法访问外网（DNS 解析失败）时，domain 类型规则（IOC003/004/006）无法通过常规 `dig` 触发。以下方案通过本地 DNS 服务器模拟 DNS 响应，使 eBPF 能捕获到包含目标域名的 DNS 应答包。

### A.1 安装 dnsmasq

```bash
# Ubuntu/Debian（dnsmasq 通常已预装或可离线安装）
apt install dnsmasq -y 2>/dev/null || echo "dnsmasq 不可用，改用 A.2 Python 方案"
```

### A.2 方案一：dnsmasq 本地 DNS

```bash
# 1. 停止 systemd-resolved（避免端口 53 冲突）
systemctl stop systemd-resolved 2>/dev/null

# 2. 启动 dnsmasq，所有查询返回 127.0.0.1
dnsmasq --no-daemon --listen-address=127.0.0.1 --address=/#/127.0.0.1 --log-queries &
DNSMASQ_PID=$!
sleep 1

# 3. 测试域名规则（指定 @127.0.0.1 使用本地 DNS）
# IOC003 - 已知矿池域名
dig @127.0.0.1 minersns.com 2>/dev/null; true

# IOC004 - 已知 C2 域名
dig @127.0.0.1 test.cobalt-strike.example.com 2>/dev/null; true

# IOC006 - 已知钓鱼域名
dig @127.0.0.1 login.phishing-example.com 2>/dev/null; true

# 4. 验证告警
grep '"indicator_type":"domain"' /tmp/ebpf_test.log

# 5. 清理
kill $DNSMASQ_PID 2>/dev/null
systemctl start systemd-resolved 2>/dev/null
```

### A.3 方案二：Python 简易 DNS 服务器

如果 dnsmasq 不可用，可用 Python 启动一个最小 DNS 应答服务：

```bash
# 1. 停止 systemd-resolved
systemctl stop systemd-resolved 2>/dev/null

# 2. 启动 Python DNS 应答服务（监听 UDP 53，返回 127.0.0.1）
python3 -c "
import socket, struct
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind(('127.0.0.1', 53))
print('Local DNS server running on 127.0.0.1:53')
while True:
    data, addr = sock.recvfrom(512)
    # 构造最小应答：复制查询头，设置 QR=1（应答），ANCOUNT=1
    resp = bytearray(data)
    resp[2] = 0x81; resp[3] = 0x80  # QR=1, RD=1, RA=1
    resp[6] = 0x00; resp[7] = 0x01  # ANCOUNT=1
    # 追加应答 RR：指向查询 name (0xc00c)，A 记录，TTL=60，RDATA=127.0.0.1
    resp += b'\xc0\x0c\x00\x01\x00\x01\x00\x00\x00\x3c\x00\x04\x7f\x00\x00\x01'
    sock.sendto(bytes(resp), addr)
" &
DNS_PID=$!
sleep 1

# 3. 执行域名测试（同 A.2 步骤 3）
dig @127.0.0.1 minersns.com 2>/dev/null; true
dig @127.0.0.1 test.cobalt-strike.example.com 2>/dev/null; true
dig @127.0.0.1 login.phishing-example.com 2>/dev/null; true

# 4. 验证 + 清理
grep '"indicator_type":"domain"' /tmp/ebpf_test.log
kill $DNS_PID 2>/dev/null
systemctl start systemd-resolved 2>/dev/null
```

### A.4 port / ip_port / ip 类型规则

这三类规则不依赖 DNS，在无外网环境下本身已可通过本地监听方式测试（正文用例 1、4 已使用此方式），无需额外处理。

> **注意**：本地 DNS 方案仅用于测试环境。测试完成后务必恢复 `systemd-resolved` 服务，避免影响系统正常 DNS 解析。
