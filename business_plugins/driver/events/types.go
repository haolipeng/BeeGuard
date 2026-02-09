// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// 事件类型常量
const (
	EventTypeExecve      uint8 = 1
	EventTypeCommitCreds uint8 = 2
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
	EventType uint8     // 事件类型标识 (EVENT_TYPE_EXECVE = 1)
	Padding1  [3]byte   // 对齐填充
	PID       uint32    // 进程ID（线程ID）
	TGID      uint32    // 线程组ID（进程ID）
	PPID      uint32    // 父进程ID
	PGID      uint32    // 进程组ID
	UID       uint32    // 用户ID
	Padding   uint32    // 对齐填充
	Comm      [16]byte  // 进程名
	ExePath   [256]byte // 可执行文件的完整路径
	Args      [512]byte // 命令行参数
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

	return &businessplugins.Record{
		DataType:  59, // execve事件类型
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":      fmt.Sprintf("%d", e.PID),
				"tgid":     fmt.Sprintf("%d", e.TGID),
				"ppid":     fmt.Sprintf("%d", e.PPID),
				"pgid":     fmt.Sprintf("%d", e.PGID),
				"uid":      fmt.Sprintf("%d", e.UID),
				"comm":     comm,
				"exe_path": exePath,
				"args":     args,
			},
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
