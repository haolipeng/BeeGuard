package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/business_plugins/detector/config"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/detector/ssh"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	configPath = flag.String("config", "config/rules", "规则配置目录路径")
)

func main() {
	flag.Parse()

	// 初始化客户端(与Agent通信)
	c := businessplugins.New()

	// 初始化日志
	l := log.New(log.Config{
		MaxSize:     1,
		Path:        "detector.log",
		FileLevel:   zapcore.InfoLevel,
		RemoteLevel: zapcore.ErrorLevel,
		MaxBackups:  10,
		Compress:    true,
		Client:      c,
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
	eng := engine.New(c)

	// 注册SSH检测器
	if cfg.SSH.Enabled {
		sshDetector := ssh.New(cfg.SSH)
		eng.Register(sshDetector)
		zap.S().Info("SSH detector registered")
	}

	// 启动检测引擎
	go eng.Run()

	// 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	zap.S().Info("detector plugin stopping...")
	eng.Stop()
	c.Close()
	zap.S().Info("detector plugin stopped")
}
