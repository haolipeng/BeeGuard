package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/whitelist"
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	wlEngine "github.com/haolipeng/BeeGuard/server/internal/whitelist"

	"github.com/gin-gonic/gin"
)

// SetupWhitelistRouter 配置白名单管理相关路由
func SetupWhitelistRouter(r *gin.RouterGroup, checker *wlEngine.Checker) {
	handler := &whitelist.Handler{
		DB:        db.GetDB(),
		Repo:      repository.NewWhitelistRepository(db.GetDB()),
		WlChecker: checker,
	}

	// 获取支持的告警类型列表
	r.GET("/types", handler.GetAlertTypes)

	// 按告警类型的白名单规则 CRUD
	r.POST("/:alert_type", handler.CreateRule)
	r.GET("/:alert_type", handler.ListRules)
	r.GET("/:alert_type/:id", handler.GetRule)
	r.PUT("/:alert_type/:id", handler.UpdateRule)
	r.DELETE("/:alert_type/:id", handler.DeleteRule)
	r.POST("/:alert_type/:id/toggle", handler.ToggleRule)
}
