package login

import (
	"net/http"

	"github.com/haolipeng/BeeGuard/server/internal/models/system"
	"github.com/haolipeng/BeeGuard/server/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token    string      `json:"token"`
	UserInfo *UserInfo   `json:"user_info"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// Handler 登录处理器
type Handler struct {
	DB *gorm.DB
}

// Login 登录接口
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 查询用户
	var user system.User
	result := h.DB.Where("username = ?", req.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "用户名或密码错误",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "数据库查询失败",
		})
		return
	}

	// 验证密码（明文比对）
	if user.Passwd != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "用户名或密码错误",
		})
		return
	}

	// 检查账号状态
	if user.AccountStatus == "disabled" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "账号已被禁用",
		})
		return
	}

	// 生成 token
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Name, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成 token 失败",
		})
		return
	}

	// 返回登录成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登录成功",
		"data": LoginResponse{
			Token: token,
			UserInfo: &UserInfo{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
				Role:     user.Role,
			},
		},
	})
}

// GetUserInfo 获取当前用户信息（需要认证中间件保护）
func (h *Handler) GetUserInfo(c *gin.Context) {
	// 从上下文中获取用户信息（由认证中间件设置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未授权",
		})
		return
	}

	var user system.User
	result := h.DB.Where("id = ?", userID).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
		},
	})
}

// Logout 登出接口
func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "登出成功",
	})
}
