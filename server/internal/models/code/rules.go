package code

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Rules 规则集配置实体
type Rules struct {
	RulesID         int64           `json:"rules_id" gorm:"primaryKey;not null;autoIncrement"` // 主键，自增
	RuleName        string          `json:"rule_name" gorm:"not null"`                         // 规则集名称
	RuleCount       *int            `json:"rule_count"`                                        // 有多少条规则数
	RuleID          *int            `json:"rule_id"`                                           // 关联codeq_rule表的rule_id
	RuleIDs         *string         `json:"rule_ids"`                                          // 关联codeq_rule表的rule_id
	ApplicableScene *string         `json:"applicable_scene"`                                  // 适用场景
	RiskCoverage    *string         `json:"risk_coverage"`                                     // 风险覆盖范围
	TotalRules      *int            `json:"total_rules"`                                       // 关联规则数
	Description     *string         `json:"description"`                                       // 规则集描述
	Status          string          `json:"status"`                                            // 启用状态，默认 ENABLED
	CreateTime      common.DateTime `json:"create_time"`                                       // 创建时间，默认当前时间
	UpdateTime      common.DateTime `json:"update_time"`                                       // 更新时间，默认当前时间并自动更新
	Deleted         *int8           `json:"deleted"`                                           // 逻辑删除标志，0=未删除，1=已删除
}

// TableName 指定表名为 codeql_rules
func (Rules) TableName() string {
	return "codeql_rules"
}
