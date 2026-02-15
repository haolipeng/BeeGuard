package main

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	businessplugins "business_plugins/lib"
	"nids/log"
)

// Detector 检测引擎
type Detector struct {
	rules   []*SuricataRule
	tracker *AttackTracker
	client  *businessplugins.Client
	logger  *log.Logger
}

// NewDetector 创建检测引擎
func NewDetector(rules []*SuricataRule, tracker *AttackTracker,
	client *businessplugins.Client, logger *log.Logger) *Detector {
	return &Detector{
		rules:   rules,
		tracker: tracker,
		client:  client,
		logger:  logger,
	}
}

// Match 对一个 HTTP 请求运行所有规则
func (d *Detector) Match(req *HTTPRequest) {
	for _, rule := range d.rules {
		matched, snippet := d.matchRule(rule, req)
		if matched {
			// 记录攻击状态
			state := d.tracker.RecordAttack(req.SrcIP, rule.SID)

			// 构建并发送告警 Record
			d.sendAlert(rule, req, state, snippet)
		}
	}
}

// matchRule 单条规则匹配（AND 逻辑：所有 Matcher 都必须命中）
func (d *Detector) matchRule(rule *SuricataRule, req *HTTPRequest) (bool, string) {
	var lastSnippet string

	for _, m := range rule.Matchers {
		// 根据 StickyBuf 选择匹配缓冲区
		buf := d.selectBuffer(m.StickyBuf, req)

		matched := false
		switch m.Type {
		case MatcherContent:
			matched, lastSnippet = matchContent(buf, m.Content, m.NoCase)
		case MatcherPCRE:
			matched, lastSnippet = matchPCRE(buf, m.PCRE)
		}

		if !matched {
			return false, ""
		}
	}

	return true, lastSnippet
}

// selectBuffer 根据 StickyBuffer 选择对应的请求字段
func (d *Detector) selectBuffer(sticky StickyBuffer, req *HTTPRequest) string {
	switch sticky {
	case StickyHTTPURI:
		return req.URI
	case StickyHTTPHeader:
		return req.Headers
	case StickyHTTPBody:
		return string(req.Body)
	case StickyHTTPMethod:
		return req.Method
	default:
		return req.RawPayload
	}
}

// matchContent 执行 content 匹配
func matchContent(buf, content string, noCase bool) (bool, string) {
	searchBuf := buf
	searchContent := content

	if noCase {
		searchBuf = strings.ToLower(buf)
		searchContent = strings.ToLower(content)
	}

	idx := strings.Index(searchBuf, searchContent)
	if idx == -1 {
		return false, ""
	}

	snippet := extractSnippet(buf, idx, len(content))
	return true, snippet
}

// matchPCRE 执行 PCRE 正则匹配
func matchPCRE(buf string, re *regexp.Regexp) (bool, string) {
	loc := re.FindStringIndex(buf)
	if loc == nil {
		return false, ""
	}

	snippet := extractSnippet(buf, loc[0], loc[1]-loc[0])
	return true, snippet
}

// extractSnippet 提取匹配位置前后各 64 字节的片段
func extractSnippet(buf string, matchStart, matchLen int) string {
	start := matchStart - 64
	if start < 0 {
		start = 0
	}
	end := matchStart + matchLen + 64
	if end > len(buf) {
		end = len(buf)
	}
	// 整体限制 256 字节
	if end-start > 256 {
		end = start + 256
	}
	return buf[start:end]
}

// sendAlert 构建并发送 DataType=6010 告警
func (d *Detector) sendAlert(rule *SuricataRule, req *HTTPRequest,
	state *AttackState, matchedSnippet string) {

	record := &businessplugins.Record{
		DataType:  6010,
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"src_ip":             req.SrcIP,
				"dst_ip":             req.DstIP,
				"dst_port":           fmt.Sprintf("%d", req.DstPort),
				"src_port":           fmt.Sprintf("%d", req.SrcPort),
				"vulnerability_name": rule.Msg,
				"attack_status":      rule.Classtype,
				"severity":           severityToString(rule.Severity),
				"sid":                fmt.Sprintf("%d", rule.SID),
				"reference":          rule.Reference,
				"attack_count":       fmt.Sprintf("%d", state.Count),
				"last_attack_time":   state.LastSeenTime.Format(time.RFC3339),
				"first_attack_time":  state.FirstSeenTime.Format(time.RFC3339),
				"matched_payload":    matchedSnippet,
				"http_method":        req.Method,
				"http_uri":           req.URI,
			},
		},
	}

	d.logger.Warn("Attack detected",
		"sid", rule.SID,
		"msg", rule.Msg,
		"severity", severityToString(rule.Severity),
		"src_ip", req.SrcIP,
		"dst_port", req.DstPort,
		"uri", req.URI,
		"count", state.Count)

	if err := d.client.SendRecord(record); err != nil {
		d.logger.Error("Failed to send alert record", "error", err, "sid", rule.SID)
	}
}

// severityToString 将数字严重级别转为字符串
func severityToString(severity int) string {
	switch severity {
	case 1:
		return "critical"
	case 2:
		return "high"
	case 3:
		return "medium"
	case 4:
		return "low"
	default:
		return "unknown"
	}
}
