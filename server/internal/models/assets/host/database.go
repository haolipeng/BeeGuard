package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Database 数据库资产实体
type Database struct {
	ID        int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID   string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_database_composite,priority:1"`
	HostName  string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP    string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType    *string         `json:"os_type,omitempty" gorm:"column:os_type"`
	DbType    string          `json:"db_type" gorm:"column:db_type;not null;uniqueIndex:idx_asset_database_composite,priority:2"`
	DbVersion string          `json:"db_version" gorm:"column:db_version;not null"`
	Port      int32           `json:"port" gorm:"column:port;not null"`
	RunUser   *string         `json:"run_user,omitempty" gorm:"column:run_user"`
	CreatedAt common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_database
func (Database) TableName() string {
	return "asset_database"
}
