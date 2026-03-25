package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Host 主机资产实体
type Host struct {
	ID        int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID   string  `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex"`
	HostName  string  `json:"host_name" gorm:"column:host_name;not null"`
	HostIP    string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	MacAddr   *string `json:"mac_addr,omitempty" gorm:"column:mac_addr"`
	OsType    *string `json:"os_type,omitempty" gorm:"column:os_type"`
	OsVersion *string `json:"os_version,omitempty" gorm:"column:os_version"`
	AgentStatus   int16            `json:"agent_status" gorm:"column:agent_status;default:0"`
	AgentVersion  *string          `json:"agent_version,omitempty" gorm:"column:agent_version"`
	LastHeartbeat *common.DateTime `json:"last_heartbeat,omitempty" gorm:"column:last_heartbeat"`
	CreatedAt     common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_host
func (Host) TableName() string {
	return "asset_host"
}
