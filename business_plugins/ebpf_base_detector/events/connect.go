// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// ConnectEvent connect 出站连接事件 - 对应 C 结构体 struct connect_event
type ConnectEvent struct {
	EventType  uint8
	Protocol   uint8
	Padding1   [2]byte
	PID        uint32
	TGID       uint32
	PPID       uint32
	UID        uint32
	RemoteIP   uint32
	RemotePort uint16
	LocalPort  uint16
	LocalIP    uint32
	RetVal     int32
	Comm       [16]byte
	ExePath    [256]byte
}

func (e *ConnectEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

func (e *ConnectEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
	protoStr := "unknown"
	if e.Protocol == 6 {
		protoStr = "tcp"
	} else if e.Protocol == 17 {
		protoStr = "udp"
	}
	return &businessplugins.Record{
		DataType:  DataTypeConnect,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid": fmt.Sprintf("%d", e.PID), "tgid": fmt.Sprintf("%d", e.TGID),
				"ppid": fmt.Sprintf("%d", e.PPID), "uid": fmt.Sprintf("%d", e.UID),
				"comm": comm, "exe_path": exePath, "protocol": protoStr,
				"remote_ip": networkIPToString(e.RemoteIP),
				"remote_port": fmt.Sprintf("%d", networkPortToHost(e.RemotePort)),
				"local_ip": networkIPToString(e.LocalIP),
				"local_port": fmt.Sprintf("%d", e.LocalPort),
				"retval": fmt.Sprintf("%d", e.RetVal),
			},
		},
	}
}
