package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"
	"driver/detector"
	"driver/ebpf"
	"driver/events"
	"driver/log"
	"driver/trusted"
)

// 默认配置文件路径
const (
	defaultConfigPath        = "config/dangerous_commands.yaml"
	defaultTrustedConfigPath = "config/trusted_executables.yaml"
)

func main() {
	// 1. 初始化客户端（FD 3/4通信）
	client := businessplugins.New()
	defer client.Close()

	// 2. 初始化日志组件
	logger := log.New()
	logger.Info("Starting eBPF driver plugin...")

	// 3. 加载高危命令检测规则
	var det *detector.Detector
	configPath := getConfigPath()
	config, err := detector.LoadRules(configPath)
	if err != nil {
		logger.Warn("Failed to load detection rules, detection disabled", "error", err, "path", configPath)
	} else {
		det, err = detector.NewDetector(config)
		if err != nil {
			logger.Warn("Failed to create detector, detection disabled", "error", err)
			det = nil
		} else {
			logger.Info("Detection rules loaded successfully",
				"version", config.Version,
				"rules", det.GetEnabledRuleCount())
		}
	}

	// 4. 加载eBPF程序
	loader, err := ebpf.NewLoader()
	if err != nil {
		logger.Fatal("Failed to load eBPF program", "error", err)
		os.Exit(1)
	}
	// 注意：loader.Close() 在退出逻辑中显式调用，不使用 defer
	// 这样可以立即中断阻塞的 Read() 调用，实现优雅退出

	logger.Info("eBPF program loaded successfully")

	// 5. 加载可信任可执行文件白名单
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

	// 6. 创建context用于优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用 WaitGroup 等待 goroutine 退出
	var wg sync.WaitGroup

	// 6. 启动事件读取循环
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// 先检查 context 是否已取消
			select {
			case <-ctx.Done():
				logger.Info("Event reading goroutine exiting...")
				return
			default:
			}

			// 6.1 从perf buffer读取事件（阻塞）
			rec, err := loader.Read()
			if err != nil {
				// 检查是否因为 context 取消导致的错误
				if ctx.Err() != nil {
					logger.Info("Event reading stopped due to context cancellation")
					return
				}
				if errors.Is(err, syscall.EINTR) {
					// 被信号中断，继续
					continue
				}
				logger.Error("Failed to read from perf buffer", "error", err)
				continue
			}

			// 6.2 检查丢失事件
			if rec.LostSamples > 0 {
				logger.Warn("Lost samples", "count", rec.LostSamples, "cpu", rec.CPU)
			}

			// 6.3 根据事件类型分发处理
			eventType := events.GetEventType(rec.RawSample)

			switch eventType {
			case events.EventTypeExecve:
				// 处理execve事件
				var evt events.ExecveEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal execve event", "error", err)
					continue
				}

				record := evt.ToRecord()

				// 高危命令检测
				if det != nil {
					comm := cstring(evt.Comm[:])
					args := argsString(evt.Args[:])

					result := det.Detect(comm, args)
					if result != nil {
						// 修改DataType为高危命令告警类型（6003），以便Server端正确处理
						record.DataType = 6003

						// 添加检测结果到record（保留原有字段供调试）
						record.Data.Fields["detection_type"] = detector.DetectionTypeDangerousCommand
						record.Data.Fields["rule_id"] = result.RuleID
						record.Data.Fields["rule_name"] = result.RuleName
						record.Data.Fields["severity"] = result.Severity
						record.Data.Fields["rule_description"] = result.Description
						record.Data.Fields["matched_pattern"] = result.MatchedPattern

						// 添加Server端期望的字段（用于告警入库）
						record.Data.Fields["command"] = args                   // 完整命令行
						record.Data.Fields["command_type"] = result.RuleID     // 使用rule_id作为命令类型
						record.Data.Fields["user"] = record.Data.Fields["uid"] // 用户ID
						if evt.UID == 0 {
							record.Data.Fields["privilege_level"] = "root"
						} else {
							record.Data.Fields["privilege_level"] = "normal"
						}
						record.Data.Fields["timestamp"] = fmt.Sprintf("%d", record.Timestamp)

						logger.Info("Dangerous command detected",
							"rule_id", result.RuleID,
							"rule_name", result.RuleName,
							"severity", result.Severity,
							"uid", evt.UID,
							"comm", comm,
							"args", args)
					}
				}

				// 发送到Agent
				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send execve record to agent", "error", err)
				}

			case events.EventTypeCommitCreds:
				// 处理提权事件
				var evt events.CommitCredsEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal commit_creds event", "error", err)
					continue
				}

				record := evt.ToRecord()

				// 记录提权告警日志
				logger.Warn("Privilege escalation detected",
					"pid", evt.PID,
					"tgid", evt.TGID,
					"ppid", evt.PPID,
					"comm", cstring(evt.Comm[:]),
					"exe_path", cstring(evt.ExePath[:]),
					"old_uid", evt.OldUID,
					"old_euid", evt.OldEUID,
					"new_uid", evt.NewUID,
					"new_euid", evt.NewEUID)

				// 发送到Agent
				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send privilege escalation record to agent", "error", err)
				}

			default:
				logger.Warn("Unknown event type", "type", eventType)
			}
		}
	}()

	// 7. 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Received termination signal, shutting down...")

	// 1. 取消 context，通知所有 goroutine 退出
	cancel()

	// 2. 关闭 loader，这会中断阻塞的 Read() 调用
	logger.Info("Closing eBPF loader...")
	loader.Close()

	// 3. 等待所有 goroutine 退出（最多等待 5 秒）
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

// getConfigPath 获取配置文件路径
// 优先使用环境变量，否则使用默认路径
func getConfigPath() string {
	if path := os.Getenv("DRIVER_CONFIG_PATH"); path != "" {
		return path
	}

	// 尝试获取可执行文件所在目录
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 回退到当前目录
	return defaultConfigPath
}

// getTrustedConfigPath 获取白名单配置文件路径
// 优先使用环境变量，否则使用默认路径
func getTrustedConfigPath() string {
	if path := os.Getenv("DRIVER_TRUSTED_CONFIG_PATH"); path != "" {
		return path
	}

	// 尝试获取可执行文件所在目录
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultTrustedConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 回退到当前目录
	return defaultTrustedConfigPath
}

// cstring 将C字符串（以\0结尾）转换为Go字符串
func cstring(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

// argsString 处理命令行参数：将NULL字节分隔的多个参数转换为空格分隔的字符串
func argsString(b []byte) string {
	// 找到实际数据的结尾（连续的NULL字节）
	end := len(b)
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			allZero := true
			for j := i; j < len(b) && j < i+4; j++ {
				if b[j] != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				end = i
				break
			}
		}
	}

	// 将NULL字节替换为空格
	result := make([]byte, end)
	copy(result, b[:end])
	for i := 0; i < len(result); i++ {
		if result[i] == 0 {
			result[i] = ' '
		}
	}

	// 去除尾部空格
	return string(bytes.TrimRight(result, " "))
}
