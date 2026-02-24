package ssh

import (
	"fmt"
	"sync"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/business_plugins/detector/config"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/engine"
	"go.uber.org/zap"
)

// Detector SSH暴力破解检测器
type Detector struct {
	mu      sync.RWMutex
	config  config.SSHConfig
	windows map[string]*engine.SlidingWindow // 每个规则一个滑动窗口
}

// New 创建SSH检测器
func New(cfg config.SSHConfig) *Detector {
	d := &Detector{
		config:  cfg,
		windows: make(map[string]*engine.SlidingWindow),
	}

	// 为每个规则创建滑动窗口
	for _, rule := range cfg.Rules {
		d.windows[rule.Name] = engine.NewSlidingWindow(
			time.Duration(rule.Timeframe)*time.Second,
			time.Duration(rule.Ignore)*time.Second,
			rule.Frequency,
		)
		zap.S().Infof("created sliding window for rule %s: timeframe=%ds, frequency=%d, ignore=%ds",
			rule.Name, rule.Timeframe, rule.Frequency, rule.Ignore)
	}

	return d
}

// Name 返回检测器名称
func (d *Detector) Name() string {
	return "ssh"
}

// DataType 返回数据类型
func (d *Detector) DataType() int {
	return businessplugins.AlertTypeSSHBruteForce
}

// LogPaths 返回监控的日志路径
func (d *Detector) LogPaths() []string {
	return d.config.LogPaths
}

// Parse 解析日志行
func (d *Detector) Parse(line string) *engine.Event {
	parsed := ParseLine(line)
	if parsed == nil {
		return nil
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// 检查白名单
	if d.config.IsWhitelisted(parsed.SourceIP) {
		return nil
	}

	// 查找匹配的规则
	var matchedRule *config.Rule
	for i := range d.config.Rules {
		rule := &d.config.Rules[i]
		if rule.Action == parsed.Action {
			matchedRule = rule
			break
		}
	}

	if matchedRule == nil {
		return nil
	}

	return &engine.Event{
		Timestamp: parsed.Timestamp,
		SourceIP:  parsed.SourceIP,
		Username:  parsed.Username,
		Action:    parsed.Action,
		Raw:       line,
		RuleName:  matchedRule.Name,
	}
}

// Check 检查是否触发告警
func (d *Detector) Check(event *engine.Event) *engine.Alert {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 获取对应规则的滑动窗口
	window, exists := d.windows[event.RuleName]
	if !exists {
		return nil
	}

	// 查找规则配置
	var rule *config.Rule
	for i := range d.config.Rules {
		if d.config.Rules[i].Name == event.RuleName {
			rule = &d.config.Rules[i]
			break
		}
	}
	if rule == nil {
		return nil
	}

	// 检查滑动窗口
	result := window.Check(event.SourceIP, event.Timestamp)
	if !result.Triggered {
		return nil
	}

	// 构造告警
	return &engine.Alert{
		AlertType:   "brute_force",
		Service:     "ssh",
		RuleName:    rule.Name,
		Description: fmt.Sprintf("%s: 检测到来自 %s 的暴力破解攻击，%d秒内失败%d次",
			rule.Description, event.SourceIP, rule.Timeframe, result.Count),
		SourceIP:   event.SourceIP,
		TargetUser: event.Username,
		Count:      result.Count,
		Timeframe:  rule.Timeframe,
		FirstSeen:  result.FirstSeen.Unix(),
		LastSeen:   result.LastSeen.Unix(),
		Level:      rule.Level,
	}
}

// UpdateConfig 更新检测器配置 (实现 engine.ConfigUpdater 接口)
func (d *Detector) UpdateConfig(data string) error {
	// 解析新配置
	newCfg, err := config.ParseSSHConfigFromJSON(data)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// 更新规则和白名单
	d.config.Rules = newCfg.Rules
	d.config.Whitelist = newCfg.Whitelist

	// 重建滑动窗口
	newWindows := make(map[string]*engine.SlidingWindow)
	for _, rule := range newCfg.Rules {
		newWindows[rule.Name] = engine.NewSlidingWindow(
			time.Duration(rule.Timeframe)*time.Second,
			time.Duration(rule.Ignore)*time.Second,
			rule.Frequency,
		)
		zap.S().Infof("updated sliding window for rule %s: timeframe=%ds, frequency=%d, ignore=%ds",
			rule.Name, rule.Timeframe, rule.Frequency, rule.Ignore)
	}
	d.windows = newWindows

	zap.S().Infof("SSH detector config updated: %d rules, %d whitelist entries",
		len(newCfg.Rules), len(newCfg.Whitelist))

	return nil
}
