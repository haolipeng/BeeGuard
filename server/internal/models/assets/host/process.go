package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Process 进程资产实体
type Process struct {
	ID        int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID   string           `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_process_composite,priority:1"`
	HostName  string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP    string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType    *string          `json:"os_type,omitempty" gorm:"column:os_type"`
	Name      string           `json:"name" gorm:"column:name;not null"`
	Status    *string          `json:"status,omitempty" gorm:"column:status"`
	Version   *string          `json:"version,omitempty" gorm:"column:version"`
	Path      string           `json:"path" gorm:"column:path;not null;uniqueIndex:idx_asset_process_composite,priority:2"`
	RunName   string           `json:"run_name" gorm:"column:run_name;not null"`
	StartTime *common.DateTime `json:"start_time,omitempty" gorm:"column:start_time"`
	CreatedAt common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_process
func (Process) TableName() string {
	return "asset_process"
}
