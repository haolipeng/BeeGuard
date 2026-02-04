package standalone

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/buffer"
	"gitlab.myinterest.top/security/agent/config"
	"gitlab.myinterest.top/security/agent/proto"
	"go.uber.org/zap"
)

// DetectionOutput 高危命令检测结果输出结构
type DetectionOutput struct {
	Timestamp  int64             `json:"timestamp"`
	DataType   int32             `json:"data_type"`
	RuleID     string            `json:"rule_id"`
	RuleName   string            `json:"rule_name"`
	Severity   string            `json:"severity"`
	Command    string            `json:"command"`
	Pattern    string            `json:"matched_pattern,omitempty"`
	PID        string            `json:"pid,omitempty"`
	UID        string            `json:"uid,omitempty"`
	ExePath    string            `json:"exe_path,omitempty"`
	AllFields  map[string]string `json:"all_fields,omitempty"`
}

// StartOutputHandler 启动 standalone 模式的输出处理
func StartOutputHandler(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	zap.S().Info("standalone output handler startup")

	cfg, err := config.Get()
	if err != nil {
		zap.S().Errorf("failed to get config: %v", err)
		return
	}

	standaloneCfg := cfg.Standalone
	if standaloneCfg == nil {
		zap.S().Error("standalone config is nil")
		return
	}

	interval := time.Duration(standaloneCfg.FlushInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var outputFile *os.File
	if standaloneCfg.Output == "file" {
		outputFile, err = os.OpenFile(standaloneCfg.OutputPath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			zap.S().Errorf("failed to open output file: %v", err)
			return
		}
		defer outputFile.Close()
		zap.S().Infof("detection results will be written to: %s", standaloneCfg.OutputPath)
	} else {
		zap.S().Info("detection results will be logged")
	}

	for {
		select {
		case <-ctx.Done():
			// 退出前处理剩余数据
			recs := buffer.ReadEncodedRecords()
			processRecords(recs, standaloneCfg, outputFile)
			zap.S().Info("standalone output handler exiting")
			return
		case <-ticker.C:
			recs := buffer.ReadEncodedRecords()
			processRecords(recs, standaloneCfg, outputFile)
		}
	}
}

// processRecords 处理检测记录
func processRecords(recs []*proto.EncodedRecord, cfg *config.StandaloneConfig, file *os.File) {
	if len(recs) == 0 {
		return
	}

	for _, rec := range recs {
		// 尝试解析 Payload
		payload := parsePayload(rec.Data)
		if payload == nil {
			continue
		}

		// 仅输出高危命令检测结果（有 rule_id 字段）
		ruleID, ok := payload.Fields["rule_id"]
		if !ok || ruleID == "" {
			continue
		}

		output := buildOutput(rec, payload)

		switch cfg.Output {
		case "log":
			logRecord(output)
		case "file":
			writeToFile(file, output)
		default:
			logRecord(output)
		}
	}
}

// parsePayload 解析 protobuf payload
func parsePayload(data []byte) *businessplugins.Payload {
	if len(data) == 0 {
		return nil
	}

	payload := &businessplugins.Payload{}
	if err := payload.Unmarshal(data); err != nil {
		// 静默忽略解析错误，可能是非标准格式的数据
		return nil
	}

	return payload
}

// buildOutput 构建输出结构
func buildOutput(rec *proto.EncodedRecord, payload *businessplugins.Payload) *DetectionOutput {
	fields := payload.Fields

	return &DetectionOutput{
		Timestamp:  rec.Timestamp,
		DataType:   rec.DataType,
		RuleID:     fields["rule_id"],
		RuleName:   fields["rule_name"],
		Severity:   fields["severity"],
		Command:    fields["command"],
		Pattern:    fields["matched_pattern"],
		PID:        fields["pid"],
		UID:        fields["uid"],
		ExePath:    fields["exe_path"],
		AllFields:  fields,
	}
}

// logRecord 将记录输出到日志
func logRecord(output *DetectionOutput) {
	zap.S().Infow("dangerous command detected",
		"rule_id", output.RuleID,
		"rule_name", output.RuleName,
		"severity", output.Severity,
		"command", output.Command,
		"matched_pattern", output.Pattern,
		"pid", output.PID,
		"uid", output.UID,
	)
}

// writeToFile 将记录写入 JSON 文件
func writeToFile(file *os.File, output *DetectionOutput) {
	data, err := json.Marshal(output)
	if err != nil {
		zap.S().Errorf("failed to marshal output: %v", err)
		return
	}
	file.Write(append(data, '\n'))
}
