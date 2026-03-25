package alert

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/alert"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FileHandler 文件查杀告警处理器结构体，用于处理与文件查杀告警相关的HTTP请求
type FileHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListFileAlerts 获取文件查杀告警列表（支持搜索查询）
func (h *FileHandler) ListFileAlerts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	threatType := c.Query("threat_type")
	fileName := c.Query("file_name")
	filePath := c.Query("file_path")
	fileMD5 := c.Query("file_md5")
	malwareFamily := c.Query("malware_family")
	statusStr := c.Query("status")
	isQuarantinedStr := c.Query("is_quarantined")
	isDeletedStr := c.Query("is_deleted")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var fileAlerts []alert.MalwareScan
	var total int64

	// 构建查询条件
	query := h.DB.Model(&alert.MalwareScan{})

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
	if threatType != "" {
		query = query.Where("threat_type = ?", threatType)
	}
	if fileName != "" {
		query = query.Where("file_name LIKE ?", "%"+fileName+"%")
	}
	if filePath != "" {
		query = query.Where("file_path LIKE ?", "%"+filePath+"%")
	}
	if fileMD5 != "" {
		query = query.Where("file_md5 = ?", fileMD5)
	}
	if malwareFamily != "" {
		query = query.Where("malware_family LIKE ?", "%"+malwareFamily+"%")
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}
	if isQuarantinedStr != "" {
		if isQuarantined, err := strconv.Atoi(isQuarantinedStr); err == nil {
			query = query.Where("is_quarantined = ?", isQuarantined)
		}
	}
	if isDeletedStr != "" {
		if isDeleted, err := strconv.Atoi(isDeletedStr); err == nil {
			query = query.Where("is_deleted = ?", isDeleted)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按扫描时间倒序排列
	result = query.Order("scan_time DESC").Limit(limit).Offset(offset).Find(&fileAlerts)
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
		"data": fileAlerts,
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

// GetFileAlertByID 根据ID获取文件查杀告警详情
func (h *FileHandler) GetFileAlertByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var fileAlert alert.MalwareScan
	result := h.DB.Where("id = ?", id).First(&fileAlert)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": fileAlert})
}

// UpdateFileAlertStatus 更新文件查杀告警状态
func (h *FileHandler) UpdateFileAlertStatus(c *gin.Context) {
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

	result := h.DB.Model(&alert.MalwareScan{}).Where("id = ?", id).Update("status", req.Status)
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

// UpdateFileAlertQuarantine 更新文件查杀告警隔离状态
func (h *FileHandler) UpdateFileAlertQuarantine(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		IsQuarantined int16 `json:"is_quarantined" binding:"required,oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.DB.Model(&alert.MalwareScan{}).Where("id = ?", id).Update("is_quarantined", req.IsQuarantined)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "隔离状态更新成功"})
}

// UpdateFileAlertDeletion 更新文件查杀告警删除状态
func (h *FileHandler) UpdateFileAlertDeletion(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		IsDeleted int16 `json:"is_deleted" binding:"required,oneof=0 1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := h.DB.Model(&alert.MalwareScan{}).Where("id = ?", id).Update("is_deleted", req.IsDeleted)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "告警记录不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除状态更新成功"})
}