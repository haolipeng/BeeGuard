package businessplugins

// 安全告警类型常量
// 所有安全告警的 DataType 统一在此定义，各插件通过引用这些常量来设置告警类型。
const (
	AlertTypeSSHBruteForce       = 6001
	AlertTypeFTPBruteForce       = 6002
	AlertTypeDangerousCommand    = 6003
	AlertTypeReverseShell        = 6004
	AlertTypeSSHAnomalyLogin     = 6005
	AlertTypePrivilegeEscalation = 6006
	AlertTypeNIDS                = 6007
	AlertTypeMaliciousRequest    = 6008
	AlertTypeSensitiveFile       = 6009

	// 容器安全告警
	AlertTypeContainerDangerousCommand = 7001
	AlertTypeContainerEscape           = 7002
	AlertTypeContainerReverseShell     = 7003
	AlertTypeContainerSensitiveFile    = 7004
)
