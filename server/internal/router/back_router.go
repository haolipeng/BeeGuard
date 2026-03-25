package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/back"
	"github.com/haolipeng/BeeGuard/server/internal/controller/code"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"

	"github.com/gin-gonic/gin"
)

// SetupBackRouter 设置后台管理相关路由
func SetupBackRouter(r *gin.RouterGroup) {
	// 规则集管理相关路由
	rulesGroup := r.Group("/rules")
	{
		// 规则集列表接口
		rulesGroup.POST("", code.CreateRule)            // 创建规则集
		rulesGroup.GET("/:id", code.GetRuleByID)        // 获取规则集详情
		rulesGroup.POST("/update/:id", code.UpdateRule) // 更新规则集
		rulesGroup.POST("/delete/:id", code.DeleteRule) // 删除规则集
		rulesGroup.GET("", code.GetAllRules)            // 获取规则集列表(支持分页和搜索)
		rulesGroup.GET("/all", code.GetAllRuleList)
	}

	// 规则详情管理相关路由
	ruleDetailGroup := r.Group("/code_rule")
	{
		ruleHandler := &back.CodeRuleHandler{DB: mysql.DB}
		// 规则详情CURD接口
		ruleDetailGroup.POST("/rules/create", ruleHandler.CreateCodeRule)       // 创建规则
		ruleDetailGroup.DELETE("/rules/delete/:id", ruleHandler.DeleteCodeRule) // 删除规则
		ruleDetailGroup.GET("/rules", ruleHandler.ListCodeRules)                // 获取规则列表(支持分页和搜索)
	}

	// 入侵检测规则管理相关路由
	hidsRuleGroup := r.Group("/hids_rules")
	{
		hidsRuleHandler := &back.HIDSRuleHandler{DB: mysql.DB}
		// 入侵检测规则 CURD 接口
		hidsRuleGroup.POST("/create", hidsRuleHandler.CreateHIDSRule) // 创建规则
		hidsRuleGroup.GET("/:id", hidsRuleHandler.GetHIDSRule)        // 获取单个规则
		hidsRuleGroup.GET("", hidsRuleHandler.ListHIDSRules)          // 获取规则列表
		//hidsRuleGroup.POST("/:id", hidsRuleHandler.UpdateHIDSRule)        // 更新规则
		hidsRuleGroup.POST("/delete/:id", hidsRuleHandler.DeleteHIDSRule) // 删除规则
	}

	// 漏洞检测规则管理相关路由
	vulnRuleGroup := r.Group("/vuln_rules")
	{
		vulnRuleHandler := &back.VulnerabilityInfoHandler{DB: mysql.DB}
		// 漏洞检测规则CURD接口
		vulnRuleGroup.POST("/create", vulnRuleHandler.CreateVulnerabilityInfo) // 创建漏洞规则
		vulnRuleGroup.GET("/:id", vulnRuleHandler.GetVulnerabilityInfo)        // 获取单个漏洞规则
		vulnRuleGroup.GET("", vulnRuleHandler.ListVulnerabilityInfos)          // 获取漏洞规则列表
		//vulnRuleGroup.POST("/:id", vulnRuleHandler.UpdateVulnerabilityInfo)        // 更新漏洞规则
		vulnRuleGroup.POST("/delete/:id", vulnRuleHandler.DeleteVulnerabilityInfo) // 删除漏洞规则
	}
}
