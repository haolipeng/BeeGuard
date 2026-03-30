package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ContainerDangerousCommand 容器高危命令告警实体
type ContainerDangerousCommand struct {
	ID             int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID        string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID         *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName       string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP         string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	ContainerID    string          `json:"container_id" gorm:"column:container_id;not null;index"`
	ContainerName  *string         `json:"container_name,omitempty" gorm:"column:container_name"`
	ImageName      *string         `json:"image_name,omitempty" gorm:"column:image_name"`
	Command        string          `json:"command" gorm:"column:command;not null"`
	CommandType    string          `json:"command_type" gorm:"column:command_type;not null;index"`
	User           string          `json:"user" gorm:"column:user;not null"`
	PrivilegeLevel string          `json:"privilege_level" gorm:"column:privilege_level;not null"`
	Status          int16           `json:"status" gorm:"column:status;not null;default:0"` // 0-待处理 1-已处理 2-已忽略
	WhitelistHit    bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	AlertTime       common.DateTime `json:"alert_time" gorm:"column:alert_time;not null"`
	CreatedAt       common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_container_dangerous_command
func (ContainerDangerousCommand) TableName() string {
	return "alert_container_dangerous_command"
}
