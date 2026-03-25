package assets

import (
	"net/http"

	"github.com/haolipeng/BeeGuard/server/internal/models/assets"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ViewHandler 资产视图处理器
type ViewHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// GetOSTypeStats 获取系统类型统计
func (h *ViewHandler) GetOSTypeStats(c *gin.Context) {
	var stats []assets.AssetOSTypeStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询系统类型统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetHostStats 获取主机统计
func (h *ViewHandler) GetHostStats(c *gin.Context) {
	var stats []assets.AssetHostStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询主机统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetDatabaseTypeStats 获取数据库类型统计
func (h *ViewHandler) GetDatabaseTypeStats(c *gin.Context) {
	var stats []assets.AssetDatabaseTypeStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询数据库类型统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetDatabaseStats 获取数据库统计
func (h *ViewHandler) GetDatabaseStats(c *gin.Context) {
	var stats []assets.AssetDatabaseStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询数据库统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetContainerStats 获取容器统计
func (h *ViewHandler) GetContainerStats(c *gin.Context) {
	var stats []assets.AssetContainerStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询容器统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetAccountStats 获取账号统计
func (h *ViewHandler) GetAccountStats(c *gin.Context) {
	var stats []assets.AssetAccountStats
	result := h.DB.Find(&stats)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询账号统计失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": stats})
}

// GetLatestAssetsTop5 获取近期更新资产
func (h *ViewHandler) GetLatestAssetsTop5(c *gin.Context) {
	var assets []assets.AssetLatestAssetsTop5
	result := h.DB.Find(&assets)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询近期更新资产失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": assets})
}
