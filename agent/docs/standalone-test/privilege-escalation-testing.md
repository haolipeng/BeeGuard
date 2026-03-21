# 本地提权检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的本地提权检测功能（DataType 6006）：eBPF 通过 `kprobe/commit_creds` Hook 监控内核凭证变更，当进程的 uid/euid 从非 0 变为 0（提权到 root），且可执行文件不在内核白名单中时，产生告警。本文档选取 2 条代表性用例进行验证，覆盖自编译 SUID 程序和系统 SUID 程序复制两种场景。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | gcc 编译器 | `gcc --version` | gcc 已安装 |
| 7 | 普通用户账号 | `id <user>` | 存在 uid != 0 的普通用户，可用 `su - <user>` 切换 |

如果任一条件不满足，测试无法进行。

---

## Step 1：启动 Agent

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
- 该行出现 → 启动成功，eBPF 程序已加载，进入 Step 2
- 该行未出现 → 启动失败，检查内核版本和 BTF 支持
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
# 搜索提权告警（按 data_type 过滤）
grep '"data_type":6006' /tmp/ebpf_test.log

# 按 exe_path 精确搜索
grep '"exe_path":"/tmp/suid_wrapper"' /tmp/ebpf_test.log

# 使用 jq 格式化输出（需安装 jq）
cat /tmp/ebpf_test.log | jq 'select(.data_type==6006)'

# 实时监控新告警
tail -f /tmp/ebpf_test.log
```

---

## Step 2：执行测试用例

打开 **Terminal B**，逐条执行以下测试命令。每执行一条后，检查 `/tmp/ebpf_test.log` 是否出现对应告警。

### 告警日志格式

每条告警在 `/tmp/ebpf_test.log` 中以一行 JSON 输出：

```json
{"timestamp":1234567890,"data_type":6006,"pid":"PID","tgid":"TGID","ppid":"PPID","uid":"UID","comm":"进程名","exe_path":"可执行文件路径","old_uid":"旧UID","old_euid":"旧EUID","new_uid":"新UID","new_euid":"新EUID"}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. `/tmp/ebpf_test.log` 中出现 `"data_type":6006` 的 JSON 行
2. `"exe_path"` 与测试用例中执行的程序路径一致
3. `"old_uid"` 和 `"old_euid"` 非 `"0"`
4. `"new_uid"` 或 `"new_euid"` 为 `"0"`

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 `/tmp/ebpf_test.log` 无任何 `"data_type":6006` 的记录
- `"exe_path"` 与预期不一致

---

### 用例 1：SUID 程序测试（推荐）

**检测原理**：编译一个带 SUID 位的 C 程序，以普通用户执行时 euid 变为 0，触发 `commit_creds` 检测。

**测试命令**（Terminal B）：

```bash
# 1. 创建 C 源文件
cat > /tmp/suid_wrapper.c << 'EOF'
#include <unistd.h>
#include <stdio.h>
int main() {
    printf("uid=%d euid=%d\n", getuid(), geteuid());
    return 0;
}
EOF

# 2. 编译并设置 SUID 位
gcc -o /tmp/suid_wrapper /tmp/suid_wrapper.c
sudo chown root:root /tmp/suid_wrapper
sudo chmod 4755 /tmp/suid_wrapper

# 3. 以普通用户执行（将 <user> 替换为实际用户名）
su - <user> -c "/tmp/suid_wrapper"
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6006,"pid":"...","tgid":"...","ppid":"...","uid":"1000","comm":"suid_wrapper","exe_path":"/tmp/suid_wrapper","old_uid":"1000","old_euid":"1000","new_uid":"0","new_euid":"0"}
```

**验证命令**：

```bash
grep '"exe_path":"/tmp/suid_wrapper"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"exe_path":"/tmp/suid_wrapper"`，`"old_uid"` 非 `"0"`，`"new_euid":"0"`。

> 说明：`/tmp/suid_wrapper` 不在内核白名单中，因此 eBPF 会将该提权事件上报到用户态。`uid` 值取决于实际普通用户的 ID。

---

### 用例 2：已有 SUID 程序复制

**检测原理**：将系统已有的 SUID 程序复制到 `/tmp` 路径（不在白名单中），以普通用户执行时触发检测。

**测试命令**（Terminal B）：

```bash
# 1. 复制 passwd 并保留 SUID 位
sudo cp /usr/bin/passwd /tmp/test_passwd
sudo chmod 4755 /tmp/test_passwd

# 2. 以普通用户执行（将 <user> 替换为实际用户名）
su - <user> -c "/tmp/test_passwd --status <user>"
```

**预期日志**（`/tmp/ebpf_test.log`）：

```json
{"timestamp":...,"data_type":6006,"pid":"...","tgid":"...","ppid":"...","uid":"1000","comm":"test_passwd","exe_path":"/tmp/test_passwd","old_uid":"1000","old_euid":"1000","new_uid":"0","new_euid":"0"}
```

**验证命令**：

```bash
grep '"exe_path":"/tmp/test_passwd"' /tmp/ebpf_test.log
```

**PASS 判定**：上述命令有输出，且 JSON 中 `"exe_path":"/tmp/test_passwd"`，`"old_uid"` 非 `"0"`，`"new_euid":"0"`。

> 说明：原始 `/usr/bin/passwd` 在白名单中不会触发告警，但 `/tmp/test_passwd` 路径不在白名单中，因此会被检测。`--status` 参数仅查看密码状态，不会修改任何密码。

---

## Step 3：记录测试结果

| # | 用例名称 | 测试程序 | 预期 | 实际 | PASS/FAIL |
|---|----------|----------|------|------|-----------|
| 1 | SUID 程序测试 | `/tmp/suid_wrapper` | 告警 | | |
| 2 | 已有 SUID 程序复制 | `/tmp/test_passwd` | 告警 | | |

---

## Step 4：清理与停止

```bash
# 1. Terminal A：按 Ctrl+C 停止 Agent

# 2. Terminal B：清理测试残留文件
sudo rm -f /tmp/suid_wrapper /tmp/suid_wrapper.c /tmp/test_passwd

# 3. 清理输出文件（可选）
rm -f /tmp/ebpf_test.log
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动报 `failed to load eBPF` | 内核不支持或无 root 权限 | 1) `whoami` 确认 root；2) `uname -r` 确认 >= 5.4；3) `ls /sys/kernel/btf/vmlinux` 确认 BTF |
| `/tmp/ebpf_test.log` 未生成 | Agent 未成功启动或路径无写权限 | 1) 确认 Terminal A 中出现 `detection results will be written to: /tmp/ebpf_test.log`；2) `ls -la /tmp/ebpf_test.log` 确认文件存在 |
| SUID 程序执行后无告警 | 程序路径在白名单中 | 确认测试程序位于 `/tmp/` 等非白名单路径；使用 `ls -la` 确认 SUID 位已设置（权限显示为 `-rwsr-xr-x`） |
| su 切换用户失败 | 普通用户不存在或密码未知 | 1) `cat /etc/passwd` 确认用户存在；2) `passwd <user>` 重置密码；3) 或使用 `useradd -m testuser && passwd testuser` 创建测试用户 |
| 告警中 old_uid=0 | 以 root 身份执行了测试程序 | 必须以普通用户身份执行 SUID 程序，使用 `su - <user> -c "..."` 切换用户 |
| gcc 编译失败 | gcc 未安装 | `apt install gcc` 或 `yum install gcc` 安装编译器 |
| 告警延迟超过 5 秒 | standalone 刷新间隔较长 | 检查配置中 `flush_interval`（默认 1 秒）；eBPF 事件本身无延迟，延迟来自用户态轮询 |
