package scanner

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	businessplugins "business_plugins/lib"
)

// MalwareResult 恶意软件检出结果
type MalwareResult struct {
	ThreatType    string // 威胁类型（Trojan/Webshell/Miner 等）
	FileName      string // 文件名
	FilePath      string // 完整路径
	FileSize      int64  // 字节数
	FileMD5       string // MD5
	FileSHA256    string // SHA256
	MalwareFamily string // 恶意软件家族名
	ScanTime      int64  // Unix 时间戳
}

// DataType 常量
const (
	// 任务接收（控制台 → 插件）
	DataTypeDBUpdate   = 6050 // 病毒库更新
	DataTypeDirScan    = 6053 // 指定目录扫描
	DataTypeFullScan   = 6057 // 全盘扫描

	// 结果上报（插件 → Agent → Server）
	DataTypeScanStatus = 6060 // 扫描任务状态
	DataTypeFileDetect = 6061 // 静态文件检出
	DataTypeProcDetect = 6062 // 进程 EXE 检出
)

// ToRecord 将 MalwareResult 转换为 Protobuf Record
func (r *MalwareResult) ToRecord(dataType int32) *businessplugins.Record {
	return &businessplugins.Record{
		DataType:  dataType,
		Timestamp: time.Now().UnixMilli(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"threat_type":      r.ThreatType,
				"file_name":        r.FileName,
				"file_path":        r.FilePath,
				"file_size":        strconv.FormatInt(r.FileSize, 10),
				"file_md5":         r.FileMD5,
				"file_sha256":      r.FileSHA256,
				"detection_engine": "ClamAV",
				"malware_family":   r.MalwareFamily,
				"scan_time":        strconv.FormatInt(r.ScanTime, 10),
			},
		},
	}
}

// NewStatusRecord 创建扫描状态上报 Record
func NewStatusRecord(status, msg string) *businessplugins.Record {
	return &businessplugins.Record{
		DataType:  DataTypeScanStatus,
		Timestamp: time.Now().UnixMilli(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"status": status,
				"msg":    msg,
			},
		},
	}
}

// ParseVirusName 解析 ClamAV 病毒名
// 格式：Type.Class.Name.UNOFFICIAL 或 Platform.Type.Name
// 示例：
//   "Linux.Trojan.Mirai"    → threatType="Trojan", malwareFamily="Mirai"
//   "Php.Webshell.eval"     → threatType="Webshell", malwareFamily="eval"
//   "Win.Trojan.Agent-123"  → threatType="Trojan", malwareFamily="Agent-123"
func ParseVirusName(virusName string) (threatType, malwareFamily string) {
	// 去除 UNOFFICIAL 后缀
	name := strings.TrimSuffix(virusName, ".UNOFFICIAL")

	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		return "Malware", virusName
	}

	// 尝试从 parts 中找到威胁类型
	threatType = "Malware"
	malwareFamily = parts[len(parts)-1]

	for _, p := range parts {
		lower := strings.ToLower(p)
		switch lower {
		case "trojan":
			threatType = "Trojan"
		case "webshell":
			threatType = "Webshell"
		case "miner", "coinminer":
			threatType = "Miner"
		case "backdoor":
			threatType = "Backdoor"
		case "ransomware", "ransom":
			threatType = "Ransomware"
		case "worm":
			threatType = "Worm"
		case "rootkit":
			threatType = "Rootkit"
		case "exploit":
			threatType = "Exploit"
		case "adware":
			threatType = "Adware"
		case "downloader":
			threatType = "Downloader"
		case "dropper":
			threatType = "Dropper"
		}
	}

	return threatType, malwareFamily
}

// FormatResult 格式化检出结果（用于日志输出）
func FormatResult(r *MalwareResult) string {
	return fmt.Sprintf("[%s] %s (%s) - %s md5=%s",
		r.ThreatType, r.FilePath, r.MalwareFamily, formatSize(r.FileSize), r.FileMD5)
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	switch {
	case size >= 1024*1024:
		return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	default:
		return fmt.Sprintf("%dB", size)
	}
}
