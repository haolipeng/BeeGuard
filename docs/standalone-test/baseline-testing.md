# 基线检查插件 — 测试指南

## 测试目标

验证 baseline 插件的安全基线检查功能（DataType 8000/8010）：插件接收检查任务后，读取系统配置文件（如 `/etc/login.defs`、`/etc/security/pwquality.conf`），按规则引擎逐项检查，返回每个检查项的 PASS/FAIL 结果。本文档使用 E2E 测试程序验证完整流程：发送任务 → 插件执行检查 → 接收并解析结果。

## 前置条件

| # | 检查项 | 检查命令 | 通过标准 |
|---|--------|---------|---------|
| 1 | 操作系统 | `uname -s` | 输出 `Linux` |
| 2 | root 权限 | `whoami` | 输出 `root` |
| 3 | 编译环境 | `go version` | Go 已安装 |
| 4 | 系统类型 | `cat /etc/os-release` | CentOS、Debian 或 Ubuntu |

如果任一条件不满足，测试无法进行。

> Baseline 插件不依赖 eBPF，无内核版本和 BTF 要求。

---

## Step 1：编译部署

**方式一：使用测试脚本（推荐）**

```bash
cd /home/work/goProject/src/company/agent/tests/e2e/baseline
chmod +x test.sh
./test.sh
```

脚本会自动完成编译 → 准备插件目录 → 运行测试。如果使用此方式，直接跳到 Step 3 验证结果。

**方式二：手动编译**

```bash
# 1. 编译 baseline 插件
cd /home/work/goProject/src/company/agent/business_plugins/baseline
go build -o baseline main.go

# 2. 准备插件目录
mkdir -p /tmp/plugin/baseline
cp baseline /tmp/plugin/baseline/baseline
chmod +x /tmp/plugin/baseline/baseline
```

**验证**：执行 `ls -la /tmp/plugin/baseline/baseline`，文件存在且有执行权限即成功。

---

## Step 2：运行 E2E 测试程序

```bash
cd /home/work/goProject/src/company/agent/tests/e2e/baseline
go run main.go 2>&1 | tee /tmp/baseline_test.log
```

> 测试程序会自动执行：启动 plugin daemon → 加载 baseline 插件 → 发送测试任务（baseline_id=1200, check_id_list=[1001,1002,1003]） → 接收并打印结果。

### 启动成功判定

在输出中，**必须**看到以下两行日志：

```
INFO  baseline plugin loaded successfully
INFO  task sent successfully to baseline plugin
```

**判定规则**：
- 两行均出现 → 启动成功，插件已加载且任务已发送，等待结果输出
- `baseline plugin loaded successfully` 未出现 → 插件加载失败，检查插件文件是否在 `/tmp/plugin/baseline/baseline`
- `task sent successfully` 未出现 → 任务发送失败，检查插件是否正常运行

### 日志位置

| 位置 | 说明 |
|------|------|
| 终端 stdout/stderr | 实时输出，**主要观察位置** |
| `/tmp/baseline_test.log` | tee 保存的完整输出，可用 grep 搜索 |
| `/tmp/plugin/baseline/baseline.stderr` | 插件进程自身的日志 |

### 搜索技巧

```bash
# 搜索检查结果摘要
grep "Baseline Check Result" /tmp/baseline_test.log

# 搜索具体检查项结果
grep "CheckID:" /tmp/baseline_test.log

# 搜索任务状态
grep "Task Status" /tmp/baseline_test.log

# 搜索失败项
grep "Result: FAIL" /tmp/baseline_test.log
```

---

## Step 3：验证测试结果

### 输出格式

测试程序接收到 DataType 8000 结果后，以格式化文本输出：

```
========== Baseline Check Result ==========
Baseline ID: 1200
Status: success
Token: test-token-123
Check Items Count: 3
  [1] CheckID: 1001, Result: PASS, Title: 检查项标题
  [2] CheckID: 1002, Result: PASS, Title: 检查项标题
  [3] CheckID: 1003, Result: FAIL, Title: 检查项标题
==========================================
```

随后输出 DataType 8010 任务状态：

```
========== Task Status ==========
Status: succeed
Token: test-token-123
Message:
================================
```

### 通用判定规则

**PASS** 条件（全部满足）：
1. 输出中出现 `Baseline Check Result` 段落
2. `Status: success` — 插件执行成功（不是指每个检查项都通过）
3. `Check Items Count` >= 1 — 有检查项结果返回
4. 每个检查项有明确的 `Result: PASS` 或 `Result: FAIL`
5. `Task Status` 段落中 `Status: succeed`

**FAIL** 条件（任一满足）：
- 30 秒内未出现 `Baseline Check Result` 输出
- `Status: error` — 插件执行出错
- `Check Items Count: 0` — 无检查项结果

> 注意：检查项 `Result: FAIL` 表示该系统配置不符合安全基线要求，属于正常检测结果，不是测试失败。

---

### 用例 1：密码失效时间检查（check_id=1）

**检查内容**：`/etc/login.defs` 中 `PASS_MAX_DAYS` 是否 <= 90

**手动验证当前系统值**（Terminal B）：

```bash
grep -E '^\s*PASS_MAX_DAYS' /etc/login.defs
```

**预期输出中的对应行**：

```
  [1] CheckID: 1, Result: PASS, Title: 设置密码失效时间<=90天
```

**PASS 判定**：
- 如果 `PASS_MAX_DAYS` <= 90 → 检查项应为 `Result: PASS`
- 如果 `PASS_MAX_DAYS` > 90 或未设置 → 检查项应为 `Result: FAIL`
- 将手动查看的系统值与检查结果对比，两者一致即 PASS

---

### 用例 2：密码修改最短周期检查（check_id=2）

**检查内容**：`/etc/login.defs` 中 `PASS_MIN_DAYS` 是否 >= 2

**手动验证当前系统值**（Terminal B）：

```bash
grep -E '^\s*PASS_MIN_DAYS' /etc/login.defs
```

**预期输出中的对应行**：

```
  [2] CheckID: 2, Result: PASS, Title: 密码修改最短周期>=2天
```

**PASS 判定**：
- 如果 `PASS_MIN_DAYS` >= 2 → 检查项应为 `Result: PASS`
- 如果 `PASS_MIN_DAYS` < 2 或未设置 → 检查项应为 `Result: FAIL`
- 将手动查看的系统值与检查结果对比，两者一致即 PASS

---

### 用例 3：密码到期警告天数检查（check_id=3）

**检查内容**：`/etc/login.defs` 中 `PASS_WARN_AGE` 是否 >= 7

**手动验证当前系统值**（Terminal B）：

```bash
grep -E '^\s*PASS_WARN_AGE' /etc/login.defs
```

**预期输出中的对应行**：

```
  [3] CheckID: 3, Result: PASS, Title: 密码到期时间警告>=7天
```

**PASS 判定**：
- 如果 `PASS_WARN_AGE` >= 7 → 检查项应为 `Result: PASS`
- 如果 `PASS_WARN_AGE` < 7 或未设置 → 检查项应为 `Result: FAIL`
- 将手动查看的系统值与检查结果对比，两者一致即 PASS

---

## Step 4：记录测试结果

| # | 检查项 | 检查内容 | 系统实际值 | 插件判定 | 结果一致 | PASS/FAIL |
|---|--------|---------|-----------|---------|---------|-----------|
| 1 | check_id=1 | PASS_MAX_DAYS <= 90 | | | | |
| 2 | check_id=2 | PASS_MIN_DAYS >= 2 | | | | |
| 3 | check_id=3 | PASS_WARN_AGE >= 7 | | | | |
| - | 任务状态 | DataType 8010 status=succeed | - | | - | |

---

## Step 5：清理与停止

```bash
# 1. 按 Ctrl+C 停止测试程序（如果仍在运行）

# 2. 清理插件目录和日志（可选）
rm -rf /tmp/plugin/baseline
rm -f /tmp/baseline_test.log
```

---

## 修改测试任务

默认测试任务发送 baseline_id=1200（CentOS 基线）的 check_id 1、2、3。如需测试其他检查项或基线：

编辑 `tests/e2e/baseline/main.go` 中的 `sendTestTask()` 函数：

```go
taskData := map[string]interface{}{
    "baseline_id":   1200,                    // 1200=CentOS, 1300=Debian, 1400=Ubuntu, 5000=弱口令
    "check_id_list": []int{1001, 1002, 1003}, // 修改为需要测试的检查项 ID
}
```

可用基线配置：

| baseline_id | 名称 | 适用系统 | 配置文件 |
|------------|------|---------|---------|
| 1200 | CentOS 基线 | CentOS | `config/linux/1200.yaml` |
| 1300 | Debian 基线 | Debian | `config/linux/1300.yaml` |
| 1400 | Ubuntu 基线 | Ubuntu | `config/linux/1400.yaml` |
| 5000 | 弱口令检查 | 全平台 | `config/linux/5000.yaml` |

修改后重新执行 `go run main.go` 即可。

---

## 常见问题排查

| 问题现象 | 可能原因 | 排查步骤 |
|---------|---------|---------|
| `plugin executable not found` | 插件文件不在预期路径 | `ls -la /tmp/plugin/baseline/baseline` 确认文件存在且有执行权限 |
| `baseline plugin not found` | 插件加载失败 | 1) 查看输出中是否有编译错误；2) `cat /tmp/plugin/baseline/baseline.stderr` 查看插件日志 |
| `failed to send task` | 插件进程未就绪 | 增加 `sendTestTask()` 前的等待时间（当前为 3 秒） |
| 30 秒无结果输出 | 插件执行超时或崩溃 | 1) `ps aux \| grep baseline` 确认插件进程存在；2) `cat /tmp/plugin/baseline/baseline.stderr` 查看错误日志 |
| `Status: error` | 基线配置文件缺失 | 确认 `business_plugins/baseline/config/linux/` 下有对应的 YAML 文件；检查 baseline_id 与系统类型是否匹配 |
| Check Items Count: 0 | check_id_list 与配置不匹配 | 对照 YAML 配置文件中的 `check_id` 字段，确认发送的 ID 列表正确 |
| 检查结果与系统实际值不一致 | 规则匹配逻辑问题 | 1) 手动查看对应配置文件的值；2) 对照 YAML 中的 `filter` 正则表达式和 `result` 判定条件 |
| `cannot find package "business_plugins/lib"` | Go 模块依赖问题 | 在 `tests/e2e/baseline/` 目录执行 `go mod tidy` |
