package alert

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// DangerousCommand 高危命令告警实体
type DangerousCommand struct {
	ID             int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID        string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostID         *int64          `json:"host_id,omitempty" gorm:"column:host_id"`
	HostName       string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP         string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	Command        string          `json:"command" gorm:"column:command;not null"`
	CommandType    string          `json:"command_type" gorm:"column:command_type;not null;index"`
	User           string          `json:"user" gorm:"column:user;not null"`
	PrivilegeLevel string          `json:"privilege_level" gorm:"column:privilege_level;not null"`
	Status         int16           `json:"status" gorm:"column:status;not null;default:0"` // 0-待处理 1-已处理 2-已忽略
	AlertTime      common.DateTime `json:"alert_time" gorm:"column:alert_time;not null"`
	CreatedAt      common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 alert_dangerous_command
func (DangerousCommand) TableName() string {
	return "alert_dangerous_command"
}

// CommandType 命令类型枚举常量
const (
	CmdTypeFileDelete          = "file_delete"          // 文件删除
	CmdTypePrivilegeEscalation = "privilege_escalation" // 权限提升
	CmdTypePermissionModify    = "permission_modify"    // 权限修改
	CmdTypeFilesystemOperation = "filesystem_operation" // 文件系统操作
	CmdTypeNetworkScan         = "network_scan"         // 网络扫描
	CmdTypeDataExfiltration    = "data_exfiltration"    // 数据外传
	CmdTypeServiceStop         = "service_stop"         // 服务停止
	CmdTypeLogTamper           = "log_tamper"           // 日志篡改
)

// PrivilegeLevel 权限级别枚举常量
const (
	PrivilegeRoot  = "root"   // root权限
	PrivilegeAdmin = "admin"  // 管理员权限
	PrivilegeUser  = "user"   // 普通用户权限
)

// Status 状态枚举常量
const (
	StatusPending   = 0 // 待处理
	StatusProcessed = 1 // 已处理
	StatusIgnored   = 2 // 已忽略
)