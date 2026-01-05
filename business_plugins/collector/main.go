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

	e.AddHandler(time.Hour, &ProcessHandler{})
	e.Run()
}
