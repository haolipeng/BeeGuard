# 恶意请求检测测试报告

**测试日期**: 2026-02-14
**测试人员**: Claude
**Agent 版本**: 891d11b-dirty
**插件**: ebpf_base_detector

---

## 测试环境

- **操作系统**: Linux 6.5.0-18-generic
- **编译方式**: `make build && make deploy`
- **启动命令**: `sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/opt/cloudsec/logs/agent.log -test`
- **规则文件**: `/opt/cloudsec/plugins/ebpf_base_detector/config/malicious_request_rules.yaml`
- **规则版本**: 1.0
- **已启用规则数**: 7

---

## 测试结果汇总

| # | 规则 ID | 规则名称 | 指标类型 | 严重程度 | 测试方法 | 测试结果 | 备注 |
|---|---------|----------|----------|---------|---------|---------|------|
| 1 | IOC001 | 已知矿池IP | ip | high | nc 94.23.23.52 80 | ⚠️ 未测试 | 需外部可达IP |
| 2 | IOC002 | 常见矿池端口 | port | medium | Python连接127.0.0.1:3333/4444 | ✅ 通过 | 检测到2条告警 |
| 3 | IOC003 | 已知矿池域名 | domain | high | nslookup pool.minexmr.com | ❌ 失败 | DNS域名解析bug |
| 4 | IOC004 | 已知C2域名 | domain | critical | nslookup test.cobalt-strike.example.com | ❌ 失败 | DNS域名解析bug |
| 5 | IOC005 | 已知C2端点 | ip_port | critical | nc 185.141.27.100 443 | ⚠️ 未测试 | 需外部可达IP |
| 6 | IOC006 | 已知钓鱼域名 | domain | high | nslookup login.phishing-example.com | ❌ 失败 | DNS域名解析bug |
| 7 | IOC007 | 已知数据泄露IP | ip | critical | nc 198.51.100.1 443 | ⚠️ 未测试 | 需外部可达IP |

**通过率**: 1/4 (25%) - 仅 port 类型规则通过测试

---

## 详细测试记录

### ✅ 测试1: IOC002 - 常见矿池端口 (port: 3333)

**执行命令**:
```bash
# 启动本地监听
python3 -c "import socket; s=socket.socket(); s.bind(('127.0.0.1', 3333)); s.listen(1); ..."
# 连接测试
python3 -c "import socket; s=socket.socket(); s.connect(('127.0.0.1', 3333)); s.close()"
```

**告警记录** (`agent.log`):
```json
{
    "timestamp": 1771025105,
    "data_type": 6008,
    "rule_id": "IOC002",
    "rule_name": "常见矿池端口",
    "severity": "medium",
    "pid": "2678032",
    "uid": "0",
    "exe_path": "/usr/bin/python3.10",
    "all_fields": {
        "detection_type": "malicious_request",
        "event_type": "connect",
        "indicator_type": "port",
        "matched_value": "3333",
        "remote_ip": "127.0.0.1",
        "remote_port": "3333",
        "protocol": "tcp",
        "threat_type": "mining"
    }
}
```

**插件日志** (`ebpf_base_detector.stderr`):
```
2026-02-14T07:24:08.143+0800	INFO	Connect event	{"pid": 2677772, "comm": "python3", "remote_ip": "127.0.0.1", "remote_port": "3333", "protocol": "tcp", "retval": 0}
2026-02-14T07:24:08.145+0800	WARN	Malicious request detected on connect	{"rule_id": "IOC002", "rule_name": "常见矿池端口", "threat_type": "mining", "matched_value": "3333", "pid": 2677772, "comm": "python3"}
```

**结论**: ✅ **通过** - 端口类型规则检测正常，成功匹配 3333 端口连接行为

---

### ✅ 测试2: IOC002 - 常见矿池端口 (port: 4444)

**执行命令**:
```bash
python3 -c "import socket; s=socket.socket(); s.bind(('127.0.0.1', 4444)); s.listen(1); ..."
python3 -c "import socket; s=socket.socket(); s.connect(('127.0.0.1', 4444)); s.close()"
```

**告警记录** (`agent.log`):
```json
{
    "timestamp": 1771025108,
    "data_type": 6008,
    "rule_id": "IOC002",
    "matched_value": "4444",
    "remote_port": "4444"
}
```

**结论**: ✅ **通过** - 成功检测到 4444 端口连接

---

### ❌ 测试3: IOC003 - 已知矿池域名

**执行命令**:
```bash
nslookup pool.minexmr.com 8.8.8.8
nslookup xmr.pool.minergate.com 8.8.8.8
```

**插件日志**:
```
# DNS 查询事件被捕获，但域名格式错误
(DNS query executed, queries sent but not captured with correct format)
```

**问题分析**:
- DNS 事件已被 eBPF 捕获
- 但域名解析存在 bug：`hids.bpf.dev.c:876-893 process_domain_name()` 函数
- 原始 DNS 域名格式（长度前缀编码）未正确转换为点分格式
- 例如：期望 `test.cobalt-strike.example.com`，实际 `.est\rcobalt-strike\u0007example\u0003com`

**结论**: ❌ **失败** - DNS 域名解析 bug 导致无法匹配

---

### ❌ 测试4: IOC004 - 已知C2域名

**执行命令**:
```bash
nslookup test.cobalt-strike.example.com 8.8.8.8
nslookup beacon.darkside.example.com 8.8.8.8
```

**插件日志** (`ebpf_base_detector.stderr`):
```
2026-02-14T07:25:44.629+0800	INFO	DNS query event	{"pid": 2678093, "comm": "isc-net-0000", "domain": ".est\rcobalt-strike\u0007example\u0003com", "query_type": "A", "dns_server": "8.8.8.8"}
```

**结论**: ❌ **失败** - 同上，DNS 域名解析 bug

---

### ❌ 测试5: IOC006 - 已知钓鱼域名

**执行命令**:
```bash
nslookup login.phishing-example.com 8.8.8.8
```

**插件日志**:
```
2026-02-14T07:25:46.930+0800	INFO	DNS query event	{"pid": 2678101, "comm": "isc-net-0000", "domain": ".ogin\u0010phishing-example\u0003com", "query_type": "A", "dns_server": "8.8.8.8"}
```

**结论**: ❌ **失败** - DNS 域名解析 bug（第一个字符被 '.' 替换）

---

## 问题总结

### 🐛 Bug #1: DNS 域名解析错误

**位置**: `business_plugins/ebpf_base_detector/ebpf/bpf/hids.bpf.dev.c:876-893`

**问题描述**:
`process_domain_name()` 函数在解析 DNS 长度前缀编码域名时存在缺陷：
```c
static __noinline int process_domain_name(char *data, char *name, int *flag, int i)
{
    char rc = *(data + 12 + i);
    int v = *flag;
    if (0 == rc) return 0;
    if (v == 0) {
        v = rc;
        name[i - 1] = 46;  // ← BUG: 第一次写入时 i=1，写到 name[0]，跳过第一个字符
    } else {
        name[i - 1] = rc;
        v = v - 1;
    }
    *flag = v;
    return 1;
}
```

**影响**:
- 所有 domain 类型规则无法正常工作
- 域名第一个字符被替换为 '.'
- 匹配引擎无法匹配正确域名

**影响规则**:
- IOC003: 已知矿池域名
- IOC004: 已知C2域名
- IOC006: 已知钓鱼域名

**修复建议**: 需要重写 DNS 域名解析逻辑，正确处理长度前缀编码

---

### ⚠️ 限制: Connect 事件仅捕获成功连接

**位置**: `hids.bpf.dev.c:958-960`

**代码**:
```c
if (syscall_nr == 42) {  // connect
    if (retval != 0)     // 仅采集成功的 connect (retval == 0)
        return 0;
```

**影响**:
- IP/ip_port 类型规则需要目标可达
- 无法测试外部 IP 地址（94.23.23.52, 185.141.27.100 等）
- 测试需要本地监听服务或使用真实恶意 IP

**建议**: 测试文档已说明，符合预期设计

---

## 功能验证

### ✅ 已验证功能

1. **eBPF 程序加载**: 正常
   - `tp_proc_exec` 和 `tp_sys_exit` 程序已挂载
   - Perf buffer 事件读取正常

2. **规则加载**: 正常
   - 7条规则全部加载成功
   - 规则索引构建正确 (ipIndex, domainIndex, portIndex, ipPortIndex)

3. **Connect 事件捕获**: 正常
   - 成功捕获 TCP connect 事件
   - retval == 0 过滤正常工作

4. **Port 类型规则匹配**: ✅ **正常**
   - 端口匹配引擎工作正常
   - 告警生成和发送正常
   - Agent 日志输出正常

5. **告警字段完整性**: 正常
   - DataType: 6008
   - detection_type: malicious_request
   - event_type: connect/dns
   - 所有必需字段齐全

### ❌ 未验证功能

1. **Domain 类型规则**: DNS 域名解析 bug 阻塞
2. **IP 类型规则**: 需外部可达 IP，环境限制未测试
3. **IP_Port 类型规则**: 需外部可达 IP:Port，环境限制未测试

---

## 测试结论

### 核心功能状态

| 功能模块 | 状态 | 说明 |
|---------|------|------|
| eBPF 事件捕获 | ✅ 正常 | Connect/DNS 事件正常捕获 |
| 规则加载 | ✅ 正常 | 7条规则全部加载 |
| Port 类型匹配 | ✅ 正常 | 检测成功，告警正常 |
| Domain 类型匹配 | ❌ 异常 | DNS 域名解析 bug |
| IP 类型匹配 | ⚠️ 未测试 | 需外部可达 IP |
| IP_Port 类型匹配 | ⚠️ 未测试 | 需外部可达 IP:Port |

### 严重性评估

- **Critical Bug**: DNS 域名解析错误
  - 影响 3/7 规则（43%）
  - 阻塞所有 domain 类型威胁检测
  - **优先级: P0**

### 建议

1. **立即修复**: DNS 域名解析 bug (`process_domain_name` 函数)
2. **扩展测试**: 在有外网访问的环境测试 IP/ip_port 类型规则
3. **增强测试**: 添加自动化测试脚本验证所有规则类型
4. **文档更新**: 将 DNS bug 标记为已知问题，提供临时解决方案

---

## 附录: 测试命令

### 端口检测测试 (可用)
```bash
# IOC002: 端口 3333/4444/5555/7777/8333/14444/14433
python3 << EOF &
import socket, time
s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(('127.0.0.1', 3333))
s.listen(1)
time.sleep(5)
s.close()
EOF
sleep 1
python3 -c "import socket; s=socket.socket(); s.connect(('127.0.0.1', 3333)); s.close()"
```

### DNS 检测测试 (目前不可用)
```bash
# IOC003/IOC004/IOC006: 域名检测
nslookup pool.minexmr.com 8.8.8.8
nslookup test.cobalt-strike.example.com 8.8.8.8
```

### 查看告警
```bash
# Agent 主日志
tail -f /opt/cloudsec/logs/agent.log | python3 -m json.tool

# 插件调试日志
tail -f /opt/cloudsec/plugins/ebpf_base_detector/ebpf_base_detector.stderr
```

---

**报告生成时间**: 2026-02-14 07:28:00
