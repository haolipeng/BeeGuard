package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-viper/mapstructure/v2"
	"github.com/karrick/godirwalk"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/engine"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/utils"
	"go.uber.org/zap"
)

// SearchDir systemd 服务文件搜索目录列表
// 按照优先级顺序搜索，从高优先级到低优先级
var SearchDir = []string{
	"/etc/systemd/system.control",
	"/run/systemd/system.control",
	"/run/systemd/transient",
	"/run/systemd/generator.early",
	"/etc/systemd/system",
	"/run/systemd/system",
	"/run/systemd/generator",
	"/usr/local/lib/systemd/system",
	"/usr/lib/systemd/system",
	"/run/systemd/generator.late",
}

// ServiceHandler 系统服务采集处理器
type ServiceHandler struct{}

func (h *ServiceHandler) Name() string {
	return "service"
}

func (h *ServiceHandler) DataType() int {
	return datatype.Service
}

// Service 服务信息结构体
type Service struct {
	Name       string `mapstructure:"name"`        // 服务名称（文件名）
	Type       string `mapstructure:"type"`        // 服务类型（simple, oneshot, dbus 等）
	Command    string `mapstructure:"command"`     // 启动命令（ExecStart）
	Path       string `mapstructure:"path"`        // 可执行文件路径（从 ExecStart 提取）
	Restart    string `mapstructure:"restart"`     // 是否自动重启（true/false）
	WorkingDir string `mapstructure:"working_dir"` // 工作目录（WorkingDirectory）
	Checksum   string `mapstructure:"checksum"`    // 文件 MD5 校验和
	BusName    string `mapstructure:"bus_name"`    // D-Bus 总线名称（如果适用）
	Status     string `mapstructure:"status"`      // 服务运行状态（active/inactive/failed）
	RunUser    string `mapstructure:"run_user"`    // 运行用户
	Version    string `mapstructure:"version"`     // 服务版本（从二进制获取）
}

// SetDefault 设置默认服务类型
// 根据服务配置推断服务类型
func (s *Service) SetDefault() {
	if s.Command != "" && s.Type == "" && s.BusName == "" {
		// 有启动命令但没有指定类型，默认为 simple
		s.Type = "simple"
	} else if s.Command == "" && s.Type == "" {
		// 没有启动命令，默认为 oneshot（一次性服务）
		s.Type = "oneshot"
	} else if s.Type == "" && s.BusName != "" {
		// 有 D-Bus 名称，默认为 dbus 类型
		s.Type = "dbus"
	}
}

// getServiceRuntimeInfo 获取服务运行时信息（状态、用户、版本）
// 通过 systemctl show 命令获取；若未解析到状态则用 systemctl is-active 回退
func getServiceRuntimeInfo(serviceName string) (status, runUser, version string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/usr/bin/systemctl", "show", serviceName,
		"--property=ActiveState,User,ExecMainPID,Version")
	output, err := cmd.Output()
	if err != nil {
		zap.S().Debugf("systemctl show %s failed: %v", serviceName, err)
		return getServiceStatusFallback(serviceName), "", ""
	}

	lines := strings.Split(string(output), "\n")
	var activeState string
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "ActiveState":
			if value != "" && value != "[not set]" {
				activeState = value
			}
		case "User":
			if value != "" && value != "[not set]" {
				runUser = value
			}
		case "Version":
			if value != "" && value != "[not set]" {
				version = value
			}
		}
	}

	// 只上报 active/inactive/failed，不组合 SubState
	if activeState != "" {
		status = activeState
	} else {
		status = getServiceStatusFallback(serviceName)
	}

	return status, runUser, version
}

// getServiceStatusFallback 在 systemctl show 未返回有效状态时，用 systemctl is-active 获取单一状态
func getServiceStatusFallback(serviceName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/systemctl", "is-active", serviceName)
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "unknown"
	}
	return s
}

// versionExtractPattern 从 --version 输出中提取 "程序名 + 版本号" 的核心部分
// 支持以下常见格式：
//
//	"openjdk version \"17.0.6\" 2023-01-17 LTS"          → openjdk version "17.0.6"
//	"Docker version 24.0.7, build afdd53b4e3"             → Docker version 24.0.7
//	"curl 7.81.0 (x86_64-pc-linux-gnu) libcurl/7.81.0"   → curl 7.81.0
//	"Python 3.11.2 (main, Mar 13 2023, 12:18:29)"         → Python 3.11.2
//	"OpenSSH_8.9p1 Ubuntu-3ubuntu0.6, OpenSSL 3.0.2"      → OpenSSH_8.9p1（回退到首词）
var versionExtractPattern = regexp.MustCompile(
	`(?i)^(` +
		`\S+\s+version\s+"[\d][\w._-]+"` + // name version "x.y.z"
		`|\S+\s+version\s+[\d][\w._+-]+` + // name version x.y.z
		`|\S+\s+[\d][\w._+-]+` + // name x.y.z
		`)`,
)

// extractCleanVersion 从 --version 输出的首行中提取简洁版本信息
func extractCleanVersion(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	if match := versionExtractPattern.FindString(line); match != "" {
		return match
	}

	// 回退：取第一个词（处理 OpenSSH_8.9p1 等名称内嵌版本号的情况）
	if fields := strings.Fields(line); len(fields) > 0 {
		return fields[0]
	}
	return line
}

// getVersionFromBinary 尝试从二进制文件获取版本号
func getVersionFromBinary(command string) string {
	if command == "" {
		return ""
	}

	// 提取可执行文件路径（ExecStart 可能包含参数）
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}
	exe := parts[0]

	// 移除可能的前缀（如 -、@、+、!、!!）
	exe = strings.TrimLeft(exe, "-@+!")

	// 检查文件是否存在
	if _, err := os.Stat(exe); err != nil {
		return ""
	}

	// 尝试 --version 参数
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, "--version")
	output, err := cmd.Output()
	if err != nil {
		// 尝试 -v 参数
		ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel2()
		cmd = exec.CommandContext(ctx2, exe, "-v")
		output, err = cmd.Output()
		if err != nil {
			return ""
		}
	}

	// 提取第一行，使用正则提取简洁版本信息
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return extractCleanVersion(lines[0])
	}
	return ""
}

// extractExePath 从 ExecStart 命令中提取可执行文件路径
// ExecStart 格式示例：
//
//	"/usr/sbin/sshd -D $SSHD_OPTS"          → /usr/sbin/sshd
//	"-/sbin/agetty -o -p -- \\u --noclear"   → /sbin/agetty  （去除 - 前缀）
//	"@/usr/bin/dbus-daemon --system"          → /usr/bin/dbus-daemon
//	""                                        → ""
func extractExePath(command string) string {
	if command == "" {
		return ""
	}
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}
	exe := strings.TrimLeft(parts[0], "-@+!")
	if exe == "" || !strings.HasPrefix(exe, "/") {
		return ""
	}
	return exe
}

func (h *ServiceHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	// 使用 Set 来去重，避免重复处理同名服务文件
	set := mapset.NewSet[string]()

	// 遍历所有搜索目录
	for _, dir := range SearchDir {
		// 使用 godirwalk 高效遍历目录
		if err := godirwalk.Walk(dir, &godirwalk.Options{
			Callback: func(path string, de *godirwalk.Dirent) error {
				// 只处理 .service 文件（普通文件或符号链接）
				if strings.HasSuffix(de.Name(), ".service") && (de.IsRegular() || de.IsSymlink()) && !set.Contains(de.Name()) {
					// 打开服务文件
					f, err := os.Open(path)
					if err != nil {
						// 如果打开失败，跳过该文件
						return nil
					}
					defer f.Close()

					// 标记该文件名已处理
					set.Add(de.Name())

					// 限制文件读取大小（最大 1MB）
					s := bufio.NewScanner(io.LimitReader(f, 1024*1024))

					// 初始化服务信息
					u := &Service{
						Name:    de.Name(),
						Restart: "false", // 默认不自动重启
					}

					// 解析服务文件内容（使用 SplitN(_, "=", 2) 保证值中可含等号，如 ExecStart=... --containerd=/path）
					for s.Scan() {
						parts := strings.SplitN(s.Text(), "=", 2)
						if len(parts) != 2 {
							continue
						}

						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])

						// 解析关键字段
						switch key {
						case "Type":
							// 服务类型
							u.Type = value
						case "ExecStart":
							// 启动命令
							u.Command = value
						case "Restart":
							// 重启策略
							if value == "no" {
								u.Restart = "false"
							} else {
								u.Restart = "true"
							}
						case "WorkingDirectory":
							// 工作目录
							u.WorkingDir = value
						case "User":
							// 运行用户
							u.RunUser = value
						}
					}

					// 计算文件 MD5 校验和
					u.Checksum, _ = utils.GetMd5(path, "")

					// 从 ExecStart 提取可执行文件路径
					u.Path = extractExePath(u.Command)

					// 设置默认服务类型
					u.SetDefault()

					// 获取服务运行时信息（状态、用户）
					status, runUser, version := getServiceRuntimeInfo(u.Name)
					u.Status = status
					// 如果服务文件中没有指定用户，使用运行时用户
					if u.RunUser == "" {
						u.RunUser = runUser
					}
					// 如果没有从systemctl获取到版本，尝试从二进制获取
					if version != "" {
						u.Version = version
					} else {
						u.Version = getVersionFromBinary(u.Command)
					}

					// 如果RunUser仍为空，默认为root
					if u.RunUser == "" {
						u.RunUser = "root"
					}

					// 创建记录
					rec := &businessplugins.Record{
						DataType:  int32(h.DataType()),
						Timestamp: time.Now().Unix(),
						Data: &businessplugins.Payload{
							Fields: make(map[string]string, 7),
						},
					}

					// 使用 mapstructure 将 Service 结构体转换为 map[string]string
					if err := mapstructure.Decode(u, &rec.Data.Fields); err != nil {
						zap.S().Warnf("Failed to decode service: %v", err)
						return nil
					}

					// 添加包序列号
					rec.Data.Fields["package_seq"] = seq

					// 发送记录到 agent
					c.SendRecord(rec)
				}
				return nil
			},
			FollowSymbolicLinks: false, // 不跟随符号链接，避免循环
		}); err != nil {
			// 如果目录不存在或无法访问，记录日志但继续处理下一个目录
			zap.S().Debugf("Failed to walk directory %s: %v", dir, err)
		}
	}

	zap.S().Infof("Service collection completed, processed %d unique services", set.Cardinality())
}
