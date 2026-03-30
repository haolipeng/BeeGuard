package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/status"
	grpcHandler "github.com/haolipeng/BeeGuard/server/internal/grpc/handler"

	"github.com/gin-gonic/gin"
)

// SetupStatusRouter 配置服务状态监控相关路由
func SetupStatusRouter(r *gin.RouterGroup, transferServer *grpcHandler.TransferServer) {
	handler := &status.Handler{
		TransferServer: transferServer,
	}

	r.GET("/server", handler.GetServerStatus)
	r.GET("/database", handler.GetDatabaseStatus)
	r.GET("/agents", handler.GetAgentsStatus)
	r.GET("/overview", handler.GetOverview)
}
