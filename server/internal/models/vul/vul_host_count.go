package vul

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// HostVulnScanTask 主机漏洞扫描任务记录实体
type HostVulnScanTask struct {
	ID            int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID       string          `json:"agent_id" gorm:"column:agent_id;not null;index:idx_hvst_agent_id"`
	HostID        *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName      string          `json:"host_name" gorm:"column:host_name;not null;size:128"`
	HostIP        string          `json:"host_ip" gorm:"column:host_ip;not null;size:45;index:idx_hvst_host_ip"`
	ScanStatus    int16           `json:"scan_status" gorm:"column:scan_status;not null;default:0;index:idx_hvst_scan_status"` // 0-进行中 1-成功 2-失败
	ScanTrigger   string          `json:"scan_trigger" gorm:"column:scan_trigger;size:16;default:auto"`                        // auto/manual
	TotalPackages *int32          `json:"total_packages,omitempty" gorm:"column:total_packages"`
	MatchedVulns  *int32          `json:"matched_vulns,omitempty" gorm:"column:matched_vulns"`
	ScanDuration  *int32          `json:"scan_duration,omitempty" gorm:"column:scan_duration"` // 扫描耗时(ms)
	ErrorMessage  *string         `json:"error_message,omitempty" gorm:"column:error_message;type:text"`
	ScanTime      common.DateTime `json:"scan_time" gorm:"column:scan_time;not null;index:idx_hvst_scan_time"`
	CreatedAt     common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 host_vuln_scan_task
func (HostVulnScanTask) TableName() string {
	return "host_vuln_scan_task"
}

// 扫描任务状态常量
const (
	ScanStatusRunning int16 = 0 // 进行中
	ScanStatusSuccess int16 = 1 // 成功
	ScanStatusFailed  int16 = 2 // 失败
)

// 扫描触发方式常量
const (
	ScanTriggerAuto   = "auto"   // 定时自动扫描
	ScanTriggerManual = "manual" // 手动触发
)

// VulnCountHost 漏洞统计主机视图实体（基于v_vuln_count_hosts视图）
type VulnCountHost struct {
	AgentID       string          `json:"agent_id" gorm:"column:agent_id"`
	HostIP        string          `json:"host_ip" gorm:"column:host_ip"`
	HostName      string          `json:"host_name" gorm:"column:host_name"`
	LastScanTime  common.DateTime `json:"last_scan_time" gorm:"column:last_scan_time"`
	FirstScanTime common.DateTime `json:"first_scan_time" gorm:"column:first_scan_time"`
	CriticalVulns int64           `json:"critical_vulns" gorm:"column:critical_vulns"`
	HighVulns     int64           `json:"high_vulns" gorm:"column:high_vulns"`
	MediumVulns   int64           `json:"medium_vulns" gorm:"column:medium_vulns"`
	LowVulns      int64           `json:"low_vulns" gorm:"column:low_vulns"`
	TotalVulns    int64           `json:"total_vulns" gorm:"column:total_vulns"`
}

// TableName 指定视图名为 v_vuln_count_hosts
func (VulnCountHost) TableName() string {
	return "v_vuln_count_hosts"
}
