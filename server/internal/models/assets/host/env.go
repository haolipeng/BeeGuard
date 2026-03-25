package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// EnvSuspicious 可疑环境变量实体
type EnvSuspicious struct {
	ID                int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID           string          `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_env_composite,priority:1"`
	HostName          string          `json:"host_name" gorm:"column:host_name;not null"`
	HostIP            string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	VarName           string          `json:"var_name" gorm:"column:var_name;not null;uniqueIndex:idx_asset_env_composite,priority:2"`
	VarValue          *string         `json:"var_value,omitempty" gorm:"column:var_value"`
	SuspiciousReasons *string         `json:"suspicious_reasons,omitempty" gorm:"column:suspicious_reasons"`
	Source            *string         `json:"source,omitempty" gorm:"column:source"`
	CreatedAt         common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_env_suspicious
func (EnvSuspicious) TableName() string {
	return "asset_env_suspicious"
}
