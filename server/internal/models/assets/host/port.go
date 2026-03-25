package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Port 端口资产实体
type Port struct {
	ID            int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID       string           `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_port_composite,priority:1"`
	HostName      string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP        string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType        *string          `json:"os_type,omitempty" gorm:"column:os_type"`
	Port          int32            `json:"port" gorm:"column:port;not null;uniqueIndex:idx_asset_port_composite,priority:2"`
	Protocol      int16            `json:"protocol" gorm:"column:protocol;not null;uniqueIndex:idx_asset_port_composite,priority:3"` // 6=TCP, 17=UDP
	ListenIP      string           `json:"listen_ip" gorm:"column:listen_ip;not null"`
	ListenProcess string           `json:"listen_process" gorm:"column:listen_process;not null"`
	RunUser       *string          `json:"run_user,omitempty" gorm:"column:run_user"`
	OsVersion     *string          `json:"os_version,omitempty" gorm:"column:os_version"`
	AgentStatus   int16            `json:"agent_status" gorm:"column:agent_status;default:0"`
	AgentVersion  *string          `json:"agent_version,omitempty" gorm:"column:agent_version"`
	ProcessTime   *common.DateTime `json:"process_time,omitempty" gorm:"column:process_time"`
	CreatedAt     common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_port
func (Port) TableName() string {
	return "asset_port"
}
