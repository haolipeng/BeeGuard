package vul

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ImageVulnDetail 镜像漏洞发现记录实体
type ImageVulnDetail struct {
	ID               int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	ScanID           int64           `json:"scan_id" gorm:"column:scan_id;not null;index:idx_ivd_scan_id"`
	AgentID          string          `json:"agent_id" gorm:"column:agent_id;not null;index:idx_ivd_agent_id"`
	ImageID          string          `json:"image_id" gorm:"column:image_id;not null;index:idx_ivd_image_id"`
	VulnID           int64           `json:"vuln_id" gorm:"column:vuln_id;not null"`
	CveID            string          `json:"cve_id" gorm:"column:cve_id;not null;index"`
	PackageName      string          `json:"package_name" gorm:"column:package_name;not null"`
	InstalledVersion *string         `json:"installed_version,omitempty" gorm:"column:installed_version"`
	FixedVersion     *string         `json:"fixed_version,omitempty" gorm:"column:fixed_version"`
	Status           int16           `json:"status" gorm:"column:status;not null;default:0"` // 0-未修复 1-已修复 2-已忽略
	ScanTime         common.DateTime `json:"scan_time" gorm:"column:scan_time;not null;index"`
	CreatedAt        common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	ImageName        string          `json:"image_name" gorm:"column:image_name"`
	VulnName         string          `json:"vuln_name" gorm:"column:vuln_name"`
	Severity         string          `json:"severity" gorm:"column:severity"`
	CVSSScore        *float64        `json:"cvss_score,omitempty" gorm:"column:cvss_score"`
	Description      *string         `json:"description,omitempty" gorm:"column:description"`
	FixSuggestion    *string         `json:"fix_suggestion,omitempty" gorm:"column:fix_suggestion"`
}

// TableName 指定表名为 image_vuln_detail
func (ImageVulnDetail) TableName() string {
	return "image_vuln_detail"
}

// ImageVulnStatus 状态枚举常量
const (
	ImageVulnStatusUnfixed = 0 // 未修复
	ImageVulnStatusFixed   = 1 // 已修复
	ImageVulnStatusIgnored = 2 // 已忽略
)

// Severity 漏洞等级枚举常量
const (
	ImageSeverityCritical = "critical" // 严重
	ImageSeverityHigh     = "high"     // 高危
	ImageSeverityMedium   = "medium"   // 中危
	ImageSeverityLow      = "low"      // 低危
)

// VulnCountImage 漏洞统计镜像视图实体（基于v_vuln_count_images视图）
type VulnCountImage struct {
	ImageID       string          `json:"image_id" gorm:"column:image_id"`
	ImageName     string          `json:"image_name" gorm:"column:image_name"`
	LastScanTime  common.DateTime `json:"last_scan_time" gorm:"column:last_scan_time"`
	FirstScanTime common.DateTime `json:"first_scan_time" gorm:"column:first_scan_time"`
	CriticalVulns int64           `json:"critical_vulns" gorm:"column:critical_vulns"`
	HighVulns     int64           `json:"high_vulns" gorm:"column:high_vulns"`
	MediumVulns   int64           `json:"medium_vulns" gorm:"column:medium_vulns"`
	LowVulns      int64           `json:"low_vulns" gorm:"column:low_vulns"`
	TotalVulns    int64           `json:"total_vulns" gorm:"column:total_vulns"`
}

// TableName 指定视图名为 v_vuln_count_images
func (VulnCountImage) TableName() string {
	return "v_vuln_count_images"
}
