package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/container"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/models/vul"
)

// VulnRepository 漏洞数据仓库
type VulnRepository struct{}

// NewVulnRepository 创建漏洞仓库实例
func NewVulnRepository() *VulnRepository {
	return &VulnRepository{}
}

// getDB 获取数据库连接
func (r *VulnRepository) getDB() *gorm.DB {
	return db.GetDB()
}

// =========================================================
// 漏洞信息 (vuln_info)
// =========================================================

// CreateOrUpdateVulnInfo 插入或更新漏洞信息（按 cve_id 去重）
func (r *VulnRepository) CreateOrUpdateVulnInfo(ctx context.Context, info *vul.VulnInfo) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cve_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"vuln_name", "severity", "cvss_score",
			"description", "fix_suggestion", "reference_urls", "updated_at",
		}),
	}).Create(info).Error

	if err != nil {
		log.Errorf("[VulnRepository] 漏洞信息写入失败: %v", err)
	}
	return err
}

// BatchCreateOrUpdateVulnInfos 批量插入或更新漏洞信息
func (r *VulnRepository) BatchCreateOrUpdateVulnInfos(ctx context.Context, infos []*vul.VulnInfo) error {
	if len(infos) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "cve_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"vuln_name", "severity", "cvss_score",
			"description", "fix_suggestion", "reference_urls", "updated_at",
		}),
	}).Create(&infos).Error

	if err != nil {
		log.Errorf("[VulnRepository] 批量漏洞信息写入失败: %v", err)
	}
	return err
}

// GetVulnInfoByCveID 根据 CVE ID 查询漏洞信息
func (r *VulnRepository) GetVulnInfoByCveID(ctx context.Context, cveID string) (*vul.VulnInfo, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var info vul.VulnInfo
	err := database.WithContext(ctx).Where("cve_id = ?", cveID).First(&info).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

// GetVulnInfosByCveIDs 批量根据 CVE ID 查询漏洞信息
func (r *VulnRepository) GetVulnInfosByCveIDs(ctx context.Context, cveIDs []string) (map[string]*vul.VulnInfo, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var infos []vul.VulnInfo
	err := database.WithContext(ctx).Where("cve_id IN ?", cveIDs).Find(&infos).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]*vul.VulnInfo, len(infos))
	for i := range infos {
		result[infos[i].CveID] = &infos[i]
	}
	return result, nil
}

// =========================================================
// 主机漏洞扫描 (host_vuln_scan_task / host_vuln_detail)
// =========================================================

// CreateHostVulnScanTask 创建主机漏洞扫描任务记录
func (r *VulnRepository) CreateHostVulnScanTask(ctx context.Context, task *vul.HostVulnScanTask) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(task).Error
	if err != nil {
		log.Errorf("[VulnRepository] 主机漏洞扫描任务写入失败: %v", err)
	}
	return err
}

// UpdateHostVulnScanTask 更新主机漏洞扫描任务记录
func (r *VulnRepository) UpdateHostVulnScanTask(ctx context.Context, task *vul.HostVulnScanTask) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Save(task).Error
	if err != nil {
		log.Errorf("[VulnRepository] 主机漏洞扫描任务更新失败: %v", err)
	}
	return err
}

// BatchCreateHostVulnDetails 批量创建主机漏洞关联记录
func (r *VulnRepository) BatchCreateHostVulnDetails(ctx context.Context, details []*vul.HostVulnDetail) error {
	if len(details) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	// 分批写入，每批 500 条
	batchSize := 500
	for i := 0; i < len(details); i += batchSize {
		end := i + batchSize
		if end > len(details) {
			end = len(details)
		}
		batch := details[i:end]
		if err := database.WithContext(ctx).Create(&batch).Error; err != nil {
			log.Errorf("[VulnRepository] 主机漏洞关联批量写入失败: %v", err)
			return err
		}
	}
	return nil
}

// DeleteHostVulnDetailsByAgent 删除指定 Agent 的主机漏洞关联记录（重新匹配前清理旧数据）
func (r *VulnRepository) DeleteHostVulnDetailsByAgent(ctx context.Context, agentID string) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Delete(&vul.HostVulnDetail{}).Error

	if err != nil {
		log.Errorf("[VulnRepository] 删除主机漏洞关联失败 (agent=%s): %v", agentID, err)
	}
	return err
}

// =========================================================
// 镜像漏洞扫描 (image_vuln_scan_task / image_vuln_detail)
// =========================================================

// CreateImageVulnScanTask 创建镜像漏洞扫描任务记录
func (r *VulnRepository) CreateImageVulnScanTask(ctx context.Context, task *vul.ImageVulnScanTask) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Create(task).Error
	if err != nil {
		log.Errorf("[VulnRepository] 镜像漏洞扫描任务写入失败: %v", err)
	}
	return err
}

// UpdateImageVulnScanTask 更新镜像漏洞扫描任务记录
func (r *VulnRepository) UpdateImageVulnScanTask(ctx context.Context, task *vul.ImageVulnScanTask) error {
	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	err := database.WithContext(ctx).Save(task).Error
	if err != nil {
		log.Errorf("[VulnRepository] 镜像漏洞扫描任务更新失败: %v", err)
	}
	return err
}

// BatchCreateImageVulnDetails 批量创建镜像漏洞关联记录
func (r *VulnRepository) BatchCreateImageVulnDetails(ctx context.Context, details []*vul.ImageVulnDetail) error {
	if len(details) == 0 {
		return nil
	}

	database := r.getDB()
	if database == nil {
		log.Warnf("[VulnRepository] 数据库未初始化，跳过写入")
		return nil
	}

	batchSize := 500
	for i := 0; i < len(details); i += batchSize {
		end := i + batchSize
		if end > len(details) {
			end = len(details)
		}
		batch := details[i:end]
		if err := database.WithContext(ctx).Create(&batch).Error; err != nil {
			log.Errorf("[VulnRepository] 镜像漏洞关联批量写入失败: %v", err)
			return err
		}
	}
	return nil
}

// DeleteImageVulnDetailsByAgentAndImage 删除指定 Agent 和镜像的漏洞关联记录
func (r *VulnRepository) DeleteImageVulnDetailsByAgentAndImage(ctx context.Context, agentID, imageID string) error {
	database := r.getDB()
	if database == nil {
		return nil
	}

	err := database.WithContext(ctx).
		Where("agent_id = ? AND image_id = ?", agentID, imageID).
		Delete(&vul.ImageVulnDetail{}).Error

	if err != nil {
		log.Errorf("[VulnRepository] 删除镜像漏洞关联失败 (agent=%s, image=%s): %v",
			agentID, imageID, err)
	}
	return err
}

// =========================================================
// 查询方法（供匹配引擎和调度器使用）
// =========================================================

// HostInfo 主机基本信息（用于漏洞匹配）
type HostInfo struct {
	AgentID   string
	HostName  string
	HostIP    string
	OsType    string
	OsVersion string
}

// GetAllHosts 获取所有在线主机列表
func (r *VulnRepository) GetAllHosts(ctx context.Context) ([]HostInfo, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var hosts []HostInfo
	err := database.WithContext(ctx).
		Model(&host.Host{}).
		Select("agent_id, host_name, host_ip, os_type, os_version").
		Find(&hosts).Error

	if err != nil {
		log.Errorf("[VulnRepository] 查询主机列表失败: %v", err)
		return nil, err
	}
	return hosts, nil
}

// PackageInfo 软件包信息（用于漏洞匹配）
type PackageInfo struct {
	Name    string
	Version string
	Type    string // dpkg, rpm, apk
}

// GetHostPackages 获取指定主机的软件包列表
func (r *VulnRepository) GetHostPackages(ctx context.Context, agentID string) ([]PackageInfo, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var packages []PackageInfo
	err := database.WithContext(ctx).
		Model(&host.Software{}).
		Select("name, version, type").
		Where("agent_id = ? AND type IN ?", agentID, []string{"dpkg", "rpm", "apk"}).
		Find(&packages).Error

	if err != nil {
		log.Errorf("[VulnRepository] 查询主机软件包失败 (agent=%s): %v", agentID, err)
		return nil, err
	}
	return packages, nil
}

// ImageInfo 镜像基本信息（用于漏洞匹配）
type ImageInfo struct {
	AgentID   string
	ImageID   string
	ImageName string
}

// GetAllImages 获取所有镜像列表
func (r *VulnRepository) GetAllImages(ctx context.Context) ([]ImageInfo, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var images []ImageInfo
	err := database.WithContext(ctx).
		Model(&container.Image{}).
		Select("agent_id, image_id, image_name").
		Find(&images).Error

	if err != nil {
		log.Errorf("[VulnRepository] 查询镜像列表失败: %v", err)
		return nil, err
	}
	return images, nil
}

// ImagePackageInfo 镜像软件包信息（用于漏洞匹配）
type ImagePackageInfo struct {
	ImageID   string
	OsVersion string
	Packages  []PackageInfo
}

// GetImagePackages 获取指定镜像的软件包列表
func (r *VulnRepository) GetImagePackages(ctx context.Context, agentID, imageID string) ([]PackageInfo, error) {
	database := r.getDB()
	if database == nil {
		return nil, nil
	}

	var packages []PackageInfo
	err := database.WithContext(ctx).
		Model(&container.ImagePackage{}).
		Select("package_name as name, package_version as version, package_type as type").
		Where("agent_id = ? AND image_id = ?", agentID, imageID).
		Find(&packages).Error

	if err != nil {
		log.Errorf("[VulnRepository] 查询镜像软件包失败 (agent=%s, image=%s): %v", agentID, imageID, err)
		return nil, err
	}
	return packages, nil
}

// GetImageOSVersion 获取镜像内的 OS 版本（从镜像软件包记录中取）
func (r *VulnRepository) GetImageOSVersion(ctx context.Context, agentID, imageID string) (string, error) {
	database := r.getDB()
	if database == nil {
		return "", nil
	}

	var osVersion string
	err := database.WithContext(ctx).
		Model(&container.ImagePackage{}).
		Select("os_version").
		Where("agent_id = ? AND image_id = ? AND os_version != ''", agentID, imageID).
		Limit(1).
		Scan(&osVersion).Error

	if err != nil {
		return "", err
	}
	return osVersion, nil
}

// NormalizeOSVersion 标准化 OS 版本为 Trivy DB 的 source 格式
// 输入: os_type="linux", os_version="Ubuntu 22.04.3 LTS"
// 输出: "ubuntu 22.04"
func NormalizeOSVersion(osType, osVersion string) string {
	if osType == "" || osVersion == "" {
		return ""
	}

	lower := strings.ToLower(osVersion)
	lowerType := strings.ToLower(osType)

	// Debian: "Debian GNU/Linux 12 (bookworm)" -> "debian 12"
	if strings.Contains(lower, "debian") || lowerType == "debian" {
		parts := strings.Fields(osVersion)
		for _, p := range parts {
			if p[0] >= '0' && p[0] <= '9' {
				// 取主版本号
				ver := strings.Split(p, ".")[0]
				return "debian " + ver
			}
		}
		return "debian"
	}

	// Ubuntu: "Ubuntu 22.04.3 LTS" or osType="ubuntu", osVersion="22.04" -> "ubuntu 22.04"
	if strings.Contains(lower, "ubuntu") || lowerType == "ubuntu" {
		parts := strings.Fields(osVersion)
		for _, p := range parts {
			if p[0] >= '0' && p[0] <= '9' {
				// 取 major.minor
				dotParts := strings.Split(p, ".")
				if len(dotParts) >= 2 {
					return "ubuntu " + dotParts[0] + "." + dotParts[1]
				}
				return "ubuntu " + p
			}
		}
		return "ubuntu"
	}

	// CentOS: "CentOS Linux 7.9.2009" -> "centos 7"
	if strings.Contains(lower, "centos") || lowerType == "centos" {
		parts := strings.Fields(osVersion)
		for _, p := range parts {
			if p[0] >= '0' && p[0] <= '9' {
				ver := strings.Split(p, ".")[0]
				return "centos " + ver
			}
		}
		return "centos"
	}

	// RHEL: "Red Hat Enterprise Linux 8.6" -> "redhat 8"
	if strings.Contains(lower, "red hat") || strings.Contains(lower, "rhel") || lowerType == "redhat" || lowerType == "rhel" {
		parts := strings.Fields(osVersion)
		for _, p := range parts {
			if p[0] >= '0' && p[0] <= '9' {
				ver := strings.Split(p, ".")[0]
				return "redhat " + ver
			}
		}
		return "redhat"
	}

	// Alpine: "Alpine Linux 3.18.4" -> "alpine 3.18"
	if strings.Contains(lower, "alpine") || lowerType == "alpine" {
		parts := strings.Fields(osVersion)
		for _, p := range parts {
			if p[0] >= '0' && p[0] <= '9' {
				dotParts := strings.Split(p, ".")
				if len(dotParts) >= 2 {
					return "alpine " + dotParts[0] + "." + dotParts[1]
				}
				return "alpine " + p
			}
		}
		return "alpine"
	}

	// Amazon Linux: "Amazon Linux 2023" -> "amazon linux 2023"
	if strings.Contains(lower, "amazon") || lowerType == "amazon" {
		parts := strings.Fields(osVersion)
		for _, p := range parts {
			if p[0] >= '0' && p[0] <= '9' {
				return "amazon linux " + p
			}
		}
		return "amazon linux"
	}

	return lower
}
