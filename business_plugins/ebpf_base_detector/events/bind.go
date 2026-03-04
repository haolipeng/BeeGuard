// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// BindEvent bind 端口绑定事件 - 对应 C 结构体 struct bind_event
type BindEvent struct {
	EventType uint8
	Protocol  uint8
	Padding1  [2]byte
	PID       uint32
	TGID      uint32
	PPID      uint32
	UID       uint32
	BindIP    uint32
	BindPort  uint16
	Padding2  uint16
	RetVal    int32
	Comm      [16]byte
	ExePath   [256]byte
}

func (e *BindEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

func (e *BindEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
	protoStr := "unknown"
	if e.Protocol == 6 {
		protoStr = "tcp"
	} else if e.Protocol == 17 {
		protoStr = "udp"
	}
	return &businessplugins.Record{
		DataType:  DataTypeBind,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid": fmt.Sprintf("%d", e.PID), "tgid": fmt.Sprintf("%d", e.TGID),
				"ppid": fmt.Sprintf("%d", e.PPID), "uid": fmt.Sprintf("%d", e.UID),
				"comm": comm, "exe_path": exePath, "protocol": protoStr,
				"bind_ip": networkIPToString(e.BindIP),
				"bind_port": fmt.Sprintf("%d", networkPortToHost(e.BindPort)),
				"retval": fmt.Sprintf("%d", e.RetVal),
			},
		},
	}
}
