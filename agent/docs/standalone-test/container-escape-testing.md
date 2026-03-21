# 容器逃逸检测（mount 设备） — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的容器逃逸检测功能（DataType 7002）：eBPF 在 `sys_exit` raw tracepoint 中捕获 mount 系统调用（syscall 165）的返回事件，通过 `mntns_id != root_mntns_id` 判断是否在容器内，用户态检测挂载源是否为宿主机块设备（`/dev/sd*`、`/dev/vd*`、`/dev/nvme*` 等），匹配成功时产生 critical 级别告警。

### 检测原理

容器逃逸的经典手法之一是在特权容器内挂载宿主机的块设备（如 `/dev/sda1`），从而获取宿主机文件系统的完整访问权限。本检测通过以下条件链判定：

1. **事件来源**：mount 系统调用成功返回（retval == 0）
2. **容器判断**：`mntns_id != root_mntns_id`（进程不在宿主机命名空间）
3. **设备匹配**：挂载源以 `/dev/sd`、`/dev/vd`、`/dev/nvme`、`/dev/xvd`、`/dev/hd` 开头

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | Docker | `docker version` | Docker 已安装且运行中 |
| 7 | 块设备 | `lsblk` | 至少有一个块设备（如 `/dev/sda1`） |

如果任一条件不满足，测试无法进行。

> **安全警告**：本测试涉及在容器内挂载宿主机磁盘设备，**请在隔离的测试环境中操作**，避免在生产环境执行。挂载操作本身是只读安全的（使用 `-o ro` 参数），但仍需谨慎。

---

## Step 1：确认宿主机块设备

```bash
lsblk -o NAME,TYPE,SIZE,MOUNTPOINT | grep -E "disk|part"
```

记录一个已有的块设备分区，例如 `/dev/sda1` 或 `/dev/vda1`。后续测试用例中以 `DEV` 代指该设备路径。

**示例输出**：

```
sda      disk  50G
├─sda1   part  49G /
├─sda2   part   1G [SWAP]
```

此例中 `DEV=/dev/sda1`。

---

## Step 2：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**看到以下日志行：

```
INFO  Container escape detector initialized
```

### 搜索技巧

```bash
# 只显示容器逃逸告警
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | grep "Container escape detected"

# 保存全部输出
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | tee /tmp/escape_test.log
```

---

## Step 3：启动特权测试容器

打开 **Terminal B**，启动一个 **特权容器**：

```bash
docker run -it --rm --privileged --name escape_test ubuntu:22.04 /bin/bash
```

> **关键**：必须使用 `--privileged` 参数，否则容器内无权执行 mount 系统调用，mount 会返回 `EPERM` 错误（retval != 0），eBPF 会跳过该事件。

---

## Step 4：执行测试用例

### 告警日志格式

```
{时间戳}  WARN  Container escape detected (mount device)  rule_name=container_escape_mount_device  severity=critical  dev_name={设备}  dir_name={挂载点}  pid={PID}  comm={进程名}  container_id={容器ID}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. Terminal A 出现包含 `Container escape detected` 的日志行
2. `rule_name=container_escape_mount_device`
3. `severity=critical`
4. `dev_name` 与挂载的设备路径一致
5. `container_id` 非空

**FAIL** 条件（任一满足）：
- 执行 mount 后 5 秒内 Terminal A 无任何 `Container escape detected` 输出
- `dev_name` 与实际挂载设备不一致

---

### 用例 1：容器内挂载宿主机磁盘分区

**测试命令**（Terminal B，容器内）：

```bash
# 将 /dev/sda1 替换为 Step 1 中确认的实际设备
mkdir -p /mnt/escape_test
mount -o ro /dev/sda1 /mnt/escape_test
```

> 使用 `-o ro`（只读挂载）降低风险。

**预期日志**（Terminal A）：

```
WARN  Container escape detected (mount device)  rule_name=container_escape_mount_device  severity=critical  dev_name=/dev/sda1  dir_name=/mnt/escape_test  pid=...  comm=mount  container_id=...
```

**PASS 判定**：出现 `Container escape detected`，`severity=critical`，`dev_name=/dev/sda1`。

**清理**：

```bash
umount /mnt/escape_test
rmdir /mnt/escape_test
```

---

### 用例 2：容器内挂载不同设备路径格式

如果宿主机有 VirtIO 设备（云服务器常见），测试 `/dev/vda*` 格式：

```bash
mkdir -p /mnt/escape_test2
mount -o ro /dev/vda1 /mnt/escape_test2
```

**PASS 判定**：同用例 1，`dev_name` 匹配 `/dev/vda1`。

**清理**：

```bash
umount /mnt/escape_test2 2>/dev/null; rmdir /mnt/escape_test2
```

---

### 用例 3（反向验证）：容器内挂载 tmpfs 不触发告警

**测试命令**（Terminal B，容器内）：

```bash
mkdir -p /mnt/tmpfs_test
mount -t tmpfs tmpfs /mnt/tmpfs_test
```

**预期**：Terminal A **不应**出现 `Container escape detected` 日志。

> 说明：tmpfs 挂载的设备名为 `tmpfs`，不以 `/dev/sd*` 等块设备前缀开头，不满足逃逸检测条件。

**PASS 判定**：5 秒内 Terminal A 无 `Container escape detected` 输出。

**清理**：

```bash
umount /mnt/tmpfs_test
rmdir /mnt/tmpfs_test
```

---

### 用例 4（反向验证）：宿主机 mount 不触发容器逃逸告警

退出容器，在 **宿主机** Terminal B 上执行：

```bash
mkdir -p /tmp/host_mount_test
mount -t tmpfs tmpfs /tmp/host_mount_test
umount /tmp/host_mount_test
rmdir /tmp/host_mount_test
```

**预期**：Terminal A **不应**出现 `Container escape detected` 日志。

> 说明：宿主机进程的 `mntns_id == root_mntns_id`，不满足容器判断条件。

**PASS 判定**：5 秒内 Terminal A 无 `Container escape detected` 输出。

---

### 用例 5（反向验证）：非特权容器 mount 失败不触发告警

```bash
# 启动一个非特权容器
docker run -it --rm --name escape_test_noprivs ubuntu:22.04 /bin/bash
```

在容器内执行：

```bash
mkdir -p /mnt/noprivs_test
mount /dev/sda1 /mnt/noprivs_test 2>&1; echo "exit code: $?"
```

**预期**：
- mount 命令失败，输出 `mount: permission denied` 或 `Operation not permitted`
- Terminal A **不应**出现 `Container escape detected` 日志

> 说明：非特权容器的 mount 系统调用返回非零值（retval != 0），eBPF 在 sys_exit 中直接跳过。

**PASS 判定**：mount 失败 + 无逃逸告警。

**清理**：

```bash
exit  # 退出容器
```

---

## Step 5：记录测试结果

| # | 测试用例 | 执行环境 | 测试命令 | 预期 | 实际 | PASS/FAIL |
|---|---------|----------|----------|------|------|-----------|
| 1 | 挂载宿主机磁盘 | 特权容器 | `mount -o ro /dev/sda1 /mnt/escape_test` | critical 告警 | | |
| 2 | 挂载 VirtIO 设备 | 特权容器 | `mount -o ro /dev/vda1 /mnt/escape_test2` | critical 告警 | | |
| 3 | 挂载 tmpfs（反向） | 特权容器 | `mount -t tmpfs tmpfs /mnt/tmpfs_test` | 无告警 | | |
| 4 | 宿主机 mount（反向） | 宿主机 | `mount -t tmpfs tmpfs /tmp/host_mount_test` | 无告警 | | |
| 5 | 非特权容器 mount（反向） | 普通容器 | `mount /dev/sda1 /mnt/noprivs_test` | mount 失败 + 无告警 | | |

---

## Step 6：清理与停止

```bash
# 1. Terminal B：退出所有容器
exit

# 2. Terminal A：按 Ctrl+C 停止 Agent

# 3. 确认容器已清理（--rm 参数已自动清理）
docker ps -a | grep escape_test
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| mount 在容器内报 `Operation not permitted` | 容器未使用 `--privileged` | 确认 `docker run` 命令包含 `--privileged` 参数 |
| mount 成功但无告警 | 设备路径不匹配块设备前缀 | 检查设备路径是否以 `/dev/sd`、`/dev/vd`、`/dev/nvme`、`/dev/xvd`、`/dev/hd` 开头 |
| mount 成功但无告警 | mntns_id 判断失败 | 搜索日志中的 `Mount event`（INFO 级别），检查 `is_container` 字段是否为 `true` |
| `root_mntns_id=0` | eBPF 自动初始化未完成 | 先在宿主机执行任意命令（如 `ls`）触发首次 execve 事件，使 eBPF 完成 root_mntns_id 初始化 |
| mount 事件完全未出现 | syscall 165 未被捕获 | 检查 eBPF 程序是否正确编译，确认 `hids.bpf.c` 中包含 `syscall_nr != 165` 的过滤逻辑 |
| `container_id` 为空 | cgroup 格式不匹配 | 在特权容器内查看 `cat /proc/1/cgroup`，确认包含 Docker 容器 ID |
| Docker 不可用 | Docker 服务未启动 | `systemctl status docker`，确认服务运行中 |
| 宿主机无块设备 | 虚拟化环境差异 | 使用 `lsblk` 查看可用设备；云服务器可能使用 `/dev/vda*` 而非 `/dev/sda*` |
