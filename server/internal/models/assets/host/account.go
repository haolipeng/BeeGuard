package host

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Account 账号资产实体
type Account struct {
	ID            int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	AgentID       string           `json:"agent_id" gorm:"column:agent_id;not null;uniqueIndex:idx_asset_account_composite,priority:1"`
	HostName      string           `json:"host_name" gorm:"column:host_name;not null"`
	HostIP        string            `json:"host_ip" gorm:"column:host_ip;type:varchar(256);not null;default:''"`
	OsType        *string          `json:"os_type,omitempty" gorm:"column:os_type"`
	Name          string           `json:"name" gorm:"column:name;not null;uniqueIndex:idx_asset_account_composite,priority:2"`
	Uid           int32            `json:"uid" gorm:"column:uid;not null"`
	Status        int16            `json:"status" gorm:"column:status;not null;default:0"` // 0=正常 1=即将过期 2=已过期
	Permission    string           `json:"permission" gorm:"column:permission;not null"`
	LoginType     *string          `json:"login_type,omitempty" gorm:"column:login_type"`
	LastLoginTime *common.DateTime `json:"last_login_time,omitempty" gorm:"column:last_login_time"`
	CreatedAt     common.DateTime  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     common.DateTime  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名为 asset_account
func (Account) TableName() string {
	return "asset_account"
}
