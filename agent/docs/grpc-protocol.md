# Agent-Server gRPC 通信协议文档

## 1. 概述

### 1.1 通信架构

Agent 和 Server 之间采用 gRPC 双向流（Bidirectional Streaming）进行通信。

```
┌─────────┐                          ┌─────────┐
│  Agent  │◄────── gRPC Stream ─────►│  Server │
└─────────┘                          └─────────┘
     │                                    │
     │  PackagedData (Agent → Server)     │
     │  Command (Server → Agent)          │
     │                                    │
```

### 1.2 连接特性

- **协议**: gRPC over HTTP/2
- **流模式**: 双向流（Bidirectional Streaming）
- **默认端口**: 50051
- **心跳**: Agent 每 100ms 发送一次 PackagedData（可能为空数据）

---

## 2. 服务定义

```protobuf
syntax = "proto3";
package grpc;

// Transfer 数据传输服务（双向流）
service Transfer {
  rpc Transfer(stream PackagedData) returns (stream Command) {}
}
```

---

## 3. 消息结构

### 3.1 Agent → Server 消息

#### PackagedData

Agent 发送给 Server 的数据包，包含采集数据和 Agent 元信息。

```protobuf
message PackagedData {
  repeated EncodedRecord records = 1;  // 编码后的数据记录列表
  string agent_id = 2;                  // Agent 唯一标识符
  repeated string ipv4 = 3;             // IPv4 地址列表
  string hostname = 4;                  // 主机名
  string version = 5;                   // Agent 版本号
  string product = 6;                   // 产品名称
  string mac_addr = 7;                  // MAC 地址
  string os_type = 8;                   // 操作系统类型 (linux/windows)
  string os_version = 9;                // 操作系统版本
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| records | repeated EncodedRecord | 否 | 编码后的数据记录列表 |
| agent_id | string | 是 | Agent 唯一标识符（UUID 格式） |
| ipv4 | repeated string | 否 | IPv4 地址列表 |
| hostname | string | 是 | 主机名 |
| version | string | 否 | Agent 版本号 |
| product | string | 是 | 产品名称（cloudsec-agent） |
| mac_addr | string | 否 | MAC 地址 |
| os_type | string | 否 | 操作系统类型 (linux/windows) |
| os_version | string | 否 | 操作系统版本 |

#### EncodedRecord

编码后的记录，用于传输采集的数据。

```protobuf
message EncodedRecord {
  int32 data_type = 1;   // 数据类型
  int64 timestamp = 2;   // Unix 时间戳（秒）
  bytes data = 3;        // 序列化后的数据（Protobuf）
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| data_type | int32 | 数据类型标识 |
| timestamp | int64 | Unix 时间戳（秒） |
| data | bytes | Protobuf 序列化的 Payload 数据 |

#### Payload

数据负载，键值对格式。

```protobuf
message Payload {
  map<string, string> fields = 1;  // 字段映射
}
```

---

### 3.2 Server → Agent 消息

#### Command

Server 下发给 Agent 的命令。

```protobuf
message Command {
  int32 ctrl = 1;              // 控制字段
  Task task = 2;               // 任务（可选）
  repeated Config configs = 3; // 插件配置列表（可选）
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| ctrl | int32 | 控制字段（保留） |
| task | Task | 任务（可选，与 configs 二选一） |
| configs | repeated Config | 插件配置列表（可选） |

#### Task

任务结构，用于下发采集任务或控制命令。

```protobuf
message Task {
  int32 data_type = 1;    // 任务数据类型
  string object_name = 2; // 目标对象名称（Agent 或插件名称）
  string data = 3;        // 任务数据（JSON 格式）
  string token = 4;       // 任务令牌
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| data_type | int32 | 是 | 任务数据类型（见 DataType 定义） |
| object_name | string | 是 | 目标对象（"cloudsec-agent" 或插件名如 "collector"） |
| data | string | 否 | 任务数据（JSON 格式） |
| token | string | 否 | 任务令牌，用于追踪任务结果 |

#### Config

插件配置，用于下发插件更新。

```protobuf
message Config {
  string name = 1;                    // 插件名称
  string type = 2;                    // 插件类型
  string version = 3;                 // 插件版本号
  string sha256 = 4;                  // SHA256 校验值
  string signature = 5;               // 插件签名
  repeated string download_urls = 6;  // 下载地址列表
  string detail = 7;                  // 详细信息（JSON 格式）
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 插件名称（如 "collector"） |
| type | string | 插件类型 |
| version | string | 插件版本号 |
| sha256 | string | 插件文件的 SHA256 校验值 |
| signature | string | 插件签名 |
| download_urls | repeated string | 插件下载地址列表 |
| detail | string | 详细配置信息（JSON 格式） |

---

## 4. DataType 定义

### 4.1 Agent 级别命令

| DataType | 名称 | 说明 |
|----------|------|------|
| 1060 | AgentShutdown | 关闭 Agent |

### 4.2 Collector 采集任务

| DataType | Handler | 说明 |
|----------|---------|------|
| 5050 | ProcessHandler | 进程采集 |
| 5051 | PortHandler | 端口采集 |
| 5052 | UserHandler | 用户账号采集 |
| 5054 | ServiceHandler | 系统服务采集 |
| 5055 | SoftwareHandler | 软件包采集 |
| 5056 | ContainerHandler | 容器采集 |
| 5057 | EnvSuspiciousHandler | 可疑环境变量检测 |
| 5058 | ImageHandler | 镜像资产采集 |
| 5059 | ImagePackageHandler | 镜像软件包采集 |
| 5060 | WebServiceHandler | Web 服务采集 |
| 5061 | DatabaseHandler | 数据库服务采集 |
| 5062 | KmodHandler | 内核模块采集 |

### 4.3 响应类型

| DataType | 名称 | 说明 |
|----------|------|------|
| 5100 | TaskResult | 任务执行结果响应 |

**TaskResult Payload 字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| status | string | "succeed" 或 "failed" |
| msg | string | 错误信息（失败时） |
| token | string | 任务令牌（与请求中的 token 对应） |

---

## 5. 通信流程

### 5.1 Agent 连接流程

```
Agent                                   Server
  │                                       │
  │──── 建立 gRPC 连接 ──────────────────>│
  │                                       │
  │<──── Transfer 双向流建立 ─────────────│
  │                                       │
  │──── PackagedData (心跳/注册) ────────>│  首次发送，Server 注册 Agent
  │                                       │
  │<──── Command (configs) ──────────────│  下发插件配置（可选）
  │                                       │
  │──── PackagedData (数据) ────────────>│  持续发送采集数据
  │         ...                           │
  │<──── Command (task) ─────────────────│  下发任务（按需）
  │                                       │
  │──── PackagedData (结果) ────────────>│  返回任务结果
  │                                       │
```

### 5.2 任务下发流程

```
HTTP Client                  Server                      Agent
     │                          │                          │
     │── POST /api/task ──────>│                          │
     │                          │                          │
     │<── 200 OK ──────────────│                          │
     │                          │                          │
     │                          │──── Command(Task) ─────>│
     │                          │                          │
     │                          │                          │ 执行采集任务
     │                          │                          │
     │                          │<── PackagedData ────────│ 返回采集数据
     │                          │    (data_type=5050...)   │
     │                          │                          │
     │                          │<── PackagedData ────────│ 返回任务结果
     │                          │    (data_type=5100)      │
     │                          │                          │
```

### 5.3 Agent 关闭流程

```
Server                                  Agent
  │                                       │
  │──── Command(Task) ──────────────────>│  data_type=1060
  │     object_name="cloudsec-agent"      │  object_name="cloudsec-agent"
  │                                       │
  │                                       │  Agent 优雅关闭
  │<──── 连接关闭 ────────────────────────│
  │                                       │
```

---

## 6. 错误处理

### 6.1 连接断开重试

Agent 端连接断开后会自动重试：

- **重试间隔**: 5 秒（可配置）
- **最大重试次数**: 10 次（可配置）
- **连接超时**: 30 秒（可配置）

> **注意：** 传输守护进程 (`transport/transfer.go`) 内部使用硬编码的 5 次重试、每次间隔 5 秒。配置文件中的 `retry_max_count` / `retry_interval` 作用于连接层。

### 6.2 任务执行失败响应

当任务执行失败时，Agent 返回 DataType=5100 的记录：

```json
{
  "status": "failed",
  "msg": "the data_type hasn't been implemented",
  "token": "task-123"
}
```

---

## 7. 示例

### 7.1 下发进程采集任务

**HTTP 请求：**

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "6bb9735d-66ee-556a-8981-62d127daf308",
    "object_name": "collector",
    "data_type": 5050,
    "data": "{}",
    "token": "task-process-001"
  }'
```

**生成的 gRPC Command：**

```
Command {
  task: Task {
    data_type: 5050
    object_name: "collector"
    data: "{}"
    token: "task-process-001"
  }
}
```

### 7.2 关闭 Agent

**HTTP 请求：**

```bash
curl -X POST http://localhost:8080/api/task \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "6bb9735d-66ee-556a-8981-62d127daf308",
    "object_name": "cloudsec-agent",
    "data_type": 1060
  }'
```

---

## 8. 配置参数

### 8.1 Agent 端配置 (config.yaml)

```yaml
server: "localhost:50051"    # Server 地址
connect_timeout: 30          # 连接超时（秒）
retry_max_count: 10          # 最大重试次数
retry_interval: 5            # 重试间隔（秒）
```

### 8.2 Server 端配置 (config.yaml)

```yaml
server:
  port: 50051                # gRPC 端口
  http_port: 8080            # HTTP API 端口
  max_recv_msg_size: 16      # 最大接收消息大小（MB）
  max_send_msg_size: 16      # 最大发送消息大小（MB）
```
