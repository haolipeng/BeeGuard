package ssh

import (
	"fmt"
	"sync"
	"time"

	"gitlab.myinterest.top/security/agent/business_plugins/detector/config"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/engine"
	"go.uber.org/zap"
)

const (
	// DataTypeSSHAnomalyLogin SSH异常登录告警数据类型
	// 注意: 6003 已被服务端高危命令告警使用，改用 6005
	DataTypeSSHAnomalyLogin = 6005
)

// Detector SSH异常登录检测器
type Detector struct {
	mu          sync.RWMutex
	config      config.SSHAnomalyConfig
	ipWhitelist map[string]bool   // 所有规则的IP白名单合集
	alertCache  map[string]time.Time // IP -> 上次告警时间（告警抑制）
}

// New 创建SSH异常登录检测器
func New(cfg config.SSHAnomalyConfig) (*Detector, error) {
	d := &Detector{
		config:      cfg,
		ipWhitelist: make(map[string]bool),
		alertCache:  make(map[string]time.Time),
	}

	// 编译IP白名单
	d.compileIPWhitelist()

	return d, nil
}

// compileIPWhitelist 从所有启用的规则中收集IP白名单
func (d *Detector) compileIPWhitelist() {
	d.ipWhitelist = make(map[string]bool)

	for _, rule := range d.config.AnomalyRules {
		if !rule.Enabled {
			continue
		}
		for _, ip := range rule.IPs {
			d.ipWhitelist[ip] = true
		}
	}

	zap.S().Infof("SSH anomaly detector: compiled %d IPs in whitelist from %d rules",
		len(d.ipWhitelist), len(d.config.AnomalyRules))
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
	return DataTypeSSHAnomalyLogin
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

	// 检查IP是否在白名单中
	if d.ipWhitelist[event.SourceIP] {
		// 正常登录，不告警
		return nil
	}

	// 检查告警抑制
	if d.shouldSuppressAlert(event.SourceIP) {
		zap.S().Debugf("SSH anomaly alert suppressed for IP %s", event.SourceIP)
		return nil
	}

	// 更新告警缓存
	d.alertCache[event.SourceIP] = time.Now()

	// 异常登录，生成告警
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

	// 重新编译IP白名单
	d.compileIPWhitelist()

	// 清空告警缓存
	d.alertCache = make(map[string]time.Time)

	zap.S().Infof("SSH anomaly detector config updated: %d rules, %d IPs in whitelist",
		len(d.config.AnomalyRules), len(d.ipWhitelist))

	return nil
}
