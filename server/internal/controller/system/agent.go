package system

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AgentInfoHandler Agent客户端信息处理器结构体
type AgentInfoHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateAgentInfo 创建Agent客户端信息
func (h *AgentInfoHandler) CreateAgentInfo(c *gin.Context) {
	// 接收数据
	var agent system.AgentInfo
	// 验证并绑定请求中的JSON数据到agent结构体
	if err := c.ShouldBindJSON(&agent); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行数据库操作
	result := h.DB.Create(&agent)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    agent,
	})
}

// GetAgentInfo 获取单个Agent客户端信息
func (h *AgentInfoHandler) GetAgentInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var agent system.AgentInfo

	result := h.DB.Where("id = ?", id).First(&agent)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": agent})
}

// ListAgentInfos 获取Agent客户端信息列表（支持搜索查询）
func (h *AgentInfoHandler) ListAgentInfos(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	osType := c.Query("os_type")
	connectionStatusStr := c.Query("connection_status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var agents []system.AgentInfo
	var total int64

	// 构建查询条件
	query := h.DB.Model(&system.AgentInfo{})

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
	if osType != "" {
		query = query.Where("os_type = ?", osType)
	}
	if connectionStatusStr != "" {
		if status, err := strconv.Atoi(connectionStatusStr); err == nil {
			query = query.Where("connection_status = ?", status)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按最后连接时间倒序排列
	result = query.Order("last_connected_at DESC NULLS LAST, created_at DESC").Limit(limit).Offset(offset).Find(&agents)
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
		"data": agents,
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

// UpdateAgentInfo 更新Agent客户端信息
func (h *AgentInfoHandler) UpdateAgentInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var agent system.AgentInfo
	// 检查Agent信息是否存在
	result := h.DB.Where("id = ?", id).First(&agent)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 接收更新数据
	var updateData system.AgentInfo
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行更新
	result = h.DB.Model(&agent).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的数据
	var updatedAgent system.AgentInfo
	h.DB.Where("id = ?", id).First(&updatedAgent)

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updatedAgent})
}

// DeleteAgentInfo 删除Agent客户端信息
func (h *AgentInfoHandler) DeleteAgentInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&system.AgentInfo{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// UpdateAgentConnectionStatus 更新Agent连接状态
func (h *AgentInfoHandler) UpdateAgentConnectionStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		ConnectionStatus int16 `json:"connection_status" binding:"oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.DB.Model(&system.AgentInfo{}).Where("id = ?", id).Updates(map[string]interface{}{
		"connection_status": req.ConnectionStatus,
		"last_connected_at":  gorm.Expr("CURRENT_TIMESTAMP"),
	})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent信息不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "连接状态更新成功"})
}
