package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/task"
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"

	"github.com/gin-gonic/gin"
)

// SetupTaskRouter 配置 Agent 任务管理相关路由
func SetupTaskRouter(r *gin.RouterGroup, transferServer *handler.TransferServer) {
	taskHandler := &task.Handler{
		DB:             db.GetDB(),
		Repo:           repository.NewTaskRepository(db.GetDB()),
		TransferServer: transferServer,
	}

	r.POST("/send", taskHandler.SendTask)
	r.GET("/history", taskHandler.ListHistory)
	r.GET("/history/:id", taskHandler.GetHistory)
	r.GET("/types", taskHandler.GetTaskTypes)
}
