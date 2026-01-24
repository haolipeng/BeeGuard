package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gitlab.myinterest.top/security/agent/agent"
	"gitlab.myinterest.top/security/agent/config"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/transport"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "Path to config file")
	testMode := flag.Bool("test", false, "Enable test mode with fixed agent ID (123456)")
	flag.Parse()

	fmt.Println("agent start running!")

	// 初始化 zap logger
	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logConfig.Build()
	if err != nil {
		fmt.Printf("failed to init zap logger: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// 设置测试模式（如果通过命令行指定）
	if *testMode {
		agent.SetTestMode()
		fmt.Println("Test mode enabled, agent ID:", agent.TestAgentID)
	}

	// 设置配置文件路径（如果通过命令行指定）
	if *configPath != "" {
		config.SetConfigPath(*configPath)
	}

	// 初始化配置
	if err := config.Init(); err != nil {
		slog.Error("failed to init config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("config initialized successfully")

	wg := &sync.WaitGroup{}
	zap.S().Info("++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++")

	Context, Cancel := context.WithCancel(context.Background())

	// 启动插件守护进程
	wg.Add(1)
	go plugin.Startup(Context, wg)

	// 启动传输守护进程（gRPC 连接）
	wg.Add(1)
	go transport.StartTransfer(Context, wg)

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
