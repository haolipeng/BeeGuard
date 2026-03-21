package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// LoadRulesFile 从文件加载并解析 Suricata 规则
func LoadRulesFile(path string) ([]*SuricataRule, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open rules file %s: %w", path, err)
	}
	defer f.Close()

	var rules []*SuricataRule
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading rules file: %w", err)
	}

	return rules, nil
}

// optionToken 解析后的选项 token
type optionToken struct {
	keyword string
	value   string
}

// parseLine 解析一行 Suricata 规则
func parseLine(line string) (*SuricataRule, error) {
	// 找到选项体的起始位置
	openParen := strings.Index(line, "(")
	if openParen == -1 {
		return nil, fmt.Errorf("missing options section (no opening parenthesis)")
	}

	closeParen := strings.LastIndex(line, ")")
	if closeParen == -1 || closeParen <= openParen {
		return nil, fmt.Errorf("missing closing parenthesis")
	}

	optionsBody := line[openParen+1 : closeParen]

	// 使用状态机解析选项体
	tokens, err := tokenizeOptions(optionsBody)
	if err != nil {
		return nil, fmt.Errorf("failed to tokenize options: %w", err)
	}

	rule := &SuricataRule{
		Severity: 3, // 默认 medium
	}

	// 当前正在构建的 matcher（content/pcre 后跟 nocase/sticky 修饰符）
	var currentMatcher *Matcher

	for _, tok := range tokens {
		keyword := strings.TrimSpace(tok.keyword)
		value := strings.TrimSpace(tok.value)

		switch keyword {
		case "msg":
			rule.Msg = unquote(value)

		case "sid":
			sid, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid sid: %s", value)
			}
			rule.SID = sid

		case "rev":
			rev, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid rev: %s", value)
			}
			rule.Rev = rev

		case "severity":
			sev, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid severity: %s", value)
			}
			rule.Severity = sev

		case "classtype":
			rule.Classtype = value

		case "reference":
			rule.Reference = value

		case "content":
			// 保存之前的 matcher
			if currentMatcher != nil {
				rule.Matchers = append(rule.Matchers, *currentMatcher)
			}
			currentMatcher = &Matcher{
				Type:    MatcherContent,
				Content: unescapeContent(unquote(value)),
			}

		case "pcre":
			// 保存之前的 matcher
			if currentMatcher != nil {
				rule.Matchers = append(rule.Matchers, *currentMatcher)
			}
			compiled, raw, err := compilePCRE(unquote(value))
			if err != nil {
				return nil, fmt.Errorf("failed to compile pcre: %w", err)
			}
			currentMatcher = &Matcher{
				Type:    MatcherPCRE,
				PCRE:    compiled,
				PCRERaw: raw,
			}

		case "nocase":
			if currentMatcher != nil {
				currentMatcher.NoCase = true
			}

		case "http.uri":
			if currentMatcher != nil {
				currentMatcher.StickyBuf = StickyHTTPURI
			}

		case "http.header":
			if currentMatcher != nil {
				currentMatcher.StickyBuf = StickyHTTPHeader
			}

		case "http.request_body":
			if currentMatcher != nil {
				currentMatcher.StickyBuf = StickyHTTPBody
			}

		case "http.method":
			if currentMatcher != nil {
				currentMatcher.StickyBuf = StickyHTTPMethod
			}
		}
	}

	// 保存最后一个 matcher
	if currentMatcher != nil {
		rule.Matchers = append(rule.Matchers, *currentMatcher)
	}

	// 验证
	if rule.Msg == "" {
		return nil, fmt.Errorf("rule missing msg")
	}
	if rule.SID == 0 {
		return nil, fmt.Errorf("rule missing sid")
	}
	if len(rule.Matchers) == 0 {
		return nil, fmt.Errorf("rule has no matchers (sid:%d)", rule.SID)
	}

	return rule, nil
}

// tokenizeOptions 使用状态机将选项体切分为 token 列表
// 处理双引号和 PCRE 分隔符内的分号
func tokenizeOptions(body string) ([]optionToken, error) {
	var tokens []optionToken
	var current strings.Builder

	inQuote := false
	inPCRE := false
	escaped := false

	for i := 0; i < len(body); i++ {
		ch := body[i]

		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			current.WriteByte(ch)
			escaped = true
			continue
		}

		if ch == '"' && !inPCRE {
			inQuote = !inQuote
			current.WriteByte(ch)
			continue
		}

		if ch == '/' && !inQuote {
			inPCRE = !inPCRE
			current.WriteByte(ch)
			continue
		}

		// 只在引号和 PCRE 之外的分号处切分
		if ch == ';' && !inQuote && !inPCRE {
			tok := strings.TrimSpace(current.String())
			if tok != "" {
				kv := splitKeyValue(tok)
				tokens = append(tokens, kv)
			}
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	// 处理末尾没有分号的情况
	tok := strings.TrimSpace(current.String())
	if tok != "" {
		kv := splitKeyValue(tok)
		tokens = append(tokens, kv)
	}

	return tokens, nil
}

// splitKeyValue 按第一个冒号分割关键字和值
func splitKeyValue(s string) optionToken {
	idx := strings.Index(s, ":")
	if idx == -1 {
		return optionToken{keyword: s}
	}
	return optionToken{
		keyword: s[:idx],
		value:   s[idx+1:],
	}
}

// unquote 移除字符串两端的双引号
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// unescapeContent 处理 Suricata content 中的 |hex| 表示法
// 例如: "abc|41 42|def" → "abcABdef"
func unescapeContent(raw string) string {
	var result strings.Builder
	i := 0

	for i < len(raw) {
		if raw[i] == '|' {
			// 找到配对的 |
			end := strings.Index(raw[i+1:], "|")
			if end == -1 {
				result.WriteByte(raw[i])
				i++
				continue
			}

			hexStr := strings.ReplaceAll(raw[i+1:i+1+end], " ", "")
			decoded, err := hex.DecodeString(hexStr)
			if err != nil {
				// hex 解码失败，原样保留
				result.WriteString(raw[i : i+2+end])
			} else {
				result.Write(decoded)
			}
			i = i + 2 + end
		} else {
			result.WriteByte(raw[i])
			i++
		}
	}

	return result.String()
}

// compilePCRE 解析并编译 Suricata PCRE 表达式
// 输入格式: /pattern/flags
// 返回编译后的正则和原始模式串
func compilePCRE(raw string) (*regexp.Regexp, string, error) {
	if len(raw) < 2 || raw[0] != '/' {
		return nil, raw, fmt.Errorf("invalid pcre format: %s", raw)
	}

	// 从末尾找最后一个 /
	lastSlash := strings.LastIndex(raw, "/")
	if lastSlash <= 0 {
		return nil, raw, fmt.Errorf("invalid pcre format: %s", raw)
	}

	pattern := raw[1:lastSlash]
	flags := raw[lastSlash+1:]

	// 映射 Suricata PCRE flags 到 Go regexp flags
	var prefix strings.Builder
	for _, f := range flags {
		switch f {
		case 'i':
			prefix.WriteString("(?i)")
		case 's':
			prefix.WriteString("(?s)")
		case 'm':
			prefix.WriteString("(?m)")
		}
	}

	goPattern := prefix.String() + pattern
	compiled, err := regexp.Compile(goPattern)
	if err != nil {
		return nil, raw, fmt.Errorf("failed to compile pcre /%s/: %w", pattern, err)
	}

	return compiled, raw, nil
}
