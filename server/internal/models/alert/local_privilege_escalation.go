package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// PrivilegeEscalation 本地提权告警实体
type PrivilegeEscalation struct {
	ID                int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID           string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID            *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName          string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP            string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	EscalatedUser     string          `json:"escalated_user" gorm:"column:escalated_user;not null"`
	ParentProcess     string          `json:"parent_process" gorm:"column:parent_process;not null"`
	ParentProcessUser string          `json:"parent_process_user" gorm:"column:parent_process_user;not null"`
	ProcessID         *int32          `json:"process_id,omitempty" gorm:"column:process_id"`
	ProcessPath       *string         `json:"process_path,omitempty" gorm:"column:process_path"`
	Status            int16           `json:"status" gorm:"column:status;not null;default:0"`
	WhitelistHit      bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID   *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	DiscoverTime      common.DateTime `json:"discover_time" gorm:"column:discover_time;not null"`
	CreatedAt         common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_privilege_escalation
func (PrivilegeEscalation) TableName() string {
	return "alert_privilege_escalation"
}

// PrivilegeEscalationStatus 状态枚举常量
const (
	PrivilegeEscalationStatusPending   = 0 // 待处理
	PrivilegeEscalationStatusProcessed = 1 // 已处理
	PrivilegeEscalationStatusIgnored   = 2 // 已忽略
)
