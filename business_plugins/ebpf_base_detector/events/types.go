// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	businessplugins "business_plugins/lib"
)

// 事件类型常量
const (
	EventTypeExecve      uint8 = 1
	EventTypeCommitCreds uint8 = 2
	EventTypeConnect     uint8 = 4
	EventTypeBind         uint8 = 5
	EventTypeAccept       uint8 = 6
	EventTypeDNS          uint8 = 7
	EventTypeFile         uint8 = 8
)

// 文件操作 action 常量
const (
	FileActionCreate uint8 = 1
	FileActionRename uint8 = 2
)

// GetEventType 从原始数据中获取事件类型
// 事件类型存储在数据的第一个字节
func GetEventType(data []byte) uint8 {
	if len(data) < 1 {
		return 0
	}
	return data[0]
}

// ExecveEvent execve事件 - 对应C结构体 struct execve_event
type ExecveEvent struct {
	EventType  uint8      // 事件类型标识 (EVENT_TYPE_EXECVE = 1)
	FDType     uint8      // 内核预过滤: 0=无, 1=stdin是socket, 2=stdout, 3=both
	Padding1   [2]byte    // 对齐填充
	PID        uint32     // 进程ID（线程ID）
	TGID       uint32     // 线程组ID（进程ID）
	PPID       uint32     // 父进程ID
	PGID       uint32     // 进程组ID
	UID        uint32     // 用户ID
	SocketPID  uint32     // 持有 socket 的进程 PID
	Comm       [16]byte   // 进程名
	ExePath    [256]byte  // 可执行文件的完整路径
	Args       [512]byte  // 命令行参数
	StdinPath  [64]byte   // FD 0 的文件路径
	StdoutPath [64]byte   // FD 1 的文件路径
	TTYName    [64]byte   // 控制终端名称
	RemoteIP   uint32     // socket 远程 IP（网络字节序）
	RemotePort uint16     // socket 远程端口（网络字节序）
	LocalPort  uint16     // socket 本地端口（主机字节序）
	LocalIP    uint32     // socket 本地 IP（网络字节序）
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *ExecveEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *ExecveEvent) ToRecord() *businessplugins.Record {
	// 将字节数组转换为字符串（C字符串以\0结尾）
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
	// 命令行参数需要特殊处理：将NULL字节替换为空格
	args := argsString(e.Args[:])

	fields := map[string]string{
		"pid":      fmt.Sprintf("%d", e.PID),
		"tgid":     fmt.Sprintf("%d", e.TGID),
		"ppid":     fmt.Sprintf("%d", e.PPID),
		"pgid":     fmt.Sprintf("%d", e.PGID),
		"uid":      fmt.Sprintf("%d", e.UID),
		"comm":     comm,
		"exe_path": exePath,
		"args":     args,
	}

	// 新增反弹 shell 增强字段
	stdinPath := cstring(e.StdinPath[:])
	stdoutPath := cstring(e.StdoutPath[:])
	ttyName := cstring(e.TTYName[:])

	fields["stdin_path"] = stdinPath
	fields["stdout_path"] = stdoutPath
	fields["tty_name"] = ttyName
	fields["socket_pid"] = fmt.Sprintf("%d", e.SocketPID)
	fields["fd_type"] = fmt.Sprintf("%d", e.FDType)

	if e.RemoteIP != 0 {
		fields["remote_ip"] = networkIPToString(e.RemoteIP)
		fields["remote_port"] = fmt.Sprintf("%d", networkPortToHost(e.RemotePort))
		fields["local_ip"] = networkIPToString(e.LocalIP)
		fields["local_port"] = fmt.Sprintf("%d", e.LocalPort)
	}

	return &businessplugins.Record{
		DataType:  59, // execve事件类型
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}

// CommitCredsEvent commit_creds提权事件 - 对应C结构体 struct commit_creds_event
type CommitCredsEvent struct {
	EventType uint8     // 事件类型标识 (EVENT_TYPE_COMMIT_CREDS = 2)
	Padding1  [3]byte   // 对齐填充
	PID       uint32    // 进程ID
	TGID      uint32    // 线程组ID
	PPID      uint32    // 父进程ID
	UID       uint32    // 当前用户ID
	OldUID    uint32    // 提权前的uid
	OldEUID   uint32    // 提权前的euid
	NewUID    uint32    // 提权后的uid
	NewEUID   uint32    // 提权后的euid
	Comm      [16]byte  // 进程名
	ExePath   [256]byte // 可执行文件路径
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *CommitCredsEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *CommitCredsEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])

	return &businessplugins.Record{
		DataType:  6006, // 本地提权告警类型
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":      fmt.Sprintf("%d", e.PID),
				"tgid":     fmt.Sprintf("%d", e.TGID),
				"ppid":     fmt.Sprintf("%d", e.PPID),
				"uid":      fmt.Sprintf("%d", e.UID),
				"old_uid":  fmt.Sprintf("%d", e.OldUID),
				"old_euid": fmt.Sprintf("%d", e.OldEUID),
				"new_uid":  fmt.Sprintf("%d", e.NewUID),
				"new_euid": fmt.Sprintf("%d", e.NewEUID),
				"comm":     comm,
				"exe_path": exePath,
			},
		},
	}
}

// ConnectEvent connect出站连接事件 - 对应C结构体 struct connect_event
type ConnectEvent struct {
	EventType  uint8      // EVENT_TYPE_CONNECT = 4
	Protocol   uint8      // 6=TCP, 17=UDP
	Padding1   [2]byte
	PID        uint32
	TGID       uint32
	PPID       uint32
	UID        uint32
	RemoteIP   uint32     // 目标 IP（网络字节序）
	RemotePort uint16     // 目标端口（网络字节序）
	LocalPort  uint16     // 本地端口
	LocalIP    uint32     // 本地 IP（网络字节序）
	RetVal     int32      // 系统调用返回值
	Comm       [16]byte
	ExePath    [256]byte
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *ConnectEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *ConnectEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])

	remoteIP := networkIPToString(e.RemoteIP)
	localIP := networkIPToString(e.LocalIP)
	remotePort := networkPortToHost(e.RemotePort)

	protoStr := "unknown"
	switch e.Protocol {
	case 6:
		protoStr = "tcp"
	case 17:
		protoStr = "udp"
	}

	return &businessplugins.Record{
		DataType:  60,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":         fmt.Sprintf("%d", e.PID),
				"tgid":        fmt.Sprintf("%d", e.TGID),
				"ppid":        fmt.Sprintf("%d", e.PPID),
				"uid":         fmt.Sprintf("%d", e.UID),
				"comm":        comm,
				"exe_path":    exePath,
				"protocol":    protoStr,
				"remote_ip":   remoteIP,
				"remote_port": fmt.Sprintf("%d", remotePort),
				"local_ip":    localIP,
				"local_port":  fmt.Sprintf("%d", e.LocalPort),
				"retval":      fmt.Sprintf("%d", e.RetVal),
			},
		},
	}
}

// BindEvent bind端口绑定事件 - 对应C结构体 struct bind_event
type BindEvent struct {
	EventType uint8      // EVENT_TYPE_BIND = 5
	Protocol  uint8
	Padding1  [2]byte
	PID       uint32
	TGID      uint32
	PPID      uint32
	UID       uint32
	BindIP    uint32     // 绑定 IP
	BindPort  uint16     // 绑定端口（网络字节序）
	Padding2  uint16
	RetVal    int32
	Comm      [16]byte
	ExePath   [256]byte
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *BindEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *BindEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])

	bindIP := networkIPToString(e.BindIP)
	bindPort := networkPortToHost(e.BindPort)

	protoStr := "unknown"
	switch e.Protocol {
	case 6:
		protoStr = "tcp"
	case 17:
		protoStr = "udp"
	}

	return &businessplugins.Record{
		DataType:  61,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":       fmt.Sprintf("%d", e.PID),
				"tgid":      fmt.Sprintf("%d", e.TGID),
				"ppid":      fmt.Sprintf("%d", e.PPID),
				"uid":       fmt.Sprintf("%d", e.UID),
				"comm":      comm,
				"exe_path":  exePath,
				"protocol":  protoStr,
				"bind_ip":   bindIP,
				"bind_port": fmt.Sprintf("%d", bindPort),
				"retval":    fmt.Sprintf("%d", e.RetVal),
			},
		},
	}
}

// AcceptEvent accept入站连接事件 - 对应C结构体 struct accept_event
type AcceptEvent struct {
	EventType  uint8      // EVENT_TYPE_ACCEPT = 6
	Protocol   uint8
	Padding1   [2]byte
	PID        uint32
	TGID       uint32
	PPID       uint32
	UID        uint32
	RemoteIP   uint32     // 连接来源 IP
	RemotePort uint16     // 连接来源端口（网络字节序）
	LocalPort  uint16     // 本地监听端口
	LocalIP    uint32
	RetVal     int32
	Comm       [16]byte
	ExePath    [256]byte
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *AcceptEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *AcceptEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])

	remoteIP := networkIPToString(e.RemoteIP)
	localIP := networkIPToString(e.LocalIP)
	remotePort := networkPortToHost(e.RemotePort)

	protoStr := "unknown"
	switch e.Protocol {
	case 6:
		protoStr = "tcp"
	case 17:
		protoStr = "udp"
	}

	return &businessplugins.Record{
		DataType:  62,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":         fmt.Sprintf("%d", e.PID),
				"tgid":        fmt.Sprintf("%d", e.TGID),
				"ppid":        fmt.Sprintf("%d", e.PPID),
				"uid":         fmt.Sprintf("%d", e.UID),
				"comm":        comm,
				"exe_path":    exePath,
				"protocol":    protoStr,
				"remote_ip":   remoteIP,
				"remote_port": fmt.Sprintf("%d", remotePort),
				"local_ip":    localIP,
				"local_port":  fmt.Sprintf("%d", e.LocalPort),
				"retval":      fmt.Sprintf("%d", e.RetVal),
			},
		},
	}
}

// DNSEvent DNS查询事件 - 对应C结构体 struct dns_event
type DNSEvent struct {
	EventType     uint8      // EVENT_TYPE_DNS = 7
	Padding1      [3]byte
	PID           uint32
	TGID          uint32
	PPID          uint32
	UID           uint32
	DNSServerIP   uint32     // DNS 服务器 IP
	DNSServerPort uint16     // DNS 服务器端口（网络字节序）
	QueryType     uint16     // DNS 查询类型
	Opcode        int32
	Rcode         int32
	Comm          [16]byte
	ExePath       [256]byte
	Domain        [256]byte  // 查询域名
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *DNSEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *DNSEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
	domain := cstring(e.Domain[:])

	serverIP := networkIPToString(e.DNSServerIP)
	serverPort := networkPortToHost(e.DNSServerPort)

	qtypeStr := fmt.Sprintf("%d", e.QueryType)
	switch e.QueryType {
	case 1:
		qtypeStr = "A"
	case 5:
		qtypeStr = "CNAME"
	case 15:
		qtypeStr = "MX"
	case 16:
		qtypeStr = "TXT"
	case 28:
		qtypeStr = "AAAA"
	}

	return &businessplugins.Record{
		DataType:  63,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":             fmt.Sprintf("%d", e.PID),
				"tgid":            fmt.Sprintf("%d", e.TGID),
				"ppid":            fmt.Sprintf("%d", e.PPID),
				"uid":             fmt.Sprintf("%d", e.UID),
				"comm":            comm,
				"exe_path":        exePath,
				"domain":          domain,
				"query_type":      qtypeStr,
				"dns_server_ip":   serverIP,
				"dns_server_port": fmt.Sprintf("%d", serverPort),
				"opcode":          fmt.Sprintf("%d", e.Opcode),
				"rcode":           fmt.Sprintf("%d", e.Rcode),
			},
		},
	}
}

// FileEvent 文件操作事件 - 对应C结构体 struct file_event
type FileEvent struct {
	EventType  uint8      // EVENT_TYPE_FILE = 8
	Action     uint8      // 1=create, 2=rename
	Padding1   [2]byte
	PID        uint32
	TGID       uint32
	PPID       uint32
	UID        uint32
	SocketPID  uint32     // 持有 socket 的进程 PID
	RemoteIP   uint32     // socket 远程 IP（网络字节序）
	RemotePort uint16     // socket 远程端口（网络字节序）
	LocalPort  uint16     // socket 本地端口
	LocalIP    uint32     // socket 本地 IP（网络字节序）
	Comm       [16]byte
	ExePath    [256]byte  // 操作进程的可执行文件路径
	NewPath    [512]byte  // 创建：文件路径；重命名：新路径
	OldPath    [512]byte  // 仅重命名有值
	SID        [32]byte   // 文件系统 ID
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *FileEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为Agent的protobuf Record格式
func (e *FileEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
	newPath := cstring(e.NewPath[:])
	oldPath := cstring(e.OldPath[:])
	sID := cstring(e.SID[:])

	actionStr := "unknown"
	switch e.Action {
	case FileActionCreate:
		actionStr = "create"
	case FileActionRename:
		actionStr = "rename"
	}

	fields := map[string]string{
		"pid":      fmt.Sprintf("%d", e.PID),
		"tgid":     fmt.Sprintf("%d", e.TGID),
		"ppid":     fmt.Sprintf("%d", e.PPID),
		"uid":      fmt.Sprintf("%d", e.UID),
		"comm":     comm,
		"exe_path": exePath,
		"action":   actionStr,
		"new_path": newPath,
		"s_id":     sID,
	}

	if oldPath != "" {
		fields["old_path"] = oldPath
	}

	if e.SocketPID != 0 {
		fields["socket_pid"] = fmt.Sprintf("%d", e.SocketPID)
		fields["remote_ip"] = networkIPToString(e.RemoteIP)
		fields["remote_port"] = fmt.Sprintf("%d", networkPortToHost(e.RemotePort))
		fields["local_ip"] = networkIPToString(e.LocalIP)
		fields["local_port"] = fmt.Sprintf("%d", e.LocalPort)
	}

	return &businessplugins.Record{
		DataType:  64, // 文件操作基础事件类型
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}

// NetworkIPToString 将网络字节序的IPv4地址转换为可读字符串（导出版本）
func NetworkIPToString(ip uint32) string {
	return networkIPToString(ip)
}

// NetworkPortToHost 将网络字节序端口转换为主机字节序（导出版本）
func NetworkPortToHost(port uint16) uint16 {
	return networkPortToHost(port)
}

// networkIPToString 将网络字节序的IPv4地址转换为可读字符串
func networkIPToString(ip uint32) string {
	return net.IP([]byte{
		byte(ip),
		byte(ip >> 8),
		byte(ip >> 16),
		byte(ip >> 24),
	}).String()
}

// networkPortToHost 将网络字节序端口转换为主机字节序
func networkPortToHost(port uint16) uint16 {
	return binary.BigEndian.Uint16([]byte{byte(port), byte(port >> 8)})
}

// cstring 将C字符串（以\0结尾）转换为Go字符串
func cstring(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

// argsString 处理命令行参数：将NULL字节分隔的多个参数转换为空格分隔的字符串
func argsString(b []byte) string {
	// 找到实际数据的结尾（连续的NULL字节）
	end := len(b)
	for i := 0; i < len(b); i++ {
		// 如果遇到连续的NULL，说明数据结束
		if b[i] == 0 {
			allZero := true
			for j := i; j < len(b) && j < i+4; j++ {
				if b[j] != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				end = i
				break
			}
		}
	}

	// 将NULL字节替换为空格
	result := make([]byte, end)
	copy(result, b[:end])
	for i := 0; i < len(result); i++ {
		if result[i] == 0 {
			result[i] = ' '
		}
	}

	// 去除尾部空格
	return string(bytes.TrimRight(result, " "))
}
