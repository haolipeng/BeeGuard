package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReverseShellHandler 反弹shell告警处理器结构体，用于处理与反弹shell告警相关的HTTP请求
type ReverseShellHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListShellAlerts 获取反弹shell告警列表（支持搜索查询）
func (h *ReverseShellHandler) ListShellAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	victimIP := c.Query("victim_ip")
	commandLine := c.Query("command_line")
	shellType := c.Query("shell_type")
	targetHost := c.Query("target_host")
	targetPort := c.Query("target_port")
	statusStr := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var shellAlerts []alert.ReverseShell
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.ReverseShell{})

	// 添加搜索条件
	if agentID != "" {
		query = query.Where("agent_id LIKE ?", "%"+agentID+"%")
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if victimIP != "" {
		query = query.Where("victim_ip LIKE ?", "%"+victimIP+"%")
	}
	if commandLine != "" {
		query = query.Where("command_line LIKE ?", "%"+commandLine+"%")
	}
	if shellType != "" {
		query = query.Where("shell_type LIKE ?", "%"+shellType+"%")
	}
	if targetHost != "" {
		query = query.Where("target_host LIKE ?", "%"+targetHost+"%")
	}
	if targetPort != "" {
		if port, err := strconv.Atoi(targetPort); err == nil {
			query = query.Where("target_port = ?", port)
		}
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按事件时间倒序排列
	result = query.Order("event_time DESC").Limit(limit).Offset(offset).Find(&shellAlerts)
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
		"data": shellAlerts,
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

// GetShellAlertByID 根据ID获取反弹shell告警详情
func (h *ReverseShellHandler) GetShellAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var shellAlert alert.ReverseShell
	result := h.DB.Where("id = ?", id).First(&shellAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": shellAlert})
}

// UpdateShellAlertStatus 更新反弹shell告警状态
func (h *ReverseShellHandler) UpdateShellAlertStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		Status int16 `json:"status" binding:"oneof=0 1 2"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.DB.Model(&alert.ReverseShell{}).Where("id = ?", id).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}