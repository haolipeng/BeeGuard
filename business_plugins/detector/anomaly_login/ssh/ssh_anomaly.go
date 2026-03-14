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
	mu            sync.RWMutex
	config        config.SSHAnomalyConfig
	ipRuleIndex   map[string][]*config.AnomalyRule // IP -> 包含该IP的规则列表
	userRuleIndex map[string][]*config.AnomalyRule // 用户名 -> 包含该用户的规则列表
	alertCache    map[string]time.Time             // IP -> 上次告警时间（告警抑制）
}

// New 创建SSH异常登录检测器
func New(cfg config.SSHAnomalyConfig) (*Detector, error) {
	d := &Detector{
		config:        cfg,
		ipRuleIndex:   make(map[string][]*config.AnomalyRule),
		userRuleIndex: make(map[string][]*config.AnomalyRule),
		alertCache:    make(map[string]time.Time),
	}

	// 编译规则索引
	d.compileRuleIndexes()

	return d, nil
}

// compileRuleIndexes 构建IP和用户名到规则的索引映射
func (d *Detector) compileRuleIndexes() {
	d.ipRuleIndex = make(map[string][]*config.AnomalyRule)
	d.userRuleIndex = make(map[string][]*config.AnomalyRule)

	for i := range d.config.AnomalyRules {
		rule := &d.config.AnomalyRules[i]
		if !rule.Enabled {
			continue
		}
		for _, ip := range rule.IPs {
			d.ipRuleIndex[ip] = append(d.ipRuleIndex[ip], rule)
		}
		for _, user := range rule.Users {
			d.userRuleIndex[user] = append(d.userRuleIndex[user], rule)
		}
	}

	zap.S().Infof("SSH anomaly detector: compiled %d IPs, %d users from %d rules",
		len(d.ipRuleIndex), len(d.userRuleIndex), len(d.config.AnomalyRules))
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

// isUserAllowed 检查用户是否在规则的允许用户列表中
func isUserAllowed(username string, rule *config.AnomalyRule) bool {
	// 没有配置用户列表，默认允许所有用户
	if len(rule.Users) == 0 {
		return true
	}

	for _, u := range rule.Users {
		if u == username {
			return true
		}
	}
	return false
}

// Check 检查是否触发告警
func (d *Detector) Check(event *engine.Event) *engine.Alert {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 关键：没有配置规则时，不上报任何告警
	if !d.hasEnabledRules() {
		return nil
	}

	// 检查IP是否在白名单中
	rules, found := d.ipRuleIndex[event.SourceIP]
	if !found {
		// IP不在白名单 -> unknown_ip 告警
		if d.shouldSuppressAlert(event.SourceIP) {
			zap.S().Debugf("SSH anomaly alert suppressed for IP %s", event.SourceIP)
			return nil
		}
		d.alertCache[event.SourceIP] = time.Now()
		return &engine.Alert{
			AlertType:   "anomaly_login",
			Service:     "ssh",
			RuleName:    "ssh_anomaly_login",
			Description: fmt.Sprintf("检测到SSH异常登录: 用户 %s 从 %s 登录，该IP不在允许的白名单中",
				event.Username, event.SourceIP),
			SourceIP:     event.SourceIP,
			TargetUser:   event.Username,
			Count:        1,
			Timeframe:    0,
			FirstSeen:    event.Timestamp.Unix(),
			LastSeen:     event.Timestamp.Unix(),
			Level:        d.config.AlertLevel,
			AbnormalType: "unknown_ip",
		}
	}

	// IP在白名单中，检查是否有规则同时满足时间和用户条件
	for _, rule := range rules {
		if isTimeAllowed(event.Timestamp, rule) && isUserAllowed(event.Username, rule) {
			// 匹配到规则：IP允许、时间允许、用户允许 -> 正常登录
			return nil
		}
	}

	// 没有任何规则同时满足，需要确定具体的异常类型
	// 检查是否存在时间允许的规则（区分 abnormal_time 和 abnormal_user）
	abnormalType := "abnormal_time"
	description := fmt.Sprintf("检测到SSH异常登录: 用户 %s 从 %s 在 %s 登录，该时间不在允许的时间段内",
		event.Username, event.SourceIP, event.Timestamp.Format("15:04"))

	for _, rule := range rules {
		if isTimeAllowed(event.Timestamp, rule) {
			// 时间允许但用户不允许 -> abnormal_user
			abnormalType = "abnormal_user"
			description = fmt.Sprintf("检测到SSH异常登录: 用户 %s 从 %s 登录，该用户不在允许的用户列表中",
				event.Username, event.SourceIP)
			break
		}
	}

	if d.shouldSuppressAlert(event.SourceIP) {
		zap.S().Debugf("SSH anomaly alert suppressed for IP %s", event.SourceIP)
		return nil
	}
	d.alertCache[event.SourceIP] = time.Now()
	return &engine.Alert{
		AlertType:    "anomaly_login",
		Service:      "ssh",
		RuleName:     "ssh_anomaly_login",
		Description:  description,
		SourceIP:     event.SourceIP,
		TargetUser:   event.Username,
		Count:        1,
		Timeframe:    0,
		FirstSeen:    event.Timestamp.Unix(),
		LastSeen:     event.Timestamp.Unix(),
		Level:        d.config.AlertLevel,
		AbnormalType: abnormalType,
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

	// 重新编译规则索引
	d.compileRuleIndexes()

	// 清空告警缓存
	d.alertCache = make(map[string]time.Time)

	zap.S().Infof("SSH anomaly detector config updated: %d rules, %d IPs indexed",
		len(d.config.AnomalyRules), len(d.ipRuleIndex))

	return nil
}
