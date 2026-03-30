package status

import (
	"net/http"
	"runtime"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"

	"github.com/gin-gonic/gin"
)

// Handler 服务状态监控控制器
type Handler struct {
	TransferServer *handler.TransferServer
}

// GetServerStatus 获取服务器自身状态
func (h *Handler) GetServerStatus(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.JSON(http.StatusOK, gin.H{
		"goroutines":     runtime.NumGoroutine(),
		"alloc_mb":       memStats.Alloc / 1024 / 1024,
		"total_alloc_mb": memStats.TotalAlloc / 1024 / 1024,
		"sys_mb":         memStats.Sys / 1024 / 1024,
		"num_gc":         memStats.NumGC,
		"go_version":     runtime.Version(),
		"num_cpu":        runtime.NumCPU(),
	})
}

// GetDatabaseStatus 获取 PostgreSQL 连接状态
func (h *Handler) GetDatabaseStatus(c *gin.Context) {
	gormDB := db.GetDB()
	if gormDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "disconnected"})
		return
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "error": err.Error()})
		return
	}

	stats := sqlDB.Stats()
	err = sqlDB.Ping()
	status := "connected"
	if err != nil {
		status = "error"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        status,
		"open_connections": stats.OpenConnections,
		"in_use":        stats.InUse,
		"idle":          stats.Idle,
		"max_open":      stats.MaxOpenConnections,
		"wait_count":    stats.WaitCount,
	})
}

// GetAgentsStatus 获取 gRPC 连接数和 Agent 在线数
func (h *Handler) GetAgentsStatus(c *gin.Context) {
	agents := h.TransferServer.GetAgents()

	c.JSON(http.StatusOK, gin.H{
		"online_count": len(agents),
	})
}

// GetOverview 获取综合状态（前端轮询用）
func (h *Handler) GetOverview(c *gin.Context) {
	// Server
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	serverStatus := gin.H{
		"goroutines": runtime.NumGoroutine(),
		"alloc_mb":   memStats.Alloc / 1024 / 1024,
		"sys_mb":     memStats.Sys / 1024 / 1024,
		"num_cpu":    runtime.NumCPU(),
	}

	// Database
	dbStatus := "unknown"
	var dbConns int
	gormDB := db.GetDB()
	if gormDB != nil {
		if sqlDB, err := gormDB.DB(); err == nil {
			if err := sqlDB.Ping(); err == nil {
				dbStatus = "connected"
			} else {
				dbStatus = "error"
			}
			stats := sqlDB.Stats()
			dbConns = stats.OpenConnections
		}
	}

	// Agents
	agents := h.TransferServer.GetAgents()

	c.JSON(http.StatusOK, gin.H{
		"server": serverStatus,
		"database": gin.H{
			"status":      dbStatus,
			"connections": dbConns,
		},
		"agents": gin.H{
			"online_count": len(agents),
		},
	})
}
