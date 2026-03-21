package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadMaliciousRequestRules 从 YAML 文件加载恶意请求检测规则
func LoadMaliciousRequestRules(path string) (*MaliciousRequestRuleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read malicious request rules file: %w", err)
	}

	var config MaliciousRequestRuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse malicious request rules file: %w", err)
	}

	if err := validateMaliciousRequestConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid malicious request rules config: %w", err)
	}

	return &config, nil
}

// validateMaliciousRequestConfig 校验恶意请求规则配置的有效性
func validateMaliciousRequestConfig(config *MaliciousRequestRuleConfig) error {
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
		case MaliciousRequestTypeIP, MaliciousRequestTypeDomain, MaliciousRequestTypePort, MaliciousRequestTypeIPPort:
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
	case MaliciousRequestTypeIP:
		if net.ParseIP(indicator) == nil {
			return fmt.Errorf("rule '%s': indicator %d: invalid IP address '%s'", ruleID, index, indicator)
		}
	case MaliciousRequestTypePort:
		port, err := strconv.Atoi(indicator)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("rule '%s': indicator %d: invalid port '%s'", ruleID, index, indicator)
		}
	case MaliciousRequestTypeIPPort:
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
	case MaliciousRequestTypeDomain:
		// 域名允许通配符前缀 *.
		d := strings.TrimPrefix(indicator, "*.")
		if d == "" {
			return fmt.Errorf("rule '%s': indicator %d: empty domain", ruleID, index)
		}
	}
	return nil
}

// LoadMaliciousRequestRulesOrDefault 加载恶意请求规则，如果失败则返回空配置
func LoadMaliciousRequestRulesOrDefault(path string) *MaliciousRequestRuleConfig {
	config, err := LoadMaliciousRequestRules(path)
	if err != nil {
		return &MaliciousRequestRuleConfig{
			Version:     "1.0",
			Description: "empty config (load failed)",
			Rules:       []MaliciousRequestRule{},
		}
	}
	return config
}

// ParseMaliciousRequestRulesFromJSON 解析服务端下发的 JSON 格式恶意请求规则
func ParseMaliciousRequestRulesFromJSON(data string) (*MaliciousRequestRuleConfig, error) {
	var config MaliciousRequestRuleConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("failed to parse malicious request rules JSON: %w", err)
	}

	if err := validateMaliciousRequestConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid malicious request rules config: %w", err)
	}

	return &config, nil
}
