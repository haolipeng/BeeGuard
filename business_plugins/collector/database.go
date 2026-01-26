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
	"gitlab.myinterest.top/security/agent/business_plugins/collector/port"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/process"
	"go.uber.org/zap"
)

// DatabaseRule 数据库识别规则
type DatabaseRule struct {
	dbType            string         // 数据库类型: mysql/postgresql/redis/mongodb
	defaultPort       int            // 默认端口
	versionRegex      *regexp.Regexp // 版本号正则
	versionArgs       []string       // 获取版本的命令参数
	versionTrimPrefix string         // 版本号前缀清理
	versionTrimSuffix string         // 版本号后缀清理
}

var (
	mysqlDbRule = &DatabaseRule{
		dbType:            "mysql",
		defaultPort:       3306,
		versionRegex:      regexp.MustCompile(`Ver\s(\d+\.)+\d+\S*`),
		versionArgs:       []string{"-V"},
		versionTrimPrefix: "Ver ",
	}
	postgresqlDbRule = &DatabaseRule{
		dbType:       "postgresql",
		defaultPort:  5432,
		versionRegex: regexp.MustCompile(`(\d+\.)+\d+`),
		versionArgs:  []string{"-V"},
	}
	redisDbRule = &DatabaseRule{
		dbType:            "redis",
		defaultPort:       6379,
		versionRegex:      regexp.MustCompile(`v=(\d+\.)+\d+`),
		versionArgs:       []string{"-v"},
		versionTrimPrefix: "v=",
	}
	mongodbDbRule = &DatabaseRule{
		dbType:            "mongodb",
		defaultPort:       27017,
		versionRegex:      regexp.MustCompile(`db\sversion\sv(\d+\.)+\d+`),
		versionArgs:       []string{"--version"},
		versionTrimPrefix: "db version v",
	}

	databaseRuleMap = map[string]*DatabaseRule{
		"mysqld":       mysqlDbRule,
		"postgres":     postgresqlDbRule,
		"redis-server": redisDbRule,
		"mongod":       mongodbDbRule,
	}
)

// DatabaseHandler 数据库服务采集处理器
type DatabaseHandler struct{}

func (h *DatabaseHandler) Name() string {
	return "database"
}

func (h *DatabaseHandler) DataType() int {
	return 5061 // 数据库服务专用DataType
}

func (h *DatabaseHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	procs, err := process.Processes(false)
	if err != nil {
		zap.S().Errorf("Failed to get processes: %v", err)
		return
	}

	// 获取端口信息用于关联数据库端口
	portMap := h.buildPortMap()

	versionCache := map[string]string{}
	reported := map[string]bool{} // 防止重复上报同类型数据库

	for _, proc := range procs {
		time.Sleep(process.TraversalInterval)

		comm, err := proc.Comm()
		if err != nil {
			continue
		}

		rule, ok := databaseRuleMap[comm]
		if !ok {
			continue
		}

		// 防止同类型数据库重复上报
		if reported[rule.dbType] {
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

		// 获取版本号
		dbVersion := versionCache[exe+pns]
		if dbVersion == "" {
			dbVersion = h.getVersion(rule, uint32(euid), uint32(egid), dir, exe)
			if dbVersion != "" {
				versionCache[exe+pns] = dbVersion
			}
		}

		// 获取运行用户
		runUser := status.Eusername
		if runUser == "" {
			runUser = status.Rusername
		}

		// 获取监听端口
		dbPort := h.getPort(proc.Pid(), portMap, rule.defaultPort)

		// 上报数据库资产
		c.SendRecord(&businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: map[string]string{
					"db_type":     rule.dbType,
					"db_version":  dbVersion,
					"port":        strconv.Itoa(dbPort),
					"run_user":    runUser,
					"package_seq": seq,
				},
			},
		})

		reported[rule.dbType] = true
		zap.S().Infof("Database collected: type=%s version=%s port=%d user=%s",
			rule.dbType, dbVersion, dbPort, runUser)
	}
}

// isChildProcess 检查是否为子进程
func (h *DatabaseHandler) isChildProcess(comm, ppid string) bool {
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

// getVersion 获取数据库版本
func (h *DatabaseHandler) getVersion(rule *DatabaseRule, uid, gid uint32, dir, exe string) string {
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
		zap.S().Debugf("Failed to get version for %s: %v", rule.dbType, err)
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
	if rule.versionTrimSuffix != "" {
		version = strings.TrimSuffix(version, rule.versionTrimSuffix)
	}
	return version
}

// buildPortMap 构建进程PID到端口的映射
func (h *DatabaseHandler) buildPortMap() map[string]int {
	portMap := make(map[string]int)

	ports, err := port.ListeningPorts()
	if err != nil {
		return portMap
	}

	for _, p := range ports {
		if p.Pid != "" {
			portNum, _ := strconv.Atoi(p.Sport)
			// 只记录第一个端口（主端口）
			if _, exists := portMap[p.Pid]; !exists {
				portMap[p.Pid] = portNum
			}
		}
	}

	return portMap
}

// getPort 获取数据库监听端口
func (h *DatabaseHandler) getPort(pid string, portMap map[string]int, defaultPort int) int {
	if p, ok := portMap[pid]; ok {
		return p
	}
	return defaultPort
}

// findConfigFile 查找配置文件（保留供扩展使用）
func findConfigFile(paths []string) string {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// getMySQLConfigPath 获取MySQL配置文件路径
func getMySQLConfigPath(cmdline string) string {
	// 从命令行参数中提取
	res := regexp.MustCompile(`--defaults-file=\S+`).Find([]byte(cmdline))
	if res != nil {
		return strings.TrimPrefix(string(res), "--defaults-file=")
	}
	// 查找默认路径
	return findConfigFile([]string{
		"/etc/my.cnf",
		"/etc/mysql/my.cnf",
		"/usr/etc/my.cnf",
	})
}

// getPostgreSQLConfigPath 获取PostgreSQL配置文件路径
func getPostgreSQLConfigPath(cmdline string, proc process.Process) string {
	// 从命令行参数中提取
	res := regexp.MustCompile(`config_file=\S+`).Find([]byte(cmdline))
	if res != nil {
		return strings.TrimPrefix(string(res), "config_file=")
	}

	// 从-D参数提取PGDATA
	pgdata := regexp.MustCompile(`-D\s\S+`).Find([]byte(cmdline))
	if pgdata != nil {
		path := filepath.Join(strings.TrimPrefix(string(pgdata), "-D "), "postgresql.conf")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 从环境变量提取PGDATA
	if envs, err := proc.Envs(); err == nil {
		if pgdataEnv, ok := envs["PGDATA"]; ok {
			path := filepath.Join(pgdataEnv, "postgresql.conf")
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}
