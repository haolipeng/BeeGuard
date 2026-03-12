// SPDX-License-Identifier: GPL-2.0
package main

import (
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/events"
)

// BuildContainerSensitiveFileRecord 从 FileEvent 和检测结果构建 DataType 7004 告警
func BuildContainerSensitiveFileRecord(evt *events.FileEvent, result *DetectionResult, pidTree string, containerMeta *ContainerMetaCache) *businessplugins.Record {
	comm := cstring(evt.Comm[:])
	exePath := cstring(evt.ExePath[:])
	newPath := cstring(evt.NewPath[:])
	oldPath := cstring(evt.OldPath[:])
	sID := cstring(evt.SID[:])

	actionStr := "unknown"
	switch evt.Action {
	case events.FileActionCreate:
		actionStr = "create"
	case events.FileActionRename:
		actionStr = "rename"
	case events.FileActionDelete:
		actionStr = "delete"
	}

	fields := map[string]string{
		"pid":             fmt.Sprintf("%d", evt.PID),
		"tgid":            fmt.Sprintf("%d", evt.TGID),
		"ppid":            fmt.Sprintf("%d", evt.PPID),
		"uid":             fmt.Sprintf("%d", evt.UID),
		"comm":            comm,
		"exe_path":        exePath,
		"action":          actionStr,
		"new_path":        newPath,
		"s_id":            sID,
		"detection_type":  DetectionTypeContainerSensitiveFile,
		"rule_id":         fmt.Sprintf("%d", result.RuleID),
		"rule_name":       result.RuleName,
		"severity":        result.Severity,
		"rule_description": result.Description,
		"matched_pattern": result.MatchedPattern,
		"pid_tree":        pidTree,
		"operator_user":   resolveUsername(evt.UID),
		"operator_process": comm,
		"timestamp":       fmt.Sprintf("%d", time.Now().Unix()),
	}

	if oldPath != "" {
		fields["old_path"] = oldPath
	}

	if evt.SocketPID != 0 {
		fields["socket_pid"] = fmt.Sprintf("%d", evt.SocketPID)
		fields["remote_ip"] = events.NetworkIPToString(evt.RemoteIP)
		fields["remote_port"] = fmt.Sprintf("%d", events.NetworkPortToHost(evt.RemotePort))
		fields["local_ip"] = events.NetworkIPToString(evt.LocalIP)
		fields["local_port"] = fmt.Sprintf("%d", evt.LocalPort)
	}

	// 填充容器元数据
	enrichContainerFields(fields, evt.TGID, containerMeta)

	return &businessplugins.Record{
		DataType:  businessplugins.AlertTypeContainerSensitiveFile,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}
