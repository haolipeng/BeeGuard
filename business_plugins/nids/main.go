package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"
	"nids/log"
)

func main() {
	// 1. 初始化客户端（FD 3/4 通信）
	client := businessplugins.New()
	defer client.Close()

	// 2. 初始化日志
	logDir := os.Getenv("LOG_DIR")
	logger := log.New(logDir)
	logger.Info("Starting NIDS plugin...")

	// 3. 加载配置
	configPath := getConfigPath()
	cfg, err := LoadConfig(configPath)
	if err != nil {
		logger.Fatal("Failed to load config", "error", err, "path", configPath)
		os.Exit(1)
	}
	logger.Info("Config loaded",
		"interface", cfg.Interface,
		"bpf_filter", cfg.BPFFilter,
		"snaplen", cfg.Snaplen,
		"max_streams", cfg.TCPReassembly.MaxStreams,
		"stream_timeout", cfg.TCPReassembly.StreamTimeout)

	// 4. 解析 Suricata 规则
	rulesPath := getRulesPath(cfg)
	rules, err := LoadRulesFile(rulesPath)
	if err != nil {
		logger.Fatal("Failed to load rules", "error", err, "path", rulesPath)
		os.Exit(1)
	}
	logger.Info("Suricata rules loaded", "count", len(rules), "path", rulesPath)

	// 5. 创建攻击追踪器
	tracker := NewAttackTracker()

	// 6. 创建检测引擎
	detector := NewDetector(rules, tracker, client, logger)

	// 7. 创建抓包器
	factory := NewHTTPStreamFactory(detector, logger, cfg)
	capture, err := NewPacketCapture(cfg, factory, logger)
	if err != nil {
		logger.Fatal("Failed to create packet capture", "error", err,
			"interface", cfg.Interface, "bpf_filter", cfg.BPFFilter)
		os.Exit(1)
	}
	logger.Info("Packet capture initialized",
		"interface", cfg.Interface,
		"bpf_filter", cfg.BPFFilter)

	// 8. 启动抓包循环
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		capture.Run(ctx)
	}()

	// 9. 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Received termination signal, shutting down...")

	// 10. 优雅退出
	cancel()
	capture.Close()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All goroutines exited gracefully")
	case <-time.After(5 * time.Second):
		logger.Warn("Timeout waiting for goroutines to exit, forcing shutdown")
	}

	logger.Info("NIDS plugin shutdown complete")
}
