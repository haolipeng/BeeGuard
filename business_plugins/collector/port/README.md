# Port 端口信息采集功能

## 功能说明

Port 功能用于采集系统上所有监听端口的详细信息。

## 第一次移植：基础结构 + procfs 方法

### 实现内容

1. **Port 结构体** (`port.go`)
   - 定义了端口信息的数据结构
   - 包含网络层信息：Family, Protocol, State, Sport, Dport, Sip, Dip, Uid, Inode, Username

2. **procfs.go**
   - `parseIP()`: 解析十六进制格式的 IP 地址
   - `procNet()`: 从 `/proc/net/` 文件读取端口信息

3. **ListeningPorts()** (`port.go`)
   - 主入口函数
   - 遍历 TCP/UDP 和 IPv4/IPv6 组合
   - 调用 `procNet()` 获取端口信息

### 工作原理

Linux 系统在以下文件中记录端口信息：
- `/proc/net/tcp` - TCP IPv4 端口
- `/proc/net/tcp6` - TCP IPv6 端口
- `/proc/net/udp` - UDP IPv4 端口
- `/proc/net/udp6` - UDP IPv6 端口

这些文件的格式：
```
sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
 0: 0100007F:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 12345
```

字段说明：
- `local_address`: 本地地址（IP:端口，十六进制）
- `rem_address`: 远程地址（IP:端口，十六进制）
- `st`: 状态（十六进制，TCP 10=LISTEN，UDP 7=UDP）
- `uid`: 用户 ID
- `inode`: Socket inode 号

### 使用方法

```go
import "gitlab.myinterest.top/security/agent/business_plugins/collector/port"

ports, err := port.ListeningPorts()
if err != nil {
    log.Fatal(err)
}

for _, p := range ports {
    fmt.Printf("Port: %s %s:%s\n", p.Protocol, p.Sip, p.Sport)
}
```

### 测试

#### 1. 运行单元测试
```bash
cd business_plugins/collector
go test ./port/... -v
```

#### 2. 运行独立测试程序
```bash
cd business_plugins/collector/port/cmd/test_port
go build -o test_port main.go
./test_port
```

测试程序会显示所有监听端口的详细信息，可以与系统命令对比：
```bash
ss -tuln | head -20
netstat -tuln | head -20
```

### 端口状态说明

- **TCP 状态 10 (0x0A)**: LISTEN - 监听状态
- **UDP 状态 7 (0x07)**: UDP - UDP 协议本身没有状态，7 表示 UDP socket

### 注意事项

1. 只采集监听状态的端口（TCP LISTEN 或 UDP）
2. IP 地址和端口号在文件中是十六进制格式，需要转换
3. IPv4 地址需要反转字节顺序（小端序）
4. 第一次移植不包含进程信息，只有网络层信息

### 后续计划

- **第二次移植**：添加进程关联功能（通过 inode 匹配）
- **第三次移植**：添加 netlink 方法（更高效）和完整集成

