package main

import (
	"math/rand"
	"runtime"
	"time"

	businessplugins "business_plugins/lib"

	"github.com/go-logr/zapr"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	runtime.GOMAXPROCS(8)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	c := businessplugins.New()
	l := log.New(
		log.Config{
			MaxSize:     1,
			Path:        "collector.log",
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

	e.AddHandler(time.Hour, &ProcessHandler{})         //进程
	e.AddHandler(time.Hour, &PortHandler{})            //端口
	e.AddHandler(time.Hour, &KmodHandler{})            //内核模块
	e.AddHandler(time.Hour*6, &ServiceHandler{})       //服务
	e.AddHandler(time.Hour*6, &SoftwareHandler{})      //软件
	e.AddHandler(time.Hour*6, &UserHandler{})          //账号和用户
	e.AddHandler(time.Hour*6, &EnvSuspiciousHandler{}) //可疑环境变量检测
	e.AddHandler(time.Hour*6, &ContainerHandler{})     //容器

	//运行engine引擎
	e.Run()
}
