package host

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WebHandler Web服务资产处理器结构体，用于处理与Web服务资产相关的HTTP请求
type WebHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListWebs 获取Web服务资产列表（支持搜索查询）
func (h *WebHandler) ListWebs(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	appName := c.Query("name")
	version := c.Query("version")
	serverType := c.Query("server_type")
	siteDomain := c.Query("site_domain")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var webs []host.Web
	var total int64

	// 构建查询条件
	query := h.DB.Model(&host.Web{})

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
	if appName != "" {
		query = query.Where("name LIKE ?", "%"+appName+"%")
	}
	if version != "" {
		query = query.Where("version LIKE ?", "%"+version+"%")
	}
	if serverType != "" {
		query = query.Where("server_type LIKE ?", "%"+serverType+"%")
	}
	if siteDomain != "" {
		query = query.Where("site_domain LIKE ?", "%"+siteDomain+"%")
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&webs)
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
		"data": webs,
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
