package analysis

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Report 分析报告模型
type Report struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	AnalysisType    string    `json:"analysis_type" gorm:"column:analysis_type;type:varchar(50);not null"`
	ScopeKey        string    `json:"scope_key" gorm:"column:scope_key;type:varchar(255);not null"`
	AlertCount      int       `json:"alert_count" gorm:"column:alert_count;not null;default:0"`
	AlertSnapshot   JSONB     `json:"alert_snapshot" gorm:"column:alert_snapshot;type:jsonb"`
	RiskLevel       string    `json:"risk_level" gorm:"column:risk_level;type:varchar(20)"`
	AttackPattern   string    `json:"attack_pattern" gorm:"column:attack_pattern;type:text"`
	AttackStage     string    `json:"attack_stage" gorm:"column:attack_stage;type:varchar(100)"`
	Summary         string    `json:"summary" gorm:"column:summary;type:text"`
	Recommendations JSONB     `json:"recommendations" gorm:"column:recommendations;type:jsonb"`
	IOCIndicators   JSONB     `json:"ioc_indicators" gorm:"column:ioc_indicators;type:jsonb"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at;type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 设置表名
func (Report) TableName() string {
	return "analysis_report"
}

// JSONB 自定义类型，用于处理 JSONB 字段
type JSONB json.RawMessage

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	*j = JSONB(bytes)
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (j *JSONB) UnmarshalJSON(data []byte) error {
	*j = JSONB(data)
	return nil
}

// ToJSONB 将任意类型转换为 JSONB
func ToJSONB(v interface{}) JSONB {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return JSONB(data)
}