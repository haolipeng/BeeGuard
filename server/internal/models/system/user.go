package system

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// User 系统用户实体
type User struct {
	ID            int64           `json:"id" gorm:"primaryKey;column:id"`           // 主键ID
	Username      string          `json:"username" gorm:"column:username;size:250"` // 用户名
	Passwd        string          `json:"passwd" gorm:"column:passwd;size:250"`     // 密码 - 确保json标签是小写
	Name          string          `json:"name" gorm:"column:name;size:250"`         // 姓名
	Role          string          `json:"role" gorm:"column:role;size:100"`         // 角色权限
	AccountStatus string          `json:"account_status" gorm:"column:account_status;size:50"` // 账号状态
	CreatedAt     common.DateTime `json:"created_at"`                                          // 创建时间
	UpdatedAt     common.DateTime `json:"updated_at"`                                          // 更新时间
}

// TableName 指定表名
func (User) TableName() string {
	return "systen_user"
}