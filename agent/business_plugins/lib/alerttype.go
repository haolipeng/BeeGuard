package businessplugins

import "shared/datatype"

// 安全告警类型常量
// 所有安全告警的 DataType 统一在此定义，各插件通过引用这些常量来设置告警类型。
const (
	AlertTypeSSHBruteForce       = datatype.AlertSSHBruteForce
	AlertTypeFTPBruteForce       = datatype.AlertFTPBruteForce
	AlertTypeDangerousCommand    = datatype.AlertDangerousCommand
	AlertTypeReverseShell        = datatype.AlertReverseShell
	AlertTypeSSHAnomalyLogin     = datatype.AlertSSHAnomalyLogin
	AlertTypePrivilegeEscalation = datatype.AlertPrivilegeEscalation
	AlertTypeNIDS                = datatype.AlertNIDS
	AlertTypeMaliciousRequest    = datatype.AlertMaliciousRequest
	AlertTypeSensitiveFile       = datatype.AlertSensitiveFile

	// 容器安全告警
	AlertTypeContainerDangerousCommand = datatype.AlertContainerDangerousCommand
	AlertTypeContainerEscape           = datatype.AlertContainerEscape
	AlertTypeContainerReverseShell     = datatype.AlertContainerReverseShell
	AlertTypeContainerSensitiveFile    = datatype.AlertContainerSensitiveFile
)
