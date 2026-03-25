package vul

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ImageVulnScanTask 镜像漏洞扫描任务记录实体
type ImageVulnScanTask struct {
	ID            int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID       string          `json:"agent_id" gorm:"column:agent_id;not null;index:idx_ivst_agent_id"`
	ImageID       string          `json:"image_id" gorm:"column:image_id;not null;index:idx_ivst_image_id"`
	ImageName     string          `json:"image_name" gorm:"column:image_name;not null"`
	ScanStatus    int16           `json:"scan_status" gorm:"column:scan_status;not null;default:0;index:idx_ivst_scan_status"` // 0-进行中 1-成功 2-失败
	ScanTrigger   string          `json:"scan_trigger" gorm:"column:scan_trigger;size:16;default:auto"`                        // auto/manual
	TotalPackages *int32          `json:"total_packages,omitempty" gorm:"column:total_packages"`
	MatchedVulns  *int32          `json:"matched_vulns,omitempty" gorm:"column:matched_vulns"`
	ScanDuration  *int32          `json:"scan_duration,omitempty" gorm:"column:scan_duration"` // 扫描耗时(ms)
	ErrorMessage  *string         `json:"error_message,omitempty" gorm:"column:error_message;type:text"`
	ScanTime      common.DateTime `json:"scan_time" gorm:"column:scan_time;not null;index:idx_ivst_scan_time"`
	CreatedAt     common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 image_vuln_scan_task
func (ImageVulnScanTask) TableName() string {
	return "image_vuln_scan_task"
}
