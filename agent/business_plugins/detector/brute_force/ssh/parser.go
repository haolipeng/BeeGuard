package ssh

import (
	"regexp"
	"strconv"
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

	// Accepted password for root from 192.168.1.100 port 22 ssh2
	acceptedPasswordRegex = regexp.MustCompile(`Accepted password for (\S+) from (\S+) port (\d+)`)

	// Accepted publickey for root from 192.168.1.100 port 22 ssh2
	acceptedPublickeyRegex = regexp.MustCompile(`Accepted publickey for (\S+) from (\S+) port (\d+)`)

	// syslog "message repeated N times" 格式
	// message repeated 5 times: [ Failed password for root from 10.107.12.70 port 32369 ssh2]
	messageRepeatedRegex = regexp.MustCompile(`message repeated (\d+) times: \[ (.*)\]`)
)

// ParsedLog 解析后的日志信息
type ParsedLog struct {
	Timestamp time.Time
	SourceIP  string
	Username  string
	Action    string // "failed", "invalid_user", or "success"
	Count     int    // 事件重复次数（默认为 1，syslog "message repeated N times" 时为 N）
}

// ParseLine 解析单行SSH日志
func ParseLine(line string) *ParsedLog {
	// 先检查是否为 "message repeated N times" 格式
	if matches := messageRepeatedRegex.FindStringSubmatch(line); matches != nil {
		count, _ := strconv.Atoi(matches[1])
		innerLine := matches[2]
		// 递归解析内部日志内容
		parsed := parseSingleLine(innerLine)
		if parsed != nil {
			parsed.Timestamp = parseTimestamp(line) // 时间戳取外层
			parsed.Count = count
			return parsed
		}
		return nil
	}

	parsed := parseSingleLine(line)
	if parsed != nil {
		parsed.Count = 1
	}
	return parsed
}

// parseSingleLine 解析单条SSH日志内容（不含 message repeated 包装）
func parseSingleLine(line string) *ParsedLog {
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

	// 尝试匹配成功登录（密码认证）
	if matches := acceptedPasswordRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLog{
			Timestamp: parseTimestamp(line),
			Username:  matches[1],
			SourceIP:  matches[2],
			Action:    "success",
		}
	}

	// 尝试匹配成功登录（公钥认证）
	if matches := acceptedPublickeyRegex.FindStringSubmatch(line); matches != nil {
		return &ParsedLog{
			Timestamp: parseTimestamp(line),
			Username:  matches[1],
			SourceIP:  matches[2],
			Action:    "success",
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
