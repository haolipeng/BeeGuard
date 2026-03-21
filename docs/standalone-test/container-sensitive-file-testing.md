# 容器核心文件监控 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的容器核心文件监控功能（DataType 7004）：eBPF 在 `security_inode_create`/`security_inode_rename`/`security_inode_unlink` Hook 中捕获文件操作事件，通过 `mntns_id != root_mntns_id` 判断操作是否发生在容器内，用户态使用独立的容器敏感文件规则集匹配文件路径，命中规则即触发容器核心文件监控告警。主机敏感文件告警（DataType 6009）逻辑不受影响。本文档选取 2 条默认规则进行验证，覆盖容器内 passwd/shadow 修改和 DNS/hosts 篡改场景。

### 与宿主机敏感文件检测的区别

| 维度 | 宿主机敏感文件（6009） | 容器核心文件（7004） |
|------|----------------------|---------------------|
| 检测范围 | 所有进程（含容器） | 仅容器内进程 |
| 规则配置 | `sensitive_file_rules.yaml`（8 条规则） | `container_sensitive_file_rules.yaml`（2 条规则） |
| 容器判断 | 无 | `mntns_id != root_mntns_id` |
| 告警字段 | 无容器信息 | 包含 `container_id`、`container_name` |
| 规则 ID 范围 | 1001-1008 | 2001-2005 |

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
grep "Container sensitive file rules loaded" /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log
```

预期输出：

```
INFO  Container sensitive file rules loaded successfully  version=1.0 rules=2
```

**判定规则**：
- 两行日志都出现 → 启动成功，进入 Step 2
- `failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3
- `Failed to load container sensitive file rules` → 检查配置文件是否部署成功

### 日志位置

| 位置 | 内容 | 说明 |
|------|------|------|
| Terminal A (stderr) | Agent 主进程日志 | 用于确认启动状态 |
| `/opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log` | 插件日志，包含 `Container sensitive file operation detected` | **推荐用此日志验证**，包含 `container_id` |
| `/tmp/ebpf_test.log` | 检测结果 JSON 输出 | **主要验证位置** |

### 搜索技巧

```bash
# 搜索容器核心文件告警（按 data_type 过滤）
grep '"data_type":7004' /tmp/ebpf_test.log

# 按规则 ID 搜索
grep '"rule_id":"2001"' /tmp/ebpf_test.log
grep '"rule_id":"2005"' /tmp/ebpf_test.log

# 使用 jq 格式化输出（需安装 jq）
cat /tmp/ebpf_test.log | jq 'select(.data_type==7004)'

# 实时监控插件日志中的容器核心文件告警
tail -f /opt/cloudsec/agent/logs/plugins/ebpf_base_detector/ebpf_base_detector.log | grep "Container sensitive file"

# 实时监控新告警
tail -f /tmp/ebpf_test.log
```

---

## Step 2：启动测试容器

打开 **Terminal B**，启动一个 Docker 容器：

```bash
docker run -it --rm --name csf_test ubuntu:22.04 /bin/bash
```

> 说明：`ubuntu:22.04` 基础镜像已内置 `touch`、`bash` 等工具，无需额外安装。如无该镜像，可先 `docker pull ubuntu:22.04`。

---

## Step 3：执行测试用例

每个测试需要 2 个终端：**Terminal A**（Agent，已在 Step 1 启动）、**Terminal B**（容器内触发端）。每执行一条后，检查 `/tmp/ebpf_test.log` 是否出现对应告警。

### 告警日志格式

每条告警在 `/tmp/ebpf_test.log` 中以一行 JSON 输出：

```json
{"timestamp":1234567890,"data_type":7004,"pid":"PID","tgid":"TGID","ppid":"PPID","uid":"UID","comm":"进程名","exe_path":"路径","action":"操作","new_path":"文件路径","rule_id":"规则ID","rule_name":"规则名","severity":"级别","container_id":"容器ID","container_id_short":"短ID"}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. `/tmp/ebpf_test.log` 中出现 `"data_type":7004` 的 JSON 行
2. `"rule_id"` 与预期规则 ID 一致
3. `"new_path"` 包含预期的文件路径
4. `"container_id"` 非空（64 位十六进制字符串）

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 `/tmp/ebpf_test.log` 无任何 `"data_type":7004` 的记录
- `"rule_id"` 或 `"new_path"` 与预期不一致
- `"container_id"` 为空

---

### 用例 1：CSF001 — 容器内 passwd 文件创建（规则 2001）

**检测原理**：在容器内创建 `/etc/passwd` 相关文件，触发 `security_inode_create` Hook，eBPF 采集事件后用户态匹配规则 2001（`^/etc/passwd`）。

**测试命令**：

Terminal B（容器内触发端）：

```bash
touch /etc/passwd_test
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":7004,"comm":"touch","exe_path":"/usr/bin/touch","action":"create","new_path":"/etc/passwd_test","rule_id":"2001","rule_name":"容器内 passwd/shadow 文件修改","severity":"critical","container_id":"..."}
```

**验证命令**：

```bash
grep '"data_type":7004' /tmp/ebpf_test.log | grep '"rule_id":"2001"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"new_path"` 包含 `/etc/passwd_test`，`"severity":"critical"`，`"container_id"` 非空。

> 测试完成后清理：`rm -f /etc/passwd_test`

---

### 用例 2：CSF002 — 容器内 resolv.conf 修改（规则 2005）

**检测原理**：在容器内向 `/etc/resolv.conf` 写入内容会触发文件重建（部分编辑器会删除旧文件再创建新文件），触发 `security_inode_create` Hook，匹配规则 2005（精确匹配 `/etc/resolv.conf`）。

**测试命令**：

Terminal B（容器内触发端）：

```bash
cp /etc/resolv.conf /etc/resolv.conf.bak
cp /etc/resolv.conf.bak /etc/resolv.conf
```

> 说明：由于 `security_inode_create` 监控的是文件创建操作，直接 `echo >> /etc/resolv.conf` 追加写入不会触发（只触发 write 系统调用）。使用 `cp` 会创建新文件覆盖旧文件，从而触发创建事件。如果 cp 未触发，也可尝试先删除再创建：`rm /etc/resolv.conf && touch /etc/resolv.conf`

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":7004,"comm":"cp","exe_path":"/usr/bin/cp","action":"create","new_path":"/etc/resolv.conf","rule_id":"2005","rule_name":"容器内 DNS/hosts 篡改","severity":"high","container_id":"..."}
```

**验证命令**：

```bash
grep '"data_type":7004' /tmp/ebpf_test.log | grep '"rule_id":"2005"'
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"new_path"` 为 `/etc/resolv.conf`，`"severity":"high"`，`"container_id"` 非空。

---

### 用例 3：CSF003 — 容器内非敏感文件（无告警）

**检测原理**：在容器内创建不匹配任何规则的文件，不应触发 7004 告警。

**测试命令**：

Terminal B（容器内触发端）：

```bash
touch /tmp/safe_test_file
```

**验证命令**：

```bash
grep '"data_type":7004' /tmp/ebpf_test.log | grep 'safe_test_file'
```

**PASS 判定**：上述命令**无输出**（不应触发容器核心文件告警）。

> 测试完成后清理：`rm -f /tmp/safe_test_file`

---

### 用例 4（反向验证）：宿主机操作敏感文件触发 6009 而非 7004

退出容器，在宿主机上执行：

Terminal B（宿主机触发端）：

```bash
touch /etc/passwd_host_test
```

**预期**：
- `/tmp/ebpf_test.log` 中出现 `"data_type":6009` 的告警（主机侧敏感文件）
- **不应**出现 `"data_type":7004` 的告警

**验证命令**：

```bash
# 应有输出（主机侧 6009 告警）
grep '"data_type":6009' /tmp/ebpf_test.log | grep 'passwd_host_test'

# 不应有输出（容器侧 7004 告警）
grep '"data_type":7004' /tmp/ebpf_test.log | grep 'passwd_host_test'
```

**PASS 判定**：第一条命令有输出，第二条命令无输出。

> 测试完成后清理：`rm -f /etc/passwd_host_test`

---

## Step 4：记录测试结果

| # | 用例 ID | 测试场景 | 执行环境 | 触发文件 | 预期规则 ID | 预期 data_type | 实际 | PASS/FAIL |
|---|---------|----------|----------|----------|------------|---------------|------|-----------|
| 1 | CSF001 | passwd 文件创建 | 容器内 | /etc/passwd_test | 2001 | 7004 | | |
| 2 | CSF002 | resolv.conf 修改 | 容器内 | /etc/resolv.conf | 2005 | 7004 | | |
| 3 | CSF003 | 非敏感文件（无告警） | 容器内 | /tmp/safe_test_file | 无 | 无 | | |
| 4 | CSF004 | 反向验证（宿主机） | 宿主机 | /etc/passwd_host_test | 1008 | 6009 | | |

---

## Step 5：清理与停止

```bash
# 1. Terminal B（容器内）：退出容器
exit

# 2. Terminal A：按 Ctrl+C 停止 Agent

# 3. 确认容器已清理（--rm 参数已自动清理）
docker ps -a | grep csf_test

# 4. 清理宿主机测试文件
rm -f /etc/passwd_host_test

# 5. 清理输出文件（可选）
rm -f /tmp/ebpf_test.log
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动报 `failed to load eBPF` | 内核不支持或无 root 权限 | 1) `whoami` 确认 root；2) `uname -r` 确认 >= 5.4；3) `ls /sys/kernel/btf/vmlinux` 确认 BTF |
| 插件日志中无 `Container sensitive file rules loaded` | 配置文件未部署 | `ls /opt/cloudsec/agent/plugins/ebpf_base_detector/config/container_sensitive_file_rules.yaml` 确认文件存在 |
| `/tmp/ebpf_test.log` 未生成 | Agent 未成功启动或路径无写权限 | 确认 Terminal A 中出现 `detection results will be written to: /tmp/ebpf_test.log` |
| 容器内 touch 操作无 7004 告警 | mntns_id 未正确填充 | 检查插件日志中文件事件的 `mntns_id` 和 `root_mntns_id` 是否不相等 |
| 出现 6009 而非 7004 告警 | 容器检测未生效或规则未加载 | 1) 检查插件日志确认容器敏感文件规则已加载；2) 检查容器内进程的 mntns_id |
| 7004 告警中 `container_id` 为空 | cgroup 格式不匹配 | 在容器内查看 `cat /proc/1/cgroup`，确认输出中包含 Docker 容器 ID |
| 宿主机操作也触发了 7004 | mntns_id 判断异常 | 检查插件日志中宿主机进程的 `mntns_id` 和 `root_mntns_id` 是否相等 |
| `cp /etc/resolv.conf` 未触发告警 | 文件系统层面未触发 create | 尝试 `rm /etc/resolv.conf && touch /etc/resolv.conf` 替代 |
| 告警延迟超过 5 秒 | standalone 刷新间隔 | eBPF 事件本身无延迟，延迟来自用户态轮询 |
| Docker 不可用 | Docker 服务未启动 | `systemctl status docker`，确认服务运行中 |
