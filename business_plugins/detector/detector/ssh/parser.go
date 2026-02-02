package ssh

import (
	"regexp"
	"time"
)

// 预编译的正则表达式
var (
	// syslog时间格式: "Jan  2 15:04:05" 或 "Jan 02 15:04:05"
	syslogTimeRegex = regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})`)

	// Failed password for root from 192.168.1.100 port 22 ssh2
	// Failed password for invalid user admin from 192.168.1.100 port 22 ssh2
	failedPasswordRegex = regexp.MustCompile(`Failed (password|publickey) for (?:invalid user )?(\S+) from (\S+)`)

	// Invalid user admin from 192.168.1.100 port 22
	// Illegal user admin from 192.168.1.100 port 22
	invalidUserRegex = regexp.MustCompile(`(?:Invalid|Illegal) user (\S+) from (\S+)`)
)

// ParsedLog 解析后的日志信息
type ParsedLog struct {
	Timestamp time.Time
	SourceIP  string
	Username  string
	Action    string // "failed" or "invalid_user"
}

// ParseLine 解析单行SSH日志
func ParseLine(line string) *ParsedLog {
	// 尝试匹配认证失败
	if matches := failedPasswordRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLog{
			Timestamp: parseTimestamp(line),
			Username:  matches[2],
			SourceIP:  matches[3],
			Action:    "failed",
		}
	}

	// 尝试匹配非法用户
	if matches := invalidUserRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLog{
			Timestamp: parseTimestamp(line),
			Username:  matches[1],
			SourceIP:  matches[2],
			Action:    "invalid_user",
		}
	}

	return nil
}

// parseTimestamp 解析syslog时间戳
func parseTimestamp(line string) time.Time {
	matches := syslogTimeRegex.FindStringSubmatch(line)
	if matches == nil {
		return time.Now()
	}

	// syslog格式不包含年份，使用当前年份
	year := time.Now().Year()
	timeStr := matches[1]

	// 解析时间
	t, err := time.Parse("Jan  2 15:04:05", timeStr)
	if err != nil {
		t, err = time.Parse("Jan 2 15:04:05", timeStr)
		if err != nil {
			return time.Now()
		}
	}

	// 设置年份
	t = t.AddDate(year, 0, 0)

	// 处理跨年情况：如果解析的时间在未来，说明是去年的日志
	if t.After(time.Now().Add(24 * time.Hour)) {
		t = t.AddDate(-1, 0, 0)
	}

	return t
}
