package mapper

import (
	"strconv"
	"strings"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// MapDNS DNS查询事件字段映射: Agent -> 数据库
// Agent字段: pid, tgid, ppid, uid, comm, exe_path, domain, query_type, timestamp
// 数据库字段: pid, tgid, ppid, uid, comm, exe_path, domain, query_type, event_time
func MapDNS(fields map[string]string, ctx *AgentContext) *host.DNS {
	pid, _ := strconv.Atoi(fields["pid"])
	tgid, _ := strconv.Atoi(fields["tgid"])
	ppid, _ := strconv.Atoi(fields["ppid"])
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

	return &host.DNS{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		PID:       pid,
		TGID:      tgid,
		PPID:      ppid,
		UID:       uid,
		Comm:      fields["comm"],
		ExePath:   fields["exe_path"],
		Domain:    fields["domain"],
		QueryType: fields["query_type"],
		PidTree:   fields["pid_tree"],
		EventTime: common.DateTime{Time: eventTime},
	}
}
