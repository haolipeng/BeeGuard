package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"
	"scanner/cronjob"
	"scanner/engine"
	"scanner/fullscan"
	"scanner/log"
	"scanner/scanner"
	"scanner/updater"
)

func main() {
	// 1. 初始化客户端（FD 3/4 通信）
	client := businessplugins.New()
	defer client.Close()

	// 2. 初始化日志
	logDir := os.Getenv("LOG_DIR")
	logger := log.New(logDir)
	logger.Info("Starting scanner plugin...")

	// 3. 加载配置
	configPath := getConfigPath()
	cfg, err := LoadConfig(configPath)
	if err != nil {
		logger.Fatal("Failed to load config", "error", err, "path", configPath)
		os.Exit(1)
	}
	logger.Info("Config loaded",
		"db_path", cfg.Scanner.Engine.DBPath,
		"max_file_size", cfg.Scanner.Engine.MaxFileSize,
		"max_scan_time", cfg.Scanner.Engine.MaxScanTime,
		"scan_dirs", len(cfg.Scanner.ScanDirs))

	// 4. 初始化 ClamAV 引擎
	eng := engine.NewClamAVEngine(cfg.Scanner.Engine.MaxFileSize, cfg.Scanner.Engine.MaxScanTime)
	if err := eng.Init(); err != nil {
		logger.Fatal("Failed to init ClamAV engine", "error", err)
		os.Exit(1)
	}
	defer eng.Close()
	logger.Info("ClamAV engine initialized")

	// 5. 加载病毒数据库
	dbPath := resolveDBPath(cfg.Scanner.Engine.DBPath)
	if err := eng.LoadDB(dbPath); err != nil {
		logger.Fatal("Failed to load virus database", "error", err, "path", dbPath)
		os.Exit(1)
	}
	logger.Info("Virus database loaded", "path", dbPath)

	// 6. ���建文件过滤器和扫描器
	filter := scanner.NewFilter(
		cfg.Scanner.Filter.PathWhitelist,
		cfg.Scanner.Filter.SkipFileTypes,
		cfg.Scanner.Filter.MinFileSize,
		cfg.Scanner.Filter.MaxFileSize,
	)
	sc := scanner.New(eng, filter, logger)

	// 7. 应用 Cgroup 资源限制
	var cgroupMgr *fullscan.CgroupManager
	if cfg.Scanner.Cgroup.Enabled {
		cgroupMgr = fullscan.NewCgroupManager("scanner_plugin", logger)
		cgroupCfg := fullscan.CgroupConfig{
			Enabled:  true,
			MemoryMB: cfg.Scanner.Cgroup.NormalMemoryMB,
			CPUQuota: cfg.Scanner.Cgroup.NormalCPUQuota,
		}
		if err := cgroupMgr.Apply(cgroupCfg); err != nil {
			logger.Warn("Failed to apply cgroup limits", "error", err)
		}
		defer cgroupMgr.Remove()
	}

	// 8. 创建更新器
	dbUpdater := updater.New(eng, dbPath, logger)

	// 9. 创建上下文和 WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	// 10. 启动结果上报协程
	resultCh := make(chan *scanner.MalwareResult, 100)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				// 排空剩余结果
				for {
					select {
					case result := <-resultCh:
						sendResult(client, result, logger)
					default:
						return
					}
				}
			case result := <-resultCh:
				sendResult(client, result, logger)
			}
		}
	}()

	// 11. 启动定时扫描
	cronDirs := make([]cronjob.ScanDir, len(cfg.Scanner.ScanDirs))
	for i, d := range cfg.Scanner.ScanDirs {
		cronDirs[i] = cronjob.ScanDir{Path: d.Path, MaxDepth: d.MaxDepth}
	}
	cron := cronjob.New(sc, cronjob.Config{
		DirScanInterval:  cfg.Scanner.Cronjob.ParseDirScanInterval(),
		ProcScanInterval: cfg.Scanner.Cronjob.ParseProcScanInterval(),
		Throttle:         cfg.Scanner.Cronjob.ParseThrottle(),
		ScanDirs:         cronDirs,
	}, resultCh, logger)

	wg.Add(1)
	go func() {
		defer wg.Done()
		cron.Run(ctx)
	}()

	// 12. 启动任务接收协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		receiveTask(ctx, client, sc, eng, dbUpdater, cgroupMgr, filter, cfg, resultCh, logger)
	}()

	// 13. 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Received termination signal, shutting down...")

	// 14. 优雅退出
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All goroutines exited gracefully")
	case <-time.After(10 * time.Second):
		logger.Warn("Timeout waiting for goroutines to exit, forcing shutdown")
	}

	logger.Info("Scanner plugin shutdown complete")
}

// sendResult 发送检出结果到 Agent
func sendResult(client *businessplugins.Client, result *scanner.MalwareResult, logger *log.Logger) {
	dataType := int32(scanner.DataTypeFileDetect)
	record := result.ToRecord(dataType)
	if err := client.SendRecord(record); err != nil {
		logger.Error("Failed to send record", "error", err, "path", result.FilePath)
	} else {
		logger.Info("Malware detected", "detail", scanner.FormatResult(result))
	}
}

// receiveTask 接收控制台下发的任务
func receiveTask(ctx context.Context, client *businessplugins.Client, sc *scanner.Scanner, eng engine.Engine, dbUpdater *updater.Updater, cgroupMgr *fullscan.CgroupManager, filter *scanner.Filter, cfg *ScannerConfig, resultCh chan<- *scanner.MalwareResult, logger *log.Logger) {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Task receiver stopped")
			return
		default:
		}

		task, err := client.ReceiveTask()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				logger.Error("Failed to receive task", "error", err)
				time.Sleep(time.Second)
				continue
			}
		}

		logger.Info("Received task", "data_type", task.DataType, "data", task.Data)

		switch task.DataType {
		case scanner.DataTypeDBUpdate:
			handleDBUpdate(ctx, dbUpdater, task, logger, client)
		case scanner.DataTypeDirScan:
			handleDirScan(ctx, sc, task, cfg, resultCh, logger, client)
		case scanner.DataTypeFullScan:
			handleFullScan(ctx, sc, filter, cgroupMgr, task, cfg, resultCh, logger, client)
		default:
			logger.Warn("Unknown task type", "data_type", task.DataType)
		}
	}
}

// handleDBUpdate 处理病毒库更新任务
func handleDBUpdate(ctx context.Context, dbUpdater *updater.Updater, task *businessplugins.Task, logger *log.Logger, client *businessplugins.Client) {
	var taskData struct {
		Version string `json:"version"`
		SHA256  string `json:"sha256"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal([]byte(task.Data), &taskData); err != nil {
		logger.Error("Failed to parse DB update task", "error", err)
		client.SendRecord(scanner.NewStatusRecord("error", "invalid task data: "+err.Error()))
		return
	}

	logger.Info("DB update task received", "version", taskData.Version, "url", taskData.URL)
	client.SendRecord(scanner.NewStatusRecord("running", "updating database to "+taskData.Version))

	if err := dbUpdater.Update(taskData.URL, taskData.SHA256); err != nil {
		logger.Error("DB update failed", "error", err)
		client.SendRecord(scanner.NewStatusRecord("error", "update failed: "+err.Error()))
		return
	}

	logger.Info("Virus database updated", "version", taskData.Version)
	client.SendRecord(scanner.NewStatusRecord("success", "db updated to "+taskData.Version))
}

// handleDirScan 处理指定目录扫描任务
func handleDirScan(ctx context.Context, sc *scanner.Scanner, task *businessplugins.Task, cfg *ScannerConfig, resultCh chan<- *scanner.MalwareResult, logger *log.Logger, client *businessplugins.Client) {
	var taskData struct {
		Path string `json:"exe"`
	}
	if err := json.Unmarshal([]byte(task.Data), &taskData); err != nil {
		logger.Error("Failed to parse dir scan task", "error", err)
		client.SendRecord(scanner.NewStatusRecord("error", "invalid task data: "+err.Error()))
		return
	}

	logger.Info("Dir scan task received", "path", taskData.Path)
	client.SendRecord(scanner.NewStatusRecord("running", "scanning "+taskData.Path))

	throttle := cfg.Scanner.Cronjob.ParseThrottle()
	if err := sc.ScanDirectory(taskData.Path, 20, throttle, resultCh); err != nil {
		logger.Error("Dir scan failed", "path", taskData.Path, "error", err)
		client.SendRecord(scanner.NewStatusRecord("error", err.Error()))
		return
	}

	logger.Info("Dir scan completed", "path", taskData.Path)
	client.SendRecord(scanner.NewStatusRecord("success", "scan completed for "+taskData.Path))
}

// handleFullScan 处理全盘扫描任务
func handleFullScan(ctx context.Context, sc *scanner.Scanner, filter *scanner.Filter, cgroupMgr *fullscan.CgroupManager, task *businessplugins.Task, cfg *ScannerConfig, resultCh chan<- *scanner.MalwareResult, logger *log.Logger, client *businessplugins.Client) {
	var taskData struct {
		Mode    string `json:"mode"`
		Workers int    `json:"workers"`
		Timeout string `json:"timeout"`
	}
	if err := json.Unmarshal([]byte(task.Data), &taskData); err != nil {
		logger.Error("Failed to parse full scan task", "error", err)
		client.SendRecord(scanner.NewStatusRecord("error", "invalid task data: "+err.Error()))
		return
	}

	logger.Info("Full scan task received", "mode", taskData.Mode, "workers", taskData.Workers)
	client.SendRecord(scanner.NewStatusRecord("running", "full scan started"))

	// 提升 cgroup 限制
	if cgroupMgr != nil && cfg.Scanner.Cgroup.Enabled {
		fullscanCgroup := fullscan.CgroupConfig{
			Enabled:  true,
			MemoryMB: cfg.Scanner.Cgroup.FullscanMemoryMB,
			CPUQuota: cfg.Scanner.Cgroup.FullscanCPUQuota,
		}
		if err := cgroupMgr.Apply(fullscanCgroup); err != nil {
			logger.Warn("Failed to apply fullscan cgroup limits", "error", err)
		}
		// 扫描完成后恢复正常限制
		defer func() {
			normalCgroup := fullscan.CgroupConfig{
				Enabled:  true,
				MemoryMB: cfg.Scanner.Cgroup.NormalMemoryMB,
				CPUQuota: cfg.Scanner.Cgroup.NormalCPUQuota,
			}
			if err := cgroupMgr.Apply(normalCgroup); err != nil {
				logger.Warn("Failed to restore normal cgroup limits", "error", err)
			}
		}()
	}

	// 确定参数
	workers := cfg.Scanner.Fullscan.MaxWorkers
	if taskData.Workers > 0 {
		workers = taskData.Workers
	}

	mode := taskData.Mode
	if mode == "" {
		mode = "full"
	}

	// 创建全盘扫描器
	fsCfg := fullscan.Config{
		MaxWorkers:   workers,
		QuickTimeout: cfg.Scanner.Fullscan.ParseQuickTimeout(),
		FullTimeout:  cfg.Scanner.Fullscan.ParseFullTimeout(),
		Throttle:     cfg.Scanner.Cronjob.ParseThrottle(),
	}
	fs := fullscan.New(sc, filter, fsCfg, logger)

	// 构建扫描目录列表
	scanDirs := make([]fullscan.ScanDir, len(cfg.Scanner.ScanDirs))
	for i, d := range cfg.Scanner.ScanDirs {
		scanDirs[i] = fullscan.ScanDir{Path: d.Path, MaxDepth: d.MaxDepth}
	}

	if err := fs.Run(ctx, mode, scanDirs, resultCh); err != nil {
		logger.Error("Full scan failed", "error", err)
		client.SendRecord(scanner.NewStatusRecord("error", err.Error()))
		return
	}

	logger.Info("Full scan completed")
	client.SendRecord(scanner.NewStatusRecord("success", "full scan completed"))
}
