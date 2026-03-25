package container

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// ImagePackage 镜像软件包实体
type ImagePackage struct {
	ID             int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID        string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_imgpkg_composite,priority:1"`
	HostName       string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP         string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	ImageID        string          `json:"image_id" gorm:"column:image_id;not null;uniqueIndex:idx_asset_imgpkg_composite,priority:2;index"`
	ImageName      string          `json:"image_name" gorm:"column:image_name;not null"`
	PackageName    string          `json:"package_name" gorm:"column:package_name;not null;uniqueIndex:idx_asset_imgpkg_composite,priority:3"`
	PackageVersion *string         `json:"package_version,omitempty" gorm:"column:package_version"`
	PackageType    string          `json:"package_type" gorm:"column:package_type;not null"` // dpkg, rpm, apk
	OsVersion      *string         `json:"os_version,omitempty" gorm:"column:os_version"`
	CreatedAt      common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_image_package
func (ImagePackage) TableName() string {
	return "asset_image_package"
}
