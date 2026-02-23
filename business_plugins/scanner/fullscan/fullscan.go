package fullscan

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"scanner/log"
	"scanner/scanner"
)

// Config 全盘扫描配置
type Config struct {
	MaxWorkers   int
	QuickTimeout time.Duration
	FullTimeout  time.Duration
	Throttle     time.Duration
}

// ScanDir 扫描目录
type ScanDir struct {
	Path     string
	MaxDepth int
}

// FullScanner 全盘扫描器
type FullScanner struct {
	scanner  *scanner.Scanner
	filter   *scanner.Filter
	config   Config
	logger   *log.Logger
}

// New 创建全盘扫描器
func New(sc *scanner.Scanner, filter *scanner.Filter, cfg Config, logger *log.Logger) *FullScanner {
	return &FullScanner{
		scanner: sc,
		filter:  filter,
		config:  cfg,
		logger:  logger,
	}
}

// Run 执行全盘扫描
// mode: "quick" 或 "full"
func (f *FullScanner) Run(ctx context.Context, mode string, scanDirs []ScanDir, resultCh chan<- *scanner.MalwareResult) error {
	var timeout time.Duration
	if mode == "quick" {
		timeout = f.config.QuickTimeout
	} else {
		timeout = f.config.FullTimeout
	}

	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	f.logger.Info("Full scan started",
		"mode", mode,
		"workers", f.config.MaxWorkers,
		"timeout", timeout,
		"dirs", len(scanDirs))

	// 文件路径通道
	fileCh := make(chan string, f.config.MaxWorkers*10)

	var wg sync.WaitGroup

	// 启动工作线程池
	for i := 0; i < f.config.MaxWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			f.worker(scanCtx, id, fileCh, resultCh)
		}(i)
	}

	// 遍历目录，发送文件到通道
	go func() {
		defer close(fileCh)
		for _, dir := range scanDirs {
			select {
			case <-scanCtx.Done():
				return
			default:
			}
			f.walkDir(scanCtx, dir.Path, dir.MaxDepth, fileCh)
		}
	}()

	// 等待工作线程完成
	wg.Wait()

	// 扫描所有进程
	f.logger.Info("Full scan: scanning processes...")
	if err := f.scanner.ScanAllProcesses(f.config.Throttle, resultCh); err != nil {
		f.logger.Error("Full scan: process scan error", "error", err)
	}

	f.logger.Info("Full scan completed", "mode", mode)
	return nil
}

// worker 工作线程
func (f *FullScanner) worker(ctx context.Context, id int, fileCh <-chan string, resultCh chan<- *scanner.MalwareResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case path, ok := <-fileCh:
			if !ok {
				return
			}

			result, err := f.scanner.ScanFile(path)
			if err != nil {
				f.logger.Warn("Full scan worker error", "worker", id, "path", path, "error", err)
				continue
			}
			if result != nil {
				resultCh <- result
			}

			// 节流
			if f.config.Throttle > 0 {
				time.Sleep(f.config.Throttle)
			}
		}
	}
}

// walkDir 遍历目录，将文件路径发送到通道
func (f *FullScanner) walkDir(ctx context.Context, root string, maxDepth int, fileCh chan<- string) {
	baseDepth := pathDepth(root)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return filepath.SkipAll
		default:
		}

		if err != nil {
			return nil
		}

		currentDepth := pathDepth(path) - baseDepth
		if d.IsDir() {
			if currentDepth >= maxDepth {
				return filepath.SkipDir
			}
			if f.filter.IsPathWhitelisted(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// 快速过滤（不含魔数检测）
		if ok, _ := f.filter.ShouldScan(path); !ok {
			return nil
		}

		select {
		case fileCh <- path:
		case <-ctx.Done():
			return filepath.SkipAll
		}

		return nil
	})
}

// pathDepth 计算路径深度
func pathDepth(path string) int {
	clean := filepath.Clean(path)
	if clean == "/" {
		return 0
	}
	count := 0
	for i := 0; i < len(clean); i++ {
		if clean[i] == os.PathSeparator {
			count++
		}
	}
	return count
}
