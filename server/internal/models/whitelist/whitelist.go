package whitelist

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// Conditions 白名单匹配条件
type Conditions struct {
	Logic string          `json:"logic"` // "and" 或 "or"
	Rules []ConditionRule `json:"rules"`
}

// ConditionRule 单条匹配规则
type ConditionRule struct {
	Field    string `json:"field"`    // 告警字段名: command, user, host_ip, source_ip 等
	Operator string `json:"operator"` // 运算符: eq, regex, contains
	Value    string `json:"value"`    // 匹配值
}

// Value 实现 driver.Valuer 接口，用于 GORM 写入 JSONB
func (c Conditions) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口，用于 GORM 读取 JSONB
func (c *Conditions) Scan(value interface{}) error {
	if value == nil {
		*c = Conditions{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Conditions: not []byte")
	}
	return json.Unmarshal(bytes, c)
}

// WhitelistRule 白名单规则基础结构体
type WhitelistRule struct {
	ID          int64           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string          `json:"name" gorm:"column:name;type:varchar(128);not null"`
	Description string          `json:"description,omitempty" gorm:"column:description;type:varchar(512)"`
	Scope       int16           `json:"scope" gorm:"column:scope;not null;default:0"`       // 0=全局, 1=指定Agent
	AgentIDs    string          `json:"agent_ids,omitempty" gorm:"column:agent_ids;type:text"`
	Conditions  Conditions      `json:"conditions" gorm:"column:conditions;type:jsonb;not null"`
	Enabled     bool            `json:"enabled" gorm:"column:enabled;not null;default:true"`
	HitCount    int64           `json:"hit_count" gorm:"column:hit_count;not null;default:0"`
	CreatedBy   string          `json:"created_by,omitempty" gorm:"column:created_by;type:varchar(64)"`
	CreatedAt   common.DateTime `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   common.DateTime `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// AlertTypeToTable 告警类型到白名单表名的映射
var AlertTypeToTable = map[string]string{
	"dangerous_command":    "whitelist_dangerous_command",
	"reverse_shell":        "whitelist_reverse_shell",
	"privilege_escalation": "whitelist_privilege_escalation",
	"abnormal_login":       "whitelist_abnormal_login",
	"brute_force":          "whitelist_brute_force",
	"malicious_request":    "whitelist_malicious_request",
	"network_attack":       "whitelist_network_attack",
	"malware_scan":         "whitelist_malware_scan",
	"fileguard":            "whitelist_fileguard",
	"container_alert":      "whitelist_container_alert",
}

// AlertTypeToAlertTable 告警类型到告警表名的映射
var AlertTypeToAlertTable = map[string]string{
	"dangerous_command":    "alert_dangerous_command",
	"reverse_shell":        "alert_reverse_shell",
	"privilege_escalation": "alert_privilege_escalation",
	"abnormal_login":       "alert_abnormal_login",
	"brute_force":          "alert_brute_force",
	"malicious_request":    "alert_malicious_request",
	"network_attack":       "alert_network_attack",
	"malware_scan":         "alert_malware_scan",
	"fileguard":            "alert_file_integrity",
	"container_alert":      "alert_container_dangerous_command", // 容器告警检查多张表
}

// ValidAlertTypes 所有有效的告警类型
var ValidAlertTypes = []string{
	"dangerous_command",
	"reverse_shell",
	"privilege_escalation",
	"abnormal_login",
	"brute_force",
	"malicious_request",
	"network_attack",
	"malware_scan",
	"fileguard",
	"container_alert",
}

// IsValidAlertType 检查告警类型是否有效
func IsValidAlertType(alertType string) bool {
	_, ok := AlertTypeToTable[alertType]
	return ok
}

// GetWhitelistTableName 获取白名单表名
func GetWhitelistTableName(alertType string) (string, error) {
	table, ok := AlertTypeToTable[alertType]
	if !ok {
		return "", fmt.Errorf("invalid alert type: %s", alertType)
	}
	return table, nil
}

// Scope 范围枚举
const (
	ScopeGlobal = 0 // 全局
	ScopeAgent  = 1 // 指定Agent
)

// Operator 运算符枚举
const (
	OperatorEq       = "eq"       // 精确匹配
	OperatorRegex    = "regex"    // 正则匹配
	OperatorContains = "contains" // 包含
)

// Logic 逻辑运算符枚举
const (
	LogicAnd = "and"
	LogicOr  = "or"
)
