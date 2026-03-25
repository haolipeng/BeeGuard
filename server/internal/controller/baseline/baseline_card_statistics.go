package baseline

import (
	"net/http"

	"github.com/haolipeng/BeeGuard/server/internal/models/baseline"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BaselineCheckHostCardStatisticsHandler 基线检查主机卡片统计视图处理器结构体
type BaselineCheckHostCardStatisticsHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListBaselineCheckHostCardStatistics 获取基线检查主机卡片统计列表 (基于视图，支持按 baseline_id 筛选)
func (h *BaselineCheckHostCardStatisticsHandler) ListBaselineCheckHostCardStatistics(c *gin.Context) {
	var cardStats []baseline.BaselineCheckHostCardStatistics

	// 构建查询
	query := h.DB.Model(&baseline.BaselineCheckHostCardStatistics{})

	// 获取可选的 baseline_id 搜索条件
	baselineID := c.Query("baseline_id")
	if baselineID != "" {
		query = query.Where("baseline_id = ?", baselineID)
	}

	// 执行查询
	result := query.Find(&cardStats)
	if result.Error != nil {
		// 返回详细错误信息用于调试
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "detail": result.Error.Error()})
		return
	}

	// 返回查询结果
	c.JSON(http.StatusOK, gin.H{
		"data": cardStats,
	})
}

// GetBaselineCheckHostCardStatistic 获取单个基线检查主机卡片统计详情
func (h *BaselineCheckHostCardStatisticsHandler) GetBaselineCheckHostCardStatistic(c *gin.Context) {
	baselineID := c.Param("baseline_id")
	if baselineID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的基线 ID"})
		return
	}

	var cardStat baseline.BaselineCheckHostCardStatistics
	result := h.DB.Where("baseline_id = ?", baselineID).First(&cardStat)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "卡片统计信息不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cardStat})
}
