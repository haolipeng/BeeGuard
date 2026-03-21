// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// DNSEvent DNS 查询事件 - 对应 C 结构体 struct dns_event
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

// ToRecord 转换为 Agent 的 protobuf Record 格式
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
		DataType:  DataTypeDNS,
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
