package infra

import (
	"log"
	"os"
	"path/filepath"
)

var Loger *log.Logger

func init() {
	// 优先使用 LOG_DIR 环境变量指定的目录
	logPath := "baseline.log"
	if logDir := os.Getenv("LOG_DIR"); logDir != "" {
		os.MkdirAll(logDir, 0755)
		logPath = filepath.Join(logDir, "baseline.log")
	}

	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	Loger = log.New(logFile, "[baseline]", log.LstdFlags|log.Lshortfile|log.LUTC)
	return
}
