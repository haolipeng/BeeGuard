package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"

	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/process"
	"go.uber.org/zap"
)

// WebServiceRule Web服务识别规则
type WebServiceRule struct {
	appName           string                                  // 应用名称: nginx/apache
	serverType        string                                  // 服务器类型
	versionRegex      *regexp.Regexp                          // 版本号正则
	versionArgs       []string                                // 获取版本的命令参数
	versionTrimPrefix string                                  // 版本号前缀清理
	confFunc          func(cmdline string, proc process.Process) string // 配置文件查找函数
}

var (
	nginxWebRule = &WebServiceRule{
		appName:           "nginx",
		serverType:        "nginx",
		versionRegex:      regexp.MustCompile(`nginx\/(\d+\.)+\d+`),
		versionArgs:       []string{"-v"},
		versionTrimPrefix: "nginx/",
		confFunc: func(cmdline string, proc process.Process) string {
			// 从命令行参数中提取 -c 参数
			res := regexp.MustCompile(`-c\s+\S+`).Find([]byte(cmdline))
			if res != nil {
				return strings.TrimSpace(strings.TrimPrefix(string(res), "-c"))
			}
			// 默认配置文件路径
			if _, err := os.Stat("/etc/nginx/nginx.conf"); err == nil {
				return "/etc/nginx/nginx.conf"
			}
			return ""
		},
	}

	apacheWebRule = &WebServiceRule{
		appName:           "apache",
		serverType:        "apache",
		versionRegex:      regexp.MustCompile(`Apache\/(\d+\.)+\d+`),
		versionArgs:       []string{"-v"},
		versionTrimPrefix: "Apache/",
		confFunc: func(cmdline string, proc process.Process) string {
			// 从命令行参数中提取 -f 参数
			res := regexp.MustCompile(`-f\s+\S+`).Find([]byte(cmdline))
			if res != nil {
				return strings.TrimSpace(strings.TrimPrefix(string(res), "-f"))
			}
			// 查找默认配置文件路径
			defaultPaths := []string{
				"/usr/local/apache2/conf/httpd.conf",
				"/etc/apache2/apache2.conf",
				"/etc/httpd/conf/httpd.conf",
				"/etc/apache2/httpd.conf",
			}
			for _, path := range defaultPaths {
				if _, err := os.Stat(path); err == nil {
					return path
				}
			}
			return ""
		},
	}

	webServiceRuleMap = map[string]*WebServiceRule{
		"nginx":   nginxWebRule,
		"apache2": apacheWebRule,
		"httpd":   apacheWebRule,
	}
)

// WebServiceHandler Web服务采集处理器
type WebServiceHandler struct{}

func (h *WebServiceHandler) Name() string {
	return "web_service"
}

func (h *WebServiceHandler) DataType() int {
	return 5060 // Web服务专用DataType
}

func (h *WebServiceHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	procs, err := process.Processes(false)
	if err != nil {
		zap.S().Errorf("Failed to get processes: %v", err)
		return
	}

	versionCache := map[string]string{}
	reported := map[string]bool{} // 防止重复上报同类型服务

	for _, proc := range procs {
		time.Sleep(process.TraversalInterval)

		comm, err := proc.Comm()
		if err != nil {
			continue
		}

		rule, ok := webServiceRuleMap[comm]
		if !ok {
			continue
		}

		// 防止同类型服务重复上报
		if reported[rule.serverType] {
			continue
		}

		stat, err := proc.Stat()
		if err != nil {
			continue
		}
		status, err := proc.Status()
		if err != nil {
			continue
		}

		// 检查是否为子进程（跳过子进程）
		if h.isChildProcess(comm, stat.Ppid) {
			continue
		}

		euid, err := strconv.ParseUint(status.Euid, 10, 64)
		if err != nil {
			continue
		}
		egid, err := strconv.ParseUint(status.Egid, 10, 64)
		if err != nil {
			continue
		}
		exe, err := proc.Exe()
		if err != nil {
			continue
		}
		pns, err := proc.Namespace("pid")
		if err != nil {
			continue
		}
		dir, err := proc.Cwd()
		if err != nil {
			continue
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			continue
		}

		// 获取版本号
		version := versionCache[exe+pns]
		if version == "" {
			version = h.getVersion(rule, uint32(euid), uint32(egid), dir, exe)
			if version != "" {
				versionCache[exe+pns] = version
			}
		}

		// 获取运行用户
		runUser := status.Eusername
		if runUser == "" {
			runUser = status.Rusername
		}

		// 获取配置文件路径
		configPath := ""
		if rule.confFunc != nil {
			configPath = rule.confFunc(cmdline, proc)
		}

		// 上报Web服务资产
		c.SendRecord(&businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: map[string]string{
					"app_name":    rule.appName,
					"version":     version,
					"server_type": rule.serverType,
					"run_user":    runUser,
					"path":        configPath,
					"package_seq": seq,
				},
			},
		})

		reported[rule.serverType] = true
		zap.S().Infof("Web service collected: app=%s version=%s type=%s user=%s path=%s",
			rule.appName, version, rule.serverType, runUser, configPath)
	}
}

// isChildProcess 检查是否为子进程
func (h *WebServiceHandler) isChildProcess(comm, ppid string) bool {
	p, err := process.NewProcess(ppid)
	if err != nil {
		return false
	}
	ppidComm, err := p.Comm()
	if err != nil {
		return false
	}
	return ppidComm == comm
}

// getVersion 获取Web服务版本
func (h *WebServiceHandler) getVersion(rule *WebServiceRule, uid, gid uint32, dir, exe string) string {
	if rule.versionRegex == nil || len(rule.versionArgs) == 0 {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, rule.versionArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		zap.S().Debugf("Failed to get version for %s: %v", rule.appName, err)
		return ""
	}

	res := rule.versionRegex.Find(output)
	if res == nil {
		return ""
	}

	version := string(res)
	if rule.versionTrimPrefix != "" {
		version = strings.TrimPrefix(version, rule.versionTrimPrefix)
	}
	return version
}

// getWebServerRoot 获取Web服务根目录（保留供扩展使用）
func getWebServerRoot(configPath string) string {
	if configPath == "" {
		return ""
	}
	return filepath.Dir(configPath)
}
