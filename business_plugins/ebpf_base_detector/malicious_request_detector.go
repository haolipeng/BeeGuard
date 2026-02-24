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
	rule   *MaliciousRequestRule
}

// MaliciousRequestDetector 恶意请求匹配引擎
type MaliciousRequestDetector struct {
	mu          sync.RWMutex
	ipIndex     map[string]*MaliciousRequestRule   // key: IP 字符串
	domainIndex map[string]*MaliciousRequestRule   // key: 精确域名
	suffixRules []domainSuffixEntry                // 通配符域名后缀匹配
	portIndex   map[uint16]*MaliciousRequestRule   // key: 端口号
	ipPortIndex map[string]*MaliciousRequestRule   // key: "ip:port"
	ruleCount   int
}

// NewMaliciousRequestDetector 从配置构建恶意请求匹配器索引
func NewMaliciousRequestDetector(config *MaliciousRequestRuleConfig) *MaliciousRequestDetector {
	m := &MaliciousRequestDetector{
		ipIndex:     make(map[string]*MaliciousRequestRule),
		domainIndex: make(map[string]*MaliciousRequestRule),
		portIndex:   make(map[uint16]*MaliciousRequestRule),
		ipPortIndex: make(map[string]*MaliciousRequestRule),
	}
	m.buildIndex(config)
	return m
}

// buildIndex 构建所有类型的索引
func (m *MaliciousRequestDetector) buildIndex(config *MaliciousRequestRuleConfig) {
	enabledCount := 0
	for i := range config.Rules {
		rule := &config.Rules[i]
		if !rule.Enabled {
			continue
		}
		enabledCount++

		switch rule.IndicatorType {
		case MaliciousRequestTypeIP:
			for _, ip := range rule.Indicators {
				m.ipIndex[ip] = rule
			}
		case MaliciousRequestTypeDomain:
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
		case MaliciousRequestTypePort:
			for _, portStr := range rule.Indicators {
				port, _ := strconv.Atoi(portStr)
				m.portIndex[uint16(port)] = rule
			}
		case MaliciousRequestTypeIPPort:
			for _, ipPort := range rule.Indicators {
				m.ipPortIndex[ipPort] = rule
			}
		}
	}
	m.ruleCount = enabledCount
}

// GetEnabledRuleCount 返回已启用的规则数量
func (m *MaliciousRequestDetector) GetEnabledRuleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ruleCount
}

// MatchConnect 对 CONNECT 事件进行恶意请求匹配
// 优先级：ip_port > ip > port（最具体优先）
func (m *MaliciousRequestDetector) MatchConnect(evt *events.ConnectEvent) *MaliciousRequestMatchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	remoteIP := events.NetworkIPToString(evt.RemoteIP)
	remotePort := events.NetworkPortToHost(evt.RemotePort)
	ipPortKey := fmt.Sprintf("%s:%d", remoteIP, remotePort)

	// 1. ip_port 精确匹配（最高优先级）
	if rule, ok := m.ipPortIndex[ipPortKey]; ok {
		return &MaliciousRequestMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: MaliciousRequestTypeIPPort,
			MatchedValue:  ipPortKey,
			Description:   rule.Description,
		}
	}

	// 2. IP 匹配
	if rule, ok := m.ipIndex[remoteIP]; ok {
		return &MaliciousRequestMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: MaliciousRequestTypeIP,
			MatchedValue:  remoteIP,
			Description:   rule.Description,
		}
	}

	// 3. 端口匹配（最低优先级）
	if rule, ok := m.portIndex[remotePort]; ok {
		return &MaliciousRequestMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: MaliciousRequestTypePort,
			MatchedValue:  fmt.Sprintf("%d", remotePort),
			Description:   rule.Description,
		}
	}

	return nil
}

// MatchDNS 对 DNS 事件进行恶意请求匹配
// 先精确匹配，再后缀匹配
func (m *MaliciousRequestDetector) MatchDNS(evt *events.DNSEvent) *MaliciousRequestMatchResult {
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
		return &MaliciousRequestMatchResult{
			RuleID:        rule.ID,
			RuleName:      rule.Name,
			Severity:      rule.Severity,
			ThreatType:    rule.ThreatType,
			IndicatorType: MaliciousRequestTypeDomain,
			MatchedValue:  domain,
			Description:   rule.Description,
		}
	}

	// 2. 后缀匹配（通配符域名 *.evil.com）
	for _, entry := range m.suffixRules {
		if strings.HasSuffix(domain, entry.suffix) {
			return &MaliciousRequestMatchResult{
				RuleID:        entry.rule.ID,
				RuleName:      entry.rule.Name,
				Severity:      entry.rule.Severity,
				ThreatType:    entry.rule.ThreatType,
				IndicatorType: MaliciousRequestTypeDomain,
				MatchedValue:  domain,
				Description:   entry.rule.Description,
			}
		}
	}

	return nil
}

// BuildMaliciousRequestConnectRecord 从 CONNECT 事件和匹配结果构建 DataType 6008 告警
func BuildMaliciousRequestConnectRecord(evt *events.ConnectEvent, result *MaliciousRequestMatchResult) *businessplugins.Record {
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
		DataType:  businessplugins.AlertTypeMaliciousRequest,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":            fmt.Sprintf("%d", evt.PID),
				"tgid":           fmt.Sprintf("%d", evt.TGID),
				"ppid":           fmt.Sprintf("%d", evt.PPID),
				"uid":            fmt.Sprintf("%d", evt.UID),
				"comm":           comm,
				"exe_path":       exePath,
				"detection_type": DetectionTypeMaliciousRequest,
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

// BuildMaliciousRequestDNSRecord 从 DNS 事件和匹配结果构建 DataType 6008 告警
func BuildMaliciousRequestDNSRecord(evt *events.DNSEvent, result *MaliciousRequestMatchResult) *businessplugins.Record {
	comm := cstring(evt.Comm[:])
	exePath := cstring(evt.ExePath[:])
	domain := cstring(evt.Domain[:])
	dnsServerIP := events.NetworkIPToString(evt.DNSServerIP)

	return &businessplugins.Record{
		DataType:  businessplugins.AlertTypeMaliciousRequest,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"pid":            fmt.Sprintf("%d", evt.PID),
				"tgid":           fmt.Sprintf("%d", evt.TGID),
				"ppid":           fmt.Sprintf("%d", evt.PPID),
				"uid":            fmt.Sprintf("%d", evt.UID),
				"comm":           comm,
				"exe_path":       exePath,
				"detection_type": DetectionTypeMaliciousRequest,
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
func (m *MaliciousRequestDetector) UpdateRules(config *MaliciousRequestRuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 重置索引
	m.ipIndex = make(map[string]*MaliciousRequestRule)
	m.domainIndex = make(map[string]*MaliciousRequestRule)
	m.suffixRules = nil
	m.portIndex = make(map[uint16]*MaliciousRequestRule)
	m.ipPortIndex = make(map[string]*MaliciousRequestRule)

	m.buildIndex(config)
}
