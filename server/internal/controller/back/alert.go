package back

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/back"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HIDSRuleHandler 入侵检测告警规则处理器结构体
type HIDSRuleHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateHIDSRule 创建入侵检测告警规则
func (h *HIDSRuleHandler) CreateHIDSRule(c *gin.Context) {
	// 接收数据
	var rule back.HIDSRule
	// 验证并绑定请求中的JSON数据到rule结构体
	if err := c.ShouldBindJSON(&rule); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 先检查规则是否已存在，避免触发数据库唯一约束错误日志
	var existingRule back.HIDSRule
	checkResult := h.DB.Where("rule_name = ?", rule.RuleName).First(&existingRule)
	if checkResult.Error == nil {
		// 规则已存在，返回提示信息
		c.JSON(http.StatusOK, gin.H{
			"message": "数据已存在，跳过创建",
			"data":    existingRule,
		})
		return
	} else if checkResult.Error != gorm.ErrRecordNotFound {
		// 其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败: " + checkResult.Error.Error()})
		return
	}

	// 规则不存在，执行插入操作
	result := h.DB.Create(&rule)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + result.Error.Error()})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    rule,
	})
}

// GetHIDSRule 获取单个入侵检测告警规则
func (h *HIDSRuleHandler) GetHIDSRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var rule back.HIDSRule

	result := h.DB.Where("id = ?", id).First(&rule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// ListHIDSRules 获取入侵检测告警规则列表（支持搜索查询）
func (h *HIDSRuleHandler) ListHIDSRules(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	ruleName := c.Query("rule_name")
	ruleLevel := c.Query("rule_level")
	threatType := c.Query("threat_type")
	ruleStatus := c.Query("rule_status")
	rulerType := c.Query("ruler_type")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var rules []back.HIDSRule
	var total int64

	// 构建查询条件
	query := h.DB.Model(&back.HIDSRule{})

	// 添加搜索条件
	if ruleName != "" {
		query = query.Where("rule_name LIKE ?", "%"+ruleName+"%")
	}
	if ruleLevel != "" {
		query = query.Where("rule_level = ?", ruleLevel)
	}
	if threatType != "" {
		query = query.Where("threat_type = ?", threatType)
	}
	if ruleStatus != "" {
		query = query.Where("rule_status = ?", ruleStatus)
	}
	if rulerType != "" {
		query = query.Where("ruler_type = ?", rulerType)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rules)
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
		"data": rules,
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

// UpdateHIDSRule 更新入侵检测告警规则
func (h *HIDSRuleHandler) UpdateHIDSRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var rule back.HIDSRule
	// 检查规则是否存在
	result := h.DB.Where("id = ?", id).First(&rule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 接收更新数据
	var updateData back.HIDSRule
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行更新
	result = h.DB.Model(&rule).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的数据
	var updatedRule back.HIDSRule
	h.DB.Where("id = ?", id).First(&updatedRule)

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updatedRule})
}

// DeleteHIDSRule 删除入侵检测告警规则
func (h *HIDSRuleHandler) DeleteHIDSRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&back.HIDSRule{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
