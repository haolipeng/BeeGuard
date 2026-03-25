package baseline

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// BaselineCheckDetail 基线检查结果明细实体
type BaselineCheckDetail struct {
	ID            int64           `json:"id" gorm:"primaryKey;not null"`                                             // 结果 ID
	TemplateID    int32           `json:"template_id" gorm:"primaryKey;not null;index:idx_bcd_template_id"`          // 基线模板 ID(冗余，复合主键)
	ResultID      int64           `json:"result_id" gorm:"not null;index:idx_bcd_result_id"`                         // 关联检查结果 ID
	ItemID        int64           `json:"item_id" gorm:"not null;index:idx_bcd_item_id"`                             // 关联检查项 ID
	ItemName      *string         `json:"item_name,omitempty" gorm:"column:item_name"`                               // 检查项名称
	BaselineID    *string         `json:"baseline_id,omitempty" gorm:"column:baseline_id;index:idx_bcd_baseline_id"` // 检测批次 ID（前端 task_id）
	AgentID       string          `json:"agent_id" gorm:"column:agent_id;not null;index:idx_bcd_agent_id"`           // Agent 唯一标识
	HostIP        *string         `json:"host_ip,omitempty" gorm:"column:host_ip"`                                   // 主机 IP(冗余)
	HostName      *string         `json:"host_name,omitempty" gorm:"column:host_name"`                               // 主机名称 (冗余)
	TemplateName  *string         `json:"template_name,omitempty" gorm:"column:template_name"`                       // 模板名称 (冗余)
	RiskLevel     *string         `json:"risk_level,omitempty" gorm:"column:risk_level"`                             // 风险等级
	Status        int16           `json:"status" gorm:"not null;index:idx_bcd_status"`                               // 检查状态：0-未通过 1-通过 2-检查异常
	ActualValue   *string         `json:"actual_value,omitempty" gorm:"type:text"`                                   // 实际值
	ExpectedValue *string         `json:"expected_value,omitempty" gorm:"type:text"`                                 // 期望值
	ErrorMessage  *string         `json:"error_message,omitempty" gorm:"size:512"`                                   // 错误信息
	CheckTime     common.DateTime `json:"check_time" gorm:"not null"`                                                // 检查时间
	CreatedAt     common.DateTime `json:"created_at"`                                                                // 创建时间
	UpdatedAt     common.DateTime `json:"updated_at"`                                                                // 更新时间
}

// TableName 指定表名为 baseline_check_detail
func (BaselineCheckDetail) TableName() string {
	return "baseline_check_detail"
}

// BaselineCheckHostView 基线检查结果主机统计视图实体
type BaselineCheckHostView struct {
	BaselineId    int8            `json:"baseline_id" gorm:"column:baseline_id"`         // 任务ID
	AgentID       string          `json:"agent_id" gorm:"column:agent_id"`               // Agent唯一标识
	HostName      string          `json:"host_name" gorm:"column:host_name"`             // 主机名称
	HostIP        string          `json:"host_ip" gorm:"column:host_ip"`                 // 主机IP
	TotalChecks   int64           `json:"total_checks" gorm:"column:total_checks"`       // 总检查项数
	PassedChecks  int64           `json:"passed_checks" gorm:"column:passed_checks"`     // 通过项数
	FailedChecks  int64           `json:"failed_checks" gorm:"column:failed_checks"`     // 未通过项数
	ErrorChecks   int64           `json:"error_checks" gorm:"column:error_checks"`       // 异常项数
	LastCheckTime common.DateTime `json:"last_check_time" gorm:"column:last_check_time"` // 最后检查时间

}

// TableName 指定视图名为 baseline_check_host_view
func (BaselineCheckHostView) TableName() string {
	return "baseline_check_host_view"
}

// BaselineCheckItemView 基线检查结果项统计视图实体
type BaselineCheckItemView struct {
	BaselineID   string `json:"baseline_id" gorm:"column:baseline_id"`     // 任务 ID
	TemplateID   int64  `json:"template_id" gorm:"column:template_id"`     // 模版 ID
	TemplateName string `json:"template_name" gorm:"column:template_name"` // 模板名称
	ItemID       int64  `json:"item_id" gorm:"column:item_id"`             // 检查项 ID
	ItemName     string `json:"item_name" gorm:"column:item_name"`         // 检查项名称
	TotalHosts   int64  `json:"total_hosts" gorm:"column:total_hosts"`     // 检查主机数
	PassedChecks int64  `json:"passed_checks" gorm:"column:passed_checks"` // 通过检查数
	FailedChecks int64  `json:"failed_checks" gorm:"column:failed_checks"` // 失败检查数
	ErrorChecks  int64  `json:"error_checks" gorm:"column:error_checks"`   // 错误检查数
}

// TableName 指定视图名为 baseline_check_item_view
func (BaselineCheckItemView) TableName() string {
	return "baseline_check_item_view"
}

// BaselineCheckHostCardStatistics 基线检查主机卡片统计视图实体
type BaselineCheckHostCardStatistics struct {
	BaselineID      string  `json:"baseline_id" gorm:"column:baseline_id;primaryKey"`  // 任务 ID
	PassCount       int64   `json:"pass_count" gorm:"column:pass_count"`               // 通过数量
	FailCount       int64   `json:"fail_count" gorm:"column:fail_count"`               // 失败数量
	ErrorCount      int64   `json:"error_count" gorm:"column:error_count"`             // 错误数量
	TotalCount      int64   `json:"total_count" gorm:"column:total_count"`             // 总数量
	PassRatePercent float64 `json:"pass_rate_percent" gorm:"column:pass_rate_percent"` // 通过率百分比
	TemplateName    string  `json:"template_name" gorm:"column:template_name"`         // 模板名称
}

// TableName 指定视图名为 baseline_check_host_card_statistics
func (BaselineCheckHostCardStatistics) TableName() string {
	return "baseline_check_host_card_statistics"
}

// CheckStatus 检查状态枚举常量
const (
	CheckStatusFailed = 0 // 未通过
	CheckStatusPassed = 1 // 通过
	CheckStatusError  = 2 // 检查异常
)
