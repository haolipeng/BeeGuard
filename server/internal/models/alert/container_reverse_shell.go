package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ContainerReverseShell 容器反弹Shell告警实体
type ContainerReverseShell struct {
	ID            int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID       string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID        *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName      string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP        string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	ContainerID   string          `json:"container_id" gorm:"column:container_id;not null;index"`
	ContainerName *string         `json:"container_name,omitempty" gorm:"column:container_name"`
	ImageName     *string         `json:"image_name,omitempty" gorm:"column:image_name"`
	PID           int32           `json:"pid" gorm:"column:pid;not null"`
	PPID          *int32          `json:"ppid,omitempty" gorm:"column:ppid"`
	UID           string          `json:"uid" gorm:"column:uid;not null"`
	Comm          string          `json:"comm" gorm:"column:comm;not null"`
	ExePath       *string         `json:"exe_path,omitempty" gorm:"column:exe_path"`
	Args          *string         `json:"args,omitempty" gorm:"column:args"`
	ShellType     *string         `json:"shell_type,omitempty" gorm:"column:shell_type;index"`
	RemoteIP      string          `json:"remote_ip" gorm:"column:remote_ip;not null;index"`
	RemotePort    int32           `json:"remote_port" gorm:"column:remote_port;not null"`
	Status          int16           `json:"status" gorm:"column:status;not null;default:0;index"` // 0-待处理 1-已处理 2-已忽略
	WhitelistHit    bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	EventTime       common.DateTime `json:"event_time" gorm:"column:event_time;not null;index"`
	CreatedAt       common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_container_reverse_shell
func (ContainerReverseShell) TableName() string {
	return "alert_container_reverse_shell"
}
