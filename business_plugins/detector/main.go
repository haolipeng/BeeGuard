package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/business_plugins/detector/config"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/detector/ftp"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/detector/ssh"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/detector/ssh_anomaly_login"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// DetectorConfigUpdateDataType Server 下发配置更新的数据类型
	DetectorConfigUpdateDataType = int32(6010)
	// DetectorTaskStatusDataType 任务状态回报数据类型
	DetectorTaskStatusDataType = int32(6011)
	// TaskStatusSuccess 任务成功状态
	TaskStatusSuccess = "succeed"
	// TaskStatusFailed 任务失败状态
	TaskStatusFailed = "failed"
)

var (
	configPath = flag.String("config", "config/rules", "规则配置目录路径")
	// detectors 存储所有检测器实例，用于配置更新
	detectors = make(map[string]engine.ConfigUpdater)
	// pluginClient 插件客户端
	pluginClient *businessplugins.Client
)

// sendTaskStatus 发送任务状态到 server
func sendTaskStatus(status string, token string, msg string) {
	record := businessplugins.Record{}
	record.DataType = DetectorTaskStatusDataType
	record.Timestamp = time.Now().Unix()

	payload := businessplugins.Payload{}
	field := make(map[string]string)
	field["status"] = status
	if token != "" {
		field["token"] = token
	}
	if msg != "" {
		field["msg"] = msg
	}
	payload.Fields = field
	record.Data = &payload

	if err := pluginClient.SendRecord(&record); err != nil {
		zap.S().Errorf("failed to send task status: %v", err)
	}
}

// handleTask 处理从 server 接收的任务
func handleTask(task *businessplugins.Task) {
	zap.S().Infof("received task: data_type=%d, object_name=%s", task.DataType, task.ObjectName)

	// 只处理配置更新任务
	if task.DataType != DetectorConfigUpdateDataType {
		zap.S().Warnf("unknown task data_type: %d", task.DataType)
		sendTaskStatus(TaskStatusFailed, task.Token, "unknown data_type")
		return
	}

	// 根据 object_name 查找对应的检测器
	detector, exists := detectors[task.ObjectName]
	if !exists {
		zap.S().Warnf("detector not found: %s", task.ObjectName)
		sendTaskStatus(TaskStatusFailed, task.Token, "detector not found: "+task.ObjectName)
		return
	}

	// 更新配置
	if err := detector.UpdateConfig(task.Data); err != nil {
		zap.S().Errorf("failed to update config for %s: %v", task.ObjectName, err)
		sendTaskStatus(TaskStatusFailed, task.Token, err.Error())
		return
	}

	zap.S().Infof("config updated for detector: %s", task.ObjectName)
	sendTaskStatus(TaskStatusSuccess, task.Token, "")
}

func main() {
	flag.Parse()

	// 初始化客户端(与Agent通信)
	pluginClient = businessplugins.New()

	// 初始化日志
	l := log.New(log.Config{
		MaxSize:     1,
		Path:        "detector.log",
		FileLevel:   zapcore.InfoLevel,
		RemoteLevel: zapcore.ErrorLevel,
		MaxBackups:  10,
		Compress:    true,
		Client:      pluginClient,
	})
	defer l.Sync()
	zap.ReplaceGlobals(l)

	zap.S().Info("detector plugin starting...")

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}
	zap.S().Infof("loaded config: %+v", cfg)

	// 创建检测引擎
	eng := engine.New(pluginClient)

	// 注册SSH检测器
	if cfg.SSH.Enabled {
		sshDetector := ssh.New(cfg.SSH)
		eng.Register(sshDetector)
		// 注册到 detectors map，用于配置更新
		detectors["ssh"] = sshDetector
		zap.S().Info("SSH detector registered")
	}

	// 注册FTP检测器
	if cfg.FTP.Enabled {
		ftpDetector := ftp.New(cfg.FTP)
		eng.Register(ftpDetector)
		// 注册到 detectors map，用于配置更新
		detectors["ftp"] = ftpDetector
		zap.S().Info("FTP detector registered")
	}

	// 注册SSH异常登录检测器
	if cfg.SSHAnomaly.Enabled {
		sshAnomalyDetector, err := ssh_anomaly_login.New(cfg.SSHAnomaly)
		if err != nil {
			zap.S().Errorf("failed to create ssh_anomaly_login detector: %v", err)
		} else {
			eng.Register(sshAnomalyDetector)
			// 注册到 detectors map，用于配置更新
			detectors["ssh_anomaly_login"] = sshAnomalyDetector
			zap.S().Info("SSH anomaly login detector registered")
		}
	}

	// 启动任务接收循环
	go func() {
		for {
			task, err := pluginClient.ReceiveTask()
			if err != nil {
				zap.S().Errorf("ReceiveTask error: %v", err)
				break
			}
			// 在 goroutine 中处理任务
			go handleTask(task)
		}
	}()

	// 启动检测引擎
	go eng.Run()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	zap.S().Info("detector plugin stopping...")

	eng.Stop()
	pluginClient.Close()
	zap.S().Info("detector plugin stopped")
}
