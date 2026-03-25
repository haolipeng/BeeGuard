package container

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Image 镜像资产实体
type Image struct {
	ID             int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID        string           `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:uk_asset_image_agent_imgid,priority:1"`
	HostName       string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP         string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	ImageID        string           `json:"image_id" gorm:"column:image_id;not null;uniqueIndex:uk_asset_image_agent_imgid,priority:2"`
	ImageName      string           `json:"image_name" gorm:"column:image_name;not null"`
	ImageVersion   *string          `json:"image_version,omitempty" gorm:"column:image_version"`
	ImageSize      *int64           `json:"image_size,omitempty" gorm:"column:image_size"`
	ContainerCount *int32           `json:"container_count,omitempty" gorm:"column:container_count;default:0"`
	BuildTime      *common.DateTime `json:"build_time,omitempty" gorm:"column:build_time"`
	Runtime        *string          `json:"runtime,omitempty" gorm:"column:runtime"`
	CreatedAt      common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_image
func (Image) TableName() string {
	return "asset_image"
}
