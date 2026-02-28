package main

import (
	"fmt"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/events"
	"ebpf_base_detector/log"
)

// eventHandlerCtx 事件处理依赖，供各 handleXxx 使用
type eventHandlerCtx struct {
	client     *businessplugins.Client
	logger     *log.Logger
	dcDetector *DangerousCommandDetector
	rsDetector *ReverseShellDetector
	mrDetector *MaliciousRequestDetector
	sfDetector *SensitiveFileDetector
}

func handleExecve(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.ExecveEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal execve event: %w", err)
	}
	pidTreeStr := buildPidTree(evt.TGID, cstring(evt.Comm[:]))
	record := evt.ToRecord()
	record.Data.Fields["pid_tree"] = pidTreeStr

	if ctx.dcDetector != nil {
		comm := cstring(evt.Comm[:])
		args := argsString(evt.Args[:])
		result := ctx.dcDetector.Detect(comm, args)
		if result != nil {
			record.DataType = businessplugins.AlertTypeDangerousCommand
			record.Data.Fields["detection_type"] = DetectionTypeDangerousCommand
			record.Data.Fields["rule_id"] = fmt.Sprintf("%d", result.RuleID)
			record.Data.Fields["rule_name"] = result.RuleName
			record.Data.Fields["severity"] = result.Severity
			record.Data.Fields["rule_description"] = result.Description
			record.Data.Fields["matched_pattern"] = result.MatchedPattern
			record.Data.Fields["command"] = args
			record.Data.Fields["command_type"] = fmt.Sprintf("%d", result.RuleID)
			record.Data.Fields["user"] = record.Data.Fields["uid"]
			if evt.UID == 0 {
				record.Data.Fields["privilege_level"] = "root"
			} else {
				record.Data.Fields["privilege_level"] = "normal"
			}
			record.Data.Fields["timestamp"] = fmt.Sprintf("%d", record.Timestamp)
			ctx.logger.Info("Dangerous command detected",
				"rule_id", result.RuleID, "rule_name", result.RuleName, "severity", result.Severity,
				"uid", evt.UID, "comm", comm, "args", args)
		}
	}

	rsResult := ctx.rsDetector.Detect(&evt)
	if rsResult != nil {
		rsRecord := BuildReverseShellRecord(&evt, rsResult, pidTreeStr)
		ctx.logger.Warn("Reverse shell detected (userspace)",
			"rule", rsResult.RuleName, "confidence", rsResult.Confidence,
			"pid", evt.PID, "tgid", evt.TGID, "comm", cstring(evt.Comm[:]),
			"exe_path", cstring(evt.ExePath[:]), "stdin_path", cstring(evt.StdinPath[:]),
			"stdout_path", cstring(evt.StdoutPath[:]), "pid_tree", pidTreeStr,
			"tty_name", cstring(evt.TTYName[:]), "socket_pid", evt.SocketPID)
		if err := ctx.client.SendRecord(rsRecord); err != nil {
			ctx.logger.Error("Failed to send reverse shell record to agent", "error", err)
		}
	}
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send execve record: %w", err)
	}
	return nil
}

func handleCommitCreds(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.CommitCredsEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal commit_creds event: %w", err)
	}
	exePath := resolveExePath(evt.TGID, cstring(evt.ExePath[:]))
	record := evt.ToRecord()
	record.Data.Fields["exe_path"] = exePath
	record.Data.Fields["escalated_user"] = resolveUsername(evt.NewUID)
	record.Data.Fields["parent_process"] = resolveParentComm(evt.PPID)
	record.Data.Fields["parent_process_user"] = resolveParentUID(evt.PPID)
	record.Data.Fields["timestamp"] = fmt.Sprintf("%d", record.Timestamp)
	ctx.logger.Warn("Privilege escalation detected",
		"pid", evt.PID, "tgid", evt.TGID, "ppid", evt.PPID, "comm", cstring(evt.Comm[:]),
		"exe_path", exePath,
		"escalated_user", record.Data.Fields["escalated_user"],
		"parent_process", record.Data.Fields["parent_process"],
		"parent_process_user", record.Data.Fields["parent_process_user"],
		"old_uid", evt.OldUID, "old_euid", evt.OldEUID, "new_uid", evt.NewUID, "new_euid", evt.NewEUID)
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send privilege escalation record: %w", err)
	}
	return nil
}

func handleConnect(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.ConnectEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal connect event: %w", err)
	}
	record := evt.ToRecord()
	ctx.logger.Info("Connect event",
		"pid", evt.PID, "comm", cstring(evt.Comm[:]),
		"remote_ip", record.Data.Fields["remote_ip"], "remote_port", record.Data.Fields["remote_port"],
		"protocol", record.Data.Fields["protocol"], "retval", evt.RetVal)
	if ctx.mrDetector != nil {
		if mrResult := ctx.mrDetector.MatchConnect(&evt); mrResult != nil {
			mrRecord := BuildMaliciousRequestConnectRecord(&evt, mrResult)
			ctx.logger.Warn("Malicious request detected on connect",
				"rule_id", mrResult.RuleID, "rule_name", mrResult.RuleName,
				"threat_type", mrResult.ThreatType, "matched_value", mrResult.MatchedValue,
				"pid", evt.PID, "comm", cstring(evt.Comm[:]))
			if err := ctx.client.SendRecord(mrRecord); err != nil {
				ctx.logger.Error("Failed to send malicious request connect record to agent", "error", err)
			}
		}
	}
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send connect record: %w", err)
	}
	return nil
}

func handleBind(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.BindEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal bind event: %w", err)
	}
	record := evt.ToRecord()
	ctx.logger.Info("Bind event",
		"pid", evt.PID, "comm", cstring(evt.Comm[:]),
		"bind_ip", record.Data.Fields["bind_ip"], "bind_port", record.Data.Fields["bind_port"],
		"protocol", record.Data.Fields["protocol"])
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send bind record: %w", err)
	}
	return nil
}

func handleAccept(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.AcceptEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal accept event: %w", err)
	}
	record := evt.ToRecord()
	ctx.logger.Info("Accept event",
		"pid", evt.PID, "comm", cstring(evt.Comm[:]),
		"remote_ip", record.Data.Fields["remote_ip"], "remote_port", record.Data.Fields["remote_port"],
		"local_port", record.Data.Fields["local_port"], "protocol", record.Data.Fields["protocol"])
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send accept record: %w", err)
	}
	return nil
}

func handleDNS(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.DNSEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal DNS event: %w", err)
	}
	record := evt.ToRecord()
	ctx.logger.Info("DNS query event",
		"pid", evt.PID, "comm", cstring(evt.Comm[:]),
		"domain", record.Data.Fields["domain"], "query_type", record.Data.Fields["query_type"],
		"dns_server", record.Data.Fields["dns_server_ip"])
	if ctx.mrDetector != nil {
		if mrResult := ctx.mrDetector.MatchDNS(&evt); mrResult != nil {
			mrRecord := BuildMaliciousRequestDNSRecord(&evt, mrResult)
			ctx.logger.Warn("Malicious request detected on DNS",
				"rule_id", mrResult.RuleID, "rule_name", mrResult.RuleName,
				"threat_type", mrResult.ThreatType, "matched_value", mrResult.MatchedValue,
				"pid", evt.PID, "comm", cstring(evt.Comm[:]))
			if err := ctx.client.SendRecord(mrRecord); err != nil {
				ctx.logger.Error("Failed to send malicious request DNS record to agent", "error", err)
			}
		}
	}
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send DNS record: %w", err)
	}
	return nil
}

func handleFile(ctx *eventHandlerCtx, raw []byte) error {
	var evt events.FileEvent
	if err := evt.UnmarshalBinary(raw); err != nil {
		return fmt.Errorf("unmarshal file event: %w", err)
	}
	record := evt.ToRecord()
	pidTreeStr := buildPidTree(evt.TGID, cstring(evt.Comm[:]))
	record.Data.Fields["pid_tree"] = pidTreeStr
	newPath := cstring(evt.NewPath[:])
	actionStr := "create"
	switch evt.Action {
	case events.FileActionRename:
		actionStr = "rename"
	case events.FileActionDelete:
		actionStr = "delete"
	}
	ctx.logger.Info("File event",
		"pid", evt.PID, "comm", cstring(evt.Comm[:]), "action", actionStr,
		"new_path", newPath, "old_path", cstring(evt.OldPath[:]), "s_id", cstring(evt.SID[:]))
	if ctx.sfDetector != nil {
		result := ctx.sfDetector.Detect(newPath)
		if result != nil {
			alertRecord := evt.ToRecord()
			alertRecord.DataType = businessplugins.AlertTypeSensitiveFile
			alertRecord.Data.Fields["detection_type"] = DetectionTypeSensitiveFile
			alertRecord.Data.Fields["rule_id"] = fmt.Sprintf("%d", result.RuleID)
			alertRecord.Data.Fields["rule_name"] = result.RuleName
			alertRecord.Data.Fields["severity"] = result.Severity
			alertRecord.Data.Fields["rule_description"] = result.Description
			alertRecord.Data.Fields["matched_pattern"] = result.MatchedPattern
			alertRecord.Data.Fields["pid_tree"] = pidTreeStr
			alertRecord.Data.Fields["operator_user"] = resolveUsername(evt.UID)
			alertRecord.Data.Fields["operator_process"] = cstring(evt.Comm[:])
			alertRecord.Data.Fields["timestamp"] = fmt.Sprintf("%d", alertRecord.Timestamp)
			ctx.logger.Warn("Sensitive file operation detected",
				"rule_id", result.RuleID, "rule_name", result.RuleName, "severity", result.Severity,
				"action", actionStr, "new_path", newPath, "pid", evt.PID, "comm", cstring(evt.Comm[:]))
			if err := ctx.client.SendRecord(alertRecord); err != nil {
				ctx.logger.Error("Failed to send sensitive file alert record to agent", "error", err)
			}
		}
	}
	if err := ctx.client.SendRecord(record); err != nil {
		return fmt.Errorf("send file event record: %w", err)
	}
	return nil
}
