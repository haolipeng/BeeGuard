package main

import (
	"fmt"
	"regexp"
	"strings"
)

// SensitiveFileDetector 敏感文件检测器
type SensitiveFileDetector struct {
	rules    []Rule
	compiled map[int64][]*regexp.Regexp // 预编译的正则表达式，key为规则ID
}

// NewSensitiveFileDetector 创建敏感文件检测器实例
func NewSensitiveFileDetector(config *RuleConfig) (*SensitiveFileDetector, error) {
	d := &SensitiveFileDetector{
		rules:    config.Rules,
		compiled: make(map[int64][]*regexp.Regexp),
	}

	// 预编译所有启用的正则表达式规则
	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		if rule.Match.Type == MatchTypeRegex {
			var patterns []*regexp.Regexp
			for _, p := range rule.Match.Patterns {
				re, err := regexp.Compile(p)
				if err != nil {
					return nil, fmt.Errorf("rule '%d': invalid regex pattern '%s': %w", rule.ID, p, err)
				}
				patterns = append(patterns, re)
			}
			d.compiled[rule.ID] = patterns
		}
	}

	return d, nil
}

// Detect 检测文件路径是否匹配敏感规则
// filePath: 文件创建/重命名的目标路径
// 返回: 检测结果（如果匹配）或nil（如果不匹配）
func (d *SensitiveFileDetector) Detect(filePath string) *DetectionResult {
	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		matched, matchedPattern := d.matchRule(&rule, filePath)
		if matched {
			return &DetectionResult{
				RuleID:         rule.ID,
				RuleName:       rule.Name,
				Severity:       rule.Severity,
				Description:    rule.Description,
				MatchedPattern: matchedPattern,
			}
		}
	}

	return nil
}

// matchRule 检测文件路径是否匹配指定规则
func (d *SensitiveFileDetector) matchRule(rule *Rule, filePath string) (matched bool, pattern string) {
	switch rule.Match.Type {
	case MatchTypeRegex:
		patterns := d.compiled[rule.ID]
		for i, re := range patterns {
			if re.MatchString(filePath) {
				return true, rule.Match.Patterns[i]
			}
		}

	case MatchTypeContains:
		for _, p := range rule.Match.Patterns {
			if strings.Contains(filePath, p) {
				return true, p
			}
		}

	case MatchTypePrefix:
		for _, p := range rule.Match.Patterns {
			if strings.HasPrefix(filePath, p) {
				return true, p
			}
		}

	case MatchTypeExact:
		for _, p := range rule.Match.Patterns {
			if filePath == p {
				return true, p
			}
		}
	}

	return false, ""
}

// GetEnabledRuleCount 返回启用的规则数量
func (d *SensitiveFileDetector) GetEnabledRuleCount() int {
	count := 0
	for _, rule := range d.rules {
		if rule.Enabled {
			count++
		}
	}
	return count
}
