package view

import (
	"net/http"

	"github.com/haolipeng/BeeGuard/server/internal/models/view"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ViewHandler 概览视图处理器
type ViewHandler struct {
	DB *gorm.DB
}

// GetCodeQLVulnSummary 获取代码仓库漏洞统计
func (h *ViewHandler) GetCodeQLVulnSummary(c *gin.Context) {
	var stats []view.CodeQLVulnSummary
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询代码仓库漏洞统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetImageVulnTop5 获取容器镜像漏洞top5
func (h *ViewHandler) GetImageVulnTop5(c *gin.Context) {
	var stats []view.ImageVulnTop5ByCVE
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询容器镜像漏洞top5失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetAlertHourlyStats 获取每小时告警趋势
func (h *ViewHandler) GetAlertHourlyStats(c *gin.Context) {
	var stats []view.TotalAlertHourlyStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询每小时告警趋势失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetAlertMonthlyStats 获取每月告警数
func (h *ViewHandler) GetAlertMonthlyStats(c *gin.Context) {
	var stats []view.TotalAlertMonthlyStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询每月告警数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetImageVulnTop2 获取容器风险-漏洞视图top2
func (h *ViewHandler) GetImageVulnTop2(c *gin.Context) {
	var stats []view.VulnCountImageVuls
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询容器风险-漏洞视图top2失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostVulnTop2 获取主机风险-漏洞视图top2
func (h *ViewHandler) GetHostVulnTop2(c *gin.Context) {
	var stats []view.VulnCountVuls
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询主机风险-漏洞视图top2失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetBaselineItemTop5 获取合规基线检测项top5
func (h *ViewHandler) GetBaselineItemTop5(c *gin.Context) {
	var stats []view.BaselineItemTop5Affected
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询合规基线检测项top5失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostStatusSummary 获取在线主机统计
func (h *ViewHandler) GetHostStatusSummary(c *gin.Context) {
	var stats []view.HostStatusSummary
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询在线主机统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostVulnStats 获取主机风险资产统计
func (h *ViewHandler) GetHostVulnStats(c *gin.Context) {
	var stats []view.HostVulnStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询主机风险资产统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostVulnPackageTop5 获取风险资产分布TOP5
func (h *ViewHandler) GetHostVulnPackageTop5(c *gin.Context) {
	var stats []view.HostVulnPackageTop5
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询风险资产分布TOP5失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetThreatTypeTotalCount 获取威胁类型统计
func (h *ViewHandler) GetThreatTypeTotalCount(c *gin.Context) {
	var stats []view.ThreatTypeTotalCount
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询威胁类型统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetVulnChartData 获取安全看板漏洞统计
func (h *ViewHandler) GetVulnChartData(c *gin.Context) {
	var stats []view.VulnChartData
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询安全看板漏洞统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostBaselineFailTop5 获取基线检查失败主机top5
func (h *ViewHandler) GetHostBaselineFailTop5(c *gin.Context) {
	var stats []view.HostBaselineFailTop5
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询基线检查不通过主机top5失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostVulnDailyStats 获取主机漏洞每日统计
func (h *ViewHandler) GetHostVulnDailyStats(c *gin.Context) {
	var stats []view.HostVulnDailyStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询主机漏洞每日统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetSecurityAlertDailyStats 获取安全告警每日统计
func (h *ViewHandler) GetSecurityAlertDailyStats(c *gin.Context) {
	var stats []view.SecurityAlertDailyStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询安全告警每日统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}
