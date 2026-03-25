package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Web Web服务资产实体
type Web struct {
	ID         int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID    string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_webservice_composite,priority:1"`
	HostName   string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP     string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType     *string         `json:"os_type,omitempty" gorm:"column:os_type"`
	Name       string          `json:"name" gorm:"column:name;not null"`
	Version    string          `json:"version" gorm:"column:version;not null"`
	ServerType string          `json:"server_type" gorm:"column:server_type;not null;uniqueIndex:idx_asset_webservice_composite,priority:2"`
	SiteDomain *string         `json:"site_domain,omitempty" gorm:"column:site_domain"`
	Path       *string         `json:"path,omitempty" gorm:"column:path"`
	CreatedAt  common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_web_service
func (Web) TableName() string {
	return "asset_web_service"
}
