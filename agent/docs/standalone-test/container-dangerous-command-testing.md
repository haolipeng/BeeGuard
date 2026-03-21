# 容器高危命令检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的容器高危命令检测功能（DataType 7001）：eBPF 在 `sched_process_exec` Hook 中捕获进程执行事件，通过 `mntns_id != root_mntns_id` 判断进程是否运行在容器内，用户态将容器内的命令行参数与 `container_dangerous_commands.yaml` 中的规则进行匹配，匹配成功时产生告警。本文档选取 3 条规则的 7 个测试用例进行验证，覆盖 regex、prefix 两种匹配方式和 medium、high 两种严重程度。

> **注意**：`sched_process_exec` 仅在 `execve()` 系统调用**成功**时触发。如果容器内不存在对应的可执行文件（如 `wget`），`execve()` 会失败，eBPF 不会捕获到该事件。因此测试前需确保容器内安装了所需命令。

### 已知限制：脚本类命令无法检测

`pip`、`npm`、`gem` 等命令本质上是脚本文件（shebang 为 `#!/usr/bin/python3`、`#!/usr/bin/node` 等），执行时内核 `comm` 字段为解释器名（如 `python3`），而非命令名本身。检测器的反误报过滤机制会验证 `comm` 与规则期望的命令名是否一致，因此这类脚本命令无法被正确检测。`container_dangerous_commands.yaml` 中的 `pip3?\s+install`、`npm\s+install`、`gem\s+install` 模式在当前实现下不会触发告警，属于已知限制。

### 与宿主机高危命令检测的区别

| 维度 | 宿主机高危命令（6003） | 容器高危命令（7001） |
|------|----------------------|---------------------|
| 检测范围 | 所有进程 | 仅容器内进程 |
| 规则文件 | `dangerous_commands.yaml` | `container_dangerous_commands.yaml` |
| 容器判断 | 无 | `mntns_id != root_mntns_id` |
| 告警字段 | 无容器信息 | 包含 `container_id`、`is_container=true` |

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | Docker | `docker version` | Docker 已安装且运行中 |

如果任一条件不满足，测试无法进行。

---

## Step 1：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test
```

### 启动成功判定

启动后查看**插件日志文件**，**必须**看到以下日志行：

```bash
grep "Container dangerous command rules loaded" /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log
```

预期输出：

```
INFO  Container dangerous command rules loaded successfully  {"version": "1.0", "rules": 3}
```

**判定规则**：
- `"rules": 3` → 启动成功，3 条容器规则全部加载，进入 Step 2
- 出现 `Failed to load container dangerous command rules` → 规则文件缺失或格式错误，检查 `container_dangerous_commands.yaml` 是否在 `/opt/cloudsec/agent/plugins/ebpf_base_detector/config/` 目录下
- `failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3

同时还应看到：

```
INFO  Container escape detector initialized
INFO  Container metadata cache initialized
```

### 日志位置

Agent 启动后产生两个日志流：

| 位置 | 内容 | 说明 |
|------|------|------|
| Terminal A (stderr) | `dangerous command detected` | Agent 主进程的 standalone 输出，包含 `rule_id`、`rule_name`、`command`、`pid` 等字段，**不包含** `container_id` |
| `/opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log` | `Container dangerous command detected` | 插件进程日志，包含 `rule_id`、`comm`、`args`、**`container_id`** 等字段，**推荐用此日志验证** |

> **重要**：Terminal A 的 stderr 中显示的是 `dangerous command detected`（不区分容器/宿主机），而包含 `container_id` 的 `Container dangerous command detected` 仅出现在插件日志文件中。测试时应以插件日志为准。

### 搜索技巧

```bash
# 方式一：实时查看插件日志中的容器高危命令告警
tail -f /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log | grep "Container dangerous command detected"

# 方式二：按规则 ID 精确搜索
grep "rule_id.*3001" /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log

# 方式三：查看 Terminal A 的 standalone 输出（不含 container_id）
# 在 Terminal A 的 stderr 中可以看到类似如下的 JSON 日志：
# INFO  dangerous command detected  {"rule_id": "3001", "rule_name": "容器内包管理器安装", ...}

# 方式四：搜索已轮转的压缩日志（日志量大时当前文件可能已被轮转）
zcat /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector-*.log.gz 2>/dev/null | grep "Container dangerous command detected"
```

> **注意**：插件日志配置了自动轮转，当容器内安装大量软件包时会产生海量文件事件，可能导致日志快速轮转。如果在当前日志文件中找不到预期告警，请同时搜索已轮转的 `.log.gz` 文件。

---

## Step 2：启动测试容器

打开 **Terminal B**，启动一个 Docker 容器并安装测试所需工具：

```bash
docker run -it --rm --name container_test ubuntu:22.04 /bin/bash
```

容器启动后，在容器内安装测试所需的命令：

```bash
apt-get update && apt-get install -y wget curl cron 2>/dev/null; true
```

> 说明：`ubuntu:22.04` 基础镜像仅包含 `apt`/`apt-get`、`useradd` 等少量命令，`wget`、`curl`、`crontab` 均未预装。上述安装命令本身也会触发规则 3001 告警，属于预期行为。如无该镜像，可先 `docker pull ubuntu:22.04`。

---

## Step 3：执行测试用例

在 **Terminal B**（容器内）逐条执行以下测试命令。每执行一条后，在插件日志中查看是否出现对应告警。

### 告警日志格式

每条告警在插件日志中以一行 JSON 结构化日志输出：

```
{时间戳}  WARN  ebpf_base_detector/event_handlers.go:91  Container dangerous command detected  {"rule_id": {ID}, "rule_name": "{名称}", "severity": "{级别}", "uid": {UID}, "comm": "{进程名}", "args": "{完整命令行}", "container_id": "{容器ID}"}
```

> 注意：`args` 字段包含完整命令行（含命令名本身），例如 `"args": "apt-get install -y curl"`，而非仅参数部分。

### 通用判定规则

**PASS** 条件（全部满足）：
1. 插件日志出现包含 `Container dangerous command detected` 的日志行
2. `rule_id` 与测试用例的规则 ID 一致
3. `comm` 与执行的命令名一致
4. `container_id` 非空（64 位十六进制字符串）

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内插件日志无任何 `Container dangerous command detected` 输出
- `rule_id` 与预期不一致
- `container_id` 为空

---

### 用例 1：规则 3001 — 容器内包管理器安装（medium）

**匹配方式**：regex，匹配 `apt(-get)?\s+install`

**测试命令**（Terminal B，容器内）：

```bash
apt-get install -y curl 2>/dev/null; true
```

**预期日志**（插件日志文件）：

```
WARN  Container dangerous command detected  {"rule_id": 3001, "rule_name": "容器内包管理器安装", "severity": "medium", "uid": 0, "comm": "apt-get", "args": "apt-get install -y curl", "container_id": "..."}
```

**PASS 判定**：出现 `Container dangerous command detected`，且 `rule_id` 为 3001，`comm` 为 `apt-get`。

> 说明：即使 `apt-get install` 因网络原因失败，eBPF 在 execve 阶段已捕获，不影响检测（前提是 `apt-get` 二进制文件存在）。

---

### 用例 2：规则 3001 — 容器内 apt install（medium）

**匹配方式**：regex，匹配 `apt(-get)?\s+install`（验证不带 `-get` 的 `apt install` 变体）

**测试命令**（Terminal B，容器内）：

```bash
apt install -y curl 2>/dev/null; true
```

**预期日志**（插件日志文件）：

```
WARN  Container dangerous command detected  {"rule_id": 3001, "rule_name": "容器内包管理器安装", "severity": "medium", "uid": 0, "comm": "apt", "args": "apt install -y curl", "container_id": "..."}
```

**PASS 判定**：`rule_id` 为 3001，`comm` 为 `apt`。

> 说明：用例 1 测试 `apt-get install`，本用例测试 `apt install`，两者共同验证正则 `apt(-get)?\s+install` 的两个分支。

---

### 用例 3：规则 3002 — 容器内 wget 下载（high）

**匹配方式**：prefix，进程名以 `wget` 开头

**前提**：容器内需存在 `wget` 命令。`ubuntu:22.04` 未预装，需先执行 `apt-get install -y wget`。

**测试命令**（Terminal B，容器内）：

```bash
wget http://example.com 2>/dev/null; true
```

**预期日志**（插件日志文件）：

```
WARN  Container dangerous command detected  {"rule_id": 3002, "rule_name": "容器内网络下载工具", "severity": "high", "uid": 0, "comm": "wget", "args": "wget http://example.com", "container_id": "..."}
```

**PASS 判定**：`rule_id` 为 3002，`comm` 为 `wget`。

---

### 用例 4：规则 3002 — 容器内 curl 请求（high）

**匹配方式**：prefix，进程名以 `curl` 开头

**前提**：容器内需存在 `curl` 命令。`ubuntu:22.04` 未预装，需先执行 `apt-get install -y curl`。

**测试命令**（Terminal B，容器内）：

```bash
curl http://example.com 2>/dev/null; true
```

**预期日志**（插件日志文件）：

```
WARN  Container dangerous command detected  {"rule_id": 3002, "rule_name": "容器内网络下载工具", "severity": "high", "uid": 0, "comm": "curl", "args": "curl http://example.com", "container_id": "..."}
```

**PASS 判定**：`rule_id` 为 3002，`comm` 为 `curl`。

---

### 用例 5：规则 3003 — 容器内用户创建（high）

**匹配方式**：regex，匹配 `useradd\s+`

**测试命令**（Terminal B，容器内）：

```bash
useradd testuser123 2>/dev/null; true
```

**预期日志**（插件日志文件）：

```
WARN  Container dangerous command detected  {"rule_id": 3003, "rule_name": "容器内系统配置修改", "severity": "high", "uid": 0, "comm": "useradd", "args": "useradd testuser123", "container_id": "..."}
```

**PASS 判定**：`rule_id` 为 3003，`comm` 为 `useradd`。

---

### 用例 6：规则 3003 — 容器内 crontab 编辑（high）

**匹配方式**：regex，匹配 `crontab\s+-e`

**前提**：容器内需存在 `crontab` 命令。`ubuntu:22.04` 未预装，需先执行 `apt-get install -y cron`。

**测试命令**（Terminal B，容器内）：

```bash
crontab -e 2>/dev/null; true
```

**预期日志**（插件日志文件）：

```
WARN  Container dangerous command detected  {"rule_id": 3003, "rule_name": "容器内系统配置修改", "severity": "high", "uid": 0, "comm": "crontab", "args": "crontab -e", "container_id": "..."}
```

**PASS 判定**：`rule_id` 为 3003，`comm` 为 `crontab`。

---

### 用例 7（反向验证）：宿主机同命令不触发容器告警

退出容器，在 **宿主机** Terminal B 上执行：

```bash
curl http://example.com 2>/dev/null; true
```

**预期**：插件日志中**不应**出现新的 `Container dangerous command detected` 日志（排除其他已运行容器产生的告警）。

> 说明：宿主机上的 curl 可能触发宿主机的 `Dangerous command detected`（DataType 6003，取决于 `dangerous_commands.yaml` 配置），但**不应**触发容器的 `Container dangerous command detected`（DataType 7001）。如果环境中有其他正在运行的容器（如带有健康检查的容器），可能会在日志中看到来自这些容器的告警，应通过 `container_id` 字段区分。

**PASS 判定**：5 秒内插件日志无新增与宿主机 curl 相关的 `Container dangerous command detected` 输出。

---

## Step 4：记录测试结果

| # | 规则 ID | 规则名称 | 严重程度 | 测试命令 | 执行环境 | 预期 | 实际 | PASS/FAIL |
|---|---------|----------|----------|----------|----------|------|------|-----------|
| 1 | 3001 | 容器内包管理器安装 | medium | `apt-get install -y curl` | 容器内 | 告警 | | |
| 2 | 3001 | 容器内包管理器安装 | medium | `apt install -y curl` | 容器内 | 告警 | | |
| 3 | 3002 | 容器内网络下载工具 | high | `wget http://example.com` | 容器内 | 告警 | | |
| 4 | 3002 | 容器内网络下载工具 | high | `curl http://example.com` | 容器内 | 告警 | | |
| 5 | 3003 | 容器内系统配置修改 | high | `useradd testuser123` | 容器内 | 告警 | | |
| 6 | 3003 | 容器内系统配置修改 | high | `crontab -e` | 容器内 | 告警 | | |
| 7 | — | 反向验证 | — | `curl http://example.com` | 宿主机 | 无告警 | | |

---

## Step 5：清理与停止

```bash
# 1. Terminal B（容器内）：退出容器
exit

# 2. Terminal A：按 Ctrl+C 停止 Agent

# 3. 确认容器已清理（--rm 参数已自动清理）
docker ps -a | grep container_test
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| 容器内命令未触发告警 | 命令不存在 | eBPF 的 `sched_process_exec` 仅在 `execve()` 成功时触发；在容器内用 `which <命令>` 确认二进制文件存在，不存在则需先安装 |
| 容器内命令未触发告警 | 容器规则文件缺失 | 1) `ls /opt/cloudsec/agent/plugins/ebpf_base_detector/config/container_dangerous_commands.yaml` 确认文件存在；2) 在插件日志中搜索 `Container dangerous command rules loaded` 确认加载成功 |
| 容器内命令未触发告警 | mntns_id 判断不生效 | 在插件日志中搜索 `is_container`，确认容器进程的 `is_container=true`；如果 `root_mntns_id=0` 说明 eBPF 自动初始化未完成，先在宿主机执行任意命令触发首次 execve |
| Terminal A 看不到 `Container dangerous command detected` | 正常行为 | Terminal A 的 stderr 显示的是 standalone 输出 `dangerous command detected`（不区分容器/宿主机）；`Container dangerous command detected`（含 `container_id`）仅出现在插件日志 `/opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log` 中 |
| `container_id` 为空 | cgroup 格式不匹配 | 在容器内查看 `cat /proc/1/cgroup`，确认输出中包含 Docker 容器 ID（64 位十六进制字符串） |
| 宿主机命令也触发了容器告警 | mntns_id 计算错误 | 检查插件日志中宿主机进程的 `mntns_id` 和 `root_mntns_id` 是否相等；如不相等说明初始化有误 |
| Docker 不可用 | Docker 服务未启动 | `systemctl status docker`，确认服务运行中 |
| 容器内 wget/curl/crontab 不存在 | ubuntu:22.04 未预装 | `ubuntu:22.04` 基础镜像仅包含 `apt`/`apt-get`、`useradd` 等少量命令；需先在容器内执行 `apt-get update && apt-get install -y wget curl cron` 安装所需工具 |
| 容器内 pip/npm/gem 未触发告警 | 脚本类命令限制 | 这些命令本质是脚本文件，内核 `comm` 字段为解释器名（如 `python3`），检测器的反误报机制会过滤掉 `comm` 与规则期望命令名不一致的事件，属于已知限制 |
| 告警出现但 `rule_id` 不符 | 命令同时匹配多条规则 | 正常现象，检查命令是否匹配了其他规则的 pattern |
| 告警延迟超过 5 秒 | standalone 刷新间隔较长 | eBPF 事件本身无延迟，延迟来自用户态轮询；检查日志轮转配置 |
| 之前的告警在当前日志中消失 | 日志轮转 | 大量容器内操作（如 `apt-get install`）会产生海量文件事件导致日志快速轮转；用 `zcat /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector-*.log.gz \| grep "Container dangerous command detected"` 搜索已归档日志 |
| 其他已运行容器产生干扰告警 | 环境中存在带健康检查的容器 | 通过 `container_id` 字段区分不同容器的告警；可用 `docker ps` 查看当前运行的容器及其 ID |
