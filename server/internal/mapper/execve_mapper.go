package mapper

import (
	"strconv"
	"strings"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// MapExecve execve事件字段映射: Agent -> 数据库
// Agent字段: pid, tgid, ppid, pgid, uid, comm, exe_path, args, timestamp
// 数据库字段: pid, tgid, ppid, pgid, uid, comm, exe_path, args, event_time
func MapExecve(fields map[string]string, ctx *AgentContext) *host.Execve {
	pid, _ := strconv.Atoi(fields["pid"])
	tgid, _ := strconv.Atoi(fields["tgid"])
	ppid, _ := strconv.Atoi(fields["ppid"])
	pgid, _ := strconv.Atoi(fields["pgid"])
	uid, _ := strconv.Atoi(fields["uid"])

	// 事件时间转换（从Unix时间戳）
	var eventTime time.Time
	if ts := fields["timestamp"]; ts != "" {
		if sec, err := strconv.ParseInt(ts, 10, 64); err == nil && sec > 0 {
			eventTime = time.Unix(sec, 0)
		}
	}
	// 如果没有timestamp或解析失败，使用当前时间
	if eventTime.IsZero() {
		eventTime = time.Now()
	}

	return &host.Execve{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		PID:       pid,
		TGID:      tgid,
		PPID:      ppid,
		PGID:      pgid,
		UID:       uid,
		Comm:      fields["comm"],
		ExePath:   fields["exe_path"],
		Args:      fields["args"],
		EventTime: common.DateTime{Time: eventTime},
	}
}
