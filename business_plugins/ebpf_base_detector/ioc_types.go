package main

// 威胁类型常量
const (
	ThreatTypeMining      = "mining"
	ThreatTypeC2          = "c2"
	ThreatTypePhishing    = "phishing"
	ThreatTypeDataLeakage = "data_leakage"
)

// 指标类型常量
const (
	IOCTypeIP     = "ip"      // IP 地址匹配
	IOCTypeDomain = "domain"  // 域名匹配（精确 + 后缀通配）
	IOCTypePort   = "port"    // 端口匹配
	IOCTypeIPPort = "ip_port" // IP:Port 复合匹配
)

// 检测类型常量
const (
	DetectionTypeIOC = "ioc"
)

// IOCRule 单条 IOC 规则
type IOCRule struct {
	ID            string   `yaml:"id" json:"id"`
	Name          string   `yaml:"name" json:"name"`
	Description   string   `yaml:"description" json:"description"`
	Severity      string   `yaml:"severity" json:"severity"`
	Enabled       bool     `yaml:"enabled" json:"enabled"`
	ThreatType    string   `yaml:"threat_type" json:"threat_type"`
	IndicatorType string   `yaml:"indicator_type" json:"indicator_type"`
	Indicators    []string `yaml:"indicators" json:"indicators"`
}

// IOCRuleConfig YAML 配置根结构
type IOCRuleConfig struct {
	Version     string    `yaml:"version" json:"version"`
	Description string    `yaml:"description" json:"description"`
	Rules       []IOCRule `yaml:"rules" json:"rules"`
}

// IOCMatchResult 匹配结果
type IOCMatchResult struct {
	RuleID        string
	RuleName      string
	Severity      string
	ThreatType    string
	IndicatorType string
	MatchedValue  string
	Description   string
}
