package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// FileIntegrity 文件完整性告警实体
type FileIntegrity struct {
	ID              int64           `json:"id" gorm:"primaryKey;not null"`              // 主键ID
	AgentID         string          `json:"agent_id" gorm:"not null;index:idx_alert_fi_agent_id"` // Agent唯一标识
	HostID          *int64          `json:"host_id,omitempty"`                           // 主机ID
	HostName        string          `json:"host_name" gorm:"not null"`                  // 主机名称
	HostIP          string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`                    // 主机IP
	RuleType        string          `json:"rule_type" gorm:"not null;index:idx_alert_fi_rule_type"` // 规则类型
	RuleName        string          `json:"rule_name" gorm:"not null"`                  // 命中规则名称
	RuleID          *int64          `json:"rule_id,omitempty"`                           // 关联规则ID
	ThreatLevel     string          `json:"threat_level" gorm:"not null;index:idx_alert_fi_threat_level"` // 威胁等级
	ThreatAction    string          `json:"threat_action" gorm:"not null"`              // 威胁行为
	FilePath        string          `json:"file_path" gorm:"not null;index:idx_alert_fi_file_path"` // 文件路径
	FileName        *string         `json:"file_name,omitempty"`                         // 文件名
	OldContentHash  *string         `json:"old_content_hash,omitempty"`                  // 原内容哈希
	NewContentHash  *string         `json:"new_content_hash,omitempty"`                  // 新内容哈希
	ChangeDetail    *string         `json:"change_detail,omitempty"`                     // 变更详情
	OperatorUser    *string         `json:"operator_user,omitempty"`                     // 操作用户
	OperatorProcess *string         `json:"operator_process,omitempty"`                  // 操作进程
	AlertDescription *string        `json:"alert_description,omitempty"`                 // 告警描述
	Status          int16           `json:"status" gorm:"not null;default:0;index:idx_alert_fi_status"` // 状态: 0-待处理 1-已处理 2-已忽略
	WhitelistHit    bool            `json:"whitelist_hit" gorm:"column:whitelist_hit;default:false"`
	WhitelistRuleID *int64          `json:"whitelist_rule_id,omitempty" gorm:"column:whitelist_rule_id"`
	AlertTime       common.DateTime `json:"alert_time" gorm:"not null;index:idx_alert_fi_alert_time"` // 告警时间
	CreatedAt       common.DateTime `json:"created_at"`                                  // 创建时间
	UpdatedAt       common.DateTime `json:"updated_at"`                                  // 更新时间
}

// TableName 指定表名为 alert_file_integrity
func (FileIntegrity) TableName() string {
	return "alert_file_integrity"
}

// RuleType 规则类型枚举常量
const (
	RuleTypeCoreFile    = "core_file"    // 核心文件
	RuleTypeConfigFile  = "config_file"  // 配置文件
	RuleTypeSystemFile  = "system_file"  // 系统文件
	RuleTypeBinaryFile  = "binary_file"  // 二进制文件
	RuleTypeLogFile     = "log_file"     // 日志文件
	RuleTypeScriptFile  = "script_file"  // 脚本文件
)

// ThreatLevel 威胁等级枚举常量
const (
	ThreatLevelLow    = "low"    // 低威胁
	ThreatLevelMedium = "medium" // 中威胁
	ThreatLevelHigh   = "high"   // 高威胁
)

// ThreatAction 威胁行为枚举常量
const (
	ThreatActionAdd    = "add"    // 新增
	ThreatActionModify = "modify" // 修改
	ThreatActionDelete = "delete" // 删除
)

// FileIntegrityStatus 状态枚举常量
const (
	FileIntegrityStatusPending   = 0 // 待处理
	FileIntegrityStatusProcessed = 1 // 已处理
	FileIntegrityStatusIgnored   = 2 // 已忽略
)