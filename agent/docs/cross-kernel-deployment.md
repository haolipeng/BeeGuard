# 跨内核部署指南（内核 > 5.10）

## 现状评估

项目已具备较好的跨内核基础：

- 使用 **CO-RE**（`BPF_CORE_READ`）+ `vmlinux.h`，不依赖硬编码偏移
- 使用 **cilium/ebpf** 库，运行时自动 BTF 重定位
- Hook 点选择稳定（raw_tracepoint、kprobe on LSM hooks）
- eBPF 对象通过 `bpf2go` 编译后嵌入 Go 二进制，部署时无需编译环境

**结论：编译一次即可在不同内核上运行，不需要在目标机器上重新编译。** 但需要确认以下几项工作。

---

## 一、目标机器环境检查（必须）

| 检查项 | 要求 | 检查命令 |
|--------|------|----------|
| 内核版本 | >= 5.8（> 5.10 已满足） | `uname -r` |
| BTF 支持 | 内核编译时启用 `CONFIG_DEBUG_INFO_BTF=y` | `ls /sys/kernel/btf/vmlinux` |
| 架构 | x86-64（当前代码硬编码了 x86 syscall 号） | `uname -m` |
| BPF 子系统 | 已启用 | `ls /sys/fs/bpf/` |

**BTF 是最关键的前置条件**。若目标机器无 `/sys/kernel/btf/vmlinux`，CO-RE 无法工作。

> 主流发行版 BTF 支持情况：
> - Ubuntu 20.10+ / 22.04 LTS — 默认开启
> - CentOS 8.2+ / RHEL 8.2+ — 默认开启
> - Debian 11+ — 默认开启
> - Amazon Linux 2 — **需手动确认**，部分 AMI 未开启

---

## 二、无 BTF 内核的兼容方案（按需）

如果部分目标机器**没有开启 BTF**，需要做以下工作：

### 方案 A：引入 BTF Hub（推荐）

[btfhub-archive](https://github.com/aquasecurity/btfhub-archive) 提供了主流发行版预生成的 BTF 文件。

工作内容：

1. 在 loader.go 中加载 eBPF 对象时，检测 `/sys/kernel/btf/vmlinux` 是否存在
2. 若不存在，根据 `uname -r` + 发行版信息匹配离线 BTF 文件
3. 使用 `cilium/ebpf` 的 `btf.Spec` 参数传入外部 BTF
4. 将所需的 BTF 文件打包或做成按需下载机制

### 方案 B：要求目标机器手动安装 BTF

- Debian/Ubuntu: `apt install linux-image-$(uname -r)-dbg`
- CentOS/RHEL: `yum install kernel-debuginfo-$(uname -r)`

---

## 三、运行时内核兼容性加固（建议）

当前代码在不同内核版本间**可能遇到的问题**及应对：

### 3.1 内核结构体字段变化处理

当前代码大量使用 `BPF_CORE_READ` 访问 `task_struct`、`file`、`dentry` 等内核结构体。CO-RE 会自动处理字段偏移变化，但**字段被移除/重命名**的情况需要主动处理。

工作内容：

- 使用 `bpf_core_field_exists()` 对关键字段做存在性检查
- 对 5.10 ~ 6.x 内核中已知的结构体变化做适配（如 `nsproxy`、`mount` 相关字段）

### 3.2 BPF Helper 可用性

当前使用的 helper 在 5.10+ 内核上均可用，**无需额外处理**：

| Helper | 最低内核版本 | 5.10 可用 |
|--------|-------------|-----------|
| `bpf_get_current_task()` | 4.8 | 是 |
| `bpf_probe_read_kernel()` | 5.5 | 是 |
| `bpf_perf_event_output()` | 4.4 | 是 |
| `bpf_get_current_pid_tgid()` | 4.2 | 是 |

### 3.3 Kprobe 符号稳定性

当前 kprobe 挂载的内核函数（`commit_creds`、`security_inode_create/rename/unlink`）在所有 5.10+ 内核中均存在。建议：

- 在 loader 中增加 kprobe attach 失败的 graceful 降级处理
- 记录日志告知哪些 hook 点未成功挂载

---

## 四、部署与分发（必须）

### 4.1 编译产物确认

由于使用了 `bpf2go`，eBPF 字节码已嵌入 Go 二进制。部署只需分发单个二进制 + 配置文件：

```
/opt/cloudsec/agent/
├── agent              # 单一二进制，包含 eBPF 字节码
├── config.yaml
└── plugin_configs/
```

**不需要在目标机器上安装 clang、llvm、libbpf-dev、linux-headers。**

### 4.2 权限要求

- 必须以 **root** 运行（eBPF 加载需要 `CAP_SYS_ADMIN` 或 `CAP_BPF`）
- 5.8+ 内核支持 `CAP_BPF`（更细粒度的权限），但建议仍用 root

### 4.3 编写部署前检查脚本

建议编写一个 `preflight-check.sh`，在部署前自动检查目标机器环境：

```bash
#!/bin/bash
# 检查内核版本 >= 5.10
# 检查架构 = x86_64
# 检查 BTF: /sys/kernel/btf/vmlinux
# 检查 BPF 文件系统: /sys/fs/bpf/
# 检查 CAP_SYS_ADMIN / root
```

---

## 五、多架构支持（可选，当前不需要）

当前代码仅支持 x86-64，原因：

- `bpf2go` 编译参数：`-target amd64 -D__TARGET_ARCH_x86`
- syscall 号硬编码：connect=42, bind=49, accept=43 等（x86-64 专用）

如果未来需要支持 ARM64（如华为鲲鹏、AWS Graviton），需要：

1. 在 `loader.go` 中增加 `bpf2go` 的 arm64 target
2. 处理 syscall 号的架构差异（通过 `#ifdef` 或 CO-RE 宏）
3. 重新生成 ARM64 的 vmlinux.h

---

## 工作优先级总结

| 优先级 | 工作项 | 工作量 |
|--------|--------|--------|
| **P0** | 确认目标机器 BTF 支持 + 编写 preflight 检查脚本 | 小 |
| **P0** | 在目标机器上实际测试部署，验证 eBPF 加载成功 | 小 |
| **P1** | loader 增加错误处理和 graceful 降级 | 中 |
| **P1** | 无 BTF 机器的 BTFHub 兼容方案 | 中 |
| **P2** | 关键结构体字段的 `bpf_core_field_exists` 检查 | 中 |
| **P3** | ARM64 多架构支持 | 大 |

**如果目标机器都是 5.10+ 且有 BTF 支持的 x86-64 机器，当前代码已经可以直接部署运行，核心工作只是 P0 的环境检查和测试验证。**
