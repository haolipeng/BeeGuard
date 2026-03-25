package vul

import (
	"encoding/json"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/models/common"

	"gorm.io/gorm"
)

// ImagesVulViewCount 镜像漏洞统计实体（基于视图 v_vuln_count_vuls）
type ImagesVulViewCount struct {
	VulnID            int64           `json:"vuln_id" gorm:"column:vuln_id"`                         // 漏洞ID
	CVEID             *string         `json:"cve_id" gorm:"column:cve_id"`                           // CVE编号
	VulnName          string          `json:"vuln_name" gorm:"column:vuln_name;not null"`            // 漏洞名称
	Severity          string          `json:"severity" gorm:"column:severity;not null"`              // 严重级别
	CVSSScore         *float64        `json:"cvss_score" gorm:"column:cvss_score"`                   // CVSS评分
	Description       *string         `json:"description" gorm:"column:description"`                 // 漏洞描述
	FixSuggestion     *string         `json:"fix_suggestion" gorm:"column:fix_suggestion"`           // 修复建议
	FirstScanTime     *time.Time      `json:"first_scan_time" gorm:"column:first_scan_time"`         // 首次扫描时间
	LastScanTime      *time.Time      `json:"last_scan_time" gorm:"column:last_scan_time"`           // 最后扫描时间
	AffectedImageCount int64          `json:"affected_image_count" gorm:"column:affected_image_count"` // 影响镜像数量
	AffectedImages    json.RawMessage `json:"affected_images" gorm:"column:affected_images"`          // 影响镜像详情（JSON数组）
}

// TableName 指定表名为视图 v_vuln_count_image_vuls
func (ImagesVulViewCount) TableName() string {
	return "v_vuln_count_image_vuls"
}

// AffectedImage 影响镜像详情结构
type AffectedImage struct {
	AgentID      *string    `json:"agent_id,omitempty"`      // Agent ID
	ImageID      *string    `json:"image_id,omitempty"`      // 镜像ID
	ImageName    string     `json:"image_name"`              // 镜像名称
	ScanTime     *time.Time `json:"scan_time,omitempty"`     // 扫描时间
	Status       *int       `json:"status,omitempty"`        // 状态
}

// ImageVulnerability 镜像漏洞基本信息实体
type ImageVulnerability struct {
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

// TableName 指定镜像漏洞表名
func (ImageVulnerability) TableName() string {
	return "image_vulnerability_info"
}

// ImageVulnStatus 状态枚举常量
const (
	ImageVulnStatusActive   = "active"   // 激活
	ImageVulnStatusInactive = "inactive" // 未激活
)