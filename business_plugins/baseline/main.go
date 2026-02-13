package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"baseline/check"

	businessplugins "business_plugins/lib"
)

var (
	// BaseLineDataType 基线检查数据类型
	BaseLineDataType = int32(8000)
	// BaseLineTaskStatusDataType 任务状态数据类型
	BaseLineTaskStatusDataType = int32(8010)
	// TaskStatusSuccess 任务成功状态
	TaskStatusSuccess = "succeed"
	// TaskStatusFailed 任务失败状态
	TaskStatusFailed = "failed"
	// pluginClient 插件客户端
	pluginClient *businessplugins.Client
)

// SendServer 发送基线检查结果到 server
func SendServer(retCheckInfo check.RetBaselineInfo, token string) (err error) {
	record := businessplugins.Record{}
	record.DataType = BaseLineDataType
	record.Timestamp = time.Now().Unix()

	// 将检查结果序列化为 JSON
	dataInfo, err := json.Marshal(retCheckInfo)
	if err != nil {
		return err
	}

	// 创建 Payload
	payload := businessplugins.Payload{}
	field := make(map[string]string, 0)
	field["data"] = string(dataInfo)
	field["token"] = token
	payload.Fields = field
	record.Data = &payload

	// 发送记录
	err = pluginClient.SendRecord(&record)
	if err != nil {
		return err
	}
	return nil
}

// TaskStatusSendServer 发送任务状态到 server
func TaskStatusSendServer(status string, token string, msg string) {
	record := businessplugins.Record{}
	record.DataType = BaseLineTaskStatusDataType
	record.Timestamp = time.Now().Unix()

	payload := businessplugins.Payload{}
	field := make(map[string]string, 0)
	field["status"] = status
	if token != "" {
		field["token"] = token
	}
	field["msg"] = msg
	payload.Fields = field
	record.Data = &payload

	_ = pluginClient.SendRecord(&record)
}

func main() {
	// 设置日志输出
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		os.MkdirAll(logDir, 0755)
		logFile, err := os.OpenFile(filepath.Join(logDir, "baseline.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			log.SetOutput(io.MultiWriter(os.Stderr, logFile))
			defer logFile.Close()
		}
	}

	// 初始化插件客户端
	pluginClient = businessplugins.New()

	// 循环接收任务
	go func() {
		for {
			// 从 agent 接收任务
			pluginsTask, err := pluginClient.ReceiveTask()
			if err != nil {
				log.Printf("ReceiveTask error: %v\n", err)
				break
			}

			// 在 goroutine 中处理任务
			go func() {
				// 执行基线分析
				retBaselineInfo, analysisErr := check.Analysis(pluginsTask.Data)

				// 发送检查结果到 server
				err = SendServer(retBaselineInfo, pluginsTask.Token)
				if err != nil {
					log.Printf("SendServer error: %v\n", err)
				} else {
					log.Printf("SendServer success: baseline_id=%d\n", retBaselineInfo.BaselineId)
				}

				// 发送任务状态
				if analysisErr != nil {
					TaskStatusSendServer(TaskStatusFailed, pluginsTask.Token, analysisErr.Error())
				} else {
					TaskStatusSendServer(TaskStatusSuccess, pluginsTask.Token, "")
				}
			}()
		}
	}()

	// 保持主程序运行
	select {}
}
