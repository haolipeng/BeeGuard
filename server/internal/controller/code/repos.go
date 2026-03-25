package code

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/code"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReposHandler 仓库处理器结构体，用于处理与仓库相关的HTTP请求
type ReposHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateRepos 创建仓库
func (h *ReposHandler) CreateRepos(c *gin.Context) {
	// 接收数据
	var repos code.Repos
	// 验证并绑定请求中的JSON数据到repos结构体
	if err := c.ShouldBindJSON(&repos); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 打印接收的 数据
	//fmt.Println("========== 接收到的所有数据 ==========")
	//fmt.Println(repos)

	// 创建一个只包含所需字段的新结构体实例
	newRepo := code.Repos{
		RepoName:      repos.RepoName,
		CodeqlRules:   repos.CodeqlRules,
		RepoURL:       repos.RepoURL,
		Language:      repos.Language,
		Description:   repos.Description,
		ScanFrequency: repos.ScanFrequency,
		Owner:         repos.Owner,
	}
	// 打印newRepo
	//fmt.Println("========== 接收到的可选数据 ==========")
	fmt.Println(newRepo)

	// 执行数据库操作，只保存指定字段
	result := h.DB.Select("repo_name", "repo_url", "language", "description", "scan_frequency", "codeql_rules", "owner").Create(&newRepo)
	//result := h.DB.Select("repo_name", "repo_url", "language").Create(&newRepo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    newRepo,
	})
}

// GetRepos 获取单个仓库
func (h *ReposHandler) GetRepos(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repos code.Repos

	result := h.DB.Where("repo_id = ?", id).First(&repos)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "仓库不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": repos})
}

// ListRepos 获取仓库列表（支持搜索查询）
func (h *ReposHandler) ListRepos(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	repoName := c.Query("repo_name")
	repoURL := c.Query("repo_url")
	status := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var repos []code.Repos
	var total int64

	// 构建查询条件
	query := h.DB.Model(&code.Repos{})

	// 添加搜索条件
	if repoName != "" {
		query = query.Where("repo_name LIKE ?", "%"+repoName+"%")
	}
	if repoURL != "" {
		query = query.Where("repo_url LIKE ?", "%"+repoURL+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&repos)
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
		"data": repos,
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

// UpdateRepos 更新仓库信息
func (h *ReposHandler) UpdateRepos(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repos code.Repos

	if err := c.ShouldBindJSON(&repos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置repo_id为URL路径参数中的ID值
	repos.RepoID = int64(id)

	//// ========== 核心：美观打印接收的参数 ==========
	//// 方案1：用json.MarshalIndent（无第三方依赖，格式化JSON，易读）
	//fmt.Println("========== 接收的Repos参数 ==========")
	//// 格式化JSON，带缩进，空字段也会显示
	//reposJSON, _ := json.MarshalIndent(repos, "", "  ")
	//
	//fmt.Printf("参数详情（JSON格式）：\n%s\n", string(reposJSON))
	//fmt.Println("====================================")
	////打印repos的RepoID
	//fmt.Println("========== RepoID ==========")
	//fmt.Println(repos.RepoID)

	//删除仓库codeql_scan_results中的数据
	h.DB.Where("repo_id = ?", id).Delete(&code.CodeqlScanResults{})
	//// 删除仓库code_vuldetail中的数据
	//h.DB.Where("repo_id = ?", id).Delete(&code.CodeVulDetail{})

	result := h.DB.Model(&code.Repos{}).Where("repo_id = ?", id).Updates(&repos)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的完整记录
	var updatedRepo code.Repos
	queryResult := h.DB.Where("repo_id = ?", id).First(&updatedRepo)
	if queryResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询更新后的数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    updatedRepo,
	})
}

// DeleteRepos 删除仓库
func (h *ReposHandler) DeleteRepos(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("repo_id = ?", id).Delete(&code.Repos{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
