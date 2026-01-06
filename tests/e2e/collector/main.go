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

	"gitlab.myinterest.top/security/agent/buffer"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 初始化 logger
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	fmt.Println("=== Collector Plugin Test ===")
	fmt.Println("Starting test agent...")

	wg := &sync.WaitGroup{}
	zap.S().Info("++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++")

	Context, Cancel := context.WithCancel(context.Background())

	// 启动 plugin daemon
	wg.Add(1)
	go plugin.Startup(Context, wg)

	// 等待插件守护进程启动
	time.Sleep(time.Second * 1)

	// 加载 collector 插件
	collectorConfig := &proto.Config{
		Name:    "collector",
		Type:    "binary",
		Version: "1.0.0",
		Sha256:  "", // 测试时可以为空
	}
	cfgs := map[string]*proto.Config{
		"collector": collectorConfig,
	}
	err := plugin.Sync(cfgs)
	if err != nil {
		zap.S().Errorf("failed to load collector plugin: %v", err)
		os.Exit(1)
	} else {
		zap.S().Info("collector plugin loaded successfully")
	}

	// 等待插件加载完成
	time.Sleep(time.Second * 2)

	// 启动结果读取 goroutine
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()
		for {
			select {
			case <-Context.Done():
				return
			case <-ticker.C:
				records := buffer.ReadEncodedRecords()
				for _, rec := range records {
					printRecord(rec)
				}
			}
		}
	}()

	// 发送测试任务（触发进程采集）
	go func() {
		time.Sleep(time.Second * 3)
		sendProcessTask()
	}()

	// 信号处理
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigs
		zap.S().Info("receive signal:", sig.String())
		zap.S().Info("wait for 5 secs to exit")
		<-time.After(time.Second * 5)
		Cancel()
	}()

	// 运行 30 秒后自动退出
	go func() {
		<-time.After(time.Second * 30)
		zap.S().Info("test timeout, exiting...")
		Cancel()
	}()

	wg.Wait()
	fmt.Println("Test completed.")
}

// sendProcessTask 发送进程采集任务给 collector 插件
func sendProcessTask() {
	plg, ok := plugin.Get("collector")
	if !ok {
		zap.S().Error("collector plugin not found")
		return
	}

	// collector 插件使用 DataType 5050 来触发进程采集
	// 根据 collector/process.go，DataType 是 5050
	task := proto.Task{
		DataType:   5050, // 进程采集的数据类型
		ObjectName: "process",
		Data:       "", // collector 插件会自动采集，不需要额外数据
		Token:      "test-process-token-" + fmt.Sprintf("%d", time.Now().Unix()),
	}

	err := plg.SendTask(task)
	if err != nil {
		zap.S().Errorf("failed to send task: %v", err)
	} else {
		zap.S().Info("process collection task sent successfully to collector plugin")
	}
}

// printRecord 打印接收到的记录
func printRecord(rec *proto.EncodedRecord) {
	zap.S().Infof("=== Received Record ===")
	zap.S().Infof("DataType: %d", rec.DataType)
	zap.S().Infof("Timestamp: %d", rec.Timestamp)

	// 进程数据的数据类型是 5050
	if rec.DataType == 5050 {
		zap.S().Infof("Data length: %d bytes", len(rec.Data))

		// 解析 protobuf Payload
		if len(rec.Data) > 0 {
			payload := &businessplugins.Payload{}
			err := payload.Unmarshal(rec.Data)
			if err != nil {
				zap.S().Errorf("Failed to unmarshal payload: %v", err)
			} else {
				fmt.Println("\n========== Process Record ==========")
				fmt.Printf("PID: %s\n", payload.Fields["pid"])
				fmt.Printf("Command: %s\n", payload.Fields["cmdline"])
				fmt.Printf("Executable: %s\n", payload.Fields["exe"])
				fmt.Printf("Working Directory: %s\n", payload.Fields["cwd"])
				fmt.Printf("PPID: %s\n", payload.Fields["ppid"])
				fmt.Printf("State: %s\n", payload.Fields["state"])
				fmt.Printf("User: %s (UID: %s)\n", payload.Fields["rusername"], payload.Fields["ruid"])
				fmt.Printf("Group: %s (GID: %s)\n", payload.Fields["rgid"], payload.Fields["rgid"])
				if nsPid, ok := payload.Fields["pns"]; ok {
					fmt.Printf("Namespace PID: %s\n", nsPid)
				}
				fmt.Println("====================================\n")
			}
		}
	} else if rec.DataType == 5100 {
		// 任务状态响应
		zap.S().Infof("Task status response received")
		if len(rec.Data) > 0 {
			payload := &businessplugins.Payload{}
			err := payload.Unmarshal(rec.Data)
			if err != nil {
				zap.S().Errorf("Failed to unmarshal payload: %v", err)
			} else {
				fmt.Println("\n========== Task Status ==========")
				fmt.Printf("Status: %s\n", payload.Fields["status"])
				fmt.Printf("Token: %s\n", payload.Fields["token"])
				fmt.Printf("Message: %s\n", payload.Fields["msg"])
				fmt.Println("================================\n")
			}
		}
	}
	zap.S().Info("========================")
}
