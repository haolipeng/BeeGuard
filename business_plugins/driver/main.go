package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	businessplugins "business_plugins/lib"
	"driver/detector"
	"driver/ebpf"
	"driver/events"
	"driver/log"
)

// 默认配置文件路径
const defaultConfigPath = "config/dangerous_commands.yaml"

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
	defer loader.Close()

	logger.Info("eBPF program loaded successfully")

	// 5. 创建context用于优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 6. 启动事件读取循环
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// 6.1 从perf buffer读取事件（阻塞）
				rec, err := loader.Read()
				if err != nil {
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

				// 6.3 反序列化事件
				var evt events.ExecveEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal event", "error", err)
					continue
				}

				// 6.4 转换为protobuf格式
				record := evt.ToRecord()

				// 6.5 高危命令检测
				if det != nil {
					comm := cstring(evt.Comm[:])
					args := argsString(evt.Args[:])

					result := det.Detect(comm, args)
					if result != nil {
						// 添加检测结果到record（保留原有字段供调试）
						record.Data.Fields["detection_type"] = detector.DetectionTypeDangerousCommand
						record.Data.Fields["rule_id"] = result.RuleID
						record.Data.Fields["rule_name"] = result.RuleName
						record.Data.Fields["severity"] = result.Severity
						record.Data.Fields["rule_description"] = result.Description
						record.Data.Fields["matched_pattern"] = result.MatchedPattern

						// 添加Server端期望的字段（用于告警入库）
						record.Data.Fields["command"] = args                      // 完整命令行
						record.Data.Fields["command_type"] = result.RuleID        // 使用rule_id作为命令类型
						record.Data.Fields["user"] = record.Data.Fields["uid"]    // 用户ID
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

				// 6.6 发送到Agent
				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send record to agent", "error", err)
				}
			}
		}
	}()

	// 7. 等待退出信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info("Received termination signal, shutting down...")
	cancel() // 通知goroutine退出
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
