package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/back"
	"github.com/haolipeng/BeeGuard/server/internal/controller/baseline"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"

	"github.com/gin-gonic/gin"
)

// SetupBaselineRouter 设置基线路由
func SetupBaselineRouter(router *gin.RouterGroup) {
	// 基线基模板管理相关路由
	baselineTemplateGroup := router.Group("/templates")
	{
		baselineTemplateHandler := &back.BaselineTemplateHandler{DB: mysql.DB}
		// 基线基模板CURD接口
		baselineTemplateGroup.POST("/create", baselineTemplateHandler.CreateBaselineTemplate)     // 创建基线模板
		baselineTemplateGroup.GET("/:id", baselineTemplateHandler.GetBaselineTemplate)            // 获取单个基线模板
		baselineTemplateGroup.GET("", baselineTemplateHandler.ListBaselineTemplates)              // 获取基线模板列表
		baselineTemplateGroup.POST("/edit/:id", baselineTemplateHandler.UpdateBaselineTemplate)   // 更新基线模板
		baselineTemplateGroup.POST("/delete/:id", baselineTemplateHandler.DeleteBaselineTemplate) // 删除基线模板
	}

	// 基线基模板与主机关联管理相关路由
	baselineLinkGroup := router.Group("/links")
	{
		baselineLinkHandler := &back.BaselineTemplateHostLinkHandler{DB: mysql.DB}
		// 基线基模板与主机关联CURD接口
		baselineLinkGroup.POST("/create", baselineLinkHandler.CreateBaselineTemplateHostLink) // 创建关联
		baselineLinkGroup.GET("", baselineLinkHandler.ListBaselineTemplateHostLinks)          // 获取关联列表
		baselineLinkGroup.POST("/delete/:id", baselineLinkHandler.DeleteBaselineTemplateHostLink)
	}

	// 基线检查项管理相关路由
	baselineItemGroup := router.Group("/items")
	{
		baselineItemHandler := &back.BaselineCheckItemHandler{DB: mysql.DB}
		// 基线检查项CURD接口
		baselineItemGroup.POST("/create", baselineItemHandler.CreateBaselineCheckItem)     // 创建检查项
		baselineItemGroup.GET("/:id", baselineItemHandler.GetBaselineCheckItem)            // 获取单个检查项
		baselineItemGroup.GET("", baselineItemHandler.ListBaselineCheckItems)              // 获取检查项列表
		baselineItemGroup.POST("/edit/:id", baselineItemHandler.UpdateBaselineCheckItem)   // 更新检查项
		baselineItemGroup.POST("/delete/:id", baselineItemHandler.DeleteBaselineCheckItem) // 删除检查项
	}

	// 基线检查结果明细管理相关路由
	baselineDetailGroup := router.Group("/details")
	{
		baselineDetailHandler := &baseline.BaselineCheckDetailHandler{DB: mysql.DB}
		// 基线检查结果明细接口
		baselineDetailGroup.GET("", baselineDetailHandler.ListBaselineCheckDetails)                    // 获取检查结果明细列表
		baselineDetailGroup.GET("/:id", baselineDetailHandler.GetBaselineCheckDetail)                  // 获取单个检查结果明细
		baselineDetailGroup.POST("/status/:id", baselineDetailHandler.UpdateBaselineCheckDetailStatus) // 更新检查结果状态
	}

	// 基线检查主机统计视图相关路由
	baselineHostViewGroup := router.Group("/host_views")
	{
		baselineHostViewHandler := &baseline.BaselineCheckHostViewHandler{DB: mysql.DB}
		// 基线检查主机统计接口（基于视图）
		baselineHostViewGroup.GET("host", baselineHostViewHandler.ListBaselineCheckHostViews)     // 获取主机统计列表
		baselineHostViewGroup.GET("/:agent_id", baselineHostViewHandler.GetBaselineCheckHostView) // 获取单个主机统计详情

	}

	// 基线检查项统计视图相关路由
	baselineItemViewGroup := router.Group("/item_views")
	{
		baselineItemViewHandler := &baseline.BaselineCheckItemViewHandler{DB: mysql.DB}
		// 基线检查检查项统计接口（基于视图）
		baselineItemViewGroup.GET("item", baselineItemViewHandler.ListBaselineCheckItemViews)
		baselineItemViewGroup.GET("/:template_id", baselineItemViewHandler.GetBaselineCheckItemView) // 获取单个检查项统计详情
	}

	// 基线检查主机卡片统计视图相关路由
	baselineCardStatsGroup := router.Group("/card_statistics")
	{
		baselineCardStatsHandler := &baseline.BaselineCheckHostCardStatisticsHandler{DB: mysql.DB}
		// 基线检查主机卡片统计接口（基于视图，不分页）
		baselineCardStatsGroup.GET("", baselineCardStatsHandler.ListBaselineCheckHostCardStatistics)                // 获取卡片统计列表
		baselineCardStatsGroup.GET("/:baseline_id", baselineCardStatsHandler.GetBaselineCheckHostCardStatistic)     // 获取单个卡片统计详情
	}
}
