package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Software 软件资产实体
type Software struct {
	ID        int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID   string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_software_composite,priority:1"`
	HostName  string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP    string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType    *string         `json:"os_type,omitempty" gorm:"column:os_type"`
	Name      string          `json:"name" gorm:"column:name;not null;uniqueIndex:idx_asset_software_composite,priority:2"`
	Version   *string         `json:"version,omitempty" gorm:"column:version"`
	Type      string          `json:"type" gorm:"column:type;not null;uniqueIndex:idx_asset_software_composite,priority:3"` // dpkg, rpm, pypi, jar
	Source    *string         `json:"source,omitempty" gorm:"column:source"`
	Status    *string         `json:"status,omitempty" gorm:"column:status"`
	Vendor    *string         `json:"vendor,omitempty" gorm:"column:vendor"`
	Path      *string         `json:"path,omitempty" gorm:"column:path"`
	CreatedAt common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_software
func (Software) TableName() string {
	return "asset_software"
}
