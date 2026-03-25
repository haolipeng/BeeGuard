package mapper

import (
	"strconv"
	"strings"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/container"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
)

// 类型转换辅助函数
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int32Ptr(n int32) *int32      { return &n }
func int64Ptr(n int64) *int64      { return &n }
func nowDateTime() common.DateTime { return common.DateTime{Time: time.Now()} }
func toDateTimePtr(t time.Time) *common.DateTime {
	if t.IsZero() {
		return nil
	}
	dt := common.DateTime{Time: t}
	return &dt
}

// parsePort 解析端口字符串，失败时记录警告并返回 0
func parsePort(s string, fieldName string) int {
	if s == "" {
		return 0
	}
	p, err := strconv.Atoi(s)
	if err != nil {
		log.Warnf("[Mapper] invalid port value for %s: %q", fieldName, s)
		return 0
	}
	if p < 0 || p > 65535 {
		log.Warnf("[Mapper] port out of range for %s: %d", fieldName, p)
		return 0
	}
	return p
}

// AgentContext Agent上下文信息（从PackagedData头部获取）
type AgentContext struct {
	AgentID      string
	HostName     string
	HostIP       []string // 完整IPv4地址列表
	AgentVersion string
	MacAddr      string // MAC地址
	OsType       string // 操作系统类型
	OsVersion    string // 操作系统版本
}

// firstIP 取IP列表的第一个元素，空列表返回空字符串
func firstIP(ips []string) string {
	if len(ips) == 0 {
		return ""
	}
	return ips[0]
}

// MapPort 端口字段映射: Agent -> 数据库
// Agent字段: sport, sip, comm, username, protocol(数字)
// 数据库字段: port, listen_ip, listen_process, run_user, protocol(数字: 6=TCP, 17=UDP)
func MapPort(fields map[string]string, ctx *AgentContext) *host.Port {
	port := parsePort(fields["sport"], "sport")

	// 协议直接使用数字: 6=TCP, 17=UDP
	protocol, _ := strconv.Atoi(fields["protocol"])

	// 进程启动时间转换
	var processTime *common.DateTime
	if ts := fields["start_time"]; ts != "" {
		if sec, err := strconv.ParseInt(ts, 10, 64); err == nil && sec > 0 {
			processTime = toDateTimePtr(time.Unix(sec, 0))
		}
	}

	return &host.Port{
		AgentID:       ctx.AgentID,
		HostName:      ctx.HostName,
		HostIP:        strings.Join(ctx.HostIP, ","),
		OsType:        strPtr(ctx.OsType),
		AgentVersion:  strPtr(ctx.AgentVersion),
		AgentStatus:   1, // 在线
		Port:          int32(port),
		Protocol:      int16(protocol),
		ListenIP:      fields["sip"],
		ListenProcess: fields["comm"],
		RunUser:       strPtr(fields["username"]),
		ProcessTime:   processTime,
		UpdatedAt:     nowDateTime(),
	}
}

// MapAccount 账号字段映射: Agent -> 数据库
// Agent字段: username, uid, is_root, is_sudo, is_expired, shell, last_login_time
// 数据库字段: username, uid, acc_status, permission, login_type, last_login_time
func MapAccount(fields map[string]string, ctx *AgentContext) *host.Account {
	uid, _ := strconv.Atoi(fields["uid"])

	// 账号状态转换: 0=正常 1=即将过期 2=已过期
	var accStatus int16 = 0 // 默认正常
	if fields["is_expired"] == "true" {
		accStatus = 2 // 已过期
	} else if fields["is_expiring_soon"] == "true" {
		accStatus = 1 // 即将过期
	}

	// 权限转换: is_root + is_sudo -> permission (使用英文)
	var permissions []string
	if fields["is_root"] == "true" {
		permissions = append(permissions, "root")
	}
	if fields["is_sudo"] == "true" {
		permissions = append(permissions, "sudo")
	}
	permission := strings.Join(permissions, ",")
	if permission == "" {
		permission = "normal" // 普通用户
	}

	// 最后登录时间转换
	var lastLoginTime *common.DateTime
	if ts := fields["last_login_time"]; ts != "" {
		if sec, err := strconv.ParseInt(ts, 10, 64); err == nil && sec > 0 {
			lastLoginTime = toDateTimePtr(time.Unix(sec, 0))
		}
	}

	return &host.Account{
		AgentID:       ctx.AgentID,
		HostName:      ctx.HostName,
		HostIP:        strings.Join(ctx.HostIP, ","),
		OsType:        strPtr(ctx.OsType),
		Name:          fields["username"],
		Uid:           int32(uid),
		Status:        accStatus, // 账号状态
		Permission:    permission,
		LoginType:     strPtr(fields["shell"]), // shell作为登录方式
		LastLoginTime: lastLoginTime,
		UpdatedAt:     nowDateTime(),
	}
}

// MapProcess 进程字段映射: Agent -> 数据库
// Agent字段: comm, state, path, rusername/eusername, start_time (mapstructure使用小写)
// 数据库字段: name(进程名), status, path, run_name(运行用户), start_time
func MapProcess(fields map[string]string, ctx *AgentContext) *host.Process {
	// 进程名: 使用comm字段 (mapstructure标签为小写)
	processName := fields["comm"]

	// 运行用户: 优先使用eusername(有效用户), 其次rusername(真实用户)
	runUser := fields["eusername"]
	if runUser == "" {
		runUser = fields["rusername"]
	}

	// 进程状态转换
	status := fields["state"]
	switch status {
	case "R":
		status = "运行中"
	case "S":
		status = "睡眠"
	case "D":
		status = "不可中断睡眠"
	case "Z":
		status = "僵尸"
	case "T":
		status = "停止"
	case "t":
		status = "跟踪停止"
	case "X":
		status = "死亡"
	}

	// 启动时间转换
	var startTime *common.DateTime
	if ts := fields["start_time"]; ts != "" {
		if sec, err := strconv.ParseInt(ts, 10, 64); err == nil && sec > 0 {
			startTime = toDateTimePtr(time.Unix(sec, 0))
		}
	}

	return &host.Process{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		OsType:    strPtr(ctx.OsType),
		Name:      processName, // 进程名称
		Status:    strPtr(status),
		Path:      fields["path"], // Agent已直接上报path字段
		RunName:   runUser,        // 运行用户
		StartTime: startTime,
		UpdatedAt: nowDateTime(),
	}
}

// MapDatabase 数据库服务字段映射: Agent -> 数据库
// Agent字段(新): db_type, db_version, port, run_user (与数据库表一致)
func MapDatabase(fields map[string]string, ctx *AgentContext) *host.Database {
	port := 0
	if p := fields["port"]; p != "" {
		port = parsePort(p, "port")
	}

	return &host.Database{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		OsType:    strPtr(ctx.OsType),
		DbType:    fields["db_type"],    // 直接使用
		DbVersion: fields["db_version"], // 直接使用
		Port:      int32(port),
		RunUser:   strPtr(fields["run_user"]),
		UpdatedAt: nowDateTime(),
	}
}

// MapWebService Web服务字段映射: Agent -> 数据库
// Agent字段: app_name, version, server_type, site_domain, path
func MapWebService(fields map[string]string, ctx *AgentContext) *host.Web {
	return &host.Web{
		AgentID:    ctx.AgentID,
		HostName:   ctx.HostName,
		HostIP:     strings.Join(ctx.HostIP, ","),
		OsType:     strPtr(ctx.OsType),
		Name:       fields["app_name"],            // 直接使用
		Version:    fields["version"],             // 直接使用
		ServerType: fields["server_type"],         // 直接使用
		SiteDomain: strPtr(fields["site_domain"]), // 站点域名
		Path:       strPtr(fields["path"]),
		UpdatedAt:  nowDateTime(),
	}
}

// MapSystemService 系统服务字段映射: Agent -> 数据库
// Agent字段: name, status, run_user, working_dir, version, command
// 数据库字段: name, status, run_user, path, version, describe
func MapSystemService(fields map[string]string, ctx *AgentContext) *host.System {
	return &host.System{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		OsType:    strPtr(ctx.OsType),
		Name:      fields["name"],            // 直接使用
		Version:   strPtr(fields["version"]), // 直接使用
		Status:    fields["status"],          // 直接使用
		RunUser:   fields["run_user"],        // 直接使用
		Path:      fields["path"],            // 可执行文件路径
		Describe:  strPtr(fields["command"]), // command -> describe
		UpdatedAt: nowDateTime(),
	}
}

// MapHost 主机字段映射
func MapHost(ctx *AgentContext) *host.Host {
	now := common.DateTime{Time: time.Now()}

	// 处理可选字段，空字符串转为 nil
	var macAddr *string
	if ctx.MacAddr != "" {
		macAddr = &ctx.MacAddr
	}

	var osType *string
	if ctx.OsType != "" {
		osType = &ctx.OsType
	}

	var osVersion *string
	if ctx.OsVersion != "" {
		osVersion = &ctx.OsVersion
	}

	var agentVersion *string
	if ctx.AgentVersion != "" {
		agentVersion = &ctx.AgentVersion
	}

	return &host.Host{
		AgentID:       ctx.AgentID,
		HostName:      ctx.HostName,
		HostIP:        strings.Join(ctx.HostIP, ","),
		MacAddr:       macAddr,
		OsType:        osType,
		OsVersion:     osVersion,
		AgentStatus:   1, // 在线
		AgentVersion:  agentVersion,
		LastHeartbeat: &now,
		UpdatedAt:     now,
		CreatedAt:     now,
	}
}

// MapSoftware 软件字段映射: Agent -> 数据库
// Agent字段: name, sversion, type, source, status, vendor, path
func MapSoftware(fields map[string]string, ctx *AgentContext) *host.Software {
	return &host.Software{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		OsType:    strPtr(ctx.OsType),
		Name:      fields["name"],
		Version:   strPtr(fields["sversion"]), // Agent使用sversion
		Type:      fields["type"],
		Source:    strPtr(fields["source"]),
		Status:    strPtr(fields["status"]),
		Vendor:    strPtr(fields["vendor"]),
		Path:      strPtr(fields["path"]),
		UpdatedAt: nowDateTime(),
	}
}

// MapContainer 容器字段映射: Agent -> 数据库
// Agent字段: id, name, state, image_id, image_name, runtime, pid, create_time
func MapContainer(fields map[string]string, ctx *AgentContext) *container.Container {
	return &container.Container{
		AgentID:     ctx.AgentID,
		HostName:    ctx.HostName,
		HostIP:      strings.Join(ctx.HostIP, ","),
		ContainerID: fields["id"],
		Name:        fields["name"],
		State:       fields["state"],
		ImageID:     strPtr(fields["image_id"]),
		ImageName:   strPtr(fields["image_name"]),
		Runtime:     strPtr(fields["runtime"]),
		Pid:         strPtr(fields["pid"]),
		CreateTime:  strPtr(fields["create_time"]),
		UpdatedAt:   nowDateTime(),
	}
}

// MapEnvSuspicious 可疑环境变量字段映射: Agent -> 数据库
// Agent字段: var_name, var_value, suspicious_reasons, source
func MapEnvSuspicious(fields map[string]string, ctx *AgentContext) *host.EnvSuspicious {
	return &host.EnvSuspicious{
		AgentID:           ctx.AgentID,
		HostName:          ctx.HostName,
		HostIP:            strings.Join(ctx.HostIP, ","),
		VarName:           fields["var_name"],
		VarValue:          strPtr(fields["var_value"]),
		SuspiciousReasons: strPtr(fields["suspicious_reasons"]),
		Source:            strPtr(fields["source"]),
		UpdatedAt:         nowDateTime(),
	}
}

// MapKmod 内核模块字段映射: Agent -> 数据库
// Agent字段: name, size, refcount, used_by, state, addr
func MapKmod(fields map[string]string, ctx *AgentContext) *host.Kmod {
	return &host.Kmod{
		AgentID:   ctx.AgentID,
		HostName:  ctx.HostName,
		HostIP:    strings.Join(ctx.HostIP, ","),
		OsType:    strPtr(ctx.OsType),
		Name:      fields["name"],
		Size:      strPtr(fields["size"]),
		RefCount:  strPtr(fields["refcount"]),
		UsedBy:    strPtr(fields["used_by"]),
		State:     strPtr(fields["state"]),
		Addr:      strPtr(fields["addr"]),
		UpdatedAt: nowDateTime(),
	}
}

// MapImagePackage 镜像软件包字段映射: Agent -> 数据库
// Agent字段: image_id, image_name, container_id, package_name, package_version, package_type, os_version
func MapImagePackage(fields map[string]string, ctx *AgentContext) *container.ImagePackage {
	return &container.ImagePackage{
		AgentID:        ctx.AgentID,
		HostName:       ctx.HostName,
		HostIP:         strings.Join(ctx.HostIP, ","),
		ImageID:        fields["image_id"],
		ImageName:      fields["image_name"],
		PackageName:    fields["package_name"],
		PackageVersion: strPtr(fields["package_version"]),
		PackageType:    fields["package_type"],
		OsVersion:      strPtr(fields["os_version"]),
		UpdatedAt:      nowDateTime(),
	}
}

// MapImage 镜像字段映射: Agent -> 数据库
// Agent字段: image_id, image_name, image_version, image_size, container_count, image_build_time, runtime
func MapImage(fields map[string]string, ctx *AgentContext) *container.Image {
	var containerCount *int32
	if c := fields["container_count"]; c != "" {
		if n, err := strconv.Atoi(c); err == nil {
			containerCount = int32Ptr(int32(n))
		}
	}

	var imageSize *int64
	if s := fields["image_size"]; s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			imageSize = int64Ptr(n)
		}
	}

	var buildTime *common.DateTime
	if bt := fields["image_build_time"]; bt != "" {
		if sec, err := strconv.ParseInt(bt, 10, 64); err == nil && sec > 0 {
			buildTime = toDateTimePtr(time.Unix(sec, 0))
		}
	}

	return &container.Image{
		AgentID:        ctx.AgentID,
		HostName:       ctx.HostName,
		HostIP:         strings.Join(ctx.HostIP, ","),
		ImageID:        fields["image_id"],
		ImageName:      fields["image_name"],
		ImageVersion:   strPtr(fields["image_version"]),
		ImageSize:      imageSize,
		ContainerCount: containerCount,
		BuildTime:      buildTime,
		Runtime:        strPtr(fields["runtime"]),
		UpdatedAt:      nowDateTime(),
	}
}
