package system

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/system"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler 系统用户处理器结构体，用于处理与系统用户相关的HTTP请求
type UserHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateUser 创建系统用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 接收数据
	var user system.User
	// 验证并绑定请求中的JSON数据到user结构体
	if err := c.ShouldBindJSON(&user); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打印接收的数据
	fmt.Println(user)

	// 创建一个只包含所需字段的新结构体实例
	newUser := system.User{
		Username:      user.Username,
		Passwd:        user.Passwd,  // 添加这一行，确保密码字段被包含
		Name:          user.Name,
		Role:          user.Role,
		AccountStatus: user.AccountStatus,
	}

	// 执行数据库操作，只保存指定字段
	result := h.DB.Select("username", "passwd", "name", "role", "account_status").Create(&newUser)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    newUser,
	})
}

// GetUser 获取单个系统用户
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var user system.User

	result := h.DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// ListUsers 获取系统用户列表（支持搜索查询）
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	username := c.Query("username")
	name := c.Query("name")
	role := c.Query("role")
	accountStatus := c.Query("account_status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var users []system.User
	var total int64

	// 构建查询条件
	query := h.DB.Model(&system.User{})

	// 添加搜索条件
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if accountStatus != "" {
		query = query.Where("account_status = ?", accountStatus)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  total,
			"per_page":     limit,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
	})
}

// UpdateUser 更新系统用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var user system.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置ID为URL路径参数中的ID值
	user.ID = int64(id)

	result := h.DB.Model(&system.User{}).Where("id = ?", id).Updates(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的完整记录
	var updatedUser system.User
	queryResult := h.DB.Where("id = ?", id).First(&updatedUser)
	if queryResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询更新后的数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    updatedUser,
	})
}

// DeleteUser 删除系统用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&system.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}