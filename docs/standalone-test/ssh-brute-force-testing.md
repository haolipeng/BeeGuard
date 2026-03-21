# SSH 暴力破解检测 — 测试指南

## 测试目标

验证 detector 插件的 SSH 暴力破解检测功能（DataType 6001）：detector 通过 `nxadm/tail` 监控系统 SSH 认证日志（`/var/log/auth.log` 或 `/var/log/secure`），解析失败登录和成功登录事件，使用滑动窗口算法判断是否存在暴力破解行为。本文档覆盖 2 种告警场景：暴力破解尝试（120 秒内 ≥6 次失败）和暴力破解成功（暴力破解告警后 10 分钟内同一 IP 成功登录）。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | SSH 服务 | `systemctl status sshd 2>/dev/null \|\| systemctl status ssh` | 服务运行中（Ubuntu/Debian 服务名为 `ssh`，CentOS/RHEL 为 `sshd`） |
| 3 | 认证日志 | `ls /var/log/auth.log 2>/dev/null \|\| ls /var/log/secure` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |
| 5 | 攻击机（可选） | 另一台可 SSH 连接被测机的机器，或使用被测机本身 | 网络可达 |

如果任一条件不满足，测试无法进行。

### 检测规则说明

detector 默认加载 2 条 SSH 暴力破解规则：

| 规则名称 | 匹配动作 | 阈值 | 时间窗口 | 告警抑制 |
|---------|---------|------|---------|---------|
| `auth_failure_brute_force` | 密码认证失败（`Failed password`） | 6 次 | 120 秒 | 60 秒 |
| `invalid_user_brute_force` | 无效用户登录（`Invalid user`） | 6 次 | 120 秒 | 60 秒 |

此外，当某 IP 触发暴力破解告警后 **10 分钟内**成功登录，会额外产生 `brute_force_success` 告警。

### 白名单说明

SSH 暴力破解的白名单由配置文件 `config/rules/ssh_brute_force.yaml` 中的 `whitelist` 字段控制。默认配置为空（`whitelist: []`），即**不过滤任何 IP**，从本机（`127.0.0.1`）或远程机器发起的 SSH 连接都会被检测。如需添加白名单 IP，修改该 YAML 文件即可。

> 注意：FTP 检测的白名单默认包含 `127.0.0.1` 和 `::1`，但 SSH 检测不同，默认无白名单。

---

## Step 1：启动 Agent

在被测机上打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=detector -output=stderr -test
```

### 启动成功判定

在 Terminal A 的 stderr 输出中，**必须**看到以下日志行：

```
INFO  detection engine starting...
```

以及 SSH watcher 启动日志（表示开始监控认证日志）：

```
INFO  watcher/watcher.go:54  watching log file: /var/log/auth.log
```

**判定规则**：
- 上述日志出现 → 启动成功，进入 Step 2
- `no log paths configured` → SSH 认证日志路径不存在，检查前置条件 3
- Agent 无输出或立即退出 → 检查编译部署是否成功

### 告警日志位置

> **重要**：detector 插件的 SSH 暴力破解告警（DataType 6001）在 standalone 模式下**不会**输出到 stderr 或 `-output` 指定的文件，因为 standalone 的 output handler 不支持该数据类型。告警通过 detector 自身的日志系统以 WARN 级别输出。

| 位置 | 说明 |
|------|------|
| Terminal A (stderr) | 操作日志（启动、watcher 状态等），**非告警验证位置** |
| `/opt/cloudsec/agent/logs/plugins/detector/detector.log` | **告警验证位置**，WARN 级别的 ALERT 日志（目录在首次启动时自动创建） |

### 搜索技巧

```bash
# 实时监控告警（在另一个终端执行）
tail -f /opt/cloudsec/agent/logs/plugins/detector/detector.log | grep "ALERT"

# 按规则名搜索
grep "auth_failure_brute_force" /opt/cloudsec/agent/logs/plugins/detector/detector.log
grep "brute_force_success" /opt/cloudsec/agent/logs/plugins/detector/detector.log

# 按来源 IP 搜索
grep "ALERT.*192.168.1.100" /opt/cloudsec/agent/logs/plugins/detector/detector.log
```

---

## Step 2：执行测试用例

测试可从**被测机本机**（`127.0.0.1`）或**远程攻击机**发起 SSH 连接。在攻击端执行测试命令。

> **提示**：从本机测试更简单（无需第二台机器），默认 SSH 白名单为空，`127.0.0.1` 不会被过滤。

### 告警日志格式

每条告警在 `detector.log` 中以 WARN 级别输出：

```
{时间戳}  WARN  engine/engine.go:103  ALERT: ssh brute force detected from {IP}, count={次数}, rule={规则名}
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. `detector.log` 中出现包含 `ALERT: ssh brute force detected` 的 WARN 日志行
2. `from` 后的 IP 与攻击机 IP 一致
3. `rule` 与预期规则名一致

**FAIL** 条件（任一满足）：
- 执行完所有失败登录后 30 秒内 `detector.log` 无任何 `ALERT` 记录
- IP 或规则名与预期不一致

---

### 用例 1：BF001 — 暴力破解尝试（密码认证失败）

**检测原理**：同一 IP 在 120 秒内密码认证失败 ≥6 次，触发 `auth_failure_brute_force` 告警。

**测试命令**（攻击端，替换 `<被测机IP>` 为实际地址，本机测试可用 `127.0.0.1`）：

```bash
# 连续 6 次用错误密码尝试 SSH 登录（每次会提示输入密码，输入任意错误密码）
# -o PubkeyAuthentication=no 强制使用密码认证，避免公钥直接登录成功
for i in $(seq 1 6); do
    ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 -o PubkeyAuthentication=no root@<被测机IP> 2>/dev/null
    # 出现密码提示后输入错误密码，连接断开后继续
done
```

或使用 `sshpass` 工具自动化（需安装 `sudo apt install sshpass`）：

```bash
for i in $(seq 1 6); do
    sshpass -p 'wrong_password' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@<被测机IP> exit 2>/dev/null
done
```

> **注意**：6 次失败必须在 120 秒内完成。如果手动输入密码较慢，需加快速度或增加尝试次数。

**预期日志**（被测机 `detector.log`）：

```
WARN  ALERT: ssh brute force detected from <攻击机IP>, count=6, rule=auth_failure_brute_force
```

**验证命令**（被测机）：

```bash
grep "ALERT.*brute force" /opt/cloudsec/agent/logs/plugins/detector/detector.log
```

**PASS 判定**：上述命令有输出，且包含攻击机 IP 和 `rule=auth_failure_brute_force`。

---

### 用例 2：BF002 — 暴力破解成功（暴力破解后成功登录）

**检测原理**：在用例 1 触发暴力破解告警后 10 分钟内，同一 IP 成功 SSH 登录，触发 `brute_force_success` 告警。

**前置条件**：用例 1 已触发告警，且距告警时间不超过 10 分钟。

**测试命令**（攻击端）：

```bash
# 用正确密码 SSH 登录被测机（密码认证）
ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@<被测机IP>
# 登录成功后输入 exit 退出
```

或使用 `sshpass`：

```bash
sshpass -p '<正确密码>' ssh -o StrictHostKeyChecking=no -o PubkeyAuthentication=no root@<被测机IP> exit
```

如果攻击端已配置公钥认证，也可以直接使用公钥登录（无需 `-o PubkeyAuthentication=no`）：

```bash
ssh root@<被测机IP> exit
```

**预期日志**（被测机 `detector.log`）：

```
WARN  ALERT: ssh brute force detected from <攻击机IP>, count=0, rule=brute_force_success
```

**验证命令**（被测机）：

```bash
grep "brute_force_success" /opt/cloudsec/agent/logs/plugins/detector/detector.log
```

**PASS 判定**：上述命令有输出，且包含攻击机 IP 和 `rule=brute_force_success`。

> 说明：此用例必须在用例 1 告警后 10 分钟内执行，否则 bruteForceIPs 记录过期，不会触发 brute_force_success 告警。

---

## Step 3：记录测试结果

| # | 用例 ID | 测试场景 | 预期规则 | 预期 | 实际 | PASS/FAIL |
|---|---------|---------|---------|------|------|-----------|
| 1 | BF001 | 暴力破解尝试（6 次错误密码） | auth_failure_brute_force | 告警 | | |
| 2 | BF002 | 暴力破解成功（告警后成功登录） | brute_force_success | 告警 | | |

---

## Step 4：清理与停止

```bash
# 1. Terminal A：按 Ctrl+C 停止 Agent

# 2. 攻击机：关闭所有 SSH 连接（如有）
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动后无 watcher 日志 | 认证日志不存在 | `ls /var/log/auth.log /var/log/secure` 确认文件存在；某些系统使用 `journald` 不生成文件日志 |
| 本机 SSH 登录不产生失败日志 | 公钥认证优先 | 被测机对攻击端配置了公钥认证时，SSH 会跳过密码直接成功登录。加 `-o PubkeyAuthentication=no` 强制密码认证 |
| 6 次失败后无告警 | 未在 120 秒窗口内完成 | 确保 6 次失败在 2 分钟内完成；用 `sshpass` 自动化可避免手动输入过慢 |
| `detector.log` 中无 ALERT | watcher 未读取到新日志行 | 1) watcher 从文件末尾开始读取，**Agent 启动后**产生的日志才会被处理；2) 确认 Agent 启动时间早于 SSH 登录尝试 |
| 在 stderr 或 `-output` 文件中找不到告警 | standalone 不支持 6001 输出 | detector 的暴力破解告警通过 `detector.log` 输出（WARN 级别），不走 standalone 的 output handler |
| `message repeated N times` 格式未识别 | syslog 聚合了重复消息 | 解析器已支持此格式，一条 `message repeated 5 times` 等价于 5 次独立事件；如仍未触发，检查总次数是否达到阈值 |
| 暴力破解成功告警未出现 | 超过 10 分钟窗口 | 成功登录必须在暴力破解告警后 10 分钟内发生；超时后 bruteForceIPs 记录被清除 |
| 连续测试时第二次不告警 | 告警抑制（60 秒） | 同一 IP 触发告警后，60 秒内不会再次告警；等待 60 秒后重新尝试 |
| 日志文件路径不同 | 系统使用 `/var/log/secure` | CentOS/RHEL 使用 `/var/log/secure`，Ubuntu/Debian 使用 `/var/log/auth.log`；detector 默认同时监控两者 |
