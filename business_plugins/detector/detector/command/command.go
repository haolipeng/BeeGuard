package command

import (
	"regexp"
	"strconv"
	"sync"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/business_plugins/detector/audit"
	"gitlab.myinterest.top/security/agent/business_plugins/detector/config"
	"go.uber.org/zap"
)

const (
	// DataTypeDangerousCommand 高危命令告警数据类型
	DataTypeDangerousCommand = 6003
	// DataTypeReverseShell 反弹Shell告警数据类型
	DataTypeReverseShell = 6004
)

// CommandAlert 高危命令告警 (按 soc_tech_doc.md 规范)
type CommandAlert struct {
	Command        string    // 执行的命令内容
	CommandType    string    // 命令类型枚举
	User           string    // 执行用户
	PrivilegeLevel string    // 权限级别
	Timestamp      time.Time // 告警时间
}

// ReverseShellAlert 反弹Shell告警 (按 soc_tech_doc.md 规范)
type ReverseShellAlert struct {
	CommandLine string    // 反弹Shell命令行
	ShellType   string    // Shell类型
	TargetHost  string    // 目标主机(攻击者IP)
	TargetPort  int       // 目标端口
	Timestamp   time.Time // 事件时间
}

// CompiledRule 编译后的规则
type CompiledRule struct {
	Name           string
	Description    string
	Category       string // reverse_shell / privilege_escalation 等
	Pattern        *regexp.Regexp
	Level          int
	CommandType    string // 命令类型枚举值
	IsReverseShell bool   // 是否为反弹Shell规则
}

// Detector 高危命令检测器
type Detector struct {
	mu           sync.RWMutex
	config       config.CommandConfig
	rules        []*CompiledRule
	auditClient  *audit.Client
	pluginClient *businessplugins.Client
	done         chan struct{}
	wg           sync.WaitGroup
}

// New 创建高危命令检测器
func New(cfg config.CommandConfig, pluginClient *businessplugins.Client) *Detector {
	d := &Detector{
		config:       cfg,
		pluginClient: pluginClient,
		done:         make(chan struct{}),
	}
	d.compileRules()
	return d
}

// Name 返回检测器名称
func (d *Detector) Name() string {
	return "command"
}

// compileRules 编译检测规则
func (d *Detector) compileRules() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.rules = make([]*CompiledRule, 0)
	for _, rule := range d.config.Rules {
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			zap.S().Errorf("failed to compile rule pattern %s: %v", rule.Name, err)
			continue
		}

		compiled := &CompiledRule{
			Name:           rule.Name,
			Description:    rule.Description,
			Category:       rule.Category,
			Pattern:        pattern,
			Level:          rule.Level,
			CommandType:    rule.CommandType,
			IsReverseShell: rule.Category == "reverse_shell",
		}
		d.rules = append(d.rules, compiled)
		zap.S().Infof("compiled command detection rule: %s (category=%s)", rule.Name, rule.Category)
	}
}

// Start 启动检测器
func (d *Detector) Start() error {
	// 创建审计客户端
	client, err := audit.New()
	if err != nil {
		return err
	}
	d.auditClient = client

	// 配置execve监控规则
	if err := d.auditClient.SetupExecveRule(); err != nil {
		d.auditClient.Close()
		return err
	}

	// 启动事件处理循环
	d.wg.Add(1)
	go d.eventLoop()

	zap.S().Info("command detector started")
	return nil
}

// eventLoop 事件处理循环
func (d *Detector) eventLoop() {
	defer d.wg.Done()

	for {
		select {
		case <-d.done:
			return
		default:
			rawEvent, err := d.auditClient.Receive()
			if err != nil {
				// 检查是否是关闭导致的错误
				select {
				case <-d.done:
					return
				default:
					zap.S().Debugf("receive audit event error: %v", err)
					continue
				}
			}

			// 解析execve事件
			event := ParseExecveEvent(rawEvent)
			if event == nil {
				continue
			}

			// 白名单检查
			if d.isWhitelisted(event) {
				continue
			}

			// 规则匹配
			d.matchAndAlert(event)
		}
	}
}

// isWhitelisted 检查是否在白名单中
func (d *Detector) isWhitelisted(event *ExecveEvent) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 检查用户白名单
	for _, user := range d.config.Whitelist.Users {
		if event.Username == user {
			return true
		}
	}

	// 检查命令路径白名单
	for _, path := range d.config.Whitelist.ExePaths {
		if len(event.Exe) >= len(path) && event.Exe[:len(path)] == path {
			return true
		}
	}

	return false
}

// matchAndAlert 匹配规则并发送告警
func (d *Detector) matchAndAlert(event *ExecveEvent) {
	d.mu.RLock()
	rules := d.rules
	d.mu.RUnlock()

	cmdline := event.Cmdline
	if cmdline == "" {
		cmdline = event.Exe
	}

	for _, rule := range rules {
		if rule.Pattern.MatchString(cmdline) {
			if rule.IsReverseShell {
				d.sendReverseShellAlert(event, rule)
			} else {
				d.sendDangerousCommandAlert(event, rule)
			}
			return // 只匹配第一个规则
		}
	}
}

// sendDangerousCommandAlert 发送高危命令告警
func (d *Detector) sendDangerousCommandAlert(event *ExecveEvent, rule *CompiledRule) {
	cmdline := event.Cmdline
	if cmdline == "" {
		cmdline = event.Exe
	}

	privilegeLevel := "normal"
	if event.UID == 0 {
		privilegeLevel = "root"
	}

	commandType := rule.CommandType
	if commandType == "" {
		commandType = rule.Category
	}

	zap.S().Warnf("[DangerousCommand] rule=%s type=%s command=%s user=%s",
		rule.Name, commandType, truncateString(cmdline, 100), event.Username)

	rec := &businessplugins.Record{
		DataType:  int32(DataTypeDangerousCommand),
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"command":         cmdline,
				"command_type":    commandType,
				"user":            event.Username,
				"privilege_level": privilegeLevel,
				"timestamp":       strconv.FormatInt(event.Timestamp.Unix(), 10),
			},
		},
	}

	if err := d.pluginClient.SendRecord(rec); err != nil {
		zap.S().Errorf("failed to send dangerous command alert: %v", err)
	}
}

// sendReverseShellAlert 发送反弹Shell告警
func (d *Detector) sendReverseShellAlert(event *ExecveEvent, rule *CompiledRule) {
	cmdline := event.Cmdline
	if cmdline == "" {
		cmdline = event.Exe
	}

	// 解析目标主机和端口
	targetHost, targetPort := parseReverseShellTarget(cmdline)

	shellType := detectShellType(event.Exe, cmdline)

	zap.S().Warnf("[ReverseShell] rule=%s shell_type=%s target=%s:%d command=%s",
		rule.Name, shellType, targetHost, targetPort, truncateString(cmdline, 100))

	rec := &businessplugins.Record{
		DataType:  int32(DataTypeReverseShell),
		Timestamp: time.Now().Unix(),
		Data: &businessplugins.Payload{
			Fields: map[string]string{
				"command_line": cmdline,
				"shell_type":   shellType,
				"target_host":  targetHost,
				"target_port":  strconv.Itoa(targetPort),
				"timestamp":    strconv.FormatInt(event.Timestamp.Unix(), 10),
			},
		},
	}

	if err := d.pluginClient.SendRecord(rec); err != nil {
		zap.S().Errorf("failed to send reverse shell alert: %v", err)
	}
}

// Stop 停止检测器
func (d *Detector) Stop() error {
	close(d.done)

	if d.auditClient != nil {
		d.auditClient.Cleanup()
		d.auditClient.Close()
	}

	d.wg.Wait()
	zap.S().Info("command detector stopped")
	return nil
}

// UpdateConfig 更新检测器配置
func (d *Detector) UpdateConfig(data string) error {
	newCfg, err := config.ParseCommandConfigFromJSON(data)
	if err != nil {
		return err
	}

	d.mu.Lock()
	d.config = *newCfg
	d.mu.Unlock()

	d.compileRules()

	zap.S().Infof("command detector config updated: %d rules", len(newCfg.Rules))
	return nil
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parseReverseShellTarget 解析反弹Shell目标
func parseReverseShellTarget(cmdline string) (host string, port int) {
	// 匹配 /dev/tcp/IP/PORT 格式
	devTcpRe := regexp.MustCompile(`/dev/tcp/(\d+\.\d+\.\d+\.\d+)/(\d+)`)
	if matches := devTcpRe.FindStringSubmatch(cmdline); len(matches) == 3 {
		host = matches[1]
		port, _ = strconv.Atoi(matches[2])
		return
	}

	// 匹配 nc IP PORT 格式
	ncRe := regexp.MustCompile(`nc\s+(\d+\.\d+\.\d+\.\d+)\s+(\d+)`)
	if matches := ncRe.FindStringSubmatch(cmdline); len(matches) == 3 {
		host = matches[1]
		port, _ = strconv.Atoi(matches[2])
		return
	}

	return "unknown", 0
}

// detectShellType 检测Shell类型
func detectShellType(exe, cmdline string) string {
	switch {
	case regexp.MustCompile(`(?i)bash`).MatchString(exe) || regexp.MustCompile(`(?i)bash`).MatchString(cmdline):
		return "bash"
	case regexp.MustCompile(`(?i)python`).MatchString(exe) || regexp.MustCompile(`(?i)python`).MatchString(cmdline):
		return "python"
	case regexp.MustCompile(`(?i)\bnc\b|netcat`).MatchString(exe) || regexp.MustCompile(`(?i)\bnc\b|netcat`).MatchString(cmdline):
		return "nc"
	case regexp.MustCompile(`(?i)perl`).MatchString(exe) || regexp.MustCompile(`(?i)perl`).MatchString(cmdline):
		return "perl"
	case regexp.MustCompile(`(?i)php`).MatchString(exe) || regexp.MustCompile(`(?i)php`).MatchString(cmdline):
		return "php"
	case regexp.MustCompile(`(?i)ruby`).MatchString(exe) || regexp.MustCompile(`(?i)ruby`).MatchString(cmdline):
		return "ruby"
	case regexp.MustCompile(`(?i)powershell`).MatchString(exe) || regexp.MustCompile(`(?i)powershell`).MatchString(cmdline):
		return "powershell"
	default:
		return "bash"
	}
}
