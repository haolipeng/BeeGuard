package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/vul"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"

	"github.com/gin-gonic/gin"
)

// SetupVulAlertRouter 配置容器漏洞扫描结果管理相关路由
func SetupVulAlertRouter(r *gin.RouterGroup) {
	// 容器漏洞扫描结果管理相关路由
	imageVulnGroup := r.Group("/vulns/image_details")
	{
		imageViewCountHandler := &vul.ImageViewCountHandler{DB: mysql.DB}
		imageVulnGroup.GET("/:id", imageViewCountHandler.GetImageVulnDetail)
		imageVulnGroup.GET("", imageViewCountHandler.ListImageVulnDetails)
	}
}
