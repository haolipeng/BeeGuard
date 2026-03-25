package vuln

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aquasecurity/trivy-db/pkg/types"
	debver "github.com/knqyf263/go-deb-version"
	rpmver "github.com/knqyf263/go-rpm-version"

	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
	"github.com/haolipeng/BeeGuard/server/internal/models/vul"
)

// Matcher 漏洞匹配引擎，负责将软件包版本与 Trivy 漏洞数据库进行匹配
type Matcher struct {
	dbMgr    *DBManager
	vulnRepo *repository.VulnRepository
}

// NewMatcher 创建漏洞匹配引擎
func NewMatcher(dbMgr *DBManager, vulnRepo *repository.VulnRepository) *Matcher {
	return &Matcher{
		dbMgr:    dbMgr,
		vulnRepo: vulnRepo,
	}
}

// MatchResult 单个包的匹配结果
type MatchResult struct {
	CveID            string
	VulnName         string
	Severity         string
	CvssScore        float64
	PackageName      string
	InstalledVersion string
	FixedVersion     string
	Description      string
	References       []string
}

// MatchHostVulns 对指定主机执行漏洞匹配
// 流程：读取主机软件包 → 逐包查询 Trivy DB → 写入匹配结果
// 返回匹配到的漏洞数量
func (m *Matcher) MatchHostVulns(ctx context.Context, host repository.HostInfo) (int, error) {
	if !m.dbMgr.IsReady() {
		return 0, fmt.Errorf("漏洞数据库未就绪")
	}

	// 标准化 OS 版本为 Trivy DB source 格式
	source := repository.NormalizeOSVersion(host.OsType, host.OsVersion)
	if source == "" {
		log.Warnf("[Matcher] 主机 OS 版本信息缺失，跳过匹配 (agent=%s)", host.AgentID)
		return 0, nil
	}

	// 获取主机软件包列表
	packages, err := m.vulnRepo.GetHostPackages(ctx, host.AgentID)
	if err != nil {
		return 0, fmt.Errorf("获取主机软件包失败 (agent=%s): %w", host.AgentID, err)
	}
	if len(packages) == 0 {
		// 降级为 Debug：避免每次主机匹配都刷屏 Info
		log.Debugf("[Matcher] 主机无软件包数据，跳过匹配 (agent=%s)", host.AgentID)
		return 0, nil
	}

	// 降级为 Debug：避免每台主机匹配刷屏 Info
	log.Debugf("[Matcher] 开始匹配主机漏洞: agent=%s, host=%s, source=%s, packages=%d",
		host.AgentID, host.HostName, source, len(packages))

	// 执行匹配
	results := m.matchPackages(ctx, source, packages)
	if len(results) == 0 {
		// 降级为 Debug：主机“未发现漏洞”在大规模资产下极高频
		log.Debugf("[Matcher] 主机未发现漏洞: agent=%s", host.AgentID)
		// 仍然写入扫描记录（匹配漏洞数为 0）
		return 0, m.saveHostScanResult(ctx, host, nil, len(packages))
	}

	// 降级为 Debug：主机“发现漏洞”同样在大规模资产下非常高频
	log.Debugf("[Matcher] 主机发现 %d 个漏洞: agent=%s", len(results), host.AgentID)

	// 写入结果
	return len(results), m.saveHostScanResult(ctx, host, results, len(packages))
}

// MatchImageVulns 对指定镜像执行漏洞匹配
// 返回匹配到的漏洞数量
func (m *Matcher) MatchImageVulns(ctx context.Context, image repository.ImageInfo, source string, packages []repository.PackageInfo) (int, error) {
	if !m.dbMgr.IsReady() {
		return 0, fmt.Errorf("漏洞数据库未就绪")
	}

	if source == "" || len(packages) == 0 {
		// 降级为 Debug：避免每个镜像匹配刷屏 Info
		log.Debugf("[Matcher] 镜像包信息不足，跳过匹配 (agent=%s, image=%s)",
			image.AgentID, image.ImageID)
		return 0, nil
	}

	// 降级为 Debug：避免每个镜像匹配刷屏 Info
	log.Debugf("[Matcher] 开始匹配镜像漏洞: agent=%s, image=%s, source=%s, packages=%d",
		image.AgentID, image.ImageName, source, len(packages))

	results := m.matchPackages(ctx, source, packages)
	if len(results) == 0 {
		// 降级为 Debug：镜像“未发现漏洞”在大规模资产下极高频
		log.Debugf("[Matcher] 镜像未发现漏洞: image=%s", image.ImageName)
		return 0, m.saveImageScanResult(ctx, image, nil, len(packages))
	}

	// 降级为 Debug：镜像“发现漏洞”同样在大规模资产下非常高频
	log.Debugf("[Matcher] 镜像发现 %d 个漏洞: image=%s", len(results), image.ImageName)

	return len(results), m.saveImageScanResult(ctx, image, results, len(packages))
}

// matchPackages 对一组软件包执行漏洞匹配，返回所有匹配结果
// 支持通过 ctx 取消，收到取消信号时立即返回已匹配的结果
func (m *Matcher) matchPackages(ctx context.Context, source string, packages []repository.PackageInfo) []MatchResult {
	var results []MatchResult
	seen := make(map[string]bool) // 采用 cve_id+pkg 去重

	for _, pkg := range packages {
		// 检查 context 是否已取消（关闭信号），及时退出
		if ctx.Err() != nil {
			// 降级为 Debug：取消场景一般不需要高频 Info
			log.Debugf("[Matcher] 匹配任务被取消，已处理 %d/%d 个包", len(seen), len(packages))
			return results
		}

		if pkg.Name == "" || pkg.Version == "" {
			continue
		}

		advisories, err := m.dbMgr.GetAdvisories(source, pkg.Name)
		if err != nil {
			// 查询失败不中断，记录警告继续下一个包
			log.Debugf("[Matcher] 查询 advisory 失败 (source=%s, pkg=%s): %v",
				source, pkg.Name, err)
			continue
		}

		for _, adv := range advisories {
			if adv.VulnerabilityID == "" {
				continue
			}

			key := adv.VulnerabilityID + "|" + pkg.Name
			if seen[key] {
				continue
			}

			// 如果状态为 not_affected，则跳过
			if adv.Status == types.StatusNotAffected {
				continue
			}

			// 版本比较：如果有 FixedVersion，检查已安装版本是否小于修复版本
			if adv.FixedVersion != "" {
				affected, err := versionLessThan(pkg.Type, pkg.Version, adv.FixedVersion)
				if err != nil {
					log.Debugf("[Matcher] 版本比较失败 (pkg=%s, installed=%s, fixed=%s, type=%s): %v",
						pkg.Name, pkg.Version, adv.FixedVersion, pkg.Type, err)
					// 版本比较失败时保守处理，视为受影响
				} else if !affected {
					// 已安装版本 >= 修复版本，不受影响
					continue
				}
			}

			// 获取漏洞详情
			vuln, err := m.dbMgr.GetVulnerability(adv.VulnerabilityID)
			if err != nil {
				log.Debugf("[Matcher] 获取漏洞详情失败 (cve=%s): %v",
					adv.VulnerabilityID, err)
				// 仍然记录，使用 advisory 中的信息
			}

			severity := determineSeverity(adv, vuln)
			cvssScore := determineCVSSScore(vuln)

			result := MatchResult{
				CveID:            adv.VulnerabilityID,
				VulnName:         vulnTitle(vuln, adv.VulnerabilityID),
				Severity:         severity,
				CvssScore:        cvssScore,
				PackageName:      pkg.Name,
				InstalledVersion: pkg.Version,
				FixedVersion:     adv.FixedVersion,
				Description:      vuln.Description,
				References:       vuln.References,
			}

			results = append(results, result)
			seen[key] = true
		}
	}

	return results
}

// versionLessThan 根据包类型比较已安装版本是否小于修复版本
func versionLessThan(pkgType, installed, fixed string) (bool, error) {
	switch pkgType {
	case "dpkg":
		v1, err := debver.NewVersion(installed)
		if err != nil {
			return false, fmt.Errorf("解析 dpkg 已安装版本 %q: %w", installed, err)
		}
		v2, err := debver.NewVersion(fixed)
		if err != nil {
			return false, fmt.Errorf("解析 dpkg 修复版本 %q: %w", fixed, err)
		}
		return v1.LessThan(v2), nil
	case "rpm":
		v1 := rpmver.NewVersion(installed)
		v2 := rpmver.NewVersion(fixed)
		return v1.LessThan(v2), nil
	case "apk":
		// APK 使用类似 semver 的格式，回退到简单字符串比较
		// 当已安装版本与修复版本完全相同时，视为不受影响
		if installed == fixed {
			return false, nil
		}
		// 对于 APK，使用 dpkg 版本比较作为近似（均支持 epoch 和 ~suffix）
		v1, err := debver.NewVersion(installed)
		if err != nil {
			return false, fmt.Errorf("解析 apk 已安装版本 %q: %w", installed, err)
		}
		v2, err := debver.NewVersion(fixed)
		if err != nil {
			return false, fmt.Errorf("解析 apk 修复版本 %q: %w", fixed, err)
		}
		return v1.LessThan(v2), nil
	default:
		// 未知包类型，保守处理视为受影响
		return true, nil
	}
}

// saveHostScanResult 保存主机漏洞扫描结果到数据库
func (m *Matcher) saveHostScanResult(ctx context.Context, host repository.HostInfo, results []MatchResult, totalPackages int) error {
	now := time.Now()
	nowDT := common.DateTime{Time: now}

	matchedVulns := int32(len(results))
	totalPkgs := int32(totalPackages)

	// 创建扫描任务记录（状态: 进行中）
	task := &vul.HostVulnScanTask{
		AgentID:       host.AgentID,
		HostName:      host.HostName,
		HostIP:        host.HostIP,
		ScanStatus:    vul.ScanStatusRunning,
		ScanTrigger:   vul.ScanTriggerAuto,
		TotalPackages: &totalPkgs,
		ScanTime:      nowDT,
	}
	if err := m.vulnRepo.CreateHostVulnScanTask(ctx, task); err != nil {
		return fmt.Errorf("创建主机扫描任务失败: %w", err)
	}

	if len(results) == 0 {
		// 无漏洞，直接标记成功
		task.ScanStatus = vul.ScanStatusSuccess
		task.MatchedVulns = &matchedVulns
		duration := int32(time.Since(now).Milliseconds())
		task.ScanDuration = &duration
		return m.vulnRepo.UpdateHostVulnScanTask(ctx, task)
	}

	// 先清理旧的漏洞关联数据
	if err := m.vulnRepo.DeleteHostVulnDetailsByAgent(ctx, host.AgentID); err != nil {
		log.Warnf("[Matcher] 清理旧漏洞关联失败 (agent=%s): %v", host.AgentID, err)
	}

	// 写入/更新漏洞信息，并创建关联记录
	var details []*vul.HostVulnDetail
	for _, r := range results {
		// 写入 vuln_info
		vulnInfo := buildVulnInfo(r)
		if err := m.vulnRepo.CreateOrUpdateVulnInfo(ctx, vulnInfo); err != nil {
			log.Warnf("[Matcher] 写入漏洞信息失败 (cve=%s): %v", r.CveID, err)
			continue
		}

		// 构建关联记录
		hostName := host.HostName
		hostIP := host.HostIP
		vulnName := r.VulnName
		severity := r.Severity
		var installedVer *string
		if r.InstalledVersion != "" {
			v := r.InstalledVersion
			installedVer = &v
		}
		var fixedVer *string
		if r.FixedVersion != "" {
			v := r.FixedVersion
			fixedVer = &v
		}
		detail := &vul.HostVulnDetail{
			ScanID:           task.ID,
			AgentID:          host.AgentID,
			VulnID:           vulnInfo.ID,
			CveID:            r.CveID,
			PackageName:      r.PackageName,
			InstalledVersion: installedVer,
			FixedVersion:     fixedVer,
			Status:           vul.VulnStatusUnfixed,
			ScanTime:         nowDT,
			HostName:         &hostName,
			HostIP:           &hostIP,
			VulnName:         &vulnName,
			Severity:         &severity,
			CvssScore:        vulnInfo.CvssScore,
			Description:      vulnInfo.Description,
			FixSuggestion:    vulnInfo.FixSuggestion,
		}
		details = append(details, detail)
	}

	// 批量写入关联记录
	if err := m.vulnRepo.BatchCreateHostVulnDetails(ctx, details); err != nil {
		// 写入失败，标记任务失败
		task.ScanStatus = vul.ScanStatusFailed
		errMsg := fmt.Sprintf("批量写入主机漏洞关联失败: %v", err)
		task.ErrorMessage = &errMsg
		_ = m.vulnRepo.UpdateHostVulnScanTask(ctx, task)
		return fmt.Errorf("%s", errMsg)
	}

	// 标记任务成功
	task.ScanStatus = vul.ScanStatusSuccess
	task.MatchedVulns = &matchedVulns
	duration := int32(time.Since(now).Milliseconds())
	task.ScanDuration = &duration
	return m.vulnRepo.UpdateHostVulnScanTask(ctx, task)
}

// saveImageScanResult 保存镜像漏洞扫描结果到数据库
func (m *Matcher) saveImageScanResult(ctx context.Context, image repository.ImageInfo, results []MatchResult, totalPackages int) error {
	now := time.Now()
	nowDT := common.DateTime{Time: now}

	matchedVulns := int32(len(results))
	totalPkgs := int32(totalPackages)

	// 创建扫描任务记录（状态: 进行中）
	task := &vul.ImageVulnScanTask{
		AgentID:       image.AgentID,
		ImageID:       image.ImageID,
		ImageName:     image.ImageName,
		ScanStatus:    vul.ScanStatusRunning,
		ScanTrigger:   vul.ScanTriggerAuto,
		TotalPackages: &totalPkgs,
		ScanTime:      nowDT,
	}
	if err := m.vulnRepo.CreateImageVulnScanTask(ctx, task); err != nil {
		return fmt.Errorf("创建镜像扫描任务失败: %w", err)
	}

	if len(results) == 0 {
		// 无漏洞，直接标记成功
		task.ScanStatus = vul.ScanStatusSuccess
		task.MatchedVulns = &matchedVulns
		duration := int32(time.Since(now).Milliseconds())
		task.ScanDuration = &duration
		return m.vulnRepo.UpdateImageVulnScanTask(ctx, task)
	}

	// 清理旧数据
	if err := m.vulnRepo.DeleteImageVulnDetailsByAgentAndImage(ctx, image.AgentID, image.ImageID); err != nil {
		log.Warnf("[Matcher] 清理旧镜像漏洞关联失败: %v", err)
	}

	var details []*vul.ImageVulnDetail
	for _, r := range results {
		vulnInfo := buildVulnInfo(r)
		if err := m.vulnRepo.CreateOrUpdateVulnInfo(ctx, vulnInfo); err != nil {
			log.Warnf("[Matcher] 写入漏洞信息失败 (cve=%s): %v", r.CveID, err)
			continue
		}

		var installedVer *string
		if r.InstalledVersion != "" {
			v := r.InstalledVersion
			installedVer = &v
		}
		var fixedVer *string
		if r.FixedVersion != "" {
			v := r.FixedVersion
			fixedVer = &v
		}
		detail := &vul.ImageVulnDetail{
			ScanID:           task.ID,
			AgentID:          image.AgentID,
			ImageID:          image.ImageID,
			VulnID:           vulnInfo.ID,
			CveID:            r.CveID,
			PackageName:      r.PackageName,
			InstalledVersion: installedVer,
			FixedVersion:     fixedVer,
			Status:           vul.VulnStatusUnfixed,
			ScanTime:         nowDT,
			ImageName:        image.ImageName,
			VulnName:         r.VulnName,
			Severity:         r.Severity,
			CVSSScore:        vulnInfo.CvssScore,
			Description:      vulnInfo.Description,
			FixSuggestion:    vulnInfo.FixSuggestion,
		}
		details = append(details, detail)
	}

	if err := m.vulnRepo.BatchCreateImageVulnDetails(ctx, details); err != nil {
		// 写入失败，标记任务失败
		task.ScanStatus = vul.ScanStatusFailed
		errMsg := fmt.Sprintf("批量写入镜像漏洞关联失败: %v", err)
		task.ErrorMessage = &errMsg
		_ = m.vulnRepo.UpdateImageVulnScanTask(ctx, task)
		return fmt.Errorf("%s", errMsg)
	}

	// 标记任务成功
	task.ScanStatus = vul.ScanStatusSuccess
	task.MatchedVulns = &matchedVulns
	duration := int32(time.Since(now).Milliseconds())
	task.ScanDuration = &duration
	return m.vulnRepo.UpdateImageVulnScanTask(ctx, task)
}

// determineSeverity 确定漏洞等级，优先使用 advisory 中的等级，其次从 vulnerability 详情中获取
func determineSeverity(adv types.Advisory, vuln types.Vulnerability) string {
	// 优先使用 advisory 自带的 severity
	if adv.Severity != types.SeverityUnknown {
		return normalizeSeverity(adv.Severity)
	}

	// 从 VendorSeverity 中获取最高等级
	if len(vuln.VendorSeverity) > 0 {
		var maxSev types.Severity
		for _, sev := range vuln.VendorSeverity {
			if sev > maxSev {
				maxSev = sev
			}
		}
		if maxSev != types.SeverityUnknown {
			return normalizeSeverity(maxSev)
		}
	}

	// 从 Severity 字段获取
	if vuln.Severity != "" {
		return strings.ToLower(vuln.Severity)
	}

	return vul.VulnSeverityMedium // 默认中危
}

// determineCVSSScore 从漏洞详情中获取 CVSS 分数，优先 V3
func determineCVSSScore(vuln types.Vulnerability) float64 {
	if len(vuln.CVSS) == 0 {
		return 0
	}

	var maxScore float64
	for _, cvss := range vuln.CVSS {
		if cvss.V3Score > maxScore {
			maxScore = cvss.V3Score
		}
		if cvss.V2Score > maxScore && maxScore == 0 {
			maxScore = cvss.V2Score
		}
	}
	return maxScore
}

// normalizeSeverity 将 Trivy Severity 转换为数据库中的字符串格式
func normalizeSeverity(sev types.Severity) string {
	switch sev {
	case types.SeverityCritical:
		return vul.VulnSeverityCritical
	case types.SeverityHigh:
		return vul.VulnSeverityHigh
	case types.SeverityMedium:
		return vul.VulnSeverityMedium
	case types.SeverityLow:
		return vul.VulnSeverityLow
	default:
		return vul.VulnSeverityMedium
	}
}

// vulnTitle 获取漏洞标题，如果 Trivy 无标题则用 CVE ID
func vulnTitle(vuln types.Vulnerability, cveID string) string {
	if vuln.Title != "" {
		return vuln.Title
	}
	return cveID
}

// buildVulnInfo 从匹配结果构建 VulnInfo 模型
func buildVulnInfo(r MatchResult) *vul.VulnInfo {
	var cvssScore *float64
	if r.CvssScore > 0 {
		cvssScore = &r.CvssScore
	}

	var refURLs *string
	if len(r.References) > 0 {
		s := strings.Join(r.References, "\n")
		refURLs = &s
	}

	var fixSuggestion *string
	if r.FixedVersion != "" {
		s := fmt.Sprintf("升级 %s 到版本 %s", r.PackageName, r.FixedVersion)
		fixSuggestion = &s
	}

	var description *string
	if r.Description != "" {
		description = &r.Description
	}

	return &vul.VulnInfo{
		CveID:         r.CveID,
		VulnName:      r.VulnName,
		Severity:      r.Severity,
		CvssScore:     cvssScore,
		Description:   description,
		FixSuggestion: fixSuggestion,
		ReferenceURLs: refURLs,
	}
}
