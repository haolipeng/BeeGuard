package main

import "regexp"

// SuricataRule 一条解析后的 Suricata 规则
type SuricataRule struct {
	// 元信息
	Msg       string // msg:"描述"
	SID       int    // sid:12345
	Rev       int    // rev:1
	Severity  int    // severity:1 (1=critical,2=high,3=medium,4=low)
	Classtype string // classtype:attempted-admin
	Reference string // reference:cve,2021-44228

	// 匹配条件列表（AND 逻辑，全部满足才命中）
	Matchers []Matcher
}

// Matcher 单个匹配条件
type Matcher struct {
	Type      MatcherType    // content 或 pcre
	Content   string         // content 字符串（已解码）
	NoCase    bool           // 大小写不敏感
	PCRE      *regexp.Regexp // 编译后的正则
	PCRERaw   string         // 原始 pcre 字符串（日志用）
	StickyBuf StickyBuffer   // 匹配位置
}

// MatcherType 匹配器类型
type MatcherType int

const (
	MatcherContent MatcherType = iota
	MatcherPCRE
)

// StickyBuffer 匹配缓冲区位置
type StickyBuffer int

const (
	StickyNone       StickyBuffer = iota // 匹配完整 payload
	StickyHTTPURI                        // http.uri
	StickyHTTPHeader                     // http.header
	StickyHTTPBody                       // http.request_body
	StickyHTTPMethod                     // http.method
)
