package cronjob

import (
	"context"
	"time"

	"scanner/log"
	"scanner/scanner"
)

// ScanDir 扫描目录配置
type ScanDir struct {
	Path     string
	MaxDepth int
}

// Config 定时扫描配置
type Config struct {
	DirScanInterval  time.Duration
	ProcScanInterval time.Duration
	Throttle         time.Duration
	ScanDirs         []ScanDir
}

// Cronjob 定时扫描调度器
type Cronjob struct {
	scanner  *scanner.Scanner
	config   Config
	resultCh chan<- *scanner.MalwareResult
	logger   *log.Logger
}

// New 创建定时扫描调度器
func New(sc *scanner.Scanner, cfg Config, resultCh chan<- *scanner.MalwareResult, logger *log.Logger) *Cronjob {
	return &Cronjob{
		scanner:  sc,
		config:   cfg,
		resultCh: resultCh,
		logger:   logger,
	}
}

// Run 运行定时扫描（阻塞，直到 ctx 被取消）
func (c *Cronjob) Run(ctx context.Context) {
	dirTicker := time.NewTicker(c.config.DirScanInterval)
	defer dirTicker.Stop()
	procTicker := time.NewTicker(c.config.ProcScanInterval)
	defer procTicker.Stop()

	c.logger.Info("Cronjob started",
		"dir_interval", c.config.DirScanInterval,
		"proc_interval", c.config.ProcScanInterval,
		"throttle", c.config.Throttle,
		"scan_dirs", len(c.config.ScanDirs))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Cronjob stopped")
			return

		case <-procTicker.C:
			c.runProcessScan(ctx)

		case <-dirTicker.C:
			c.runDirScan(ctx)
		}
	}
}

// runProcessScan 执行进程扫描
func (c *Cronjob) runProcessScan(ctx context.Context) {
	c.logger.Info("Starting scheduled process scan...")
	if err := c.scanner.ScanAllProcesses(c.config.Throttle, c.resultCh); err != nil {
		c.logger.Error("Scheduled process scan failed", "error", err)
	} else {
		c.logger.Info("Scheduled process scan completed")
	}
}

// runDirScan 执行目录扫描
func (c *Cronjob) runDirScan(ctx context.Context) {
	c.logger.Info("Starting scheduled directory scan...")
	for _, dir := range c.config.ScanDirs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		c.logger.Info("Scanning directory", "path", dir.Path, "max_depth", dir.MaxDepth)
		if err := c.scanner.ScanDirectory(dir.Path, dir.MaxDepth, c.config.Throttle, c.resultCh); err != nil {
			c.logger.Error("Directory scan error", "path", dir.Path, "error", err)
		}
	}
	c.logger.Info("Scheduled directory scan completed")
}
