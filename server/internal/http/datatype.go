package http

import "shared/datatype"

// Agent 级别命令
const (
	DataTypeAgentShutdown  int32 = datatype.AgentShutdown
	DataTypeAgentUninstall int32 = datatype.AgentUninstall
)

// Collector 采集任务
const (
	DataTypeProcess       int32 = datatype.Process
	DataTypePort          int32 = datatype.Port
	DataTypeUser          int32 = datatype.User
	DataTypeService       int32 = datatype.Service
	DataTypeSoftware      int32 = datatype.Software
	DataTypeContainer     int32 = datatype.Container
	DataTypeEnvSuspicious int32 = datatype.EnvSuspicious
	DataTypeKmod          int32 = datatype.Kmod
)

// 响应类型
const (
	DataTypeTaskResult int32 = datatype.TaskResult
)

// Detector 检测器配置
const (
	DataTypeDetectorConfigUpdate int32 = datatype.DetectorConfigUpdate
)

// Scanner 扫描任务
const (
	DataTypeScannerScan int32 = datatype.ScannerDirScan
)

// Baseline 基线检测
const (
	DataTypeBaselineCheck int32 = datatype.BaselineCheck
)

// 目标对象名称
const (
	ObjectNameAgent     = "cloudsec-agent" // Agent 自身
	ObjectNameCollector = "collector"      // Collector 插件
	ObjectNameBaseline  = "baseline"       // Baseline 插件
	ObjectNameScanner   = "scanner"        // Scanner 插件
)

// validTaskDataTypes 有效的任务 DataType 集合
var validTaskDataTypes = map[int32]bool{
	DataTypeAgentShutdown:        true,
	DataTypeAgentUninstall:       true,
	DataTypeProcess:              true,
	DataTypePort:                 true,
	DataTypeUser:                 true,
	DataTypeService:              true,
	DataTypeSoftware:             true,
	DataTypeContainer:            true,
	DataTypeEnvSuspicious:        true,
	DataTypeKmod:                 true,
	DataTypeDetectorConfigUpdate: true,
	DataTypeScannerScan:          true,
	DataTypeBaselineCheck:        true,
}

// agentDataTypes Agent 级别的 DataType
var agentDataTypes = map[int32]bool{
	DataTypeAgentShutdown:  true,
	DataTypeAgentUninstall: true,
}

// IsValidTaskDataType 验证是否为有效的任务 DataType
func IsValidTaskDataType(dt int32) bool {
	return validTaskDataTypes[dt]
}

// IsAgentDataType 验证是否为 Agent 级别的 DataType
func IsAgentDataType(dt int32) bool {
	return agentDataTypes[dt]
}

// GetDataTypeName 获取 DataType 的名称
func GetDataTypeName(dt int32) string {
	switch dt {
	case DataTypeAgentShutdown:
		return "AgentShutdown"
	case DataTypeAgentUninstall:
		return "AgentUninstall"
	case DataTypeProcess:
		return "Process"
	case DataTypePort:
		return "Port"
	case DataTypeUser:
		return "User"
	case DataTypeService:
		return "Service"
	case DataTypeSoftware:
		return "Software"
	case DataTypeContainer:
		return "Container"
	case DataTypeEnvSuspicious:
		return "EnvSuspicious"
	case DataTypeKmod:
		return "Kmod"
	case DataTypeTaskResult:
		return "TaskResult"
	case DataTypeDetectorConfigUpdate:
		return "DetectorConfigUpdate"
	case DataTypeBaselineCheck:
		return "BaselineCheck"
	case DataTypeScannerScan:
		return "ScannerScan"
	default:
		return "Unknown"
	}
}
