// SPDX-License-Identifier: GPL-2.0
package main

import (
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/events"
)

// ContainerReverseShellDetector 容器侧反弹 Shell 检测器
// 仅在 IsContainer() == true 时调用，规则独立于主机侧 ReverseShellDetector
type ContainerReverseShellDetector struct{}

// ContainerReverseShellResult 检测结果
type ContainerReverseShellResult struct {
	RuleName    string // "container_stdin_socket" / "container_stdout_socket"
	Confidence  string // "high"
	Description string
}

// Detect 对容器内的 ExecveEvent 执行反弹 Shell 检测规则
// 返回 nil 表示未命中任何规则
func (d *ContainerReverseShellDetector) Detect(evt *events.ExecveEvent) *ContainerReverseShellResult {
	// 规则 1: stdin (FD 0) 指向网络 socket
	// fd_type bit 0 = stdin 是 S_IFSOCK，由 eBPF 内核层 i_mode 检查设置
	if evt.FDType&1 != 0 {
		return &ContainerReverseShellResult{
			RuleName:    "container_stdin_socket",
			Confidence:  "high",
			Description: "container process stdin (fd 0) is connected to a network socket",
		}
	}

	// 规则 2: stdout (FD 1) 指向网络 socket
	// fd_type bit 1 = stdout 是 S_IFSOCK
	if evt.FDType&2 != 0 {
		return &ContainerReverseShellResult{
			RuleName:    "container_stdout_socket",
			Confidence:  "high",
			Description: "container process stdout (fd 1) is connected to a network socket",
		}
	}

	return nil
}

// BuildContainerReverseShellRecord 从 ExecveEvent 和检测结果构建 DataType 7003 告警
func BuildContainerReverseShellRecord(evt *events.ExecveEvent, result *ContainerReverseShellResult, pidTree string, containerMeta *ContainerMetaCache) *businessplugins.Record {
	comm := cstring(evt.Comm[:])
	exePath := cstring(evt.ExePath[:])
	args := argsString(evt.Args[:])
	stdinPath := cstring(evt.StdinPath[:])
	stdoutPath := cstring(evt.StdoutPath[:])
	ttyName := cstring(evt.TTYName[:])

	fields := map[string]string{
		"pid":            fmt.Sprintf("%d", evt.PID),
		"tgid":           fmt.Sprintf("%d", evt.TGID),
		"ppid":           fmt.Sprintf("%d", evt.PPID),
		"pgid":           fmt.Sprintf("%d", evt.PGID),
		"uid":            fmt.Sprintf("%d", evt.UID),
		"comm":           comm,
		"exe_path":       exePath,
		"args":           args,
		"fd_type":        fmt.Sprintf("%d", evt.FDType),
		"stdin_path":     stdinPath,
		"stdout_path":    stdoutPath,
		"pid_tree":       pidTree,
		"tty_name":       ttyName,
		"socket_pid":     fmt.Sprintf("%d", evt.SocketPID),
		"detection_type": DetectionTypeContainerReverseShell,
		"rule_name":      result.RuleName,
		"confidence":     result.Confidence,
		"description":    result.Description,
	}

	if evt.RemoteIP != 0 {
		fields["remote_ip"] = events.NetworkIPToString(evt.RemoteIP)
		fields["remote_port"] = fmt.Sprintf("%d", events.NetworkPortToHost(evt.RemotePort))
		fields["local_ip"] = events.NetworkIPToString(evt.LocalIP)
		fields["local_port"] = fmt.Sprintf("%d", evt.LocalPort)
	}

	// 填充容器元数据
	enrichContainerFields(fields, evt.TGID, containerMeta)

	return &businessplugins.Record{
		DataType:  businessplugins.AlertTypeContainerReverseShell,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}
