# Baseline 插件测试

这是一个独立的测试程序，用于测试 baseline 插件的完整流程，不会影响主程序。

## 测试流程概述

完整的测试流程包括：

1. **编译 baseline 插件** - 将插件源码编译为可执行文件
2. **准备插件目录** - 将插件文件放到 agent 可以找到的位置
3. **运行测试程序** - 启动测试 agent，自动执行：
   - 加载 baseline 插件
   - 发送测试任务给插件
   - 接收插件返回的结果
   - 格式化打印结果

## 快速开始（推荐）

使用测试脚本一键执行所有步骤：

```bash
cd /home/work/goProject/src/BeeGuard/agent/tests/e2e/baseline
chmod +x test.sh
./test.sh
```

脚本会自动完成：
- 编译 baseline 插件
- 准备插件目录
- 运行测试程序

## 手动执行步骤

如果想手动执行，可以按照以下步骤：

### 步骤 1: 编译 baseline 插件

```bash
cd /home/work/goProject/src/BeeGuard/agent/business_plugins/baseline
go mod tidy
go build -o baseline main.go
```

编译成功后会在当前目录生成 `baseline` 可执行文件。

### 步骤 2: 准备插件目录

Agent 会在 `/tmp/plugin/{插件名}/` 目录下查找插件可执行文件：

```bash
# 创建插件目录
mkdir -p /tmp/plugin/baseline

# 复制编译好的插件
cp /home/work/goProject/src/BeeGuard/agent/business_plugins/baseline/baseline /tmp/plugin/baseline/baseline

# 确保插件有执行权限
chmod +x /tmp/plugin/baseline/baseline
```

### 步骤 3: 运行测试程序

```bash
cd /home/work/goProject/src/BeeGuard/agent/tests/e2e/baseline
go mod tidy
go run main.go
```

或者先编译再运行：

```bash
cd /home/work/goProject/src/BeeGuard/agent/tests/e2e/baseline
go build -o test_agent main.go
./test_agent
```

### 步骤 4: 观察输出

测试程序会输出详细的执行信息：

1. **初始化阶段**
   - Logger 初始化
   - Plugin daemon 启动

2. **插件加载阶段**
   - 插件配置同步
   - 插件进程启动确认

3. **任务执行阶段**
   - 任务发送确认
   - 接收到的基线检查结果（格式化输出）
   - 任务状态信息

**示例输出：**

```
=== Baseline Plugin Test ===
Starting test agent...
INFO    ... running ...
INFO    plugin daemon startup
INFO    syncing plugins...
INFO    plugin has been loaded
INFO    baseline plugin loaded successfully
INFO    task sent successfully to baseline plugin

========== Baseline Check Result ==========
Baseline ID: 1200
Status: success
Token: test-token-123
Check Items Count: 3
  [1] CheckID: 1001, Result: PASS, Title: 检查项 1001
  [2] CheckID: 1002, Result: PASS, Title: 检查项 1002
  [3] CheckID: 1003, Result: FAIL, Title: 检查项 1003
==========================================

========== Task Status ==========
Status: succeed
Token: test-token-123
Message: 
================================
```

### 步骤 5: 停止测试

按 `Ctrl+C` 发送 SIGTERM 或 SIGINT 信号，测试程序会优雅退出。

## 测试脚本说明

`test.sh` 脚本提供了完整的自动化测试流程：

- **自动编译** - 自动编译 baseline 插件
- **自动准备** - 自动创建目录并复制插件文件
- **错误检查** - 检查每个步骤是否成功
- **彩色输出** - 使用颜色标识不同状态（成功/警告/错误）

脚本执行时会显示详细的进度信息，如果任何步骤失败，会立即停止并显示错误信息。

## 测试任务说明

### 默认测试任务

测试程序会发送以下任务给 baseline 插件：

```json
{
  "baseline_id": 1200,
  "check_id_list": [1001, 1002, 1003]
}
```

### 修改测试任务

编辑 `test/main.go` 中的 `sendTestTask()` 函数可以修改测试任务内容：

```go
taskData := map[string]interface{}{
    "baseline_id":  1200,                    // 修改基线 ID
    "check_id_list": []int{1001, 1002, 1003}, // 修改检查项列表
}
```

修改后重新运行测试程序即可。

## 故障排查

### 问题 1: 插件文件未找到

**错误信息**: `plugin executable not found: /tmp/plugin/baseline/baseline`

**解决方法**:
- 确保插件已成功编译
- 检查插件文件是否在正确位置：`/tmp/plugin/baseline/baseline`
- 检查文件权限：`chmod +x /tmp/plugin/baseline/baseline`
- 验证文件是否存在：`ls -lh /tmp/plugin/baseline/baseline`

### 问题 2: 插件加载失败

**可能原因**:
- 插件编译失败
- 插件依赖缺失
- 模块路径配置错误

**解决方法**:
- 检查 baseline 插件编译日志
- 运行 `go mod tidy` 确保依赖正确
- 检查 `business_plugins/baseline/go.mod` 中的 replace 指令
- 查看 agent 日志中的详细错误信息

### 问题 3: 任务发送失败

**错误信息**: `baseline plugin not found` 或 `failed to send task`

**解决方法**:
- 确认插件已成功加载（查看日志中的 "plugin has been loaded"）
- 检查插件名称是否正确（应该是 "baseline"）
- 等待插件完全启动（可能需要几秒钟）

### 问题 4: 没有收到结果

**可能原因**:
- 插件进程异常退出
- 数据缓冲区问题
- 通信管道问题

**解决方法**:
- 查看插件 stderr 日志：`cat /tmp/plugin/baseline/baseline.stderr`
- 检查插件进程是否运行：`ps aux | grep baseline`
- 查看 agent 日志中的错误信息
- 检查 buffer 模块是否正常工作

### 问题 5: 模块路径错误

**错误信息**: `cannot find package "business_plugins/lib"`

**解决方法**:
- 确保 `business_plugins/lib/go.mod` 存在
- 运行 `go mod tidy` 更新依赖
- 检查 `business_plugins/baseline/go.mod` 中的 replace 指令是否正确

## 目录结构说明

```
agent/
├── test/                    # 测试目录
│   ├── main.go             # 测试程序主文件
│   ├── test.sh             # 自动化测试脚本
│   └── README.md           # 本文档
├── business_plugins/
│   ├── baseline/            # baseline 插件
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── check/
│   └── lib/                 # 插件库（独立模块）
│       ├── client.go
│       ├── bridge.pb.go
│       └── go.mod
└── plugin/                  # Agent 插件管理模块
```

## 技术细节

### 插件通信协议

- **任务发送**: Agent 通过管道（pipe）发送 protobuf 编码的任务数据
- **结果接收**: 插件通过管道发送 protobuf 编码的结果数据
- **数据格式**: 使用 protobuf wire format，前 4 字节为长度前缀

### 数据流向

```
Agent (test/main.go)
  ↓ SendTask()
Plugin Manager (plugin/plugin.go)
  ↓ 通过管道发送
Baseline Plugin (baseline/main.go)
  ↓ ReceiveTask() → Analysis() → SendRecord()
Plugin Manager
  ↓ 接收数据 → buffer.WriteEncodedRecord()
Buffer (buffer/buffer.go)
  ↓ ReadEncodedRecords()
Test Program
  ↓ printRecord()
控制台输出
```

### 关键数据类型

- **DataType 8000**: 基线检查结果
- **DataType 8010**: 任务状态信息
- **Task**: 任务结构（DataType, ObjectName, Data, Token）

## 下一步

测试成功后，可以：

1. **修改检查逻辑** - 在 `baseline/check/analysis.go` 中实现真实的检查逻辑
2. **添加更多测试** - 测试不同的任务类型和参数
3. **集成到主程序** - 将测试逻辑整合到主 agent 程序中
4. **开发新插件** - 参考 baseline 插件开发其他业务插件

