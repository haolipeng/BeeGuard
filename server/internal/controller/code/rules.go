package code

import (
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/models/code"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateRule 创建规则集
func CreateRule(c *gin.Context) {
	var rule code.Rules
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认状态
	if rule.Status == "" {
		rule.Status = "ENABLED"
	}

	result := db.GetDB().Create(&rule)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建成功", "data": rule})
}

// GetRuleByID 根据ID获取规则集
func GetRuleByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var rule code.Rules
	result := db.GetDB().Where("rules_id = ? AND (deleted IS NULL OR deleted = 0)", id).First(&rule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则集不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// UpdateRule 更新规则集
func UpdateRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	// 定义可更新的字段结构体
	var updateData struct {
		RuleName        *string `json:"rule_name"`        // 规则集名称
		RuleIDs         *string `json:"rule_ids"`         // 关联codeq_rule表的rule_id
		Description     *string `json:"description"`      // 规则集描述
		ApplicableScene *string `json:"applicable_scene"` // 适用场景
		RiskCoverage    *string `json:"risk_coverage"`    // 风险覆盖范围
		Status          *string `json:"status"`           // 启用状态
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查规则集是否存在
	var existingRule code.Rules
	result := db.GetDB().Where("rules_id = ? AND (deleted IS NULL OR deleted = 0)", id).First(&existingRule)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "规则集不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 构建更新映射
	updateFields := make(map[string]interface{})

	if updateData.RuleName != nil {
		updateFields["rule_name"] = *updateData.RuleName
	}

	if updateData.RuleIDs != nil {
		updateFields["rule_ids"] = *updateData.RuleIDs
	}
	if updateData.Description != nil {
		updateFields["description"] = *updateData.Description
	}
	if updateData.ApplicableScene != nil {
		updateFields["applicable_scene"] = *updateData.ApplicableScene
	}
	if updateData.RiskCoverage != nil {
		updateFields["risk_coverage"] = *updateData.RiskCoverage
	}
	if updateData.Status != nil {
		// 验证状态值
		if *updateData.Status != "ENABLED" && *updateData.Status != "DISABLED" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "状态值必须为 ENABLED 或 DISABLED"})
			return
		}
		updateFields["status"] = *updateData.Status
	}

	// 如果没有要更新的字段
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供要更新的字段"})
		return
	}

	// 执行更新
	result = db.GetDB().Model(&code.Rules{}).Where("rules_id = ?", id).Updates(updateFields)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的数据
	var updatedRule code.Rules
	db.GetDB().Where("rules_id = ?", id).First(&updatedRule)

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updatedRule})
}

// DeleteRule 删除规则集(逻辑删除)
func DeleteRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	deleted := int8(1)
	result := db.GetDB().Model(&code.Rules{}).Where("rules_id = ?", id).Update("deleted", deleted)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则集不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetAllRuleList 获取所有规则集(简化格式用于前端选择框)
func GetAllRuleList(c *gin.Context) {
	var rules []code.Rules

	// 查询所有未删除的规则集
	result := db.GetDB().Where("deleted IS NULL OR deleted = 0").Find(&rules)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 转换为前端需要的格式
	response := make([]map[string]interface{}, len(rules))
	for i, rule := range rules {
		response[i] = map[string]interface{}{
			"label": rule.RuleName,
			"value": rule.RuleName, // 可以根据需要调整为其他字段
		}
	}

	c.JSON(http.StatusOK, response)
}
func GetAllRules(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	ruleName := c.Query("rule_name")
	status := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var rules []code.Rules
	var total int64

	// 构建查询条件
	query := db.GetDB().Model(&code.Rules{}).Where("deleted IS NULL OR deleted = 0")

	// 添加搜索条件
	if ruleName != "" {
		query = query.Where("rule_name LIKE ?", "%"+ruleName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据
	result = query.Order("rules_id DESC").Limit(limit).Offset(offset).Find(&rules)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  rules,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
