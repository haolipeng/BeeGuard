// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// MountEvent mount 事件 - 对应 C 结构体 struct mount_event
type MountEvent struct {
	EventType   uint8      // EVENT_TYPE_MOUNT = 9
	Padding1    [3]byte    // 对齐填充
	PID         uint32     // 进程ID（线程ID）
	TGID        uint32     // 线程组ID（进程ID）
	PPID        uint32     // 父进程ID
	UID         uint32     // 用户ID
	MntnsID     uint64     // mount 命名空间 ID
	RootMntnsID uint64     // 宿主机 mount 命名空间 ID
	Comm        [16]byte   // 进程名
	ExePath     [256]byte  // 可执行文件路径
	DevName     [256]byte  // 挂载源设备
	DirName     [256]byte  // 挂载目标路径
	FsType      [32]byte   // 文件系统类型
	Flags       uint32     // mount 标志
	RetVal      int32      // 系统调用返回值
}

// UnmarshalBinary 从二进制数据反序列化事件
func (e *MountEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

// ToRecord 转换为 Agent 的 protobuf Record 格式
func (e *MountEvent) ToRecord() *businessplugins.Record {
	fields := map[string]string{
		"pid":           fmt.Sprintf("%d", e.PID),
		"tgid":          fmt.Sprintf("%d", e.TGID),
		"ppid":          fmt.Sprintf("%d", e.PPID),
		"uid":           fmt.Sprintf("%d", e.UID),
		"comm":          cstring(e.Comm[:]),
		"exe_path":      cstring(e.ExePath[:]),
		"dev_name":      cstring(e.DevName[:]),
		"dir_name":      cstring(e.DirName[:]),
		"fs_type":       cstring(e.FsType[:]),
		"flags":         fmt.Sprintf("%d", e.Flags),
		"retval":        fmt.Sprintf("%d", e.RetVal),
		"mntns_id":      fmt.Sprintf("%d", e.MntnsID),
		"root_mntns_id": fmt.Sprintf("%d", e.RootMntnsID),
		"is_container":  fmt.Sprintf("%t", e.MntnsID != e.RootMntnsID && e.RootMntnsID != 0),
	}

	return &businessplugins.Record{
		DataType:  DataTypeMount,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}
