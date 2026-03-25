package vul

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/vul"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HostVulnDetailHandler 主机漏洞详情处理器结构体，用于处理与主机漏洞详情相关的HTTP请求
type HostVulnDetailHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateHostVulnDetail 创建主机漏洞详情
func (h *HostVulnDetailHandler) CreateHostVulnDetail(c *gin.Context) {
	// 接收数据
	var hostVulnDetail vul.HostVulnDetail
	// 验证并绑定请求中的JSON数据到hostVulnDetail结构体
	if err := c.ShouldBindJSON(&hostVulnDetail); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打印接收的数据
	fmt.Println(hostVulnDetail)

	// 执行数据库操作
	result := h.DB.Create(&hostVulnDetail)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    hostVulnDetail,
	})
}

// GetHostVulnDetail 获取单个主机漏洞详情
func (h *HostVulnDetailHandler) GetHostVulnDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var hostVulnDetail vul.HostVulnDetail

	result := h.DB.Where("id = ?", id).First(&hostVulnDetail)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "主机漏洞详情不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hostVulnDetail})
}

// ListHostVulnDetails 获取主机漏洞详情列表（支持搜索查询）
func (h *HostVulnDetailHandler) ListHostVulnDetails(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	Vuln := c.Query("vuln_id")
	cveID := c.Query("cve_id")
	hostIp := c.Query("host_ip")
	status := c.Query("status")
	hostName := c.Query("host_name")
	severity := c.Query("severity")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var hostVulnDetails []vul.HostVulnDetail
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.HostVulnDetail{})

	// 添加搜索条件
	if agentID != "" {
		query = query.Where("agent_id LIKE ?", "%"+agentID+"%")
	}
	if Vuln != "" {
		//query = query.Where("vuln_id LIKE ?", "%"+Vuln+"%")
		query = query.Where("vuln_id = ?", Vuln)
	}
	if cveID != "" {
		query = query.Where("cve_id LIKE ?", "%"+cveID+"%")
	}
	if hostIp != "" {
		query = query.Where("host_ip = ?", hostIp)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// 先获取总记录数
	result := query.Count(&total)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 分页查询数据，按扫描时间倒序排列
	result = query.Order("scan_time DESC").Limit(limit).Offset(offset).Find(&hostVulnDetails)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 计算总页数
	totalPages := int(total / int64(limit))
	if total%int64(limit) > 0 {
		totalPages++
	}

	//打印 数据
	//fmt.Println(hostVulnDetails)
	//os.Exit(0)

	// 返回分页结果
	c.JSON(http.StatusOK, gin.H{
		"data": hostVulnDetails,
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

// UpdateHostVulnDetail 更新主机漏洞详情信息
func (h *HostVulnDetailHandler) UpdateHostVulnDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var hostVulnDetail vul.HostVulnDetail

	if err := c.ShouldBindJSON(&hostVulnDetail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置ID为URL路径参数中的ID值
	hostVulnDetail.ID = id

	result := h.DB.Model(&vul.HostVulnDetail{}).Where("id = ?", id).Updates(&hostVulnDetail)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的完整记录
	var updatedHostVulnDetail vul.HostVulnDetail
	queryResult := h.DB.Where("id = ?", id).First(&updatedHostVulnDetail)
	if queryResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询更新后的数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    updatedHostVulnDetail,
	})
}

// DeleteHostVulnDetail 删除主机漏洞详情
func (h *HostVulnDetailHandler) DeleteHostVulnDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Unscoped().Where("id = ?", id).Delete(&vul.HostVulnDetail{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
