package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/controller/analysis"
	installCtrl "github.com/haolipeng/BeeGuard/server/internal/controller/install"
	"github.com/haolipeng/BeeGuard/server/internal/controller/login"
	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"
	"github.com/haolipeng/BeeGuard/server/internal/http"
	"github.com/haolipeng/BeeGuard/server/internal/middleware"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由，如果提供了 transferServer 则同时注册 gRPC 相关路由
func SetupRouter(transferServer ...*handler.TransferServer) *gin.Engine {
	r := gin.Default()

	// 使用CORS中间件
	r.Use(middleware.CORSMiddleware())
	// 其他中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 健康检查（无需认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 登录相关路由（无需认证）
	loginHandler := &login.Handler{DB: mysql.DB}
	r.POST("/api1/auth/login", loginHandler.Login)
	r.POST("/api1/auth/logout", loginHandler.Logout)

	// Agent 一键安装相关路由（无需认证）
	if config.AppConfig.Install.Enabled {
		installHandler := installCtrl.NewController()
		r.GET("/install.sh", installHandler.GetInstallScript)
		agentInstallGroup := r.Group("/api1/agent")
		{
			agentInstallGroup.GET("/download", installHandler.DownloadPackage)
			agentInstallGroup.GET("/packages", installHandler.ListPackages)
		}
	}

	// 需要认证的 API 路由组
	apiGroup := r.Group("/api1")
	apiGroup.Use(middleware.AuthMiddleware())
	{
		// 用户信息（需要认证）
		apiGroup.GET("/user/info", loginHandler.GetUserInfo)

		// 系统管理相关路由
		systemGroup := apiGroup.Group("/system")
		SetupSystemRouter(systemGroup)

		// 代码安全管理相关路由
		codeGroup := apiGroup.Group("")
		SetupCodeRouter(codeGroup)

		// 后台管理相关路由
		backGroup := apiGroup.Group("/back")
		SetupBackRouter(backGroup)

		// 资产管理相关路由
		assetsGroup := apiGroup.Group("/assets")
		SetupAssetsRouter(assetsGroup)

		// 告警管理相关路由
		alertGroup := apiGroup.Group("/alerts")
		SetupAlertRouter(alertGroup)

		// 漏洞信息管理相关路由
		vulnGroup := apiGroup.Group("/vulns")
		SetupVulnInfoRouter(vulnGroup)

		// 基线路由管理
		baselineGroup := apiGroup.Group("/baseline")
		SetupBaselineRouter(baselineGroup)

		// 概览视图路由管理
		viewGroup := apiGroup.Group("/views")
		SetupViewRouter(viewGroup)

		// 容器漏洞扫描结果管理相关路由
		imageVulnGroup := apiGroup.Group("")
		SetupVulAlertRouter(imageVulnGroup)

		// AI 分析模块相关路由
		analysisGroup := apiGroup.Group("/analysis")
		{
			analysisCtrl := analysis.NewController()
			// 手动触发全局分析任务，对所有待分析的告警进行 AI 分析
			analysisGroup.POST("/trigger", analysisCtrl.TriggerAnalysis)
			// 按主机 IP 分析，针对指定主机的所有告警进行 AI 关联分析
			analysisGroup.POST("/host", analysisCtrl.AnalyzeHost)
			// 按攻击源 IP 分析，针对指��攻击源的所有告警进行 AI 关联分析
			analysisGroup.POST("/source", analysisCtrl.AnalyzeSourceIP)
			// 直接分析提交的告警数据，通过 JSON 传入告警列表进行实时 AI 分析
			analysisGroup.POST("/alerts", analysisCtrl.AnalyzeAlerts)
			// 获取分析报告列表（从内存或文件系统），返回所有已生成的分析报告概要
			analysisGroup.GET("/reports", analysisCtrl.GetReports)
			// 获取单个分析报告详情（从文件系统），根据文件名加载完整报告内容
			analysisGroup.GET("/reports/:filename", analysisCtrl.GetReport)
			// 获取分析模块运行统计信息，包括处理数量、队列状态等指标
			analysisGroup.GET("/stats", analysisCtrl.GetStats)
			// 从数据库获取分析报告列表，支持分页和条件过滤
			analysisGroup.GET("/db_reports", analysisCtrl.ListReportsFromDB)
			// 获取数据库中分析报告的统计信息，包括总数、风险等级分布等
			analysisGroup.GET("/db_reports/stats", analysisCtrl.GetDBReportStats)
			// 从数据库获取单个分析报告详情，根据报告 ID 返回完整内容
			analysisGroup.GET("/db_reports/:id", analysisCtrl.GetReportFromDB)
			// 从数据库删除指定的分析报告，根据报告 ID 执行删除操作
			analysisGroup.DELETE("/db_reports/:id", analysisCtrl.DeleteReportFromDB)
		}

		// 白名单管理路由（需要 transferServer 中的 WlChecker）
		if len(transferServer) > 0 && transferServer[0] != nil {
			whitelistGroup := apiGroup.Group("/whitelist")
			SetupWhitelistRouter(whitelistGroup, transferServer[0].WlChecker)

			// Agent 任务管理路由
			taskGroup := apiGroup.Group("/tasks")
			SetupTaskRouter(taskGroup, transferServer[0])

			// 服务状态路由
			statusGroup := apiGroup.Group("/status")
			SetupStatusRouter(statusGroup, transferServer[0])
		}
	}

	// 如果提供了 transferServer，注册 gRPC 相关路由
	if len(transferServer) > 0 && transferServer[0] != nil {
		http.RegisterGRPCRoutes(r, transferServer[0])
	}

	return r
}
