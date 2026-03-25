package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PasswdHandler 暴力破解告警处理器结构体，用于处理与暴力破解告警相关的HTTP请求
type PasswdHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListPasswdAlerts 获取暴力破解告警列表（支持搜索查询）
func (h *PasswdHandler) ListPasswdAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	sourceIP := c.Query("source_ip")
	sourceLocation := c.Query("source_location")
	attackType := c.Query("attack_type")
	targetIP := c.Query("target_ip")
	targetPortStr := c.Query("target_port")
	username := c.Query("username")
	attemptCountStr := c.Query("attempt_count")
	statusStr := c.Query("status")
	isBlockedStr := c.Query("is_blocked")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var passwdAlerts []alert.BruteForce
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.BruteForce{})

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
	if sourceIP != "" {
		query = query.Where("source_ip LIKE ?", "%"+sourceIP+"%")
	}
	if sourceLocation != "" {
		query = query.Where("source_location LIKE ?", "%"+sourceLocation+"%")
	}
	if attackType != "" {
		query = query.Where("attack_type = ?", attackType)
	}
	if targetIP != "" {
		query = query.Where("target_ip LIKE ?", "%"+targetIP+"%")
	}
	if targetPortStr != "" {
		if targetPort, err := strconv.Atoi(targetPortStr); err == nil {
			query = query.Where("target_port = ?", targetPort)
		}
	}
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if attemptCountStr != "" {
		if attemptCount, err := strconv.Atoi(attemptCountStr); err == nil {
			query = query.Where("attempt_count >= ?", attemptCount)
		}
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}
	if isBlockedStr != "" {
		if isBlocked, err := strconv.Atoi(isBlockedStr); err == nil {
			query = query.Where("is_blocked = ?", isBlocked)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按攻击时间倒序排列
	result = query.Order("attack_time DESC").Limit(limit).Offset(offset).Find(&passwdAlerts)
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
		"data": passwdAlerts,
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

// GetPasswdAlertByID 根据ID获取暴力破解告警详情
func (h *PasswdHandler) GetPasswdAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var passwdAlert alert.BruteForce
	result := h.DB.Where("id = ?", id).First(&passwdAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": passwdAlert})
}

// UpdatePasswdAlertStatus 更新暴力破解告警状态
func (h *PasswdHandler) UpdatePasswdAlertStatus(c *gin.Context) {
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

	result := h.DB.Model(&alert.BruteForce{}).Where("id = ?", id).Update("status", req.Status)
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

// UpdatePasswdAlertBlockStatus 更新暴力破解告警封禁状态
func (h *PasswdHandler) UpdatePasswdAlertBlockStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		IsBlocked int16 `json:"is_blocked" binding:"required,oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.DB.Model(&alert.BruteForce{}).Where("id = ?", id).Update("is_blocked", req.IsBlocked)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "封禁状态更新成功"})
}