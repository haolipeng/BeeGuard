// SPDX-License-Identifier: GPL-2.0
package main

import (
	"fmt"
	"time"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/events"
)

// ReverseShellDetector 用户态反弹 shell 检测器
// 基于 enriched execve 事件中的 stdin/stdout 路径、TTY、socket 信息进行规则判定
type ReverseShellDetector struct{}

// ReverseShellResult 检测结果
type ReverseShellResult struct {
	RuleName    string // "stdin_socket" / "stdout_socket" / "no_tty_with_socket"
	Confidence  string // "high" / "medium"
	Description string
}

// Detect 对 enriched ExecveEvent 执行反弹 shell 检测规则
// 返回 nil 表示未命中任何规则
func (d *ReverseShellDetector) Detect(evt *events.ExecveEvent) *ReverseShellResult {
	ttyName := cstring(evt.TTYName[:])

	// 基础规则: stdin 是 socket（高置信度）
	// fd_type 由内核 i_mode 检查直接推导，bit 0 = stdin 是 S_IFSOCK
	if evt.FDType&1 != 0 {
		return &ReverseShellResult{
			RuleName:    "stdin_socket",
			Confidence:  "high",
			Description: "stdin (fd 0) is connected to a socket",
		}
	}

	// 基础规则: stdout 是 socket（高置信度）
	// fd_type bit 1 = stdout 是 S_IFSOCK
	if evt.FDType&2 != 0 {
		return &ReverseShellResult{
			RuleName:    "stdout_socket",
			Confidence:  "high",
			Description: "stdout (fd 1) is connected to a socket",
		}
	}

	// 关联规则: 无 TTY + 有 socket 连接（中等置信度）
	if ttyName == "" && evt.SocketPID > 0 {
		return &ReverseShellResult{
			RuleName:    "no_tty_with_socket",
			Confidence:  "medium",
			Description: "process has no controlling terminal but parent chain has active socket",
		}
	}

	return nil
}

// BuildReverseShellRecord 从 enriched ExecveEvent 和检测结果构建 DataType 6004 告警
func BuildReverseShellRecord(evt *events.ExecveEvent, result *ReverseShellResult, pidTree string) *businessplugins.Record {
	comm := cstring(evt.Comm[:])
	exePath := cstring(evt.ExePath[:])
	args := argsString(evt.Args[:])
	stdinPath := cstring(evt.StdinPath[:])
	stdoutPath := cstring(evt.StdoutPath[:])
	ttyName := cstring(evt.TTYName[:])

	fields := map[string]string{
		"pid":         fmt.Sprintf("%d", evt.PID),
		"tgid":        fmt.Sprintf("%d", evt.TGID),
		"ppid":        fmt.Sprintf("%d", evt.PPID),
		"pgid":        fmt.Sprintf("%d", evt.PGID),
		"uid":         fmt.Sprintf("%d", evt.UID),
		"comm":        comm,
		"exe_path":    exePath,
		"args":        args,
		"fd_type":     fmt.Sprintf("%d", evt.FDType),
		"stdin_path":  stdinPath,
		"stdout_path": stdoutPath,
		"pid_tree":    pidTree,
		"tty_name":    ttyName,
		"socket_pid":  fmt.Sprintf("%d", evt.SocketPID),
		"rule_name":   result.RuleName,
		"confidence":  result.Confidence,
		"description": result.Description,
	}

	if evt.RemoteIP != 0 {
		fields["remote_ip"] = events.NetworkIPToString(evt.RemoteIP)
		fields["remote_port"] = fmt.Sprintf("%d", events.NetworkPortToHost(evt.RemotePort))
		fields["local_ip"] = events.NetworkIPToString(evt.LocalIP)
		fields["local_port"] = fmt.Sprintf("%d", evt.LocalPort)
	}

	return &businessplugins.Record{
		DataType:  businessplugins.AlertTypeReverseShell,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: fields,
		},
	}
}
