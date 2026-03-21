package ftp

import (
	"regexp"
	"time"
)

// 预编译的正则表达式
var (
	// vsftpd时间格式: "Sun Jul 13 10:31:10 2008"
	vsftpdTimeRegex = regexp.MustCompile(`^(\w{3}\s+\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2}\s+\d{4})`)

	// FAIL LOGIN: Sun Jul 13 10:31:10 2008 [pid 27521] [test] FAIL LOGIN: Client "84.140.234.76"
	failLoginRegex = regexp.MustCompile(`\[pid \d+\] \[(\S+)\] FAIL LOGIN: Client "(\S+)"`)

	// OK LOGIN: Sun Jul 13 10:31:25 2008 [pid 27528] [admin] OK LOGIN: Client "192.168.1.1"
	okLoginRegex = regexp.MustCompile(`\[pid \d+\] \[(\S+)\] OK LOGIN: Client "(\S+)"`)

	// CONNECT: Sun Jul 13 10:31:05 2008 [pid 27528] CONNECT: Client "84.140.234.76"
	connectRegex = regexp.MustCompile(`\[pid \d+\] CONNECT: Client "(\S+)"`)
)

// ParsedLog 解析后的日志信息
type ParsedLog struct {
	Timestamp time.Time
	SourceIP  string
	Username  string
	Action    string // "failed" or "connect"
}

// ParseLine 解析单行vsftpd日志
func ParseLine(line string) *ParsedLog {
	// 尝试匹配登录失败
	if matches := failLoginRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLog{
			Timestamp: parseTimestamp(line),
			Username:  matches[1],
			SourceIP:  matches[2],
			Action:    "failed",
		}
	}

	// 尝试匹配连接事件
	if matches := connectRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLog{
			Timestamp: parseTimestamp(line),
			Username:  "",
			SourceIP:  matches[1],
			Action:    "connect",
		}
	}

	return nil
}

// parseTimestamp 解析vsftpd时间戳
func parseTimestamp(line string) time.Time {
	matches := vsftpdTimeRegex.FindStringSubmatch(line)
	if matches == nil {
		return time.Now()
	}

	timeStr := matches[1]

	// vsftpd格式: "Sun Jul 13 10:31:10 2008"
	t, err := time.Parse("Mon Jan 2 15:04:05 2006", timeStr)
	if err != nil {
		// 尝试另一种格式（月份日期可能有两位数）
		t, err = time.Parse("Mon Jan  2 15:04:05 2006", timeStr)
		if err != nil {
			return time.Now()
		}
	}

	return t
}
