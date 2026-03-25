package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/vul"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"

	"github.com/gin-gonic/gin"
)

// SetupVulnInfoRouter 设置漏洞信息管理相关路由
func SetupVulnInfoRouter(r *gin.RouterGroup) {
	// 漏洞信息相关路由
	// 漏洞发现-主机漏洞-漏洞视角
	vulnInfoGroup := r.Group("vul")
	{
		vulnInfoHandler := &vul.VulnInfoHandler{DB: mysql.DB}
		// 漏洞主机统计接口（基于视图）
		vulnInfoGroup.GET("/hostscount", vulnInfoHandler.ListVulnWithHosts) // 获取漏洞主机统计列表
	}
	// 主机漏洞扫描任务相关路由
	hostVulnScanTaskGroup := r.Group("host")
	{
		hostVulnScanTaskHandler := &vul.HostVulnScanTaskHandler{DB: mysql.DB}
		// 主机视角-基于视图的漏洞统计接口
		hostVulnScanTaskGroup.GET("/stats", hostVulnScanTaskHandler.ListVulnCountHosts) // 获取主机漏洞统计列表(支持分页和搜索)
	}

	// 主机漏洞详情相关路由
	hostVulnDetailGroup := r.Group("hostdetail")
	{
		hostVulnDetailHandler := &vul.HostVulnDetailHandler{DB: mysql.DB}
		// 主机漏洞详情CURD接口
		hostVulnDetailGroup.GET("/counts", hostVulnDetailHandler.ListHostVulnDetails)   // 获取主机漏洞详情列表(支持分页和搜索)
		hostVulnDetailGroup.GET("/counts/:id", hostVulnDetailHandler.GetHostVulnDetail) // 获取单个主机漏洞详情
	}

	// 镜像漏洞统计相关路由
	imageViewCountGroup := r.Group("image")
	{
		imageViewCountHandler := &vul.ImageViewCountHandler{DB: mysql.DB}
		// 镜像视角-基于视图的漏洞统计接口
		imageViewCountGroup.GET("/imagecount", imageViewCountHandler.ListVulnCountImages) // 获取镜像漏洞统计列表(支持分页和搜索)

		// 镜像漏洞详情CURD接口
		imageViewCountGroup.GET("/details", imageViewCountHandler.ListImageVulnDetails)   // 获取镜像漏洞详情列表(支持分页和搜索)
		imageViewCountGroup.GET("/details/:id", imageViewCountHandler.GetImageVulnDetail) // 获取单个镜像漏洞详情

		// 镜像漏洞统计相关路由（基于视图v_vuln_count_vuls）
		imagesVulViewCountGroup := r.Group("imagevul")
		{
			imagesVulViewCountHandler := &vul.ImagesVulViewCountHandler{DB: mysql.DB}
			// 漏洞视图 统计接口（基于视图）
			imagesVulViewCountGroup.GET("/vulcounts", imagesVulViewCountHandler.ListImagesVulViewCounts) // 获取镜像漏洞统计列表
			imagesVulViewCountGroup.GET("/counts/:id", imagesVulViewCountHandler.GetImagesVulViewCount)  // 获取单个镜像漏洞统计详情

			// 镜像漏洞信息CURD接口
			imagesVulViewCountGroup.GET("/infos", imagesVulViewCountHandler.ListImageVulnerabilities)  // 获取镜像漏洞信息列表(支持分页和搜索)
			imagesVulViewCountGroup.GET("/infos/:id", imagesVulViewCountHandler.GetImageVulnerability) // 获取单个镜像漏洞信息
		}
	}
}
