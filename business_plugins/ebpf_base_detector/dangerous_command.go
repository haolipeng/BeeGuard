package main

import (
	"fmt"
	"regexp"
	"strings"
)

// DangerousCommandDetector 高危命令检测器
type DangerousCommandDetector struct {
	rules    []Rule
	compiled map[int64][]*regexp.Regexp // 预编译的正则表达式，key为规则ID
}

// NewDangerousCommandDetector 创建检测器实例
func NewDangerousCommandDetector(config *RuleConfig) (*DangerousCommandDetector, error) {
	d := &DangerousCommandDetector{
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

// Detect 检测命令是否匹配任何规则
// comm: 进程名（如 rm, curl, wget）
// args: 命令行参数
// 返回: 检测结果（如果匹配）或nil（如果不匹配）
func (d *DangerousCommandDetector) Detect(comm, args string) *DetectionResult {
	// 构建完整命令行用于匹配
	fullCmd := comm + " " + args

	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		matched, matchedPattern := d.matchRule(&rule, comm, fullCmd)
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

// matchRule 检测命令是否匹配指定规则
func (d *DangerousCommandDetector) matchRule(rule *Rule, comm, fullCmd string) (matched bool, pattern string) {
	switch rule.Match.Type {
	case MatchTypeRegex:
		// 使用预编译的正则表达式
		patterns := d.compiled[rule.ID]
		for i, re := range patterns {
			if re.MatchString(fullCmd) {
				return true, rule.Match.Patterns[i]
			}
		}

	case MatchTypeContains:
		// 包含匹配：检查fullCmd是否包含任一模式
		for _, p := range rule.Match.Patterns {
			if strings.Contains(fullCmd, p) {
				return true, p
			}
		}

	case MatchTypePrefix:
		// 前缀匹配：检查comm是否以任一模式开头
		for _, p := range rule.Match.Patterns {
			if strings.HasPrefix(comm, p) {
				return true, p
			}
		}

	case MatchTypeExact:
		// 精确匹配：检查comm是否与任一模式完全相同
		for _, p := range rule.Match.Patterns {
			if comm == p {
				return true, p
			}
		}
	}

	return false, ""
}

// DetectAll 检测命令是否匹配所有规则，返回所有匹配结果
// 与Detect不同，这个方法会返回所有匹配的规则，而不是第一个
func (d *DangerousCommandDetector) DetectAll(comm, args string) []*DetectionResult {
	var results []*DetectionResult
	fullCmd := comm + " " + args

	for _, rule := range d.rules {
		if !rule.Enabled {
			continue
		}

		matched, matchedPattern := d.matchRule(&rule, comm, fullCmd)
		if matched {
			results = append(results, &DetectionResult{
				RuleID:         rule.ID,
				RuleName:       rule.Name,
				Severity:       rule.Severity,
				Description:    rule.Description,
				MatchedPattern: matchedPattern,
			})
		}
	}

	return results
}

// GetEnabledRuleCount 返回启用的规则数量
func (d *DangerousCommandDetector) GetEnabledRuleCount() int {
	count := 0
	for _, rule := range d.rules {
		if rule.Enabled {
			count++
		}
	}
	return count
}

// GetRuleIDs 返回所有启用的规则ID
func (d *DangerousCommandDetector) GetRuleIDs() []int64 {
	var ids []int64
	for _, rule := range d.rules {
		if rule.Enabled {
			ids = append(ids, rule.ID)
		}
	}
	return ids
}
