# Agent 架构设计文档

本文档描述 Agent 的整体架构、模块职责和数据流。

---

## 一、系统架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                         Agent 主进程                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌──────────────────┐   │
│  │ config  │  │  agent  │  │ buffer  │  │    transport     │   │
│  │ 配置管理 │  │ 状态管理 │  │ 数据缓冲 │  │  (或 standalone) │   │
│  └─────────┘  └─────────┘  └─────────┘  └──────────────────┘   │
│                      ↑                           ↑              │
│                      │ IPC (管道)                │ gRPC         │
│  ┌───────────────────┴───────────────────────────┘              │
│  │                 plugin 插件管理                               │
│  └──────────────────────────────────────────────────────────────┘
│         ↓              ↓              ↓              ↓          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │collector │  │ baseline │  │ detector │  │ebpf_base_detector│
│  │ 资产采集  │  │ 基线检查  │  │ 威胁检测  │  │ eBPF监控        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│         ↓              ↓                                        │
│  ┌──────────┐  ┌──────────┐                                     │
│  │   nids   │  │ scanner  │                                     │
│  │ 网络入侵  │  │ 病毒扫描  │                                     │
│  └──────────┘  └──────────┘                                     │
└─────────────────────────────────────────────────────────────────┘
                              ↓
                    ┌──────────────────┐
                    │   gRPC Server    │
                    │    (hcids)       │
                    └──────────────────┘
```

---

## 二、启动流程

```
main.go
   │
   ├─ 1. 命令行参数解析
   │     -config: 配置文件路径
   │     -test: 测试模式（固定 Agent ID）
   │     -standalone: 独立模式
   │     -output: 输出方式 (stderr/文件路径)
   │     -plugins: 加载的插件列表
   │
   ├─ 2. 加载配置 (config.Init)
   │
   ├─ 3. 初始化日志 (zap，依赖配置中的日志目录)
   │
   ├─ 4. 启动守护进程
   │     ├─ plugin.Startup() - 插件管理
   │     └─ transport.StartTransfer() 或 standalone.StartOutputHandler()
   │
   ├─ 5. 信号监听 (SIGTERM/SIGINT)
   │
   └─ 6. 优雅关闭 (5秒超时)
```

---

## 三、数据流

```
┌──────────────────────────────────────────────────────────────┐
│                      PLUGIN (子进程)                          │
│  Collector/Baseline/Detector/ebpf_base_detector/NIDS/Scanner   │
│  └─> 生成 EncodedRecord                                       │
│       └─> 写入 rx 管道                                        │
└──────────────────────────────────────────────────────────────┘
                           ↓
                    [rx 管道读取]
                           ↓
┌──────────────────────────────────────────────────────────────┐
│                    BUFFER (内存缓冲)                          │
│  - 固定容量缓冲区: [8192]*EncodedRecord                        │
│  - WriteEncodedRecord() / ReadEncodedRecords()               │
│  - 互斥锁 (sync.Mutex) 保证线程安全                            │
└──────────────────────────────────────────────────────────────┘
                           ↓
                    [100ms 定时轮询]
                           ↓
┌──────────────────────────────────────────────────────────────┐
│                  TRANSPORT (传输层)                           │
├─ handleSend(): 读取 buffer → 组装 PackagedData → 发送         │
│  └─ handleReceive(): 接收 Command → 转发 Task 给插件          │
└──────────────────────────────────────────────────────────────┘
                           ↓
                    [gRPC 双向流]
                           ↓
                      ┌──────────┐
                      │  SERVER  │
                      └──────────┘
```

---

## 四、核心模块

### 4.1 config/ - 配置管理

**职责：** 加载和管理全局配置

**关键结构：**
```go
type Config struct {
    Server           string              // gRPC 服务器地址
    ConnectTimeout   int                 // 连接超时(秒)
    WorkingDirectory string              // 工作目录
    PluginsDirectory string              // 插件目录
    LogDirectory     string              // 日志目录
    RetryMaxCount    int                 // 最大重试次数
    RetryInterval    int                 // 重试间隔(秒)
    Standalone       *StandaloneConfig   // 独立模式配置
    Log              *LogConfig          // 日志配置
}
```

**配置文件优先级：**
1. 命令行 `-config` 参数
2. `/etc/cloudsec-agent/agent.yaml`
3. 当前目录 `agent.yaml`

### 4.2 agent/ - 状态管理

**职责：** Agent ID 生成和运行状态管理

**关键功能：**
- `GenerateIDFromDMIAndMAC()` - 基于硬件信息生成 ID（优先）
- `GenerateIDFromMachineID()` - 基于 machine-id 生成（回退）
- `SetRunning()` / `SetAbnormal()` - 状态管理

### 4.3 plugin/ - 插件管理

**职责：** 插件生命周期管理和进程间通信

**插件生命周期：**
```
Load() ──────────────────────────────────────────> Shutdown()
   │                                                    │
   ├─ 检查插件文件                                       ├─ 关闭管道
   ├─ 创建双向管道                                       ├─ 等待退出(10s)
   ├─ 启动子进程                                        └─ 强制杀死
   └─ 启动 3 个 goroutine:
       ├─ 等待进程退出
       ├─ 接收插件数据
       └─ 发送任务给插件
```

**两种通信协议：**
| 协议 | 格式 | 适用插件 |
|------|------|----------|
| 标准 Protobuf | `[4字节长度][protobuf]` | ebpf_base_detector, nids, scanner, detector |
| 优化格式 | Varint 编码字段序列 | collector, baseline |

### 4.4 buffer/ - 数据缓冲

**职责：** 固定容量缓冲区，解耦插件数据生产和网络传输

**特点：**
- 固定容量 8192 条记录
- 互斥锁 (sync.Mutex) 保证线程安全
- 100ms 批量读取，减少网络开销

### 4.5 transport/ - 网络传输

**职责：** gRPC 双向流通信

**两个工作协程：**
- `handleSend()` - 100ms 轮询读取 buffer，组装 PackagedData 发送
- `handleReceive()` - 接收 Command，转发 Task 给插件

**重试策略：**
- 传输守护进程内部使用硬编码 6 次重试（`retries > 5`）、每次间隔 5 秒
- 配置文件中的 `RetryMaxCount` / `RetryInterval` 作用于连接层

### 4.6 standalone/ - 独立模式

**职责：** 替代 transport，本地输出检测结果

**输出方式：**
- `stderr` - 通过 zap 日志输出到标准错误
- 文件路径 - JSON 格式写入指定文件

---

## 五、插件架构

### 5.1 插件类型

| 插件 | 功能 | 数据类型 | 协议 |
|------|------|----------|------|
| collector | 资产采集 | 5050-5062 | 优化格式 |
| baseline | 基线检查 | 8000, 8010 | 优化格式 |
| detector | 威胁检测 | 6001, 6002, 6005 | 标准 Protobuf |
| ebpf_base_detector | eBPF 进程监控 | 59-65, 6003-6009, 7001-7004 | 标准 Protobuf |
| nids | 网络入侵检测 | 6007 | 标准 Protobuf |
| scanner | 病毒扫描 | 6050-6061 | 标准 Protobuf |

### 5.2 插件通信

```
Agent 主进程                          Plugin 子进程
     │                                      │
     │ ←── rx 管道 (Agent 接收) ────────────│ 数据上报
     │                                      │
     │ ──── tx 管道 (Agent 发送) ──────────→│ 任务下发
     │                                      │
```

**文件描述符：**
- fd 3: 输入管道（插件接收任务）
- fd 4: 输出管道（插件发送数据）

---

## 六、关键数据结构

### 6.1 EncodedRecord

```protobuf
message EncodedRecord {
  int32 data_type = 1;    // 数据类型标识
  int64 timestamp = 2;    // 时间戳
  bytes data = 3;         // 序列化的 payload
}
```

### 6.2 PackagedData

```protobuf
message PackagedData {
  repeated EncodedRecord records = 1;
  string agent_id = 2;
  repeated string ipv4 = 3;
  string hostname = 4;
  string version = 5;
  string product = 6;
  string mac_addr = 7;
  string os_type = 8;
  string os_version = 9;
}
```

### 6.3 Task

```protobuf
message Task {
  int32 data_type = 1;
  string object_name = 2;
  string data = 3;
  string token = 4;
}
```

---

## 七、设计要点

### 并发模型
- 插件管理：1 个主守护进程 + N 个插件进程各 3 个协程
- 传输层：1 个守护进程 + 2 个工作协程 (send/receive)
- 缓冲区：线程安全的固定容量缓冲 (sync.Mutex)

### 资源管理
- 插件生命周期完整管理（加载 → 运行 → 关闭）
- 进程组管理（Setpgid 便于批量杀死）
- 优雅关闭和强制清理

### 两种运行模式
1. **Server 模式**（默认）：gRPC 传输 + 远程命令控制
2. **Standalone 模式**：本地输出 + 无网络依赖

---

## 相关文档

- [DataType 详细说明](data-types.md) — 各插件的 DataType 定义和数据字段
- [配置文件详解](configuration.md) — Agent 及插件配置项说明
- [插件开发指南](plugin-development.md) — 插件接口规范和开发模式
- [编译部署文档](build-deploy.md) — 编译、打包和部署流程
- [gRPC 协议说明](grpc-protocol.md) — Agent 与 Server 的通信协议
