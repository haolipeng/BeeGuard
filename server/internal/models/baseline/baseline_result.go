package baseline

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// CheckResult 基线检查结果表
type CheckResult struct {
	ID          int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	BaselineID  string          `json:"baseline_id" gorm:"column:baseline_id;not null;index"`
	TemplateID  int64           `json:"template_id" gorm:"column:template_id;index"`
	AgentID     string          `json:"agent_id" gorm:"column:agent_id;not null;index"`
	HostIP      string          `json:"host_ip" gorm:"column:host_ip;not null"`
	HostName    string          `json:"host_name" gorm:"column:host_name;not null"`
	TotalItems  int             `json:"total_items" gorm:"column:total_items;default:0"`
	PassedItems int             `json:"passed_items" gorm:"column:passed_items;default:0"`
	FailedItems int             `json:"failed_items" gorm:"column:failed_items;default:0"`
	ErrorItems  int             `json:"error_items" gorm:"column:error_items;default:0"`
	CheckTime   common.DateTime `json:"check_time" gorm:"column:check_time"`
	CreatedAt   common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (CheckResult) TableName() string {
	return "baseline_check_result"
}
