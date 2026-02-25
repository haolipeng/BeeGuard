package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
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

	var dcDetector *DangerousCommandDetector
	configPath := getConfigPath()
	config, err := LoadRules(configPath)
	if err != nil {
		logger.Warn("Failed to load detection rules, detection disabled", "error", err, "path", configPath)
	} else {
		dcDetector, err = NewDangerousCommandDetector(config)
		if err != nil {
			logger.Warn("Failed to create detector, detection disabled", "error", err)
			dcDetector = nil
		} else {
			logger.Info("Detection rules loaded successfully",
				"version", config.Version,
				"rules", dcDetector.GetEnabledRuleCount())
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

	loader, err := ebpf.NewLoader()
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
			}

			eventType := events.GetEventType(rec.RawSample)
			evtCtx := &eventHandlerCtx{
				client:     client,
				logger:     logger,
				dcDetector: dcDetector,
				rsDetector: rsDetector,
				mrDetector: mrDetector,
				sfDetector: sfDetector,
			}
			var handlerErr error
			switch eventType {
			case events.EventTypeExecve:
				handlerErr = handleExecve(evtCtx, rec.RawSample)
			case events.EventTypeCommitCreds:
				handlerErr = handleCommitCreds(evtCtx, rec.RawSample)
			case events.EventTypeConnect:
				handlerErr = handleConnect(evtCtx, rec.RawSample)
			case events.EventTypeBind:
				handlerErr = handleBind(evtCtx, rec.RawSample)
			case events.EventTypeAccept:
				handlerErr = handleAccept(evtCtx, rec.RawSample)
			case events.EventTypeDNS:
				handlerErr = handleDNS(evtCtx, rec.RawSample)
			case events.EventTypeFile:
				handlerErr = handleFile(evtCtx, rec.RawSample)
			default:
				logger.Warn("Unknown event type", "type", eventType)
			}
			if handlerErr != nil {
				logger.Error("Event handler failed", "type", eventType, "error", handlerErr)
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
