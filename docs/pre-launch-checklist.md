# Agent 上线前检查清单

本文档记录了 Agent 项目上线前需要完成的所有事项，包括 bug 修复、BTF 兼容性适配、测试体系建设、生产环境加固等。

---

## 一、Bug 修复

### P0 - 阻塞上线

| # | 问题 | 文件 | 说明 |
|---|------|------|------|
| 1 | **Signal handler 资源泄漏** | `ebpf_base_detector/main.go:180-182` | 调用了 `signal.Notify()` 但缺少 `signal.Stop(sig)` 和 `close(sig)`，导致信号 goroutine 泄漏 |
| 2 | **MaliciousRequestDetector 数据竞争** | `malicious_request_detector.go:82` | `buildIndex()` 写入 `ruleCount` 时未持锁，但 `GetEnabledRuleCount()` 读取时用了 `RLock`，存在 data race |
| 3 | **Perf buffer 丢失事件无上报** | `ebpf_base_detector/main.go:142-144` | `LostSamples` 只打了日志，没有上报到 Server 端，生产环境中无法感知数据丢失 |

### P1 - 高优先级

| # | 问题 | 文件 | 说明 |
|---|------|------|------|
| 4 | **端口号转换错误静默忽略** | `malicious_request_detector.go:73` | `strconv.Atoi(portStr)` 错误被 `_` 吞掉，无效端口会变成 0，干扰检测 |
| 5 | **/proc 读取竞态** | `proc.go` | 进程在 perf 事件到达用户态之间可能已退出，`resolveExePath` / `buildPidTree` 未做容错 |
| 6 | **时间戳精度不一致** | `reverse_shell.go:98` vs `event_handlers.go` | 告警用 `Unix()`（秒），事件用 `UnixMilli()`（毫秒），上报到 Server 后关联分析困难 |
| 7 | **白名单加载失败静默降级** | `main.go:87-89, 104-106` | trusted map 加载失败只打 Warn，系统在无白名单保护下运行，应至少上报告警 |

### 修复参考

#### Bug #1: Signal handler 资源泄漏

```go
sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
<-sig
signal.Stop(sig)  // 新增
close(sig)        // 新增
```

#### Bug #2: MaliciousRequestDetector 数据竞争

```go
// 在 buildIndex() 中保护 ruleCount 写入：
m.mu.Lock()
m.ruleCount = enabledCount
m.mu.Unlock()
```

#### Bug #4: 端口号转换错误

```go
port, err := strconv.Atoi(portStr)
if err != nil {
    // 记录日志并跳过该规则
    continue
}
if port < 0 || port > 65535 {
    continue
}
m.portIndex[uint16(port)] = rule
```

---

## 二、BTF 适配与内��兼容性

当前代码完全依赖宿主机自带 BTF（`/sys/kernel/btf/vmlinux`），没有任何兼容层。这在生产环境中是最大的部署风险。

### 当前状态

| 组件 | 状态 | 说明 |
|------|------|------|
| vmlinux.h | ✅ 已有 | 预生成的 x86_64 版本，包含 CO-RE pragma |
| BPF_CORE_READ | ✅ 正确 | 50+ 处用法，嵌套结构体访问正确 |
| PT_REGS_PARM_CORE | ✅ 正确 | CO-RE 感知的参数提取 |
| cilium/ebpf 版本 | ⚠️ 过旧 | v0.12.0，缺少现代错误处理 |
| BTF 运行时检测 | ❌ 缺失 | 未检查 `/sys/kernel/btf/vmlinux` |
| 内核版本检测 | ❌ 缺失 | 实际需要 5.8+，但未校验 |
| BTFHub 支持 | ❌ 未实现 | 无动态 BTF 获取能力 |
| 多架构支持 | ❌ 仅 x86 | 无 ARM64/RISCV 支持 |
| 错误信息分类 | ❌ 过于笼统 | 不区分 BTF/版本/架构/权限等失败原因 |
| 优雅降级 | ❌ 无 | eBPF 加载失败则整个插件不可用 |

### 必须做

| # | 事项 | 说明 |
|---|------|------|
| 1 | **启动时检测 BTF 可用性** | `loader.go` 加载 eBPF 前检查 `/sys/kernel/btf/vmlinux` 是否存在，不存在时给出明确错误信息而非通用 `failed to load eBPF objects` |
| 2 | **最低内核版本检测** | 代码实际需要 **5.8+**（因为用了 `__builtin_preserve_field_info` 做 bitfield CO-RE），但文档只写了 "5.x"，需要在启动时校验 `uname -r` |
| 3 | **错误信息分类** | 区分 BTF 不可用 / 内核版本过低 / 架构不匹配 / 权限不足等不同失败原因，方便运维排查 |
| 4 | **vmlinux.h 生成文档化** | 当前 `vmlinux.h` 是预生成的，没有记录基于哪个内核版本生成，需要文档化并加入 CI |

### 建议做（覆盖更多环境）

| # | 事项 | 说明 |
|---|------|------|
| 5 | **集成 BTFHub** | 对没有内置 BTF 的旧内核（5.2-5.7），使用 [BTFHub](https://github.com/aquasecurity/btfhub-archive) 提供的预编译 BTF 文件做 CO-RE 重定位 |
| 6 | **升级 cilium/ebpf** | 当前 `v0.12.0` 较旧，`v0.15+` 有更好的 BTF 自动检测和错误上下文 |
| 7 | **ARM64 支持** | 当前只有 `bpf_bpfel_x86.go`（`-target amd64`），如果有 ARM 服务器需加 `-target arm64` 生成对应字节码 |
| 8 | **优雅降级机制** | eBPF 加载失败时，插件应上报自身状态为 "disabled" 而非让整个 agent 受影响 |

### BTF 检测参考实现

```go
func checkBTFSupport() error {
    // 1. 检查内核版本
    var uname syscall.Utsname
    if err := syscall.Uname(&uname); err != nil {
        return fmt.Errorf("failed to get kernel version: %w", err)
    }
    release := unix.ByteSliceToString(uname.Release[:])
    major, minor, err := parseKernelVersion(release)
    if err != nil {
        return fmt.Errorf("failed to parse kernel version %q: %w", release, err)
    }
    if major < 5 || (major == 5 && minor < 8) {
        return fmt.Errorf("kernel %d.%d is not supported, minimum required: 5.8", major, minor)
    }

    // 2. 检查 BTF 可用性
    if _, err := os.Stat("/sys/kernel/btf/vmlinux"); os.IsNotExist(err) {
        return fmt.Errorf("BTF not available: /sys/kernel/btf/vmlinux not found, " +
            "ensure CONFIG_DEBUG_INFO_BTF=y in kernel config")
    }

    // 3. 检查权限
    if os.Geteuid() != 0 {
        return fmt.Errorf("root privileges required for eBPF program loading")
    }

    return nil
}
```

---

## 三、测试体系建设

当前 ebpf_base_detector **没有任何自动化测试**，其他插件测试覆盖也不完整。

### 3.1 单元测试（必须补充）

| 模块 | 测试内容 | 优先级 |
|------|----------|--------|
| `dangerous_command.go` | 规则匹配正确性、边界情况（空字符串、超长命令）、regex 编译失败处理 | P0 |
| `reverse_shell.go` | 各 fd_type 组合的检测结果、无 TTY 场景、误报场景 | P0 |
| `sensitive_file.go` | 路径匹配、白名单过滤 | P0 |
| `malicious_request_detector.go` | 端口匹配、规则更新并发安全、index 构建 | P1 |
| `proc.go` | 进程退出后的容错、pid tree 构建深度限制 | P1 |
| `util.go` | `argsString()` 对空参数、超长参数、含空字节参数的处理 | P1 |
| `events/types.go` | `UnmarshalBinary()` 对畸形数据的处理 | P1 |
| `trusted/` | hash 计算一致性、map 填充/查询 | P2 |

### 3.2 eBPF 集成测试

| 测试项 | 方法 | 说明 |
|--------|------|------|
| **Hook 挂载** | 加载 eBPF 后验证所有 5 个 hook 返回有效 link | 验证内核兼容性 |
| **Execve 事件** | `fork+exec` 一个已知进程，验证 perf event 内容 | 端到端数据流 |
| **Commit_creds** | `sudo -u nobody true` 触发提权，验证事件 | 提权检测 |
| **Connect 事件** | `curl` 外部地址，验证连接事件 | 网络监控 |
| **DNS 事件** | `nslookup` 已知域名，验证解析事件 | DNS 审计 |
| **File 事件** | 创建/重命名敏感路径文件，验证文件事件 | 文件监控 |

### 3.3 性能测试

| 测试项 | 指标 | 方法 |
|--------|------|------|
| **事件吞吐量** | 每秒可处理多少事件不丢失 | 高并发 `fork+exec`（如 `stress-ng --fork`），监控 `LostSamples` |
| **CPU 开销** | eBPF hook 引入的 CPU overhead | 对比开启/关闭 eBPF 时同一负载下的 CPU 使用率（建议 < 3%） |
| **内存占用** | Perf buffer + Go 用户态内存 | 长时间运行后观察 RSS 增长曲线 |
| **Regex 匹配延迟** | 每条命令的正则匹配耗时 | 构造 1000 条不同长度的命令，benchmark `Detect()` |
| **Perf buffer 满载** | 丢失率 vs 缓冲区大小 | 当前 32 页（128KB/CPU），测试不同大小下的丢失率 |
| **Proc 读取开销** | `buildPidTree()` 延迟 | 在高 PID 翻转场景下测量 |

### 3.4 稳定性测试

| 测试项 | 持续时间 | 关注点 |
|--------|----------|--------|
| **长时间运行** | 72 小时 | 内存泄漏（RSS 是否持续增长）、goroutine 泄漏、fd 泄漏 |
| **高压运行** | 24 小时 + 持续高负载 | LostSamples 趋势、CPU 平稳性、无 panic |
| **进程频繁创建/退出** | `while true; do /bin/true; done` | /proc 竞态是否导致错误日志暴增 |
| **eBPF map 满载** | 填满 trusted_exes map（2048 条） | LRU 淘汰是否正常工作 |
| **规则热更新** | 运行中更新 YAML 规则 | 并发安全，不 panic |
| **异常恢复** | kill -9 agent 后重启 | eBPF 程序是否正确卸载、重新挂载 |

---

## 四、生产环境加固

### 4.1 可观测性

| 事项 | 说明 |
|------|------|
| **Metrics 上报** | 事件处理速率、丢失事件数、各检测器匹配次数、perf buffer 使用率 |
| **健康检查** | 定期验证 eBPF 程序仍在运行（link 未断开），不是"挂了但没人知道" |
| **日志分级** | 生产环境确保默认 Warn 级别，可动态切换到 Debug |

### 4.2 安全加固

| 事项 | 说明 |
|------|------|
| **规则文件完整性** | YAML 配置文件增加校验（hash/签名），防止被篡改导致白名单绕过 |
| **eBPF 程序签名** | 验证加载的 .o 文件完整性 |
| **权限最小化** | 检查是否可以用 `CAP_BPF + CAP_PERFMON` 替代 root |

### 4.3 部署与运维

| 事项 | 说明 |
|------|------|
| **Systemd 集成** | Watchdog、自动重启、OOM score 调整 |
| **Cgroup 资源限制** | 给 agent 自身设置 CPU/内存上限，防止因 bug 拖垮宿主机 |
| **升级方案** | 热升级还是重启升级？eBPF 程序替换期间是否有监控空窗期 |
| **回滚方案** | 配置回滚、二进制回滚的自动化脚本 |

---

## 五、当前未提交代码审查

当前 `dev` 分支有未提交的重构（从 `main.go` 拆分出 `event_handlers.go`、`proc.go`、`util.go`、`config_path.go`）。

**检查项：**

1. 确认重构后功能等价（无逻辑变更）
2. 提交前跑一遍完整的手动验证流程（各技能脚本）
3. 使用 `scripts/test-*.sh` 逐个验证检测能力

---

## 六、建议的执行顺序

```
Phase 1 - 阻塞上线
├── 修复 P0 bug（signal leak、data race、丢失事件上报）
├── BTF 检测 + 内核版本校验 + 错误信息细化
├── 补充核心检测器单元测试
└── 提交当前 dev 分支代码

Phase 2 - 上线前验证
├── eBPF 集成测试
├── 性能测试 + 压测
├── 72 小时稳定性测试
└── Metrics / 健康检查接入

Phase 3 - 上线后迭代
├── BTFHub 集成
├── ARM64 支持
├── cilium/ebpf 升级
├── 优雅降级机制
└── 规则文件签名
```
