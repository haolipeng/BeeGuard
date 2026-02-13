package main

// DetectionResult 检测结果
type DetectionResult struct {
	RuleID         string // 规则ID
	RuleName       string // 规则名称
	Severity       string // 严重级别: critical/high/medium/low
	Description    string // 规则描述
	MatchedPattern string // 匹配的模式
}

// Rule 检测规则
type Rule struct {
	ID          string `yaml:"id"`          // 规则唯一标识
	Name        string `yaml:"name"`        // 规则名称
	Description string `yaml:"description"` // 规则描述
	Severity    string `yaml:"severity"`    // 严重级别: critical/high/medium/low
	Enabled     bool   `yaml:"enabled"`     // 是否启用
	Match       Match  `yaml:"match"`       // 匹配配置
}

// Match 匹配配置
type Match struct {
	Type     string   `yaml:"type"`     // 匹配类型: regex/contains/prefix/exact
	Patterns []string `yaml:"patterns"` // 匹配模式列表
}

// RuleConfig YAML配置文件结构
type RuleConfig struct {
	Version     string `yaml:"version"`     // 配置版本
	Description string `yaml:"description"` // 配置描述
	Rules       []Rule `yaml:"rules"`       // 规则列表
}

// 严重级别常量
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)

// 匹配类型常量
const (
	MatchTypeRegex    = "regex"    // 正则表达式匹配
	MatchTypeContains = "contains" // 包含匹配
	MatchTypePrefix   = "prefix"   // 前缀匹配
	MatchTypeExact    = "exact"    // 精确匹配
)

// 检测类型常量
const (
	DetectionTypeDangerousCommand = "dangerous_command" // 高危命令
)
