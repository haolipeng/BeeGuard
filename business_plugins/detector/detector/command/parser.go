package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-libaudit/v2"
	"github.com/elastic/go-libaudit/v2/auparse"
	"go.uber.org/zap"
)

// ExecveEvent execve事件
type ExecveEvent struct {
	Timestamp time.Time
	PID       int
	PPID      int
	UID       int
	GID       int
	Username  string // comm 字段
	Exe       string // 可执行文件路径
	Cmdline   string // 完整命令行
	Cwd       string // 工作目录
	TTY       string
	Key       string // 审计规则key
}

// ParseExecveEvent 解析execve审计事件
func ParseExecveEvent(raw *libaudit.RawAuditMessage) *ExecveEvent {
	if raw == nil {
		return nil
	}

	// 解析审计消息
	msg, err := auparse.ParseLogLine(string(raw.Data))
	if err != nil {
		zap.S().Debugf("failed to parse audit log line: %v", err)
		return nil
	}

	// 只处理SYSCALL类型
	if msg.RecordType != auparse.AUDIT_SYSCALL {
		return nil
	}

	// 检查是否是execve (syscall 59 on x86_64, 11 on x86)
	syscallNum := msg.Data["syscall"]
	if syscallNum != "59" && syscallNum != "11" && syscallNum != "execve" {
		return nil
	}

	// 检查是否成功执行
	if msg.Data["success"] != "yes" {
		return nil
	}

	event := &ExecveEvent{
		Timestamp: msg.Timestamp,
	}

	// 提取进程信息
	if pid, err := strconv.Atoi(msg.Data["pid"]); err == nil {
		event.PID = pid
	}
	if ppid, err := strconv.Atoi(msg.Data["ppid"]); err == nil {
		event.PPID = ppid
	}
	if uid, err := strconv.Atoi(msg.Data["uid"]); err == nil {
		event.UID = uid
	}
	if gid, err := strconv.Atoi(msg.Data["gid"]); err == nil {
		event.GID = gid
	}

	// 提取其他字段
	event.Exe = unquoteAuditValue(msg.Data["exe"])
	event.TTY = msg.Data["tty"]
	event.Username = msg.Data["comm"]
	event.Key = msg.Data["key"]

	// 命令行参数需要从 EXECVE 记录中获取
	// 简化版：使用 comm 和 exe 组合
	if event.Username != "" {
		event.Cmdline = event.Username
	} else {
		event.Cmdline = event.Exe
	}

	return event
}

// ParseExecveArgs 从 EXECVE 类型记录中解析命令行参数
// 审计系统会将完整命令行分成多条记录
func ParseExecveArgs(raw *libaudit.RawAuditMessage) []string {
	if raw == nil {
		return nil
	}

	msg, err := auparse.ParseLogLine(string(raw.Data))
	if err != nil {
		return nil
	}

	if msg.RecordType != auparse.AUDIT_EXECVE {
		return nil
	}

	// 获取参数数量
	argc, err := strconv.Atoi(msg.Data["argc"])
	if err != nil {
		return nil
	}

	args := make([]string, 0, argc)
	for i := 0; i < argc; i++ {
		key := "a" + strconv.Itoa(i)
		if val, ok := msg.Data[key]; ok {
			args = append(args, unquoteAuditValue(val))
		}
	}

	return args
}

// unquoteAuditValue 解码审计值
// 审计系统可能会将字符串编码为十六进制
func unquoteAuditValue(s string) string {
	if s == "" {
		return s
	}

	// 移除引号
	s = strings.Trim(s, "\"")

	// 检查是否是十六进制编码
	if len(s) > 0 && isHexString(s) {
		decoded, err := hexDecode(s)
		if err == nil {
			return decoded
		}
	}

	return s
}

// isHexString 检查是否是十六进制字符串
func isHexString(s string) bool {
	if len(s)%2 != 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// hexDecode 解码十六进制字符串
func hexDecode(s string) (string, error) {
	bytes := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		b, err := strconv.ParseUint(s[i:i+2], 16, 8)
		if err != nil {
			return "", err
		}
		bytes[i/2] = byte(b)
	}
	// 去除结尾的 null 字符
	result := strings.TrimRight(string(bytes), "\x00")
	return result, nil
}
