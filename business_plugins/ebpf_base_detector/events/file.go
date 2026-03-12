// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// FileEvent 文件操作事件 - 对应 C 结构体 struct file_event
type FileEvent struct {
	EventType  uint8      // EVENT_TYPE_FILE = 8
	Action     uint8      // 1=create, 2=rename, 3=delete
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
	MntnsID    uint64     // 当前进程 mount 命名空间 ID
	RootMntnsID uint64   // 宿主机 mount 命名空间 ID
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *FileEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为 Agent 的 protobuf Record 格式
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
	case FileActionDelete:
		actionStr = "delete"
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
		DataType:  DataTypeFile,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}
