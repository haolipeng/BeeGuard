package vul

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/haolipeng/BeeGuard/server/internal/models/vul"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HostVulnScanTaskHandler 主机漏洞扫描任务处理器结构体，用于处理与主机漏洞扫描任务相关的HTTP请求
type HostVulnScanTaskHandler struct {
	// DB 数据库连接实例
	DB *gorm.DB
}

// CreateHostVulnScanTask 创建主机漏洞扫描任务
func (h *HostVulnScanTaskHandler) CreateHostVulnScanTask(c *gin.Context) {
	// 接收数据
	var task vul.HostVulnScanTask
	// 验证并绑定请求中的JSON数据到task结构体
	if err := c.ShouldBindJSON(&task); err != nil {
		// 数据验证失败时返回错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打印接收的数据
	fmt.Println(task)

	// 执行数据库操作
	result := h.DB.Create(&task)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 返回成功响应，包含创建的记录信息
	c.JSON(http.StatusOK, gin.H{
		"message": "创建成功",
		"data":    task,
	})
}

// GetHostVulnScanTask 获取单个主机漏洞扫描任务
func (h *HostVulnScanTaskHandler) GetHostVulnScanTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var task vul.HostVulnScanTask

	result := h.DB.Where("id = ?", id).First(&task)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "主机漏洞扫描任务不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// ListHostVulnScanTasks 获取主机漏洞扫描任务列表（支持搜索查询）
func (h *HostVulnScanTaskHandler) ListHostVulnScanTasks(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	agentID := c.Query("agent_id")
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var tasks []vul.HostVulnScanTask
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.HostVulnScanTask{})

	// 添加搜索条件
	if agentID != "" {
		query = query.Where("agent_id LIKE ?", "%"+agentID+"%")
	}
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
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
	result = query.Order("scan_time DESC").Limit(limit).Offset(offset).Find(&tasks)
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
		"data": tasks,
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

// UpdateHostVulnScanTask 更新主机漏洞扫描任务
func (h *HostVulnScanTaskHandler) UpdateHostVulnScanTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var task vul.HostVulnScanTask

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置ID为URL路径参数中的ID值
	task.ID = id

	result := h.DB.Model(&vul.HostVulnScanTask{}).Where("id = ?", id).Updates(&task)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	// 查询更新后的完整记录
	var updatedTask vul.HostVulnScanTask
	queryResult := h.DB.Where("id = ?", id).First(&updatedTask)
	if queryResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询更新后的数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"data":    updatedTask,
	})
}

// DeleteHostVulnScanTask 删除主机漏洞扫描任务
func (h *HostVulnScanTaskHandler) DeleteHostVulnScanTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	result := h.DB.Where("id = ?", id).Delete(&vul.HostVulnScanTask{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ListVulnCountHosts 获取漏洞统计主机列表（基于视图v_vuln_count_hosts）
func (h *HostVulnScanTaskHandler) ListVulnCountHosts(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 获取搜索条件参数
	hostName := c.Query("host_name")
	hostIP := c.Query("host_ip")
	minVulns := c.Query("min_vulns")
	maxVulns := c.Query("max_vulns")
	severity := c.Query("severity") // critical, high, medium, low

	// 确保页码至少为1
	if page < 1 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	var vulnCountHosts []vul.VulnCountHost
	var total int64

	// 构建查询条件
	query := h.DB.Model(&vul.VulnCountHost{})

	// 添加搜索条件
	if hostName != "" {
		query = query.Where("host_name LIKE ?", "%"+hostName+"%")
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
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
	result = query.Order("total_vulns DESC").Limit(limit).Offset(offset).Find(&vulnCountHosts)
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
		"data": vulnCountHosts,
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

// GetVulnCountHostByIP 根据主机IP获取漏洞统计详情
func (h *HostVulnScanTaskHandler) GetVulnCountHostByIP(c *gin.Context) {
	hostIP := c.Param("host_ip")
	if hostIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "主机IP不能为空"})
		return
	}

	var vulnCountHost vul.VulnCountHost

	result := h.DB.Where("host_ip = ?", hostIP).First(&vulnCountHost)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "主机漏洞统计信息不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": vulnCountHost})
}
