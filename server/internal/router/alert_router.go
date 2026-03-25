package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/alert"
	"github.com/haolipeng/BeeGuard/server/internal/db"

	"github.com/gin-gonic/gin"
)

// SetupAlertRouter 设置告警管理相关路由
func SetupAlertRouter(r *gin.RouterGroup) {
	// 高危命令告警相关路由
	commandGroup := r.Group("/command")
	{
		commandHandler := &alert.CommandHandler{DB: db.GetDB()}
		// 高危命令告警CURD接口
		commandGroup.GET("/commands", commandHandler.ListCommands)                    // 获取高危命令告警列表(支持分页和搜索)
		commandGroup.POST("/commands/status/:id", commandHandler.UpdateCommandStatus) // 更新高危命令告警状态
	}

	// 反弹shell告警相关路由
	shellGroup := r.Group("/shell")
	{
		shellHandler := &alert.ReverseShellHandler{DB: db.GetDB()}
		// 反弹shell告警CURD接口
		shellGroup.GET("/shells", shellHandler.ListShellAlerts) // 获取反弹shell告警列表(支持分页和搜索)
		//shellGroup.GET("/alerts/:id", shellHandler.GetShellAlertByID)             // 根据ID获取反弹shell告警详情
		shellGroup.POST("/shells/status/:id", shellHandler.UpdateShellAlertStatus) // 更新反弹shell告警状态
	}

	// 本地提权告警相关路由
	localGroup := r.Group("/local")
	{
		localHandler := &alert.ShellHandler{DB: db.GetDB()}
		// 本地提权告警CURD接口
		localGroup.GET("/alerts", localHandler.ListShellAlerts) // 获取本地提权告警列表(支持分页和搜索)
		//localGroup.GET("/alerts/:id", localHandler.GetShellAlertByID)              // 根据ID获取本地提权告警详情
		localGroup.POST("/alerts/status/:id", localHandler.UpdateShellAlertStatus) // 更新本地提权告警状态
	}

	// 异常登录告警相关路由
	loginGroup := r.Group("/login")
	{
		loginHandler := &alert.LoginHandler{DB: db.GetDB()}
		// 异常登录告警CURD接口
		loginGroup.GET("/alerts", loginHandler.ListLoginAlerts) // 获取异常登录告警列表(支持分页和搜索)
		//loginGroup.GET("/alerts/:id", loginHandler.GetLoginAlertByID)                    // 根据ID获取异常登录告警详情
		loginGroup.POST("/alerts/status/:id", loginHandler.UpdateLoginAlertStatus) // 更新异常登录告警状态
		//loginGroup.GET("/alerts/whitelist/:id/", loginHandler.UpdateLoginAlertWhitelist) // 更新异常登录告警白名单状态
	}

	// 暴力破解告警相关路由
	passwdGroup := r.Group("/passwd")
	{
		passwdHandler := &alert.PasswdHandler{DB: db.GetDB()}
		// 暴力破解告警CURD接口
		passwdGroup.GET("/alerts", passwdHandler.ListPasswdAlerts) // 获取暴力破解告警列表(支持分页和搜索)
		//passwdGroup.GET("/alerts/:id", passwdHandler.GetPasswdAlertByID)                  // 根据ID获取暴力破解告警详情
		passwdGroup.POST("/alerts/status/:id", passwdHandler.UpdatePasswdAlertStatus) // 更新暴力破解告警状态
		//passwdGroup.GET("/alerts/block/:id/", passwdHandler.UpdatePasswdAlertBlockStatus) // 更新暴力破解告警封禁状态
	}

	// 恶意请求告警相关路由
	requestGroup := r.Group("/request")
	{
		requestHandler := &alert.RequestHandler{DB: db.GetDB()}
		// 恶意请求告警CURD接口
		requestGroup.GET("/alerts", requestHandler.ListRequestAlerts) // 获取恶意请求告警列表(支持分页和搜索)
		//requestGroup.GET("/alerts/:id", requestHandler.GetRequestAlertByID)              // 根据ID获取恶意请求告警详情
		requestGroup.POST("/alerts/status/:id", requestHandler.UpdateRequestAlertStatus) // 更新恶意请求告警状态
	}

	// 网络攻击告警相关路由
	networkGroup := r.Group("/network")
	{
		networkHandler := &alert.NetworkHandler{DB: db.GetDB()}
		// 网络攻击告警CURD接口
		networkGroup.GET("/attacks", networkHandler.ListNetworkAlerts) // 获取网络攻击告警列表(支持分页和搜索)
		//networkGroup.GET("/attacks/:id", networkHandler.GetNetworkAlertByID)              // 根据ID获取网络攻击告警详情
		networkGroup.POST("/attacks/status/:id", networkHandler.UpdateNetworkAlertStatus) // 更新网络攻击告警状态
	}

	// 文件查杀告警相关路由
	fileGroup := r.Group("/file")
	{
		fileHandler := &alert.FileHandler{DB: db.GetDB()}
		// 文件查杀告警CURD接口
		fileGroup.GET("/scans", fileHandler.ListFileAlerts) // 获取文件查杀告警列表(支持分页和搜索)
		//fileGroup.GET("/scans/:id", fileHandler.GetFileAlertByID)                      // 根据ID获取文件查杀告警详情
		fileGroup.POST("/scans/status/:id", fileHandler.UpdateFileAlertStatus) // 更新文件查杀告警状态
		//fileGroup.GET("/scans/quarantine/:id/", fileHandler.UpdateFileAlertQuarantine) // 更新文件查杀告警隔离状态
		//fileGroup.GET("/scans/delete/:id/", fileHandler.UpdateFileAlertDeletion)       // 更新文件查杀告警删除状态
	}

	// 文件完整性告警相关路由
	fileGuardGroup := r.Group("/fileguard")
	{
		fileGuardHandler := &alert.FileGuardHandler{DB: db.GetDB()}
		// 文件完整性告警CURD接口
		fileGuardGroup.GET("/alerts", fileGuardHandler.ListFileGuardAlerts) // 获取文件完整性告警列表(支持分页和搜索)
		//fileGuardGroup.GET("/alerts/:id", fileGuardHandler.GetFileGuardAlertByID)              // 根据ID获取文件完整性告警详情
		fileGuardGroup.POST("/alerts/status/:id", fileGuardHandler.UpdateFileGuardAlertStatus) // 更新文件完整性告警状态
	}
}
