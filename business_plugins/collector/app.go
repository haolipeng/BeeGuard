package main

import (
	businessplugins "business_plugins/lib"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/process"
	"go.uber.org/zap"
)

var (
	apacheRule = &AppRule{
		name:              "apache",
		_type:             "web_service",
		versionRegex:      regexp.MustCompile(`Apache\/(\d+\.)+\d+`),
		versionArgs:       []string{"-v"},
		versionTrimPrefix: "Apache/",
		confFunc: func(rc RuleContext) string {
			res := regexp.MustCompile(`-f\s\S+`).Find([]byte(rc.cmdline))
			if res != nil {
				return strings.TrimPrefix(string(res), "-f ")
			}
			rootPath := "/"
			for _, path := range []string{
				"/usr/local/apache2/conf/httpd.conf", "/etc/apache2/apache2.conf",
				"/etc/httpd/conf/httpd.conf", "/etc/apache2/httpd.conf"} {
				if _, err := os.Stat(filepath.Join(rootPath, path)); err == nil {
					return path
				}
			}
			return ""
		},
	}
	nginxRule = &AppRule{
		name:              "nginx",
		_type:             "web_service",
		versionRegex:      regexp.MustCompile(`nginx\/(\d+\.)+\d+`),
		versionTrimPrefix: "nginx/",
		versionArgs:       []string{"-v"},
		confFunc: func(rc RuleContext) string {
			res := regexp.MustCompile(`-c\s\S+`).Find([]byte(rc.cmdline))
			if res != nil {
				return strings.TrimPrefix(string(res), "-c ")
			}
			rootPath := "/"
			if _, err := os.Stat(filepath.Join(rootPath, "/etc/nginx/nginx.conf")); err == nil {
				return "/etc/nginx/nginx.conf"
			}
			return ""
		},
	}
	redisRule = &AppRule{
		name:              "redis",
		_type:             "database",
		versionRegex:      regexp.MustCompile(`v=(\d+\.)+\d+`),
		versionArgs:       []string{"-v"},
		versionTrimPrefix: "v=",
		confFunc: func(rc RuleContext) string {
			res := regexp.MustCompile(`\S+\.conf`).Find([]byte(rc.cmdline))
			if res != nil {
				return string(res)
			}
			return ""
		},
	}

	mysqlRule = &AppRule{
		name:              "mysql",
		_type:             "database",
		versionRegex:      regexp.MustCompile(`Ver\s(\d+\.)+\d+\S+`),
		versionArgs:       []string{"-V"},
		versionTrimPrefix: "Ver ",
		confFunc: func(rc RuleContext) string {
			res := regexp.MustCompile(`--defaults-file=\S+`).Find([]byte(rc.cmdline))
			if res != nil {
				return strings.TrimPrefix(string(res), "--defaults-file=")
			}
			rootPath := "/"
			for _, path := range []string{"/etc/my.cnf", "/etc/mysql/my.cnf", "/usr/etc/my.cnf"} {
				if _, err := os.Stat(filepath.Join(rootPath, path)); err == nil {
					return path
				}
			}
			return ""
		},
	}
	postgresqlRule = &AppRule{
		name:         "postgresql",
		_type:        "database",
		versionRegex: regexp.MustCompile(`(\d+\.)+\d+`),
		versionArgs:  []string{"-V"},
		confFunc: func(rc RuleContext) string {
			res := regexp.MustCompile(`config_file=\S+`).Find([]byte(rc.cmdline))
			if res != nil {
				return strings.TrimPrefix(string(res), "config_file=")
			}
			rootPath := "/"
			pgdata := regexp.MustCompile(`-D\s\S+`).Find([]byte(rc.cmdline))
			if pgdata != nil {
				path := filepath.Join(strings.TrimPrefix(string(pgdata), "-D "), "postgresql.conf")
				if _, err := os.Stat(filepath.Join(rootPath, path)); err == nil {
					return path
				}
			}
			if envs, err := rc.proc.Envs(); err == nil {
				if pgdata, ok := envs["PGDATA"]; ok {
					path := filepath.Join(pgdata, "postgresql.conf")
					if _, err := os.Stat(filepath.Join(rootPath, path)); err == nil {
						return path
					}
				}
			}
			return ""
		},
	}
	mongodbRule = &AppRule{
		name:              "mongodb",
		_type:             "database",
		versionRegex:      regexp.MustCompile(`db\sversion\sv(\d+\.)+\d+`),
		versionTrimPrefix: "db version v",
		versionArgs:       []string{"--version"},
		confFunc: func(rc RuleContext) string {
			res := regexp.MustCompile(`--config(=|\s+)\S+`).Find([]byte(rc.cmdline))
			if res != nil {
				return strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(string(res), "--config"), "="))
			}
			res = regexp.MustCompile(`-f\s+\S+`).Find([]byte(rc.cmdline))
			if res != nil {
				return strings.TrimSpace(strings.TrimPrefix(string(res), "-f"))
			}
			return ""
		},
	}

	ruleMap = map[string]*AppRule{
		"apache2":      apacheRule,
		"httpd":        apacheRule,
		"nginx":        nginxRule,
		"redis-server": redisRule,
		"mysqld":       mysqlRule,
		"postgres":     postgresqlRule,
		"mongod":       mongodbRule,
	}
)

type App struct {
	Name    string
	Version string
	Type    string
	Conf    string
	Matched bool
}

type AppRule struct {
	name              string
	versionRegex      *regexp.Regexp
	versionArgs       []string
	_type             string
	versionTrimPrefix string
	versionTrimSuffix string
	confFunc          func(RuleContext) string
}

type RuleContext struct {
	comm       string
	uid        uint32
	gid        uint32
	dir        string
	exe        string
	cmdline    string
	ppid       string
	proc       process.Process
	appVersion string
}

func (r *AppRule) GenerateApp(rc RuleContext) ([]byte, *App) {
	var output []byte
	var app *App
	p, err := process.NewProcess(rc.ppid)
	if err != nil {
		return nil, nil
	}
	//获取父进程名称
	ppidComm, err := p.Comm()
	if err != nil {
		return nil, nil
	}
	//如果当前进程名称和其父进程名称相同，则被过滤掉
	if ppidComm == rc.comm {
		return nil, nil
	}

	app = &App{}
	app.Name = r.name
	app.Type = r._type
	if rc.appVersion != "" {
		//优先使用缓存的版本号
		app.Version = rc.appVersion
	} else {
		//执行命令获取版本号
		if r.versionRegex != nil {
			var err error
			output, err = ExecAs(rc.uid, rc.gid, rc.dir, rc.exe, r.versionArgs...)
			if err != nil {
				zap.S().Warnf("app exec failed: %v", err)
				return nil, app
			}
			//使用正则表达式提取版本号
			res := r.versionRegex.Find(output)
			if res != nil {
				app.Version = string(res)
				//清理版本的前缀
				if r.versionTrimPrefix != "" {
					app.Version = strings.TrimPrefix(app.Version, r.versionTrimPrefix)
				}
				//清理版本的后缀
				if r.versionTrimSuffix != "" {
					app.Version = strings.TrimSuffix(app.Version, r.versionTrimSuffix)
				}
				app.Matched = true
			}
		}
	}
	if r.confFunc != nil {
		app.Conf = r.confFunc(rc)
	}
	return output, app
}

// ExecAs 使用指定用户执行命令
// uid, gid: 执行命令的用户ID和组ID
// dir: 工作目录
// name: 可执行文件路径
// arg: 命令参数
func ExecAs(uid, gid uint32, dir string, name string, arg ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uid,
			Gid: gid,
		},
	}
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

type AppHandler struct{}

func (h *AppHandler) Name() string {
	return "app"
}
func (h *AppHandler) DataType() int {
	return 5060
}

func (h *AppHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	procs, err := process.Processes(false)
	if err != nil {
		return
	}
	versionCache := map[string]string{}
	for _, proc := range procs {
		time.Sleep(process.TraversalInterval)
		comm, err := proc.Comm()
		if err != nil {
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

		euid, err := strconv.ParseUint(status.Euid, 10, 64)
		if err != nil {
			continue
		}
		egid, err := strconv.ParseUint(status.Egid, 10, 64)
		if err != nil {
			continue
		}
		cmdline, err := proc.Cmdline()
		if err != nil {
			continue
		}
		exe, err := proc.Exe()
		if err != nil {
			continue
		}
		//PID Namespace，用于判断进程是否在容器中
		pns, err := proc.Namespace("pid")
		if err != nil {
			continue
		}
		dir, err := proc.Cwd()
		if err != nil {
			continue
		}

		version := versionCache[exe+pns]
		//根据进程名称，规则
		if rule, ok := ruleMap[comm]; ok {
			_, app := rule.GenerateApp(RuleContext{
				uid:        uint32(euid),
				gid:        uint32(egid),
				ppid:       stat.Ppid,
				exe:        exe,
				cmdline:    cmdline,
				proc:       proc,
				appVersion: version,
				comm:       comm,
				dir:        dir,
			})
			if app != nil {
				versionCache[pns+exe] = version
				c.SendRecord(&businessplugins.Record{
					DataType:  int32(h.DataType()),
					Timestamp: time.Now().Unix(),
					Data: &businessplugins.Payload{
						Fields: map[string]string{
							"name":        app.Name,
							"type":        app.Type,
							"sversion":    app.Version,
							"conf":        app.Conf,
							"pid":         proc.Pid(),
							"exe":         exe,
							"start_time":  stat.StartTime,
							"package_seq": seq,
						},
					},
				})
			}
		}
	}
}
