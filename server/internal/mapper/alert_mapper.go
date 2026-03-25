package mapper

import (
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/geoip"
	"github.com/haolipeng/BeeGuard/server/internal/models/alert"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// MapBruteForceAlert 暴力破解告警字段映射
// Agent字段: service, source_ip, target_user, count, first_seen, last_seen
func MapBruteForceAlert(fields map[string]string, ctx *AgentContext, geoIPSvc *geoip.Service) *alert.BruteForce {
	attemptCount, _ := strconv.Atoi(fields["count"])

	// 时间戳转换
	var firstAttackTime, attackTime time.Time
	if ts, err := strconv.ParseInt(fields["first_seen"], 10, 64); err == nil && ts > 0 {
		firstAttackTime = time.Unix(ts, 0)
	}
	if ts, err := strconv.ParseInt(fields["last_seen"], 10, 64); err == nil && ts > 0 {
		attackTime = time.Unix(ts, 0)
	}
	if attackTime.IsZero() {
		attackTime = time.Now()
	}

	// 根据攻击类型确定目标端口
	attackType := fields["service"] // ssh/ftp
	var targetPort int32
	switch attackType {
	case "ssh":
		targetPort = 22
	case "ftp":
		targetPort = 21
	}

	// GeoIP 查询
	sourceIP := fields["source_ip"]
	var sourceLocation *string
	if geoIPSvc != nil && sourceIP != "" {
		result := geoIPSvc.Query(sourceIP)
		if result.Country != "" && result.Country != "Unknown" {
			sourceLocation = &result.Country
		}
	}

	result := fields["result"]

	return &alert.BruteForce{
		AgentID:         ctx.AgentID,
		HostName:        ctx.HostName,
		HostIP:          strings.Join(ctx.HostIP, ","),
		SourceIP:        sourceIP,
		SourceLocation:  sourceLocation,
		AttackType:      attackType,
		TargetIP:        firstIP(ctx.HostIP), // 目标IP即为主机IP
		TargetPort:      &targetPort,
		Username:        fields["target_user"],
		AttemptCount:    int32(attemptCount),
		Result:          result,
		AttackTime:      common.DateTime{Time: attackTime},
		FirstAttackTime: toDateTimePtr(firstAttackTime),
		Status:          0, // 默认待处理
		UpdatedAt:       common.DateTime{Time: time.Now()},
	}
}

// MapDangerousCommandAlert 高危命令告警字段映射
// Agent字段: command, command_type, user, privilege_level, timestamp
func MapDangerousCommandAlert(fields map[string]string, ctx *AgentContext) *alert.DangerousCommand {
	var alertTime time.Time
	if ts, err := strconv.ParseInt(fields["timestamp"], 10, 64); err == nil && ts > 0 {
		alertTime = time.Unix(ts, 0)
	} else {
		alertTime = time.Now()
	}

	return &alert.DangerousCommand{
		AgentID:        ctx.AgentID,
		HostName:       ctx.HostName,
		HostIP:         strings.Join(ctx.HostIP, ","),
		Command:        fields["command"],
		CommandType:    fields["command_type"],
		User:           fields["user"],
		PrivilegeLevel: fields["privilege_level"],
		Status:         0,
		AlertTime:      common.DateTime{Time: alertTime},
		UpdatedAt:      common.DateTime{Time: time.Now()},
	}
}

// MapReverseShellAlert 反弹Shell告警字段映射
// Agent字段: comm, exe_path, args, remote_ip, remote_port, pid, ppid, uid, tgid
func MapReverseShellAlert(fields map[string]string, agentCtx *AgentContext, timestamp int64) *alert.ReverseShell {
	// 构建 CommandLine: 优先使用 args，若为空则用 exe_path + " " + args
	commandLine := fields["args"]
	if commandLine == "" {
		exePath := fields["exe_path"]
		if exePath != "" {
			commandLine = exePath
		}
	}

	// 推断 ShellType
	shellType := inferShellType(fields["comm"], fields["exe_path"])
	var shellTypePtr *string
	if shellType != "" {
		shellTypePtr = &shellType
	}

	// 目标端口
	targetPort := parsePort(fields["remote_port"], "remote_port")

	// 事件时间: 使用Record级别的timestamp
	var eventTime time.Time
	if timestamp > 0 {
		eventTime = time.Unix(timestamp, 0)
	} else {
		eventTime = time.Now()
	}

	return &alert.ReverseShell{
		AgentID:     agentCtx.AgentID,
		HostName:    agentCtx.HostName,
		VictimIP:    firstIP(agentCtx.HostIP),
		CommandLine: commandLine,
		ShellType:   shellTypePtr,
		TargetHost:  fields["remote_ip"],
		TargetPort:  int32(targetPort),
		Status:      0,
		EventTime:   common.DateTime{Time: eventTime},
		UpdatedAt:   common.DateTime{Time: time.Now()},
	}
}

// inferShellType 从 comm 或 exe_path 推断 Shell 类型
func inferShellType(comm, exePath string) string {
	// 取 exe_path 的 basename 用于匹配
	baseName := filepath.Base(exePath)
	// 统一转小写便于匹配
	commLower := strings.ToLower(comm)
	baseLower := strings.ToLower(baseName)

	// 按优先级匹配（先匹配更具体的，避免 "nc" 误匹配 "ncat" 等）
	patterns := []struct {
		keywords  []string
		shellType string
	}{
		{[]string{"python"}, alert.ShellTypePython},
		{[]string{"ncat", "netcat"}, alert.ShellTypeNc},
		{[]string{"perl"}, alert.ShellTypePerl},
		{[]string{"php"}, alert.ShellTypePHP},
		{[]string{"ruby"}, alert.ShellTypeRuby},
		{[]string{"bash", "sh"}, alert.ShellTypeBash},
		{[]string{"nc"}, alert.ShellTypeNc},
	}

	for _, p := range patterns {
		for _, kw := range p.keywords {
			if matchCommand(commLower, kw) || matchCommand(baseLower, kw) {
				return p.shellType
			}
		}
	}

	// 未匹配到已知类型，返回 comm 原值
	if comm != "" {
		return comm
	}
	return baseName
}

// matchCommand 精确匹配命令名，兼容带版本号后缀的情况
// 例如 keyword="python" 能匹配 "python", "python3", "python3.11"
// 但 keyword="sh" 不会匹配 "fish", "zsh", "ssh"
func matchCommand(name, keyword string) bool {
	if name == keyword {
		return true
	}
	if strings.HasPrefix(name, keyword) && len(name) > len(keyword) {
		// keyword 后面紧跟非字母字符（数字、点号等）才算匹配
		next := name[len(keyword)]
		return next < 'a' || next > 'z'
	}
	return false
}

// MapAbnormalLoginAlert 异常登录告警字段映射
// Agent字段: alert_type, service, rule_name, description, source_ip, target_user, count, timeframe, first_seen, last_seen, level
func MapAbnormalLoginAlert(fields map[string]string, ctx *AgentContext, geoIPSvc *geoip.Service) *alert.AbnormalLogin {
	// 时间戳转换 - 使用 last_seen 作为登录时间
	var loginTime time.Time
	if ts, err := strconv.ParseInt(fields["last_seen"], 10, 64); err == nil && ts > 0 {
		loginTime = time.Unix(ts, 0)
	} else {
		loginTime = time.Now()
	}

	// 风险等级转换: level(1-10) -> risk_level(string)
	level, _ := strconv.Atoi(fields["level"])
	riskLevel := convertRiskLevel(level)

	abnormalType := fields["abnormal_type"]
	switch abnormalType {
	case alert.AbnormalTypeUnknownIP, alert.AbnormalTypeTime, alert.AbnormalTypeUser:
		// 合法值，保持不变
	default:
		abnormalType = alert.AbnormalTypeUnknownIP
	}

	// GeoIP 查询
	sourceIP := fields["source_ip"]
	var sourceCountry *string
	if geoIPSvc != nil && sourceIP != "" {
		result := geoIPSvc.Query(sourceIP)
		if result.Country != "" && result.Country != "Unknown" {
			sourceCountry = &result.Country
		}
	}

	return &alert.AbnormalLogin{
		AgentID:        ctx.AgentID,
		HostName:       ctx.HostName,
		HostIP:         strings.Join(ctx.HostIP, ","),
		SourceIP:       sourceIP,
		SourceLocation: sourceCountry,
		LoginUser:      fields["target_user"], // Agent: target_user -> DB: login_user
		LoginTime:      common.DateTime{Time: loginTime},
		RiskLevel:      riskLevel,
		AbnormalType:   &abnormalType,
		Status:         0, // 待处理
		UpdatedAt:      common.DateTime{Time: time.Now()},
	}
}

// MapPrivilegeEscalationAlert 本地提权告警字段映射
// Agent字段: escalated_user, parent_process, parent_process_user, tgid, exe_path, timestamp
func MapPrivilegeEscalationAlert(fields map[string]string, ctx *AgentContext) *alert.PrivilegeEscalation {
	processID, _ := strconv.Atoi(fields["tgid"])
	processID32 := int32(processID)

	var discoverTime time.Time
	if ts, err := strconv.ParseInt(fields["timestamp"], 10, 64); err == nil && ts > 0 {
		discoverTime = time.Unix(ts, 0)
	} else {
		discoverTime = time.Now()
	}

	processPath := fields["exe_path"]
	var processPathPtr *string
	if processPath != "" {
		processPathPtr = &processPath
	}

	return &alert.PrivilegeEscalation{
		AgentID:           ctx.AgentID,
		HostName:          ctx.HostName,
		HostIP:            strings.Join(ctx.HostIP, ","),
		EscalatedUser:     fields["escalated_user"],
		ParentProcess:     fields["parent_process"],
		ParentProcessUser: fields["parent_process_user"],
		ProcessID:         &processID32,
		ProcessPath:       processPathPtr,
		Status:            0,
		DiscoverTime:      common.DateTime{Time: discoverTime},
		UpdatedAt:         common.DateTime{Time: time.Now()},
	}
}

// convertRiskLevel 风险等级转换: 数字(1-10) -> 字符串
func convertRiskLevel(level int) string {
	switch {
	case level >= 8:
		return alert.RiskLevelCritical // 危急
	case level >= 6:
		return alert.RiskLevelHigh // 高危
	case level >= 4:
		return alert.RiskLevelMedium // 中危
	case level >= 2:
		return alert.RiskLevelLow // 低危
	default:
		return alert.RiskLevelLow // 默认低危
	}
}

// MapMaliciousRequestAlert 恶意请求告警字段映射（DataType 6008）
// Agent字段: event_type, rule_id, rule_name, severity, threat_type, indicator_type,
//
//	matched_value, remote_ip, remote_port, domain, description, timestamp等
func MapMaliciousRequestAlert(fields map[string]string, ctx *AgentContext, timestamp int64) *alert.MaliciousRequest {
	// 提取事件时间
	var eventTime time.Time
	if timestamp > 0 {
		eventTime = time.Unix(timestamp, 0)
	} else if ts, err := strconv.ParseInt(fields["timestamp"], 10, 64); err == nil && ts > 0 {
		eventTime = time.Unix(ts, 0)
	} else {
		eventTime = time.Now()
	}

	// 提取域名或IP（优先domain）
	var maliciousDomain string
	var maliciousIP *string

	eventType := fields["event_type"] // "connect" 或 "dns"

	if eventType == "dns" {
		// DNS事件：domain字段存在
		domain := fields["domain"]
		if domain == "" {
			domain = fields["matched_value"]
		}
		maliciousDomain = normalizeDomain(domain)

		// DNS事件也可能有remote_ip（DNS解析结果）
		if remoteIP := fields["remote_ip"]; remoteIP != "" {
			maliciousIP = &remoteIP
		}
	} else {
		// CONNECT事件：只有remote_ip
		if remoteIP := fields["remote_ip"]; remoteIP != "" {
			// 检查remote_ip是否是有效的IP地址
			if isValidIP(remoteIP) {
				maliciousIP = &remoteIP
			} else {
				// 如果不是IP，可能是域名（异常情况）
				maliciousDomain = normalizeDomain(remoteIP)
			}
		}
	}

	// Fallback：如果两者都为空，使用matched_value
	if maliciousDomain == "" && maliciousIP == nil {
		matchedValue := fields["matched_value"]
		if isValidIP(matchedValue) {
			maliciousIP = &matchedValue
		} else {
			maliciousDomain = normalizeDomain(matchedValue)
		}
	}

	// 字段映射
	threatType := fields["threat_type"]
	policyType := mapThreatTypeToPolicyType(threatType)
	description := fields["description"]

	eventTimeDT := common.DateTime{Time: eventTime}

	return &alert.MaliciousRequest{
		AgentID:          ctx.AgentID,
		HostName:         ctx.HostName,
		HostIP:           strings.Join(ctx.HostIP, ","),
		PolicyType:       policyType,
		PolicyName:       fields["rule_name"],
		MaliciousDomain:  maliciousDomain,
		MaliciousIP:      maliciousIP,
		RequestCount:     1, // 初始值，聚合时会更新
		FirstRequestTime: &eventTimeDT,
		LastRequestTime:  &eventTimeDT,
		RiskDescription:  &description,
		Status:           0, // 0-待处理
	}
}

// mapThreatTypeToPolicyType 威胁类型映射到策略类型
func mapThreatTypeToPolicyType(threatType string) string {
	switch strings.ToLower(threatType) {
	case "mining":
		return "behavior_analyze"
	case "c2":
		return "threat_intel"
	case "phishing":
		return "url_filter"
	case "botnet":
		return "ip_blacklist"
	case "ransomware":
		return "threat_intel"
	default:
		return "domain_filter"
	}
}

// normalizeDomain 规范化域名（去除协议、端口、尾部点号，转小写）
func normalizeDomain(domain string) string {
	if domain == "" {
		return ""
	}
	domain = strings.ToLower(domain)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.Split(domain, ":")[0]   // 去除端口
	domain = strings.TrimSuffix(domain, ".") // DNS FQDN格式
	return domain
}

// isValidIP 检查字符串是否为有效IP地址
func isValidIP(s string) bool {
	return net.ParseIP(s) != nil
}

// MapMalwareScanAlert 恶意文件扫描告警字段映射（DataType 6061/6062）
// Agent字段: threat_type, file_name, file_path, file_size, file_md5, file_sha256,
//
//	detection_engine, malware_family, scan_time
func MapMalwareScanAlert(fields map[string]string, ctx *AgentContext) *alert.MalwareScan {
	var scanTime time.Time
	if ts, err := strconv.ParseInt(fields["scan_time"], 10, 64); err == nil && ts > 0 {
		scanTime = time.Unix(ts, 0)
	} else {
		scanTime = time.Now()
	}

	var fileSize *int64
	if s, err := strconv.ParseInt(fields["file_size"], 10, 64); err == nil {
		fileSize = &s
	}

	var fileMD5, fileSHA256, detectionEngine, malwareFamily *string
	if v := fields["file_md5"]; v != "" {
		fileMD5 = &v
	}
	if v := fields["file_sha256"]; v != "" {
		fileSHA256 = &v
	}
	if v := fields["detection_engine"]; v != "" {
		detectionEngine = &v
	}
	if v := fields["malware_family"]; v != "" {
		malwareFamily = &v
	}

	// 规范化威胁类型（转小写以匹配数据库枚举值）
	threatType := strings.ToLower(fields["threat_type"])

	return &alert.MalwareScan{
		AgentID:         ctx.AgentID,
		HostName:        ctx.HostName,
		HostIP:          strings.Join(ctx.HostIP, ","),
		ThreatType:      threatType,
		FileName:        fields["file_name"],
		FilePath:        fields["file_path"],
		FileSize:        fileSize,
		FileMD5:         fileMD5,
		FileSHA256:      fileSHA256,
		DetectionEngine: detectionEngine,
		MalwareFamily:   malwareFamily,
		Status:          0, // 待处理
		ScanTime:        common.DateTime{Time: scanTime},
		UpdatedAt:       common.DateTime{Time: time.Now()},
	}
}

// MapContainerDangerousCommandAlert 容器高危命令告警字段映射
// Agent字段: command, rule_id, rule_name, severity, uid, timestamp, container_id, container_name, image_name
func MapContainerDangerousCommandAlert(fields map[string]string, ctx *AgentContext) *alert.ContainerDangerousCommand {
	var alertTime time.Time
	if ts, err := strconv.ParseInt(fields["timestamp"], 10, 64); err == nil && ts > 0 {
		alertTime = time.Unix(ts, 0)
	} else {
		alertTime = time.Now()
	}

	// privilege_level: uid=="0" → root, 否则 normal
	privilegeLevel := "normal"
	if fields["uid"] == "0" {
		privilegeLevel = "root"
	}

	// container_name: 空字符串则设为 nil
	var containerName *string
	if v := fields["container_name"]; v != "" {
		containerName = &v
	}

	// image_name: 空字符串则设为 nil
	var imageName *string
	if v := fields["image_name"]; v != "" {
		imageName = &v
	}

	return &alert.ContainerDangerousCommand{
		AgentID:        ctx.AgentID,
		HostName:       ctx.HostName,
		HostIP:         strings.Join(ctx.HostIP, ","),
		ContainerID:    fields["container_id"],
		ContainerName:  containerName,
		ImageName:      imageName,
		Command:        fields["command"],
		CommandType:    fields["rule_id"],
		User:           fields["uid"],
		PrivilegeLevel: privilegeLevel,
		Status:         0,
		AlertTime:      common.DateTime{Time: alertTime},
		UpdatedAt:      common.DateTime{Time: time.Now()},
	}
}

// MapContainerReverseShellAlert 容器反弹Shell告警字段映射
// Agent字段: pid, ppid, uid, comm, exe_path, args, remote_ip, remote_port, container_id, container_name, image_name
func MapContainerReverseShellAlert(fields map[string]string, ctx *AgentContext, timestamp int64) *alert.ContainerReverseShell {
	// pid / ppid / remote_port
	pid, _ := strconv.Atoi(fields["pid"])
	remotePort := parsePort(fields["remote_port"], "remote_port")

	var ppid *int32
	if v, err := strconv.Atoi(fields["ppid"]); err == nil {
		v32 := int32(v)
		ppid = &v32
	}

	// nullable 字符串字段：空字符串转 nil
	var containerName *string
	if v := fields["container_name"]; v != "" {
		containerName = &v
	}
	var imageName *string
	if v := fields["image_name"]; v != "" {
		imageName = &v
	}
	var exePath *string
	if v := fields["exe_path"]; v != "" {
		exePath = &v
	}
	var args *string
	if v := fields["args"]; v != "" {
		args = &v
	}

	// 推断 shell_type（复用 inferShellType）
	shellType := inferShellType(fields["comm"], fields["exe_path"])
	var shellTypePtr *string
	if shellType != "" {
		shellTypePtr = &shellType
	}

	// 事件时间：优先使用 Record 级 timestamp
	var eventTime time.Time
	if timestamp > 0 {
		eventTime = time.Unix(timestamp, 0)
	} else {
		eventTime = time.Now()
	}

	return &alert.ContainerReverseShell{
		AgentID:       ctx.AgentID,
		HostName:      ctx.HostName,
		HostIP:        strings.Join(ctx.HostIP, ","),
		ContainerID:   fields["container_id"],
		ContainerName: containerName,
		ImageName:     imageName,
		PID:           int32(pid),
		PPID:          ppid,
		UID:           fields["uid"],
		Comm:          fields["comm"],
		ExePath:       exePath,
		Args:          args,
		ShellType:     shellTypePtr,
		RemoteIP:      fields["remote_ip"],
		RemotePort:    int32(remotePort),
		Status:        0,
		EventTime:     common.DateTime{Time: eventTime},
		UpdatedAt:     common.DateTime{Time: time.Now()},
	}
}

// MapContainerSensitiveFileAlert 容器核心文件监控告警字段映射
// Agent字段: action, new_path, old_path, rule_id, rule_name, severity, rule_description,
//
//	matched_pattern, operator_user, operator_process, container_id, container_name, image_name, timestamp
func MapContainerSensitiveFileAlert(fields map[string]string, ctx *AgentContext) *alert.ContainerSensitiveFile {
	var alertTime time.Time
	if ts, err := strconv.ParseInt(fields["timestamp"], 10, 64); err == nil && ts > 0 {
		alertTime = time.Unix(ts, 0)
	} else {
		alertTime = time.Now()
	}

	// nullable 字符串字段：空字符串转 nil
	var containerName *string
	if v := fields["container_name"]; v != "" {
		containerName = &v
	}
	var imageName *string
	if v := fields["image_name"]; v != "" {
		imageName = &v
	}
	var ruleDescription *string
	if v := fields["rule_description"]; v != "" {
		ruleDescription = &v
	}
	var matchedPattern *string
	if v := fields["matched_pattern"]; v != "" {
		matchedPattern = &v
	}
	var oldPath *string
	if v := fields["old_path"]; v != "" {
		oldPath = &v
	}
	var operatorUser *string
	if v := fields["operator_user"]; v != "" {
		operatorUser = &v
	}
	var operatorProcess *string
	if v := fields["operator_process"]; v != "" {
		operatorProcess = &v
	}

	return &alert.ContainerSensitiveFile{
		AgentID:         ctx.AgentID,
		HostName:        ctx.HostName,
		HostIP:          strings.Join(ctx.HostIP, ","),
		ContainerID:     fields["container_id"],
		ContainerName:   containerName,
		ImageName:       imageName,
		RuleID:          fields["rule_id"],
		RuleName:        fields["rule_name"],
		Severity:        fields["severity"],
		RuleDescription: ruleDescription,
		MatchedPattern:  matchedPattern,
		Action:          fields["action"],
		FilePath:        fields["new_path"],
		OldPath:         oldPath,
		OperatorUser:    operatorUser,
		OperatorProcess: operatorProcess,
		Status:          0,
		AlertTime:       common.DateTime{Time: alertTime},
		UpdatedAt:       common.DateTime{Time: time.Now()},
	}
}

// MapNetworkAttackAlert 网络攻击告警字段映射
// Agent字段: src_ip, dst_ip, dst_port, src_port, vulnerability_name, attack_status,
//
//	sid, reference, attack_count, first_attack_time, last_attack_time, matched_payload
func MapNetworkAttackAlert(fields map[string]string, ctx *AgentContext, timestamp int64,
	geoIPSvc *geoip.Service) *alert.NetworkAttack {

	// 端口和计数转换
	targetPort := parsePort(fields["dst_port"], "dst_port")
	attackCount, _ := strconv.Atoi(fields["attack_count"])

	// 时间处理：优先使用 timestamp 参数
	var firstAttackTime, lastAttackTime time.Time
	if timestamp > 0 {
		lastAttackTime = time.Unix(timestamp, 0)
	} else if ts, err := strconv.ParseInt(fields["last_attack_time"], 10, 64); err == nil && ts > 0 {
		lastAttackTime = time.Unix(ts, 0)
	} else {
		lastAttackTime = time.Now()
	}

	if ts, err := strconv.ParseInt(fields["first_attack_time"], 10, 64); err == nil && ts > 0 {
		firstAttackTime = time.Unix(ts, 0)
	} else {
		firstAttackTime = lastAttackTime
	}

	// GeoIP 查询
	attackerIP := fields["src_ip"]
	var attackerCountry, attackerLocation *string
	if geoIPSvc != nil && attackerIP != "" {
		result := geoIPSvc.Query(attackerIP)
		if result.Country != "" && result.Country != "Unknown" {
			attackerCountry = &result.Country
		}
		if result.Location != "" && result.Location != "Unknown" {
			attackerLocation = &result.Location
		}
	}

	// vulnerability_id: 使用 reference 字段（可能为空字符串）
	vulnerabilityID := fields["reference"]
	var vulnIDPtr *string
	if vulnerabilityID != "" {
		vulnIDPtr = &vulnerabilityID
	}

	// attack_payload
	attackPayload := fields["matched_payload"]
	var payloadPtr *string
	if attackPayload != "" {
		payloadPtr = &attackPayload
	}

	firstAttackTimeDT := common.DateTime{Time: firstAttackTime}

	return &alert.NetworkAttack{
		AgentID:           ctx.AgentID,
		HostName:          ctx.HostName,
		HostIP:            strings.Join(ctx.HostIP, ","),
		TargetPort:        int32(targetPort),
		AttackerIP:        attackerIP,
		AttackerLocation:  attackerLocation,
		AttackerCountry:   attackerCountry,
		VulnerabilityName: fields["vulnerability_name"],
		VulnerabilityID:   vulnIDPtr,
		AttackStatus:      alert.AttackStatusDetected, // 统一映射为 "detected",等开始实现响应处置功能后，才有阻断网络攻击的功能
		AttackCount:       int32(attackCount),
		FirstAttackTime:   &firstAttackTimeDT,
		LastAttackTime:    common.DateTime{Time: lastAttackTime},
		AttackPayload:     payloadPtr,
		Status:            0, // 0-待处理
	}
}
