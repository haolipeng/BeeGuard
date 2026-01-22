# Agent-Server gRPC 通道端到端测试文档

| 事项                     | 操作人 | 时间       |
| ------------------------ | ------ | ---------- |
| 创建grpc通道的端到端测试 | 郝立鹏 | 2026-01-22 |
|                          |        |            |
|                          |        |            |



## 概述

Agent 和 Server 通过双向流 gRPC 服务 `Transfer` 通信：
- **Agent → Server**: `PackagedData` (数据包含 agent 元信息 + EncodedRecord 记录)
- **Server → Agent**: `Command` (控制命令、任务、插件配置)

## 前置条件

### 1. 创建 Agent 配置文件

在 Agent 目录下创建 `config.yaml`：

```bash
cd /home/work/goProject/src/company/agent
```

`config.yaml` 内容：
```yaml
server: "localhost:50051"
connect_timeout: 30
working_directory: "/tmp/cloudsec-agent"
retry_max_count: 10
retry_interval: 5
```

### 2. 代码修改

#### main.go 修改

确保 `main.go` 包含以下内容：
- 初始化配置：`config.Init()`
- 启动传输守护进程：`transport.StartTransfer()`

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gitlab.myinterest.top/security/agent/config"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/transport"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("agent start running!")

	// 初始化配置
	if err := config.Init(); err != nil {
		slog.Error("failed to init config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("config initialized successfully")

	wg := &sync.WaitGroup{}
	zap.S().Info("++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++")

	Context, Cancel := context.WithCancel(context.Background())

	// 启动插件守护进程
	wg.Add(1)
	go plugin.Startup(Context, wg)

	// 启动传输守护进程（gRPC 连接）
	wg.Add(1)
	go transport.StartTransfer(Context, wg)

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigs
		zap.S().Error("receive signal:", sig.String())
		zap.S().Info("wait for 5 secs to exit")
		<-time.After(time.Second * 5)
		Cancel()
	}()

	wg.Wait()
}
```

#### transport/transfer.go 修改

修改 `handleSend` 函数，即使没有 records 也发送心跳包：

```go
case <-ticker.C:
	recs := buffer.ReadEncodedRecords()
	// 即使没有 records 也发送心跳包（包含 agent 元信息）

	// 获取主机信息
	// ... 其余代码
```

## 测试步骤

### 步骤 1：启动 Server（终端 1）

```bash
cd /home/work/goProject/src/company/server
go run main.go -port 50051
```

**预期输出：**
```
gRPC Server 启动，监听端口 :50051
```

### 步骤 2：启动 Agent（终端 2）

```bash
cd /home/work/goProject/src/company/agent
go run main.go
```

**预期 Agent 输出：**
```
agent start running!
2026/01/21 21:35:43 INFO config initialized successfully
2026/01/21 21:35:43 INFO transfer daemon startup
2026/01/21 21:35:43 INFO dialing server server=localhost:50051 retry=0 timeout=30s
2026/01/21 21:35:43 INFO connected to server server=localhost:50051
2026/01/21 21:35:43 INFO get connection successfully
2026/01/21 21:35:43 INFO receive handler running
2026/01/21 21:35:43 INFO send handler running
```

**预期 Server 日志：**
```
[Transfer] Agent 连接 agent_id=6bb9735d-66ee-556a-8981-62d127daf308 hostname=ubuntu version=
```

### 步骤 3：验证心跳传输

Agent 启动后会自动发送心跳包（每 100ms），Server 端会持续接收连接活动。

### 步骤 4：停止 Agent

按 `Ctrl+C` 停止 Agent。

**预期 Server 日志：**
```
[Transfer] Agent 断开 agent_id=6bb9735d-66ee-556a-8981-62d127daf308
```



## 验证清单

| 检查项 | 预期结果 | 状态 |
|--------|----------|------|
| Server 启动 | 日志显示 `gRPC Server 启动，监听端口 :50051` | ✅ |
| Agent 连接 | Server 日志显示 `[Transfer] Agent 连接 agent_id=xxx` | ✅ |
| Agent 无错误 | Agent 日志无 ERROR 级别输出 | ✅ |
| 心跳发送 | Agent 持续发送心跳包（每 100ms） | ✅ |
| 断开连接 | Server 日志显示 `[Transfer] Agent 断开 agent_id=xxx` | ✅ |

## 心跳包内容

心跳包 (`PackagedData`) 包含以下字段：
- `agent_id`: Agent 唯一标识（基于 DMI/MAC 或 machine-id 生成）
- `hostname`: 主机名
- `ipv4`: IP 地址列表
- `version`: Agent 版本号
- `product`: 产品名称（cloudsec-agent）
- `records`: 数据记录（心跳包可为空）

## 故障排查

| 问题 | 解决方案 |
|------|----------|
| 连接失败 | 检查 `config.yaml` 中 server 地址是否正确 |
| 配置文件找不到 | 确保在 agent 目录下创建 `config.yaml` 或放置于 `/etc/cloudsec-agent/config.yaml` |
| 端口冲突 | 修改 server 启动端口 `-port 50052`，同时更新 `config.yaml` |
| Agent ID 为空 | 检查 `/etc/machine-id` 或 DMI 信息是否可读 |
