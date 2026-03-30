package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ContainerSensitiveFile 容器核心文件监控告警实体
type ContainerSensitiveFile struct {
	ID              int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID         string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID          *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName        string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP          string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	ContainerID     string          `json:"container_id" gorm:"column:container_id;not null;index"`
	ContainerName   *string         `json:"container_name,omitempty" gorm:"column:container_name"`
	ImageName       *string         `json:"image_name,omitempty" gorm:"column:image_name"`
	RuleID          string          `json:"rule_id" gorm:"column:rule_id;not null"`
	RuleName        string          `json:"rule_name" gorm:"column:rule_name;not null"`
	Severity        string          `json:"severity" gorm:"column:severity;not null;index"`
	RuleDescription *string         `json:"rule_description,omitempty" gorm:"column:rule_description"`
	MatchedPattern  *string         `json:"matched_pattern,omitempty" gorm:"column:matched_pattern"`
	Action          string          `json:"action" gorm:"column:action;not null"`
	FilePath        string          `json:"file_path" gorm:"column:file_path;not null"`
	OldPath         *string         `json:"old_path,omitempty" gorm:"column:old_path"`
	OperatorUser    *string         `json:"operator_user,omitempty" gorm:"column:operator_user"`
	OperatorProcess *string         `json:"operator_process,omitempty" gorm:"column:operator_process"`
	Status          int16           `json:"status" gorm:"column:status;not null;default:0"` // 0-待处理 1-已处理 2-已忽略
	WhitelistHit    bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	AlertTime       common.DateTime `json:"alert_time" gorm:"column:alert_time;not null"`
	CreatedAt       common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_container_sensitive_file
func (ContainerSensitiveFile) TableName() string {
	return "alert_container_sensitive_file"
}
