package view

import "github.com/haolipeng/BeeGuard/server/internal/models/common"

// CodeQLVulnSummary 代码仓库漏洞统计视图
type CodeQLVulnSummary struct {
	RepoID        int64           `json:"repo_id" gorm:"column:repo_id"`
	RepoName      string          `json:"repo_name" gorm:"column:repo_name"`
	ProjectName   string          `json:"project_name" gorm:"column:project_name"`
	TotalVulns    int64           `json:"total_vulns" gorm:"column:total_vulns"`
	CriticalCount int64           `json:"critical_count" gorm:"column:critical_count"`
	HighCount     int64           `json:"high_count" gorm:"column:high_count"`
	MediumCount   int64           `json:"medium_count" gorm:"column:medium_count"`
	LowCount      int64           `json:"low_count" gorm:"column:low_count"`
	LastScanTime  common.DateTime `json:"last_scan_time" gorm:"column:last_scan_time"`
}

func (CodeQLVulnSummary) TableName() string {
	return "v_views_codeql_vuln_summary"
}

// ImageVulnTop5ByCVE 容器镜像漏洞top5视图
type ImageVulnTop5ByCVE struct {
	CVEID                string   `json:"cve_id" gorm:"column:cve_id"`
	VulnName             string   `json:"vuln_name" gorm:"column:vuln_name"`
	Severity             string   `json:"severity" gorm:"column:severity"`
	CVSSScore            *float64 `json:"cvss_score" gorm:"column:cvss_score"`
	AffectedImageCount   int64    `json:"affected_image_count" gorm:"column:affected_image_count"`
	TotalInstances       int64    `json:"total_instances" gorm:"column:total_instances"`
	PendingInstances     int64    `json:"pending_instances" gorm:"column:pending_instances"`
	FixedInstances       int64    `json:"fixed_instances" gorm:"column:fixed_instances"`
	AffectedImagesSample string   `json:"affected_images_sample" gorm:"column:affected_images_sample"`
}

func (ImageVulnTop5ByCVE) TableName() string {
	return "v_views_image_vuln_top5_by_cve_all"
}

// TotalAlertHourlyStats 每小时告警趋势视图
type TotalAlertHourlyStats struct {
	HourBucket     common.DateTime `json:"hour_bucket" gorm:"column:hour_bucket"`
	TotalAlerts    int64           `json:"total_alerts" gorm:"column:total_alerts"`
	PendingCount   int64           `json:"pending_count" gorm:"column:pending_count"`
	ProcessedCount int64           `json:"processed_count" gorm:"column:processed_count"`
	IgnoredCount   int64           `json:"ignored_count" gorm:"column:ignored_count"`
}

func (TotalAlertHourlyStats) TableName() string {
	return "v_views_total_alert_hourly_stats"
}

// TotalAlertMonthlyStats 每月告警数视图
type TotalAlertMonthlyStats struct {
	MonthBucket    common.DateTime `json:"month_bucket" gorm:"column:month_bucket"`
	TotalAlerts    int64           `json:"total_alerts" gorm:"column:total_alerts"`
	PendingCount   int64           `json:"pending_count" gorm:"column:pending_count"`
	ProcessedCount int64           `json:"processed_count" gorm:"column:processed_count"`
	IgnoredCount   int64           `json:"ignored_count" gorm:"column:ignored_count"`
	AvgDailyAlerts float64         `json:"avg_daily_alerts" gorm:"column:avg_daily_alerts"`
}

func (TotalAlertMonthlyStats) TableName() string {
	return "v_views_total_alert_monthly_stats"
}

// VulnCountImageVuls 容器风险-漏洞视图top2
type VulnCountImageVuls struct {
	VulnID             int64           `json:"vuln_id" gorm:"column:vuln_id"`
	CVEID              string          `json:"cve_id" gorm:"column:cve_id"`
	VulnName           string          `json:"vuln_name" gorm:"column:vuln_name"`
	Severity           string          `json:"severity" gorm:"column:severity"`
	CVSSScore          *float64        `json:"cvss_score" gorm:"column:cvss_score"`
	Description        string          `json:"description" gorm:"column:description"`
	FixSuggestion      string          `json:"fix_suggestion" gorm:"column:fix_suggestion"`
	FirstScanTime      common.DateTime `json:"first_scan_time" gorm:"column:first_scan_time"`
	LastScanTime       common.DateTime `json:"last_scan_time" gorm:"column:last_scan_time"`
	AffectedImageCount int64           `json:"affected_image_count" gorm:"column:affected_image_count"`
	AffectedImages     string          `json:"affected_images" gorm:"column:affected_images"`
}

func (VulnCountImageVuls) TableName() string {
	return "v_views_vuln_count_image_vuls"
}

// VulnCountVuls 主机风险-漏洞视图top2
type VulnCountVuls struct {
	VulnID            int64           `json:"vuln_id" gorm:"column:vuln_id"`
	CVEID             string          `json:"cve_id" gorm:"column:cve_id"`
	VulnName          string          `json:"vuln_name" gorm:"column:vuln_name"`
	Severity          string          `json:"severity" gorm:"column:severity"`
	CVSSScore         *float64        `json:"cvss_score" gorm:"column:cvss_score"`
	Description       string          `json:"description" gorm:"column:description"`
	FixSuggestion     string          `json:"fix_suggestion" gorm:"column:fix_suggestion"`
	FirstScanTime     common.DateTime `json:"first_scan_time" gorm:"column:first_scan_time"`
	LastScanTime      common.DateTime `json:"last_scan_time" gorm:"column:last_scan_time"`
	AffectedHostCount int64           `json:"affected_host_count" gorm:"column:affected_host_count"`
	AffectedHosts     string          `json:"affected_hosts" gorm:"column:affected_hosts"`
}

func (VulnCountVuls) TableName() string {
	return "v_views_vuln_count_vuls"
}

// BaselineItemTop5Affected 合规基线检测项top5视图
type BaselineItemTop5Affected struct {
	ItemID        int64  `json:"item_id" gorm:"column:item_id"`
	ItemName      string `json:"item_name" gorm:"column:item_name"`
	CheckCount    int64  `json:"check_count" gorm:"column:check_count"`
	FailedCount   int64  `json:"failed_count" gorm:"column:failed_count"`
	FailedRate    string `json:"failed_rate" gorm:"column:failed_rate"`
	AffectedHosts string `json:"affected_hosts" gorm:"column:affected_hosts"`
}

func (BaselineItemTop5Affected) TableName() string {
	return "v_views_baseline_item_top5_affected"
}

// HostStatusSummary 在线主机视图
type HostStatusSummary struct {
	TotalHosts   int64  `json:"total_hosts" gorm:"column:total_hosts"`
	OnlineHosts  int64  `json:"online_hosts" gorm:"column:online_hosts"`
	OfflineHosts int64  `json:"offline_hosts" gorm:"column:offline_hosts"`
	OnlineRate   string `json:"online_rate" gorm:"column:online_rate"`
}

func (HostStatusSummary) TableName() string {
	return "v_views_host_status_summary"
}

// HostVulnStats 主机风险资产视图
type HostVulnStats struct {
	TotalCount int64   `json:"total_count" gorm:"column:total_count"` // 总风险主机数
	TodayCount int64   `json:"today_count" gorm:"column:today_count"` // 今日新增风险主机数
	Percentage float64 `json:"percentage" gorm:"column:percentage"`   // 今日新增占比百分比
}

func (HostVulnStats) TableName() string {
	return "v_views_host_vuln_stats"
}

// HostVulnPackageTop5 风险资产分布TOP5视图
type HostVulnPackageTop5 struct {
	PackageName     string `json:"package_name" gorm:"column:package_name"`
	OccurrenceCount int64  `json:"occurrence_count" gorm:"column:occurrence_count"`
	Rank            int64  `json:"rank" gorm:"column:rank"`
}

func (HostVulnPackageTop5) TableName() string {
	return "v_view_host_vuln_package_top5"
}

// ThreatTypeTotalCount 威胁类型统计视图
type ThreatTypeTotalCount struct {
	ThreatType string `json:"threat_type" gorm:"column:threat_type"`
	Count      int64  `json:"count" gorm:"column:count"`
}

func (ThreatTypeTotalCount) TableName() string {
	return "v_view_threat_type_total_count"
}

// VulnChartData 安全看板漏洞统计视图
type VulnChartData struct {
	ID                string `json:"id" gorm:"column:id"`
	Title             string `json:"title" gorm:"column:title"`
	Severity          string `json:"severity" gorm:"column:severity"`
	AffectedHostCount int64  `json:"affected_host_count" gorm:"column:affected_host_count"`
}

func (VulnChartData) TableName() string {
	return "v_view_vuln_chart_data"
}

// HostBaselineFailTop5 基线检查失败主机top5视图
type HostBaselineFailTop5 struct {
	AgentID   string `json:"agent_id" gorm:"column:agent_id"`
	HostIP    string `json:"host_ip" gorm:"column:host_ip"`
	HostName  string `json:"host_name" gorm:"column:host_name"`
	FailCount int64  `json:"fail_count" gorm:"column:fail_count"`
}

func (HostBaselineFailTop5) TableName() string {
	return "v_views_host_baseline_fail_top5"
}

// HostVulnDailyStats 主机漏洞每日统计视图
type HostVulnDailyStats struct {
	TotalVulnCount     int64   `json:"total_vuln_count" gorm:"column:total_vuln_count"`         // 主机漏洞总数
	TodayNewCount      int64   `json:"today_new_count" gorm:"column:today_new_count"`           // 今日新增主机漏洞数
	TodayNewPercentage float64 `json:"today_new_percentage" gorm:"column:today_new_percentage"` // 今日新增占比(%)
}

func (HostVulnDailyStats) TableName() string {
	return "v_views_host_vuln_daily_stats"
}

// SecurityAlertDailyStats 安全告警每日统计视图
type SecurityAlertDailyStats struct {
	TotalAlertCount    int64   `json:"total_alert_count" gorm:"column:total_alert_count"`       // 安全告警总数
	TodayNewCount      int64   `json:"today_new_count" gorm:"column:today_new_count"`           // 今日新增安全告警数
	TodayNewPercentage float64 `json:"today_new_percentage" gorm:"column:today_new_percentage"` // 今日新增占比(%)
}

func (SecurityAlertDailyStats) TableName() string {
	return "v_views_security_alert_daily_stats"
}
