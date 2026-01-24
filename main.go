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

	"gitlab.myinterest.top/security/agent/config"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/transport"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	fmt.Println("agent start running!")

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
		zap.S().Error("receive signal:", sig.String())
		zap.S().Info("wait for 5 secs to exit")
		<-time.After(time.Second * 5)
		Cancel()
	}()

	wg.Wait()
}
