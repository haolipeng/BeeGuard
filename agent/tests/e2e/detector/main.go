package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/agent"
	"gitlab.myinterest.top/security/agent/buffer"
	"gitlab.myinterest.top/security/agent/config"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

// 临时日志文件路径
const (
	tmpSSHLog = "/tmp/test_ssh_auth.log"
	tmpFTPLog = "/tmp/test_ftp.log"
)

// 规则配置文件路径（相对于 build/plugins/detector/）
var rulesDir = "../../../build/plugins/detector/config/rules"

// ruleFile 规则文件信息
type ruleFile struct {
	path    string
	backup  []byte // 原始内容备份
}

func main() {
	// 初始化 logger
	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := logConfig.Build()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	fmt.Println("=== Detector Plugin E2E Test ===")

	// 创建临时日志文件
	if err := createTempLogFiles(); err != nil {
		zap.S().Fatalf("failed to create temp log files: %v", err)
	}
	defer cleanupTempLogFiles()

	// 备份并修改规则配置文件
	ruleFiles, err := patchRuleConfigs()
	if err != nil {
		zap.S().Fatalf("failed to patch rule configs: %v", err)
	}
	defer restoreRuleConfigs(ruleFiles)

	// 初始化 agent 配置
	if err := config.Init(); err != nil {
		zap.S().Fatalf("failed to init config: %v", err)
	}

	cfg, _ := config.Get()
	agent.PluginsDirectory = cfg.PluginsDirectory
	agent.PluginsDirectory = "../../../build/plugins"

	if err := config.SetStandalone(true, "stderr", []string{"detector"}); err != nil {
		zap.S().Fatalf("failed to set standalone mode: %v", err)
	}
	zap.S().Info("standalone mode enabled, plugins directory: ../../../build/plugins")

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	// 启动 plugin daemon
	wg.Add(1)
	go plugin.Startup(ctx, wg)

	// 等待 detector 启动并开始 tail
	zap.S().Info("waiting for detector plugin to start...")
	time.Sleep(5 * time.Second)

	// 注入模拟攻击日志
	zap.S().Info("injecting simulated attack logs...")
	injectSSHBruteForce()
	injectFTPBruteForce()
	zap.S().Info("attack logs injected, waiting for alerts...")

	// 结果跟踪
	sshAlertReceived := false
	ftpAlertReceived := false

	// 轮询读取告警
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				records := buffer.ReadEncodedRecords()
				for _, rec := range records {
					printAlertRecord(rec)
					switch rec.DataType {
					case 6001:
						sshAlertReceived = true
					case 6002:
						ftpAlertReceived = true
					}
					// 两种告警都收到后提前结束
					if sshAlertReceived && ftpAlertReceived {
						printResults(sshAlertReceived, ftpAlertReceived)
						cancel()
						return
					}
				}
			}
		}
	}()

	// 信号处理
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigs
		zap.S().Infof("received signal: %s", sig.String())
		cancel()
	}()

	// 超时退出
	go func() {
		<-time.After(60 * time.Second)
		zap.S().Warn("test timeout, exiting...")
		printResults(sshAlertReceived, ftpAlertReceived)
		cancel()
	}()

	wg.Wait()

	if !sshAlertReceived || !ftpAlertReceived {
		printResults(sshAlertReceived, ftpAlertReceived)
	}
	fmt.Println("Test completed.")
}

// createTempLogFiles 创建临时日志文件
func createTempLogFiles() error {
	for _, path := range []string{tmpSSHLog, tmpFTPLog} {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create %s: %w", path, err)
		}
		f.Close()
	}
	zap.S().Infof("created temp log files: %s, %s", tmpSSHLog, tmpFTPLog)
	return nil
}

// cleanupTempLogFiles 清理临时日志文件
func cleanupTempLogFiles() {
	os.Remove(tmpSSHLog)
	os.Remove(tmpFTPLog)
	zap.S().Info("cleaned up temp log files")
}

// patchRuleConfigs 备份并修改规则配置文件，使其指向临时日志文件
func patchRuleConfigs() ([]ruleFile, error) {
	var files []ruleFile

	// 修改 SSH 规则配置
	sshPath := rulesDir + "/ssh_brute_force.yaml"
	if err := patchYAMLConfig(sshPath, "ssh", []string{tmpSSHLog}, &files); err != nil {
		restoreRuleConfigs(files)
		return nil, fmt.Errorf("failed to patch ssh config: %w", err)
	}

	// 修改 FTP 规则配置
	ftpPath := rulesDir + "/ftp_brute_force.yaml"
	if err := patchYAMLConfig(ftpPath, "ftp", []string{tmpFTPLog}, &files); err != nil {
		restoreRuleConfigs(files)
		return nil, fmt.Errorf("failed to patch ftp config: %w", err)
	}

	return files, nil
}

// patchYAMLConfig 修改单个 YAML 配置文件
func patchYAMLConfig(path string, rootKey string, logPaths []string, files *[]ruleFile) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	// 备份原始内容
	*files = append(*files, ruleFile{path: path, backup: data})

	// 解析 YAML
	var content map[string]any
	if err := yaml.Unmarshal(data, &content); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// 修改配置
	if section, ok := content[rootKey].(map[string]any); ok {
		section["log_paths"] = logPaths
		section["whitelist"] = []string{}
	} else {
		return fmt.Errorf("root key %q not found in %s", rootKey, path)
	}

	// 写回文件
	newData, err := yaml.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", path, err)
	}
	if err := os.WriteFile(path, newData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	zap.S().Infof("patched config: %s (log_paths=%v, whitelist=[])", path, logPaths)
	return nil
}

// restoreRuleConfigs 恢复原始规则配置文件
func restoreRuleConfigs(files []ruleFile) {
	for _, f := range files {
		if err := os.WriteFile(f.path, f.backup, 0644); err != nil {
			zap.S().Errorf("failed to restore %s: %v", f.path, err)
		} else {
			zap.S().Infof("restored config: %s", f.path)
		}
	}
}

// injectSSHBruteForce 注入模拟 SSH 暴力破解日志（8条，超过阈值6）
func injectSSHBruteForce() {
	now := time.Now()
	f, err := os.OpenFile(tmpSSHLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		zap.S().Errorf("failed to open SSH log: %v", err)
		return
	}
	defer f.Close()

	for i := 0; i < 8; i++ {
		t := now.Add(time.Duration(i) * time.Second)
		// 格式匹配 sshd failedPasswordRegex: "Failed password for <user> from <ip>"
		line := fmt.Sprintf("%s testhost sshd[%d]: Failed password for root from 192.168.1.200 port 22 ssh2\n",
			t.Format("Jan  2 15:04:05"), 12345+i)
		if _, err := f.WriteString(line); err != nil {
			zap.S().Errorf("failed to write SSH log line: %v", err)
			return
		}
	}
	zap.S().Info("injected 8 SSH brute force log entries")
}

// injectFTPBruteForce 注入模拟 FTP 暴力破解日志（8条，超过阈值6）
func injectFTPBruteForce() {
	now := time.Now()
	f, err := os.OpenFile(tmpFTPLog, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		zap.S().Errorf("failed to open FTP log: %v", err)
		return
	}
	defer f.Close()

	for i := 0; i < 8; i++ {
		t := now.Add(time.Duration(i) * time.Second)
		// 格式匹配 vsftpd failLoginRegex: "[pid <pid>] [<user>] FAIL LOGIN: Client "<ip>""
		line := fmt.Sprintf("%s [pid %d] [testuser] FAIL LOGIN: Client \"192.168.1.201\"\n",
			t.Format("Mon Jan  2 15:04:05 2006"), 12345+i)
		if _, err := f.WriteString(line); err != nil {
			zap.S().Errorf("failed to write FTP log line: %v", err)
			return
		}
	}
	zap.S().Info("injected 8 FTP brute force log entries")
}

// printAlertRecord 打印告警记录
func printAlertRecord(rec *proto.EncodedRecord) {
	payload := unmarshalPayload(rec)
	if payload == nil {
		zap.S().Warnf("failed to unmarshal alert record (DataType=%d)", rec.DataType)
		return
	}

	var label string
	switch rec.DataType {
	case 6001:
		label = "SSH Brute Force Alert"
	case 6002:
		label = "FTP Brute Force Alert"
	case 6005:
		label = "SSH Anomaly Login Alert"
	case 6011:
		label = "Task Status"
	default:
		label = fmt.Sprintf("Unknown Alert (DataType=%d)", rec.DataType)
	}

	fmt.Printf("\n========== %s (DataType %d) ==========\n", label, rec.DataType)
	fmt.Printf("  alert_type:  %s\n", payload.Fields["alert_type"])
	fmt.Printf("  service:     %s\n", payload.Fields["service"])
	fmt.Printf("  rule_name:   %s\n", payload.Fields["rule_name"])
	fmt.Printf("  description: %s\n", payload.Fields["description"])
	fmt.Printf("  source_ip:   %s\n", payload.Fields["source_ip"])
	fmt.Printf("  target_user: %s\n", payload.Fields["target_user"])
	fmt.Printf("  count:       %s\n", payload.Fields["count"])
	fmt.Printf("  timeframe:   %s\n", payload.Fields["timeframe"])
	fmt.Printf("  first_seen:  %s\n", payload.Fields["first_seen"])
	fmt.Printf("  last_seen:   %s\n", payload.Fields["last_seen"])
	fmt.Printf("  level:       %s\n", payload.Fields["level"])
	fmt.Println("==========================================")
}

// unmarshalPayload 解析 EncodedRecord 的 Data 为 Payload
func unmarshalPayload(rec *proto.EncodedRecord) *businessplugins.Payload {
	if len(rec.Data) == 0 {
		return nil
	}
	payload := &businessplugins.Payload{}
	if err := payload.Unmarshal(rec.Data); err != nil {
		zap.S().Errorf("failed to unmarshal payload: %v", err)
		return nil
	}
	return payload
}

// printResults 打印测试结果
func printResults(sshOK, ftpOK bool) {
	fmt.Println()
	fmt.Println("========== Test Results ==========")
	if sshOK {
		fmt.Println("[PASS] SSH brute force alert received")
	} else {
		fmt.Println("[FAIL] SSH brute force alert NOT received")
	}
	if ftpOK {
		fmt.Println("[PASS] FTP brute force alert received")
	} else {
		fmt.Println("[FAIL] FTP brute force alert NOT received")
	}
	fmt.Println("==================================")
}
