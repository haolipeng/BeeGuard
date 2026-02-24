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

// Detector SSH异常登录检测器
type Detector struct {
	mu          sync.RWMutex
	config      config.SSHAnomalyConfig
	ipRuleIndex map[string][]*config.AnomalyRule // IP -> 包含该IP的规则列表
	alertCache  map[string]time.Time             // IP -> 上次告警时间（告警抑制）
}

// New 创建SSH异常登录检测器
func New(cfg config.SSHAnomalyConfig) (*Detector, error) {
	d := &Detector{
		config:      cfg,
		ipRuleIndex: make(map[string][]*config.AnomalyRule),
		alertCache:  make(map[string]time.Time),
	}

	// 编译IP到规则的索引
	d.compileIPRuleIndex()

	return d, nil
}

// compileIPRuleIndex 构建IP到规则的索引映射
func (d *Detector) compileIPRuleIndex() {
	d.ipRuleIndex = make(map[string][]*config.AnomalyRule)

	for i := range d.config.AnomalyRules {
		rule := &d.config.AnomalyRules[i]
		if !rule.Enabled {
			continue
		}
		for _, ip := range rule.IPs {
			d.ipRuleIndex[ip] = append(d.ipRuleIndex[ip], rule)
		}
	}

	zap.S().Infof("SSH anomaly detector: compiled %d IPs from %d rules",
		len(d.ipRuleIndex), len(d.config.AnomalyRules))
}

// parseTimeString 解析 "HH:MM" 格式的时间字符串
func parseTimeString(s string) (hour, minute int, err error) {
	var h, m int
	n, err := fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil || n != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s, expected HH:MM", s)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("invalid time value: %s", s)
	}
	return h, m, nil
}

// isTimeInRange 检查时间是否在单个时间段内
func isTimeInRange(t time.Time, tr config.TimeRange) bool {
	startHour, startMin, err := parseTimeString(tr.Start)
	if err != nil {
		zap.S().Warnf("invalid time range start: %v", err)
		return true // 配置错误时默认允许
	}

	endHour, endMin, err := parseTimeString(tr.End)
	if err != nil {
		zap.S().Warnf("invalid time range end: %v", err)
		return true // 配置错误时默认允许
	}

	// 检查 start >= end 的无效配置
	startMins := startHour*60 + startMin
	endMins := endHour*60 + endMin
	if startMins >= endMins {
		zap.S().Warnf("invalid time range: start %s >= end %s", tr.Start, tr.End)
		return true // 配置错误时默认允许
	}

	// 获取事件时间的时分
	eventMins := t.Hour()*60 + t.Minute()

	return eventMins >= startMins && eventMins <= endMins
}

// isTimeAllowed 检查时间是否在规则的任意时间段内
func isTimeAllowed(t time.Time, rule *config.AnomalyRule) bool {
	// 没有配置时间段，默认全天允许
	if len(rule.TimeRanges) == 0 {
		return true
	}

	// 检查是否在任一时间段内
	for _, tr := range rule.TimeRanges {
		if isTimeInRange(t, tr) {
			return true
		}
	}
	return false
}

// hasEnabledRules 检查是否有启用的规则
func (d *Detector) hasEnabledRules() bool {
	for _, rule := range d.config.AnomalyRules {
		if rule.Enabled && len(rule.IPs) > 0 {
			return true
		}
	}
	return false
}

// Name 返回检测器名称
func (d *Detector) Name() string {
	return "ssh_anomaly_login"
}

// DataType 返回数据类型
func (d *Detector) DataType() int {
	return businessplugins.AlertTypeSSHAnomalyLogin
}

// LogPaths 返回监控的日志路径
func (d *Detector) LogPaths() []string {
	return d.config.LogPaths
}

// Parse 解析日志行
func (d *Detector) Parse(line string) *engine.Event {
	// 只解析成功登录事件
	parsed := ParseSuccessLogin(line)
	if parsed == nil {
		return nil
	}

	return &engine.Event{
		Timestamp: parsed.Timestamp,
		SourceIP:  parsed.SourceIP,
		Username:  parsed.Username,
		Action:    "success_login",
		Raw:       line,
		RuleName:  "",
	}
}

// Check 检查是否触发告警
func (d *Detector) Check(event *engine.Event) *engine.Alert {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 关键：没有配置规则时，不上报任何告警
	if !d.hasEnabledRules() {
		return nil
	}

	// 检查IP是否在白名单中，并验证时间段
	rules, found := d.ipRuleIndex[event.SourceIP]
	if found {
		for _, rule := range rules {
			if isTimeAllowed(event.Timestamp, rule) {
				// IP匹配且时间允许，正常登录
				return nil
			}
		}
		// IP在白名单但时间不允许 -> 生成时间异常告警
		if d.shouldSuppressAlert(event.SourceIP) {
			zap.S().Debugf("SSH anomaly alert suppressed for IP %s", event.SourceIP)
			return nil
		}
		d.alertCache[event.SourceIP] = time.Now()
		return &engine.Alert{
			AlertType:   "anomaly_login",
			Service:     "ssh",
			RuleName:    "ssh_anomaly_login",
			Description: fmt.Sprintf("检测到SSH异常登录: 用户 %s 从 %s 在 %s 登录，该时间不在允许的时间段内",
				event.Username, event.SourceIP, event.Timestamp.Format("15:04")),
			SourceIP:   event.SourceIP,
			TargetUser: event.Username,
			Count:      1,
			Timeframe:  0,
			FirstSeen:  event.Timestamp.Unix(),
			LastSeen:   event.Timestamp.Unix(),
			Level:      d.config.AlertLevel,
		}
	}

	// IP不在白名单，检查告警抑制
	if d.shouldSuppressAlert(event.SourceIP) {
		zap.S().Debugf("SSH anomaly alert suppressed for IP %s", event.SourceIP)
		return nil
	}

	// 更新告警缓存
	d.alertCache[event.SourceIP] = time.Now()

	// IP异常登录，生成告警
	return &engine.Alert{
		AlertType:   "anomaly_login",
		Service:     "ssh",
		RuleName:    "ssh_anomaly_login",
		Description: fmt.Sprintf("检测到SSH异常登录: 用户 %s 从 %s 登录，该IP不在允许的白名单中",
			event.Username, event.SourceIP),
		SourceIP:   event.SourceIP,
		TargetUser: event.Username,
		Count:      1,
		Timeframe:  0,
		FirstSeen:  event.Timestamp.Unix(),
		LastSeen:   event.Timestamp.Unix(),
		Level:      d.config.AlertLevel,
	}
}

// shouldSuppressAlert 检查是否应该抑制告警
func (d *Detector) shouldSuppressAlert(ip string) bool {
	lastAlert, exists := d.alertCache[ip]
	if !exists {
		return false
	}

	ignoreTime := time.Duration(d.config.IgnoreTime) * time.Second
	return time.Since(lastAlert) < ignoreTime
}

// UpdateConfig 更新检测器配置（实现 engine.ConfigUpdater 接口）
func (d *Detector) UpdateConfig(data string) error {
	newCfg, err := config.ParseSSHAnomalyConfigFromJSON(data)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.config = *newCfg

	// 重新编译IP到规则的索引
	d.compileIPRuleIndex()

	// 清空告警缓存
	d.alertCache = make(map[string]time.Time)

	zap.S().Infof("SSH anomaly detector config updated: %d rules, %d IPs indexed",
		len(d.config.AnomalyRules), len(d.ipRuleIndex))

	return nil
}
