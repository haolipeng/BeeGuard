package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"gitlab.myinterest.top/security/agent/agent"
	"gitlab.myinterest.top/security/agent/config"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/standalone"
	"gitlab.myinterest.top/security/agent/transport"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "Path to config file")
	testMode := flag.Bool("test", false, "Enable test mode with fixed agent ID (123456)")
	standaloneMode := flag.Bool("standalone", false, "Enable standalone mode (no gRPC transport)")
	outputMode := flag.String("output", "", "Standalone output: stderr (default) or file path")
	pluginsList := flag.String("plugins", "", "Comma-separated list of plugins to load (standalone mode)")
	flag.Parse()

	fmt.Println("agent start running!")

	// 设置测试模式（如果通过命令行指定）
	if *testMode {
		agent.SetTestMode()
		fmt.Println("Test mode enabled, agent ID:", agent.TestAgentID)
	}

	// 设置配置文件路径（如果通过命令行指���）
	if *configPath != "" {
		config.SetConfigPath(*configPath)
	}

	// 初始化配置
	if err := config.Init(); err != nil {
		slog.Error("failed to init config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("config initialized successfully")

	// 设置 standalone 模式（通过命令行指定）
	if *standaloneMode {
		var plugins []string
		if *pluginsList != "" {
			plugins = strings.Split(*pluginsList, ",")
		}
		if err := config.SetStandalone(true, *outputMode, plugins); err != nil {
			slog.Error("failed to set standalone mode", slog.String("error", err.Error()))
			os.Exit(1)
		}
		fmt.Println("Standalone mode enabled")
	}

	// 将配置同��到 agent 包
	cfg, _ := config.Get()
	agent.WorkingDirectory = cfg.WorkingDirectory
	agent.PluginsDirectory = cfg.PluginsDirectory
	agent.LogDirectory = cfg.LogDirectory

	// 初始化 zap logger（在配置加载后，使用 LogDirectory）
	logger, err := initLogger(cfg)
	if err != nil {
		fmt.Printf("failed to init zap logger: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	wg := &sync.WaitGroup{}
	zap.S().Info("++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++")

	Context, Cancel := context.WithCancel(context.Background())

	// 启动插件守护进程
	wg.Add(1)
	go plugin.Startup(Context, wg)

	// 根据模式启动不同的传输守护进程
	wg.Add(1)
	if config.IsStandalone() {
		zap.S().Info("running in standalone mode, transport disabled")
		go standalone.StartOutputHandler(Context, wg)
	} else {
		go transport.StartTransfer(Context, wg)
	}

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigs
		zap.S().Warnf("receive signal: %s, agent will shutdown...", sig.String())
		zap.S().Info("waiting 5 seconds for graceful shutdown...")
		<-time.After(time.Second * 5)
		Cancel()
	}()

	wg.Wait()

	zap.S().Info("all goroutines exited, agent shutdown complete")
	fmt.Println("agent stopped")
}

// initLogger 初始化 zap logger，同时输出到 stderr 和文件
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// stderr 输出
	stderrCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapcore.DebugLevel)

	// 如果 LogDirectory 非空，添加文件输出
	if cfg.LogDirectory != "" {
		logDir := filepath.Join(cfg.LogDirectory, "agent")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}

		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(logDir, "agent.log"),
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			Compress:   cfg.Log.Compress,
		})
		fileCore := zapcore.NewCore(encoder, fileWriter, zapcore.DebugLevel)

		return zap.New(zapcore.NewTee(stderrCore, fileCore), zap.AddCaller()), nil
	}

	return zap.New(stderrCore, zap.AddCaller()), nil
}
