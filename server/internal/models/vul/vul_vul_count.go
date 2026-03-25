package vul

import (
	"encoding/json"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/models/common"

	"gorm.io/gorm"
)

// VulnWithHosts 漏洞主机统计实体（基于视图 v_vuln_with_hosts）
type VulnWithHosts struct {
	VulnID            int64           `json:"vuln_id" gorm:"column:vuln_id"`                         // 漏洞ID
	CVEID             *string         `json:"cve_id" gorm:"column:cve_id"`                           // CVE编号
	VulnName          string          `json:"vuln_name" gorm:"column:vuln_name;not null"`            // 漏洞名称
	Severity          string          `json:"severity" gorm:"column:severity;not null"`              // 严重级别
	CVSSScore         *float64        `json:"cvss_score" gorm:"column:cvss_score"`                   // CVSS评分
	Description       *string         `json:"description" gorm:"column:description"`                 // 漏洞描述
	FixSuggestion     *string         `json:"fix_suggestion" gorm:"column:fix_suggestion"`           // 修复建议
	FirstScanTime     *time.Time      `json:"first_scan_time" gorm:"column:first_scan_time"`         // 首次扫描时间
	LastScanTime      *time.Time      `json:"last_scan_time" gorm:"column:last_scan_time"`           // 最后扫描时间
	AffectedHostCount int64           `json:"affected_host_count" gorm:"column:affected_host_count"` // 影响主机数量
	AffectedHosts     json.RawMessage `json:"affected_hosts" gorm:"column:affected_hosts"`           // 影响主机详情（JSON数组）
}

// TableName 指定表名为视图 v_vuln_with_hosts
func (VulnWithHosts) TableName() string {
	return "v_vuln_count_vuls"
}

// AffectedHost 影响主机详情结构
type AffectedHost struct {
	HostID   *int64     `json:"host_id,omitempty"`   // 主机ID
	HostName string     `json:"host_name"`           // 主机名称
	HostIP   string     `json:"host_ip"`             // 主机IP
	ScanTime *time.Time `json:"scan_time,omitempty"` // 扫描时间
	Status   *int       `json:"status,omitempty"`    // 状态
}

// Vulnerability 漏洞基本信息实体
type Vulnerability struct {
	ID            int64           `json:"id" gorm:"primaryKey;not null;autoIncrement"` // 主键ID
	CVEID         *string         `json:"cve_id" gorm:"index"`                         // CVE编号
	VulnName      string          `json:"vuln_name" gorm:"not null;size:255"`          // 漏洞名称
	Severity      string          `json:"severity" gorm:"not null;size:20"`            // 严重级别: critical/high/medium/low
	CVSSScore     *float64        `json:"cvss_score"`                                  // CVSS评分
	Description   *string         `json:"description" gorm:"type:text"`                // 漏洞描述
	FixSuggestion *string         `json:"fix_suggestion" gorm:"type:text"`             // 修复建议
	Reference     *string         `json:"reference" gorm:"type:text"`                  // 参考链接
	PublishDate   *time.Time      `json:"publish_date"`                                // 发布日期
	UpdateTime    *time.Time      `json:"update_time"`                                 // 更新时间
	Status        string          `json:"status" gorm:"not null;default:'active'"`     // 状态: active/inactive
	CreatedAt     common.DateTime `json:"created_at"`
	UpdatedAt     common.DateTime `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// TableName 指定漏洞表名
func (Vulnerability) TableName() string {
	return "vulnerability_info"
}

// Severity 严重级别枚举常量
const (
	SeverityCritical = "critical" // 严重
	SeverityHigh     = "high"     // 高危
	SeverityMedium   = "medium"   // 中危
	SeverityLow      = "low"      // 低危
)

// VulnStatus 状态枚举常量
const (
	VulnStatusActive   = "active"   // 激活
	VulnStatusInactive = "inactive" // 未激活
)
