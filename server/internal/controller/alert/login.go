package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginHandler 异常登录告警处理器结构体，用于处理与异常登录告警相关的HTTP请求
type LoginHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListLoginAlerts 获取异常登录告警列表（支持搜索查询）
func (h *LoginHandler) ListLoginAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	sourceIP := c.Query("source_ip")
	sourceLocation := c.Query("source_location")
	loginUser := c.Query("login_user")
	riskLevel := c.Query("risk_level")
	abnormalType := c.Query("abnormal_type")
	statusStr := c.Query("status")
	isWhitelistStr := c.Query("is_whitelist")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var loginAlerts []alert.AbnormalLogin
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.AbnormalLogin{})

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
	if loginUser != "" {
		query = query.Where("login_user LIKE ?", "%"+loginUser+"%")
	}
	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	if abnormalType != "" {
		query = query.Where("abnormal_type LIKE ?", "%"+abnormalType+"%")
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}
	if isWhitelistStr != "" {
		if isWhitelist, err := strconv.Atoi(isWhitelistStr); err == nil {
			query = query.Where("is_whitelist = ?", isWhitelist)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按登录时间倒序排列
	result = query.Order("login_time DESC").Limit(limit).Offset(offset).Find(&loginAlerts)
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
		"data": loginAlerts,
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

// GetLoginAlertByID 根据ID获取异常登录告警详情
func (h *LoginHandler) GetLoginAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var loginAlert alert.AbnormalLogin
	result := h.DB.Where("id = ?", id).First(&loginAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": loginAlert})
}

// UpdateLoginAlertStatus 更新异常登录告警状态
func (h *LoginHandler) UpdateLoginAlertStatus(c *gin.Context) {
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

	result := h.DB.Model(&alert.AbnormalLogin{}).Where("id = ?", id).Update("status", req.Status)
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

// UpdateLoginAlertWhitelist 更新异常登录告警白名单状态
func (h *LoginHandler) UpdateLoginAlertWhitelist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		IsWhitelist int16 `json:"is_whitelist" binding:"required,oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.DB.Model(&alert.AbnormalLogin{}).Where("id = ?", id).Update("is_whitelist", req.IsWhitelist)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "白名单状态更新成功"})
}