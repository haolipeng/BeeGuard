package host

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HostHandler 主机资产处理器结构体，用于处理与主机资产相关的HTTP请求
type HostHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListHosts 获取主机资产列表（支持搜索查询）
func (h *HostHandler) ListHosts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	agentStatus := c.Query("agent_status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var hosts []host.Host
	var total int64

	// 构建查询条件，使用 SELECT 显式指定字段并从 agent_info 表关联查询 connection_status
	query := h.DB.Model(&host.Host{}).Select("asset_host.*, agent_info.connection_status AS agent_status")

	// 使用 JOIN 关联 agent_info 表
	query = query.Joins("LEFT JOIN agent_info ON asset_host.agent_id = agent_info.agent_id")

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
	if agentStatus != "" {
		status, err := strconv.Atoi(agentStatus)
		if err == nil {
			query = query.Where("agent_info.connection_status = ?", status)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("asset_host.created_at DESC").Limit(limit).Offset(offset).Find(&hosts)
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
		"data": hosts,
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
