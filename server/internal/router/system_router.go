package router

import (
	"github.com/haolipeng/BeeGuard/server/internal/controller/system"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"

	"github.com/gin-gonic/gin"
)

// SetupSystemRouter 配置系统管理相关路由
func SetupSystemRouter(r *gin.RouterGroup) {
	// user 管理
	userHandler := &system.UserHandler{DB: mysql.DB}
	r.POST("/users/create", userHandler.CreateUser) // 创建用户 curd 接口 1
	//	curl -X POST "http://localhost:8080/api1/system/users/create" \ -H "Content-Type: application/json" \ -d '{
	//	"username": "test",
	//		"passwd": "123456",
	//		"name": "测试用户",
	//		"role": "admin",
	//		"account_status": "active"
	//}'

	r.GET("/users/:id", userHandler.GetUser)            // 获取单个用户 curl -X GET "http://localhost:8080/api1/system/users/1"
	r.GET("/users/edit/:id", userHandler.UpdateUser)    // 更新用户 curl -X POST "http://localhost:8080/api1/system/users/edit/1"
	r.POST("/users/delete/:id", userHandler.DeleteUser) // 删除用户
	r.GET("/users", userHandler.ListUsers)              // 获取用户列表 (支持分页和搜索)

	// Agent 客户端管理相关路由
	agentGroup := r.Group("/agents")
	{
		agentHandler := &system.AgentInfoHandler{DB: mysql.DB}
		agentGroup.POST("/create", agentHandler.CreateAgentInfo)
		agentGroup.GET("/:id", agentHandler.GetAgentInfo)
		agentGroup.GET("", agentHandler.ListAgentInfos)
		agentGroup.POST("/edit/:id", agentHandler.UpdateAgentInfo)
		agentGroup.POST("/delete/:id", agentHandler.DeleteAgentInfo)
		agentGroup.POST("/connection/:id", agentHandler.UpdateAgentConnectionStatus)
	}
}
