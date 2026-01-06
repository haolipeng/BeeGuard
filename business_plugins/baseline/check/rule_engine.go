package check

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type RuleStruct struct {
	Type    string      `yaml:"type" bson:"type"`
	Param   []string    `yaml:"param" bson:"param"`
	Filter  string      `yaml:"filter" bson:"filter"`
	Require string      `yaml:"require" bson:"require"`
	Result  interface{} `yaml:"result" bson:"result"`
}

type BaselineCheck struct {
	Condition string       `yaml:"condition" bson:"condition"`
	Rules     []RuleStruct `yaml:"rules" bson:"rules"`
}

type CheckInfo struct {
	CheckId       int           `yaml:"check_id" bson:"check_id"`
	Type          string        `yaml:"type" bson:"type"`
	Title         string        `yaml:"title" bson:"title"`
	TitleCn       string        `yaml:"title_cn" bson:"title_cn"`
	Description   string        `yaml:"description" bson:"description"`
	Solution      string        `yaml:"solution" bson:"solution"`
	Security      string        `yaml:"security" bson:"security"`
	TypeCn        string        `yaml:"type_cn" bson:"type_cn"`
	DescriptionCn string        `yaml:"description_cn" bson:"description_cn"`
	SolutionCn    string        `yaml:"solution_cn" bson:"solution_cn"`
	Check         BaselineCheck `yaml:"check" bson:"check"`
}

type BaselineInfo struct {
	BaselineId      int         `yaml:"baseline_id" bson:"baseline_id"`
	BaselineVersion string      `yaml:"baseline_version" bson:"baseline_version"`
	CheckList       []CheckInfo `yaml:"check_list" bson:"check_list"`
}

const (
	SuccessCode      = 1
	FailCode         = 2
	ErrorCode        = -1
	ErrorConfigWrite = -2 // Configuration writing is not standardized
	ErrorFile        = -3 // File read and write exception
)

// AnalysisRule Rule parsing engine
// 实现三种检测方式：
// 1. 检测 root 启动的业务进程比例
// 2. 检测非系统进程是否由 root 启动
// 3. 检测不安全服务是否已禁用
func AnalysisRule(check BaselineCheck) (ifPass bool, err error) {
	// 根据规则类型进行检测
	for _, rule := range check.Rules {
		switch rule.Type {
		case "root_process_ratio":
			// 检测 1: root 启动的业务进程比例
			ratio, err := checkRootProcessRatio()
			if err != nil {
				return false, err
			}
			// 从配置文件 rule.Result 读取阈值
			if rule.Result == nil {
				return false, fmt.Errorf("threshold not configured in rule.Result for root_process_ratio")
			}

			var threshold float64
			var parseErr error
			if thresholdStr, ok := rule.Result.(string); ok {
				threshold, parseErr = strconv.ParseFloat(thresholdStr, 64)
			} else if thresholdNum, ok := rule.Result.(float64); ok {
				threshold = thresholdNum
			} else if thresholdInt, ok := rule.Result.(int); ok {
				threshold = float64(thresholdInt)
			} else {
				return false, fmt.Errorf("invalid threshold type in rule.Result: %T, expected string, float64, or int", rule.Result)
			}

			if parseErr != nil {
				return false, fmt.Errorf("failed to parse threshold: %v", parseErr)
			}

			if ratio > threshold {
				return false, fmt.Errorf("root process ratio %.2f%% exceeds threshold %.2f%%", ratio, threshold)
			}

		case "non_system_root_process":
			// 检测 2: 非系统进程是否由 root 启动
			hasNonSystemRoot, err := checkNonSystemRootProcess()
			if err != nil {
				return false, err
			}
			if hasNonSystemRoot {
				return false, fmt.Errorf("found non-system processes running as root")
			}

		case "insecure_services":
			// 检测 3: 不安全服务是否已禁用
			services := []string{"telnet", "rsh", "ftp", "tftp", "smb"}
			if len(rule.Param) > 0 && rule.Param[0] != "" {
				services = strings.Split(rule.Param[0], ",")
				for i := range services {
					services[i] = strings.TrimSpace(services[i])
				}
			}
			enabled, err := checkInsecureServices(services)
			if err != nil {
				return false, err
			}
			if enabled {
				return false, fmt.Errorf("insecure services are enabled")
			}

		default:
			// 其他类型规则，默认通过
			return true, nil
		}
	}

	return true, nil
}

// checkRootProcessRatio 检测 root 启动的业务进程比例
func checkRootProcessRatio() (ratio float64, err error) {
	// 获取所有进程信息
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get process list: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	totalProcesses := 0
	rootProcesses := 0

	// 系统进程关键字
	systemKeywords := []string{
		"kernel", "systemd", "init", "kthreadd", "ksoftirqd",
		"migration", "rcu_", "watchdog", "kworker", "kswapd",
		"khugepaged", "netns", "cgroup", "devtmpfs",
	}

	for i, line := range lines {
		if i == 0 {
			continue // 跳过表头
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		user := fields[0]
		command := strings.Join(fields[10:], " ")

		// 判断是否为系统进程
		isSystemProcess := false
		for _, keyword := range systemKeywords {
			if strings.Contains(command, keyword) {
				isSystemProcess = true
				break
			}
		}

		// 只统计业务进程
		if !isSystemProcess {
			totalProcesses++
			if user == "root" {
				rootProcesses++
			}
		}
	}

	if totalProcesses == 0 {
		return 0, nil
	}

	ratio = float64(rootProcesses) * 100.0 / float64(totalProcesses)
	return ratio, nil
}

// checkNonSystemRootProcess 检测非系统进程是否由 root 启动
func checkNonSystemRootProcess() (hasNonSystemRoot bool, err error) {
	// 获取所有进程信息
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get process list: %v", err)
	}

	lines := strings.Split(string(output), "\n")

	// 系统进程关键字
	systemKeywords := []string{
		"kernel", "systemd", "init", "kthreadd", "ksoftirqd",
		"migration", "rcu_", "watchdog", "kworker", "kswapd",
		"khugepaged", "netns", "cgroup", "devtmpfs",
		"sshd", "rsyslog", "cron", "dbus", "NetworkManager",
	}

	for i, line := range lines {
		if i == 0 {
			continue // 跳过表头
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		user := fields[0]
		command := strings.Join(fields[10:], " ")

		// 判断是否为系统进程
		isSystemProcess := false
		for _, keyword := range systemKeywords {
			if strings.Contains(command, keyword) {
				isSystemProcess = true
				break
			}
		}

		// 如果非系统进程且以 root 运行，返回 true
		if !isSystemProcess && user == "root" {
			return true, nil
		}
	}

	return false, nil
}

// checkInsecureServices 检测不安全服务是否已禁用
func checkInsecureServices(services []string) (enabled bool, err error) {
	enabledServices := []string{}

	// 检查 systemd 服务
	for _, service := range services {
		// 检查常见的服务名称变体
		serviceNames := []string{
			service,
			service + ".service",
			service + "d",
			service + "d.service",
		}

		for _, serviceName := range serviceNames {
			// 使用 systemctl 检查服务状态
			cmd := exec.Command("systemctl", "is-enabled", serviceName)
			output, err := cmd.Output()
			if err != nil {
				// 服务不存在或无法检查，跳过
				continue
			}

			status := strings.TrimSpace(string(output))
			if status == "enabled" || status == "enabled-runtime" {
				enabledServices = append(enabledServices, serviceName)
			}
		}

		// 也检查 inetd/xinetd 服务
		if checkInetdService(service) {
			enabledServices = append(enabledServices, service+" (inetd/xinetd)")
		}
	}

	// 如果找到启用的服务，返回 true
	return len(enabledServices) > 0, nil
}

// checkInetdService 检查 inetd/xinetd 配置中的服务
func checkInetdService(service string) bool {
	// 检查 /etc/inetd.conf
	if checkInetdConfig("/etc/inetd.conf", service) {
		return true
	}

	// 检查 /etc/xinetd.d/ 目录下的配置文件
	dir := "/etc/xinetd.d/"
	files, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filePath := dir + file.Name()
		if checkInetdConfig(filePath, service) {
			return true
		}
	}

	return false
}

// checkInetdConfig 检查 inetd 配置文件
func checkInetdConfig(filePath string, service string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 检查是否包含服务名称且未注释
		if strings.Contains(line, service) && !strings.HasPrefix(line, "#") {
			return true
		}
	}
	return false
}
