package detector

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadRules 从YAML文件加载检测规则
func LoadRules(path string) (*RuleConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rules file: %w", err)
	}

	// 验证规则配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid rules config: %w", err)
	}

	return &config, nil
}

// validateConfig 验证规则配置的有效性
func validateConfig(config *RuleConfig) error {
	if config.Version == "" {
		return fmt.Errorf("missing version field")
	}

	seenIDs := make(map[string]bool)
	for i, rule := range config.Rules {
		// 检查规则ID唯一性
		if rule.ID == "" {
			return fmt.Errorf("rule %d: missing id", i)
		}
		if seenIDs[rule.ID] {
			return fmt.Errorf("rule %d: duplicate id '%s'", i, rule.ID)
		}
		seenIDs[rule.ID] = true

		// 检查规则名称
		if rule.Name == "" {
			return fmt.Errorf("rule '%s': missing name", rule.ID)
		}

		// 检查严重级别
		switch rule.Severity {
		case SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow:
			// 有效
		default:
			return fmt.Errorf("rule '%s': invalid severity '%s'", rule.ID, rule.Severity)
		}

		// 检查匹配类型
		switch rule.Match.Type {
		case MatchTypeRegex, MatchTypeContains, MatchTypePrefix, MatchTypeExact:
			// 有效
		default:
			return fmt.Errorf("rule '%s': invalid match type '%s'", rule.ID, rule.Match.Type)
		}

		// 检查模式列表
		if len(rule.Match.Patterns) == 0 {
			return fmt.Errorf("rule '%s': no patterns defined", rule.ID)
		}
	}

	return nil
}

// LoadRulesOrDefault 加载规则，如果失败则返回空配置
func LoadRulesOrDefault(path string) *RuleConfig {
	config, err := LoadRules(path)
	if err != nil {
		return &RuleConfig{
			Version:     "1.0",
			Description: "empty config (load failed)",
			Rules:       []Rule{},
		}
	}
	return config
}
