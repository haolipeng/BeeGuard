package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ReverseShell 反弹shell告警实体
type ReverseShell struct {
	ID          int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID     string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID      *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName    string          `json:"host_name" gorm:"column:host_name;not null"`
	VictimIP    string          `json:"victim_ip" gorm:"column:victim_ip;not null"`
	CommandLine string          `json:"command_line" gorm:"column:command_line;not null"`
	ShellType   *string         `json:"shell_type,omitempty" gorm:"column:shell_type"`
	TargetHost  string          `json:"target_host" gorm:"column:target_host;not null"`
	TargetPort  int32           `json:"target_port" gorm:"column:target_port;not null"`
	Status          int16           `json:"status" gorm:"column:status;not null;default:0"`
	WhitelistHit    bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	EventTime       common.DateTime `json:"event_time" gorm:"column:event_time;not null"`
	CreatedAt       common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_reverse_shell
func (ReverseShell) TableName() string {
	return "alert_reverse_shell"
}

// ShellType 枚举常量
const (
	ShellTypeBash       = "bash"
	ShellTypePython     = "python"
	ShellTypeNc         = "nc"
	ShellTypePerl       = "perl"
	ShellTypePHP        = "php"
	ShellTypeRuby       = "ruby"
	ShellTypePowershell = "powershell"
)

// ReverseShellStatus 状态枚举常量
const (
	ReverseShellStatusPending   = 0 // 待处理
	ReverseShellStatusProcessed = 1 // 已处理
	ReverseShellStatusIgnored   = 2 // 已忽略
)
