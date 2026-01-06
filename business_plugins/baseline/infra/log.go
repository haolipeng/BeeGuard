package infra

import (
	"log"
	"os"
)

var Loger *log.Logger

// 创建插件的日志文件
func init() {
	file := "baseline.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	Loger = log.New(logFile, "[baseline]", log.LstdFlags|log.Lshortfile|log.LUTC)
	return
}
