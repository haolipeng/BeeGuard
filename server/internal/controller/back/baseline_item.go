package back

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/back"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BaselineCheckItemHandler 基线检查项处理器结构体
type BaselineCheckItemHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateBaselineCheckItem 创建基线检查项
func (h *BaselineCheckItemHandler) CreateBaselineCheckItem(c *gin.Context) {
	// 接收数据
	var item back.BaselineCheckItem
	// 验证并绑定请求中的JSON数据到item结构体
	if err := c.ShouldBindJSON(&item); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行数据库操作
	result := h.DB.Create(&item)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    item,
	})
}

// GetBaselineCheckItem 获取单个基线检查项
func (h *BaselineCheckItemHandler) GetBaselineCheckItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var item back.BaselineCheckItem

	result := h.DB.Where("id = ?", id).First(&item)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "检查项不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

// ListBaselineCheckItems 获取基线检查项列表（支持搜索查询）
func (h *BaselineCheckItemHandler) ListBaselineCheckItems(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	templateIDStr := c.Query("template_id")
	itemName := c.Query("item_name")
	category := c.Query("category")
	riskLevel := c.Query("risk_level")
	isEnabledStr := c.Query("is_enabled")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var items []back.BaselineCheckItem
	var total int64

	// 构建查询条件
	query := h.DB.Model(&back.BaselineCheckItem{})

	// 添加搜索条件
	if templateIDStr != "" {
		if templateID, err := strconv.ParseInt(templateIDStr, 10, 64); err == nil {
			query = query.Where("template_id = ?", templateID)
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
	if isEnabledStr != "" {
		if isEnabled, err := strconv.Atoi(isEnabledStr); err == nil {
			query = query.Where("is_enabled = ?", isEnabled)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items)
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
		"data": items,
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

// UpdateBaselineCheckItem 更新基线检查项
func (h *BaselineCheckItemHandler) UpdateBaselineCheckItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var item back.BaselineCheckItem
	// 检查检查项是否存在
	result := h.DB.Where("id = ?", id).First(&item)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "检查项不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 接收更新数据
	var updateData back.BaselineCheckItem
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行更新
	result = h.DB.Model(&item).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的数据
	var updatedItem back.BaselineCheckItem
	h.DB.Where("id = ?", id).First(&updatedItem)

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updatedItem})
}

// DeleteBaselineCheckItem 删除基线检查项
func (h *BaselineCheckItemHandler) DeleteBaselineCheckItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&back.BaselineCheckItem{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}