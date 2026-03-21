// SPDX-License-Identifier: GPL-2.0
package events

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
)

// CommitCredsEvent commit_creds 提权事件 - 对应 C 结构体 struct commit_creds_event
type CommitCredsEvent struct {
	EventType uint8
	Padding1  [3]byte
	PID       uint32
	TGID      uint32
	PPID      uint32
	UID       uint32
	OldUID    uint32
	OldEUID   uint32
	NewUID    uint32
	NewEUID   uint32
	Comm      [16]byte
	ExePath   [256]byte
}

func (e *CommitCredsEvent) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	return binary.Read(buf, binary.LittleEndian, e)
}

func (e *CommitCredsEvent) ToRecord() *businessplugins.Record {
	comm := cstring(e.Comm[:])
	exePath := cstring(e.ExePath[:])
	return &businessplugins.Record{
		DataType:  businessplugins.AlertTypePrivilegeEscalation,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid": fmt.Sprintf("%d", e.PID), "tgid": fmt.Sprintf("%d", e.TGID),
				"ppid": fmt.Sprintf("%d", e.PPID), "uid": fmt.Sprintf("%d", e.UID),
				"old_uid": fmt.Sprintf("%d", e.OldUID), "old_euid": fmt.Sprintf("%d", e.OldEUID),
				"new_uid": fmt.Sprintf("%d", e.NewUID), "new_euid": fmt.Sprintf("%d", e.NewEUID),
				"comm": comm, "exe_path": exePath,
			},
		},
	}
}
