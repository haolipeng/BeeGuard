package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// AbnormalLogin 异常登录告警实体
type AbnormalLogin struct {
	ID             int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID        string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID         *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName       string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP         string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	SourceIP       string          `json:"source_ip" gorm:"column:source_ip;not null;index"`
	SourceLocation *string         `json:"source_location,omitempty" gorm:"column:source_location"`
	LoginUser      string          `json:"login_user" gorm:"column:login_user;not null"`
	LoginTime      common.DateTime `json:"login_time" gorm:"column:login_time;not null"`
	RiskLevel      string          `json:"risk_level" gorm:"column:risk_level;not null"`
	AbnormalType   *string         `json:"abnormal_type,omitempty" gorm:"column:abnormal_type"`
	Status          int16           `json:"status" gorm:"column:status;not null;default:0"`
	WhitelistHit    bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	IsWhitelist     *int16          `json:"is_whitelist,omitempty" gorm:"column:is_whitelist;default:0"`
	CreatedAt       common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_abnormal_login
func (AbnormalLogin) TableName() string {
	return "alert_abnormal_login"
}

// RiskLevel 风险等级枚举常量
const (
	RiskLevelLow      = "low"      // 低风险
	RiskLevelMedium   = "medium"   // 中风险
	RiskLevelHigh     = "high"     // 高风险
	RiskLevelCritical = "critical" // 危急
)

// AbnormalType 异常类型枚举常量
const (
	AbnormalTypeUnknownIP = "unknown_ip"    // 未知IP
	AbnormalTypeTime      = "abnormal_time" // 异常时间
	AbnormalTypeUser      = "abnormal_user" // 异常用户
)

// LoginStatus 状态枚举常量
const (
	LoginStatusPending   = 0 // 待处理
	LoginStatusProcessed = 1 // 已处理
	LoginStatusIgnored   = 2 // 已忽略
)
