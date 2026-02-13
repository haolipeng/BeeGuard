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
	MaliciousRequestTypeIP     = "ip"      // IP 地址匹配
	MaliciousRequestTypeDomain = "domain"  // 域名匹配（精确 + 后缀通配）
	MaliciousRequestTypePort   = "port"    // 端口匹配
	MaliciousRequestTypeIPPort = "ip_port" // IP:Port 复��匹配
)

// 检测类型常量
const (
	DetectionTypeMaliciousRequest = "malicious_request"
)

// MaliciousRequestRule 单条恶意请求检测规则
type MaliciousRequestRule struct {
	ID            string   `yaml:"id" json:"id"`
	Name          string   `yaml:"name" json:"name"`
	Description   string   `yaml:"description" json:"description"`
	Severity      string   `yaml:"severity" json:"severity"`
	Enabled       bool     `yaml:"enabled" json:"enabled"`
	ThreatType    string   `yaml:"threat_type" json:"threat_type"`
	IndicatorType string   `yaml:"indicator_type" json:"indicator_type"`
	Indicators    []string `yaml:"indicators" json:"indicators"`
}

// MaliciousRequestRuleConfig YAML 配置根结构
type MaliciousRequestRuleConfig struct {
	Version     string                 `yaml:"version" json:"version"`
	Description string                 `yaml:"description" json:"description"`
	Rules       []MaliciousRequestRule `yaml:"rules" json:"rules"`
}

// MaliciousRequestMatchResult 匹配结果
type MaliciousRequestMatchResult struct {
	RuleID        string
	RuleName      string
	Severity      string
	ThreatType    string
	IndicatorType string
	MatchedValue  string
	Description   string
}
