package code

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/models/code"
)

// CreateRepoScanList 创建仓库扫描记录
func CreateRepoScanList(c *gin.Context) {
	var repoScanList code.RepoScanList

	// 绑定JSON数据
	if err := c.ShouldBindJSON(&repoScanList); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取数据库连接
	database := db.GetDB()

	// 保存到数据库
	if err := database.Create(&repoScanList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    repoScanList,
	})
}

// GetRepoScanListByID 根据ID获取仓库扫描记录
func GetRepoScanListByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repoScanList code.RepoScanList
	database := db.GetDB()

	// 查询repo_id字段等于id的记录
	if err := database.Where("repo_id = ?", uint(id)).First(&repoScanList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "查询成功",
		"data":    repoScanList,
	})
}

// GetAllRepoScanLists 获取所有仓库扫描记录
func GetAllRepoScanLists(c *gin.Context) {
	var repoScanLists []code.RepoScanList // 定义变量，用来存储查询到的扫描列表数据
	database := db.GetDB()                  // 获取数据库连接

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 构建查询条件
	query := database.Model(&code.RepoScanList{})

	// 添加repo_id查询条件
	if repoIDStr := c.Query("repo_id"); repoIDStr != "" {
		if repoID, err := strconv.ParseInt(repoIDStr, 10, 64); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的repo_id参数"})
			return
		} else {
			// 明确设置repo_id查询条件
			query = query.Where("repo_id = ?", repoID)
			fmt.Printf("正在查询 repo_id=%d 的记录\n", repoID)
		}
	}

	// 添加其他可选查询条件
	if repoName := c.Query("repo_name"); repoName != "" {
		query = query.Where("repo_name LIKE ?", "%"+repoName+"%")
	}
	if severity := c.Query("severity"); severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if scanType := c.Query("scan_type"); scanType != "" {
		query = query.Where("scan_type = ?", scanType)
	}
	if branch := c.Query("branch"); branch != "" {
		query = query.Where("branch LIKE ?", "%"+branch+"%")
	}

	// 查询总数用于分页
	var total int64
	query.Count(&total) // 查询总数
	fmt.Printf("总记录数: %d\n", total)

	// 执行数据库查询，获取分页后的扫描列表数据
	if err := query.Offset(offset).Limit(pageSize).Find(&repoScanLists).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}
	// 输出查询结果数量
	fmt.Printf("查询结果数量: %d\n", len(repoScanLists))
	if len(repoScanLists) > 0 {
		for _, item := range repoScanLists {
			fmt.Printf("ID: %d, RepoID: %d\n", item.ID, item.RepoID)
		}
	}

	// 如果查询成功，将返回包含以下内容的JSON响应：
	c.JSON(http.StatusOK, gin.H{
		"message":  "查询成功",
		"data":     repoScanLists,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// UpdateRepoScanList 更新仓库扫描记录
func UpdateRepoScanList(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repoScanList code.RepoScanList
	database := db.GetDB()

	// 检查记录是否存在
	if err := database.First(&repoScanList, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	// 绑定更新数据
	var updateData code.RepoScanList
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新记录
	if err := database.Model(&repoScanList).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	// 重新查询以获取完整数据
	database.First(&repoScanList, uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    repoScanList,
	})
}

// DeleteRepoScanList 删除仓库扫描记录
func DeleteRepoScanList(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var repoScanList code.RepoScanList
	database := db.GetDB()

	// 检查记录是否存在
	if err := database.First(&repoScanList, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "记录不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
		return
	}

	// 删除记录
	if err := database.Delete(&repoScanList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}
