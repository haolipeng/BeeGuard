package middleware

import (
	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// CORSMiddleware 处理跨域请求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查来源是否在允许列表中
		allowed := false
		for _, allowedOrigin := range config.AppConfig.Server.CORS.AllowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}
		
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// 如果没有匹配的源，但配置中有通配符，则允许所有
			for _, allowedOrigin := range config.AppConfig.Server.CORS.AllowedOrigins {
				if allowedOrigin == "*" {
					c.Header("Access-Control-Allow-Origin", "*")
					break
				}
			}
		}
		
		// 允许的请求方法
		methods := strings.Join(config.AppConfig.Server.CORS.AllowedMethods, ", ")
		c.Header("Access-Control-Allow-Methods", methods)
		
		// 允许的请求头
		headers := strings.Join(config.AppConfig.Server.CORS.AllowedHeaders, ", ")
		c.Header("Access-Control-Allow-Headers", headers)
		
		// 暴露的响应头
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, X-CSRF-Token")
		
		// 是否允许携带cookie
		if config.AppConfig.Server.CORS.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 如果是预检请求，直接返回200状态码
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// 继续执行其他中间件或处理函数
		c.Next()
	}
}