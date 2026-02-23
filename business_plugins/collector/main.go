package main

import (
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	businessplugins "business_plugins/lib"

	"github.com/go-logr/zapr"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	runOnce  = flag.Bool("run-once", false, "立即执行一次后退出（用于测试）")
	handlers = flag.String("handler", "", "指定Handler名称，多个用逗号分隔（如: process,port,database）")
)

func init() {
	runtime.GOMAXPROCS(8)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	c := businessplugins.New()
	logPath := "collector.log"
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		os.MkdirAll(logDir, 0755)
		logPath = filepath.Join(logDir, "collector.log")
	}
	l := log.New(
		log.Config{
			MaxSize:     1,
			Path:        logPath,
			FileLevel:   zapcore.InfoLevel,
			RemoteLevel: zapcore.ErrorLevel,
			MaxBackups:  10,
			Compress:    true,
			Client:      c,
		},
	)
	defer l.Sync()
	zap.ReplaceGlobals(l)
	e := engine.New(c, zapr.NewLogger(l))

	e.AddHandler(time.Hour, &ProcessHandler{}) //进程
	e.AddHandler(time.Hour, &PortHandler{})    //端口
	//e.AddHandler(time.Hour, &KmodHandler{})          //内核模块
	e.AddHandler(time.Hour*6, &ServiceHandler{})       //服务
	e.AddHandler(time.Hour*6, &SoftwareHandler{})      //软件
	e.AddHandler(time.Hour*6, &UserHandler{})          //账号和用户
	e.AddHandler(time.Hour*6, &EnvSuspiciousHandler{}) //可疑环境变量检测
	e.AddHandler(time.Hour*6, &ContainerHandler{})     //容器
	e.AddHandler(time.Hour*6, &ImageHandler{})         //镜像资产
	e.AddHandler(time.Hour*6, &ImagePackageHandler{})  //镜像软件包
	e.AddHandler(time.Hour*6, &DatabaseHandler{})      //数据库服务
	e.AddHandler(time.Hour*6, &WebServiceHandler{})    //Web服务

	// 判断执行模式
	if *runOnce {
		var names []string
		if *handlers != "" {
			names = strings.Split(*handlers, ",")
		}
		e.RunOnce(names)
		return // 执行完毕退出
	}

	//运行engine引擎
	e.Run()
}
