package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	businessplugins "business_plugins/lib"
	"ebpf_base_detector/events"
)

// domainSuffixEntry 域名后缀匹配项
type domainSuffixEntry struct {
	suffix string   // 如 ".evil.com"
	rule   *IOCRule
}

// IOCMatcher IOC 匹配引擎
type IOCMatcher struct {
	mu          sync.RWMutex
	ipIndex     map[string]*IOCRule   // key: IP 字符串
	domainIndex map[string]*IOCRule   // key: 精确域名
	suffixRules []domainSuffixEntry   // 通配符域名后缀匹配
	portIndex   map[uint16]*IOCRule   // key: 端口号
	ipPortIndex map[string]*IOCRule   // key: "ip:port"
	ruleCount   int
}

// NewIOCMatcher 从配置构建 IOC 匹配器索引
func NewIOCMatcher(config *IOCRuleConfig) *IOCMatcher {
	m := &IOCMatcher{
		ipIndex:     make(map[string]*IOCRule),
		domainIndex: make(map[string]*IOCRule),
		portIndex:   make(map[uint16]*IOCRule),
		ipPortIndex: make(map[string]*IOCRule),
	}
	m.buildIndex(config)
	return m
}

// buildIndex 构建所有类型的索引
func (m *IOCMatcher) buildIndex(config *IOCRuleConfig) {
	enabledCount := 0
	for i := range config.Rules {
		rule := &config.Rules[i]
		if !rule.Enabled {
			continue
		}
		enabledCount++

		switch rule.IndicatorType {
		case IOCTypeIP:
			for _, ip := range rule.Indicators {
				m.ipIndex[ip] = rule
			}
		case IOCTypeDomain:
			for _, domain := range rule.Indicators {
				d := strings.ToLower(domain)
				if strings.HasPrefix(d, "*.") {
					// 通配符域名，存储后缀（如 "*.evil.com" -> ".evil.com"）
					m.suffixRules = append(m.suffixRules, domainSuffixEntry{
						suffix: d[1:], // 去掉 "*"，保留 ".evil.com"
						rule:   rule,
					})
				} else {
					m.domainIndex[d] = rule
				}
			}
		case IOCTypePort:
			for _, portStr := range rule.Indicators {
				port, _ := strconv.Atoi(portStr)
				m.portIndex[uint16(port)] = rule
			}
		case IOCTypeIPPort:
			for _, ipPort := range rule.Indicators {
				m.ipPortIndex[ipPort] = rule
			}
		}
	}
	m.ruleCount = enabledCount
}

// GetEnabledRuleCount 返回已启用的规则数量
func (m *IOCMatcher) GetEnabledRuleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ruleCount
}

// MatchConnect 对 CONNECT 事件进行 IOC 匹配
// 优先级：ip_port > ip > port（最具体优先）
func (m *IOCMatcher) MatchConnect(evt *events.ConnectEvent) *IOCMatchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	remoteIP := events.NetworkIPToString(evt.RemoteIP)
	remotePort := events.NetworkPortToHost(evt.RemotePort)
	ipPortKey := fmt.Sprintf("%s:%d", remoteIP, remotePort)

	// 1. ip_port 精确匹配（最高优先级）
	if rule, ok := m.ipPortIndex[ipPortKey]; ok {
		return &IOCMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: IOCTypeIPPort,
			MatchedValue:  ipPortKey,
			Description:   rule.Description,
		}
	}

	// 2. IP 匹配
	if rule, ok := m.ipIndex[remoteIP]; ok {
		return &IOCMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: IOCTypeIP,
			MatchedValue:  remoteIP,
			Description:   rule.Description,
		}
	}

	// 3. 端口匹配（最低优先级）
	if rule, ok := m.portIndex[remotePort]; ok {
		return &IOCMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: IOCTypePort,
			MatchedValue:  fmt.Sprintf("%d", remotePort),
			Description:   rule.Description,
		}
	}

	return nil
}

// MatchDNS 对 DNS 事件进行 IOC 匹配
// 先精确匹配，再后缀匹配
func (m *IOCMatcher) MatchDNS(evt *events.DNSEvent) *IOCMatchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domain := strings.ToLower(cstring(evt.Domain[:]))
	// 去除末尾点号（DNS 全限定域名格式）
	domain = strings.TrimSuffix(domain, ".")

	if domain == "" {
		return nil
	}

	// 1. 精确域名匹配
	if rule, ok := m.domainIndex[domain]; ok {
		return &IOCMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: IOCTypeDomain,
			MatchedValue:  domain,
			Description:   rule.Description,
		}
	}

	// 2. 后缀匹配（通配符域名 *.evil.com）
	for _, entry := range m.suffixRules {
		if strings.HasSuffix(domain, entry.suffix) {
			return &IOCMatchResult{
				RuleID:        entry.rule.ID,
				RuleName:      entry.rule.Name,
				Severity:      entry.rule.Severity,
				ThreatType:    entry.rule.ThreatType,
				IndicatorType: IOCTypeDomain,
				MatchedValue:  domain,
				Description:   entry.rule.Description,
			}
		}
	}

	return nil
}

// BuildIOCConnectRecord 从 CONNECT 事件和匹配结果构建 DataType 6008 告警
func BuildIOCConnectRecord(evt *events.ConnectEvent, result *IOCMatchResult) *businessplugins.Record {
	comm := cstring(evt.Comm[:])
	exePath := cstring(evt.ExePath[:])
	remoteIP := events.NetworkIPToString(evt.RemoteIP)
	remotePort := events.NetworkPortToHost(evt.RemotePort)

	protoStr := "unknown"
	switch evt.Protocol {
	case 6:
		protoStr = "tcp"
	case 17:
		protoStr = "udp"
	}

	return &businessplugins.Record{
		DataType:  6008,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":            fmt.Sprintf("%d", evt.PID),
				"tgid":           fmt.Sprintf("%d", evt.TGID),
				"ppid":           fmt.Sprintf("%d", evt.PPID),
				"uid":            fmt.Sprintf("%d", evt.UID),
				"comm":           comm,
				"exe_path":       exePath,
				"detection_type": DetectionTypeIOC,
				"event_type":     "connect",
				"rule_id":        result.RuleID,
				"rule_name":      result.RuleName,
				"severity":       result.Severity,
				"threat_type":    result.ThreatType,
				"indicator_type": result.IndicatorType,
				"matched_value":  result.MatchedValue,
				"description":    result.Description,
				"remote_ip":      remoteIP,
				"remote_port":    fmt.Sprintf("%d", remotePort),
				"protocol":       protoStr,
			},
		},
	}
}

// BuildIOCDNSRecord 从 DNS 事件和匹配结果构建 DataType 6008 告警
func BuildIOCDNSRecord(evt *events.DNSEvent, result *IOCMatchResult) *businessplugins.Record {
	comm := cstring(evt.Comm[:])
	exePath := cstring(evt.ExePath[:])
	domain := cstring(evt.Domain[:])
	dnsServerIP := events.NetworkIPToString(evt.DNSServerIP)

	return &businessplugins.Record{
		DataType:  6008,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":            fmt.Sprintf("%d", evt.PID),
				"tgid":           fmt.Sprintf("%d", evt.TGID),
				"ppid":           fmt.Sprintf("%d", evt.PPID),
				"uid":            fmt.Sprintf("%d", evt.UID),
				"comm":           comm,
				"exe_path":       exePath,
				"detection_type": DetectionTypeIOC,
				"event_type":     "dns",
				"rule_id":        result.RuleID,
				"rule_name":      result.RuleName,
				"severity":       result.Severity,
				"threat_type":    result.ThreatType,
				"indicator_type": result.IndicatorType,
				"matched_value":  result.MatchedValue,
				"description":    result.Description,
				"domain":         domain,
				"dns_server_ip":  dnsServerIP,
			},
		},
	}
}

// UpdateRules 原子替换规则索引
func (m *IOCMatcher) UpdateRules(config *IOCRuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重置索引
	m.ipIndex = make(map[string]*IOCRule)
	m.domainIndex = make(map[string]*IOCRule)
	m.suffixRules = nil
	m.portIndex = make(map[uint16]*IOCRule)
	m.ipPortIndex = make(map[string]*IOCRule)

	m.buildIndex(config)
}
