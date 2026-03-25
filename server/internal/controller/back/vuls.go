package back

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/vul"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VulnerabilityInfoHandler 漏洞检测规则处理器结构体
type VulnerabilityInfoHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateVulnerabilityInfo 创建漏洞检测规则
func (h *VulnerabilityInfoHandler) CreateVulnerabilityInfo(c *gin.Context) {
	// 接收数据
	var vuln vul.VulnInfo
	// 验证并绑定请求中的JSON数据到vuln结构体
	if err := c.ShouldBindJSON(&vuln); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//打印接收到的数据（格式化输出）
	fmt.Printf("=== 接收到的漏洞数据 ===\n")
	fmt.Printf("ID: %d\n", vuln.ID)
	//os.Exit(0)

	// 执行数据库操作
	result := h.DB.Create(&vuln)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    vuln,
	})
}

// GetVulnerabilityInfo 获取单个漏洞检测规则
func (h *VulnerabilityInfoHandler) GetVulnerabilityInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var vuln vul.VulnInfo

	result := h.DB.Where("id = ?", id).First(&vuln)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "漏洞规则不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vuln})
}

// ListVulnerabilityInfos 获取漏洞检测规则列表（支持搜索查询）
func (h *VulnerabilityInfoHandler) ListVulnerabilityInfos(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	cveID := c.Query("cve_id")
	vulnName := c.Query("vuln_name")
	severity := c.Query("severity")
	statusStr := c.Query("status")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var vulns []vul.VulnInfo
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.VulnInfo{})
	//打印接收到的数据（格式化输出）

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
	if statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", status)
		}
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按创建时间倒序排列
	result = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&vulns)
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
		"data": vulns,
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

// UpdateVulnerabilityInfo 更新漏洞检测规则
func (h *VulnerabilityInfoHandler) UpdateVulnerabilityInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var vuln vul.VulnInfo
	// 检查漏洞规则是否存在
	result := h.DB.Where("id = ?", id).First(&vuln)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "漏洞规则不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 接收更新数据
	var updateData vul.VulnInfo
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 执行更新
	result = h.DB.Model(&vuln).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的数据
	var updatedVuln vul.VulnInfo
	h.DB.Where("id = ?", id).First(&updatedVuln)

	c.JSON(http.StatusOK, gin.H{"message": "更新成功", "data": updatedVuln})
}

// DeleteVulnerabilityInfo 删除漏洞检测规则
func (h *VulnerabilityInfoHandler) DeleteVulnerabilityInfo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&vul.VulnInfo{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
