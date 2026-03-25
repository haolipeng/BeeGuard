package vul

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// VulnInfo 漏洞信息表（主机/容器共用）
type VulnInfo struct {
	ID            int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	CveID         string          `json:"cve_id" gorm:"column:cve_id;not null;uniqueIndex"`
	VulnName      string          `json:"vuln_name" gorm:"column:vuln_name;not null"`
	Severity      string          `json:"severity" gorm:"column:severity;not null;index"` // critical/high/medium/low
	CvssScore     *float64        `json:"cvss_score,omitempty" gorm:"column:cvss_score;index"`
	Description   *string         `json:"description,omitempty" gorm:"column:description;type:text"`
	FixSuggestion *string         `json:"fix_suggestion,omitempty" gorm:"column:fix_suggestion;type:text"`
	ReferenceURLs *string         `json:"reference_urls,omitempty" gorm:"column:reference_urls;type:text"`
	CreatedAt     common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (VulnInfo) TableName() string {
	return "vuln_info"
}

// 漏洞状态常量
const (
	VulnStatusUnfixed int16 = 0 // 未修复
	VulnStatusFixed   int16 = 1 // 已修复
	VulnStatusIgnored int16 = 2 // 已忽略
)

// 漏洞等级常量
const (
	VulnSeverityCritical = "critical"
	VulnSeverityHigh     = "high"
	VulnSeverityMedium   = "medium"
	VulnSeverityLow      = "low"
)
