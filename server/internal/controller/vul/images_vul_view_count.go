package vul

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/vul"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImagesVulViewCountHandler 镜像漏洞统计处理器结构体，用于处理与镜像漏洞统计相关的HTTP请求
type ImagesVulViewCountHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListImagesVulViewCounts 获取镜像漏洞统计列表（基于视图）
func (h *ImagesVulViewCountHandler) ListImagesVulViewCounts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	cveID := c.Query("cve_id")
	vulnName := c.Query("vuln_name")
	severity := c.Query("severity")
	minImageCount := c.Query("min_image_count")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var imagesVulViewCounts []vul.ImagesVulViewCount
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.ImagesVulViewCount{})

	// 添加搜索条件
	if cveID != "" {
		query = query.Where("cve_id LIKE ?", "%"+cveID+"%")
	}
	if vulnName != "" {
		query = query.Where("vuln_name LIKE ?", "%"+vulnName+"%")
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if minImageCount != "" {
		if count, err := strconv.Atoi(minImageCount); err == nil {
			query = query.Where("affected_image_count >= ?", count)
		}
	}
	// 添加时间范围查询条件
	if startTime != "" {
		query = query.Where("first_scan_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("last_scan_time <= ?", endTime)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按影响镜像数量倒序排列
	result = query.Order("affected_image_count DESC, last_scan_time DESC").Limit(limit).Offset(offset).Find(&imagesVulViewCounts)
	if result.Error != nil {
		// 输出详细错误信息到日志
		fmt.Printf("数据库查询错误: %v\n", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "details": result.Error.Error()})
		return
	}

	// 解析受影响镜像详情
	for i := range imagesVulViewCounts {
		if len(imagesVulViewCounts[i].AffectedImages) > 0 {
			var affectedImages []vul.AffectedImage
			if err := json.Unmarshal(imagesVulViewCounts[i].AffectedImages, &affectedImages); err == nil {
				// 只保留前5个镜像信息以减少响应大小
				if len(affectedImages) > 5 {
					affectedImages = affectedImages[:5]
				}
				// 重新序列化截取后的数据
				if jsonData, err := json.Marshal(affectedImages); err == nil {
					imagesVulViewCounts[i].AffectedImages = jsonData
				}
			}
		}
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": imagesVulViewCounts,
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

// GetImagesVulViewCount 获取单个镜像漏洞统计详情
func (h *ImagesVulViewCountHandler) GetImagesVulViewCount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var imagesVulViewCount vul.ImagesVulViewCount

	result := h.DB.Where("vuln_id = ?", id).First(&imagesVulViewCount)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "镜像漏洞统计信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 解析受影响镜像详情
	if len(imagesVulViewCount.AffectedImages) > 0 {
		var affectedImages []vul.AffectedImage
		if err := json.Unmarshal(imagesVulViewCount.AffectedImages, &affectedImages); err == nil {
			// 重新序列化完整的数据
			if jsonData, err := json.Marshal(affectedImages); err == nil {
				imagesVulViewCount.AffectedImages = jsonData
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": imagesVulViewCount})
}

// CreateImageVulnerability 创建镜像漏洞信息
func (h *ImagesVulViewCountHandler) CreateImageVulnerability(c *gin.Context) {
	// 接收数据
	var imageVulnerability vul.ImageVulnerability
	// 验证并绑定请求中的JSON数据到imageVulnerability结构体
	if err := c.ShouldBindJSON(&imageVulnerability); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行数据库操作
	result := h.DB.Create(&imageVulnerability)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    imageVulnerability,
	})
}

// GetImageVulnerability 获取单个镜像漏洞信息
func (h *ImagesVulViewCountHandler) GetImageVulnerability(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var imageVulnerability vul.ImageVulnerability

	result := h.DB.Where("id = ?", id).First(&imageVulnerability)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "镜像漏洞信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": imageVulnerability})
}

// ListImageVulnerabilities 获取镜像漏洞信息列表（支持搜索查询）
func (h *ImagesVulViewCountHandler) ListImageVulnerabilities(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	cveID := c.Query("cve_id")
	vulnName := c.Query("vuln_name")
	severity := c.Query("severity")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var imageVulnerabilities []vul.ImageVulnerability
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.ImageVulnerability{})

	// 添加搜索条件
	if cveID != "" {
		query = query.Where("cve_id LIKE ?", "%"+cveID+"%")
	}
	if vulnName != "" {
		query = query.Where("vuln_name LIKE ?", "%"+vulnName+"%")
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	// 添加时间范围查询条件
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&imageVulnerabilities)
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
		"data": imageVulnerabilities,
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

// UpdateImageVulnerability 更新镜像漏洞信息
func (h *ImagesVulViewCountHandler) UpdateImageVulnerability(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var imageVulnerability vul.ImageVulnerability

	if err := c.ShouldBindJSON(&imageVulnerability); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置ID为URL路径参数中的ID值
	imageVulnerability.ID = id

	result := h.DB.Model(&vul.ImageVulnerability{}).Where("id = ?", id).Updates(&imageVulnerability)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的完整记录
	var updatedImageVuln vul.ImageVulnerability
	queryResult := h.DB.Where("id = ?", id).First(&updatedImageVuln)
	if queryResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询更新后的数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    updatedImageVuln,
	})
}

// DeleteImageVulnerability 删除镜像漏洞信息
func (h *ImagesVulViewCountHandler) DeleteImageVulnerability(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&vul.ImageVulnerability{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
