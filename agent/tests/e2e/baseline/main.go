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

	"github.com/haolipeng/BeeGuard/agent/buffer"
	"github.com/haolipeng/BeeGuard/agent/plugin"
	"github.com/haolipeng/BeeGuard/agent/proto"
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

	fmt.Println("=== Baseline Plugin Test ===")
	fmt.Println("Starting test agent...")

	wg := &sync.WaitGroup{}
	zap.S().Info("++++++++++++++++++++++++++++++running++++++++++++++++++++++++++++++")

	Context, Cancel := context.WithCancel(context.Background())

	// 启动 plugin daemon
	wg.Add(1)
	go plugin.Startup(Context, wg)

	// 等待插件守护进程启动
	time.Sleep(time.Second * 1)

	// 加载 baseline 插件
	baselineConfig := &proto.Config{
		Name:    "baseline",
		Type:    "binary",
		Version: "1.0.0",
		Sha256:  "", // 测试时可以为空
	}
	cfgs := map[string]*proto.Config{
		"baseline": baselineConfig,
	}
	err := plugin.Sync(cfgs)
	if err != nil {
		zap.S().Errorf("failed to load baseline plugin: %v", err)
		os.Exit(1)
	} else {
		zap.S().Info("baseline plugin loaded successfully")
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

	// 发送测试任务
	go func() {
		time.Sleep(time.Second * 3)
		sendTestTask()
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

	wg.Wait()
	fmt.Println("Test completed.")
}

// sendTestTask 发送测试任务给 baseline 插件
func sendTestTask() {
	plg, ok := plugin.Get("baseline")
	if !ok {
		zap.S().Error("baseline plugin not found")
		return
	}

	// 创建测试任务数据
	// Ubuntu 用 1400（config/linux/1400.yaml），CentOS 用 1200，Debian 用 1300；check_id_list 需与对应 yaml 中的 check_id 一致
	taskData := map[string]interface{}{
		"baseline_id":   1400, // Ubuntu 基线；改为 1200 可测 CentOS 基线
		"check_id_list": []int{1, 2, 3},
	}
	taskDataJSON, _ := json.Marshal(taskData)

	task := proto.Task{
		DataType:   100, // 任务类型
		ObjectName: "baseline",
		Data:       string(taskDataJSON),
		Token:      "test-token-123",
	}

	err := plg.SendTask(task)
	if err != nil {
		zap.S().Errorf("failed to send task: %v", err)
	} else {
		zap.S().Info("task sent successfully to baseline plugin")
	}
}

// toFloat64 从 JSON 解析的 interface{} 安全转为 float64
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	}
	return 0
}

// getStr 从 map[string]interface{} 安全取字符串
func getStr(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// printRecord 打印接收到的记录
func printRecord(rec *proto.EncodedRecord) {
	zap.S().Infof("=== Received Record ===")
	zap.S().Infof("DataType: %d", rec.DataType)
	zap.S().Infof("Timestamp: %d", rec.Timestamp)

	// 如果是 baseline 的结果（DataType 8000 或 8010）
	if rec.DataType == 8000 || rec.DataType == 8010 {
		zap.S().Infof("Data length: %d bytes", len(rec.Data))

		// 解析 protobuf Payload
		if len(rec.Data) > 0 {
			payload := &businessplugins.Payload{}
			err := payload.Unmarshal(rec.Data)
			if err != nil {
				zap.S().Errorf("Failed to unmarshal payload: %v", err)
			} else {
				zap.S().Infof("Payload Fields: %+v", payload.Fields)

				// 如果是基线检查结果（DataType 8000），解析并打印检测结果
				if rec.DataType == 8000 {
					if dataStr, ok := payload.Fields["data"]; ok {
						var baselineResult map[string]interface{}
						if err := json.Unmarshal([]byte(dataStr), &baselineResult); err == nil {
							fmt.Println("\n========== Baseline Check Result ==========")
							if v, ok := baselineResult["baseline_id"]; ok {
								fmt.Printf("Baseline ID: %.0f\n", toFloat64(v))
							}
							fmt.Printf("Status: %s\n", getStr(baselineResult, "status"))
							fmt.Printf("Token: %s\n", payload.Fields["token"])
							if msg := getStr(baselineResult, "msg"); msg != "" {
								fmt.Printf("Message: %s\n", msg)
							}
							checkList, _ := baselineResult["check_list"].([]interface{})
							fmt.Printf("Check Items Count: %d\n", len(checkList))
							for i, item := range checkList {
								itemMap, _ := item.(map[string]interface{})
								checkID := toFloat64(itemMap["check_id"])
								resultVal := toFloat64(itemMap["result"])
								resultStr := "PASS"
								if resultVal == 1 {
									resultStr = "FAIL"
								} else if resultVal == 2 {
									resultStr = "ERROR"
								}
								titleCn := getStr(itemMap, "title_cn")
								if titleCn == "" {
									titleCn = getStr(itemMap, "title")
								}
								fmt.Printf("  [%d] CheckID: %.0f, Result: %s, Title: %s\n",
									i+1, checkID, resultStr, titleCn)
								if msg := getStr(itemMap, "msg"); msg != "" {
									fmt.Printf("       Msg: %s\n", msg)
								}
							}
							fmt.Println("==========================================\n")
						}
					}
				}

				// 如果是任务状态（DataType 8010）
				if rec.DataType == 8010 {
					fmt.Println("\n========== Task Status ==========")
					fmt.Printf("Status: %s\n", payload.Fields["status"])
					fmt.Printf("Token: %s\n", payload.Fields["token"])
					fmt.Printf("Message: %s\n", payload.Fields["msg"])
					fmt.Println("================================\n")
				}
			}
		}
	}
	zap.S().Info("========================")
}
