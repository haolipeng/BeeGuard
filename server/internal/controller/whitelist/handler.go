package whitelist

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	wlModel "github.com/haolipeng/BeeGuard/server/internal/models/whitelist"
	"github.com/haolipeng/BeeGuard/server/internal/whitelist"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 白名单管理控制器
type Handler struct {
	DB        *gorm.DB
	Repo      *repository.WhitelistRepository
	WlChecker *whitelist.Checker
}

// CreateRule 创建白名单规则
func (h *Handler) CreateRule(c *gin.Context) {
	alertType := c.Param("alert_type")
	if !wlModel.IsValidAlertType(alertType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警类型: " + alertType})
		return
	}

	var rule wlModel.WhitelistRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.Create(c.Request.Context(), alertType, &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建规则失败"})
		return
	}

	// 触发异步追溯匹配
	h.WlChecker.InvalidateCache(alertType)
	h.WlChecker.RetroactiveCheck(alertType, &rule)

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// ListRules 查询白名单规则列表
func (h *Handler) ListRules(c *gin.Context) {
	alertType := c.Param("alert_type")
	if !wlModel.IsValidAlertType(alertType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警类型: " + alertType})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}

	rules, total, err := h.Repo.List(c.Request.Context(), alertType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询规则列表失败"})
		return
	}

	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

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

// GetRule 获取单条规则详情
func (h *Handler) GetRule(c *gin.Context) {
	alertType := c.Param("alert_type")
	if !wlModel.IsValidAlertType(alertType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警类型: " + alertType})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	rule, err := h.Repo.GetByID(c.Request.Context(), alertType, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rule})
}

// UpdateRule 更新白名单规则
func (h *Handler) UpdateRule(c *gin.Context) {
	alertType := c.Param("alert_type")
	if !wlModel.IsValidAlertType(alertType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警类型: " + alertType})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 如果 conditions 是 map，序列化为 JSONB
	if cond, ok := updates["conditions"]; ok {
		condBytes, err := json.Marshal(cond)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conditions 格式错误"})
			return
		}
		var conditions wlModel.Conditions
		if err := json.Unmarshal(condBytes, &conditions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "conditions 格式错误"})
			return
		}
		updates["conditions"] = conditions
	}

	if err := h.Repo.Update(c.Request.Context(), alertType, id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新规则失败"})
		return
	}

	// 清除缓存，追溯匹配
	h.WlChecker.InvalidateCache(alertType)
	rule, _ := h.Repo.GetByID(c.Request.Context(), alertType, id)
	if rule != nil {
		h.WlChecker.RetroactiveCheck(alertType, rule)
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则更新成功"})
}

// DeleteRule 删除白名单规则
func (h *Handler) DeleteRule(c *gin.Context) {
	alertType := c.Param("alert_type")
	if !wlModel.IsValidAlertType(alertType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警类型: " + alertType})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.Repo.Delete(c.Request.Context(), alertType, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除规则失败"})
		return
	}

	// 清除缓存，恢复被命中的告警
	h.WlChecker.InvalidateCache(alertType)
	h.WlChecker.RestoreOnDelete(alertType, id)

	c.JSON(http.StatusOK, gin.H{"message": "规则删除成功"})
}

// ToggleRule 启用/禁用规则
func (h *Handler) ToggleRule(c *gin.Context) {
	alertType := c.Param("alert_type")
	if !wlModel.IsValidAlertType(alertType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的告警类型: " + alertType})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	rule, err := h.Repo.GetByID(c.Request.Context(), alertType, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "规则不存在"})
		return
	}

	newEnabled := !rule.Enabled
	if err := h.Repo.Update(c.Request.Context(), alertType, id, map[string]interface{}{"enabled": newEnabled}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "切换规则状态失败"})
		return
	}

	h.WlChecker.InvalidateCache(alertType)

	status := "已禁用"
	if newEnabled {
		status = "已启用"
	}
	c.JSON(http.StatusOK, gin.H{"message": "规则" + status, "enabled": newEnabled})
}

// GetAlertTypes 获取所有支持的告警类型列表
func (h *Handler) GetAlertTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": wlModel.ValidAlertTypes})
}
