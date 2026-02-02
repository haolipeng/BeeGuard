package ssh

import (
	"regexp"
	"time"
)

// 预编译的正则表达式
var (
	// syslog时间格式: "Jan  2 15:04:05" 或 "Jan 02 15:04:05"
	syslogTimeRegex = regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})`)

	// Accepted password for root from 192.168.1.100 port 22 ssh2
	acceptedPasswordRegex = regexp.MustCompile(`Accepted password for (\S+) from (\S+) port (\d+)`)

	// Accepted publickey for root from 192.168.1.100 port 22 ssh2
	acceptedPublickeyRegex = regexp.MustCompile(`Accepted publickey for (\S+) from (\S+) port (\d+)`)
)

// ParsedLogin 解析后的成功登录信息
type ParsedLogin struct {
	Timestamp time.Time
	Username  string
	SourceIP  string
	Port      string
	AuthType  string // "password" or "publickey"
}

// ParseSuccessLogin 解析成功登录日志行
func ParseSuccessLogin(line string) *ParsedLogin {
	// 尝试匹配密码认证成功
	if matches := acceptedPasswordRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLogin{
			Timestamp: parseTimestamp(line),
			Username:  matches[1],
			SourceIP:  matches[2],
			Port:      matches[3],
			AuthType:  "password",
		}
	}

	// 尝试匹配公钥认证成功
	if matches := acceptedPublickeyRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLogin{
			Timestamp: parseTimestamp(line),
			Username:  matches[1],
			SourceIP:  matches[2],
			Port:      matches[3],
			AuthType:  "publickey",
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
