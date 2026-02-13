package detector

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadIOCRules 从 YAML 文件加载 IOC 检测规则
func LoadIOCRules(path string) (*IOCRuleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read IOC rules file: %w", err)
	}

	var config IOCRuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse IOC rules file: %w", err)
	}

	if err := validateIOCConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid IOC rules config: %w", err)
	}

	return &config, nil
}

// validateIOCConfig 校验 IOC 规则配置的有效性
func validateIOCConfig(config *IOCRuleConfig) error {
	if config.Version == "" {
		return fmt.Errorf("missing version field")
	}

	seenIDs := make(map[string]bool)
	for i, rule := range config.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule %d: missing id", i)
		}
		if seenIDs[rule.ID] {
			return fmt.Errorf("rule %d: duplicate id '%s'", i, rule.ID)
		}
		seenIDs[rule.ID] = true

		if rule.Name == "" {
			return fmt.Errorf("rule '%s': missing name", rule.ID)
		}

		// 校验严重级别
		switch rule.Severity {
		case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow:
			// 有效
		default:
			return fmt.Errorf("rule '%s': invalid severity '%s'", rule.ID, rule.Severity)
		}

		// 校验指标类型
		switch rule.IndicatorType {
		case IOCTypeIP, IOCTypeDomain, IOCTypePort, IOCTypeIPPort:
			// 有效
		default:
			return fmt.Errorf("rule '%s': invalid indicator_type '%s'", rule.ID, rule.IndicatorType)
		}

		// 校验威胁类型
		switch rule.ThreatType {
		case ThreatTypeMining, ThreatTypeC2, ThreatTypePhishing, ThreatTypeDataLeakage:
			// 有效
		default:
			return fmt.Errorf("rule '%s': invalid threat_type '%s'", rule.ID, rule.ThreatType)
		}

		// 校验指标列表非空
		if len(rule.Indicators) == 0 {
			return fmt.Errorf("rule '%s': no indicators defined", rule.ID)
		}

		// 校验指标值格式
		for j, indicator := range rule.Indicators {
			if err := validateIndicator(rule.ID, rule.IndicatorType, indicator, j); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateIndicator 校验单个指标值的格式
func validateIndicator(ruleID, indicatorType, indicator string, index int) error {
	switch indicatorType {
	case IOCTypeIP:
		if net.ParseIP(indicator) == nil {
			return fmt.Errorf("rule '%s': indicator %d: invalid IP address '%s'", ruleID, index, indicator)
		}
	case IOCTypePort:
		port, err := strconv.Atoi(indicator)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("rule '%s': indicator %d: invalid port '%s'", ruleID, index, indicator)
		}
	case IOCTypeIPPort:
		host, portStr, err := net.SplitHostPort(indicator)
		if err != nil {
			return fmt.Errorf("rule '%s': indicator %d: invalid ip:port format '%s'", ruleID, index, indicator)
		}
		if net.ParseIP(host) == nil {
			return fmt.Errorf("rule '%s': indicator %d: invalid IP in '%s'", ruleID, index, indicator)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("rule '%s': indicator %d: invalid port in '%s'", ruleID, index, indicator)
		}
	case IOCTypeDomain:
		// 域名允许通配符前缀 *.
		d := strings.TrimPrefix(indicator, "*.")
		if d == "" {
			return fmt.Errorf("rule '%s': indicator %d: empty domain", ruleID, index)
		}
	}
	return nil
}

// LoadIOCRulesOrDefault 加载 IOC 规则，如果失败则返回空配置
func LoadIOCRulesOrDefault(path string) *IOCRuleConfig {
	config, err := LoadIOCRules(path)
	if err != nil {
		return &IOCRuleConfig{
			Version:     "1.0",
			Description: "empty config (load failed)",
			Rules:       []IOCRule{},
		}
	}
	return config
}

// ParseIOCRulesFromJSON 解析服务端下发的 JSON 格式 IOC 规则
func ParseIOCRulesFromJSON(data string) (*IOCRuleConfig, error) {
	var config IOCRuleConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("failed to parse IOC rules JSON: %w", err)
	}

	if err := validateIOCConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid IOC rules config: %w", err)
	}

	return &config, nil
}
