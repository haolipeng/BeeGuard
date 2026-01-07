package main

import (
	"context"
	"encoding/json"
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

// jsonWriter JSON 文件写入器
type jsonWriter struct {
	file *os.File
	enc  *json.Encoder
	mu   sync.Mutex
}

// 配置：是否将记录写入 JSON 文件
var (
	// enableJSONOutput 控制是否将接收到的记录写入 JSON 文件
	// 设置为 true 启用 JSON 文件输出，设置为 false 禁用
	enableJSONOutput = true

	// jsonOutputFile 指定 JSON 输出文件的路径
	jsonOutputFile = "collector_records.json"

	// jsonWriterInst JSON 文件写入器实例
	jsonWriterInst *jsonWriter
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

	// 初始化 JSON 文件写入器（如果启用）
	if enableJSONOutput {
		var err error
		jsonWriterInst, err = newJSONWriter(jsonOutputFile)
		if err != nil {
			zap.S().Warnf("Failed to initialize JSON writer: %v, JSON output disabled", err)
			enableJSONOutput = false
		} else {
			zap.S().Infof("JSON output enabled, writing to: %s", jsonOutputFile)
			defer jsonWriterInst.Close()
		}
	}

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
					// 如果启用 JSON 输出，将记录写入文件
					if enableJSONOutput && jsonWriterInst != nil {
						writeRecordToJSON(rec)
					}
				}
			}
		}
	}()

	// 发送测试任务（触发进程采集、端口采集和内核模块采集）
	go func() {
		time.Sleep(time.Second * 3)
		sendProcessTask()
		time.Sleep(time.Second * 2)
		sendPortTask()
		time.Sleep(time.Second * 2)
		sendKmodTask()
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

// sendPortTask 发送端口采集任务给 collector 插件
func sendPortTask() {
	//获取collector插件的实例
	plg, ok := plugin.Get("collector")
	if !ok {
		zap.S().Error("collector plugin not found")
		return
	}

	// collector 插件使用 DataType 5051 来触发端口采集
	// 根据 collector/port.go，DataType 是 5051
	task := proto.Task{
		DataType:   5051, // 端口采集的数据类型
		ObjectName: "port",
		Data:       "", // collector 插件会自动采集，不需要额外数据
		Token:      "test-port-token-" + fmt.Sprintf("%d", time.Now().Unix()),
	}

	//发送任务给collector插件
	err := plg.SendTask(task)
	if err != nil {
		zap.S().Errorf("failed to send port task: %v", err)
	} else {
		zap.S().Info("port collection task sent successfully to collector plugin")
	}
}

// sendKmodTask 发送内核模块采集任务给 collector 插件
func sendKmodTask() {
	//获取collector插件的实例
	plg, ok := plugin.Get("collector")
	if !ok {
		zap.S().Error("collector plugin not found")
		return
	}

	// collector 插件使用 DataType 5062 来触发内核模块采集
	// 根据 collector/kmod.go，DataType 是 5062
	task := proto.Task{
		DataType:   5062, // 内核模块采集的数据类型
		ObjectName: "kmod",
		Data:       "", // collector 插件会自动采集，不需要额外数据
		Token:      "test-kmod-token-" + fmt.Sprintf("%d", time.Now().Unix()),
	}

	//发送任务给collector插件
	err := plg.SendTask(task)
	if err != nil {
		zap.S().Errorf("failed to send kmod task: %v", err)
	} else {
		zap.S().Info("kmod collection task sent successfully to collector plugin")
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
				fmt.Println("====================================")
				fmt.Println()
			}
		}
	} else if rec.DataType == 5051 {
		// 端口数据的数据类型是 5051
		zap.S().Infof("Data length: %d bytes", len(rec.Data))

		// 解析 protobuf Payload
		if len(rec.Data) > 0 {
			payload := &businessplugins.Payload{}
			err := payload.Unmarshal(rec.Data)
			if err != nil {
				zap.S().Errorf("Failed to unmarshal payload: %v", err)
			} else {
				fmt.Println("\n========== Port Record ==========")
				fmt.Printf("Protocol: %s", payload.Fields["protocol"])
				if payload.Fields["protocol"] == "6" {
					fmt.Print(" (TCP)")
				} else if payload.Fields["protocol"] == "17" {
					fmt.Print(" (UDP)")
				}
				fmt.Println()
				fmt.Printf("Family: %s", payload.Fields["family"])
				if payload.Fields["family"] == "2" {
					fmt.Print(" (IPv4)")
				} else if payload.Fields["family"] == "10" {
					fmt.Print(" (IPv6)")
				}
				fmt.Println()
				fmt.Printf("Local:  %s:%s\n", payload.Fields["sip"], payload.Fields["sport"])
				fmt.Printf("Remote: %s:%s\n", payload.Fields["dip"], payload.Fields["dport"])
				fmt.Printf("State: %s", payload.Fields["state"])
				if payload.Fields["state"] == "10" {
					fmt.Print(" (LISTEN)")
				} else if payload.Fields["state"] == "7" {
					fmt.Print(" (UDP)")
				}
				fmt.Println()
				fmt.Printf("UID: %s (%s)\n", payload.Fields["uid"], payload.Fields["username"])
				fmt.Printf("Inode: %s\n", payload.Fields["inode"])
				fmt.Println("=================================")
				fmt.Println()
			}
		}
	} else if rec.DataType == 5062 {
		// 内核模块数据的数据类型是 5062
		zap.S().Infof("Data length: %d bytes", len(rec.Data))

		// 解析 protobuf Payload
		if len(rec.Data) > 0 {
			payload := &businessplugins.Payload{}
			err := payload.Unmarshal(rec.Data)
			if err != nil {
				zap.S().Errorf("Failed to unmarshal payload: %v", err)
			} else {
				fmt.Println("\n========== Kernel Module Record ==========")
				fmt.Printf("Name: %s\n", payload.Fields["name"])
				fmt.Printf("Size: %s bytes\n", payload.Fields["size"])
				fmt.Printf("RefCount: %s\n", payload.Fields["refcount"])
				fmt.Printf("Used By: %s\n", payload.Fields["used_by"])
				fmt.Printf("State: %s\n", payload.Fields["state"])
				fmt.Printf("Address: %s\n", payload.Fields["addr"])
				fmt.Println("==========================================")
				fmt.Println()
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
				fmt.Println("================================")
				fmt.Println()
			}
		}
	}
	zap.S().Info("========================")
}

// newJSONWriter 创建新的 JSON 文件写入器
func newJSONWriter(filename string) (*jsonWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}

	// 写入 JSON 数组的开始标记（如果是新文件）
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// 如果文件为空，写入数组开始标记
	if stat.Size() == 0 {
		if _, err := file.WriteString("[\n"); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write array start: %w", err)
		}
	}

	return &jsonWriter{
		file: file,
		enc:  json.NewEncoder(file),
	}, nil
}

// Close 关闭 JSON 文件写入器
func (jw *jsonWriter) Close() error {
	jw.mu.Lock()
	defer jw.mu.Unlock()

	if jw.file == nil {
		return nil
	}

	// 写入数组结束标记
	if _, err := jw.file.WriteString("\n]"); err != nil {
		jw.file.Close()
		jw.file = nil
		return err
	}

	err := jw.file.Close()
	jw.file = nil
	return err
}

// writeRecordToJSON 将记录写入 JSON 文件
func writeRecordToJSON(rec *proto.EncodedRecord) {
	if jsonWriterInst == nil {
		return
	}

	jsonWriterInst.mu.Lock()
	defer jsonWriterInst.mu.Unlock()

	// 构建 JSON 记录结构
	record := map[string]interface{}{
		"data_type": rec.DataType,
		"timestamp": rec.Timestamp,
	}

	// 解析 Payload（如果存在）
	if len(rec.Data) > 0 {
		payload := &businessplugins.Payload{}
		if err := payload.Unmarshal(rec.Data); err == nil {
			record["data"] = payload.Fields
		} else {
			// 如果解析失败，将原始数据作为 base64 字符串存储
			record["data_raw"] = fmt.Sprintf("%x", rec.Data)
		}
	}

	// 检查文件位置，如果不是第一个记录，需要添加逗号
	stat, err := jsonWriterInst.file.Stat()
	if err == nil && stat.Size() > 2 { // 大于 "[\n" 的长度
		if _, err := jsonWriterInst.file.WriteString(",\n"); err != nil {
			zap.S().Errorf("Failed to write comma separator: %v", err)
			return
		}
	}

	// 写入 JSON 记录
	if err := jsonWriterInst.enc.Encode(record); err != nil {
		zap.S().Errorf("Failed to encode record to JSON: %v", err)
		return
	}

	// 刷新缓冲区，确保数据写入磁盘
	if err := jsonWriterInst.file.Sync(); err != nil {
		zap.S().Warnf("Failed to sync JSON file: %v", err)
	}
}
