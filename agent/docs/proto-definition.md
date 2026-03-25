# Proto 定义文档

本文档详细说明了 Agent 与 Server 通信的 protobuf 消息定义和服务接口。

## 目录

- [消息类型](#消息类型)
  - [PackagedData](#packageddata)
  - [EncodedRecord](#encodedrecord)
  - [Record](#record)
  - [Payload](#payload)
  - [Command](#command)
  - [Task](#task)
  - [Config](#config)
- [服务定义](#服务定义)
  - [Transfer 服务](#transfer-服务)
- [数据流转](#数据流转)
- [使用示例](#使用示例)

## 消息类型

### PackagedData

Agent 发送给 Server 的数据包，包含批量采集的数据记录和 Agent 元信息。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `records` | `repeated EncodedRecord` | 编码后的数据记录列表（批量发送） |
| `agent_id` | `string` | Agent 唯一标识符 |
| `ipv4` | `repeated string` | IPv4 地址列表（不区分内网/公网） |
| `hostname` | `string` | 主机名 |
| `version` | `string` | Agent 版本号 |
| `product` | `string` | 产品名称（如 "cloudsec-agent"） |
| `mac_addr` | `string` | MAC 地址 |
| `os_type` | `string` | 操作系统类型 (linux/windows) |
| `os_version` | `string` | 操作系统版本 |

**使用场景：**
- Agent 定期批量发送采集的数据
- 每次发送包含多个 `EncodedRecord`
- 同时携带 Agent 元信息，便于 Server 识别和处理

### EncodedRecord

编码后的记录，用于传输层。`Data` 字段是 `Payload` 序列化后的字节数组。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `data_type` | `int32` | 数据类型（用于区分不同的采集数据类型） |
| `timestamp` | `int64` | 时间戳（Unix 时间戳，秒） |
| `data` | `bytes` | 序列化后的数据（`Payload` 的 protobuf 序列化结果） |

**使用场景：**
- 存储在缓冲区中
- 批量打包到 `PackagedData` 中发送
- 已序列化，适合网络传输

### Record

未编码的记录，Agent 内部使用。`Data` 字段是结构化的 `Payload` 对象。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `data_type` | `int32` | 数据类型 |
| `timestamp` | `int64` | 时间戳 |
| `data` | `Payload` | 结构化的数据负载（键值对） |

**使用场景：**
- 插件生成数据时使用
- Agent 内部处理时使用
- 需要转换为 `EncodedRecord` 后才能发送

**转换关系：**
```
Record (Payload) → 序列化 → EncodedRecord ([]byte) → 打包 → PackagedData
```

### Payload

数据负载，使用键值对格式存储数据。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `fields` | `map<string, string>` | 字段映射（键值对） |

**使用场景：**
- 存储采集到的结构化数据
- 所有字段值都是字符串类型
- 序列化后存储在 `EncodedRecord.Data` 中

### Command

Server 下发给 Agent 的命令结构体。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ctrl` | `int32` | 控制字段（预留，用于未来扩展控制命令） |
| `task` | `Task` | 任务（可选）<br>- 如果 `ObjectName == agent.Product`，则是给 Agent 的任务（如 Agent 更新、设置元数据等）<br>- 否则是给指定插件的任务（`ObjectName` 为插件名称） |
| `configs` | `repeated Config` | 插件配置列表（可选）<br>用于插件安装、更新、配置等操作 |

**使用场景：**
- Server 通过 `Transfer` 服务的双向流下发命令
- 可以同时包含任务和配置
- Agent 接收后根据内容进行处理

### Task

任务结构，用于 Server 向 Agent 或插件下发任务。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `data_type` | `int32` | 任务数据类型（用于区分不同的任务类型） |
| `object_name` | `string` | 目标对象名称<br>- 如果等于 `agent.Product`，则任务发送给 Agent 本身<br>- 否则为插件名称，任务发送给对应的插件 |
| `data` | `string` | 任务数据（JSON 格式的字符串） |
| `token` | `string` | 任务令牌（用于任务追踪和日志关联） |

**任务类型示例：**

| data_type | 说明 | 目标 |
|-----------|------|------|
| 1050 | 文件上传任务 | Agent |
| 1051 | 设置元数据（IDC、Region） | Agent |
| 1060 | Agent 关闭 | Agent |
| 其他 | 插件特定任务 | 插件 |

### Config

插件配置，用于插件安装、更新、配置等操作。

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | `string` | 插件名称（唯一标识） |
| `type` | `string` | 插件类型（如 "collector", "baseline" 等） |
| `version` | `string` | 插件版本号 |
| `sha256` | `string` | 插件文件的 SHA256 校验值（用于验证文件完整性） |
| `signature` | `string` | 插件签名（用于验证插件完整性） |
| `download_urls` | `repeated string` | 插件下载地址列表（支持多个下载源） |
| `detail` | `string` | 插件详细信息（JSON 格式的字符串，包含额外配置） |

**使用场景：**
- 插件安装：提供下载地址和校验信息
- 插件更新：版本号不同时触发更新
- 插件配置：通过 `detail` 字段传递配置信息

## 服务定义

### Transfer 服务

数据传输服务，提供 Agent 与 Server 之间的双向流通信。

**服务定义：**
```protobuf
service Transfer {
  rpc Transfer(stream PackagedData) returns (stream Command) {}
}
```

**工作流程：**

1. **建立连接**：Agent 作为客户端，Server 作为服务端，建立双向流连接
2. **数据发送**：Agent 持续发送 `PackagedData`（包含采集的数据）
3. **命令接收**：Server 可以随时通过 `Command` 流下发任务或配置
4. **命令处理**：Agent 接收 `Command` 后处理任务或更新插件配置
5. **连接保持**：连接保持活跃，直到一方主动关闭或发生错误

**双向流特性：**
- **客户端流（Client Stream）**：Agent 可以持续发送多个 `PackagedData`
- **服务端流（Server Stream）**：Server 可以持续发送多个 `Command`
- **异步通信**：发送和接收可以独立进行，互不阻塞

**生成的接口：**

**客户端接口（Agent 使用）：**
```go
type TransferClient interface {
    Transfer(ctx context.Context, opts ...grpc.CallOption) (Transfer_TransferClient, error)
}

type Transfer_TransferClient interface {
    Send(*PackagedData) error  // 发送数据
    Recv() (*Command, error)   // 接收命令
    grpc.ClientStream
}
```

**服务端接口（Server 使用）：**
```go
type TransferServer interface {
    Transfer(Transfer_TransferServer) error
}

type Transfer_TransferServer interface {
    Send(*Command) error        // 发送命令
    Recv() (*PackagedData, error) // 接收数据
    grpc.ServerStream
}
```

## 数据流转

### Agent 端数据流转

```
插件生成数据
    ↓
Record (Payload)
    ↓
序列化 Payload
    ↓
EncodedRecord ([]byte)
    ↓
写入缓冲区 buffer.WriteEncodedRecord()
    ↓
批量读取 buffer.ReadEncodedRecords()
    ↓
打包 PackagedData
    ↓
通过 Transfer 服务发送 stream.Send()
```

### Server 端命令流转

```
Server 生成命令
    ↓
Command (Task/Config)
    ↓
通过 Transfer 服务发送 stream.Send()
    ↓
Agent 接收 stream.Recv()
    ↓
处理命令
    ├─ Task → 发送给 Agent 或插件
    └─ Config → 更新插件配置
```

## 使用示例

### Agent 端：发送数据

```go
// 1. 建立连接
conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
if err != nil {
    return err
}
defer conn.Close()

// 2. 创建客户端
client := proto.NewTransferClient(conn)

// 3. 建立双向流
stream, err := client.Transfer(ctx)
if err != nil {
    return err
}
defer stream.CloseSend()

// 4. 启动发送协程
go func() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // 从缓冲区读取数据
            records := buffer.ReadEncodedRecords()

            // 打包数据（即使 records 为空也发送，作为心跳）
            pkg := &proto.PackagedData{
                Records:   records,
                AgentId:   agent.ID,
                Ipv4:      host.IPv4.Load().([]string),
                Hostname:  host.Name.Load().(string),
                Version:   agent.Version,
                Product:   agent.Product,
                MacAddr:   host.MACAddr,
                OsType:    host.OSType,
                OsVersion: host.OSVersion,
            }
            
            // 发送数据
            if err := stream.Send(pkg); err != nil {
                return
            }
            
            // 归还记录到对象池
            buffer.PutEncodedRecords(records)
        }
    }
}()

// 5. 接收命令
for {
    cmd, err := stream.Recv()
    if err != nil {
        return err
    }
    
    // 处理命令
    handleCommand(cmd)
}
```

### Agent 端：处理命令

```go
func handleCommand(cmd *proto.Command) {
    // 处理任务
    if cmd.Task != nil {
        if cmd.Task.ObjectName == agent.Product {
            // Agent 任务
            handleAgentTask(cmd.Task)
        } else {
            // 插件任务
            plg, ok := plugin.Get(cmd.Task.ObjectName)
            if ok {
                plg.SendTask(*cmd.Task)
            }
        }
    }
    
    // 处理配置
    if len(cmd.Configs) > 0 {
        cfgs := make(map[string]*proto.Config)
        for _, cfg := range cmd.Configs {
            cfgs[cfg.Name] = cfg
        }
        plugin.Sync(cfgs)
    }
}
```

### Record 到 EncodedRecord 的转换

```go
func WriteRecord(rec *proto.Record) error {
    // 1. 从对象池获取 EncodedRecord
    erec := buffer.GetEncodedRecord()
    
    // 2. 复制基本信息
    erec.DataType = rec.DataType
    erec.Timestamp = rec.Timestamp
    
    // 3. 序列化 Payload
    if cap(erec.Data) < rec.Data.Size() {
        erec.Data = make([]byte, rec.Data.Size())
    } else {
        erec.Data = erec.Data[:rec.Data.Size()]
    }
    _, err := rec.Data.MarshalTo(erec.Data)
    if err != nil {
        return err
    }
    
    // 4. 写入缓冲区
    return buffer.WriteEncodedRecord(erec)
}
```

## 注意事项

1. **字段编号**：protobuf 字段编号一旦确定，不应随意修改，以保证向后兼容
2. **序列化格式**：Agent 端使用 gogo protobuf 生成代码；Server 端使用标准 protobuf
3. **双向流**：Transfer 服务是双向流，发送和接收可以同时进行
4. **错误处理**：流关闭或错误时，需要重新建立连接
