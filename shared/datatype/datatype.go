package datatype

// Agent 命令 (1060-1099)
const (
	AgentShutdown  = 1060
	AgentUninstall = 1061
)

// 资产采集 (5050-5099)
const (
	Process       = 5050
	Port          = 5051
	User          = 5052
	Service       = 5054
	Software      = 5055
	Container     = 5056
	EnvSuspicious = 5057
	Image         = 5058
	ImagePackage  = 5059
	WebService    = 5060
	Database      = 5061
	Kmod          = 5062
)

// 任务结果
const (
	TaskResult = 5100
)

// eBPF 实时事件
const (
	EventExecve   = 59
	EventConnect  = 60
	EventBind     = 61
	EventAccept   = 62
	EventDNS      = 63
	EventFile     = 64
	EventMount    = 65
	EventPerfLoss = 66
)

// 安全告警 (6001-6099)
const (
	AlertSSHBruteForce       = 6001
	AlertFTPBruteForce       = 6002
	AlertDangerousCommand    = 6003
	AlertReverseShell        = 6004
	AlertSSHAnomalyLogin     = 6005
	AlertPrivilegeEscalation = 6006
	AlertNIDS                = 6007
	AlertMaliciousRequest    = 6008
	AlertSensitiveFile       = 6009
)

// 检测器配置
const (
	DetectorConfigUpdate = 6010
)

// 病毒扫描 (6050-6069)
const (
	ScannerDBUpdate   = 6050
	ScannerDirScan    = 6053
	ScannerFullScan   = 6057
	ScannerScanStatus = 6060
	ScannerFileDetect = 6061
	ScannerProcDetect = 6062
)

// 容器安全告警 (7001-7099)
const (
	AlertContainerDangerousCommand = 7001
	AlertContainerEscape           = 7002
	AlertContainerReverseShell     = 7003
	AlertContainerSensitiveFile    = 7004
)

// 基线检测 (8000-8099)
const (
	BaselineCheck      = 8000
	BaselineTaskStatus = 8010
)
