package back

import (
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// CodeqlRule 规则详情配置实体
type CodeqlRule struct {
	RuleID                      int64           `json:"rule_id" gorm:"primaryKey;not null;autoIncrement"` // 主键，自增
	Enabled                     *bool           `json:"enabled"`                                           // 是否启用，默认 TRUE
	ID                          *string         `json:"id"`                                                // 规则id：go/incomplete-hostname-regexp
	Code                        *string         `json:"code"`                                              // 编程语言
	ShortDescriptionText        *string         `json:"short_description_text"`                            // 简短描述
	FullDescriptionText         *string         `json:"full_description_text"`                             // 完整描述
	DefaultConfigurationEnabled *string         `json:"default_configuration_enabled"`                     // 默认配置启用true=1,
	DefaultConfigurationLevel   *string         `json:"default_configuration_level"`                       // 默认配置级别
	PropertiesTags              *string         `json:"properties_tags"`                                   // 标签
	PropertiesDescription       *string         `json:"properties_description"`                            // 描述
	PropertiesKind              *string         `json:"properties_kind"`                                   // 问题类型
	PropertiesPrecision         *string         `json:"properties_precision"`                              // 高中低
	PropertiesProblemSeverity   *string         `json:"properties_problem_severity"`                       // 属性问题严重程度,warning
	PropertiesSecuritySeverity  *string         `json:"properties_security_severity"`                      // 属性安全级别,打分
	CreateTime                  common.DateTime `json:"create_time"`                                       // 创建时间，默认当前时间
	UpdateTime                  common.DateTime `json:"update_time"`                                       // 更新时间，默认当前时间并自动更新
	Deleted                     *int8           `json:"deleted"`                                           // 逻辑删除标志，0=未删除，1=已删除
}

// TableName 指定表名为 codeql_rule
func (CodeqlRule) TableName() string {
	return "codeql_rule"
}
