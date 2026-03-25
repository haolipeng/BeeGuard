package back

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// BaselineTemplate 基线模板实体
type BaselineTemplate struct {
	ID           int64           `json:"id" gorm:"primaryKey;not null"`                               // 模板ID
	TemplateName string          `json:"template_name" gorm:"not null"`                               // 模板名称
	TemplateType string          `json:"template_type" gorm:"not null"`                               // 基线类型: os_security/db_security/middleware_security
	OSType       *string         `json:"os_type,omitempty"`                                           // 操作系统类型: linux/windows
	Version      *string         `json:"version,omitempty"`                                           // 版本
	ItemCount    *int32          `json:"item_count,omitempty"`                                        // 检查项数量
	Description  *string         `json:"description,omitempty"`                                       // 模板描述
	IsEnabled    int16           `json:"is_enabled" gorm:"not null;default:1"`                        // 是否启用: 0-禁用 1-启用
	BaselineIDs  interface{}     `json:"baseline_ids,omitempty" gorm:"type:text;column:baseline_ids"` // 基线ID列表（接收数组或字符串，存储为文本）
	CreatedAt    common.DateTime `json:"created_at"`                                                  // 创建时间
	UpdatedAt    common.DateTime `json:"updated_at"`                                                  // 更新时间
}

// TableName 指定表名为 baseline_template
func (BaselineTemplate) TableName() string {
	return "baseline_template"
}

// BaselineTemplateHostLink 基线模板与主机关联实体
type BaselineTemplateHostLink struct {
	ID            int64           `json:"id" gorm:"primaryKey;not null"`     // 关联记录主键
	TemplateID    int64           `json:"template_id" gorm:"not null;index"` // 基线模板ID
	TemplateName  string          `json:"template_name" gorm:"not null"`
	TargetRange   *string         `json:"target_range" gorm:"not null;type:text"` // 目标范围，存储主机ID列表的JSON格式
	ScanFrequency string          `json:"scan_frequency" gorm:"not null"`         // 扫描频率
	CreatedAt     common.DateTime `json:"created_at"`                             // 创建时间
	UpdatedAt     common.DateTime `json:"updated_at"`                             // 更新时间
}

// TableName 指定表名为 baseline_template_host_link
func (BaselineTemplateHostLink) TableName() string {
	return "baseline_template_host_link"
}

// BaselineCheckItem 基线检查项实体
type BaselineCheckItem struct {
	ID            int64           `json:"id" gorm:"primaryKey;not null"`     // 检查项ID
	TemplateID    int64           `json:"template_id" gorm:"not null;index"` // 关联基线模板ID
	ItemName      string          `json:"item_name" gorm:"not null"`         // 检查项名称
	Category      string          `json:"category" gorm:"not null"`          // 检查分类
	RiskLevel     string          `json:"risk_level" gorm:"not null"`        // 风险等级: high/medium/low
	CheckRules    string          `json:"check_rules" gorm:"not null"`       // 检查规则
	FixSuggestion string          `json:"fix_suggestion" gorm:"not null"`    // 修复建议
	FixScript     string          `json:"fix_script" gorm:"not null"`        // 修复脚本
	CreatedAt     common.DateTime `json:"created_at"`                        // 创建时间
	UpdatedAt     common.DateTime `json:"updated_at"`                        // 更新时间
}

// TableName 指定表名为 baseline_check_item
func (BaselineCheckItem) TableName() string {
	return "baseline_check_item"
}

// OSType 操作系统类型枚举常量
const (
	OSTypeLinux   = "linux"
	OSTypeWindows = "windows"
)

// RiskLevel 风险等级枚举常量
const (
	RiskLevelHigh   = "high"
	RiskLevelMedium = "medium"
	RiskLevelLow    = "low"
)

// EnabledStatus 启用状态枚举常量
const (
	StatusDisabled = 0 // 禁用
	StatusEnabled  = 1 // 启用
)

// BaselineType 基线类型枚举常量
const (
	BaselineTypeOSSecurity         = "os_security"         // 操作系统安全基线
	BaselineTypeDBSecurity         = "db_security"         // 数据库安全基线
	BaselineTypeMiddlewareSecurity = "middleware_security" // 中间件安全基线
)

// GetBaselineIDsAsString 获取基线ID列表的字符串表示
func (bt *BaselineTemplate) GetBaselineIDsAsString() *string {
	if bt.BaselineIDs == nil {
		return nil
	}

	switch v := bt.BaselineIDs.(type) {
	case string:
		return &v
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		// 转换为逗号分隔的字符串
		var ids []string
		for _, item := range v {
			if id, ok := item.(float64); ok {
				ids = append(ids, fmt.Sprintf("%.0f", id))
			} else if idStr, ok := item.(string); ok {
				ids = append(ids, idStr)
			}
		}
		result := strings.Join(ids, ",")
		return &result
	case []string:
		if len(v) == 0 {
			return nil
		}
		result := strings.Join(v, ",")
		return &result
	case []int:
		if len(v) == 0 {
			return nil
		}
		var ids []string
		for _, id := range v {
			ids = append(ids, strconv.Itoa(id))
		}
		result := strings.Join(ids, ",")
		return &result
	case []int64:
		if len(v) == 0 {
			return nil
		}
		var ids []string
		for _, id := range v {
			ids = append(ids, strconv.FormatInt(id, 10))
		}
		result := strings.Join(ids, ",")
		return &result
	default:
		return nil
	}
}

// SetBaselineIDsFromInterface 设置基线ID列表
func (bt *BaselineTemplate) SetBaselineIDsFromInterface(data interface{}) {
	bt.BaselineIDs = data
}
