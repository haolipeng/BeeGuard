package vul

import (
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/vul"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImageViewCountHandler 镜像漏洞统计处理器结构体，用于处理与镜像漏洞统计相关的HTTP请求
type ImageViewCountHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// ListVulnCountImages 获取漏洞统计镜像列表（基于视图v_vuln_count_images）
func (h *ImageViewCountHandler) ListVulnCountImages(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	imageName := c.Query("image_name")
	imageID := c.Query("image_id")
	minVulns := c.Query("min_vulns")
	maxVulns := c.Query("max_vulns")
	severity := c.Query("severity") // critical, high, medium, low

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var vulnCountImages []vul.VulnCountImage
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.VulnCountImage{})

	// 添加搜索条件
	if imageName != "" {
		query = query.Where("image_name LIKE ?", "%"+imageName+"%")
	}
	if imageID != "" {
		query = query.Where("image_id LIKE ?", "%"+imageID+"%")
	}
	if minVulns != "" {
		if min, err := strconv.Atoi(minVulns); err == nil {
			query = query.Where("total_vulns >= ?", min)
		}
	}
	if maxVulns != "" {
		if max, err := strconv.Atoi(maxVulns); err == nil {
			query = query.Where("total_vulns <= ?", max)
		}
	}
	// 根据严重级别筛选
	if severity != "" {
		switch severity {
		case "critical":
			query = query.Where("critical_vulns > 0")
		case "high":
			query = query.Where("high_vulns > 0")
		case "medium":
			query = query.Where("medium_vulns > 0")
		case "low":
			query = query.Where("low_vulns > 0")
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按总漏洞数倒序排列
	result = query.Order("total_vulns DESC").Limit(limit).Offset(offset).Find(&vulnCountImages)
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
		"data": vulnCountImages,
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

// GetVulnCountImageByID 根据镜像ID获取漏洞统计详情
func (h *ImageViewCountHandler) GetVulnCountImageByID(c *gin.Context) {
	imageID := c.Param("image_id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "镜像ID不能为空"})
		return
	}

	var vulnCountImage vul.VulnCountImage

	result := h.DB.Where("image_id = ?", imageID).First(&vulnCountImage)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "镜像漏洞统计信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vulnCountImage})
}

// ListImageVulnDetails 获取镜像漏洞详情列表（支持搜索查询）
func (h *ImageViewCountHandler) ListImageVulnDetails(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	imageID := c.Query("image_id")
	imageName := c.Query("image_name")
	cveID := c.Query("cve_id")
	packageName := c.Query("package_name")
	severity := c.Query("severity")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var imageVulnDetails []vul.ImageVulnDetail
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.ImageVulnDetail{})

	// 添加搜索条件
	if imageID != "" {
		query = query.Where("image_id LIKE ?", "%"+imageID+"%")
	}
	if imageName != "" {
		query = query.Where("image_name LIKE ?", "%"+imageName+"%")
	}
	if cveID != "" {
		query = query.Where("cve_id LIKE ?", "%"+cveID+"%")
	}
	if packageName != "" {
		query = query.Where("package_name LIKE ?", "%"+packageName+"%")
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if startTime != "" {
		query = query.Where("scan_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("scan_time <= ?", endTime)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按扫描时间倒序排列
	result = query.Order("scan_time DESC").Limit(limit).Offset(offset).Find(&imageVulnDetails)
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
		"data": imageVulnDetails,
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

// GetImageVulnDetail 获取单个镜像漏洞详情
func (h *ImageViewCountHandler) GetImageVulnDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var imageVulnDetail vul.ImageVulnDetail

	result := h.DB.Where("id = ?", id).First(&imageVulnDetail)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "镜像漏洞详情不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": imageVulnDetail})
}
