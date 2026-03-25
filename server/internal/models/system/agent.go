package system

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// AgentInfo Agent客户端信息实体
type AgentInfo struct {
	ID              int64           `json:"id" gorm:"primaryKey;not null"`              // 主键ID
	AgentID         string          `json:"agent_id" gorm:"not null;default:''"`       // Agent唯一标识(如AGT-20251225-001)
	AgentVersion    *string         `json:"agent_version,omitempty"`                     // 安装版本(如2.1.5)
	ConnectionStatus int16          `json:"connection_status" gorm:"not null;default:0"` // 连接状态: 0-已断开 1-已连接
	HostName        string          `json:"host_name" gorm:"not null"`                  // 主机名
	HostIP          string            `json:"host_ip" gorm:"type:varchar(256);not null;default:''"` // IP地址列表(逗号分隔)
	OSType          string          `json:"os_type" gorm:"not null"`                    // 操作系统类型: linux/windows
	OSVersion       *string         `json:"os_version,omitempty"`                        // 操作系统版本(如Ubuntu 20.04.3 LTS)
	OSArch          *string         `json:"os_arch,omitempty"`                           // CPU架构(如x86_64, aarch64)
	CPUCount        *int32          `json:"cpu_count,omitempty"`                         // CPU核数
	MemoryTotal     *int64          `json:"memory_total,omitempty"`                      // 内存总量(字节)
	DiskTotal       *int64          `json:"disk_total,omitempty"`                        // 磁盘总量(字节)
	LastConnectedAt *common.DateTime `json:"last_connected_at,omitempty"`                // 最后连接时间
	RegisteredAt    common.DateTime `json:"registered_at" gorm:"not null"`              // Agent首次注册时间
	CreatedAt       common.DateTime `json:"created_at"`                                  // 创建时间
	UpdatedAt       common.DateTime `json:"updated_at"`                                  // 更新时间
}

// TableName 指定表名为 agent_info
func (AgentInfo) TableName() string {
	return "agent_info"
}

// ConnectionStatus 连接状态枚举常量
const (
	ConnectionStatusDisconnected = 0 // 已断开
	ConnectionStatusConnected    = 1 // 已连接
)

// OSType 操作系统类型枚举常量
const (
	OSTypeLinux   = "linux"
	OSTypeWindows = "windows"
)
