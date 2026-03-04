// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// ExecveEvent execve 事件 - 对应 C 结构体 struct execve_event
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

// ToRecord 转换为 Agent 的 protobuf Record 格式
func (e *ExecveEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
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

	fields["stdin_path"] = cstring(e.StdinPath[:])
	fields["stdout_path"] = cstring(e.StdoutPath[:])
	fields["tty_name"] = cstring(e.TTYName[:])
	fields["socket_pid"] = fmt.Sprintf("%d", e.SocketPID)
	fields["fd_type"] = fmt.Sprintf("%d", e.FDType)

	if e.RemoteIP != 0 {
		fields["remote_ip"] = networkIPToString(e.RemoteIP)
		fields["remote_port"] = fmt.Sprintf("%d", networkPortToHost(e.RemotePort))
		fields["local_ip"] = networkIPToString(e.LocalIP)
		fields["local_port"] = fmt.Sprintf("%d", e.LocalPort)
	}

	return &businessplugins.Record{
		DataType:  DataTypeExecve,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}
