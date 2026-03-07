# 容器高危命令检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的容器高危命令检测功能（DataType 7001）：eBPF 在 `sched_process_exec` Hook 中捕获进程执行事件，通过 `mntns_id != root_mntns_id` 判断进程是否运行在容器内，用户态将容器内的命令行参数与 `container_dangerous_commands.yaml` 中的规则进行匹配，匹配成功时产生告警。本文档选取 3 条规则的 7 个测试用例进行验证，覆盖 regex、prefix 两种匹配方式和 medium、high 两种严重程度。

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
| 5 | 编译环境 | `go version` | Go 已安装 |
| 6 | Docker | `docker version` | Docker 已安装且运行中 |

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
# 确认二进制文件存在
ls -la /opt/cloudsec/bin/agent /opt/cloudsec/plugins/ebpf_base_detector/ebpf_base_detector

# 确认容器高危命令规则文件存在
ls -la /opt/cloudsec/plugins/ebpf_base_detector/config/container_dangerous_commands.yaml
```

两个二进制文件和规则文件都存在即成功。

---

## Step 2：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**看到以下日志行：

```
INFO  Container dangerous command rules loaded successfully  version=1.0  rules=3
```

**判定规则**：
- `rules=3` → 启动成功，3 条容器规则全部加载，进入 Step 3
- 出现 `Failed to load container dangerous command rules` → 规则文件缺失或格式错误，检查 `container_dangerous_commands.yaml` 是否在 `/opt/cloudsec/plugins/ebpf_base_detector/config/` 目录下
- `failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3

同时还应看到：

```
INFO  Container escape detector initialized
INFO  Container metadata cache initialized
```

### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stderr) | 实时输出，**主要观察位置** |
| `/opt/cloudsec/logs/ebpf_base_detector.log` | 同内容持久化文件，可用 grep 搜索 |

### 搜索技巧

```bash
# 方式一：只显示容器高危命令告警
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | grep "Container dangerous command detected"

# 方式二：保存全部输出到文件，在另一个终端搜索
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | tee /tmp/container_test.log
# 另一个终端
grep "Container dangerous command" /tmp/container_test.log

# 方式三：按规则 ID 精确搜索
grep "rule_id=3001" /tmp/container_test.log
```

---

## Step 3：启动测试容器

打开 **Terminal B**，启动一个 Docker 容器：

```bash
docker run -it --rm --name container_test ubuntu:22.04 /bin/bash
```

> 说明：使用 `ubuntu:22.04` 镜像，容器内有 `apt`、`useradd` 等命令。如无该镜像，可先 `docker pull ubuntu:22.04`。

---

## Step 4：执行测试用例

在 **Terminal B**（容器内）逐条执行以下测试命令。每执行一条后，回到 Terminal A 查看是否出现对应告警。

### 告警日志格式

每条告警在 Terminal A 中以一行结构化日志输出：

```
{时间戳}  WARN  Container dangerous command detected  rule_id={ID}  rule_name={名称}  severity={级别}  uid={UID}  comm={进程名}  args={命令参数}  container_id={容器ID}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. Terminal A 出现包含 `Container dangerous command detected` 的日志行
2. `rule_id` 与测试用例的规则 ID 一致
3. `comm` 与执行的命令名一致
4. `container_id` 非空（64 位十六进制字符串）

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 Terminal A 无任何 `Container dangerous command detected` 输出
- `rule_id` 与预期不一致
- `container_id` 为空

---

### 用例 1：规则 3001 — 容器内包管理器安装（medium）

**匹配方式**：regex，匹配 `apt(-get)?\s+install`

**测试命令**（Terminal B，容器内）：

```bash
apt-get install -y curl 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
WARN  Container dangerous command detected  rule_id=3001  rule_name=容器内包管理器安装  severity=medium  uid=0  comm=apt-get  args=install -y curl  container_id=...
```

**PASS 判定**：出现 `Container dangerous command detected`，且 `rule_id=3001`，`comm=apt-get`。

> 说明：即使 `apt-get install` 因网络原因失败，eBPF 在 execve 阶段已捕获，不影响检测。

---

### 用例 2：规则 3001 — 容器内 pip 安装（medium）

**匹配方式**：regex，匹配 `pip3?\s+install`

**测试命令**（Terminal B，容器内）：

```bash
pip install requests 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
WARN  Container dangerous command detected  rule_id=3001  rule_name=容器内包管理器安装  severity=medium  comm=pip  args=install requests  container_id=...
```

**PASS 判定**：`rule_id=3001`，`comm=pip`。

> 说明：容器内可能未安装 pip，命令报错不影响 eBPF 捕获。

---

### 用例 3：规则 3002 — 容器内 wget 下载（high）

**匹配方式**：prefix，进程名以 `wget` 开头

**测试命令**（Terminal B，容器内）：

```bash
wget http://example.com 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
WARN  Container dangerous command detected  rule_id=3002  rule_name=容器内网络下载工具  severity=high  comm=wget  container_id=...
```

**PASS 判定**：`rule_id=3002`，`comm=wget`。

---

### 用例 4：规则 3002 — 容器内 curl 请求（high）

**匹配方式**：prefix，进程名以 `curl` 开头

**测试命令**（Terminal B，容器内）：

```bash
curl http://example.com 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
WARN  Container dangerous command detected  rule_id=3002  rule_name=容器内网络下载工具  severity=high  comm=curl  container_id=...
```

**PASS 判定**：`rule_id=3002`，`comm=curl`。

---

### 用例 5：规则 3003 — 容器内用户创建（high）

**匹配方式**：regex，匹配 `useradd\s+`

**测试命令**（Terminal B，容器内）：

```bash
useradd testuser123 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
WARN  Container dangerous command detected  rule_id=3003  rule_name=容器内系统配置修改  severity=high  comm=useradd  args=testuser123  container_id=...
```

**PASS 判定**：`rule_id=3003`，`comm=useradd`。

---

### 用例 6：规则 3003 — 容器内 crontab 编辑（high）

**匹配方式**：regex，匹配 `crontab\s+-e`

**测试命令**（Terminal B，容器内）：

```bash
crontab -e 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
WARN  Container dangerous command detected  rule_id=3003  rule_name=容器内系统配置修改  severity=high  comm=crontab  args=-e  container_id=...
```

**PASS 判定**：`rule_id=3003`，`comm=crontab`。

---

### 用例 7（反向验证）：宿主机同命令不触发容器告警

退出容器，在 **宿主机** Terminal B 上执行：

```bash
curl http://example.com 2>/dev/null; true
```

**预期**：Terminal A **不应**出现 `Container dangerous command detected` 日志。

> 说明：宿主机上的 curl 可能触发宿主机的 `Dangerous command detected`（DataType 6003，取决于 `dangerous_commands.yaml` 配置），但**不应**触发容器的 `Container dangerous command detected`（DataType 7001）。

**PASS 判定**：5 秒内 Terminal A 无 `Container dangerous command detected` 输出。

---

## Step 5：记录测试结果

| # | 规则 ID | 规则名称 | 严重程度 | 测试命令 | 执行环境 | 预期 | 实际 | PASS/FAIL |
|---|---------|----------|----------|----------|----------|------|------|-----------|
| 1 | 3001 | 容器内包管理器安装 | medium | `apt-get install -y curl` | 容器内 | 告警 | | |
| 2 | 3001 | 容器内包管理器安装 | medium | `pip install requests` | 容器内 | 告警 | | |
| 3 | 3002 | 容器内网络下载工具 | high | `wget http://example.com` | 容器内 | 告警 | | |
| 4 | 3002 | 容器内网络下载工具 | high | `curl http://example.com` | 容器内 | 告警 | | |
| 5 | 3003 | 容器内系统配置修改 | high | `useradd testuser123` | 容器内 | 告警 | | |
| 6 | 3003 | 容器内系统配置修改 | high | `crontab -e` | 容器内 | 告警 | | |
| 7 | — | 反向验证 | — | `curl http://example.com` | 宿主机 | 无告警 | | |

---

## Step 6：清理与停止

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
| 容器内命令未触发告警 | 容器规则文件缺失 | 1) `ls /opt/cloudsec/plugins/ebpf_base_detector/config/container_dangerous_commands.yaml` 确认文件存在；2) 检查启动日志是否有 `Container dangerous command rules loaded` |
| 容器内命令未触发告警 | mntns_id 判断不生效 | 在日志中搜索 `is_container`，确认容器进程的 `is_container=true`；如果 `root_mntns_id=0` 说明 eBPF 自动初始化未完成，先在宿主机执行任意命令触发首次 execve |
| `container_id` 为空 | cgroup 格式不匹配 | 在容器内查看 `cat /proc/1/cgroup`，确认输出中包含 Docker 容器 ID（64 位十六进制字符串） |
| 宿主机命令也触发了容器告警 | mntns_id ���算错误 | 检查日志中宿主机进程的 `mntns_id` 和 `root_mntns_id` 是否相等；如不相等说明初始化有误 |
| Docker 不可用 | Docker 服务未启动 | `systemctl status docker`，确认服务运行中 |
| 容器内 apt/wget 不存在 | 镜像过于精简 | 换用 `ubuntu:22.04` 或 `debian:12` 镜像，这些镜像自带基本命令 |
| 告警出现但 `rule_id` 不符 | 命令同时匹配多条规则 | 正常现象，检查命令是否匹配了其他规则的 pattern |
| 告警延迟超过 5 秒 | standalone 刷新间隔较长 | eBPF 事件本身无延迟，延迟来自用户态轮询；检查日志轮转配置 |
