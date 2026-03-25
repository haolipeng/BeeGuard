package http

import (
	"fmt"

	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"
	"github.com/haolipeng/BeeGuard/server/internal/log"

	"github.com/gin-gonic/gin"
)

// Server HTTP API 服务器
type Server struct {
	transferServer *handler.TransferServer
	router         *gin.Engine
	httpPort       int
}

// NewServer 创建新的 HTTP API 服务器
func NewServer(ts *handler.TransferServer, port int) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[API] %s %s %s %d %s\n",
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	s := &Server{
		transferServer: ts,
		router:         r,
		httpPort:       port,
	}

	return s
}

// Start 启动 HTTP 服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.httpPort)
	log.Infof("[API] HTTP server starting on %s", addr)
	return s.router.Run(addr)
}

// RegisterGRPCRoutes 将 gRPC 管理相关的路由注册到指定的 gin.Engine
// 用于将路由集成到其他路由组中
func RegisterGRPCRoutes(r *gin.Engine, transferServer *handler.TransferServer) {
	apiServer := &Server{
		transferServer: transferServer,
	}

	// gRPC 管理相关路由
	api := r.Group("/api")
	{
		api.POST("/task", apiServer.SendTask)
		api.POST("/config", apiServer.SendPluginConfig)
		api.POST("/detector/config", apiServer.SendDetectorConfig)
		api.POST("/baseline/check", apiServer.SendBaselineCheck)
		api.POST("/agent/uninstall", apiServer.UninstallAgent)
		api.GET("/agents", apiServer.ListAgents)
		api.GET("/agents/:id", apiServer.GetAgent)
	}
}
