package baseline

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/baseline"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BaselineCheckHostViewHandler 基线检查结果主机统计视图处理器结构体
type BaselineCheckHostViewHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListBaselineCheckHostViews 获取基线检查结果主机统计列表（基于视图）
func (h *BaselineCheckHostViewHandler) ListBaselineCheckHostViews(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	baselineID := c.Query("baseline_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	minTotalChecksStr := c.Query("min_total_checks")
	minFailedChecksStr := c.Query("min_failed_checks")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var hostViews []baseline.BaselineCheckHostView
	var total int64

	// 构建查询条件
	query := h.DB.Model(&baseline.BaselineCheckHostView{})

	// 添加搜索条件
	if agentID != "" {
		query = query.Where("agent_id LIKE ?", "%"+agentID+"%")
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
	}
	if minTotalChecksStr != "" {
		if minTotalChecks, err := strconv.Atoi(minTotalChecksStr); err == nil {
			query = query.Where("total_checks >= ?", minTotalChecks)
		}
	}
	if minFailedChecksStr != "" {
		if minFailedChecks, err := strconv.Atoi(minFailedChecksStr); err == nil {
			query = query.Where("failed_checks >= ?", minFailedChecks)
		}
	}
	if baselineID != "" {
		query = query.Where("baseline_id = ?", baselineID)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列 (移除可能不存在的last_check_time字段)
	//result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&hostViews)
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&hostViews)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": hostViews,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  total,
			"per_page":     limit,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
	})
}

// GetBaselineCheckHostView 获取单个基线检查结果主机统计详情
func (h *BaselineCheckHostViewHandler) GetBaselineCheckHostView(c *gin.Context) {
	agentID := c.Param("agent_id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的Agent ID"})
		return
	}

	var hostView baseline.BaselineCheckHostView
	result := h.DB.Where("agent_id = ?", agentID).First(&hostView)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "主机统计信息不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hostView})
}
