package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// MaliciousRequest 恶意请求告警实体
type MaliciousRequest struct {
	ID               int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID          string           `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID           *int64           `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName         string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP           string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	PolicyType       string           `json:"policy_type" gorm:"column:policy_type;not null"`
	PolicyName       string           `json:"policy_name" gorm:"column:policy_name;not null"`
	MaliciousDomain  string           `json:"malicious_domain" gorm:"column:malicious_domain;not null;index"`
	MaliciousIP      *string          `json:"malicious_ip,omitempty" gorm:"column:malicious_ip;index"`
	RequestCount     int32            `json:"request_count" gorm:"column:request_count;not null;default:1"`
	FirstRequestTime *common.DateTime `json:"first_request_time,omitempty" gorm:"column:first_request_time"`
	LastRequestTime  *common.DateTime `json:"last_request_time,omitempty" gorm:"column:last_request_time"`
	RiskDescription  *string          `json:"risk_description,omitempty" gorm:"column:risk_description"`
	Status           int16            `json:"status" gorm:"column:status;not null;default:0"`
	CreatedAt        common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_malicious_request
func (MaliciousRequest) TableName() string {
	return "alert_malicious_request"
}

// PolicyType 策略类型枚举常量
const (
	PolicyTypeDNSFilter       = "dns_filter"       // DNS过滤
	PolicyTypeIPBlacklist     = "ip_blacklist"      // IP黑名单
	PolicyTypeURLFilter       = "url_filter"        // URL过滤
	PolicyTypeDomainFilter    = "domain_filter"     // 域名过滤
	PolicyTypeBehaviorAnalyze = "behavior_analyze"  // 行为分析
	PolicyTypeThreatIntel     = "threat_intel"      // 威胁情报
)

// MaliciousRequestStatus 状态枚举常量
const (
	MaliciousRequestStatusPending   = 0 // 待处理
	MaliciousRequestStatusProcessed = 1 // 已处理
	MaliciousRequestStatusIgnored   = 2 // 已忽略
)
