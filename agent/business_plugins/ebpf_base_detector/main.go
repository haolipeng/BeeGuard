package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/ebpf"
	"ebpf_base_detector/events"
	"ebpf_base_detector/log"
	"ebpf_base_detector/trusted"
)

func main() {
	client := businessplugins.New()
	defer client.Close()

	logDir := os.Getenv("LOG_DIR")
	logger := log.New(logDir)
	logger.Info("Starting eBPF driver plugin...")

	var hdcDetector *DangerousCommandDetector
	configPath := getConfigPath()
	config, err := LoadRules(configPath)
	if err != nil {
		logger.Warn("Failed to load detection rules, detection disabled", "error", err, "path", configPath)
	} else {
		hdcDetector, err = NewDangerousCommandDetector(config)
		if err != nil {
			logger.Warn("Failed to create detector, detection disabled", "error", err)
			hdcDetector = nil
		} else {
			logger.Info("Detection rules loaded successfully",
				"version", config.Version,
				"rules", hdcDetector.GetEnabledRuleCount())
		}
	}

	// 容器高危命令检测器（独立规则集）
	var cdcDetector *DangerousCommandDetector
	cdcConfigPath := getContainerDangerousCommandConfigPath()
	cdcConfig, err := LoadRules(cdcConfigPath)
	if err != nil {
		logger.Warn("Failed to load container dangerous command rules, container command detection disabled",
			"error", err, "path", cdcConfigPath)
	} else {
		cdcDetector, err = NewDangerousCommandDetector(cdcConfig)
		if err != nil {
			logger.Warn("Failed to create container dangerous command detector, detection disabled", "error", err)
			cdcDetector = nil
		} else {
			logger.Info("Container dangerous command rules loaded successfully",
				"version", cdcConfig.Version,
				"rules", cdcDetector.GetEnabledRuleCount())
		}
	}

	rsDetector := &ReverseShellDetector{}

	var mrDetector *MaliciousRequestDetector
	mrConfigPath := getMaliciousRequestConfigPath()
	mrConfig, err := LoadMaliciousRequestRules(mrConfigPath)
	if err != nil {
		logger.Warn("Failed to load malicious request rules, malicious request detection disabled", "error", err, "path", mrConfigPath)
	} else {
		mrDetector = NewMaliciousRequestDetector(mrConfig)
		logger.Info("Malicious request rules loaded", "version", mrConfig.Version, "rules", mrDetector.GetEnabledRuleCount())
	}

	var sfDetector *SensitiveFileDetector
	sfConfigPath := getSensitiveFileConfigPath()
	sfConfig, err := LoadRules(sfConfigPath)
	if err != nil {
		logger.Warn("Failed to load sensitive file rules, sensitive file detection disabled", "error", err, "path", sfConfigPath)
	} else {
		sfDetector, err = NewSensitiveFileDetector(sfConfig)
		if err != nil {
			logger.Warn("Failed to create sensitive file detector, detection disabled", "error", err)
			sfDetector = nil
		} else {
			logger.Info("Sensitive file rules loaded successfully",
				"version", sfConfig.Version,
				"rules", sfDetector.GetEnabledRuleCount())
		}
	}

	// 容器逃逸检测器
	ceDetector := NewContainerEscapeDetector()
	logger.Info("Container escape detector initialized")

	// 容器反弹 Shell 检测器
	crsDetector := &ContainerReverseShellDetector{}
	logger.Info("Container reverse shell detector initialized")

	// 容器敏感文件检测器（独立规则集）
	var csfDetector *SensitiveFileDetector
	csfConfigPath := getContainerSensitiveFileConfigPath()
	csfConfig, err := LoadRules(csfConfigPath)
	if err != nil {
		logger.Warn("Failed to load container sensitive file rules, container sensitive file detection disabled",
			"error", err, "path", csfConfigPath)
	} else {
		csfDetector, err = NewSensitiveFileDetector(csfConfig)
		if err != nil {
			logger.Warn("Failed to create container sensitive file detector, detection disabled", "error", err)
			csfDetector = nil
		} else {
			logger.Info("Container sensitive file rules loaded successfully",
				"version", csfConfig.Version,
				"rules", csfDetector.GetEnabledRuleCount())
		}
	}

	// 容器元数据缓存
	containerMeta := NewContainerMetaCache(5 * time.Minute)
	logger.Info("Container metadata cache initialized")

	loader, err := ebpf.NewLoader(getBTFDir(), logger)
	if err != nil {
		logger.Fatal("Failed to load eBPF program", "error", err)
		os.Exit(1)
	}
	logger.Info("eBPF program loaded successfully")

	trustedConfigPath := getTrustedConfigPath()
	trustedConfig, err := trusted.LoadConfig(trustedConfigPath)
	if err != nil {
		logger.Warn("Failed to load trusted executables config, whitelist disabled",
			"error", err, "path", trustedConfigPath)
	} else {
		trustedMap := loader.GetTrustedExesMap()
		count, err := trusted.PopulateTrustedExesMap(trustedMap, trustedConfig, logger)
		if err != nil {
			logger.Warn("Failed to populate trusted executables map", "error", err)
		} else {
			logger.Info("Trusted executables whitelist loaded",
				"count", count,
				"enabled", trustedConfig.Enabled)
		}
	}

	fileWhitelistPath := getFileMonitorWhitelistPath()
	fileWhitelistConfig, err := trusted.LoadConfig(fileWhitelistPath)
	if err != nil {
		logger.Warn("Failed to load file monitor whitelist config, whitelist disabled",
			"error", err, "path", fileWhitelistPath)
	} else {
		fileTrustedMap := loader.GetFileTrustedExesMap()
		count, err := trusted.PopulateTrustedExesMap(fileTrustedMap, fileWhitelistConfig, logger)
		if err != nil {
			logger.Warn("Failed to populate file monitor whitelist map", "error", err)
		} else {
			logger.Info("File monitor whitelist loaded",
				"count", count,
				"enabled", fileWhitelistConfig.Enabled)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var totalLostSamples uint64

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				logger.Info("Event reading goroutine exiting...")
				return
			default:
			}

			rec, err := loader.Read()
			if err != nil {
				if ctx.Err() != nil {
					logger.Info("Event reading stopped due to context cancellation")
					return
				}
				if errors.Is(err, syscall.EINTR) {
					continue
				}
				logger.Error("Failed to read from perf buffer", "error", err)
				continue
			}

			if rec.LostSamples > 0 {
				logger.Warn("Lost samples", "count", rec.LostSamples, "cpu", rec.CPU)
				atomic.AddUint64(&totalLostSamples, rec.LostSamples)
			}

			eventType := events.GetEventType(rec.RawSample)
			evtCtx := &eventHandlerCtx{
				client:        client,
				logger:        logger,
				dcDetector:    hdcDetector, // 主机高危命令检测器
				cdcDetector:   cdcDetector, // 容器高危命令检测器
				rsDetector:    rsDetector,
				mrDetector:    mrDetector,
				sfDetector:    sfDetector,
				csfDetector:   csfDetector,
				ceDetector:    ceDetector,
				crsDetector:   crsDetector,
				containerMeta: containerMeta,
			}
			var handlerErr error
			switch eventType {
			case events.EventTypeExecve: // 执行命令事件
				handlerErr = handleExecve(evtCtx, rec.RawSample)
			case events.EventTypeCommitCreds: // 提权事件
				handlerErr = handleCommitCreds(evtCtx, rec.RawSample)
			case events.EventTypeConnect: // 连接事件
				handlerErr = handleConnect(evtCtx, rec.RawSample)
			case events.EventTypeDNS: // DNS事件
				handlerErr = handleDNS(evtCtx, rec.RawSample)
			case events.EventTypeFile: // 文件操作事件
				handlerErr = handleFile(evtCtx, rec.RawSample)
			case events.EventTypeMount: // mount 事件
				handlerErr = handleMount(evtCtx, rec.RawSample)
			default:
				logger.Warn("Unknown event type", "type", eventType) // 未知事件类型
			}
			if handlerErr != nil {
				logger.Error("Event handler failed", "type", eventType, "error", handlerErr)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lost := atomic.SwapUint64(&totalLostSamples, 0)
				if lost == 0 {
					continue
				}
				logger.Warn("Reporting perf event loss to server", "lost_count", lost)
				rec := &businessplugins.Record{
					DataType:  events.DataTypePerfEventLoss,
					Timestamp: time.Now().Unix(),
					Data: &businessplugins.Payload{
						Fields: map[string]string{
							"lost_count":      fmt.Sprintf("%d", lost),
							"report_interval": "30",
						},
					},
				}
				if err := client.SendRecord(rec); err != nil {
					logger.Error("Failed to send perf event loss record", "error", err)
				}
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Received termination signal, shutting down...")
	cancel()
	logger.Info("Closing eBPF loader...")
	loader.Close()

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

	logger.Info("Driver plugin shutdown complete")
}
