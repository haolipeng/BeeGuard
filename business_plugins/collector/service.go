package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	businessplugins "business_plugins/lib"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-viper/mapstructure/v2"
	"github.com/karrick/godirwalk"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/utils"
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
	return 5054 // 系统服务采集的数据类型
}

// Service 服务信息结构体
type Service struct {
	Name       string `mapstructure:"name"`        // 服务名称（文件名）
	Type       string `mapstructure:"type"`        // 服务类型（simple, oneshot, dbus 等）
	Command    string `mapstructure:"command"`     // 启动命令（ExecStart）
	Restart    string `mapstructure:"restart"`     // 是否自动重启（true/false）
	WorkingDir string `mapstructure:"working_dir"` // 工作目录（WorkingDirectory）
	Checksum   string `mapstructure:"checksum"`    // 文件 MD5 校验和
	BusName    string `mapstructure:"bus_name"`    // D-Bus 总线名称（如果适用）
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

					// 解析服务文件内容
					for s.Scan() {
						// 按等号分割键值对
						fields := strings.Split(s.Text(), "=")
						if len(fields) != 2 {
							continue
						}

						key := strings.TrimSpace(fields[0])
						value := strings.TrimSpace(fields[1])

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
						}
					}

					// 计算文件 MD5 校验和
					u.Checksum, _ = utils.GetMd5(path, "")

					// 设置默认服务类型
					u.SetDefault()

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
