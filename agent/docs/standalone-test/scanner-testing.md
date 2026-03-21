# 漏洞扫描（病毒查杀）插件 — 测试指南

## 测试目标

验证 scanner 插件的病毒木马查杀功能（DataType 6060/6061）：插件使用 ClamAV 引擎扫描文件系统和运行进程，检测木马、Webshell、挖矿程序、后门等恶意文件。匹配成功时产生检测告警（DataType 6061）和扫描状态（DataType 6060）。本文档使用 EICAR 标准测试文件验证检测流程，覆盖目录扫描和进程扫描两种模式。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | root 权限 | `whoami` | 输出 `root` |
| 3 | ClamAV 开发库 | `dpkg -l libclamav-dev 2>/dev/null || rpm -q clamav-devel 2>/dev/null` | 已安装 |
| 5 | 病毒库文件 | `ls /var/lib/clamav/` | 目录下有 `.cvd` 或 `.cld` 文件 |

如果任一条件不满足，测试无法进行。

> ClamAV 安装：`apt install clamav libclamav-dev clamav-freshclam`（Debian/Ubuntu）或 `yum install clamav clamav-devel`（CentOS）。安装后执行 `sudo freshclam` 下载病毒库到 `/var/lib/clamav/`。

---

## Step 1：启动 Agent

打开 **Terminal A**，执行：

```bash
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=scanner -output=/tmp/scanner_test.log -test
```

### 启动成功判定

查看输出日志文件，**必须**依次看到以下日志行：

```bash
grep -E "ClamAV engine initialized|Virus database loaded" /tmp/scanner_test.log
```

预期输出：

```
INFO  ClamAV engine initialized
INFO  Virus database loaded  path=/var/lib/clamav
```

**判定规则**：
- 两行均出现 → 启动成功，ClamAV 引擎和病毒库加载完成，进入 Step 2
- `Failed to init ClamAV engine` → ClamAV 库链接失败，检查前置条件 4
- `Failed to load virus database` → 病毒库文件缺失，检查前置条件 5

### 日志位置

| 位置 | 说明 |
|------|------|
| `/tmp/scanner_test.log` | standalone 模式检测输出，**主要验证位置** |
| `/opt/cloudsec/agent/logs/scanner.log` | 插件内部运行日志 |

### 搜索技巧

```bash
# 搜索恶意文件检测结果
grep "Malware detected" /opt/cloudsec/agent/logs/scanner.log

# 搜索扫描状态
grep "DataType: 6060\|DataType: 6061" /tmp/scanner_test.log

# 按威胁类型搜索
grep "threat_type" /tmp/scanner_test.log
```

---

## Step 2：执行测试用例

打开 **Terminal B**，执行以下测试命令。

### 检测记录格式

scanner 插件检测到恶意文件时，在插件日志中输出：

```
{时间戳}  INFO  Malware detected  detail=[{威胁类型}] {文件路径} ({恶意家族}) - {文件大小} md5={MD5哈希}
```

检测结果通过 IPC 发送给 Agent，standalone 模式下写入 `-output` 指定的文件。记录字段：

| 字段 | 说明 | 示例值 |
|------|------|--------|
| `threat_type` | 威胁分类 | Trojan / Webshell / Miner / Backdoor |
| `file_name` | 文件名 | eicar_test.com |
| `file_path` | 完整路径 | /tmp/scanner_test/eicar_test.com |
| `file_size` | 文件大小（字节） | 68 |
| `file_md5` | MD5 哈希 | 44d88612fea8a8f36de82e1278abb02f |
| `file_sha256` | SHA256 哈希 | 275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f |
| `detection_engine` | 检测引擎 | ClamAV |
| `malware_family` | 恶意软件家族 | Eicar-Signature |
| `scan_time` | 检测时间戳 | 1709280000 |

### 通用判定规则

**PASS** 条件（全部满足）：
1. 插件日志中出现 `Malware detected` 行
2. `file_path` 与放置的测试文件路径一致
3. standalone 输出文件中有对应的 DataType 6061 记录

**FAIL** 条件（任一满足）：
- 放置测试文件后 60 秒内无 `Malware detected` 输出
- 检测到的文件路径与预期不一致

---

### 用例 1：EICAR 标准测试文件检测

**检测原理**：EICAR 是国际公认的杀毒软件测试标准文件，所有合格的杀毒引擎都能识别。文件内容固定为 68 字节的 ASCII 字符串，无实际危害。

**测试命令**（Terminal B）：

```bash
# 1. 创建测试目录
mkdir -p /tmp/scanner_test

# 2. 创建 EICAR 标准测试文件
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /tmp/scanner_test/eicar_test.com
```

等待自动扫描触发（插件的定时目录扫描周期为 24 小时，进程扫描为 1 小时）。如果不想等待，通过 hcids HTTP API 手动触发目录扫描（需启动 hcids Server）：

```bash
# 手动触发扫描（需要 hcids Server 运行且 Agent 已连接）
curl -X POST http://localhost:8081/api/task \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"123456","object_name":"scanner","data_type":6053,"data":"{\"path\":\"/tmp/scanner_test\"}","token":"test-scan-001"}'
```

**预期日志**（插件日志 `/opt/cloudsec/agent/logs/scanner.log`）：

```
INFO  Malware detected  detail=[Malware] /tmp/scanner_test/eicar_test.com (Eicar-Signature) - 68B md5=44d88612fea8a8f36de82e1278abb02f
```

**PASS 判定**：
1. `Malware detected` 出现
2. `file_path` 为 `/tmp/scanner_test/eicar_test.com`
3. `md5=44d88612fea8a8f36de82e1278abb02f`（EICAR 文件的标准 MD5）

> 说明：EICAR 文件是杀毒测试行业标准，所有 ClamAV 签名库都包含该检测规则。

---

### 用例 2：多文件批量检测

**检测原理**：在同一目录放置多个测试文件，验证扫描器能逐一检测。

**测试命令**（Terminal B）：

```bash
# 创建多个 EICAR 变体文件
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /tmp/scanner_test/eicar_1.exe
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /tmp/scanner_test/eicar_2.sh
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /tmp/scanner_test/eicar_3.py
```

**预期日志**：每个文件各产生一条 `Malware detected` 日志。

```bash
# 验证检测数量
grep "Malware detected" /opt/cloudsec/agent/logs/scanner.log | grep "/tmp/scanner_test/" | wc -l
```

**PASS 判定**：检测到的文件数 >= 3（加上用例 1 的文件共 4 个）。

---

### 用例 3：白名单路径验证（反向测试）

**检测原理**：验证白名单路径下的文件不被扫描。`/opt/cloudsec/` 在默认白名单中。

**测试命令**（Terminal B）：

```bash
# 在白名单路径创建 EICAR 文件
mkdir -p /opt/cloudsec/test_whitelist
echo 'X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*' > /opt/cloudsec/test_whitelist/eicar_safe.com
```

**预期结果**：该文件**不应**被检测到。

```bash
# 验证白名单路径文件未被扫描
grep "eicar_safe.com" /opt/cloudsec/agent/logs/scanner.log
```

**PASS 判定**：grep 无输出（白名单路径下的文件被过滤，不触发检测）。

---

## Step 3：记录测试结果

| # | 用例名称 | 测试文件 | 预期 | 实际 | PASS/FAIL |
|---|----------|---------|------|------|-----------|
| 1 | EICAR 标准测试 | `/tmp/scanner_test/eicar_test.com` | 检测到 Malware | | |
| 2 | 多文件批量检测 | `/tmp/scanner_test/eicar_*.{exe,sh,py}` | 每个文件均检测到 | | |
| 3 | 白名单路径验证 | `/opt/cloudsec/test_whitelist/eicar_safe.com` | 未检测到 | | |

---

## Step 4：清理与停止

```bash
# 1. Terminal A：按 Ctrl+C 停止 Agent

# 2. Terminal B：清理测试文件
rm -rf /tmp/scanner_test
rm -rf /opt/cloudsec/test_whitelist
rm -f /tmp/scanner_test.log
```

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| 编译报 `cannot find -lclamav` | ClamAV 开发库未安装 | `apt install libclamav-dev`（Debian/Ubuntu）或 `yum install clamav-devel`（CentOS） |
| `Failed to init ClamAV engine` | libclamav 运行时库缺失 | `ldconfig -p \| grep clamav` 确认库文件存在；如缺失执行 `ldconfig` |
| `Failed to load virus database` | 病毒库文件缺失 | 1) `ls /var/lib/clamav/` 确认有 `.cvd` 或 `.cld` 文件；2) 使用 `sudo freshclam` 下载最新病毒库 |
| EICAR 文件未被检测 | 扫描尚未触发 | 1) 定时扫描默认 24 小时，首次需手动触发或等待；2) 通过 hcids API 发送 6053 任务触发目录扫描 |
| 白名单路径文件被检测到 | `scanner.yaml` 白名单配置错误 | 检查 `config/scanner.yaml` 中 `filter.path_whitelist` 是否包含 `/opt/cloudsec` |
| 扫描导致系统卡顿 | cgroup 资源限制未生效 | 检查 `scanner.yaml` 中 `cgroup.enabled: true`；`cat /sys/fs/cgroup/*/scanner/` 确认 cgroup 已创建 |
| 插件启动后立即退出 | 配置文件缺失 | `ls /opt/cloudsec/agent/plugins/scanner/config/scanner.yaml` 确认配置文件存在 |
| 日志文件无内容 | 输出路径错误 | 确认启动命令使用 `-output=/tmp/scanner_test.log`；检查 `/opt/cloudsec/agent/logs/scanner.log` 插件日志 |
