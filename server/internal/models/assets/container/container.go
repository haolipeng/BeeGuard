package container

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Container 容器资产实体
type Container struct {
	ID          int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID     string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_container_composite,priority:1"`
	HostName    string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP      string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	ContainerID string          `json:"container_id" gorm:"column:container_id;not null;uniqueIndex:idx_asset_container_composite,priority:2"`
	Name        string          `json:"name" gorm:"column:name;not null"`
	State       string          `json:"state" gorm:"column:state;not null"`
	ImageID     *string         `json:"image_id,omitempty" gorm:"column:image_id"`
	ImageName   *string         `json:"image_name,omitempty" gorm:"column:image_name"`
	Runtime     *string         `json:"runtime,omitempty" gorm:"column:runtime"`
	Pid         *string         `json:"pid,omitempty" gorm:"column:pid"`
	CreateTime  *string         `json:"create_time,omitempty" gorm:"column:create_time"`
	CreatedAt   common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_container
func (Container) TableName() string {
	return "asset_container"
}
