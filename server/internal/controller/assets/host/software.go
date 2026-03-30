package host

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SoftwareHandler 软件资产处理器结构体
type SoftwareHandler struct {
	DB *gorm.DB
}

// ListSoftware 获取软件资产列表（支持搜索查询）
func (h *SoftwareHandler) ListSoftware(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	name := c.Query("name")
	softwareType := c.Query("type")

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	var software []host.Software
	var total int64

	query := h.DB.Model(&host.Software{})

	if agentID != "" {
		query = query.Where("agent_id LIKE ?", "%"+agentID+"%")
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if softwareType != "" {
		query = query.Where("type = ?", softwareType)
	}

	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&software)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": software,
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
