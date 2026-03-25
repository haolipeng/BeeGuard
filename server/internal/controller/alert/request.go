package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequestHandler 恶意请求告警处理器结构体，用于处理与恶意请求告警相关的HTTP请求
type RequestHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListRequestAlerts 获取恶意请求告警列表（支持搜索查询）
func (h *RequestHandler) ListRequestAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	policyType := c.Query("policy_type")
	policyName := c.Query("policy_name")
	maliciousDomain := c.Query("malicious_domain")
	maliciousIP := c.Query("malicious_ip")
	requestCountStr := c.Query("request_count")
	statusStr := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var requestAlerts []alert.MaliciousRequest
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.MaliciousRequest{})

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
	if policyType != "" {
		query = query.Where("policy_type = ?", policyType)
	}
	if policyName != "" {
		query = query.Where("policy_name LIKE ?", "%"+policyName+"%")
	}
	if maliciousDomain != "" {
		query = query.Where("malicious_domain LIKE ?", "%"+maliciousDomain+"%")
	}
	if maliciousIP != "" {
		query = query.Where("malicious_ip LIKE ?", "%"+maliciousIP+"%")
	}
	if requestCountStr != "" {
		if requestCount, err := strconv.Atoi(requestCountStr); err == nil {
			query = query.Where("request_count >= ?", requestCount)
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

	// 分页查询数据，按最后请求时间倒序排列
	result = query.Order("last_request_time DESC NULLS LAST, created_at DESC").Limit(limit).Offset(offset).Find(&requestAlerts)
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
		"data": requestAlerts,
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

// GetRequestAlertByID 根据ID获取恶意请求告警详情
func (h *RequestHandler) GetRequestAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var requestAlert alert.MaliciousRequest
	result := h.DB.Where("id = ?", id).First(&requestAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requestAlert})
}

// UpdateRequestAlertStatus 更新恶意请求告警状态
func (h *RequestHandler) UpdateRequestAlertStatus(c *gin.Context) {
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

	result := h.DB.Model(&alert.MaliciousRequest{}).Where("id = ?", id).Update("status", req.Status)
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