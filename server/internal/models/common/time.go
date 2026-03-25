package common

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// DateTime 自定义时间类型，用于统一时间格式输出
type DateTime struct {
	time.Time
}

// MarshalJSON 实现 JSON 序列化接口，格式化时间为 "2006-01-02 15:04:05"
func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, dt.Time.Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON 实现 JSON 反序列化接口
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	str := string(data[1 : len(data)-1]) // 移除引号
	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return err
	}
	dt.Time = t
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (dt DateTime) Value() (driver.Value, error) {
	return dt.Time, nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		dt.Time = v
		return nil
	default:
		return fmt.Errorf("cannot scan %T into DateTime", value)
	}
}