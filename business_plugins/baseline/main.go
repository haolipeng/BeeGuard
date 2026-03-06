package main

import (
	"baseline/check"
	"baseline/infra"
	"encoding/json"
	"runtime"
	"strconv"
	"time"

	businessplugins "business_plugins/lib"
)

var (
	BaseLineDataType           = int32(8000)
	BaseLineTaskStatusDataType = int32(8010)
	TaskStatusSuccess          = "succeed"
	TaskStatusFailed           = "failed"
	pluginClient               *businessplugins.Client
)

// SendServer send result to server
func SendServer(retCheckInfo check.RetBaselineInfo, token string) (err error) {
	record := businessplugins.Record{}
	record.DataType = BaseLineDataType
	record.Timestamp = time.Now().Unix()

	dataInfo, err := json.Marshal(retCheckInfo)
	if err != nil {
		return err
	}

	payload := businessplugins.Payload{}
	field := make(map[string]string, 0)
	field["data"] = string(dataInfo)
	field["token"] = token
	field["template_name"] = retCheckInfo.TemplateName
	field["template_id"] = strconv.Itoa(retCheckInfo.TemplateId)
	field["baseline_id"] = retCheckInfo.BaselineId
	payload.Fields = field
	record.Data = &payload

	err = pluginClient.SendRecord(&record)
	if err != nil {
		return err
	}
	return nil
}

// TaskStatusSendServer send task result to server
func TaskStatusSendServer(status string, token string, msg string, baselineId string) {
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
	field["baseline_id"] = baselineId
	payload.Fields = field
	record.Data = &payload

	_ = pluginClient.SendRecord(&record)
}

func main() {
	runtime.GOMAXPROCS(4)
	pluginClient = businessplugins.New()

	go func() {
		for {
			// get result from leader
			pluginsTask, err := pluginClient.ReceiveTask()
			if err != nil {
				infra.Loger.Println("getTask error:", err.Error())
				break
			}
			go func() {
				// start baseline analysis
				retBaselineInfo, analysisErr := check.Analysis(pluginsTask.Data)

				// send request to server
				sendErr := SendServer(retBaselineInfo, pluginsTask.Token)
				if sendErr != nil {
					infra.Loger.Println("sendServer error:", sendErr)
				} else {
					infra.Loger.Println("sendServer success:", retBaselineInfo.BaselineId)
				}

				// report task result
				if analysisErr != nil {
					TaskStatusSendServer(TaskStatusFailed, pluginsTask.Token, analysisErr.Error(), retBaselineInfo.BaselineId)
				} else {
					TaskStatusSendServer(TaskStatusSuccess, pluginsTask.Token, "", retBaselineInfo.BaselineId)
				}
			}()
		}
	}()

	select {}
}
