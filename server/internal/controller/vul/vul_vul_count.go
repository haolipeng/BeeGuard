package vul

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/vul"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VulnInfoHandler 漏洞信息处理器结构体，用于处理与漏洞信息相关的HTTP请求
type VulnInfoHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateVulnInfo 创建漏洞信息
func (h *VulnInfoHandler) CreateVulnInfo(c *gin.Context) {
	// 接收数据
	var vulnerability vul.Vulnerability
	// 验证并绑定请求中的JSON数据到vulnerability结构体
	if err := c.ShouldBindJSON(&vulnerability); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行数据库操作
	result := h.DB.Create(&vulnerability)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    vulnerability,
	})
}

// GetVulnInfo 获取单个漏洞信息
func (h *VulnInfoHandler) GetVulnInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var vulnerability vul.Vulnerability

	result := h.DB.Where("id = ?", id).First(&vulnerability)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "漏洞信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vulnerability})
}

// ListVulnInfos 获取漏洞信息列表（支持搜索查询）
func (h *VulnInfoHandler) ListVulnInfos(c *gin.Context) {
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

	var vulnerabilities []vul.Vulnerability
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.Vulnerability{})

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
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&vulnerabilities)
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
		"data": vulnerabilities,
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

// UpdateVulnInfo 更新漏洞信息
func (h *VulnInfoHandler) UpdateVulnInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var vulnerability vul.Vulnerability

	if err := c.ShouldBindJSON(&vulnerability); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置ID为URL路径参数中的ID值
	vulnerability.ID = id

	result := h.DB.Model(&vul.Vulnerability{}).Where("id = ?", id).Updates(&vulnerability)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的完整记录
	var updatedVuln vul.Vulnerability
	queryResult := h.DB.Where("id = ?", id).First(&updatedVuln)
	if queryResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询更新后的数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    updatedVuln,
	})
}

// DeleteVulnInfo 删除漏洞信息
func (h *VulnInfoHandler) DeleteVulnInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&vul.Vulnerability{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListVulnWithHosts 获取漏洞主机统计列表（基于视图）
func (h *VulnInfoHandler) ListVulnWithHosts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	cveID := c.Query("cve_id")
	vulnName := c.Query("vuln_name")
	severity := c.Query("severity")
	minHostCount := c.Query("min_host_count")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var vulnWithHosts []vul.VulnWithHosts
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.VulnWithHosts{})

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
	if minHostCount != "" {
		if count, err := strconv.Atoi(minHostCount); err == nil {
			query = query.Where("affected_host_count >= ?", count)
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

	// 分页查询数据，按影响主机数量倒序排列
	result = query.Order("affected_host_count DESC, last_scan_time DESC").Limit(limit).Offset(offset).Find(&vulnWithHosts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 解析受影响主机详情
	for i := range vulnWithHosts {
		if len(vulnWithHosts[i].AffectedHosts) > 0 {
			var affectedHosts []vul.AffectedHost
			if err := json.Unmarshal(vulnWithHosts[i].AffectedHosts, &affectedHosts); err == nil {
				// 只保留前5个主机信息以减少响应大小
				if len(affectedHosts) > 5 {
					affectedHosts = affectedHosts[:5]
				}
				// 重新序列化截取后的数据
				if jsonData, err := json.Marshal(affectedHosts); err == nil {
					vulnWithHosts[i].AffectedHosts = jsonData
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
		"data": vulnWithHosts,
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

// GetVulnWithHosts 获取单个漏洞主机统计详情
func (h *VulnInfoHandler) GetVulnWithHosts(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}
	//打印
	//log.Println("GetVulnWithHosts called with ID:", id)
	//os.Exit(0)

	var vulnWithHosts vul.VulnWithHosts

	result := h.DB.Where("vuln_id = ?", id).First(&vulnWithHosts)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "漏洞统计信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 解析受影响主机详情
	if len(vulnWithHosts.AffectedHosts) > 0 {
		var affectedHosts []vul.AffectedHost
		if err := json.Unmarshal(vulnWithHosts.AffectedHosts, &affectedHosts); err == nil {
			// 重新序列化完整的数据
			if jsonData, err := json.Marshal(affectedHosts); err == nil {
				vulnWithHosts.AffectedHosts = jsonData
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": vulnWithHosts})
}
