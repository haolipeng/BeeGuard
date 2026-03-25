package back

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/haolipeng/BeeGuard/server/internal/models/back"
	"gorm.io/gorm"
)

// BaselineTemplateHandler 基线模板处理器结构体
type BaselineTemplateHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateBaselineTemplate 创建基线模板
func (h *BaselineTemplateHandler) CreateBaselineTemplate(c *gin.Context) {
	// 接收数据
	var template back.BaselineTemplate
	// 验证并绑定请求中的JSON数据到template结构体
	if err := c.ShouldBindJSON(&template); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打印接收到的数据
	fmt.Printf("Received data: %+v\n", template)

	// 将baseline_ids数组转换为字符串格式存储
	baselineIDsStr := template.GetBaselineIDsAsString()

	// 创建一个新的模板实例用于数据库存储
	templateToSave := back.BaselineTemplate{
		//ID:           template.ID,
		TemplateName: template.TemplateName,
		TemplateType: template.TemplateType,
		OSType:       template.OSType,
		Version:      template.Version,
		ItemCount:    template.ItemCount,
		Description:  template.Description,
		IsEnabled:    template.IsEnabled,
		//CreatedAt:    template.CreatedAt,
		//UpdatedAt:    template.UpdatedAt,
	}

	// 如果有baseline_ids，存储转换后的字符串并计算数量
	if baselineIDsStr != nil {
		templateToSave.BaselineIDs = *baselineIDsStr
		// 计算baseline_ids的数量
		idsCount := len(strings.Split(*baselineIDsStr, ","))
		// 将数量存储到item_count字段
		idsCountInt32 := int32(idsCount)
		templateToSave.ItemCount = &idsCountInt32
	}

	//判断baseline_ids的值数量，
	// 执行数据库操作
	result := h.DB.Create(&templateToSave)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    templateToSave,
	})
}

// GetBaselineTemplate 获取单个基线模板
func (h *BaselineTemplateHandler) GetBaselineTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var template back.BaselineTemplate

	result := h.DB.Where("id = ?", id).First(&template)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "基线模板不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 将数据库中的字符串格式转换回数组格式
	if idsStr, ok := template.BaselineIDs.(string); ok && idsStr != "" {
		// 分割字符串
		idsArray := strings.Split(idsStr, ",")
		var ids []int64
		for _, idStr := range idsArray {
			if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
				ids = append(ids, id)
			}
		}
		template.BaselineIDs = ids
	}

	c.JSON(http.StatusOK, gin.H{"data": template})
}

// ListBaselineTemplates 获取基线模板列表（支持搜索查询）
func (h *BaselineTemplateHandler) ListBaselineTemplates(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	templateName := c.Query("template_name")
	templateType := c.Query("template_type")
	osType := c.Query("os_type")
	isEnabledStr := c.Query("is_enabled")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var templates []back.BaselineTemplate
	var total int64

	// 构建查询条件
	query := h.DB.Model(&back.BaselineTemplate{})

	// 添加搜索条件
	if templateName != "" {
		query = query.Where("template_name LIKE ?", "%"+templateName+"%")
	}
	if templateType != "" {
		query = query.Where("template_type = ?", templateType)
	}
	if osType != "" {
		query = query.Where("os_type = ?", osType)
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
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&templates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 将数据库中的字符串格式转换回数组格式
	for i := range templates {
		if idsStr, ok := templates[i].BaselineIDs.(string); ok && idsStr != "" {
			// 分割字符串
			idsArray := strings.Split(idsStr, ",")
			var ids []int64
			for _, idStr := range idsArray {
				if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
					ids = append(ids, id)
				}
			}
			templates[i].BaselineIDs = ids
		}
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": templates,
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

// UpdateBaselineTemplate 更新基线模板
func (h *BaselineTemplateHandler) UpdateBaselineTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var template back.BaselineTemplate
	// 检查基线模板是否存在
	result := h.DB.Where("id = ?", id).First(&template)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "基线模板不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 接收更新数据
	var updateData back.BaselineTemplate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 处理baseline_ids转换
	updateMap := make(map[string]interface{})

	// 只更新非零值字段
	if updateData.TemplateName != "" {
		updateMap["template_name"] = updateData.TemplateName
	}
	if updateData.TemplateType != "" {
		updateMap["template_type"] = updateData.TemplateType
	}
	if updateData.OSType != nil {
		updateMap["os_type"] = updateData.OSType
	}
	if updateData.Version != nil {
		updateMap["version"] = updateData.Version
	}
	if updateData.ItemCount != nil {
		updateMap["item_count"] = updateData.ItemCount
	}
	if updateData.Description != nil {
		updateMap["description"] = updateData.Description
	}
	if updateData.IsEnabled != template.IsEnabled {
		updateMap["is_enabled"] = updateData.IsEnabled
	}

	// 特别处理baseline_ids字段
	if updateData.BaselineIDs != nil {
		// 创建临时实例来使用转换方法
		tempTemplate := back.BaselineTemplate{BaselineIDs: updateData.BaselineIDs}
		baselineIDsStr := tempTemplate.GetBaselineIDsAsString()
		if baselineIDsStr != nil {
			updateMap["baseline_ids"] = *baselineIDsStr
			// 计算baseline_ids的数量并更新item_count字段
			idsCount := len(strings.Split(*baselineIDsStr, ","))
			updateMap["item_count"] = int32(idsCount)
		}
	}

	// 执行更新
	if len(updateMap) > 0 {
		result = h.DB.Model(&template).Updates(updateMap)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
			return
		}
	}

	// 查询更新后的数据
	var updatedTemplate back.BaselineTemplate
	h.DB.Where("id = ?", id).First(&updatedTemplate)

	// 将数据库中的字符串格式转换回数组格式
	if idsStr, ok := updatedTemplate.BaselineIDs.(string); ok && idsStr != "" {
		// 分割字符串
		idsArray := strings.Split(idsStr, ",")
		var ids []int64
		for _, idStr := range idsArray {
			if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
				ids = append(ids, id)
			}
		}
		updatedTemplate.BaselineIDs = ids
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updatedTemplate})
}

// DeleteBaselineTemplate 删除基线模板
func (h *BaselineTemplateHandler) DeleteBaselineTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&back.BaselineTemplate{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// BaselineTemplateHostLinkHandler 基线模板与主机关联处理器结构体
type BaselineTemplateHostLinkHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// DeleteBaselineTemplateHostLink 删除基线模板与主机关联
func (h *BaselineTemplateHostLinkHandler) DeleteBaselineTemplateHostLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	result := h.DB.Where("id = ?", id).Delete(&back.BaselineTemplateHostLink{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	return
}

// CreateBaselineTemplateHostLinkRequest 创建基线模板与主机关联请求结构体
type CreateBaselineTemplateHostLinkRequest struct {
	TemplateID    string  `json:"template_id" binding:"required"`    // 基线模板ID(字符串类型)
	TemplateName  string  `json:"template_name" binding:"required"`  // 基线模板名称
	TargetRange   *string `json:"target_range,omitempty"`            // 目标范围
	ScanFrequency string  `json:"scan_frequency" binding:"required"` // 扫描频率
}

// CreateBaselineTemplateHostLink 创建基线模板与主机关联
func (h *BaselineTemplateHostLinkHandler) CreateBaselineTemplateHostLink(c *gin.Context) {
	// 接收数据
	var req CreateBaselineTemplateHostLinkRequest
	// 验证并绑定请求中的JSON数据到req结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换 template_id 从字符串到 int64
	templateID, err := strconv.ParseInt(req.TemplateID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 template_id 格式"})
		return
	}

	// 创建数据库实体
	link := back.BaselineTemplateHostLink{
		TemplateID:    templateID,
		TemplateName:  req.TemplateName,
		TargetRange:   req.TargetRange,
		ScanFrequency: req.ScanFrequency,
	}

	// 执行数据库操作
	result := h.DB.Create(&link)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败：" + result.Error.Error()})
		return
	}
	//打印创建的记录，link.ID
	fmt.Println("创建的记录 ID:", link.ID)
	
	// 创建成功后，调用基线检查接口
	if link.TargetRange != nil && *link.TargetRange != "" {
		h.triggerBaselineCheck(link.ID, link.TemplateID, *link.TargetRange)
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    link,
	})
}

// triggerBaselineCheck 触发基线检查任务
func (h *BaselineTemplateHostLinkHandler) triggerBaselineCheck(linkID, templateID int64, targetRange string) {
	// 解析 TargetRange 获取 agent_ids
	var agentIDs []string
	if err := json.Unmarshal([]byte(targetRange), &agentIDs); err != nil {
		fmt.Printf("[基线检查] 解析 TargetRange 失败：%v\n", err)
		return
	}

	if len(agentIDs) == 0 {
		fmt.Println("[基线检查] TargetRange 为空，跳过触发")
		return
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"agent_ids":   agentIDs,
		"template_id": templateID,
		"baseline_id": fmt.Sprintf("task-%d", linkID),
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("[基线检查] 序列化请求体失败：%v\n", err)
		return
	}

	// 发送 POST 请求到基线检查接口
	apiURL := "http://127.0.0.1:8081/api/baseline/check"
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		fmt.Printf("[基线检查] 调用 API 失败：%v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		fmt.Printf("[基线检查] 解析响应失败：%v\n", err)
		return
	}

	fmt.Printf("[基线检查] 触发成功 - link_id=%d, template_id=%d, agents=%v, response=%v\n",
		linkID, templateID, agentIDs, respBody)
}

// ListBaselineTemplateHostLinks 获取基线模板与主机关联列表
func (h *BaselineTemplateHostLinkHandler) ListBaselineTemplateHostLinks(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	templateIDStr := c.Query("template_id")
	TemplateName := c.Query("template_name")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var links []back.BaselineTemplateHostLink
	var total int64

	// 构建查询条件
	query := h.DB.Model(&back.BaselineTemplateHostLink{})

	// 添加搜索条件
	if templateIDStr != "" {
		if templateID, err := strconv.ParseInt(templateIDStr, 10, 64); err == nil {
			query = query.Where("template_id = ?", templateID)
		}
	}
	if TemplateName != "" {
		query = query.Where("template_name LIKE ?", "%"+TemplateName+"%")
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&links)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 为每个 link 统计已检查的主机数
	type LinkWithStats struct {
		back.BaselineTemplateHostLink
		CheckedHostCount int `json:"checked_host_count"` // 已检查的主机数
		TotalHostCount   int `json:"total_host_count"`   // 总主机数
	}

	var linksWithStats []LinkWithStats
	for _, link := range links {
		item := LinkWithStats{
			BaselineTemplateHostLink: link,
		}

		// 解析 target_range 获取总主机数
		if link.TargetRange != nil && *link.TargetRange != "" {
			var agentIDs []string
			if err := json.Unmarshal([]byte(*link.TargetRange), &agentIDs); err == nil {
				item.TotalHostCount = len(agentIDs)
			}
		}

		// 查询 baseline_check_detail 表统计已检查的主机数
		// baseline_id 可能是 "task-{link.ID}" 格式，也可能是 "{link.ID}" 格式
		baselineIDWithPrefix := fmt.Sprintf("task-%d", link.ID)
		baselineIDWithoutPrefix := strconv.FormatInt(link.ID, 10)

		var checkedCount int64
		// 先尝试 task-{id} 格式
		h.DB.Table("baseline_check_detail").
			Where("baseline_id = ?", baselineIDWithPrefix).
			Distinct("agent_id").
			Count(&checkedCount)

		// 如果没找到，尝试纯数字格式
		if checkedCount == 0 {
			h.DB.Table("baseline_check_detail").
				Where("baseline_id = ?", baselineIDWithoutPrefix).
				Distinct("agent_id").
				Count(&checkedCount)
		}
		item.CheckedHostCount = int(checkedCount)

		linksWithStats = append(linksWithStats, item)
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": linksWithStats,
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
