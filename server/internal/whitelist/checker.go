package whitelist

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"time"

	wlModel "github.com/haolipeng/BeeGuard/server/internal/models/whitelist"

	"gorm.io/gorm"
)

// Checker 白名单同步检查器
type Checker struct {
	db    *gorm.DB
	cache sync.Map // map[string]*cacheEntry  alertType -> rules
}

type cacheEntry struct {
	rules     []wlModel.WhitelistRule
	expiresAt time.Time
}

const cacheTTL = 30 * time.Second

// NewChecker 创建白名单检查器
func NewChecker(db *gorm.DB) *Checker {
	return &Checker{db: db}
}

// Check 检查单条告警是否命中白名单
// alertType: "dangerous_command", "reverse_shell" 等
// fields: 告警字段 map (field_name → value)
// agentID: Agent 标识
// 返回：是否命中、命中的规则 ID
func (c *Checker) Check(alertType string, fields map[string]string, agentID string) (bool, int64) {
	rules := c.loadRules(alertType)
	if len(rules) == 0 {
		return false, 0
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// 检查 scope
		if rule.Scope == wlModel.ScopeAgent {
			if !c.matchAgentScope(rule.AgentIDs, agentID) {
				continue
			}
		}

		// 评估条件
		if c.evaluateConditions(&rule.Conditions, fields) {
			// 异步更新命中计数
			go c.incrementHitCount(alertType, rule.ID)
			return true, rule.ID
		}
	}

	return false, 0
}

// InvalidateCache 清除指定告警类型的缓存
func (c *Checker) InvalidateCache(alertType string) {
	c.cache.Delete(alertType)
}

// InvalidateAllCache 清除所有缓存
func (c *Checker) InvalidateAllCache() {
	c.cache.Range(func(key, _ interface{}) bool {
		c.cache.Delete(key)
		return true
	})
}

// loadRules 从缓存或数据库加载规则列表
func (c *Checker) loadRules(alertType string) []wlModel.WhitelistRule {
	// 检查缓存
	if entry, ok := c.cache.Load(alertType); ok {
		ce := entry.(*cacheEntry)
		if time.Now().Before(ce.expiresAt) {
			return ce.rules
		}
	}

	// 从数据库加载
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return nil
	}

	var rules []wlModel.WhitelistRule
	result := c.db.Table(tableName).Where("enabled = ?", true).Find(&rules)
	if result.Error != nil {
		return nil
	}

	// 写入缓存
	c.cache.Store(alertType, &cacheEntry{
		rules:     rules,
		expiresAt: time.Now().Add(cacheTTL),
	})

	return rules
}

// matchAgentScope 检查 agentID 是否在白名单规则的 agent_ids 列表中
func (c *Checker) matchAgentScope(agentIDsJSON string, agentID string) bool {
	if agentIDsJSON == "" {
		return false
	}

	var agentIDs []string
	if err := json.Unmarshal([]byte(agentIDsJSON), &agentIDs); err != nil {
		return false
	}

	for _, id := range agentIDs {
		if id == agentID {
			return true
		}
	}
	return false
}

// evaluateConditions 评估匹配条件
func (c *Checker) evaluateConditions(conditions *wlModel.Conditions, fields map[string]string) bool {
	if len(conditions.Rules) == 0 {
		return false
	}

	if conditions.Logic == wlModel.LogicAnd {
		for _, rule := range conditions.Rules {
			if !c.evaluateRule(&rule, fields) {
				return false
			}
		}
		return true
	}

	// OR 逻辑
	for _, rule := range conditions.Rules {
		if c.evaluateRule(&rule, fields) {
			return true
		}
	}
	return false
}

// evaluateRule 评估单条匹配规则
func (c *Checker) evaluateRule(rule *wlModel.ConditionRule, fields map[string]string) bool {
	fieldValue, ok := fields[rule.Field]
	if !ok {
		return false
	}

	switch rule.Operator {
	case wlModel.OperatorEq:
		return fieldValue == rule.Value
	case wlModel.OperatorRegex:
		matched, err := regexp.MatchString(rule.Value, fieldValue)
		if err != nil {
			return false
		}
		return matched
	case wlModel.OperatorContains:
		return strings.Contains(fieldValue, rule.Value)
	default:
		return false
	}
}

// incrementHitCount 异步更新命中计数
func (c *Checker) incrementHitCount(alertType string, ruleID int64) {
	tableName, err := wlModel.GetWhitelistTableName(alertType)
	if err != nil {
		return
	}

	c.db.Table(tableName).Where("id = ?", ruleID).
		UpdateColumn("hit_count", gorm.Expr("hit_count + 1"))
}
