package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NetworkHandler 网络攻击告警处理器结构体，用于处理与网络攻击告警相关的HTTP请求
type NetworkHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListNetworkAlerts 获取网络攻击告警列表（支持搜索查询）
func (h *NetworkHandler) ListNetworkAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	targetPortStr := c.Query("target_port")
	attackerIP := c.Query("attacker_ip")
	attackerLocation := c.Query("attacker_location")
	attackerCountry := c.Query("attacker_country")
	vulnerabilityName := c.Query("vulnerability_name")
	vulnerabilityID := c.Query("vulnerability_id")
	attackStatus := c.Query("attack_status")
	attackCountStr := c.Query("attack_count")
	statusStr := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var networkAlerts []alert.NetworkAttack
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.NetworkAttack{})

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
	if targetPortStr != "" {
		if targetPort, err := strconv.Atoi(targetPortStr); err == nil {
			query = query.Where("target_port = ?", targetPort)
		}
	}
	if attackerIP != "" {
		query = query.Where("attacker_ip LIKE ?", "%"+attackerIP+"%")
	}
	if attackerLocation != "" {
		query = query.Where("attacker_location LIKE ?", "%"+attackerLocation+"%")
	}
	if attackerCountry != "" {
		query = query.Where("attacker_country LIKE ?", "%"+attackerCountry+"%")
	}
	if vulnerabilityName != "" {
		query = query.Where("vulnerability_name LIKE ?", "%"+vulnerabilityName+"%")
	}
	if vulnerabilityID != "" {
		query = query.Where("vulnerability_id LIKE ?", "%"+vulnerabilityID+"%")
	}
	if attackStatus != "" {
		query = query.Where("attack_status = ?", attackStatus)
	}
	if attackCountStr != "" {
		if attackCount, err := strconv.Atoi(attackCountStr); err == nil {
			query = query.Where("attack_count >= ?", attackCount)
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

	// 分页查询数据，按最后攻击时间倒序排列
	result = query.Order("last_attack_time DESC").Limit(limit).Offset(offset).Find(&networkAlerts)
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
		"data": networkAlerts,
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

// GetNetworkAlertByID 根据ID获取网络攻击告警详情
func (h *NetworkHandler) GetNetworkAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var networkAlert alert.NetworkAttack
	result := h.DB.Where("id = ?", id).First(&networkAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": networkAlert})
}

// UpdateNetworkAlertStatus 更新网络攻击告警状态
func (h *NetworkHandler) UpdateNetworkAlertStatus(c *gin.Context) {
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

	result := h.DB.Model(&alert.NetworkAttack{}).Where("id = ?", id).Update("status", req.Status)
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