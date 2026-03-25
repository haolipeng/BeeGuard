package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FileGuardHandler 文件完整性告警处理器结构体，用于处理与文件完整性告警相关的HTTP请求
type FileGuardHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListFileGuardAlerts 获取文件完整性告警列表（支持搜索查询）
func (h *FileGuardHandler) ListFileGuardAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	ruleType := c.Query("rule_type")
	ruleName := c.Query("rule_name")
	ruleIDStr := c.Query("rule_id")
	threatLevel := c.Query("threat_level")
	threatAction := c.Query("threat_action")
	filePath := c.Query("file_path")
	fileName := c.Query("file_name")
	operatorUser := c.Query("operator_user")
	statusStr := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var fileGuardAlerts []alert.FileIntegrity
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.FileIntegrity{})

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
	if ruleType != "" {
		query = query.Where("rule_type = ?", ruleType)
	}
	if ruleName != "" {
		query = query.Where("rule_name LIKE ?", "%"+ruleName+"%")
	}
	if ruleIDStr != "" {
		if ruleID, err := strconv.ParseInt(ruleIDStr, 10, 64); err == nil {
			query = query.Where("rule_id = ?", ruleID)
		}
	}
	if threatLevel != "" {
		query = query.Where("threat_level = ?", threatLevel)
	}
	if threatAction != "" {
		query = query.Where("threat_action = ?", threatAction)
	}
	if filePath != "" {
		query = query.Where("file_path LIKE ?", "%"+filePath+"%")
	}
	if fileName != "" {
		query = query.Where("file_name LIKE ?", "%"+fileName+"%")
	}
	if operatorUser != "" {
		query = query.Where("operator_user LIKE ?", "%"+operatorUser+"%")
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

	// 分页查询数据，按告警时间倒序排列
	result = query.Order("alert_time DESC").Limit(limit).Offset(offset).Find(&fileGuardAlerts)
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
		"data": fileGuardAlerts,
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

// GetFileGuardAlertByID 根据ID获取文件完整性告警详情
func (h *FileGuardHandler) GetFileGuardAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var fileGuardAlert alert.FileIntegrity
	result := h.DB.Where("id = ?", id).First(&fileGuardAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fileGuardAlert})
}

// UpdateFileGuardAlertStatus 更新文件完整性告警状态
func (h *FileGuardHandler) UpdateFileGuardAlertStatus(c *gin.Context) {
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

	result := h.DB.Model(&alert.FileIntegrity{}).Where("id = ?", id).Update("status", req.Status)
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