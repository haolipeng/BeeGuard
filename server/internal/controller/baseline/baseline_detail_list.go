package baseline

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/baseline"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BaselineCheckDetailHandler 基线检查结果明细处理器结构体
type BaselineCheckDetailHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListBaselineCheckDetails 获取基线检查结果明细列表（支持搜索查询）
func (h *BaselineCheckDetailHandler) ListBaselineCheckDetails(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	resultIDStr := c.Query("result_id")
	baselineIDStr := c.Query("baseline_id")
	hostIDStr := c.Query("host_id")
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	itemIDStr := c.Query("item_id")
	itemName := c.Query("item_name")
	category := c.Query("category")
	riskLevel := c.Query("risk_level")
	statusStr := c.Query("status")
	templateIDStr := c.Query("template_id")
	baselineName := c.Query("baseline_name")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var details []baseline.BaselineCheckDetail
	var total int64

	// 构建查询条件
	query := h.DB.Model(&baseline.BaselineCheckDetail{})

	// 添加搜索条件
	if resultIDStr != "" {
		if resultID, err := strconv.ParseInt(resultIDStr, 10, 64); err == nil {
			query = query.Where("result_id = ?", resultID)
		}
	}
	if baselineIDStr != "" {
		query = query.Where("baseline_id = ?", baselineIDStr)
	}
	if hostIDStr != "" {
		if hostID, err := strconv.ParseInt(hostIDStr, 10, 64); err == nil {
			query = query.Where("host_id = ?", hostID)
		}
	}
	if agentID != "" {
		//不要模糊匹配，要严格匹配
		query = query.Where("agent_id = ?", agentID)
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
	}
	// 根据检查项ID进行精确匹配查询
	if itemIDStr != "" {
		if itemID, err := strconv.ParseInt(itemIDStr, 10, 64); err == nil {
			query = query.Where("item_id = ?", itemID)
		}
	}
	if itemName != "" {
		query = query.Where("item_name LIKE ?", "%"+itemName+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}
	if templateIDStr != "" {
		if templateID, err := strconv.ParseInt(templateIDStr, 10, 64); err == nil {
			query = query.Where("template_id = ?", templateID)
		}
	}
	if baselineName != "" {
		query = query.Where("baseline_name LIKE ?", "%"+baselineName+"%")
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按检查时间倒序排列
	result = query.Order("check_time DESC").Limit(limit).Offset(offset).Find(&details)
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
		"data": details,
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

// GetBaselineCheckDetail 获取单个基线检查结果明细
func (h *BaselineCheckDetailHandler) GetBaselineCheckDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var detail baseline.BaselineCheckDetail
	result := h.DB.Where("id = ?", id).First(&detail)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "检查结果明细不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": detail})
}

// UpdateBaselineCheckDetailStatus 更新基线检查结果明细状态
func (h *BaselineCheckDetailHandler) UpdateBaselineCheckDetailStatus(c *gin.Context) {
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

	result := h.DB.Model(&baseline.BaselineCheckDetail{}).Where("id = ?", id).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "检查结果明细不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}
