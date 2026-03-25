package common

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringList 自定义字符串数组类型，用于 JSONB 列存储
// 数据库中存储为 JSON 数组（如 ["192.168.1.10","192.168.1.11"]）
type StringList []string

// Value 实现 driver.Valuer 接口，序列化为 JSON 写入 JSONB 列
func (sl StringList) Value() (driver.Value, error) {
	if sl == nil {
		return "[]", nil
	}
	data, err := json.Marshal(sl)
	if err != nil {
		return nil, fmt.Errorf("StringList.Value: %w", err)
	}
	return string(data), nil
}

// Scan 实现 sql.Scanner 接口，从 JSONB 列读取
func (sl *StringList) Scan(value interface{}) error {
	if value == nil {
		*sl = StringList{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("StringList.Scan: unsupported type %T", value)
	}

	var result []string
	if err := json.Unmarshal(bytes, &result); err != nil {
		return fmt.Errorf("StringList.Scan: %w", err)
	}
	*sl = result
	return nil
}

// MarshalJSON 实现 json.Marshaler 接口，API 输出为 JSON 数组
func (sl StringList) MarshalJSON() ([]byte, error) {
	if sl == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(sl))
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (sl *StringList) UnmarshalJSON(data []byte) error {
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	*sl = result
	return nil
}

// First 返回第一个元素，空列表返回空字符串
func (sl StringList) First() string {
	if len(sl) == 0 {
		return ""
	}
	return sl[0]
}
