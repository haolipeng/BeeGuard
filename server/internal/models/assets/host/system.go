package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// System 系统服务资产实体
type System struct {
	ID        int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID   string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_sysservice_composite,priority:1"`
	HostName  string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP    string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType    *string         `json:"os_type,omitempty" gorm:"column:os_type"`
	Name      string          `json:"name" gorm:"column:name;not null;uniqueIndex:idx_asset_sysservice_composite,priority:2"`
	Version   *string         `json:"version,omitempty" gorm:"column:version"`
	Status    string          `json:"status" gorm:"column:status;not null"`
	RunUser   string          `json:"run_user" gorm:"column:run_user;not null"`
	Path      string          `json:"path" gorm:"column:path;not null"`
	Describe  *string         `json:"describe,omitempty" gorm:"column:describe"`
	CreatedAt common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_system_service
func (System) TableName() string {
	return "asset_system_service"
}
