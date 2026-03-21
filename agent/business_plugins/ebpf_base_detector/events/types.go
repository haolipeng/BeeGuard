// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"net"

	"shared/datatype"
)

// 事件类型常量（与 eBPF 内核侧一致）
const (
	EventTypeExecve      uint8 = 1
	EventTypeCommitCreds uint8 = 2
	EventTypeConnect     uint8 = 4
	EventTypeDNS         uint8 = 7
	EventTypeFile        uint8 = 8
	EventTypeMount       uint8 = 9
)

// 上报 DataType 常量（与平台约定一致，避免硬编码）
const (
	DataTypeExecve        int32 = datatype.EventExecve
	DataTypeConnect       int32 = datatype.EventConnect
	DataTypeDNS           int32 = datatype.EventDNS
	DataTypeFile          int32 = datatype.EventFile
	DataTypeMount         int32 = datatype.EventMount
	DataTypePerfEventLoss int32 = datatype.EventPerfLoss
)

// 文件操作 action 常量
const (
	FileActionCreate uint8 = 1
	FileActionRename uint8 = 2
	FileActionDelete uint8 = 3
)

// GetEventType 从原始数据中获取事件类型
func GetEventType(data []byte) uint8 {
	if len(data) < 1 {
		return 0
	}
	return data[0]
}

// NetworkIPToString 将网络字节序的 IPv4 地址转换为可读字符串（导出版本）
func NetworkIPToString(ip uint32) string {
	return networkIPToString(ip)
}

// NetworkPortToHost 将网络字节序端口转换为主机字节序（导出版本）
func NetworkPortToHost(port uint16) uint16 {
	return networkPortToHost(port)
}

func networkIPToString(ip uint32) string {
	return net.IP([]byte{
		byte(ip), byte(ip >> 8), byte(ip >> 16), byte(ip >> 24),
	}).String()
}

func networkPortToHost(port uint16) uint16 {
	return binary.BigEndian.Uint16([]byte{byte(port), byte(port >> 8)})
}

func cstring(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

func argsString(b []byte) string {
	end := len(b)
	for i := 0; i < len(b); i++ {
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
	result := make([]byte, end)
	copy(result, b[:end])
	for i := 0; i < len(result); i++ {
		if result[i] == 0 {
			result[i] = ' '
		}
	}
	return string(bytes.TrimRight(result, " "))
}
