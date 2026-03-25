package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/code"
	"github.com/haolipeng/BeeGuard/server/internal/db"

	"github.com/gin-gonic/gin"
)

// SetupCodeRouter 设置代码安全管理相关路由
func SetupCodeRouter(r *gin.RouterGroup) {
	// 代码仓库列表相关路由
	codeGroup := r.Group("/code_repos")
	{
		reposHandler := &code.ReposHandler{DB: db.GetDB()}
		codeGroup.POST("/repos/create", reposHandler.CreateRepos)
		codeGroup.GET("/repos/:id", reposHandler.GetRepos)
		codeGroup.POST("/repos/edit/:id", reposHandler.UpdateRepos)
		codeGroup.POST("/repos/delete/:id", reposHandler.DeleteRepos)
		codeGroup.GET("/repos", reposHandler.ListRepos)
	}

	// 仓库扫描结果列表相关路由
	codeScanResultGroup := r.Group("/code_scan_result")
	{
		codeScanResultGroup.POST("/results", code.CreateRepoScanResult)
		codeScanResultGroup.GET("/results/:id", code.GetRepoScanResultByID)
		codeScanResultGroup.POST("/results/status/:id", code.UpdateRepoScanResult)
		codeScanResultGroup.GET("/results/all", code.GetAllRepoScanResults)
		codeScanResultGroup.GET("/vul_detail", code.GetVulDetailByScanResultIDAndPath)
	}

	// 规则集管理相关路由
	rulesGroup := r.Group("/rules")
	{
		rulesGroup.POST("/rules", code.CreateRule)
		rulesGroup.GET("/rules/:id", code.GetRuleByID)
		rulesGroup.POST("/rules/update/:id", code.UpdateRule)
		rulesGroup.POST("/rules/delete/:id", code.DeleteRule)
		rulesGroup.GET("/rules", code.GetAllRules)
		rulesGroup.GET("/rules_all", code.GetAllRuleList)
	}
}