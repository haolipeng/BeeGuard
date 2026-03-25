package back

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/back"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCodeRule 创建规则
type CodeRuleHandler struct {
	DB *gorm.DB
}

// CreateCodeRule 创建规则
type CreateCodeRuleRequest struct {
	Enabled                     *bool   `json:"enabled"`                       // 是否启用
	ID                          *string `json:"id"`                            // 规则id
	Code                        *string `json:"code"`                          // 编程语言
	ShortDescriptionText        *string `json:"short_description_text"`        // 简短描述
	FullDescriptionText         *string `json:"full_description_text"`         // 完整描述
	DefaultConfigurationEnabled *string `json:"default_configuration_enabled"` // 默认配置启用
	DefaultConfigurationLevel   *string `json:"default_configuration_level"`   // 默认配置级别
	PropertiesTags              *string `json:"properties_tags"`               // 标签
	PropertiesDescription       *string `json:"properties_description"`        // 描述
	PropertiesKind              *string `json:"properties_kind"`               // 问题类型
	PropertiesPrecision         *string `json:"properties_precision"`          // 精度
	PropertiesProblemSeverity   *string `json:"properties_problem_severity"`   // 问题严重程度
	PropertiesSecuritySeverity  *string `json:"properties_security_severity"`  // 安全级别
}

func (h *CodeRuleHandler) CreateCodeRule(c *gin.Context) {
	var req CreateCodeRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建规则实体
	rule := back.CodeqlRule{
		Enabled:                     req.Enabled,
		ID:                          req.ID,
		Code:                        req.Code,
		ShortDescriptionText:        req.ShortDescriptionText,
		FullDescriptionText:         req.FullDescriptionText,
		DefaultConfigurationEnabled: req.DefaultConfigurationEnabled,
		DefaultConfigurationLevel:   req.DefaultConfigurationLevel,
		PropertiesTags:              req.PropertiesTags,
		PropertiesDescription:       req.PropertiesDescription,
		PropertiesKind:              req.PropertiesKind,
		PropertiesPrecision:         req.PropertiesPrecision,
		PropertiesProblemSeverity:   req.PropertiesProblemSeverity,
		PropertiesSecuritySeverity:  req.PropertiesSecuritySeverity,
	}

	result := h.DB.Create(&rule)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建规则失败: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则创建成功", "data": rule})
}

// GetCodeRule 获取规则详情
type GetCodeRuleResponse struct {
	RuleID                      int64   `json:"rule_id"`
	Enabled                     *bool   `json:"enabled"`
	ID                          *string `json:"id"`
	Code                        *string `json:"code"`
	ShortDescriptionText        *string `json:"short_description_text"`
	FullDescriptionText         *string `json:"full_description_text"`
	DefaultConfigurationEnabled *string `json:"default_configuration_enabled"`
	DefaultConfigurationLevel   *string `json:"default_configuration_level"`
	PropertiesTags              *string `json:"properties_tags"`
	PropertiesDescription       *string `json:"properties_description"`
	PropertiesKind              *string `json:"properties_kind"`
	PropertiesPrecision         *string `json:"properties_precision"`
	PropertiesProblemSeverity   *string `json:"properties_problem_severity"`
	PropertiesSecuritySeverity  *string `json:"properties_security_severity"`
	CreateTime                  string  `json:"create_time"`
	UpdateTime                  string  `json:"update_time"`
}

// GetCodeRule 获取规则详情
func (h *CodeRuleHandler) GetCodeRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var rule back.CodeqlRule
	result := h.DB.Where("rule_id = ? AND (deleted IS NULL OR deleted = 0)", id).First(&rule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	response := GetCodeRuleResponse{
		RuleID:                      rule.RuleID,
		Enabled:                     rule.Enabled,
		ID:                          rule.ID,
		Code:                        rule.Code,
		ShortDescriptionText:        rule.ShortDescriptionText,
		FullDescriptionText:         rule.FullDescriptionText,
		DefaultConfigurationEnabled: rule.DefaultConfigurationEnabled,
		DefaultConfigurationLevel:   rule.DefaultConfigurationLevel,
		PropertiesTags:              rule.PropertiesTags,
		PropertiesDescription:       rule.PropertiesDescription,
		PropertiesKind:              rule.PropertiesKind,
		PropertiesPrecision:         rule.PropertiesPrecision,
		PropertiesProblemSeverity:   rule.PropertiesProblemSeverity,
		PropertiesSecuritySeverity:  rule.PropertiesSecuritySeverity,
		CreateTime:                  rule.CreateTime.Time.Format("2006-01-02 15:04:05"),
		UpdateTime:                  rule.UpdateTime.Time.Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateCodeRule 更新规则
type UpdateCodeRuleRequest struct {
	Enabled                     *bool   `json:"enabled"`                       // 是否启用
	Code                        *string `json:"code"`                          // 编程语言
	ShortDescriptionText        *string `json:"short_description_text"`        // 简短描述
	FullDescriptionText         *string `json:"full_description_text"`         // 完整描述
	DefaultConfigurationEnabled *string `json:"default_configuration_enabled"` // 默认配置启用
	DefaultConfigurationLevel   *string `json:"default_configuration_level"`   // 默认配置级别
	PropertiesTags              *string `json:"properties_tags"`               // 标签
	PropertiesDescription       *string `json:"properties_description"`        // 描述
	PropertiesKind              *string `json:"properties_kind"`               // 问题类型
	PropertiesPrecision         *string `json:"properties_precision"`          // 精度
	PropertiesProblemSeverity   *string `json:"properties_problem_severity"`   // 问题严重程度
	PropertiesSecuritySeverity  *string `json:"properties_security_severity"`  // 安全级别
}

// UpdateCodeRule 更新规则
func (h *CodeRuleHandler) UpdateCodeRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req UpdateCodeRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查规则是否存在
	var existingRule back.CodeqlRule
	result := h.DB.Where("rule_id = ? AND (deleted IS NULL OR deleted = 0)", id).First(&existingRule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 构建更新映射
	updateFields := make(map[string]interface{})

	if req.Enabled != nil {
		updateFields["enabled"] = *req.Enabled
	}
	if req.Code != nil {
		updateFields["code"] = *req.Code
	}
	if req.ShortDescriptionText != nil {
		updateFields["shortdescription_text"] = *req.ShortDescriptionText
	}
	if req.FullDescriptionText != nil {
		updateFields["fulldescription_text"] = *req.FullDescriptionText
	}
	if req.DefaultConfigurationEnabled != nil {
		updateFields["defaultconfiguration_enabled"] = *req.DefaultConfigurationEnabled
	}
	if req.DefaultConfigurationLevel != nil {
		updateFields["defaultconfiguration_level"] = *req.DefaultConfigurationLevel
	}
	if req.PropertiesTags != nil {
		updateFields["properties_tags"] = *req.PropertiesTags
	}
	if req.PropertiesDescription != nil {
		updateFields["properties_description"] = *req.PropertiesDescription
	}
	if req.PropertiesKind != nil {
		updateFields["properties_kind"] = *req.PropertiesKind
	}
	if req.PropertiesPrecision != nil {
		updateFields["properties_precision"] = *req.PropertiesPrecision
	}
	if req.PropertiesProblemSeverity != nil {
		updateFields["properties_problem_severity"] = *req.PropertiesProblemSeverity
	}
	if req.PropertiesSecuritySeverity != nil {
		updateFields["properties_security_severity"] = *req.PropertiesSecuritySeverity
	}

	// 如果没有要更新的字段
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供要更新的字段"})
		return
	}

	// 执行更新
	result = h.DB.Model(&back.CodeqlRule{}).Where("rule_id = ?", id).Updates(updateFields)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + result.Error.Error()})
		return
	}

	// 查询更新后的数据
	var updatedRule back.CodeqlRule
	h.DB.Where("rule_id = ?", id).First(&updatedRule)

	c.JSON(http.StatusOK, gin.H{"message": "规则更新成功", "data": updatedRule})
}

// DeleteCodeRule 删除规则(逻辑删除)
func (h *CodeRuleHandler) DeleteCodeRule(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	deleted := int8(1)
	result := h.DB.Model(&back.CodeqlRule{}).Where("rule_id = ?", id).Update("deleted", deleted)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则删除成功"})
}

// ListCodeRules 获取规则列表
type ListCodeRulesResponse struct {
	Data  []GetCodeRuleResponse `json:"data"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// ListCodeRules 获取规则列表（支持搜索查询）
func (h *CodeRuleHandler) ListCodeRules(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	idFilter := c.Query("id")
	codeFilter := c.Query("code")
	enabledStr := c.Query("enabled")
	shortDescriptionTextFilter := c.Query("short_description_text")
	//id := c.Query("id")
	propertiesPrecisionFilter := c.Query("properties_precision")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var rules []back.CodeqlRule
	var total int64

	// 构建查询条件
	query := h.DB.Model(&back.CodeqlRule{}).Where("deleted IS NULL OR deleted = 0")

	// 添加搜索条件
	if idFilter != "" {
		query = query.Where("id LIKE ?", "%"+idFilter+"%")
	}
	if codeFilter != "" {
		query = query.Where("code LIKE ?", "%"+codeFilter+"%")
	}
	if enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			query = query.Where("enabled = ?", enabled)
		}
	}
	if shortDescriptionTextFilter != "" {
		query = query.Where("short_description_text LIKE ?", "%"+shortDescriptionTextFilter+"%")
	}
	if propertiesPrecisionFilter != "" {
		query = query.Where("properties_precision LIKE ?", "%"+propertiesPrecisionFilter+"%")
	}
	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败: " + result.Error.Error()})
		return
	}

	// 分页查询数据，按rule_id倒序排列
	result = query.Order("rule_id DESC").Limit(limit).Offset(offset).Find(&rules)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 转换为响应格式
	responseData := make([]GetCodeRuleResponse, len(rules))
	for i, rule := range rules {
		responseData[i] = GetCodeRuleResponse{
			RuleID:                      rule.RuleID,
			Enabled:                     rule.Enabled,
			ID:                          rule.ID,
			Code:                        rule.Code,
			ShortDescriptionText:        rule.ShortDescriptionText,
			FullDescriptionText:         rule.FullDescriptionText,
			DefaultConfigurationEnabled: rule.DefaultConfigurationEnabled,
			DefaultConfigurationLevel:   rule.DefaultConfigurationLevel,
			PropertiesTags:              rule.PropertiesTags,
			PropertiesDescription:       rule.PropertiesDescription,
			PropertiesKind:              rule.PropertiesKind,
			PropertiesPrecision:         rule.PropertiesPrecision,
			PropertiesProblemSeverity:   rule.PropertiesProblemSeverity,
			PropertiesSecuritySeverity:  rule.PropertiesSecuritySeverity,
			CreateTime:                  rule.CreateTime.Time.Format("2006-01-02 15:04:05"),
			UpdateTime:                  rule.UpdateTime.Time.Format("2006-01-02 15:04:05"),
		}
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": responseData,
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

// GetSimpleRuleList 获取简化规则列表(用于前端选择框)
type SimpleRuleItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// GetSimpleRuleList 获取简化规则列表(用于前端选择框)
func (h *CodeRuleHandler) GetSimpleRuleList(c *gin.Context) {
	var rules []back.CodeqlRule

	// 查询所有启用且未删除的规则
	result := h.DB.Where("enabled = ? AND (deleted IS NULL OR deleted = 0)", true).Find(&rules)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + result.Error.Error()})
		return
	}

	// 转换为前端需要的格式
	response := make([]SimpleRuleItem, len(rules))
	for i, rule := range rules {
		label := "未知规则"
		if rule.ShortDescriptionText != nil && *rule.ShortDescriptionText != "" {
			label = *rule.ShortDescriptionText
		} else if rule.ID != nil && *rule.ID != "" {
			label = *rule.ID
		}
		response[i] = SimpleRuleItem{
			Label: label,
			Value: rule.RuleID,
		}
	}

	c.JSON(http.StatusOK, response)
}
