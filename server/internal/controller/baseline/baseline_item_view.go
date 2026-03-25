package baseline

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/baseline"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BaselineCheckItemViewHandler 基线检查结果项统计视图处理器结构体
type BaselineCheckItemViewHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListBaselineCheckItemViews 获取基线检查结果项统计列表（基于视图）
func (h *BaselineCheckItemViewHandler) ListBaselineCheckItemViews(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	templateIDStr := c.Query("template_id")
	baselineID := c.Query("baseline_id")
	itemIDStr := c.Query("item_id")
	baselineName := c.Query("baseline_name")
	minTotalHostsStr := c.Query("min_total_hosts")
	minPassedChecksStr := c.Query("min_passed_checks")
	minFailedChecksStr := c.Query("min_failed_checks")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var itemViews []baseline.BaselineCheckItemView
	var total int64

	// 构建查询条件
	query := h.DB.Model(&baseline.BaselineCheckItemView{})

	// 添加搜索条件
	if templateIDStr != "" {
		if templateID, err := strconv.ParseInt(templateIDStr, 10, 64); err == nil {
			query = query.Where("template_id = ?", templateID)
		}
	}
	if baselineName != "" {
		query = query.Where("baseline_name LIKE ?", "%"+baselineName+"%")
	}
	if minTotalHostsStr != "" {
		if minTotalHosts, err := strconv.Atoi(minTotalHostsStr); err == nil {
			query = query.Where("total_hosts >= ?", minTotalHosts)
		}
	}
	if minPassedChecksStr != "" {
		if minPassedChecks, err := strconv.Atoi(minPassedChecksStr); err == nil {
			query = query.Where("passed_checks >= ?", minPassedChecks)
		}
	}
	if minFailedChecksStr != "" {
		if minFailedChecks, err := strconv.Atoi(minFailedChecksStr); err == nil {
			query = query.Where("failed_checks >= ?", minFailedChecks)
		}
	}
	if baselineID != "" {
		query = query.Where("baseline_id = ?", baselineID)
	}
	if itemIDStr != "" {
		if itemID, err := strconv.ParseInt(itemIDStr, 10, 64); err == nil {
			query = query.Where("item_id = ?", itemID)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按模板ID排序
	result = query.Order("template_id ASC").Limit(limit).Offset(offset).Find(&itemViews)
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
		"data": itemViews,
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

// GetBaselineCheckItemView 获取单个基线检查结果项统计详情
func (h *BaselineCheckItemViewHandler) GetBaselineCheckItemView(c *gin.Context) {
	templateIDStr := c.Param("template_id")
	if templateIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID"})
		return
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模板ID格式"})
		return
	}

	var itemView baseline.BaselineCheckItemView
	result := h.DB.Where("template_id = ?", templateID).First(&itemView)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "基线项统计信息不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": itemView})
}
