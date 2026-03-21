# Codex 代码审查报告

> 审查日期: 2026-02-26
> 审查范围: Agent (客户端) + HCIDS (服务端)
> 代码生成工具: Codex

## 总览

| 模块 | Critical | High | Medium | Low | 合计 |
|------|----------|------|--------|-----|------|
| Agent (客户端) | 4 | 9 | 13 | 14 | **40** |
| HCIDS (服务端) | 5 | 6 | 11 | 8 | **30** |
| **合计** | **9** | **15** | **24** | **22** | **70** |

---

# 一、Agent (客户端) 审查报告

## Critical 级别 (4个)

### C-01: 命令注入漏洞 - baseline 插件的 CommandCheck 函数

- **文件**: `agent/business_plugins/baseline/check/rules.go:27-50`
- **描述**: `CommandCheck` 函数接收来自 YAML 配置文件（或服务端下发的 JSON 数据）的 `param[0]` 字符串，直接通过 `strings.Split(command, " ")` 分割后传入 `exec.Command`。如果配置文件来源于服务端下发（`taskData.BaselineInfo` 不为 nil，见 `analysis.go:70-72`），攻击者控制服务端后可以通过下发恶意基线规则实现任意命令执行。即使命令来自本地 YAML 文件，简单的空格分割也不能正确处理包含空格的参数。
- **修复建议**: 对命令参数进行严格白名单校验；禁止直接从不可信来源执行任意命令；或使用 `exec.Command` 时显式指定允许的命令路径。

### C-02: gRPC 通信未加密 - 中间人攻击风险

- **文件**: `agent/transport/connection.go:26-30`
- **描述**: `dialOptions` 使用 `insecure.NewCredentials()` 建立 gRPC 连接，所有 Agent 与 Server 之间的通信均为明文传输。攻击者可以进行中间人攻击，窃取安全数据（如主机信息、进程列表、基线检查结果），或者篡改服务端下发的命令（包括插件配置、任务指令）来控制 Agent。
- **修复建议**: 使用 TLS 加密通信，配置客户端证书进行双向认证。

### C-03: 服务端可通过 gRPC 命令远程关闭 Agent 且无鉴权

- **文件**: `agent/transport/transfer.go:182-186`
- **描述**: 收到 `DataType == 1060` 的 Task 后直接调用 `agent.Cancel()` 关闭整个 Agent 进程。结合 C-02 的明文通信问题，攻击者可以伪造服务端命令远程关闭目标主机上的安全 Agent，使其失去安全防护能力。
- **修复建议**: 对关键控制命令增加签名验证或 HMAC 校验机制。

### C-04: 插件签名验证被注释掉 - 可执行任意插件

- **文件**: `agent/plugin/plugin_linux.go:77-87`
- **描述**: 插件签名验证和下载逻辑被完全注释掉（`utils.CheckSignature` 和 `utils.Download`）。这意味着只要在 `PluginsDirectory` 下放置同名可执行文件，Agent 就会以 root 权限启动它。配合路径可控，攻击者可以植入恶意插件。
- **修复建议**: 实现并启用插件签名验证；至少在插件启动前进行 SHA256 校验。

---

## High 级别 (9个)

### H-01: Ring Buffer 数据静默丢弃 - 安全事件丢失

- **文件**: `agent/buffer/buffer.go:17-25`
- **描述**: `WriteEncodedRecord` 函数在 buffer 满时（offset >= 8192）静默丢弃数据，既不返回错误，也不记录日志。虽然定义了 `ErrbufferOverflow` 错误变量，但从未被使用。安全事件（如提权检测、反弹 shell 检测）如果因 buffer 满而丢失，将导致严重的安全监控盲区。

```go
if offset < len(buf) {
    buf[offset] = rec
    offset++
}
// 这里 offset >= len(buf) 时没有任何处理
```

- **修复建议**: Buffer 满时返回 `ErrbufferOverflow` 错误，并让调用方记录日志或实施背压策略。

### H-02: sync.Pool 对象复用后数据污染

- **文件**: `agent/buffer/pool.go:30-35` 与 `agent/transport/transfer.go:147-149`
- **描述**: `PutEncodedRecords` 仅重置了 `Data` 字段（`rec.Data = rec.Data[:0]`），但未清除 `DataType` 和 `Timestamp` 字段。而在 `handleSend` 中，发送完数据后立即将 record 放回 pool，但 `buffer.WriteEncodedRecord` 写入的 record 并非从 pool 获取的（plugin.go 中 `ReceiveData` 使用 `&proto.EncodedRecord{}`），导致 pool 中积累非池化对象，造成混乱。
- **修复建议**: 统一使用 pool 分配 record；回收时清除所有字段。

### H-03: 全局变量数据竞争 - agent/id.go 包级别变量

- **文件**: `agent/agent/id.go:12-42`
- **描述**: `ID`、`TestMode`、`WorkingDirectory`、`PluginsDirectory`、`LogDirectory` 等全局变量在 `init()` 函数中初始化，又在 `main.go:70-72` 被重新赋值，还可能被其他 goroutine 读取。这些变量没有任何同步保护。
- **修复建议**: 使用 `sync.Once` 或原子操作保护这些全局变量，或将它们封装到一个只读配置结构体中。

### H-04: transport/transfer.go 中 GetState 存在数据竞争

- **文件**: `agent/transport/transfer.go:25-33`
- **描述**: `GetState` 函数通过 `atomic.SwapUint64` 操作 `txCnt` 和 `rxCnt`，但 `updateTime` 变量的读写没有任何同步保护。多个 goroutine 同时调用 `GetState` 时会产生数据竞争。
- **修复建议**: 用 mutex 保护 `updateTime` 的读写，或将其改为原子操作。

### H-05: Plugin.GetState 中 updateTime 的非原子读写

- **文件**: `agent/plugin/plugin.go:56-67`
- **描述**: `GetState` 方法中 `p.updateTime` 的读写没有加锁保护，而 `rxBytes/txBytes/rxCnt/txCnt` 使用了原子操作。
- **修复建议**: 将整个方法用 `p.mu.Lock()` 保护，或将 `updateTime` 改为原子操作。

### H-06: PersistID 使用相对路径写文件

- **文件**: `agent/agent/id.go:183-193`
- **描述**: `PersistID` 函数接收 `workingDir` 参数但完全忽略它，直接使用相对路径 `"machine-id"` 写文件。文件位置取决于当前工作目录，可能写入意想不到的位置。`init()` 函数同样使用相对路径读取。
- **修复建议**: 使用 `filepath.Join(workingDir, "machine-id")` 构建绝对路径。

### H-07: 异常登录检测器 Check 方法在 RLock 下修改 alertCache

- **文件**: `agent/business_plugins/detector/anomaly_login/ssh/ssh_anomaly.go:156-220`
- **描述**: `Check` 方法获取的是 `d.mu.RLock()`（读锁），但在第 179 行和第 203 行修改了 `d.alertCache` map。多个 goroutine 同时调用 `Check` 时，对 map 的并发写入会导致 Go runtime panic。

```go
d.mu.RLock() // 第 157 行 - 获取读锁
...
d.alertCache[event.SourceIP] = time.Now() // 第 179 行 - 写操作!
```

- **修复建议**: 将涉及 `alertCache` 写操作的部分改为使用写锁（`d.mu.Lock()`）。

### H-08: 管道文件描述符泄漏（Load 函数错误路径）

- **文件**: `agent/plugin/plugin_linux.go:98-110`
- **描述**: 创建了两对管道（`rx_r/rx_w` 和 `tx_r/tx_w`），但如果第二对管道创建失败，第一对管道不会被关闭。如果 `cmd.Start()` 失败，部分管道也不会被关闭。
- **修复建议**: 使用 `defer` 或在每个错误路径上确保关闭所有已创建的文件描述符。

### H-09: baseline 插件无限增长的 goroutine

- **文件**: `agent/business_plugins/baseline/main.go:74-101`
- **描述**: 每收到一个任务就启动一个新的 goroutine，没有任何并发控制。如果服务端快速下发大量任务，会导致无限制的 goroutine 增长，最终耗尽内存。
- **修复建议**: 使用 worker pool 或 semaphore 限制并发执行的任务数量。

---

## Medium 级别 (13个)

### M-01: config.globalConfig 没有并发保护

- **文件**: `agent/config/config.go:230-263`
- **描述**: `SetStandalone` 函数直接修改 `globalConfig` 的字段，`Get()` 和 `IsStandalone()` 函数直接读取。虽然 `Init()` 使用了 `sync.Once`，但后续的写操作和读操作之间没有同步。
- **修复建议**: 使用 `sync.RWMutex` 保护 `globalConfig` 的所有读写操作。

### M-02: resource.GetProcResouce 潜在的除零 panic

- **文件**: `agent/resource/resource.go:95-100`
- **描述**: 当 `startAt == now.Unix()` 时（进程刚启动），`float64(now.Unix() - startAt)` 为 0，导致除零产生 `+Inf` 或 `NaN`。
- **修复建议**: 在除法前检查分母是否为零。

### M-03: baseline BindYaml 文件句柄泄漏

- **文件**: `agent/business_plugins/baseline/infra/yaml.go:10-18`
- **描述**: `BindYaml` 函数打开文件后没有调用 `f.Close()`。每次调用都会泄漏一个文件描述符。

```go
if f, err := os.Open(filePath); err != nil {
} else {
    err = yaml.NewDecoder(f).Decode(yamlMap)
    return err // f 从未被 Close!
}
```

- **修复建议**: 添加 `defer f.Close()`。

### M-04: baseline infra/log.go 文件句柄泄漏

- **文件**: `agent/business_plugins/baseline/infra/log.go:11-25`
- **描述**: `init()` 函数打开日志文件后赋给 `log.New()`，但 `logFile` 句柄从未被关闭。
- **修复建议**: 保存 `logFile` 句柄的引用以便程序退出时清理。

### M-05: resource.GetDNS 没有 defer Close

- **文件**: `agent/resource/resource_linux.go:29-41`
- **描述**: `GetDNS` 中打开文件后用 `f.Close()` 关闭，但没有使用 `defer`。如果 scanner 使用过程中发生 panic，文件句柄会泄漏。
- **修复建议**: 改用 `defer f.Close()`。

### M-06: collector engine handler 通道死锁风险

- **文件**: `agent/business_plugins/collector/engine/engine.go:76-94`
- **描述**: `handler.Handle` 方法使用容量为 1 的 channel 作为互斥锁。如果 `h.Handler.Handle()` panic，信号不会被写回，导致后续所有调用永远阻塞。
- **修复建议**: 使用标准的 `sync.Mutex` 替代 channel 互斥机制，并使用 `defer` 确保 panic 时也能正确释放锁。

### M-07: main.go 中 Context/Cancel 命名与 agent 包冲突

- **文件**: `agent/main.go:86` 与 `agent/agent/id.go:14`
- **描述**: `main.go` 创建了本地的 `Context, Cancel`，而 `agent` 包也定义了全局的 `Context, Cancel`。`transfer.go:184` 调用的是 `agent.Cancel()`，这是 agent 包的全局 Cancel，与 main.go 中的 Cancel 无关。这意味着服务端发送 1060 命令时调用 `agent.Cancel()` 不会触发 main.go 中的 `Cancel()`，信号处理 goroutine 和插件守护进程都不会被关闭。
- **修复建议**: 统一使用一个 Context/Cancel，或将 `agent.Cancel()` 与 `main.go` 的 Cancel 关联起来。

### M-08: sync.Map 重新赋值非线程安全

- **文件**: `agent/plugin/plugin.go:281`
- **描述**: 在 Startup 的 `ctx.Done()` 分支中，执行 `m = &sync.Map{}` 直接替换了包级别的 `sync.Map` 指针。其他 goroutine 正在使用旧 map 时会导致不一致。
- **修复建议**: 使用 `Range + Delete` 清空现有 map，而不是替换指针。

### M-09: standalone output 文件写入未检查错误

- **文件**: `agent/standalone/output.go:226`
- **描述**: `writeJSON` 函数中 `file.Write(append(data, '\n'))` 的返回值被忽略。磁盘满或文件系统只读时写入失败不会被察觉。
- **修复建议**: 检查 `file.Write` 的返回错误并记录日志。

### M-10: ReceiveData 中基于不可信长度的内存分配

- **文件**: `agent/plugin/plugin.go:86-143`
- **描述**: `ReceiveData` 函数从管道读取 `uint32` 长度值，直接用它计算需要读取的数据大小。如果插件发送的长度值被篡改或损坏，可能导致分配极大的内存缓冲区。`ReceiveStandardRecord` 同样直接用 `length` 分配 `make([]byte, length)`，没有上限检查。
- **修复建议**: 对从管道读取的长度值设置合理的上限（如 10MB），超过则返回错误。

### M-11: collector/engine BeforeDawn 函数返回 -1 导致 panic

- **文件**: `agent/business_plugins/collector/engine/engine.go:103-105, 122-132`
- **描述**: `BeforeDawn()` 返回 `time.Duration(-1)`，当 interval 不是 -1 且 minutes <= 0 时会执行 `panic("unknown interval")`。使用 magic number 表示特殊调度模式是脆弱的设计。
- **修复建议**: 使用显式的枚举类型或独立字段表示调度模式。

### M-12: baseline Analysis 函数缺少 default 分支

- **文件**: `agent/business_plugins/baseline/check/analysis.go:134-148`
- **描述**: `switch data.(type)` 只处理 `int` 和 `string` 类型，缺少 `default` 分支。传入其他类型时 `taskData` 保持零值，导致后续行为不可预期。
- **修复建议**: 添加 `default` 分支返回错误。

### M-13: SlidingWindow 内存无界增长

- **文件**: `agent/business_plugins/detector/engine/window.go`
- **描述**: `SlidingWindow` 的 `events` map 按 IP 存储事件列表，但 `cleanupLoop` 目前只打印 debug 日志，没有调用 `Cleanup()` 方法。高流量暴力破解场景下 map 会持续增长。
- **修复建议**: 在 `cleanupLoop` 中定期调用 `Cleanup()` 方法。

---

## Low 级别 (14个)

### L-01: rand.Seed 已废弃

- **文件**: `agent/business_plugins/collector/main.go:28`, `agent/business_plugins/baseline/main.go`
- **描述**: `rand.Seed(time.Now().UnixNano())` 在 Go 1.20+ 中已废弃（自动播种）。
- **修复建议**: 移除 `rand.Seed` 调用。

### L-02: GOMAXPROCS 硬编码

- **文件**: `agent/business_plugins/collector/main.go:27`（`GOMAXPROCS(8)`）, `agent/business_plugins/baseline/main.go:71`（`GOMAXPROCS(4)`）
- **描述**: 硬编码 GOMAXPROCS 在核心数少的机器上浪费调度开销，核心数多的机器上限制并行度。
- **修复建议**: 移除硬编码，使用 Go 运行时默认值。

### L-03: 使用已废弃的 ioutil.ReadAll

- **文件**: `agent/business_plugins/baseline/check/rules.go:209`
- **描述**: `ioutil.ReadAll` 在 Go 1.16 中已废弃。
- **修复建议**: 替换为 `io.ReadAll(file)`。

### L-04: os.MkdirAll 错误被忽略

- **文件**: `agent/business_plugins/collector/main.go:37`, `agent/business_plugins/detector/main.go:104`, `agent/business_plugins/baseline/infra/log.go:15`
- **描述**: 多处 `os.MkdirAll` 的返回错误被忽略。
- **修复建议**: 检查 `MkdirAll` 的返回错误。

### L-05: 通过执行 cat 命令读取文件

- **文件**: `agent/business_plugins/baseline/linux/os_system.go:10`
- **描述**: `GetSystemType` 使用 `exec.Command("cat", "/etc/issue")` 读取文件。
- **修复建议**: 替换为 `os.ReadFile("/etc/issue")`。

### L-06: 不精确的系统类型判断

- **文件**: `agent/business_plugins/baseline/linux/os_system.go:9-20`
- **描述**: 只检查了 Ubuntu 和 Debian，其他所有系统都被归类为 centos，可能导致不正确的基线规则被应用。
- **修复建议**: 使用 `/etc/os-release` 进行更精确的系统类型判断。

### L-07: 不支持的 OS 导致 baseline 插件静默退出

- **文件**: `agent/business_plugins/baseline/main.go:127-129`
- **描述**: 当系统类型不是 centos/debian/ubuntu 时执行 `return`，整个插件进程终止。
- **修复建议**: 使用 `continue` 跳过本次任务，而非终止进程；记录警告日志。

### L-08: fmt.Println 调试输出残留

- **文件**: `agent/plugin/plugin_linux.go:238`, `agent/business_plugins/baseline/check/rules.go:255`
- **描述**: `fmt.Println("send task", n)` 和 `fmt.Println(username)` 是调试代码残留。
- **修复建议**: 改用结构化日志记录器（zap）或移除。

### L-09: 被注释的原子操作代码

- **文件**: `agent/plugin/plugin.go:140-141`, `agent/plugin/plugin_linux.go:239-240`
- **描述**: 多处原子计数器操作被注释掉，导致 `Plugin.GetState()` 返回的 TPS/速度统计始终为零，监控功能失效。
- **修复建议**: 恢复或移除这些注释代码。

### L-10: grpc.DialContext 已废弃

- **文件**: `agent/transport/connection.go:134`
- **描述**: `grpc.DialContext` 已被标记为废弃。
- **修复建议**: 迁移到 `grpc.NewClient` API。

### L-11: cstring 和 argsString 函数重复定义

- **文件**: `agent/business_plugins/ebpf_base_detector/util.go` 与 `agent/business_plugins/ebpf_base_detector/events/types.go`
- **描述**: 两个包中有完全相同的实现。
- **修复建议**: 在 events 包中导出这两个函数，main 包直接引用。

### L-12: collector/engine Cache.Put 可能 panic

- **文件**: `agent/business_plugins/collector/engine/engine.go:51-55`
- **描述**: `Cache.Put` 在 `c.m[dt]` 不存在时会 panic（写入 nil map）。
- **修复建议**: 在 `Put` 中检查并初始化内层 map。

### L-13: ExeItem 结构体 Name 字段过大

- **文件**: `agent/business_plugins/ebpf_base_detector/trusted/types.go:13-18`
- **描述**: `ExeItem.Name` 是 `[2048]byte`，但路径最大限制为 255/256 字节，浪费内存。
- **修复建议**: 将 `Name` 缩小为 `[256]byte`。

### L-14: Client.Close 方法中的 goroutine 泄漏

- **文件**: `agent/business_plugins/lib/client.go:33-42, 116-120`
- **描述**: `New()` 启动的定时刷新 goroutine 在 `Close()` 时没有被通知退出。
- **修复建议**: 添加 done channel，Close 时关闭该 channel 以通知 goroutine 退出。

---

# 二、HCIDS (服务端) 审查报告

## Critical 级别 (5个)

### C-01: 所有 REST API 端点缺少认证和授权机制

- **文件**: `hcids/internal/router/router.go:28-160`
- **描述**: 整个 HTTP API 没有任何认证中间件（无 JWT、无 API Key、无 Session），任何人可以直接访问所有端点，包括：向 Agent 下发任意命令（POST /api/task）、关闭 Agent（DataType 1060）、创建/删除系统用户、删除告警规则、发送基线检测任务等。这是一个安全平台本身最致命的安全漏洞。
- **修复建议**: 在路由组上添加认证中间件，至少使用 JWT Token 或 API Key 鉴权：

```go
authMiddleware := middleware.AuthMiddleware()
api := r.Group("/api1", authMiddleware)
```

### C-02: 密码明文存储

- **文件**: `hcids/internal/controller/system/user.go:35-44`, `hcids/internal/models/system/user.go:11`
- **描述**: 创建用户时密码直接明文写入数据库（`Passwd: user.Passwd`），没有任何哈希处理。User 模型的 `Passwd` 字段直接以 JSON 形式返回给客户端（`json:"passwd"`），在 GetUser 和 ListUsers 接口中会泄露所有用户密码。
- **修复建议**: 使用 bcrypt 哈希密码，并在 JSON 序列化时忽略密码字段：

```go
// 模型中
Passwd string `json:"-" gorm:"column:passwd;size:250"`

// 创建时
hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(user.Passwd), bcrypt.DefaultCost)
newUser.Passwd = string(hashedPwd)
```

### C-03: CORS 配置允许所有来源 + 携带凭证

- **文件**: `hcids/internal/middleware/cors.go:11-60`, `hcids/conf/server.yaml:14-31`
- **描述**: 默认配置 `AllowedOrigins: ["*"]` 且 `AllowCredentials: true`。根据 CORS 规范，`Access-Control-Allow-Origin` 设为请求方的 Origin 加上 `Access-Control-Allow-Credentials: true` 等同于完全信任任何来源。攻击者可以从任意恶意网站通过浏览器访问该安全平台的所有 API。
- **修复建议**: 明确配置允许的域名列表，不使用通配符：

```yaml
cors:
  allowed_origins:
    - "https://security-dashboard.example.com"
  allow_credentials: true
```

### C-04: 数据库密码硬编码在配置文件中

- **文件**: `hcids/conf/server.yaml:38`
- **描述**: 数据库密码 `password: "root"` 硬编码在版本控制的配置文件中。如果代码仓库泄露，数据库凭据直接暴露。
- **修复建议**: 使用环境变量注入敏感配置：

```go
password := os.Getenv("DB_PASSWORD")
```

### C-05: gRPC 通道无 TLS、无认证

- **文件**: `hcids/cmd/main.go:108-111`
- **描述**: gRPC Server 创建时没有配置 TLS 证书，也没有任何 interceptor 进行 Agent 身份验证。任何人可以伪造 Agent 向服务端发送虚假安全事件数据，也可以冒充已有 Agent 接管其命令通道。攻击者可以：(1) 注入虚假安全告警制造混乱；(2) 冒充 Agent 获取下发的安全检测命令；(3) 污染资产数据和漏洞扫描结果。
- **修复建议**: 添加 mTLS 或 Token 认证 interceptor：

```go
grpcServer := grpc.NewServer(
    grpc.Creds(credentials.NewTLS(tlsConfig)),
    grpc.UnaryInterceptor(authInterceptor),
    grpc.StreamInterceptor(streamAuthInterceptor),
)
```

---

## High 级别 (6个)

### H-01: 分页参数 limit 未设上限，可导致 OOM

- **文件**: 所有 controller 的 List 方法，例如 `hcids/internal/controller/alert/command.go:23`, `hcids/internal/controller/assets/host/host.go:23`, `hcids/internal/controller/system/user.go:84`
- **描述**: `limit` 参数直接从用户请求获取，无上限检查。攻击者可传入 `limit=999999999`，导致一次性加载全量数据到内存，造成 OOM 或极大的 DB 压力。
- **修复建议**:

```go
if limit <= 0 || limit > 100 {
    limit = 10
}
```

### H-02: HTTP 服务未优雅关闭

- **文件**: `hcids/cmd/main.go:122-127`
- **描述**: HTTP 服务通过 `go func() { httpRouter.Run(httpAddr) }()` 启动后，在收到关闭信号时只调用了 `grpcServer.GracefulStop()`，没有关闭 HTTP 服务。HTTP 请求可能在进程退出时被强制中断。
- **修复建议**: 使用 `http.Server` 并在关闭信号处理中调用 `httpServer.Shutdown(ctx)`。

### H-03: 无请求速率限制

- **文件**: `hcids/internal/router/router.go`（全文）
- **描述**: 所有 API 端点没有速率限制。配合 C-01（无认证），攻击者可以对系统发起大量请求，导致数据库过载。
- **修复建议**: 添加速率限制中间件。

### H-04: Agent 连接被冒充后的指令劫持风险

- **文件**: `hcids/internal/grpc/handler/transfer.go:273-319`
- **描述**: `registerAgent` 方法中，如果一个新连接使用已有的 `agent_id`，会直接覆盖 `s.agents[pkg.AgentId]`，旧的命令通道被丢弃，但旧的 Transfer goroutine 仍在运行中，造成 goroutine 泄漏。攻击者可通过伪造 agent_id 劫持合法 Agent 的命令通道。
- **修复建议**: 在注册新 Agent 时，检查是否已有同 ID 连接，先关闭旧连接再注册新连接。

### H-05: NormalizeOSVersion 函数存在 panic 风险

- **文件**: `hcids/internal/db/repository/vuln_repository.go:429, 443, 458, 470, 482, 497`
- **描述**: 在解析 OS 版本字符串时，代码直接访问 `p[0]` 而没有检查数组边界。
- **修复建议**: 在访问 `p[0]` 前添加 `len(p) > 0` 检查。

### H-06: 删除操作缺少权限控制和确认机制

- **文件**: `hcids/internal/controller/system/user.go:193-207`, `hcids/internal/controller/code/repos.go:205-219`, `hcids/internal/controller/back/alert.go:203-217`
- **描述**: 所有删除接口只接收 ID 即可删除，无任何权限验证。结合 C-01 无认证问题，任何人可以删除任何数据。
- **修复建议**: 添加权限验证中间件和操作日志记录。

---

## Medium 级别 (11个)

### M-01: DSN 中的 sslmode=disable

- **文件**: `hcids/internal/db/postgres.go:18`, `hcids/internal/mysql/db.go:25-26`
- **描述**: PostgreSQL 连接字符串中硬编码 `sslmode=disable`，数据库通信以明文传输。
- **修复建议**: 在配置中添加 `sslmode` 配置项，生产环境使用 `sslmode=verify-full`。

### M-02: 全局变量 db 无线程安全保护

- **文件**: `hcids/internal/db/postgres.go:14`, `hcids/internal/mysql/db.go:14`
- **描述**: 全局变量 `db` 在 `Init` 和 `GetDB` 中无 `sync.Once` 或 mutex 保护。
- **修复建议**: 使用 `sync.Once` 初始化，或通过依赖注入传递 db 连接。

### M-03: 数据库连接池配置不完整

- **文件**: `hcids/internal/db/postgres.go:34-35`
- **描述**: 缺少 `ConnMaxLifetime` 和 `ConnMaxIdleTime` 配置，可能导致连接长时间不回收，被数据库端超时关闭后客户端复用失效连接。
- **修复建议**:

```go
sqlDB.SetConnMaxLifetime(30 * time.Minute)
sqlDB.SetConnMaxIdleTime(5 * time.Minute)
```

### M-04: eBPF 事件表缺少数据过期清理

- **文件**: `hcids/internal/db/repository/execve_repository.go`, `hcids/internal/db/repository/connect_repository.go`, `hcids/internal/db/repository/dns_repository.go`
- **描述**: `event_execve`、`event_connect`、`event_dns` 三张表只插入不删除。虽然 `ExecveRepository` 有 `DeleteOldRecords` 方法，但从未被调用。长期运行后磁盘爆满。
- **修复建议**: 定时调用清理方法，或使用 PostgreSQL 分区表。

### M-05: User 模型表名拼写错误

- **文件**: `hcids/internal/models/system/user.go:21`
- **描述**: `TableName()` 返回 `"systen_user"`（应为 `"system_user"`）。
- **修复建议**: 确认数据库实际表名并修正拼写。

### M-06: 多处 fmt.Println 调试输出残留

- **文件**: `hcids/internal/controller/system/user.go:32`, `hcids/internal/controller/code/repos.go:45`, `hcids/internal/controller/back/baseline.go:32`
- **描述**: 多处使用 `fmt.Println` 打印用户提交的数据，生产环境中敏感数据会被输出到标准输出。
- **修复建议**: 移除所有 `fmt.Println`，改用结构化日志。

### M-07: LIKE 查询中用户输入未转义通配符

- **文件**: 所有 controller 的 List 方法
- **描述**: 查询条件中直接将用户输入拼接到 LIKE 模式中（`"%"+agentID+"%"`），用户可在输入中包含 `%` 或 `_` 通配符绕过搜索意图。
- **修复建议**:

```go
escaped := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(input)
query = query.Where("agent_id LIKE ?", "%"+escaped+"%")
```

### M-08: DateTime UnmarshalJSON 缺少边界检查

- **文件**: `hcids/internal/models/common/time.go:24`
- **描述**: 直接使用 `data[1:len(data)-1]` 去除引号，没有检查 `data` 长度。空字符串或单字符输入会导致 panic。
- **修复建议**:

```go
func (dt *DateTime) UnmarshalJSON(data []byte) error {
    if len(data) < 2 {
        return fmt.Errorf("invalid datetime format")
    }
    // ...
}
```

### M-09: 错误信息泄露内部细节

- **文件**: `hcids/internal/http/handler.go:55`, `hcids/internal/controller/back/alert.go:43, 49`
- **描述**: 多处将 `err.Error()` 直接返回给客户端，可能暴露内部实现细节。
- **修复建议**: 返回通用错误消息，将详细错误记录到日志。

### M-10: gRPC 数据处理缺少 context 超时控制

- **文件**: `hcids/internal/grpc/handler/transfer.go:392`
- **描述**: `handlePackagedData` 中使用 `context.Background()` 创建 context，没有设置超时。如果数据库操作阻塞，处理 goroutine 会永久挂起。
- **修复建议**:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

### M-11: 漏洞调度器 vulnCount 始终返回 0

- **文件**: `hcids/internal/vuln/scheduler.go:112-143, 147-203`
- **描述**: `matchAllHosts` 和 `matchAllImages` 方法的返回值 `vulnCount` 从未被赋值，始终为 0。日志中漏洞数量统计功能失效。
- **修复建议**: 在匹配成功后累加 vulnCount。

---

## Low 级别 (8个)

### L-01: 包命名不当：mysql 包实际使用 PostgreSQL

- **文件**: `hcids/internal/mysql/db.go`
- **描述**: 包名为 `mysql`，但实际使用的是 `gorm.io/driver/postgres` PostgreSQL 驱动，造成认知混淆。
- **修复建议**: 将包名重命名为 `database` 或合并到 `db` 包。

### L-02: 路由组织混乱，存在重复注册

- **文件**: `hcids/internal/router/router.go:72-81` vs `hcids/internal/router/back_router.go:14-23`
- **描述**: 规则集路由在 `/api1/rules` 和 `/api1/back/rules` 下各注册了一次，完全重复。
- **修复建议**: 统一路由组织，去除重复注册。

### L-03: UpdateUser 路由使用 GET 方法

- **文件**: `hcids/internal/router/system_router.go:24`
- **描述**: `r.GET("/users/edit/:id", userHandler.UpdateUser)` 使用 GET 方法处理更新操作，但 UpdateUser 内部调用 `c.ShouldBindJSON` 解析 request body。GET 请求通常不包含 body。
- **修复建议**: 改为 `r.PUT` 或 `r.POST`。

### L-04: 无数据库迁移版本管理

- **文件**: `hcids/internal/models/init_menu.go:10-41`
- **描述**: `AutoMigrate` 函数已被完全注释掉。SQL 迁移文件使用编号命名但没有版本管理工具，依赖手动执行。
- **修复建议**: 集成 golang-migrate 或 goose 进行自动化迁移管理。

### L-05: GetAgents 返回内部指针

- **文件**: `hcids/internal/grpc/handler/transfer.go:860-869`
- **描述**: `GetAgents()` 在持有 RLock 时将 `*AgentInfo` 指针放入返回的 slice，释放锁后调用方仍持有这些指针。其他 goroutine 可能同时修改 AgentInfo 字段，造成数据竞争。
- **修复建议**: 返回 AgentInfo 的值拷贝而非指针。

### L-06: cron 表达式解析过于简陋

- **文件**: `hcids/internal/vuln/scheduler.go:218-247`
- **描述**: `parseCronToInterval` 只处理最简单的 cron 格式，大部分有效表达式退化为默认 24 小时。
- **修复建议**: 使用 `robfig/cron` 库进行标准 cron 解析。

### L-07: tar 解压缺少大小限制（decompression bomb）

- **文件**: `hcids/internal/vuln/dbmanager.go:265-270`
- **描述**: 虽然使用 `filepath.Base(header.Name)` 防止了路径穿越，但 `io.Copy(f, tr)` 没有限制解压大小。
- **修复建议**: 使用 `io.LimitReader` 限制单文件解压大小。

### L-08: DSN 中遗留 MySQL 配置项

- **文件**: `hcids/internal/config/config.go:27-29`, `hcids/conf/server.yaml:41`
- **描述**: `DatabaseConfig` 结构体包含 `Charset: "utf8mb4"`、`ParseTime`、`Loc` 等 MySQL 专用字段，但实际使用 PostgreSQL。
- **修复建议**: 移除 MySQL 专用配置项，添加 PostgreSQL 相关配置。

---

# 三、优先修复建议

## P0 - 立即修复（安全产品自身安全问题）

| # | 问题 | 模块 | 描述 |
|---|------|------|------|
| 1 | HCIDS C-01 | 服务端 | API 认证机制完全缺失 |
| 2 | HCIDS C-02 | 服务端 | 密码明文存储和 API 泄露 |
| 3 | Agent C-02 + HCIDS C-05 | 双端 | gRPC 通道无 TLS/无认证 |
| 4 | Agent C-04 | 客户端 | 插件签名验证被注释 |
| 5 | HCIDS C-03 | 服务端 | CORS 配置允许全域 |
| 6 | Agent C-01 | 客户端 | 命令注入漏洞 |
| 7 | Agent C-03 | 客户端 | 无鉴权远程关闭 Agent |

## P1 - 尽快修复

| # | 问题 | 模块 | 描述 |
|---|------|------|------|
| 1 | HCIDS H-01 | 服务端 | 分页 limit 上限 |
| 2 | Agent H-07 | 客户端 | RLock 下写 map 致 panic |
| 3 | Agent H-01 | 客户端 | Ring Buffer 安全事件静默丢弃 |
| 4 | HCIDS H-04 | 服务端 | Agent 连接冒充劫持 |
| 5 | Agent M-07 | 客户端 | 双 Context/Cancel 导致关闭命令无效 |
| 6 | HCIDS C-04 | 服务端 | 数据库密码改用环境变量 |
| 7 | Agent H-08 | 客户端 | 管道文件描述符泄漏 |
| 8 | Agent M-03 | 客户端 | BindYaml 文件句柄泄漏 |

## P2 - 计划修复

| # | 问题 | 模块 | 描述 |
|---|------|------|------|
| 1 | HCIDS M-04 | 服务端 | 事件表数据过期清理 |
| 2 | HCIDS M-01 | 服务端 | 数据库连接启用 SSL |
| 3 | HCIDS M-10 | 服务端 | context 超时控制 |
| 4 | Agent M-13 | 客户端 | SlidingWindow 内存无界增长 |
| 5 | Agent H-09 | 客户端 | baseline 插件 goroutine 无限增长 |
| 6 | 其余 Medium/Low 问题 | 双端 | 逐步修复 |

---

# 四、总体评价

## 架构层面

两个模块的整体架构设计合理：

- **Agent**: 插件化架构清晰，Ring Buffer + gRPC 双向流的数据流设计得当，eBPF 内核监控与用户态检测的分层合理。
- **HCIDS**: gRPC 接收 -> Mapper 转换 -> Repository 持久化的分层清晰，Controller/Router/Model 的 MVC 结构规范。

## 核心问题

作为安全产品，其自身的安全防护几乎为零。9 个 Critical 问题中有 7 个与认证、加密、签名相关。这是 AI 生成代码的典型特征 -- 功能实现完整但安全加固缺失。

## 建议

1. 上线前必须解决所有 Critical 和 High 级别问题。
2. 重点关注 gRPC 双向 TLS 认证、API 鉴权、密码哈希这三个方向。
3. 建议引入 `go vet`、`staticcheck`、`golangci-lint` 等静态分析工具到 CI 流程中。
4. 建议使用 `-race` 标志运行测试以检测数据竞争问题。
