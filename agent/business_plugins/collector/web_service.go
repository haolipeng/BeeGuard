package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

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

	// site_domain 解析相关正则
	nginxServerNameRegex  = regexp.MustCompile(`(?i)^\s*server_name\s+(.+?)\s*;`)
	nginxIncludeRegex     = regexp.MustCompile(`(?i)^\s*include\s+(.+?)\s*;`)
	apacheServerNameRegex = regexp.MustCompile(`(?i)^\s*ServerName\s+(\S+)`)
	apacheServerAliasRegex = regexp.MustCompile(`(?i)^\s*ServerAlias\s+(.+)`)
	apacheIncludeRegex    = regexp.MustCompile(`(?i)^\s*(?:Include|IncludeOptional)\s+(.+?)\s*$`)
)

// WebServiceHandler Web服务采集处理器
type WebServiceHandler struct{}

func (h *WebServiceHandler) Name() string {
	return "web_service"
}

func (h *WebServiceHandler) DataType() int {
	return datatype.WebService
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

		// 提取站点域名
		siteDomain := h.extractSiteDomains(configPath, rule.serverType)

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
					"site_domain": siteDomain,
					"package_seq": seq,
				},
			},
		})

		reported[rule.serverType] = true
		zap.S().Infof("Web service collected: app=%s version=%s type=%s user=%s path=%s domains=%s",
			rule.appName, version, rule.serverType, runUser, configPath, siteDomain)
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

// extractSiteDomains 提取Web服务配置中的站点域名
func (h *WebServiceHandler) extractSiteDomains(configPath, serverType string) string {
	if configPath == "" {
		return ""
	}

	var domains []string
	switch serverType {
	case "nginx":
		domains = parseNginxDomains(configPath)
	case "apache":
		domains = parseApacheDomains(configPath)
	default:
		return ""
	}

	// 去重
	seen := make(map[string]bool, len(domains))
	unique := make([]string, 0, len(domains))
	for _, d := range domains {
		lower := strings.ToLower(d)
		if !seen[lower] {
			seen[lower] = true
			unique = append(unique, d)
		}
	}

	result := strings.Join(unique, ",")
	// VARCHAR(255) 截断：在最后一个逗号处截断，避免切断域名
	if len(result) > 255 {
		result = result[:255]
		if idx := strings.LastIndex(result, ","); idx > 0 {
			result = result[:idx]
		}
	}
	return result
}

// parseNginxDomains 解析 nginx 配置中的域名（主配置 + 一级 include）
func parseNginxDomains(configPath string) []string {
	domains, includes := parseNginxConfigFile(configPath)

	configDir := filepath.Dir(configPath)
	for _, inc := range includes {
		if !filepath.IsAbs(inc) {
			inc = filepath.Join(configDir, inc)
		}
		// 展开 glob 模式
		matches, err := filepath.Glob(inc)
		if err != nil {
			zap.S().Debugf("Failed to glob nginx include %s: %v", inc, err)
			continue
		}
		for _, m := range matches {
			d, _ := parseNginxConfigFile(m)
			domains = append(domains, d...)
		}
	}
	return domains
}

// parseNginxConfigFile 逐行扫描单个 nginx 配置文件，返回域名和 include 路径
func parseNginxConfigFile(path string) (domains []string, includes []string) {
	f, err := os.Open(path)
	if err != nil {
		zap.S().Debugf("Failed to open nginx config %s: %v", path, err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(io.LimitReader(f, 10*1024*1024))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过注释
		if strings.HasPrefix(line, "#") {
			continue
		}

		if m := nginxServerNameRegex.FindStringSubmatch(line); m != nil {
			names := strings.Fields(m[1])
			for _, name := range names {
				if isValidDomain(name) {
					domains = append(domains, name)
				}
			}
		}

		if m := nginxIncludeRegex.FindStringSubmatch(line); m != nil {
			includes = append(includes, m[1])
		}
	}
	return
}

// parseApacheDomains 解析 apache 配置中的域名（主配置 + 一级 Include）
func parseApacheDomains(configPath string) []string {
	domains, includes := parseApacheConfigFile(configPath)

	configDir := filepath.Dir(configPath)
	for _, inc := range includes {
		if !filepath.IsAbs(inc) {
			inc = filepath.Join(configDir, inc)
		}
		matches, err := filepath.Glob(inc)
		if err != nil {
			zap.S().Debugf("Failed to glob apache include %s: %v", inc, err)
			continue
		}
		for _, m := range matches {
			d, _ := parseApacheConfigFile(m)
			domains = append(domains, d...)
		}
	}
	return domains
}

// parseApacheConfigFile 逐行扫描单个 apache 配置文件，返回域名和 Include 路径
func parseApacheConfigFile(path string) (domains []string, includes []string) {
	f, err := os.Open(path)
	if err != nil {
		zap.S().Debugf("Failed to open apache config %s: %v", path, err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(io.LimitReader(f, 10*1024*1024))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		if m := apacheServerNameRegex.FindStringSubmatch(line); m != nil {
			name := strings.TrimSpace(m[1])
			if isValidDomain(name) {
				domains = append(domains, name)
			}
		}

		if m := apacheServerAliasRegex.FindStringSubmatch(line); m != nil {
			aliases := strings.Fields(m[1])
			for _, alias := range aliases {
				if isValidDomain(alias) {
					domains = append(domains, alias)
				}
			}
		}

		if m := apacheIncludeRegex.FindStringSubmatch(line); m != nil {
			includes = append(includes, m[1])
		}
	}
	return
}

// isValidDomain 过滤无意义的域名值
func isValidDomain(name string) bool {
	if name == "" || name == "_" || name == "*" || strings.EqualFold(name, "localhost") {
		return false
	}
	return true
}

// getWebServerRoot 获取Web服务根目录（保留供扩展使用）
func getWebServerRoot(configPath string) string {
	if configPath == "" {
		return ""
	}
	return filepath.Dir(configPath)
}
