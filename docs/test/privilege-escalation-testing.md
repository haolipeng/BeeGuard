# 本地提权检测测试指南

本文档描述 ebpf_base_detector 插件本地提权检测功能的测试方法，面向技术研发人员。

---

## 1. 概述

ebpf_base_detector 插件通过 eBPF Hook `kprobe/commit_creds` 检测进程的 uid/euid 提权行为。

**检测原理:**

- **Hook 点**: `kprobe/commit_creds`（内核凭证变更函数）
- **检测条件**: 原 uid 和 euid 都非 0，新 uid 或 euid 为 0（提权到 root）
- **白名单机制**: 内核层过滤 sudo、su、pkexec 等合法提权程序，基于可执行文件绝对路径匹配，使用 Murmur OAAT64 哈希算法快速比对，不产生用户态事件
- **告警类型**: DataType 60

**关键源文件:**

| 文件 | 说明 |
|------|------|
| `business_plugins/ebpf_base_detector/ebpf/bpf/hids.bpf.c` | eBPF 内核代码（commit_creds hook） |
| `business_plugins/ebpf_base_detector/main.go` | 事件处理逻辑 |
| `business_plugins/ebpf_base_detector/config/privilege_escalation_whitelist.yaml` | 白名单配置 |

---

## 2. 环境要求

| 项目 | 要求 |
|------|------|
| 内核版本 | >= 5.x |
| BTF 支持 | `/sys/kernel/btf/vmlinux` 存在 |
| 编译依赖 | clang、llvm、libbpf-dev、linux-headers |
| 运行权限 | root |

**环境检查:**

```bash
# 检查 BTF 支持
ls /sys/kernel/btf/vmlinux

# 检查内核版本
cat /proc/version
```

---

## 3. 编译

```bash
# 编译 ebpf_base_detector 插件
make build-driver
```

编译输出路径: `build/` 目录。

---

## 4. 部署

```bash
# 部署 ebpf_base_detector 插件
make deploy-driver

# 确认部署
ls -la /opt/cloudsec/plugins/ebpf_base_detector/
ls -la /opt/cloudsec/plugins/ebpf_base_detector/config/
```

**部署目录结构:**

```
/opt/cloudsec/
├── agent                          # agent 主程序
└── plugins/
    └── ebpf_base_detector/
        ├── ebpf_base_detector                 # 插件二进制
        └── config/
            ├── dangerous_commands.yaml
            └── privilege_escalation_whitelist.yaml   # 白名单配置
```

---

## 5. 启动 Agent

Standalone 模式允许不连接 gRPC Server 进行本地测试，检测结果输出到日志或文件。

### 命令行参数

| 参数 | 说明 | 示例 |
|------|------|------|
| `-config` | 配置文件路径 | `-config=agent-standalone.yaml` |
| `-standalone` | 启用 standalone 模式 | `-standalone` |
| `-output` | 输出方式 (stderr/文件路径) | `-output=stderr` |
| `-plugins` | 加载的插件列表 | `-plugins=ebpf_base_detector` |
| `-test` | 测试模式（固定 agent ID） | `-test` |

### 启动方式一：命令行参数（推荐）

```bash
# Agent 运行日志输出到 /opt/cloudsec/logs/agent.log
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test
```

### 启动方式二：配置文件

创建 `agent-standalone.yaml`:

```yaml
# Agent Standalone 模式配置
working_directory: "/opt/cloudsec/data/agent"
plugins_directory: "/opt/cloudsec/plugins"

standalone:
  enabled: true
  output: "stderr"                   # "stderr" 或文件路径（如 "/tmp/results.json"）
  flush_interval: 1                  # 刷新间隔（秒）
  plugins:
    - ebpf_base_detector                         # 仅加载 ebpf_base_detector 插件
```

```bash
cd /opt/cloudsec
sudo ./bin/agent -config=agent-standalone.yaml -test
```

### 保存日志

```bash
cd /opt/cloudsec
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | tee ebpf_base_detector.log
```

---

## 6. 触发本地提权

由于本地提权涉及到特权操作，需要在授权的测试环境中进行。以下提供多种测试方法。

### 方法一：SUID 程序测试（推荐）

创建一个带 SUID 位的测试程序：

```bash
# 1. 创建 C 包装器
cat > /tmp/suid_wrapper.c << 'EOF'
#include <unistd.h>
#include <stdio.h>
int main() {
    printf("Before: uid=%d euid=%d\n", getuid(), geteuid());
    setuid(0);
    setgid(0);
    printf("After: uid=%d euid=%d\n", getuid(), geteuid());
    execl("/bin/bash", "bash", "-c", "id", NULL);
    return 0;
}
EOF

# 2. 编译并设置 SUID 位（需要 root）
gcc -o /tmp/suid_wrapper /tmp/suid_wrapper.c
sudo chown root:root /tmp/suid_wrapper
sudo chmod 4755 /tmp/suid_wrapper

# 3. 以普通用户身份运行
su - haolipeng -c "/tmp/suid_wrapper"
```

**预期结果:**
- Agent 日志中出现 `WARN Privilege escalation detected`
- `exe_path=/tmp/suid_wrapper`
- `old_uid=1000`（你的用户 ID），`new_uid=0`

---

### 方法二：已有 SUID 程序

使用系统中已有的 SUID 程序，复制一份到不在白名单的路径：

```bash
# 查找系统中的 SUID 程序
find /usr/bin /bin /usr/sbin -perm -4000 -type f 2>/dev/null

# 创建一个不在白名单中的 SUID 副本
sudo cp /usr/bin/passwd /tmp/test_passwd
sudo chmod 4755 /tmp/test_passwd
/tmp/test_passwd
```

**预期结果:**
- `/tmp/test_passwd` 不在白名单，应触发 `WARN Privilege escalation detected`

---

### 方法三：Python/Perl 命令

在某些特殊配置的测试环境中可以尝试：

```bash
# Python 调用 setuid（通常会失败，除非在特殊环境）
python3 -c "import os; os.setuid(0); os.system('id')"

# Perl 调用 setuid（通常会失败）
perl -e 'use POSIX; POSIX::setuid(0); system("id")'
```

> **注意:** 这些命令在正常环境下会因权限不足而失败，只能在特殊测试环境中触发检测。

---

### 方法四：白名单验证（反向验证）

测试合法提权程序**不应**触发告警：

```bash
# 使用 sudo（应该不触发告警）
sudo id

# 使用 su（应该不触发告警）
su - root -c "id"
```

**预期结果:** 日志中**不应该**出现 `WARN Privilege escalation detected`，因为 sudo 和 su 在白名单中。

---

### 方法五：完整测试程序

如果以上方法都无法触发，使用以下测试程序：

```c
/* privilege_escalation_test.c
 * 用途：测试本地提权检测功能
 * 编译：gcc -o privesc_test privilege_escalation_test.c
 * 运行：需要在具有 SUID 漏洞的环境中测试
 *
 * 注意：此程序仅用于测试，在正常环境中会因权限不足而失败
 */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>

int main() {
    printf("Current UID: %d, EUID: %d\n", getuid(), geteuid());
    printf("Attempting privilege escalation...\n");

    // 方法 1: 直接调用 setuid (仅在特定环境有效)
    if (setuid(0) == 0) {
        printf("Successfully escalated to root via setuid\n");
        printf("  New UID: %d, EUID: %d\n", getuid(), geteuid());
        system("id");
        return 0;
    } else {
        perror("setuid failed (expected in normal environment)");
    }

    // 方法 2: 测试 setreuid
    if (setreuid(0, 0) == 0) {
        printf("Successfully escalated to root via setreuid\n");
        return 0;
    } else {
        perror("setreuid failed (expected in normal environment)");
    }

    printf("\nTest requires a vulnerable environment (e.g., SUID exploit)\n");
    printf("In normal conditions, all privilege escalation attempts will fail.\n");

    return 1;
}
```

**编译和运行:**

```bash
gcc -o /tmp/privesc_test privilege_escalation_test.c
/tmp/privesc_test
```

> **注意:**
> - 在正常环境中，此程序会因权限不足而失败，这是预期行为
> - 要触发提权检测，需要在具有 SUID 漏洞的测试环境中运行
> - 或者使用已知的 CVE 漏洞利用工具（仅限授权的安全测试环境）

---

## 7. 验证检测结果

### 日志查看位置

检测结果的输出位置取决于启动时的 `-output` 参数：

| 输出方式 | 查看方法 |
|----------|----------|
| `-output=stderr`（默认） | 直接查看 agent 启动终端的 stderr 输出 |
| `-output=stderr` + tee 重定向 | 查看保存的日志文件：`grep "Privilege escalation" ebpf_base_detector.log` |
| `-output=<文件路径>` | 查看指定的输出文件：`cat /tmp/results.json` |

如果 agent 在后台运行或日志已滚动，可用以下方式回溯：

```bash
# 从 tee 保存的日志中搜索
grep "Privilege escalation" ebpf_base_detector.log

# 从文件输出中搜索（当 -output 指定为文件路径时）
cat /tmp/results.json | grep -i "privesc\|privilege"
```

### 成功判断标准

1. 日志中出现 `WARN  Privilege escalation detected`
2. `old_uid` 和 `old_euid` 非 0
3. `new_uid` 或 `new_euid` 为 0
4. `exe_path` 不在白名单中

### 日志示例

```
WARN  Privilege escalation detected
    pid=12345  tgid=12345  ppid=12344  comm=privesc_test
    exe_path=/tmp/privesc_test  uid=1000
    old_uid=1000  old_euid=1000  new_uid=0  new_euid=0
```

### 查看内核调试日志

```bash
# 查看 eBPF 内核日志（需要 root 权限）
sudo cat /sys/kernel/debug/tracing/trace_pipe | grep hids
```

**内核日志示例:**

```
hids: commit_creds pid=12345 tgid=12345 ppid=12344
hids: commit_creds uid=1000 old_uid=1000 old_euid=1000
hids: commit_creds new_uid=0 new_euid=0
```

---

## 8. 白名单配置

### 配置文件路径

`/opt/cloudsec/plugins/ebpf_base_detector/config/privilege_escalation_whitelist.yaml`

### 默认白名单内容

```yaml
version: "1.0"
description: "Trusted executables whitelist for privilege escalation filtering"

# 可信任的可执行文件（绝对路径）
# 这些进程的提权行为不会触发告警
# 要求：每个条目必须是绝对路径，以 "/" 开头
trusted_executables:
  - "/usr/bin/sudo"
  - "/usr/bin/su"
  - "/usr/bin/pkexec"
  - "/usr/lib/polkit-1/polkit-agent-helper-1"
  - "/usr/bin/doas"
  - "/usr/lib/systemd/systemd"
  - "/usr/lib/systemd/systemd-logind"
  - "/usr/sbin/unix_chkpwd"

enabled: true
log_filtered_events: false  # 是否记录被过滤的事件
```

### 匹配机制

- 基于可执行文件的**绝对路径**（通过 eBPF dentry 链遍历获取）
- 使用 Murmur OAAT64 哈希算法进行快速匹配
- 内核层直接过滤，不产生用户态事件
- 路径最深支持 16 层目录，最长 255 字节

### 修改方式

编辑配置文件后重启 agent 生效。

---

## 9. 清理测试环境

```bash
# 清理方法一的测试文件
sudo rm -f /tmp/suid_wrapper /tmp/suid_wrapper.c

# 清理方法二的测试文件
sudo rm -f /tmp/test_passwd

# 清理方法五的测试文件
sudo rm -f /tmp/privesc_test
```
