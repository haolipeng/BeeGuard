package code

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/models/code"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateRepoScanResult 创建代码审计仓库结果
func CreateRepoScanResult(c *gin.Context) {
	var repoScanResult code.CodeqlScanResults
	if err := c.ShouldBindJSON(&repoScanResult); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := db.GetDB().Create(&repoScanResult)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建成功", "data": repoScanResult})
}

// GetRepoScanResultByID 根据ID获取代码审计仓库结果
func GetRepoScanResultByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repoScanResult code.CodeqlScanResults

	result := db.GetDB().Where("id = ?", id).First(&repoScanResult)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "结果不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": repoScanResult})
}

// GetAllRepoScanResults 获取代码审计仓库结果列表(支持分页和repo_id过滤)
func GetAllRepoScanResults(c *gin.Context) {
	// 获取用户传过来的参数
	repoIDStr := c.Query("repo_id")
	repoName := c.Query("repo_name") // 仓库名称搜索
	ruleName := c.Query("rule_name") // 漏洞名称/规则名称搜索

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var repoScanResults []code.CodeqlScanResults
	var total int64

	// 构建查询条件
	query := db.GetDB().Model(&code.CodeqlScanResults{})

	// 添加仓库ID过滤条件
	if repoIDStr != "" {
		repoID, err := strconv.Atoi(repoIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的repo_id参数"})
			return
		}
		query = query.Where("repo_id = ?", repoID)
	}

	// 添加仓库名称模糊搜索条件
	if repoName != "" {
		// 使用更严格的匹配方式，确保只返回包含指定名称的记录
		query = query.Where("LOWER(repo_name) LIKE LOWER(?)", "%"+repoName+"%")
		fmt.Printf("搜索仓库名称: %s\n", repoName)
	}

	// 添加漏洞名称/规则名称模糊搜索条件
	if ruleName != "" {
		// 使用更严格的匹配方式
		query = query.Where("LOWER(rule_name) LIKE LOWER(?)", "%"+ruleName+"%")
		fmt.Printf("搜索规则名称: %s\n", ruleName)
	}

	// 先获取总记录数
	result := query.Order("id DESC").Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据
	result = query.Order("id DESC").Limit(limit).Offset(offset).Find(&repoScanResults)
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
		"data": repoScanResults,
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

// UpdateRepoScanResult 更新代码审计仓库结果
func UpdateRepoScanResult(c *gin.Context) {
	//会传入repo_id和status
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repoScanResult code.CodeqlScanResults
	if err := c.ShouldBindJSON(&repoScanResult); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := db.GetDB().Model(&code.CodeqlScanResults{}).Where("id = ?", id).Updates(&repoScanResult)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": repoScanResult})
}

// DeleteRepoScanResult 删除代码审计仓库结果
func DeleteRepoScanResult(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := db.GetDB().Where("result_id = ?", id).Delete(&code.CodeqlScanResults{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetVulDetailByScanResultIDAndPath 根据scan_results_id和path获取漏洞详情
func GetVulDetailByScanResultIDAndPath(c *gin.Context) {
	scanResultsIDStr := c.Query("scan_results_id")
	path := c.Query("path")

	if scanResultsIDStr == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scan_results_id和path参数不能为空"})
		return
	}

	scanResultsID, err := strconv.Atoi(scanResultsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的scan_results_id"})
		return
	}

	var vulDetail code.CodeVulDetail

	result := db.GetDB().Where("scan_results_id = ? AND path = ?", scanResultsID, path).First(&vulDetail)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "漏洞详情不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vulDetail})
}
