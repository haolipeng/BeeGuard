# 高危命令检测 — 测试指南

## 测试目标

验证 ebpf_base_detector 插件的高危命令检测功能（DataType 6003）：eBPF 在 `sched_process_exec` Hook 中捕获进程执行事件，用户态将命令行参数与 `dangerous_commands.yaml` 中的规则进行匹配，匹配成功时产生告警。本文档选取 4 条代表性规则进行验证，覆盖 regex、prefix 两种匹配方式和 critical、high 两种严重程度。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | 内核版本 | `uname -r` | 版本 >= 5.4 |
| 3 | BTF 支持 | `ls /sys/kernel/btf/vmlinux` | 文件存在 |
| 4 | root 权限 | `whoami` | 输出 `root` |

如果任一条件不满足，测试无法进行。

---

## Step 1：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test
```

### 启动成功判定（加载规则的日志并没有显示）

在 Terminal A 的 stderr 输出中，**必须**看到以下与“规则加载”相关的日志（其它规则组不要求精确条数，只要成功加载日志出现即可）。

1. 危险命令规则（`dangerous_commands.yaml`）
需要出现包含 `Detection rules loaded` 且 `rules>=1`（或 `count>=1`）的日志，用于确认危险命令规则加载完成。

**判定规则**：

- `rules>=1` 或 `count>=1` → 启动成功，进入 Step 2
- `rules=0`/`count=0` 或该行未出现 → 启动失败，检查 `dangerous_commands.yaml` 是否在 `/opt/cloudsec/agent/plugins/ebpf_base_detector/config/` 目录下

2. 其它规则组（不硬校验条数）
需要能在 stderr 中看到（至少出现一次）以下日志（条数不强制校验）：

- `Container dangerous command rules loaded`
- `Malicious request rules loaded`
- `Sensitive file rules loaded`
- `Container sensitive file rules loaded`

如果上述日志缺失/不出现，通常表示对应配置文件不存在或加载失败（但不影响你先继续观察 Step 2 的告警是否能触发）。

`failed to load eBPF` 错误 → 内核不支持，检查前置条件 2、3

### 日志位置

| 位置 | 说明 |
|------|------|
| Terminal A (stderr) | 实时输出，**主要观察位置** |
| `/opt/cloudsec/agent/logs/ebpf_base_detector.log` | 同内容持久化文件，可用 grep 搜索 |

### 搜索技巧

如果 Terminal A 输出内容较多，可使用 grep 过滤：

```bash
# 方式一：启动时只显示告警（Terminal A）
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | grep "Dangerous command detected"

# 方式二：保存全部输出到文件，在另一个终端搜索
sudo ./bin/agent -standalone -plugins=ebpf_base_detector -output=stderr -test 2>&1 | tee /tmp/ebpf_test.log
# 另一个终端
grep "Dangerous command detected" /tmp/ebpf_test.log

# 方式三：按规则 ID 精确搜索
grep "rule_id=2001" /tmp/ebpf_test.log
```

---

## Step 2：执行测试用例

### 方式一：自动化测试脚本（推荐）

打开 **Terminal B**，执行自动化测试脚本：

```bash
cd /home/work/goProject/src/BeeGuard/agent
sudo bash tests/test-dangerous-commands.sh
```

**脚本优势**：
- 自动执行所有4个测试用例，每个间隔2秒
- 提供彩色进度显示和完成状态
- 自动处理文件创建和清理
- 测试结束后汇总所有预期告警信息

**预期结果**：脚本执行完成后，在 Terminal A 中应该看到以下4条告警：

```
INFO  Dangerous command detected  rule_id=2001  rule_name=危险删除操作  severity=critical  uid=0  comm=rm  args=-rf /tmp/dc001_nonexistent_test_dir
INFO  Dangerous command detected  rule_id=2002  rule_name=敏感文件访问  severity=high  uid=0  comm=cat  args=/etc/passwd
INFO  Dangerous command detected  rule_id=2003  rule_name=危险权限修改  severity=high  uid=0  comm=chmod  args=777 /tmp/dc003_test
INFO  Dangerous command detected  rule_id=2009  rule_name=内核模块操作  severity=high  uid=0  comm=insmod  args=/tmp/nonexistent.ko
```

---

### 方式二：手动执行测试用例（备选）

如果需要手动测试或调试特定用例，可以逐条执行以下测试命令。每执行一条后，回到 Terminal A 查看是否出现对应告警。

#### 告警日志格式

每条告警在 Terminal A 中以一行结构化日志输出：

```
{时间戳}  INFO  ebpf_base_detector/event_handlers.go:51  Dangerous command detected  rule_id={ID}  rule_name={名称}  severity={级别}  uid={UID}  comm={进程名}  args={命令参数}
```

#### 通用判定规则

**PASS** 条件（全部满足）：
1. Terminal A 出现包含 `Dangerous command detected` 的日志行
2. `rule_id` 与测试用例的规则 ID 一致
3. `comm` 与执行的命令名一致
4. `args` 中包含执行的命令参数

**FAIL** 条件（任一满足）：
- 执行命令后 5 秒内 Terminal A 无任何 `Dangerous command detected` 输出
- `rule_id` 与预期不一致

---

#### 用例 1：规则 2001 — 危险删除操作（critical）

**匹配方式**：regex，匹配 `rm\s+.*-rf\s+/`

**测试命令**（Terminal B）：

```bash
rm -rf /tmp/dc001_nonexistent_test_dir
```

**预期日志**（Terminal A）：

```
INFO  Dangerous command detected  rule_id=2001  rule_name=危险删除操作  severity=critical  uid=0  comm=rm  args=-rf /tmp/dc001_nonexistent_test_dir
```

**PASS 判定**：出现 `Dangerous command detected`，且 `rule_id=2001`，`comm=rm`。

> 说明：目标路径不存在，rm 执行无实际影响，但 eBPF 在 execve 时即捕获。

---

#### 用例 2：规则 2002 — 敏感文件访问（high）

**匹配方式**：regex，匹配 `cat\s+.*/etc/(passwd|shadow|sudoers)`

**测试命令**（Terminal B）：

```bash
cat /etc/passwd > /dev/null
```

**预期日志**（Terminal A）：

```
INFO  Dangerous command detected  rule_id=2002  rule_name=敏感文件访问  severity=high  uid=0  comm=cat  args=/etc/passwd
```

**PASS 判定**：`rule_id=2002`，`comm=cat`，`args` 包含 `/etc/passwd`。

---

#### 用例 3：规则 2003 — 危险权限修改（high）

**匹配方式**：regex，匹配 `chmod\s+.*777\s+/`

**测试命令**（Terminal B）：

```bash
touch /tmp/dc003_test && chmod 777 /tmp/dc003_test && rm -f /tmp/dc003_test
```

**预期日志**（Terminal A）：

```
INFO  Dangerous command detected  rule_id=2003  rule_name=危险权限修改  severity=high  uid=0  comm=chmod  args=777 /tmp/dc003_test
```

**PASS 判定**：`rule_id=2003`，`comm=chmod`，`args` 包含 `777`。

> 说明：`rm -f` 不匹配 2001 的 `-rf` 模式，不会产生额外告警。

---

#### 用例 4：规则 2009 — 内核模块操作（high）

**匹配方式**：prefix，进程名以 `insmod` 开头

**测试命令**（Terminal B）：

```bash
insmod /tmp/nonexistent.ko 2>/dev/null; true
```

**预期日志**（Terminal A）：

```
INFO  Dangerous command detected  rule_id=2009  rule_name=内核模块操作  severity=high  uid=0  comm=insmod  args=/tmp/nonexistent.ko
```

**PASS 判定**：`rule_id=2009`，`comm=insmod`。

> 说明：模块不存在会报错，但 eBPF 在 execve 阶段已捕获，不影响检测。

---

## Step 3：记录测试结果

| # | 规则 ID | 规则名称 | 严重程度 | 测试命令 | 预期 | 实际 | PASS/FAIL |
|---|---------|----------|----------|----------|------|------|-----------|
| 1 | 2001 | 危险删除操作 | critical | `rm -rf /tmp/dc001_nonexistent_test_dir` | 告警 | | |
| 2 | 2002 | 敏感文件访问 | high | `cat /etc/passwd > /dev/null` | 告警 | | |
| 3 | 2003 | 危险权限修改 | high | `chmod 777 /tmp/dc003_test` | 告警 | | |
| 4 | 2009 | 内核模块操作 | high | `insmod /tmp/nonexistent.ko` | 告警 | | |

---

## Step 4：清理与停止

```bash
# 1. Terminal A：按 Ctrl+C 停止 Agent

# 2. Terminal B：清理测试残留文件（如果手动执行测试）
rm -f /tmp/dc003_test /tmp/dc003_suid
```

> **注意**：如果使用自动化测试脚本，文件清理会自动完成。

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| Agent 启动报 `failed to load eBPF` | 内核不支持或无 root 权限 | 1) `whoami` 确认 root；2) `uname -r` 确认 >= 5.4；3) `ls /sys/kernel/btf/vmlinux` 确认 BTF |
| Terminal A 无任何输出 | 输出重定向错误 | 确认启动命令使用 `-output=stderr`，而非文件路径 |
| 规则加载 count=0 | 配置文件缺失或格式错误 | 1) `ls /opt/cloudsec/agent/plugins/ebpf_base_detector/config/dangerous_commands.yaml` 确认文件存在；2) 用 `python3 -c "import yaml; yaml.safe_load(open('...'))"` 检查 YAML 语法 |
| 命令执行了但无告警 | 命令参数不匹配规则 | 1) 对照规则的 `patterns` 检查命令行是否匹配；2) 在日志文件中搜索：`grep "rule_id" /opt/cloudsec/agent/logs/ebpf_base_detector.log` |
| 管道命令未触发告警 | 管道中每个子命令是独立 execve | 使用 `bash -c '完整管道命令'` 包装，让 eBPF 在 bash 进程 execve 时捕获完整命令行 |
| 脚本执行权限问题 | 脚本不可执行 | `chmod +x tests/test-dangerous-commands.sh` |

---

## 测试脚本说明

### 脚本位置
`/home/work/goProject/src/BeeGuard/agent/tests/test-dangerous-commands.sh`

### 脚本功能
- 自动执行4个测试用例，每个间隔2秒
- 彩色输出显示测试进度
- 自动权限检查和错误处理
- 测试完成后汇总预期告警信息
- 自动清理测试文件

### 使用方法
```bash
# 确保脚本可执行
chmod +x tests/test-dangerous-commands.sh

# 执行测试
sudo bash tests/test-dangerous-commands.sh
``` 的 execve 参数中捕获完整字符串 |
| bash 内建命令未触发 | 内建命令不产生 execve 事件 | `history`、`export`、`cd` 等是 shell 内建命令，eBPF 无法捕获；需 `bash -c "..."` 包装 |
| 告警出现但 rule_id 不符预期 | 命令同时匹配多条规则 | 正常现象，一条命令可能触发多条规则告警 |
| 告警延迟超过 5 秒 | standalone 刷新间隔较长 | 检查配置中 `flush_interval`（默认 1 秒）；eBPF 事件本身无延迟，延迟来自用户态轮询 |
