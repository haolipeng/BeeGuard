package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/ebpf"
	"ebpf_base_detector/events"
	"ebpf_base_detector/log"
	"ebpf_base_detector/trusted"
)

// 默认配置文件路径
const (
	defaultConfigPath                 = "config/dangerous_commands.yaml"
	defaultTrustedConfigPath          = "config/privilege_escalation_whitelist.yaml"
	defaultMaliciousRequestConfigPath = "config/malicious_request_rules.yaml"
	defaultSensitiveFileConfigPath    = "config/sensitive_file_rules.yaml"
	defaultFileMonitorWhitelistPath   = "config/file_monitor_whitelist.yaml"
)

func main() {
	// 1. 初始化客户端（FD 3/4通信）
	client := businessplugins.New()
	defer client.Close()

	// 2. 初始化日志组件
	logDir := os.Getenv("LOG_DIR")
	logger := log.New(logDir)
	logger.Info("Starting eBPF driver plugin...")

	// 3. 加载高危命令检测规则
	var det *DangerousCommandDetector
	configPath := getConfigPath()
	config, err := LoadRules(configPath)
	if err != nil {
		logger.Warn("Failed to load detection rules, detection disabled", "error", err, "path", configPath)
	} else {
		det, err = NewDangerousCommandDetector(config)
		if err != nil {
			logger.Warn("Failed to create detector, detection disabled", "error", err)
			det = nil
		} else {
			logger.Info("Detection rules loaded successfully",
				"version", config.Version,
				"rules", det.GetEnabledRuleCount())
		}
	}

	// 3.1 初始化用户态反弹 shell 检测器
	rsDetector := &ReverseShellDetector{}

	// 3.2 初始化恶意请求检测器
	var mrDetector *MaliciousRequestDetector
	mrConfigPath := getMaliciousRequestConfigPath()
	mrConfig, err := LoadMaliciousRequestRules(mrConfigPath)
	if err != nil {
		logger.Warn("Failed to load malicious request rules, malicious request detection disabled", "error", err, "path", mrConfigPath)
	} else {
		mrDetector = NewMaliciousRequestDetector(mrConfig)
		logger.Info("Malicious request rules loaded", "version", mrConfig.Version, "rules", mrDetector.GetEnabledRuleCount())
	}

	// 3.3 初始化敏感文件检测器
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

	// 5.1 加载文件监控白名单
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

	// 6. 创建context用于优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用 WaitGroup 等待 goroutine 退出
	var wg sync.WaitGroup

	// 6. 启���事件读取循环
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

				// 用户态补充 pid_tree（从 BPF 移到用户态，减少验证器指令数）
				pidTreeStr := buildPidTree(evt.TGID, cstring(evt.Comm[:]))

				// fd_type 已由内核通过 i_mode 检查直接推导，无需用户态处理

				record := evt.ToRecord()
				record.Data.Fields["pid_tree"] = pidTreeStr

				// 高危命令检测
				if det != nil {
					comm := cstring(evt.Comm[:])
					args := argsString(evt.Args[:])

					result := det.Detect(comm, args)
					if result != nil {
						// 修改DataType为高危命令告警类型（6003），以便Server端正确处理
						record.DataType = 6003

						// 添加检测结果到record（保留原有字段供调试）
						record.Data.Fields["detection_type"] = DetectionTypeDangerousCommand
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

				// 用户态反弹 shell 检测
				rsResult := rsDetector.Detect(&evt)
				if rsResult != nil {
					rsRecord := BuildReverseShellRecord(&evt, rsResult, pidTreeStr)
					logger.Warn("Reverse shell detected (userspace)",
						"rule", rsResult.RuleName,
						"confidence", rsResult.Confidence,
						"pid", evt.PID,
						"tgid", evt.TGID,
						"comm", cstring(evt.Comm[:]),
						"exe_path", cstring(evt.ExePath[:]),
						"stdin_path", cstring(evt.StdinPath[:]),
						"stdout_path", cstring(evt.StdoutPath[:]),
						"pid_tree", pidTreeStr,
						"tty_name", cstring(evt.TTYName[:]),
						"socket_pid", evt.SocketPID)
					if err := client.SendRecord(rsRecord); err != nil {
						logger.Error("Failed to send reverse shell record to agent", "error", err)
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

				// eBPF 在 kprobe 上下文中 dentry 遍历可能失败，仅返回文件名
				// 通过 /proc/<pid>/exe 补全完整路径
				exePath := resolveExePath(evt.TGID, cstring(evt.ExePath[:]))

				record := evt.ToRecord()
				record.Data.Fields["exe_path"] = exePath

				// userspace 丰富字段：通过 /proc 和 /etc/passwd 补充服务端需要的高级字段
				record.Data.Fields["escalated_user"] = resolveUsername(evt.NewUID)
				record.Data.Fields["parent_process"] = resolveParentComm(evt.PPID)
				record.Data.Fields["parent_process_user"] = resolveParentUID(evt.PPID)
				record.Data.Fields["timestamp"] = fmt.Sprintf("%d", record.Timestamp)

				// 记录提权告警日志
				logger.Warn("Privilege escalation detected",
					"pid", evt.PID,
					"tgid", evt.TGID,
					"ppid", evt.PPID,
					"comm", cstring(evt.Comm[:]),
					"exe_path", exePath,
					"escalated_user", record.Data.Fields["escalated_user"],
					"parent_process", record.Data.Fields["parent_process"],
					"parent_process_user", record.Data.Fields["parent_process_user"],
					"old_uid", evt.OldUID,
					"old_euid", evt.OldEUID,
					"new_uid", evt.NewUID,
					"new_euid", evt.NewEUID)

				// 发送到Agent
				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send privilege escalation record to agent", "error", err)
				}

			case events.EventTypeConnect:
				// 处理出站连接事件
				var evt events.ConnectEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal connect event", "error", err)
					continue
				}

				record := evt.ToRecord()

				logger.Info("Connect event",
					"pid", evt.PID,
					"comm", cstring(evt.Comm[:]),
					"remote_ip", record.Data.Fields["remote_ip"],
					"remote_port", record.Data.Fields["remote_port"],
					"protocol", record.Data.Fields["protocol"],
					"retval", evt.RetVal)

				// 恶意请求匹配
				if mrDetector != nil {
					if mrResult := mrDetector.MatchConnect(&evt); mrResult != nil {
						mrRecord := BuildMaliciousRequestConnectRecord(&evt, mrResult)
						logger.Warn("Malicious request detected on connect",
							"rule_id", mrResult.RuleID,
							"rule_name", mrResult.RuleName,
							"threat_type", mrResult.ThreatType,
							"matched_value", mrResult.MatchedValue,
							"pid", evt.PID,
							"comm", cstring(evt.Comm[:]))
						if err := client.SendRecord(mrRecord); err != nil {
							logger.Error("Failed to send malicious request connect record to agent", "error", err)
						}
					}
				}

				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send connect record to agent", "error", err)
				}

			case events.EventTypeBind:
				// 处理端口绑定事件
				var evt events.BindEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal bind event", "error", err)
					continue
				}

				record := evt.ToRecord()

				logger.Info("Bind event",
					"pid", evt.PID,
					"comm", cstring(evt.Comm[:]),
					"bind_ip", record.Data.Fields["bind_ip"],
					"bind_port", record.Data.Fields["bind_port"],
					"protocol", record.Data.Fields["protocol"])

				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send bind record to agent", "error", err)
				}

			case events.EventTypeAccept:
				// 处理入站连接事件
				var evt events.AcceptEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal accept event", "error", err)
					continue
				}

				record := evt.ToRecord()

				logger.Info("Accept event",
					"pid", evt.PID,
					"comm", cstring(evt.Comm[:]),
					"remote_ip", record.Data.Fields["remote_ip"],
					"remote_port", record.Data.Fields["remote_port"],
					"local_port", record.Data.Fields["local_port"],
					"protocol", record.Data.Fields["protocol"])

				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send accept record to agent", "error", err)
				}

			case events.EventTypeDNS:
				// 处理DNS查询事件
				var evt events.DNSEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal DNS event", "error", err)
					continue
				}

				record := evt.ToRecord()

				logger.Info("DNS query event",
					"pid", evt.PID,
					"comm", cstring(evt.Comm[:]),
					"domain", record.Data.Fields["domain"],
					"query_type", record.Data.Fields["query_type"],
					"dns_server", record.Data.Fields["dns_server_ip"])

				// 恶意请求匹配
				if mrDetector != nil {
					if mrResult := mrDetector.MatchDNS(&evt); mrResult != nil {
						mrRecord := BuildMaliciousRequestDNSRecord(&evt, mrResult)
						logger.Warn("Malicious request detected on DNS",
							"rule_id", mrResult.RuleID,
							"rule_name", mrResult.RuleName,
							"threat_type", mrResult.ThreatType,
							"matched_value", mrResult.MatchedValue,
							"pid", evt.PID,
							"comm", cstring(evt.Comm[:]))
						if err := client.SendRecord(mrRecord); err != nil {
							logger.Error("Failed to send malicious request DNS record to agent", "error", err)
						}
					}
				}

				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send DNS record to agent", "error", err)
				}

			case events.EventTypeFile:
				// 处理文件操作事件
				var evt events.FileEvent
				if err := evt.UnmarshalBinary(rec.RawSample); err != nil {
					logger.Error("Failed to unmarshal file event", "error", err)
					continue
				}

				// 构建基础 Record (DataType=64)
				record := evt.ToRecord()

				// 用户态补充 pid_tree
				pidTreeStr := buildPidTree(evt.TGID, cstring(evt.Comm[:]))
				record.Data.Fields["pid_tree"] = pidTreeStr

				newPath := cstring(evt.NewPath[:])
				actionStr := "create"
				if evt.Action == events.FileActionRename {
					actionStr = "rename"
				}

				logger.Info("File event",
					"pid", evt.PID,
					"comm", cstring(evt.Comm[:]),
					"action", actionStr,
					"new_path", newPath,
					"old_path", cstring(evt.OldPath[:]),
					"s_id", cstring(evt.SID[:]))

				// 敏感文件检测
				if sfDetector != nil {
					result := sfDetector.Detect(newPath)
					if result != nil {
						// 发送告警 Record (DataType=6009)
						alertRecord := evt.ToRecord()
						alertRecord.DataType = 6009
						alertRecord.Data.Fields["detection_type"] = DetectionTypeSensitiveFile
						alertRecord.Data.Fields["rule_id"] = result.RuleID
						alertRecord.Data.Fields["rule_name"] = result.RuleName
						alertRecord.Data.Fields["severity"] = result.Severity
						alertRecord.Data.Fields["rule_description"] = result.Description
						alertRecord.Data.Fields["matched_pattern"] = result.MatchedPattern
						alertRecord.Data.Fields["pid_tree"] = pidTreeStr

						logger.Warn("Sensitive file operation detected",
							"rule_id", result.RuleID,
							"rule_name", result.RuleName,
							"severity", result.Severity,
							"action", actionStr,
							"new_path", newPath,
							"pid", evt.PID,
							"comm", cstring(evt.Comm[:]))

						if err := client.SendRecord(alertRecord); err != nil {
							logger.Error("Failed to send sensitive file alert record to agent", "error", err)
						}
					}
				}

				// 发送基础事件 Record
				if err := client.SendRecord(record); err != nil {
					logger.Error("Failed to send file event record to agent", "error", err)
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

// getMaliciousRequestConfigPath 获取恶意请求规则配置文件路径
// 优先使用环境变量，否则使用默认路径
func getMaliciousRequestConfigPath() string {
	if path := os.Getenv("DRIVER_MALICIOUS_REQUEST_CONFIG_PATH"); path != "" {
		return path
	}

	// 尝试获取可执行文件所在目录
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultMaliciousRequestConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	// 回退到当前目录
	return defaultMaliciousRequestConfigPath
}

// getSensitiveFileConfigPath 获取敏感文件规则配置文件路径
// 优先使用环境变量，否则使用默认路径
func getSensitiveFileConfigPath() string {
	if path := os.Getenv("DRIVER_SENSITIVE_FILE_CONFIG_PATH"); path != "" {
		return path
	}

	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultSensitiveFileConfigPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return defaultSensitiveFileConfigPath
}

// getFileMonitorWhitelistPath 获取文件监控白名单配置文件路径
// 优先使用环境变量，否则使用默认路径
func getFileMonitorWhitelistPath() string {
	if path := os.Getenv("DRIVER_FILE_MONITOR_WHITELIST_PATH"); path != "" {
		return path
	}

	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		configPath := filepath.Join(dir, defaultFileMonitorWhitelistPath)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
	}

	return defaultFileMonitorWhitelistPath
}

// resolveExePath 补全可执行文件的完整路径
// eBPF 在 kprobe 上下文中 dentry 遍历可能失败，仅返回文件名
// 通过 /proc/<pid>/exe readlink 获取完整路径
func resolveExePath(tgid uint32, ebpfPath string) string {
	if len(ebpfPath) > 0 && ebpfPath[0] == '/' {
		return ebpfPath
	}
	link, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", tgid))
	if err == nil {
		return link
	}
	return ebpfPath
}

// resolveParentComm 读取父进程名称
// 通过 /proc/<ppid>/comm 获取父进程的命令名
func resolveParentComm(ppid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// resolveParentUID 读取父进程的 UID
// 通过 /proc/<ppid>/status 中的 Uid 行获取真实 UID
func resolveParentUID(ppid uint32) string {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", ppid))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1] // 真实 UID
			}
			break
		}
	}
	return ""
}

// resolveUsername 将 UID 解析为用户名
// 通过读取 /etc/passwd 将数字 UID 映射为用户名（如 0 → "root"）
func resolveUsername(uid uint32) string {
	uidStr := fmt.Sprintf("%d", uid)

	f, err := os.Open("/etc/passwd")
	if err != nil {
		return uidStr
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[2] == uidStr {
			return parts[0]
		}
	}
	return uidStr
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

// buildPidTree 在用户态构建进程链字符串
// 格式: "PID<comm<PID<comm<..."（与原 BPF 版本格式一致）
// 从当前进程向上遍历最多 8 层
func buildPidTree(tgid uint32, comm string) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%d<%s", tgid, comm))

	pid := tgid
	for i := 0; i < 7; i++ {
		ppid := readPPid(pid)
		if ppid == 0 || ppid == pid {
			break
		}
		parentComm := resolveParentComm(ppid)
		if parentComm == "" {
			break
		}
		buf.WriteString(fmt.Sprintf("<%d<%s", ppid, parentComm))
		pid = ppid
	}

	return buf.String()
}

// readPPid 从 /proc/<pid>/status 读取父进程 PID
func readPPid(pid uint32) uint32 {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PPid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var ppid uint32
				if n, _ := fmt.Sscanf(fields[1], "%d", &ppid); n == 1 {
					return ppid
				}
			}
			break
		}
	}
	return 0
}

