package vul

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// HostVulnDetail 主机漏洞发现记录实体
type HostVulnDetail struct {
	ID               int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	ScanID           int64           `json:"scan_id" gorm:"column:scan_id;not null;index"`
	AgentID          string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID           *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	VulnID           int64           `json:"vuln_id" gorm:"column:vuln_id;not null;index"`
	CveID            string          `json:"cve_id" gorm:"column:cve_id;not null;index"`
	PackageName      string          `json:"package_name" gorm:"column:package_name;not null"`
	InstalledVersion *string         `json:"installed_version,omitempty" gorm:"column:installed_version"`
	FixedVersion     *string         `json:"fixed_version,omitempty" gorm:"column:fixed_version"`
	Status           int16           `json:"status" gorm:"column:status;not null;default:0;index:idx_hvd_status"` // 0-未修复 1-已修复 2-已忽略
	ScanTime         common.DateTime `json:"scan_time" gorm:"column:scan_time;not null;index:idx_hvd_scan_time"`
	CreatedAt        common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	HostName         *string         `json:"host_name,omitempty" gorm:"column:host_name"`
	HostIP           *string         `json:"host_ip,omitempty" gorm:"column:host_ip"`
	VulnName         *string         `json:"vuln_name,omitempty" gorm:"column:vuln_name"`
	Severity         *string         `json:"severity,omitempty" gorm:"column:severity"`
	CvssScore        *float64        `json:"cvss_score,omitempty" gorm:"column:cvss_score"`
	Description      *string         `json:"description,omitempty" gorm:"column:description"`
	FixSuggestion    *string         `json:"fix_suggestion,omitempty" gorm:"column:fix_suggestion"`
}

// TableName 指定表名
func (HostVulnDetail) TableName() string {
	return "host_vuln_detail"
}
