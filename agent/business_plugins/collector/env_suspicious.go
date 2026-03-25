package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

	"github.com/go-viper/mapstructure/v2"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/engine"
	"go.uber.org/zap"
)

// EnvSuspiciousHandler 可疑环境变量检测处理器
type EnvSuspiciousHandler struct{}

func (h *EnvSuspiciousHandler) Name() string {
	return "env_suspicious"
}

func (h *EnvSuspiciousHandler) DataType() int {
	return datatype.EnvSuspicious
}

// 可疑环境变量的特征模式
var suspiciousPatterns = []struct {
	name        string
	pattern     *regexp.Regexp
	description string
}{
	{
		name:        "Base64编码",
		pattern:     regexp.MustCompile(`^[A-Za-z0-9+/]{20,}={0,2}$`),
		description: "可能包含Base64编码的敏感信息",
	},
	{
		name:        "可疑路径",
		pattern:     regexp.MustCompile(`(/tmp|/var/tmp|/dev/shm|/proc/self|\.\./).*`),
		description: "指向临时目录或可疑路径",
	},
	{
		name:        "可疑URL",
		pattern:     regexp.MustCompile(`(http|https|ftp|tcp)://[^\s]+`),
		description: "包含网络地址，可能是C2服务器",
	},
	{
		name:        "可疑命令",
		pattern:     regexp.MustCompile(`(bash|sh|python|perl|nc|netcat|wget|curl|socat).*`),
		description: "包含可执行命令",
	},
	{
		name:        "敏感关键词",
		pattern:     regexp.MustCompile(`(?i)(password|passwd|secret|token|key|api|credential|auth)`),
		description: "包含敏感关键词",
	},
}

// 可疑的环境变量名
// 这些环境变量可能被攻击者利用进行权限提升、代码注入、数据泄露等攻击
var suspiciousVarNames = []string{
	"LD_PRELOAD",      // 动态链接库预加载，可被用于劫持系统调用
	"LD_LIBRARY_PATH", // 动态链接库搜索路径，可被用于加载恶意库文件
	"PROMPT_COMMAND",  // Bash 提示符命令，每次显示提示符时都会执行
	"PS1",             // Bash 提示符变量，可能被用于隐藏恶意命令
	"PATH",            // 可执行文件搜索路径，可被用于路径劫持攻击
	"HISTFILE",        // 命令历史文件路径，可被用于隐藏命令执行痕迹
	"HISTCONTROL",     // 命令历史控制，可被用于隐藏敏感命令
	"HTTP_PROXY",      // 代理服务器设置，可被用于中间人攻击或数据泄露
	"HTTPS_PROXY",     // 代理服务器设置，可被用于中间人攻击或数据泄露
	"NO_PROXY",        // 代理排除列表，可能被用于绕过安全检测
	"TMPDIR",          // 临时目录路径，可被用于放置恶意文件或隐藏攻击载荷
	"TMP",             // 临时目录路径
	"TEMP",            // 临时目录路径
}

// EnvSuspicious 可疑环境变量信息结构体
type EnvSuspicious struct {
	VarName           string `mapstructure:"var_name"`           // 环境变量名
	VarValue          string `mapstructure:"var_value"`          // 环境变量值
	SuspiciousReasons string `mapstructure:"suspicious_reasons"` // 可疑原因（多个原因用分号分隔）
	Source            string `mapstructure:"source"`             // 来源（如 /etc/environment, /etc/profile 等）
}

// getSystemEnvs 获取系统环境变量
// 从多个地方获取：/etc/environment、/etc/profile、/etc/profile.d/*.sh等
// 整理环境变量以映射表作为返回值
func getSystemEnvs() map[string]string {
	envs := make(map[string]string)

	// 1. 读取 /etc/environment（系统级环境变量，Debian/Ubuntu）
	if f, err := os.Open("/etc/environment"); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// 跳过注释和空行
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				// 移除引号
				value := strings.Trim(parts[1], `"'`)
				envs[parts[0]] = value
			}
		}
		f.Close()
	}

	// 2. 读取 /etc/profile（系统级配置）
	if f, err := os.Open("/etc/profile"); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// 查找 export VAR=value 格式
			if strings.HasPrefix(line, "export ") {
				line = strings.TrimPrefix(line, "export ")
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					value := strings.Trim(parts[1], `"'`)
					envs[parts[0]] = value
				}
			}
		}
		f.Close()
	}

	// 3. 读取 /etc/profile.d/*.sh（系统级配置脚本）
	if dir, err := os.Open("/etc/profile.d"); err == nil {
		files, _ := dir.Readdirnames(0)
		dir.Close()
		for _, file := range files {
			if strings.HasSuffix(file, ".sh") {
				filePath := fmt.Sprintf("/etc/profile.d/%s", file)
				if f, err := os.Open(filePath); err == nil {
					scanner := bufio.NewScanner(f)
					for scanner.Scan() {
						line := strings.TrimSpace(scanner.Text())
						if strings.HasPrefix(line, "export ") {
							line = strings.TrimPrefix(line, "export ")
							parts := strings.SplitN(line, "=", 2)
							if len(parts) == 2 {
								value := strings.Trim(parts[1], `"'`)
								envs[parts[0]] = value
							}
						}
					}
					f.Close()
				}
			}
		}
	}

	return envs
}

// checkSuspiciousEnv 检查单个环境变量是否可疑
func checkSuspiciousEnv(key, value string) []string {
	var suspiciousReasons []string

	// 检查变量名是否可疑（使用精确匹配）
	for _, suspiciousName := range suspiciousVarNames {
		if key == suspiciousName {
			suspiciousReasons = append(suspiciousReasons, fmt.Sprintf("可疑变量名: %s", suspiciousName))
			break
		}
	}

	// 检查变量值是否匹配可疑模式
	for _, pattern := range suspiciousPatterns {
		if pattern.pattern.MatchString(value) {
			suspiciousReasons = append(suspiciousReasons, fmt.Sprintf("%s: %s", pattern.name, pattern.description))
		}
	}

	// 检查值是否异常长（可能包含编码数据）
	if len(value) > 500 {
		suspiciousReasons = append(suspiciousReasons, "值异常长，可能包含编码数据")
	}

	// 检查是否包含特殊字符
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		suspiciousReasons = append(suspiciousReasons, "包含换行符，可能是命令注入")
	}

	return suspiciousReasons
}

// analyzeEnvs 分析环境变量并返回可疑项
func analyzeEnvs(envs map[string]string) []*EnvSuspicious {
	var suspicious []*EnvSuspicious

	for key, value := range envs {
		reasons := checkSuspiciousEnv(key, value)
		if len(reasons) > 0 {
			suspicious = append(suspicious, &EnvSuspicious{
				VarName:           key,
				VarValue:          value,
				SuspiciousReasons: strings.Join(reasons, "; "),
				Source:            "system", // 可以进一步细化来源
			})
		}
	}

	return suspicious
}

func (h *EnvSuspiciousHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	// 获取系统环境变量
	envs := getSystemEnvs()
	if len(envs) == 0 {
		zap.S().Warn("No system environment variables found")
		return
	}

	zap.S().Infof("Collected %d system environment variables", len(envs))

	// 分析可疑环境变量
	suspicious := analyzeEnvs(envs)
	if len(suspicious) == 0 {
		zap.S().Info("No suspicious environment variables found")
		// 即使没有可疑项，也发送一条记录表示检测完成
		rec := &businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: map[string]string{
					"total_envs":       fmt.Sprintf("%d", len(envs)),
					"suspicious_count": "0",
					"package_seq":      seq,
				},
			},
		}
		c.SendRecord(rec)
		return
	}

	zap.S().Infof("Found %d suspicious environment variables", len(suspicious))

	// 发送每个可疑环境变量记录
	for _, env := range suspicious {
		rec := &businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: make(map[string]string, 5),
			},
		}

		// 使用 mapstructure 将结构体转换为 map
		err := mapstructure.Decode(env, &rec.Data.Fields)
		if err != nil {
			zap.S().Warnf("Failed to decode suspicious env: %v", err)
			continue
		}

		// 添加包序列号
		rec.Data.Fields["package_seq"] = seq
		rec.Data.Fields["total_envs"] = fmt.Sprintf("%d", len(envs))
		rec.Data.Fields["suspicious_count"] = fmt.Sprintf("%d", len(suspicious))

		c.SendRecord(rec)
	}

	zap.S().Infof("Environment suspicious detection completed, sent %d records", len(suspicious))
}
