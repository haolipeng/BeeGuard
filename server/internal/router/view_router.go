package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/view"
	"github.com/haolipeng/BeeGuard/server/internal/db"

	"github.com/gin-gonic/gin"
)

// SetupViewRouter 设置概览视图相关路由
func SetupViewRouter(r *gin.RouterGroup) {
	viewHandler := &view.ViewHandler{DB: db.GetDB()}

	// 在线主机统计
	r.GET("/host-status-summary", viewHandler.GetHostStatusSummary)

	// 主机风险资产统计
	r.GET("/host-vuln-stats", viewHandler.GetHostVulnStats)

	// 代码仓库漏洞统计
	r.GET("/codeql-vuln-summary", viewHandler.GetCodeQLVulnSummary)

	// 容器镜像漏洞top5
	r.GET("/image-vuln-top5", viewHandler.GetImageVulnTop5)

	// 每小时告警趋势
	r.GET("/alert-hourly-stats", viewHandler.GetAlertHourlyStats)

	// 每月告警数
	r.GET("/alert-monthly-stats", viewHandler.GetAlertMonthlyStats)

	// 容器风险-漏洞视图top2
	r.GET("/image-vuln-top2", viewHandler.GetImageVulnTop2)

	// 主机风险-漏洞视图top2
	r.GET("/host-vuln-top2", viewHandler.GetHostVulnTop2)

	// 合规基线检测项top5
	r.GET("/baseline-item-top5", viewHandler.GetBaselineItemTop5)

	// 风险资产分布TOP5
	r.GET("/host-vuln-package-top5", viewHandler.GetHostVulnPackageTop5)

	// 威胁类型统计
	r.GET("/threat-type-total-count", viewHandler.GetThreatTypeTotalCount)

	// 安全看板漏洞统计
	r.GET("/vuln-chart-data", viewHandler.GetVulnChartData)

	// 基线检查不通过主机top5
	r.GET("/host-baseline-fail-top5", viewHandler.GetHostBaselineFailTop5)

	// 主机漏洞每日统计
	r.GET("/host-vuln-daily-stats", viewHandler.GetHostVulnDailyStats)

	// 安全告警每日统计
	r.GET("/security-alert-daily-stats", viewHandler.GetSecurityAlertDailyStats)
}
