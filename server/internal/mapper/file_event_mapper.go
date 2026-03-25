package mapper

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// MapFileEvent 文件事件字段映射: Agent -> 数据库
// Agent字段: pid, tgid, ppid, uid, comm, exe_path, action, new_path, old_path, s_id, pid_tree,
//
//	socket_pid, remote_ip, remote_port, local_ip, local_port, timestamp
func MapFileEvent(fields map[string]string, ctx *AgentContext) *host.FileEvent {
	pid, _ := strconv.Atoi(fields["pid"])
	tgid, _ := strconv.Atoi(fields["tgid"])
	ppid, _ := strconv.Atoi(fields["ppid"])
	uid, _ := strconv.Atoi(fields["uid"])
	socketPID, _ := strconv.Atoi(fields["socket_pid"])
	remotePort := parsePort(fields["remote_port"], "remote_port")
	localPort := parsePort(fields["local_port"], "local_port")

	// 事件时间转换（从Unix时间戳）
	var eventTime time.Time
	if ts := fields["timestamp"]; ts != "" {
		if sec, err := strconv.ParseInt(ts, 10, 64); err == nil && sec > 0 {
			eventTime = time.Unix(sec, 0)
		}
	}
	if eventTime.IsZero() {
		eventTime = time.Now()
	}

	return &host.FileEvent{
		AgentID:    ctx.AgentID,
		HostName:   ctx.HostName,
		HostIP:     strings.Join(ctx.HostIP, ","),
		PID:        pid,
		TGID:       tgid,
		PPID:       ppid,
		UID:        uid,
		Comm:       fields["comm"],
		ExePath:    fields["exe_path"],
		Action:     fields["action"],
		NewPath:    fields["new_path"],
		OldPath:    fields["old_path"],
		SID:        fields["s_id"],
		PidTree:    fields["pid_tree"],
		SocketPID:  socketPID,
		RemoteIP:   fields["remote_ip"],
		RemotePort: remotePort,
		LocalIP:    fields["local_ip"],
		LocalPort:  localPort,
		EventTime:  common.DateTime{Time: eventTime},
	}
}

// MapFileIntegrityAlert 敏感文件告警字段映射: Agent -> 数据库
// Agent字段: detection_type, rule_id, rule_name, severity, rule_description, matched_pattern,
//
//	action, new_path, operator_user, operator_process, pid_tree, timestamp
func MapFileIntegrityAlert(fields map[string]string, ctx *AgentContext) *alert.FileIntegrity {
	// 解析 rule_id（int64）
	ruleID, _ := strconv.ParseInt(fields["rule_id"], 10, 64)
	var ruleIDPtr *int64
	if ruleID > 0 {
		ruleIDPtr = &ruleID
	}

	// 从 new_path 解析文件名
	newPath := fields["new_path"]
	var fileNamePtr *string
	if newPath != "" {
		fileName := filepath.Base(newPath)
		fileNamePtr = &fileName
	}

	// 操作用户
	var operatorUserPtr *string
	if v := fields["operator_user"]; v != "" {
		operatorUserPtr = &v
	}

	// 操作进程
	var operatorProcessPtr *string
	if v := fields["operator_process"]; v != "" {
		operatorProcessPtr = &v
	}

	// 告警描述
	var alertDescPtr *string
	if v := fields["rule_description"]; v != "" {
		alertDescPtr = &v
	}

	// 告警时间
	var alertTime time.Time
	if ts, err := strconv.ParseInt(fields["timestamp"], 10, 64); err == nil && ts > 0 {
		alertTime = time.Unix(ts, 0)
	} else {
		alertTime = time.Now()
	}

	return &alert.FileIntegrity{
		AgentID:          ctx.AgentID,
		HostName:         ctx.HostName,
		HostIP:           strings.Join(ctx.HostIP, ","),
		RuleType:         fields["detection_type"],
		RuleName:         fields["rule_name"],
		RuleID:           ruleIDPtr,
		ThreatLevel:      fields["severity"],
		ThreatAction:     fields["action"],
		FilePath:         newPath,
		FileName:         fileNamePtr,
		OperatorUser:     operatorUserPtr,
		OperatorProcess:  operatorProcessPtr,
		AlertDescription: alertDescPtr,
		Status:           0,
		AlertTime:        common.DateTime{Time: alertTime},
		UpdatedAt:        common.DateTime{Time: time.Now()},
	}
}
