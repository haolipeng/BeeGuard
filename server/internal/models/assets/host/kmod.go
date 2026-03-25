package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Kmod 内核模块资产实体
type Kmod struct {
	ID        int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID   string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_kmod_composite,priority:1"`
	HostName  string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP    string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType    *string         `json:"os_type,omitempty" gorm:"column:os_type"`
	Name      string          `json:"name" gorm:"column:name;not null;uniqueIndex:idx_asset_kmod_composite,priority:2"`
	Size      *string         `json:"size,omitempty" gorm:"column:size"`
	RefCount  *string         `json:"refcount,omitempty" gorm:"column:refcount"`
	UsedBy    *string         `json:"used_by,omitempty" gorm:"column:used_by"`
	State     *string         `json:"state,omitempty" gorm:"column:state"`
	Addr      *string         `json:"addr,omitempty" gorm:"column:addr"`
	CreatedAt common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_kmod
func (Kmod) TableName() string {
	return "asset_kmod"
}
