package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/assets"
	"github.com/haolipeng/BeeGuard/server/internal/controller/assets/container"
	"github.com/haolipeng/BeeGuard/server/internal/controller/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/db"

	"github.com/gin-gonic/gin"
)

// SetupAssetsRouter 设置资产管理相关路由
func SetupAssetsRouter(r *gin.RouterGroup) {
	// 主机资产管理相关路由
	hostGroup := r.Group("/host")
	{
		hostHandler := &host.HostHandler{DB: db.GetDB()}
		// 主机资产CURD接口
		hostGroup.GET("/hosts", hostHandler.ListHosts) // 获取主机资产列表(支持分页和搜索)
	}

	// 端口资产管理相关路由
	portGroup := r.Group("/port")
	{
		portHandler := &host.PortHandler{DB: db.GetDB()}
		// 端口资产CURD接口
		portGroup.GET("/ports", portHandler.ListPorts) // 获取端口资产列表(支持分页和搜索)
	}

	// 账号资产管理相关路由
	accountGroup := r.Group("/account")
	{
		accountHandler := &host.AccountHandler{DB: db.GetDB()}
		// 账号资产CURD接口
		accountGroup.GET("/accounts", accountHandler.ListAccounts) // 获取账号资产列表(支持分页和搜索)
	}

	// 进程资产管理相关路由
	processGroup := r.Group("/process")
	{
		processHandler := &host.ProcessHandler{DB: db.GetDB()}
		// 进程资产查询接口
		processGroup.GET("/processes", processHandler.ListProcesses) // 获取进程资产列表(支持分页和搜索)
	}

	// 数据库资产管理相关路由
	databaseGroup := r.Group("/database")
	{
		databaseHandler := &host.DatabaseHandler{DB: db.GetDB()}
		// 数据库资产查询接口
		databaseGroup.GET("/databases", databaseHandler.ListDatabases) // 获取数据库资产列表(支持分页和搜索)
	}

	// Web服务资产管理相关路由
	webGroup := r.Group("/web")
	{
		webHandler := &host.WebHandler{DB: db.GetDB()}
		// Web服务资产查询接口
		webGroup.GET("/webs", webHandler.ListWebs) // 获取Web服务资产列表(支持分页和搜索)
	}

	// 系统服务资产管理相关路由
	systemGroup := r.Group("/system")
	{
		systemHandler := &host.SystemHandler{DB: db.GetDB()}
		// 系统服务资产查询接口
		systemGroup.GET("/systems", systemHandler.ListSystems) // 获取系统服务资产列表(支持分页和搜索)
	}

	// 容器资产管理相关路由
	containerGroup := r.Group("/container")
	{
		containerHandler := &container.ContainerHandler{DB: db.GetDB()}
		// 容器资产查询接口
		containerGroup.GET("/containers", containerHandler.ListContainers) // 获取容器资产列表(支持分页和搜索)

		imageHandler := &container.ImageHandler{DB: db.GetDB()}
		// 镜像资产查询接口
		containerGroup.GET("/images", imageHandler.ListImages) // 获取镜像资产列表(支持分页和搜索)
	}

	// 资产视图相关路由
	viewGroup := r.Group("/view")
	{
		viewHandler := &assets.ViewHandler{DB: db.GetDB()}

		viewGroup.GET("/os-type-stats", viewHandler.GetOSTypeStats) // 系统类型统计

		viewGroup.GET("/host-stats", viewHandler.GetHostStats) // 主机统计

		viewGroup.GET("/database-type-stats", viewHandler.GetDatabaseTypeStats) // 数据库类型统计

		viewGroup.GET("/database-stats", viewHandler.GetDatabaseStats) // 数据库统计

		viewGroup.GET("/container-stats", viewHandler.GetContainerStats) // 容器统计

		viewGroup.GET("/account-stats", viewHandler.GetAccountStats) // 账号统计

		viewGroup.GET("/latest-assets-top5", viewHandler.GetLatestAssetsTop5) // 近期更新资产
	}
}
