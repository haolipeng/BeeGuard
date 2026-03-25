package back

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// HIDSRule 入侵检测告警规则实体
type HIDSRule struct {
	ID              int32            `json:"id" gorm:"primaryKey;not null"`                     // 规则序号，自增主键
	RuleName        string           `json:"rule_name" gorm:"not null;size:100"`                // 规则名称
	RuleFeature     string           `json:"rule_feature" gorm:"not null"`                      // 规则特征（如匹配的IP、端口、行为特征等）
	RuleLevel       string           `json:"rule_level" gorm:"not null;size:20"`                // 规则级别：低/中/高/紧急
	ThreatType      string           `json:"threat_type" gorm:"not null;size:50"`               // 威胁类型
	TriggerAction   string           `json:"trigger_action" gorm:"not null"`                    // 触发动作（如：发送邮件告警、阻断IP、记录日志等）
	RuleStatus      string           `json:"rule_status" gorm:"not null;default:'未生效';size:20"` // 规则状态：未生效/生效中/已停用/已删除
	EffectiveTime   *common.DateTime `json:"effective_time,omitempty"`                          // 规则生效时间
	RuleDescription *string          `json:"rule_description,omitempty"`                        // 规则详细描述
	CreatedAt       common.DateTime  `json:"created_at"`                                        // 规则创建时间
	UpdatedAt       common.DateTime  `json:"updated_at"`                                        // 规则更新时间
	//RulerType       string           `json:"ruler_type" gorm:"size:256"`               // 规则分类:高危命令、反弹shell、本地提权、异常登录、密码破解、恶意请求、网络攻击、文件查杀、核心文件监控
	RulerType string `json:"ruler_type"` // 规则分类:高危命令、反弹shell、本地提权、异常登录、密码破解、恶意请求、网络攻击、文件查杀、核心文件监控
}

// TableName 指定表名为 hids_rules
func (HIDSRule) TableName() string {
	return "hids_rules"
}

// RuleLevel 规则级别枚举常量
const (
	RuleLevelLow      = "低"
	RuleLevelMedium   = "中"
	RuleLevelHigh     = "高"
	RuleLevelCritical = "紧急"
)

// ThreatType 威胁类型枚举常量
const (
	ThreatTypeBruteForce   = "暴力破解"
	ThreatTypePortScan     = "端口扫描"
	ThreatTypeSQLInjection = "SQL注入"
	ThreatTypeXSS          = "XSS攻击"
	ThreatTypeMalware      = "恶意代码"
	ThreatTypeDDoS         = "DDoS"
	ThreatTypeOther        = "其他"
)

// RuleStatus 规则状态枚举常量
const (
	RuleStatusInactive = "未生效"
	RuleStatusActive   = "生效中"
	RuleStatusDisabled = "已停用"
	RuleStatusDeleted  = "已删除"
)

// RulerType 规则分类枚举常量
const (
	RulerTypeHighRiskCommand  = "高危命令"
	RulerTypeReverseShell     = "反弹shell"
	RulerTypeLocalPrivilege   = "本地提权"
	RulerTypeAbnormalLogin    = "异常登录"
	RulerTypePasswordCracking = "密码破解"
	RulerTypeMaliciousRequest = "恶意请求"
	RulerTypeNetworkAttack    = "网络攻击"
	RulerTypeFileScan         = "文件查杀"
	RulerTypeCoreFileMonitor  = "核心文件监控"
)
