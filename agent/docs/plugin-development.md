# 插件开发指南

本文档描述如何开发 Agent 插件，包括接口规范、生命周期和数据上报方式。

---

## 一、插件架构概述

### 1.1 插件类型

| 插件类型 | 功能 | 协议 | 示例 |
|----------|------|------|------|
| 数据采集 | 定时采集系统信息 | 优化格式 | collector |
| 任务执行 | 接收任务并返回结果 | 优化格式 | baseline |
| 实时检测 | 日志分析/事件检测 | 标准 Protobuf | detector |
| 内核监控 | eBPF 系统调用监控 | 标准 Protobuf | ebpf_base_detector |
| 网络入侵检测 | 网络流量分析 | 标准 Protobuf | nids |
| 病毒扫描 | 恶意文件检测 | 标准 Protobuf | scanner |

### 1.2 通信机制

```
Agent 主进程                          Plugin 子进程
     │                                      │
     │ ←── fd 4 (Plugin 输出) ─────────────│ SendRecord()
     │                                      │
     │ ──── fd 3 (Plugin 输入) ───────────→│ ReceiveTask()
     │                                      │
```

插件通过两个文件描述符与 Agent 通信：
- **fd 3**: 输入管道，接收 Agent 下发的任务
- **fd 4**: 输出管道，向 Agent 发送数据

---

## 二、开发环境

### 2.1 目录结构

```
business_plugins/
├── lib/                    # 公共库
│   ├── bridge.proto        # Protobuf 定义
│   ├── bridge.pb.go        # 生成的代码
│   └── client.go           # 客户端接口
│
├── collector/              # 采集插件示例
│   ├── main.go
│   └── engine/             # 调度引擎
│
├── baseline/               # 基线检查插件
├── detector/               # 威胁检测插件
├── ebpf_base_detector/     # eBPF 驱动插件
├── nids/                   # 网络入侵检测插件
└── scanner/                # 病毒扫描插件
```

### 2.2 依赖库

```go
import (
    businessplugins "company/agent/business_plugins/lib"
)
```

---

## 三、核心接口

### 3.1 Client 接口

**文件：** `business_plugins/lib/client.go`

```go
// 创建客户端
client := businessplugins.New()

// 发送数据记录
err := client.SendRecord(&businessplugins.Record{
    DataType:  5050,
    Timestamp: time.Now().Unix(),
    Data: &businessplugins.Payload{
        Fields: map[string]string{
            "key": "value",
        },
    },
})

// 接收任务
task, err := client.ReceiveTask()

// 刷新缓冲区
client.Flush()

// 关闭连接
client.Close()
```

### 3.2 数据结构

**Record - 数据记录：**
```protobuf
message Record {
  int32 data_type = 1;      // 数据类型标识
  int64 timestamp = 2;      // 时间戳
  Payload data = 3;         // 数据内容
}

message Payload {
  map<string, string> fields = 1;  // 键值对数据
}
```

**Task - 任务：**
```protobuf
message Task {
  int32 data_type = 1;      // 任务类型
  string object_name = 2;   // 对象名称
  string data = 3;          // 任务数据 (JSON)
  string token = 4;         // 任务令牌
}
```

---

## 四、插件开发模式

### 4.1 数据采集插件

适用于定时采集系统信息的场景。

**示例：进程采集器**

```go
package main

import (
    "time"
    businessplugins "company/agent/business_plugins/lib"
)

const DataTypeProcess = 5050

func main() {
    client := businessplugins.New()
    defer client.Close()

    ticker := time.NewTicker(time.Hour)
    for range ticker.C {
        collectProcesses(client)
    }
}

func collectProcesses(client *businessplugins.Client) {
    // 采集进程信息
    processes := getProcessList()

    for _, p := range processes {
        record := &businessplugins.Record{
            DataType:  DataTypeProcess,
            Timestamp: time.Now().Unix(),
            Data: &businessplugins.Payload{
                Fields: map[string]string{
                    "pid":     p.PID,
                    "name":    p.Name,
                    "cmdline": p.Cmdline,
                    "exe":     p.Exe,
                },
            },
        }
        client.SendRecord(record)
    }
    client.Flush()
}
```

### 4.2 任务执行插件

适用于接收 Server 任务并返回结果的场景。

**示例：基线检查插件**

```go
package main

import (
    "encoding/json"
    businessplugins "company/agent/business_plugins/lib"
)

const (
    DataTypeBaselineResult = 8000
    DataTypeTaskStatus     = 8010
)

func main() {
    client := businessplugins.New()
    defer client.Close()

    for {
        // 接收任务
        task, err := client.ReceiveTask()
        if err != nil {
            continue
        }

        // 执行基线检查
        result := executeBaseline(task)

        // 发送结果
        client.SendRecord(&businessplugins.Record{
            DataType:  DataTypeBaselineResult,
            Timestamp: time.Now().Unix(),
            Data: &businessplugins.Payload{
                Fields: map[string]string{
                    "token":  task.Token,
                    "result": result,
                },
            },
        })

        // 发送任务状态
        client.SendRecord(&businessplugins.Record{
            DataType:  DataTypeTaskStatus,
            Timestamp: time.Now().Unix(),
            Data: &businessplugins.Payload{
                Fields: map[string]string{
                    "token":  task.Token,
                    "status": "succeed",
                },
            },
        })
        client.Flush()
    }
}
```

### 4.3 实时检测插件

适用于日志监控和事件检测的场景。

**示例：暴力破解检测**

```go
package main

import (
    "bufio"
    "os"
    businessplugins "company/agent/business_plugins/lib"
)

const DataTypeSSHBruteForce = 6001

func main() {
    client := businessplugins.New()
    defer client.Close()

    // 监控日志文件
    file, _ := os.Open("/var/log/auth.log")
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()

        // 检测暴力破解
        if alert := detectBruteForce(line); alert != nil {
            client.SendRecord(&businessplugins.Record{
                DataType:  DataTypeSSHBruteForce,
                Timestamp: time.Now().Unix(),
                Data: &businessplugins.Payload{
                    Fields: map[string]string{
                        "source_ip": alert.SourceIP,
                        "count":     alert.Count,
                        "rule":      alert.Rule,
                    },
                },
            })
            client.Flush()
        }
    }
}
```

---

## 五、Handler 接口（采集插件）

collector 插件使用 Handler 接口和 Engine 调度器。

### 5.1 Handler 接口定义

```go
type Handler interface {
    Handle(c *businessplugins.Client, cache *Cache, seq string)
    Name() string
    DataType() int
}
```

### 5.2 实现示例

```go
type ProcessHandler struct{}

func (h *ProcessHandler) Name() string {
    return "process"
}

func (h *ProcessHandler) DataType() int {
    return 5050
}

func (h *ProcessHandler) Handle(c *businessplugins.Client, cache *Cache, seq string) {
    // 采集逻辑
    processes := collectProcesses()
    for _, p := range processes {
        c.SendRecord(buildRecord(p))
    }
}
```

### 5.3 注册 Handler

```go
func main() {
    client := businessplugins.New()
    engine := engine.New(client, logger)

    // 注册采集器，设置执行间隔
    engine.AddHandler(time.Hour, &ProcessHandler{})
    engine.AddHandler(time.Hour*6, &ServiceHandler{})
    engine.AddHandler(time.Hour*24, &SoftwareHandler{})

    engine.Run()
}
```

---

## 六、编译和部署

### 6.1 编译插件

```bash
cd business_plugins/your_plugin
go build -o your_plugin
```

### 6.2 部署结构

```
/opt/cloudsec/agent/plugins/
├── your_plugin/
│   ├── your_plugin          # 可执行文件
│   └── config/              # 配置目录（可选）
│       └── rules.yaml
```

### 6.3 插件配置

在 Agent 主配置中指定插件目录：

```yaml
plugins_directory: "/opt/cloudsec/agent/plugins"
```

---

## 七、调试技巧

### 7.1 本地测试

使用 Standalone 模式测试插件：

```bash
# 仅加载你的插件
cd /opt/cloudsec/agent
sudo ./bin/agent -standalone -plugins=your_plugin -output=/opt/cloudsec/agent/logs/agent.log -test
```

### 7.2 日志输出

插件内使用 zap 日志：

```go
import "go.uber.org/zap"

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    logger.Info("plugin started")
}
```

### 7.3 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 插件无法启动 | 文件权限 | `chmod +x plugin` |
| 数据未上报 | 未调用 Flush | 确保发送后调用 `client.Flush()` |
| 任务未收到 | 阻塞在 ReceiveTask | 检查 Agent 是否连接 Server |

---

## 八、DataType 分配

| 范围 | 用途 |
|------|------|
| 5050-5099 | Collector 采集数据 |
| 5100-5199 | Collector 状态/日志 |
| 59-65 | ebpf_base_detector eBPF 事件 |
| 6001-6009 | 安全告警（多插件共用） |
| 6010-6099 | Detector 状态/日志 |
| 6050-6061 | Scanner 扫描结果 |
| 6007 | NIDS 网络攻击告警 |
| 7001-7004 | 容器安全告警 |
| 8000-8099 | Baseline 检查结果 |

新插件开发时，请向项目负责人申请 DataType 范围。详细的字段定义见 [DataType 详细说明](data-types.md)。

---

## 相关文档

- [架构设计文档](architecture.md) — 插件架构和通信机制
- [DataType 详细说明](data-types.md) — 各 DataType 的完整字段定义
- [配置文件详解](configuration.md) — 插件配置文件格式
- [功能测试文档](testing.md) — 插件测试流程
