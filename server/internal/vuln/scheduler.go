package vuln

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// Scheduler 漏洞匹配定时调度器
type Scheduler struct {
	cfg      *config.VulnConfig
	dbMgr    *DBManager
	matcher  *Matcher
	vulnRepo *repository.VulnRepository
	stopCh   chan struct{}
	cancel   context.CancelFunc // 取消正在执行的漏洞匹配任务
	wg       sync.WaitGroup
}

// NewScheduler 创建漏洞匹配调度器
func NewScheduler(cfg *config.VulnConfig, dbMgr *DBManager) *Scheduler {
	vulnRepo := repository.NewVulnRepository()
	matcher := NewMatcher(dbMgr, vulnRepo)

	return &Scheduler{
		cfg:      cfg,
		dbMgr:    dbMgr,
		matcher:  matcher,
		vulnRepo: vulnRepo,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动调度器，解析 cron 表达式并定时执行漏洞匹配
func (s *Scheduler) Start() {
	interval := parseCronToInterval(s.cfg.ScanCron)
	log.Infof("[VulnScheduler] 调度器已启动，匹配间隔: %s", interval)

	// 在 Start 中创建 context，避免与 Stop 的 data race
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.wg.Add(1)
	go s.run(ctx, interval)
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.cancel() // 取消正在执行的漏洞匹配任务
	s.wg.Wait()
	log.Infof("[VulnScheduler] 调度器已停止")
}

// run 主循环：定时执行漏洞匹配
func (s *Scheduler) run(ctx context.Context, interval time.Duration) {
	defer s.wg.Done()

	// 启动后延迟 30 秒执行第一次匹配（等待系统稳定）
	select {
	case <-time.After(30 * time.Second):
	case <-s.stopCh:
		return
	}

	// 执行第一次匹配
	s.executeMatchAll(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.executeMatchAll(ctx)
		case <-s.stopCh:
			return
		}
	}
}

// executeMatchAll 执行一次完整的漏洞匹配（主机 + 镜像）
func (s *Scheduler) executeMatchAll(ctx context.Context) {
	startTime := time.Now()

	log.Infof("[VulnScheduler] 开始执行漏洞匹配任务...")

	// 检查并更新漏洞数据库，暂时关闭，否则测试环境每次需要重新下载漏洞数据库
	/*if s.dbMgr.NeedsUpdate() {
		if err := s.dbMgr.Update(ctx); err != nil {
			log.Errorf("[VulnScheduler] 漏洞数据库更新失败: %v", err)
			// 如果数据库仍可用，继续匹配
			if !s.dbMgr.IsReady() {
				log.Errorf("[VulnScheduler] 漏洞数据库不可用，跳过本次匹配")
				return
			}
		}
	}*/

	// 执行主机漏洞匹配
	hostCount, hostVulnCount := s.matchAllHosts(ctx)

	// 执行镜像漏洞匹配（当前暂无镜像包数据，预留接口）
	imageCount, imageVulnCount := s.matchAllImages(ctx)

	elapsed := time.Since(startTime)
	log.Infof("[VulnScheduler] 漏洞匹配任务完成: 耗时=%s, 主机=%d(发现%d个漏洞), 镜像=%d(发现%d个漏洞)",
		elapsed, hostCount, hostVulnCount, imageCount, imageVulnCount)
}

// matchAllHosts 对所有主机执行漏洞匹配
func (s *Scheduler) matchAllHosts(ctx context.Context) (hostCount, vulnCount int) {
	hosts, err := s.vulnRepo.GetAllHosts(ctx)
	if err != nil {
		log.Errorf("[VulnScheduler] 获取主机列表失败: %v", err)
		return 0, 0
	}

	if len(hosts) == 0 {
		log.Infof("[VulnScheduler] 无主机数据，跳过主机漏洞匹配")
		return 0, 0
	}

	log.Infof("[VulnScheduler] 开始匹配 %d 台主机的漏洞...", len(hosts))

	for _, host := range hosts {
		select {
		case <-s.stopCh:
			log.Infof("[VulnScheduler] 收到停止信号，中断主机匹配")
			return hostCount, vulnCount
		default:
		}

		matched, err := s.matcher.MatchHostVulns(ctx, host)
		if err != nil {
			log.Errorf("[VulnScheduler] 主机漏洞匹配失败 (agent=%s, host=%s): %v",
				host.AgentID, host.HostName, err)
			continue
		}
		hostCount++
		vulnCount += matched
	}

	return hostCount, vulnCount
}

// matchAllImages 对所有镜像执行漏洞匹配
// 从 asset_image_package 表读取镜像内软件包，与 Trivy 漏洞库进行匹配
func (s *Scheduler) matchAllImages(ctx context.Context) (imageCount, vulnCount int) {
	images, err := s.vulnRepo.GetAllImages(ctx)
	if err != nil {
		log.Errorf("[VulnScheduler] 获取镜像列表失败: %v", err)
		return 0, 0
	}

	if len(images) == 0 {
		log.Infof("[VulnScheduler] 无镜像数据，跳过镜像漏洞匹配")
		return 0, 0
	}

	log.Infof("[VulnScheduler] 开始匹配 %d 个镜像的漏洞...", len(images))

	for _, image := range images {
		select {
		case <-s.stopCh:
			log.Infof("[VulnScheduler] 收到停止信号，中断镜像匹配")
			return imageCount, vulnCount
		default:
		}

		// 获取镜像内 OS 版本
		osVersion, err := s.vulnRepo.GetImageOSVersion(ctx, image.AgentID, image.ImageID)
		if err != nil {
			log.Errorf("[VulnScheduler] 获取镜像OS版本失败 (image=%s): %v", image.ImageName, err)
			continue
		}

		// 标准化 OS 版本为 Trivy DB source 格式
		source := repository.NormalizeOSVersion("linux", osVersion)
		if source == "" {
			log.Debugf("[VulnScheduler] 镜像OS版本信息缺失，跳过 (image=%s)", image.ImageName)
			continue
		}

		// 获取镜像软件包列表
		packages, err := s.vulnRepo.GetImagePackages(ctx, image.AgentID, image.ImageID)
		if err != nil {
			log.Errorf("[VulnScheduler] 获取镜像软件包失败 (image=%s): %v", image.ImageName, err)
			continue
		}

		if len(packages) == 0 {
			continue
		}

		matched, err := s.matcher.MatchImageVulns(ctx, image, source, packages)
		if err != nil {
			log.Errorf("[VulnScheduler] 镜像漏洞匹配失败 (agent=%s, image=%s): %v",
				image.AgentID, image.ImageName, err)
			continue
		}
		imageCount++
		vulnCount += matched
	}

	return imageCount, vulnCount
}

// RunOnce 手动触发一次漏洞匹配（供 HTTP API 调用）
func (s *Scheduler) RunOnce(ctx context.Context) error {
	if !s.dbMgr.IsReady() {
		return fmt.Errorf("漏洞数据库未就绪")
	}

	go s.executeMatchAll(ctx)
	return nil
}

// parseCronToInterval 将简单 cron 表达式转换为 time.Duration
// 支持: "0 2 * * *" (每天凌晨2点) → 24h
// 复杂的 cron 表达式退化为默认 24h
func parseCronToInterval(cron string) time.Duration {
	if cron == "" {
		return 24 * time.Hour
	}

	// 简单解析：检查是否为每日/每小时模式
	// "0 2 * * *" → 每天 → 24h
	// "0 */6 * * *" → 每6小时 → 6h
	// 其他 → 默认24h
	fields := splitFields(cron)
	if len(fields) < 5 {
		return 24 * time.Hour
	}

	// 如果小时字段包含 */N，解析为每 N 小时
	if len(fields[1]) > 2 && fields[1][:2] == "*/" {
		hours := 0
		for _, c := range fields[1][2:] {
			if c >= '0' && c <= '9' {
				hours = hours*10 + int(c-'0')
			}
		}
		if hours > 0 && hours <= 24 {
			return time.Duration(hours) * time.Hour
		}
	}

	// 默认每天执行
	return 24 * time.Hour
}

// splitFields 按空白字符分割字符串
func splitFields(s string) []string {
	var fields []string
	current := ""
	for _, c := range s {
		if c == ' ' || c == '\t' {
			if current != "" {
				fields = append(fields, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		fields = append(fields, current)
	}
	return fields
}
