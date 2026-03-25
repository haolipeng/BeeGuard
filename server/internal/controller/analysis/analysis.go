package analysis

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	analysispkg "github.com/haolipeng/BeeGuard/server/internal/analysis"
)

// Controller 分析控制器
type Controller struct{}

// NewController 创建控制器
func NewController() *Controller {
	return &Controller{}
}

// TriggerAnalysis 手动触发分析
// POST /api/analysis/trigger
func (c *Controller) TriggerAnalysis(ctx *gin.Context) {
	if err := analysispkg.TriggerAnalysis(ctx.Request.Context()); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "分析任务已触发"})
}

// AnalyzeHost 分析指定主机
// POST /api/analysis/host
func (c *Controller) AnalyzeHost(ctx *gin.Context) {
	var req struct {
		HostIP string `json:"host_ip" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少 host_ip 参数"})
		return
	}

	report, err := analysispkg.AnalyzeByHost(ctx.Request.Context(), req.HostIP)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if report == nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "该主机无待分析告警"})
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// AnalyzeSourceIP 分析指定攻击源
// POST /api/analysis/source
func (c *Controller) AnalyzeSourceIP(ctx *gin.Context) {
	var req struct {
		SourceIP string `json:"source_ip" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少 source_ip 参数"})
		return
	}

	report, err := analysispkg.AnalyzeBySourceIP(ctx.Request.Context(), req.SourceIP)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if report == nil {
		ctx.JSON(http.StatusOK, gin.H{"message": "该攻击源无待分析告警"})
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// GetReports 获取分析报告列表
// GET /api/analysis/reports
func (c *Controller) GetReports(ctx *gin.Context) {
	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	reports, err := engine.GetReports()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"reports": reports, "count": len(reports)})
}

// GetReport 获取单个报告详情
// GET /api/analysis/reports/:filename
func (c *Controller) GetReport(ctx *gin.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件名"})
		return
	}

	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	report, err := engine.LoadReport(filename)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "报告不存在"})
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// GetStats 获取分析模块统计
// GET /api/analysis/stats
func (c *Controller) GetStats(ctx *gin.Context) {
	stats := analysispkg.Stats()
	if stats == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// ListReportsFromDB 从数据库获取报告列表
// GET /api1/analysis/db_reports
func (c *Controller) ListReportsFromDB(ctx *gin.Context) {
	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 解析过滤参数
	riskLevel := ctx.Query("risk_level")
	analysisType := ctx.Query("analysis_type")

	reports, total, err := engine.GetReportsFromDB(ctx.Request.Context(), page, pageSize, riskLevel, analysisType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"reports":   reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetReportFromDB 从数据库获取单个报告
// GET /api1/analysis/db_reports/:id
func (c *Controller) GetReportFromDB(ctx *gin.Context) {
	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	report, err := engine.GetReportFromDB(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "报告不存在"})
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// DeleteReportFromDB 删除数据库中的报告
// DELETE /api1/analysis/db_reports/:id
func (c *Controller) DeleteReportFromDB(ctx *gin.Context) {
	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的报告ID"})
		return
	}

	if err := engine.DeleteReportFromDB(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "报告已删除"})
}

// GetDBReportStats 获取数据库报告统计
// GET /api1/analysis/db_reports/stats
func (c *Controller) GetDBReportStats(ctx *gin.Context) {
	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	stats, err := engine.GetDBReportStats(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// AnalyzeAlerts 手动分析告警（直接传入告警数据）
// POST /api1/analysis/alerts
func (c *Controller) AnalyzeAlerts(ctx *gin.Context) {
	engine := analysispkg.GetEngine()
	if engine == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "分析模块未启用"})
		return
	}

	var req struct {
		Alerts []analysispkg.AlertContext `json:"alerts" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少 alerts 参数"})
		return
	}

	if len(req.Alerts) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "告警数据不能为空"})
		return
	}

	result, err := engine.AnalyzeAlerts(ctx.Request.Context(), req.Alerts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
