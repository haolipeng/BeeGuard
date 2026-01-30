package engine

import (
	"time"
)

// Event 检测事件
type Event struct {
	Timestamp time.Time // 事件时间
	SourceIP  string    // 源IP地址
	Username  string    // 用户名(可选)
	Action    string    // 动作类型: "failed", "invalid_user"
	Raw       string    // 原始日志行
	RuleName  string    // 匹配的规则名称
}

// Alert 告警信息
type Alert struct {
	AlertType   string `mapstructure:"alert_type"`   // 告警类型: "brute_force"
	Service     string `mapstructure:"service"`      // 服务名: "ssh"
	RuleName    string `mapstructure:"rule_name"`    // 规则名称
	Description string `mapstructure:"description"`  // 告警描述
	SourceIP    string `mapstructure:"source_ip"`    // 攻击源IP
	TargetUser  string `mapstructure:"target_user"`  // 目标用户名
	Count       int    `mapstructure:"count"`        // 失败次数
	Timeframe   int    `mapstructure:"timeframe"`    // 时间窗口(秒)
	FirstSeen   int64  `mapstructure:"first_seen"`   // 首次事件时间
	LastSeen    int64  `mapstructure:"last_seen"`    // 最后事件时间
	Level       int    `mapstructure:"level"`        // 告警级别
}

// Detector 检测器接口(可扩展框架核心)
type Detector interface {
	// Name 返回检测器名称，如 "ssh", "mysql", "ftp"
	Name() string

	// DataType 返回告警数据类型ID
	DataType() int

	// LogPaths 返回需要监控的日志文件路径
	LogPaths() []string

	// Parse 解析日志行，返回事件或nil(不匹配)
	Parse(line string) *Event

	// Check 检查事件是否触发告警
	Check(event *Event) *Alert
}

// ConfigUpdater 支持动态配置更新的检测器接口
type ConfigUpdater interface {
	// UpdateConfig 更新检测器配置
	// data 为 JSON 格式的配置数据
	UpdateConfig(data string) error
}
